package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"focus-helper/src/pkg/models"
	"focus-helper/src/pkg/state"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)


func createMockAppState() *state.AppState {
	config := &models.Config{
		ProfileName:              "test",
		Username:                 "TestUser",
		CurrentSessionID:         "test-session-123",
		ContinuousUsageStartTime: time.Now().Add(-2 * time.Hour),
		LastActivityTime:         time.Now().Add(-5 * time.Minute),
		Hyperfocus: &models.HyperfocusState{
			Level:     "medium",
			StartTime: time.Now().Add(-1 * time.Hour),
		},
		AlertLevels: []models.AlertLevel{
			{
				Enabled:   true,
				Level:     "low",
				Threshold: models.Duration{Duration: 30 * time.Minute},
				Tolerance: 0,
				Actions: []models.ActionConfig{
					{Type: models.ActionSound, SoundFile: "alert1.mp3"},
					{Type: models.ActionSpeak, Text: "Low level alert"},
				},
			},
			{
				Enabled:   true,
				Level:     "medium",
				Threshold: models.Duration{Duration: 1 * time.Hour},
				Tolerance: 1.5,
				Actions: []models.ActionConfig{
					{Type: models.ActionSound, SoundFile: "alert2.mp3"},
					{Type: models.ActionSpeak, Text: "Medium level alert"},
				},
			},
			{
				Enabled:   true,
				Level:     "high",
				Threshold: models.Duration{Duration: 2 * time.Hour},
				Tolerance: 2.0,
				Actions: []models.ActionConfig{
					{Type: models.ActionSound, SoundFile: "alert3.mp3"},
					{Type: models.ActionSpeak, Text: "High level alert"},
				},
			},
		},
	}

	return &state.AppState{
		AppConfig:                config,
		CurrentSessionID:         "test-session-123",
		ContinuousUsageStartTime: time.Now().Add(-2 * time.Hour),
		LastActivityTime:         time.Now().Add(-5 * time.Minute),
		Hyperfocus: &models.HyperfocusState{
			Level:     "medium",
			StartTime: time.Now().Add(-1 * time.Hour),
		},
		LastTriggeredLevels: make(map[int]time.Time),
	}
}


type MCPClient struct {
	baseURL string
	client  *http.Client
}

func NewMCPClient(baseURL string) *MCPClient {
	return &MCPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *MCPClient) Call(method string, params interface{}) (*MCPResponse, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/mcp", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var mcpResp MCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
		return nil, err
	}

	return &mcpResp, nil
}

