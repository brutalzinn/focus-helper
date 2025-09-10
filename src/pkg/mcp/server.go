package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"focus-helper/src/pkg/actions"
	"focus-helper/src/pkg/database"
	"focus-helper/src/pkg/models"
	"focus-helper/src/pkg/state"
	"log"
	"net/http"
	"sync"
	"time"
)

type MCPRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type SessionInfo struct {
	SessionID           string        `json:"session_id"`
	Subject             string        `json:"subject"`
	StartTime           time.Time     `json:"start_time"`
	CurrentTime         time.Time     `json:"current_time"`
	Duration            time.Duration `json:"duration"`
	IsActive            bool          `json:"is_active"`
	HyperfocusLevel     string        `json:"hyperfocus_level,omitempty"`
	HyperfocusStartTime *time.Time    `json:"hyperfocus_start_time,omitempty"`
	HyperfocusDuration  time.Duration `json:"hyperfocus_duration,omitempty"`
	LastActivityTime    time.Time     `json:"last_activity_time"`
	IdleDuration        time.Duration `json:"idle_duration"`
}

type AlertLevelInfo struct {
	Index     int           `json:"index"`
	Level     string        `json:"level"`
	Enabled   bool          `json:"enabled"`
	Threshold time.Duration `json:"threshold"`
	Tolerance float64       `json:"tolerance"`
}

type HyperfocusStatus struct {
	IsActive      bool          `json:"is_active"`
	Level         string        `json:"level,omitempty"`
	StartTime     *time.Time    `json:"start_time,omitempty"`
	Duration      time.Duration `json:"duration,omitempty"`
	ThresholdMet  bool          `json:"threshold_met"`
	NextThreshold time.Duration `json:"next_threshold,omitempty"`
}

type AppStatus struct {
	IsRunning          bool           `json:"is_running"`
	Uptime             time.Duration  `json:"uptime"`
	LastActivityTime   time.Time      `json:"last_activity_time"`
	IdleDuration       time.Duration  `json:"idle_duration"`
	TotalSessions      int            `json:"total_sessions"`
	ActiveSessionCount int            `json:"active_session_count"`
	HyperfocusCount    int            `json:"hyperfocus_count"`
	LastAlertTime      *time.Time     `json:"last_alert_time,omitempty"`
	Configuration      *models.Config `json:"configuration,omitempty"`
}

type MCPServer struct {
	appState  *state.AppState
	db        *sql.DB
	mu        sync.RWMutex
	startTime time.Time
}

func NewMCPServer(appState *state.AppState, db *sql.DB) *MCPServer {
	return &MCPServer{
		appState:  appState,
		db:        db,
		startTime: time.Now(),
	}
}

func (s *MCPServer) Start(port int) error {
	http.HandleFunc("/mcp", s.handleRequest)
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/status", s.handleStatus)

	log.Printf("MCP server starting on port %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *MCPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "focus-helper-mcp",
		"version": "1.0.0",
	})
}

