package activity

import (
	"context"
	"focus-helper/pkg/actions"
	"focus-helper/pkg/database"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/state"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-vgo/robotgo"
)

type Activity struct {
	lastMouseX int
	lastMouseY int
	AppState   *state.AppState
}

func NewActivity(appState *state.AppState) *Activity {
	x, y := robotgo.Location()
	return &Activity{lastMouseX: x, lastMouseY: y, AppState: appState}
}

func (m *Activity) HasActivity() bool {
	currentX, currentY := robotgo.Location()
	if currentX != m.lastMouseX || currentY != m.lastMouseY {
		m.lastMouseX = currentX
		m.lastMouseY = currentY
		return true
	}
	return false
}

func DetectSubject(associations map[string]string) string {
	title := robotgo.GetTitle()
	if title == "" {
		return "Unknown"
	}
	for keyword, activity := range associations {
		if strings.Contains(strings.ToLower(title), strings.ToLower(keyword)) {
			return activity
		}
	}
	return "General Use"
}

func (activityActivity *Activity) ActivityLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Activity monitor started.")
	activityCheckTicker := time.NewTicker(activityActivity.AppState.AppConfig.ActivityCheckRate.Duration)
	defer activityCheckTicker.Stop()
	subjectCheckTicker := time.NewTicker(1 * time.Minute)
	defer subjectCheckTicker.Stop()
	wasActive := !activityActivity.isIdle()
	for {
		select {
		case <-activityCheckTicker.C:
			if activityActivity.HasActivity() {
				activityActivity.AppState.LastActivityTime = time.Now()
			}
			isCurrentlyIdle := activityActivity.isIdle()
			if !isCurrentlyIdle {
				if !wasActive {
					activityActivity.resetState()
				}
				wasActive = true
				usageDuration := time.Since(activityActivity.AppState.ContinuousUsageStartTime)
				alertIndex := -1
				log.Printf("IA DETECTOR ENABLED: %v", activityActivity.AppState.AppConfig.IADetectorEnabled)
				if activityActivity.AppState.AppConfig.IADetectorEnabled && activityActivity.AppState.LLMAdapter != nil {
					history, _ := database.GetRecentHistorySummary(activityActivity.AppState.DB)
					currentWindow := DetectSubject(activityActivity.AppState.AppConfig.HyperfocusAssociations)
					index, err := llm.AnalyzeHyperfocus(
						activityActivity.AppState.LLMAdapter,
						activityActivity.AppState.Language.Get("detector_prompt"),
						len(activityActivity.AppState.AppConfig.AlertLevels),
						history,
						currentWindow,
						usageDuration,
					)
					if err != nil {
						log.Printf("AI detector failed: %v. Using progressive time-based fallback.", err)
						alertIndex = activityActivity.progressiveTimeCheck(usageDuration)
					} else {
						log.Printf("AI Analyst determined alert index: %d", index)
						alertIndex = index
					}
				} else {
					alertIndex = activityActivity.progressiveTimeCheck(usageDuration)
				}
				if alertIndex != -1 && !activityActivity.AppState.WarnedIndexes[alertIndex] {
					alertLevel := activityActivity.AppState.AppConfig.AlertLevels[alertIndex]
					log.Printf("[WARNING] HYPERFOCUS DETECTED: Level %s (Index %d, Duration: %v)", alertLevel.Level, alertIndex, usageDuration.Round(time.Second))
					subject := DetectSubject(activityActivity.AppState.AppConfig.HyperfocusAssociations)
					database.LogHyperfocusEvent(activityActivity.AppState.DB, alertIndex+1, activityActivity.AppState.ContinuousUsageStartTime, activityActivity.AppState.LastActivityTime, subject)
					if activityActivity.AppState.Hyperfocus == nil || activityActivity.AppState.Hyperfocus.Level != alertLevel.Level {
						activityActivity.AppState.Hyperfocus = &models.HyperfocusState{
							Level:     alertLevel.Level,
							StartTime: time.Now(),
						}
					}
					go actions.ExecuteSequence(alertLevel.Actions)
					activityActivity.AppState.WarnedIndexes[alertIndex] = true
				}
			} else {
				if wasActive {
					log.Println("User became idle. Finalizing session.")
					sessionDuration := time.Since(activityActivity.AppState.ContinuousUsageStartTime)
					if sessionDuration >= activityActivity.AppState.AppConfig.HyperfocusMinDuration.Duration {
						mainSubject := activityActivity.getMainSubject()
						database.LogHyperfocusSession(activityActivity.AppState.DB, activityActivity.AppState.ContinuousUsageStartTime, activityActivity.AppState.LastActivityTime, mainSubject)
					} else {
						log.Printf("Session ended, duration (%v) was too short to be logged as hyperfocus.", sessionDuration.Round(time.Second))
					}
				}
				wasActive = false
			}
		case <-subjectCheckTicker.C:
			if !activityActivity.isIdle() {
				subject := DetectSubject(activityActivity.AppState.AppConfig.HyperfocusAssociations)
				activityActivity.AppState.SubjectFrequency[subject]++
				log.Printf("Subject detected: %s. Frequencies: %v", subject, activityActivity.AppState.SubjectFrequency)
			}

		case <-ctx.Done():
			log.Println("Stopping activity monitor due to shutdown signal...")
			return
		}
	}
}

func (m Activity) getMainSubject() string {
	if len(m.AppState.SubjectFrequency) == 0 {
		return DetectSubject(m.AppState.AppConfig.HyperfocusAssociations)
	}
	mainSubject := "Unknown"
	maxCount := 0
	for subject, count := range m.AppState.SubjectFrequency {
		if count > maxCount {
			maxCount = count
			mainSubject = subject
		}
	}
	return mainSubject
}
func (m Activity) progressiveTimeCheck(usageDuration time.Duration) int {
	highestIndex := -1
	for i, level := range m.AppState.AppConfig.AlertLevels {
		if level.Enabled && usageDuration >= level.Threshold.Duration {
			highestIndex = i
		}
	}
	return highestIndex
}

func (m Activity) isIdle() bool {
	return time.Since(m.AppState.LastActivityTime) > m.AppState.AppConfig.IdleTimeout.Duration
}

func (m Activity) resetState() {
	log.Println("User is back. Reset app activityActivity.AppState.AppState.")
	now := time.Now()
	m.AppState.ContinuousUsageStartTime = now
	m.AppState.LastActivityTime = now
	m.AppState.WarnedIndexes = make(map[int]bool)
	m.AppState.Hyperfocus = nil
	m.AppState.SubjectFrequency = make(map[string]int)
	action := models.ActionConfig{
		Type:   models.ActionSpeakIA,
		Prompt: m.AppState.Language.Get("user_idle_back"),
	}
	actions.Execute(action)
}