func (c *MCPClient) HealthCheck() error {
	resp, err := c.client.Get(c.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func TestMCPServerIntegration(t *testing.T) {

	appState := createMockAppState()


	server := NewMCPServer(appState)


	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp" {
			server.handleRequest(w, r)
		} else if r.URL.Path == "/health" {
			server.handleHealth(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()


	client := NewMCPClient(testServer.URL)

	t.Run("Health Check", func(t *testing.T) {
		err := client.HealthCheck()
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})

	t.Run("Ping Method", func(t *testing.T) {
		resp, err := client.Call("ping", nil)
		if err != nil {
			t.Fatalf("Ping call failed: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Ping returned error: %v", resp.Error)
		}

		if resp.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC 2.0, got %s", resp.JSONRPC)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be map, got %T", resp.Result)
		}

		if result["pong"] != "focus-helper-mcp" {
			t.Errorf("Expected pong 'focus-helper-mcp', got %v", result["pong"])
		}
	})

	t.Run("Get Session Info", func(t *testing.T) {
		resp, err := client.Call("get_session_info", nil)
		if err != nil {
			t.Fatalf("Get session info call failed: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Get session info returned error: %v", resp.Error)
		}

		sessionInfo, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected session info to be map, got %T", resp.Result)
		}


		if sessionInfo["session_id"] != "test-session-123" {
			t.Errorf("Expected session_id 'test-session-123', got %v", sessionInfo["session_id"])
		}

		if sessionInfo["is_active"] != true {
			t.Errorf("Expected is_active true, got %v", sessionInfo["is_active"])
		}

		if sessionInfo["hyperfocus_level"] != "medium" {
			t.Errorf("Expected hyperfocus_level 'medium', got %v", sessionInfo["hyperfocus_level"])
		}
	})

	t.Run("Get Alert Levels", func(t *testing.T) {
		resp, err := client.Call("get_alert_levels", nil)
		if err != nil {
			t.Fatalf("Get alert levels call failed: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Get alert levels returned error: %v", resp.Error)
		}

		alertLevels, ok := resp.Result.([]interface{})
		if !ok {
			t.Fatalf("Expected alert levels to be array, got %T", resp.Result)
		}

		if len(alertLevels) != 3 {
			t.Errorf("Expected 3 alert levels, got %d", len(alertLevels))
		}


		level1, ok := alertLevels[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected alert level to be map, got %T", alertLevels[0])
		}

		if level1["level"] != "low" {
			t.Errorf("Expected first level 'low', got %v", level1["level"])
		}

		if level1["enabled"] != true {
			t.Errorf("Expected first level enabled true, got %v", level1["enabled"])
		}
	})

	t.Run("Get Hyperfocus Status", func(t *testing.T) {
		resp, err := client.Call("get_hyperfocus_status", nil)
		if err != nil {
			t.Fatalf("Get hyperfocus status call failed: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Get hyperfocus status returned error: %v", resp.Error)
		}

		status, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected status to be map, got %T", resp.Result)
		}

		if status["is_hyperfocus"] != true {
			t.Errorf("Expected is_hyperfocus true, got %v", status["is_hyperfocus"])
		}

		if status["level"] != "medium" {
			t.Errorf("Expected level 'medium', got %v", status["level"])
		}
	})

	t.Run("Trigger Alert - Valid Index", func(t *testing.T) {
		resp, err := client.Call("trigger_alert", map[string]interface{}{
			"alert_index": 0,
		})
		if err != nil {
			t.Fatalf("Trigger alert call failed: %v", err)
		}

		if resp.Error != nil {
			t.Errorf("Trigger alert returned error: %v", resp.Error)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be map, got %T", resp.Result)
		}

		if result["success"] != true {
			t.Errorf("Expected success true, got %v", result["success"])
		}

		if result["level"] != "low" {
			t.Errorf("Expected level 'low', got %v", result["level"])
		}
	})

	t.Run("Trigger Alert - Invalid Index", func(t *testing.T) {
		resp, err := client.Call("trigger_alert", map[string]interface{}{
			"alert_index": 999,
		})
		if err != nil {
			t.Fatalf("Trigger alert call failed: %v", err)
		}

		if resp.Error == nil {
			t.Error("Expected error for invalid alert index")
		}

		if resp.Error.Code != -32602 {
			t.Errorf("Expected error code -32602, got %d", resp.Error.Code)
		}
	})

	t.Run("Trigger Alert - Missing Parameters", func(t *testing.T) {
		resp, err := client.Call("trigger_alert", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Trigger alert call failed: %v", err)
		}

		if resp.Error == nil {
			t.Error("Expected error for missing parameters")
		}

		if resp.Error.Code != -32602 {
			t.Errorf("Expected error code -32602, got %d", resp.Error.Code)
		}
	})

	t.Run("Unknown Method", func(t *testing.T) {
		resp, err := client.Call("unknown_method", nil)
		if err != nil {
			t.Fatalf("Unknown method call failed: %v", err)
		}

		if resp.Error == nil {
			t.Error("Expected error for unknown method")
		}

		if resp.Error.Code != -32601 {
			t.Errorf("Expected error code -32601, got %d", resp.Error.Code)
		}
	})
}

func TestMCPServerConcurrentAccess(t *testing.T) {
	appState := createMockAppState()
	server := NewMCPServer(appState)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp" {
			server.handleRequest(w, r)
		} else if r.URL.Path == "/health" {
			server.handleHealth(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	client := NewMCPClient(testServer.URL)


	numGoroutines := 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {

			var err error
			switch id % 4 {
			case 0:
				_, err = client.Call("ping", nil)
			case 1:
				_, err = client.Call("get_session_info", nil)
			case 2:
				_, err = client.Call("get_alert_levels", nil)
			case 3:
				_, err = client.Call("get_hyperfocus_status", nil)
			}
			results <- err
		}(i)
	}


	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent request failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent requests timed out")
		}
	}
}

