package activity

import (
	"focus-helper/pkg/config"
	"focus-helper/pkg/database"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"log"
	"time"
)

func (activityMonitor *Monitor) MonitorActivityLoop() {
	activityCheckTicker := time.NewTicker(activityMonitor.deps.AppConfig.ActivityCheckRate.Duration)
	defer activityCheckTicker.Stop()
	subjectCheckTicker := time.NewTicker(1 * time.Minute)
	defer subjectCheckTicker.Stop()
	wasActive := !activityMonitor.isIdle()
	for {
		select {
		case <-activityCheckTicker.C:
			if activityMonitor.HasActivity() {
				activityMonitor.deps.AppState.LastActivityTime = time.Now()
			}
			isCurrentlyIdle := activityMonitor.isIdle()
			if !isCurrentlyIdle {
				if !wasActive {
					activityMonitor.resetState()
				}
				wasActive = true
				usageDuration := time.Since(activityMonitor.deps.AppState.ContinuousUsageStartTime)
				var alertIndex = -1
				if activityMonitor.deps.AppConfig.IADetectorEnabled && activityMonitor.deps.AppState.LLMAdapter != nil {
					history, _ := database.GetRecentHistorySummary(activityMonitor.deps.DB)
					currentWindow := DetectSubject(activityMonitor.deps.AppConfig.HyperfocusAssociations)
					index, err := llm.AnalyzeHyperfocus(
						activityMonitor.deps.LLMAdapter,
						activityMonitor.deps.AppState.Language.Get("detector_prompt"),
						len(activityMonitor.deps.AppConfig.AlertLevels),
						history,
						currentWindow,
						usageDuration,
					)
					if err != nil {
						log.Printf("AI detector failed: %v. Using progressive time-based fallback.", err)
						alertIndex = activityMonitor.progressiveTimeCheck(usageDuration)
					} else {
						log.Printf("AI Analyst determined alert index: %d", index)
						alertIndex = index
					}
				} else {
					alertIndex = activityMonitor.progressiveTimeCheck(usageDuration)
				}
				if alertIndex != -1 && !activityMonitor.deps.AppState.WarnedIndexes[alertIndex] {
					alertLevel := activityMonitor.deps.AppConfig.AlertLevels[alertIndex]
					log.Printf("[WARNING] HYPERFOCUS DETECTED: Level %s (Index %d, Duration: %v)", alertLevel.Level, alertIndex, usageDuration.Round(time.Second))
					subject := DetectSubject(activityMonitor.deps.AppConfig.HyperfocusAssociations)
					database.LogHyperfocusEvent(activityMonitor.deps.DB, alertIndex+1, activityMonitor.deps.AppState.ContinuousUsageStartTime, activityMonitor.deps.AppState.LastActivityTime, subject)
					if activityMonitor.deps.AppState.Hyperfocus == nil || activityMonitor.deps.AppState.Hyperfocus.Level != alertLevel.Level {
						activityMonitor.deps.AppState.Hyperfocus = &models.HyperfocusState{
							Level:     alertLevel.Level,
							StartTime: time.Now(),
						}
					}
					for _, action := range alertLevel.Actions {
						go activityMonitor.deps.ActionExecutor.Execute(action)
					}
					activityMonitor.deps.AppState.WarnedIndexes[alertIndex] = true
				}
			} else {
				if wasActive {
					log.Println("User became idle. Finalizing session.")
					sessionDuration := time.Since(activityMonitor.deps.AppState.ContinuousUsageStartTime)
					if sessionDuration >= activityMonitor.deps.AppConfig.HyperfocusMinDuration.Duration {
						mainSubject := activityMonitor.getMainSubject()
						database.LogHyperfocusSession(activityMonitor.deps.DB, activityMonitor.deps.AppState.ContinuousUsageStartTime, activityMonitor.deps.AppState.LastActivityTime, mainSubject)
					} else {
						log.Printf("Session ended, duration (%v) was too short to be logged as hyperfocus.", sessionDuration.Round(time.Second))
					}
				}
				wasActive = false
			}

		case <-subjectCheckTicker.C:
			if !activityMonitor.isIdle() {
				subject := DetectSubject(activityMonitor.deps.AppConfig.HyperfocusAssociations)
				activityMonitor.deps.AppState.SubjectFrequency[subject]++
				log.Printf("Subject detected: %s. Frequencies: %v", subject, activityMonitor.deps.AppState.SubjectFrequency)
			}
		}
	}
}

func (m Monitor) getMainSubject() string {
	if len(m.deps.AppState.SubjectFrequency) == 0 {
		return DetectSubject(m.deps.AppConfig.HyperfocusAssociations)
	}
	mainSubject := "Unknown"
	maxCount := 0
	for subject, count := range m.deps.AppState.SubjectFrequency {
		if count > maxCount {
			maxCount = count
			mainSubject = subject
		}
	}
	return mainSubject
}
func (m Monitor) progressiveTimeCheck(usageDuration time.Duration) int {
	highestIndex := -1
	for i, level := range m.deps.AppConfig.AlertLevels {
		if level.Enabled && usageDuration >= level.Threshold.Duration {
			highestIndex = i
		}
	}
	return highestIndex
}

func (m Monitor) isIdle() bool {
	return time.Since(m.deps.AppState.LastActivityTime) > m.deps.AppConfig.IdleTimeout.Duration
}

func (m Monitor) resetState() {
	log.Println("User is back. Reset app activityMonitor.deps.AppState.")
	now := time.Now()
	m.deps.AppState.ContinuousUsageStartTime = now
	m.deps.AppState.LastActivityTime = now
	m.deps.AppState.WarnedIndexes = make(map[int]bool)
	m.deps.AppState.Hyperfocus = nil
	m.deps.AppState.SubjectFrequency = make(map[string]int)
	action := models.ActionConfig{
		Type:   config.ActionSpeakIA,
		Prompt: "Informe ao %username% que ele retornou da ociosidade e que seus contadores foram reiniciados.",
	}
	go m.deps.ActionExecutor.Execute(action)
}
