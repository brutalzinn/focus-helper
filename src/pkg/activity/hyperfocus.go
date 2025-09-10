package activity

import (
	"focus-helper/src/pkg/actions"
	"focus-helper/src/pkg/database"
	"focus-helper/src/pkg/llm"
	"focus-helper/src/pkg/models"
	"focus-helper/src/pkg/state"
	"log"
	"time"
)


type HyperfocusDetector struct {
	appState *state.AppState
}


func NewHyperfocusDetector(appState *state.AppState) *HyperfocusDetector {
	return &HyperfocusDetector{appState: appState}
}


func (hd *HyperfocusDetector) CheckHyperfocus(alertIndex int, usageDuration time.Duration) {
	log.Printf("Checking hyperfocus... %s", time.Now())
	if alertIndex != -1 {
		alertLevel := hd.appState.AppConfig.AlertLevels[alertIndex]
		log.Printf("Checking hyperfocus level... %s", alertLevel.Level)
		interval := alertLevel.Threshold.Duration
		if alertLevel.Tolerance > 0 {
			interval = time.Duration(float64(alertLevel.Threshold.Duration) / alertLevel.Tolerance)
		}
		lastTriggered, ok := hd.appState.LastTriggeredLevels[alertIndex]
		if !ok || time.Since(lastTriggered) >= interval {
			log.Printf("[WARNING] HYPERFOCUS DETECTED: Level %s (Index %d, Duration: %v)",
				alertLevel.Level, alertIndex, usageDuration.Round(time.Second))
			subject := DetectSubject(hd.appState.AppConfig.HyperfocusAssociations)
			database.LogHyperfocusEvent(hd.appState.DB, alertIndex+1, hd.appState.ContinuousUsageStartTime, hd.appState.LastActivityTime, subject)
			if hd.appState.Hyperfocus == nil || hd.appState.Hyperfocus.Level != alertLevel.Level {
				hd.appState.Hyperfocus = &models.HyperfocusState{
					Level:     alertLevel.Level,
					StartTime: time.Now(),
				}
			}
			go actions.ExecuteSequence(alertLevel.Actions)
			if hd.appState.LastTriggeredLevels == nil {
				hd.appState.LastTriggeredLevels = make(map[int]time.Time)
			}
			hd.appState.LastTriggeredLevels[alertIndex] = time.Now()
		}
	}
}


func (hd *HyperfocusDetector) ProgressiveTimeCheck(usageDuration time.Duration) int {
	highestIndex := -1
	for i, level := range hd.appState.AppConfig.AlertLevels {
		if level.Enabled && usageDuration >= level.Threshold.Duration {
			highestIndex = i
		}
	}
	return highestIndex
}


func (hd *HyperfocusDetector) AnalyzeWithAI(history, currentWindow string, usageDuration time.Duration) (int, error) {
	return llm.AnalyzeHyperfocus(
		hd.appState.LLMAdapter,
		hd.appState.Language.Get("detector_prompt"),
		len(hd.appState.AppConfig.AlertLevels),
		history,
		currentWindow,
		usageDuration,
	)
}


func (hd *HyperfocusDetector) HandleIdleTransition() {
	if hd.appState.Hyperfocus != nil {
		idleDuration := time.Since(hd.appState.LastActivityTime)
		hfDuration := time.Since(hd.appState.Hyperfocus.StartTime)
		if idleDuration > 0 && hfDuration > 0 {
			proportion := float64(idleDuration) / float64(hfDuration)
			levels := hd.appState.AppConfig.AlertLevels
			currentIndex := -1
			for i, lvl := range levels {
				if lvl.Level == hd.appState.Hyperfocus.Level {
					currentIndex = i
					break
				}
			}
			if currentIndex > -1 {
				stepsDown := int(proportion * float64(len(levels)))
				newIndex := currentIndex - stepsDown
				if newIndex < 0 {
					log.Println("Idle time fully reset hyperfocus level.")
					hd.resetHyperfocusState()
				} else {
					log.Printf("Idle reduced hyperfocus from %s to %s",
						levels[currentIndex].Level, levels[newIndex].Level)
					hd.appState.Hyperfocus.Level = levels[newIndex].Level
					hd.appState.Hyperfocus.StartTime = time.Now()
				}
			}
		}
	}
}


func (hd *HyperfocusDetector) resetHyperfocusState() {
	hd.appState.Hyperfocus = nil
	hd.appState.WarnedIndexes = make(map[int]bool)
	hd.appState.SubjectFrequency = make(map[string]int)
}