func TestMCPServerErrorHandling(t *testing.T) {
	appState := createMockAppState()
	server := NewMCPServer(appState)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp" {
			server.handleRequest(w, r)
		} else if r.URL.Path == "/health" {
			server.handleHealth(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	client := NewMCPClient(testServer.URL)

	t.Run("Invalid JSON Request", func(t *testing.T) {
		resp, err := client.client.Post(testServer.URL+"/mcp", "application/json",
			bytes.NewBufferString("invalid json"))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Wrong HTTP Method", func(t *testing.T) {
		resp, err := client.client.Get(testServer.URL + "/mcp")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}

func TestMCPServerCORS(t *testing.T) {
	appState := createMockAppState()
	server := NewMCPServer(appState)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp" {
			server.handleRequest(w, r)
		} else if r.URL.Path == "/health" {
			server.handleHealth(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()


	req, err := http.NewRequest("OPTIONS", testServer.URL+"/mcp", nil)
	if err != nil {
		t.Fatalf("Failed to create OPTIONS request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS request failed: %v", err)
	}
	defer resp.Body.Close()


	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS header Access-Control-Allow-Origin: *, got %s",
			resp.Header.Get("Access-Control-Allow-Origin"))
	}

	if resp.Header.Get("Access-Control-Allow-Methods") != "POST, OPTIONS" {
		t.Errorf("Expected CORS header Access-Control-Allow-Methods: POST, OPTIONS, got %s",
			resp.Header.Get("Access-Control-Allow-Methods"))
	}
}


func TestExternalAppSimulation(t *testing.T) {
	appState := createMockAppState()
	server := NewMCPServer(appState)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp" {
			server.handleRequest(w, r)
		} else if r.URL.Path == "/health" {
			server.handleHealth(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	client := NewMCPClient(testServer.URL)

	t.Run("External App Workflow", func(t *testing.T) {

		err := client.HealthCheck()
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}


		sessionResp, err := client.Call("get_session_info", nil)
		if err != nil {
			t.Fatalf("Get session info failed: %v", err)
		}
		if sessionResp.Error != nil {
			t.Fatalf("Session info error: %v", sessionResp.Error)
		}

		sessionInfo := sessionResp.Result.(map[string]interface{})
		t.Logf("Current session: %+v", sessionInfo)


		hyperfocusResp, err := client.Call("get_hyperfocus_status", nil)
		if err != nil {
			t.Fatalf("Get hyperfocus status failed: %v", err)
		}
		if hyperfocusResp.Error != nil {
			t.Fatalf("Hyperfocus status error: %v", hyperfocusResp.Error)
		}

		hyperfocusStatus := hyperfocusResp.Result.(map[string]interface{})
		t.Logf("Hyperfocus status: %+v", hyperfocusStatus)


		alertsResp, err := client.Call("get_alert_levels", nil)
		if err != nil {
			t.Fatalf("Get alert levels failed: %v", err)
		}
		if alertsResp.Error != nil {
			t.Fatalf("Alert levels error: %v", alertsResp.Error)
		}

		alertLevels := alertsResp.Result.([]interface{})
		t.Logf("Available alert levels: %+v", alertLevels)


		sessionDuration := sessionInfo["duration"].(float64)
		duration := time.Duration(sessionDuration)

		var alertIndex int
		if duration > 2*time.Hour {
			alertIndex = 2 // High alert
		} else if duration > 1*time.Hour {
			alertIndex = 1 // Medium alert
		} else {
			alertIndex = 0 // Low alert
		}

		triggerResp, err := client.Call("trigger_alert", map[string]interface{}{
			"alert_index": alertIndex,
		})
		if err != nil {
			t.Fatalf("Trigger alert failed: %v", err)
		}
		if triggerResp.Error != nil {
			t.Fatalf("Trigger alert error: %v", triggerResp.Error)
		}

		triggerResult := triggerResp.Result.(map[string]interface{})
		t.Logf("Alert triggered: %+v", triggerResult)


		if triggerResult["success"] != true {
			t.Errorf("Expected alert trigger success, got %v", triggerResult["success"])
		}
	})
}

func BenchmarkMCPServerRequests(b *testing.B) {
	appState := createMockAppState()
	server := NewMCPServer(appState)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mcp" {
			server.handleRequest(w, r)
		} else if r.URL.Path == "/health" {
			server.handleHealth(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer testServer.Close()

	client := NewMCPClient(testServer.URL)

	b.Run("Ping", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.Call("ping", nil)
			if err != nil {
				b.Fatalf("Ping failed: %v", err)
			}
		}
	})

	b.Run("GetSessionInfo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.Call("get_session_info", nil)
			if err != nil {
				b.Fatalf("Get session info failed: %v", err)
			}
		}
	})

	b.Run("GetAlertLevels", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.Call("get_alert_levels", nil)
			if err != nil {
				b.Fatalf("Get alert levels failed: %v", err)
			}
		}
	})
}