func (s *MCPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	status := s.getAppStatus()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (s *MCPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	if req.JSONRPC != "2.0" {
		s.sendError(w, req.ID, -32600, "Invalid Request", "jsonrpc must be '2.0'")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	response := s.handleMethod(req)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *MCPServer) handleMethod(req MCPRequest) MCPResponse {
	switch req.Method {
	case "get_session_info":
		return s.handleGetSessionInfo(req)
	case "get_alert_levels":
		return s.handleGetAlertLevels(req)
	case "trigger_alert":
		return s.handleTriggerAlert(req)
	case "trigger_hyperfocus":
		return s.handleTriggerHyperfocus(req)
	case "get_hyperfocus_status":
		return s.handleGetHyperfocusStatus(req)
	case "get_app_status":
		return s.handleGetAppStatus(req)
	case "start_session":
		return s.handleStartSession(req)
	case "end_session":
		return s.handleEndSession(req)
	case "update_session_subject":
		return s.handleUpdateSessionSubject(req)
	case "get_session_history":
		return s.handleGetSessionHistory(req)
	case "get_wellbeing_questions":
		return s.handleGetWellbeingQuestions(req)
	case "submit_wellbeing_response":
		return s.handleSubmitWellbeingResponse(req)
	case "get_statistics":
		return s.handleGetStatistics(req)
	case "ping":
		return s.handlePing(req)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *MCPServer) handleGetSessionInfo(req MCPRequest) MCPResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	sessionInfo := SessionInfo{
		SessionID:        s.appState.CurrentSessionID,
		Subject:          "Unknown",
		StartTime:        now,
		CurrentTime:      now,
		Duration:         0,
		IsActive:         s.appState.CurrentSessionID != "",
		LastActivityTime: s.appState.LastActivityTime,
		IdleDuration:     now.Sub(s.appState.LastActivityTime),
	}

	if s.appState.CurrentSessionID != "" {
		session, err := database.GetCurrentSessionWithTimeout(s.db, 24*time.Hour)
		if err == nil && session != nil {
			sessionInfo.Subject = session.Subject
			sessionInfo.StartTime = session.StartTime
			sessionInfo.Duration = now.Sub(session.StartTime)
		} else {
			sessionInfo.Subject = "General Use"
			sessionInfo.StartTime = s.appState.ContinuousUsageStartTime
			sessionInfo.Duration = now.Sub(s.appState.ContinuousUsageStartTime)
		}
	}

	if s.appState.Hyperfocus != nil {
		sessionInfo.HyperfocusLevel = s.appState.Hyperfocus.Level
		sessionInfo.HyperfocusStartTime = &s.appState.Hyperfocus.StartTime
		sessionInfo.HyperfocusDuration = now.Sub(s.appState.Hyperfocus.StartTime)
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  sessionInfo,
	}
}

func (s *MCPServer) handleGetAlertLevels(req MCPRequest) MCPResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alertLevels := make([]AlertLevelInfo, len(s.appState.AppConfig.AlertLevels))
	for i, level := range s.appState.AppConfig.AlertLevels {
		alertLevels[i] = AlertLevelInfo{
			Index:     i,
			Level:     level.Level,
			Enabled:   level.Enabled,
			Threshold: level.Threshold.Duration,
			Tolerance: level.Tolerance,
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  alertLevels,
	}
}

func (s *MCPServer) handleTriggerAlert(req MCPRequest) MCPResponse {
	params, ok := req.Params.(map[string]any)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	indexFloat, ok := params["level_index"].(float64)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params: level_index required",
			},
		}
	}

	index := int(indexFloat)
	if index < 0 || index >= len(s.appState.AppConfig.AlertLevels) {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid level_index",
			},
		}
	}

	alertLevel := s.appState.AppConfig.AlertLevels[index]
	if !alertLevel.Enabled {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Alert level is disabled",
			},
		}
	}

	s.mu.Lock()
	s.triggerHyperfocusLevel(alertLevel.Level, index)
	s.mu.Unlock()

	go actions.ExecuteSequence(alertLevel.Actions)

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"success":      true,
			"level":        alertLevel.Level,
			"level_index":  index,
			"triggered_at": time.Now(),
		},
	}
}

func (s *MCPServer) handleTriggerHyperfocus(req MCPRequest) MCPResponse {
	params, ok := req.Params.(map[string]any)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	level, ok := params["level"].(string)
	if !ok || level == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params: level required",
			},
		}
	}

	index := -1
	for i, alertLevel := range s.appState.AppConfig.AlertLevels {
		if alertLevel.Level == level {
			index = i
			break
		}
	}

	if index == -1 {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid level: " + level,
			},
		}
	}

	alertLevel := s.appState.AppConfig.AlertLevels[index]
	if !alertLevel.Enabled {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Alert level is disabled",
			},
		}
	}

	s.mu.Lock()
	s.triggerHyperfocusLevel(alertLevel.Level, index)
	s.mu.Unlock()

	go actions.ExecuteSequence(alertLevel.Actions)

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"success":      true,
			"level":        alertLevel.Level,
			"level_index":  index,
			"triggered_at": time.Now(),
		},
	}
}

func (s *MCPServer) handleGetHyperfocusStatus(req MCPRequest) MCPResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	status := HyperfocusStatus{
		IsActive: s.appState.Hyperfocus != nil,
	}

	if s.appState.Hyperfocus != nil {
		status.Level = s.appState.Hyperfocus.Level
		status.StartTime = &s.appState.Hyperfocus.StartTime
		status.Duration = now.Sub(s.appState.Hyperfocus.StartTime)

		for i, level := range s.appState.AppConfig.AlertLevels {
			if level.Enabled && now.Sub(s.appState.ContinuousUsageStartTime) >= level.Threshold.Duration {
				status.ThresholdMet = true
				if i+1 < len(s.appState.AppConfig.AlertLevels) {
					nextLevel := s.appState.AppConfig.AlertLevels[i+1]
					if nextLevel.Enabled {
						status.NextThreshold = nextLevel.Threshold.Duration - now.Sub(s.appState.ContinuousUsageStartTime)
					}
				}
				break
			}
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  status,
	}
}

func (s *MCPServer) handleGetAppStatus(req MCPRequest) MCPResponse {
	status := s.getAppStatus()
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  status,
	}
}

