package activity

import (
	"focus-helper/src/pkg/database"
	"focus-helper/src/pkg/state"
	"log"
)


type SessionManager struct {
	appState *state.AppState
}


func NewSessionManager(appState *state.AppState) *SessionManager {
	return &SessionManager{appState: appState}
}


func (sm *SessionManager) CreateSessionIfNeeded() {
	if sm.appState.CurrentSessionID == "" {
		subject := DetectSubject(sm.appState.AppConfig.HyperfocusAssociations)
		sessionID, err := database.CreateSession(sm.appState.DB, subject)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
		} else {
			sm.appState.CurrentSessionID = sessionID
			log.Printf("Created new session: %s for subject: %s", sessionID, subject)
		}
	}
}


func (sm *SessionManager) UpdateSessionSubject(subject string) {
	if sm.appState.CurrentSessionID != "" {
		if err := database.UpdateSessionSubject(sm.appState.DB, sm.appState.CurrentSessionID, subject); err != nil {
			log.Printf("Failed to update session subject: %v", err)
		}
	}
}


func (sm *SessionManager) EndSession() {
	if sm.appState.CurrentSessionID != "" {
		if err := database.EndSession(sm.appState.DB, sm.appState.CurrentSessionID); err != nil {
			log.Printf("Failed to end session: %v", err)
		} else {
			log.Printf("Ended session: %s", sm.appState.CurrentSessionID)
		}
		sm.appState.CurrentSessionID = ""
	}
}
