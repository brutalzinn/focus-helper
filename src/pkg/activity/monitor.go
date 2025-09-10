package activity

import (
	"context"
	"focus-helper/src/pkg/actions"
	"focus-helper/src/pkg/database"
	"focus-helper/src/pkg/models"
	"focus-helper/src/pkg/state"
	"log"
	"sync"
	"time"
)


type Activity struct {
	lastMouseX         int
	lastMouseY         int
	appState           *state.AppState
	sessionManager     *SessionManager
	hyperfocusDetector *HyperfocusDetector
}


func NewActivity(appState *state.AppState) *Activity {
	x, y := robot.Location()
	return &Activity{
		lastMouseX:         x,
		lastMouseY:         y,
		appState:           appState,
		sessionManager:     NewSessionManager(appState),
		hyperfocusDetector: NewHyperfocusDetector(appState),
	}
}


func (m *Activity) HasActivity() bool {
	currentX, currentY := robot.Location()
	if currentX != m.lastMouseX || currentY != m.lastMouseY {
		m.lastMouseX = currentX
		m.lastMouseY = currentY
		return true
	}
	return false
}


func (m *Activity) ActivityLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Activity monitor started.")

	activityCheckTicker := time.NewTicker(m.appState.AppConfig.ActivityCheckRate.Duration)
	defer activityCheckTicker.Stop()

	subjectCheckTicker := time.NewTicker(1 * time.Minute)
	defer subjectCheckTicker.Stop()

	wasActive := !m.isIdle()

	for {
		select {
		case <-activityCheckTicker.C:
			m.handleActivityCheck(wasActive)
			wasActive = !m.isIdle()

		case <-subjectCheckTicker.C:
			m.handleSubjectCheck()

		case <-ctx.Done():
			log.Println("Stopping activity monitor due to shutdown signal...")
			return
		}
	}
}


func (m *Activity) handleActivityCheck(wasActive bool) {
	if m.HasActivity() {
		m.appState.LastActivityTime = time.Now()
	}

	isCurrentlyIdle := m.isIdle()

	if !isCurrentlyIdle {
		if !wasActive {
			m.resetState()
			m.sessionManager.CreateSessionIfNeeded()
		}

		usageDuration := time.Since(m.appState.ContinuousUsageStartTime)
		alertIndex := m.determineAlertIndex(usageDuration)
		m.hyperfocusDetector.CheckHyperfocus(alertIndex, usageDuration)
	} else {
		if wasActive {
			m.handleIdle()
		}
	}
}


func (m *Activity) handleSubjectCheck() {
	if !m.isIdle() {
		subject := DetectSubject(m.appState.AppConfig.HyperfocusAssociations)
		m.appState.SubjectFrequency[subject]++
		log.Printf("Subject detected: %s. Frequencies: %v", subject, m.appState.SubjectFrequency)
		m.sessionManager.UpdateSessionSubject(subject)
	}
}


func (m *Activity) determineAlertIndex(usageDuration time.Duration) int {
	alertIndex := -1

	log.Printf("IA DETECTOR ENABLED: %v", m.appState.AppConfig.IADetectorEnabled)

	if m.appState.AppConfig.IADetectorEnabled && m.appState.LLMAdapter != nil {

		timeSinceLastLLMCall := time.Since(m.appState.LastLLMCallTime)
		llmInterval := m.appState.AppConfig.LLMCallInterval.Duration

		if timeSinceLastLLMCall < llmInterval {
			log.Printf("LLM call interval not reached yet. Time since last call: %v, required interval: %v. Using progressive check.",
				timeSinceLastLLMCall.Round(time.Second), llmInterval)
			return m.hyperfocusDetector.ProgressiveTimeCheck(usageDuration)
		}

		history, _ := database.GetRecentHistorySummary(m.appState.DB)
		if history == "" {
			log.Println("No history found. Skipping AI analyze, using progressive check.")
			return m.hyperfocusDetector.ProgressiveTimeCheck(usageDuration)
		}

		m.appState.AnalyzeMu.Lock()
		if m.appState.IsAnalyzing {
			m.appState.AnalyzeMu.Unlock()
			return -1
		}
		m.appState.IsAnalyzing = true
		m.appState.AnalyzeMu.Unlock()


		m.appState.LastLLMCallTime = time.Now()

		go func() {
			defer func() {
				m.appState.AnalyzeMu.Lock()
				m.appState.IsAnalyzing = false
				m.appState.AnalyzeMu.Unlock()
			}()

			currentWindow := DetectSubject(m.appState.AppConfig.HyperfocusAssociations)
			index, err := m.hyperfocusDetector.AnalyzeWithAI(history, currentWindow, usageDuration)
			if err != nil {
				log.Printf("AI detector failed: %v. Using progressive time-based fallback.", err)
				alertIndex = m.hyperfocusDetector.ProgressiveTimeCheck(usageDuration)
			} else {
				log.Printf("AI Analyst determined alert index: %d", index)
				alertIndex = index
			}
			m.hyperfocusDetector.CheckHyperfocus(alertIndex, usageDuration)
		}()
	} else {
		alertIndex = m.hyperfocusDetector.ProgressiveTimeCheck(usageDuration)
	}

	return alertIndex
}


func (m *Activity) isIdle() bool {
	return time.Since(m.appState.LastActivityTime) > m.appState.AppConfig.IdleTimeout.Duration
}


func (m *Activity) resetState() {
	log.Println("User is back. Reset app state.")
	now := time.Now()
	m.appState.ContinuousUsageStartTime = now
	m.appState.LastActivityTime = now
	m.hyperfocusDetector.resetHyperfocusState()

	action := models.ActionConfig{
		Type:   models.ActionSpeakIA,
		Prompt: m.appState.Language.Get("user_idle_back"),
	}
	actions.Execute(action)
}


func (m *Activity) handleIdle() {
	log.Println("User became idle. Finalizing session.")
	sessionDuration := time.Since(m.appState.ContinuousUsageStartTime)

	m.hyperfocusDetector.HandleIdleTransition()
	m.sessionManager.EndSession()

	if sessionDuration >= m.appState.AppConfig.HyperfocusMinDuration.Duration {
		mainSubject := GetMainSubject(m.appState.SubjectFrequency, m.appState.AppConfig.HyperfocusAssociations)
		database.LogHyperfocusSession(
			m.appState.DB,
			m.appState.ContinuousUsageStartTime,
			m.appState.LastActivityTime,
			mainSubject,
		)
	} else {
		log.Printf("Session ended, duration (%v) too short to log.", sessionDuration.Round(time.Second))
	}
}