func (s *MCPServer) handleStartSession(req MCPRequest) MCPResponse {
	params, ok := req.Params.(map[string]any)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	subject := "General Use"
	if subj, ok := params["subject"].(string); ok && subj != "" {
		subject = subj
	}

	sessionID, err := database.CreateSession(s.db, subject)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to create session",
				Data:    err.Error(),
			},
		}
	}

	s.appState.CurrentSessionID = sessionID
	s.appState.ContinuousUsageStartTime = time.Now()

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"session_id": sessionID,
			"subject":    subject,
			"started_at": time.Now(),
		},
	}
}

func (s *MCPServer) handleEndSession(req MCPRequest) MCPResponse {
	if s.appState.CurrentSessionID == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "No active session to end",
			},
		}
	}

	err := database.EndSession(s.db, s.appState.CurrentSessionID)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to end session",
				Data:    err.Error(),
			},
		}
	}

	sessionID := s.appState.CurrentSessionID
	s.appState.CurrentSessionID = ""

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"session_id": sessionID,
			"ended_at":   time.Now(),
		},
	}
}

func (s *MCPServer) handleUpdateSessionSubject(req MCPRequest) MCPResponse {
	params, ok := req.Params.(map[string]any)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	subject, ok := params["subject"].(string)
	if !ok || subject == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params: subject required",
			},
		}
	}

	if s.appState.CurrentSessionID == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "No active session to update",
			},
		}
	}

	err := database.UpdateSessionSubject(s.db, s.appState.CurrentSessionID, subject)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to update session subject",
				Data:    err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"session_id": s.appState.CurrentSessionID,
			"subject":    subject,
			"updated_at": time.Now(),
		},
	}
}

func (s *MCPServer) handleGetSessionHistory(req MCPRequest) MCPResponse {
	params, ok := req.Params.(map[string]any)
	if !ok {
		params = make(map[string]any)
	}

	limit := 10
	if l, ok := params["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	offset := 0
	if o, ok := params["offset"].(float64); ok && o >= 0 {
		offset = int(o)
	}

	sessions, err := database.GetSessions(s.db, limit, offset)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to get session history",
				Data:    err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  sessions,
	}
}

func (s *MCPServer) handleGetWellbeingQuestions(req MCPRequest) MCPResponse {
	enabled := s.appState.AppConfig.WellbeingQuestionsEnabled
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"enabled": enabled,
		},
	}
}

func (s *MCPServer) handleSubmitWellbeingResponse(req MCPRequest) MCPResponse {
	params, ok := req.Params.(map[string]any)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	questionID, ok := params["question_id"].(string)
	if !ok || questionID == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params: question_id required",
			},
		}
	}

	response, ok := params["response"].(string)
	if !ok || response == "" {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params: response required",
			},
		}
	}

	err := database.SaveWellbeingResponse(s.db, questionID, response, s.appState.CurrentSessionID)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to save wellbeing response",
				Data:    err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"question_id":  questionID,
			"response":     response,
			"submitted_at": time.Now(),
		},
	}
}

func (s *MCPServer) handleGetStatistics(req MCPRequest) MCPResponse {
	stats, err := database.GetStatistics(s.db)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to get statistics",
				Data:    err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  stats,
	}
}

func (s *MCPServer) handlePing(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]string{
			"pong": "focus-helper-mcp",
		},
	}
}

func (s *MCPServer) getAppStatus() AppStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	status := AppStatus{
		IsRunning:        true,
		Uptime:           now.Sub(s.startTime),
		LastActivityTime: s.appState.LastActivityTime,
		IdleDuration:     now.Sub(s.appState.LastActivityTime),
		Configuration:    s.appState.AppConfig,
	}

	if s.appState.Hyperfocus != nil {
		status.LastAlertTime = &s.appState.Hyperfocus.StartTime
	}

	totalSessions, _ := database.GetTotalSessions(s.db)
	status.TotalSessions = totalSessions

	activeSessions, _ := database.GetActiveSessions(s.db)
	status.ActiveSessionCount = len(activeSessions)

	hyperfocusCount, _ := database.GetHyperfocusCount(s.db)
	status.HyperfocusCount = hyperfocusCount

	return status
}

func (s *MCPServer) triggerHyperfocusLevel(level string, index int) {
	now := time.Now()

	if s.appState.Hyperfocus == nil {
		s.appState.Hyperfocus = &models.HyperfocusState{
			Level:     level,
			StartTime: now,
		}
	} else {
		s.appState.Hyperfocus.Level = level
		s.appState.Hyperfocus.StartTime = now
	}

	log.Printf("Hyperfocus level triggered: %s (index: %d)", level, index)

	if s.appState.CurrentSessionID != "" {
		database.LogHyperfocusEvent(s.db, index, now, now, "MCP Triggered")
	}
}

func (s *MCPServer) sendError(w http.ResponseWriter, id any, code int, message, data string) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
