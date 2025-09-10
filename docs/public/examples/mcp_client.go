package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MCPRequest represents a Model Context Protocol request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a Model Context Protocol response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SessionInfo represents current session information
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

// AlertLevelInfo represents alert level information
type AlertLevelInfo struct {
	Index     int           `json:"index"`
	Level     string        `json:"level"`
	Enabled   bool          `json:"enabled"`
	Threshold time.Duration `json:"threshold"`
	Tolerance float64       `json:"tolerance"`
}

// MCPClient handles communication with the MCP server
type MCPClient struct {
	baseURL string
	client  *http.Client
}

// NewMCPClient creates a new MCP client
func NewMCPClient(baseURL string) *MCPClient {
	return &MCPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// sendRequest sends a request to the MCP server
func (c *MCPClient) sendRequest(method string, params interface{}) (*MCPResponse, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      time.Now().UnixNano(),
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, err
	}

	return &mcpResp, nil
}

// GetSessionInfo retrieves current session information
func (c *MCPClient) GetSessionInfo() (*SessionInfo, error) {
	resp, err := c.sendRequest("get_session_info", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	// Convert result to SessionInfo
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var sessionInfo SessionInfo
	if err := json.Unmarshal(resultBytes, &sessionInfo); err != nil {
		return nil, err
	}

	return &sessionInfo, nil
}

// GetAlertLevels retrieves available alert levels
func (c *MCPClient) GetAlertLevels() ([]AlertLevelInfo, error) {
	resp, err := c.sendRequest("get_alert_levels", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	// Convert result to []AlertLevelInfo
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var alertLevels []AlertLevelInfo
	if err := json.Unmarshal(resultBytes, &alertLevels); err != nil {
		return nil, err
	}

	return alertLevels, nil
}

// TriggerAlert triggers a specific alert level
func (c *MCPClient) TriggerAlert(alertIndex int) error {
	params := map[string]interface{}{
		"alert_index": alertIndex,
	}

	resp, err := c.sendRequest("trigger_alert", params)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	return nil
}

// GetHyperfocusStatus retrieves current hyperfocus status
func (c *MCPClient) GetHyperfocusStatus() (map[string]interface{}, error) {
	resp, err := c.sendRequest("get_hyperfocus_status", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	// Convert result to map
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	return resultMap, nil
}

// Ping tests the connection to the MCP server
func (c *MCPClient) Ping() error {
	resp, err := c.sendRequest("ping", nil)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	return nil
}

func main() {
	// Create MCP client
	client := NewMCPClient("http://localhost:8089")

	// Test connection
	fmt.Println("Testing MCP server connection...")
	if err := client.Ping(); err != nil {
		fmt.Printf("❌ Failed to connect to MCP server: %v\n", err)
		return
	}
	fmt.Println("✅ Connected to MCP server")

	// Get session information
	fmt.Println("\n📊 Current Session Information:")
	sessionInfo, err := client.GetSessionInfo()
	if err != nil {
		fmt.Printf("❌ Failed to get session info: %v\n", err)
	} else {
		fmt.Printf("Session ID: %s\n", sessionInfo.SessionID)
		fmt.Printf("Subject: %s\n", sessionInfo.Subject)
		fmt.Printf("Start Time: %s\n", sessionInfo.StartTime.Format(time.RFC3339))
		fmt.Printf("Current Time: %s\n", sessionInfo.CurrentTime.Format(time.RFC3339))
		fmt.Printf("Duration: %s\n", sessionInfo.Duration.Round(time.Second))
		fmt.Printf("Is Active: %t\n", sessionInfo.IsActive)
		fmt.Printf("Last Activity: %s\n", sessionInfo.LastActivityTime.Format(time.RFC3339))
		fmt.Printf("Idle Duration: %s\n", sessionInfo.IdleDuration.Round(time.Second))
		
		if sessionInfo.HyperfocusLevel != "" {
			fmt.Printf("Hyperfocus Level: %s\n", sessionInfo.HyperfocusLevel)
			fmt.Printf("Hyperfocus Start: %s\n", sessionInfo.HyperfocusStartTime.Format(time.RFC3339))
			fmt.Printf("Hyperfocus Duration: %s\n", sessionInfo.HyperfocusDuration.Round(time.Second))
		}
	}

	// Get alert levels
	fmt.Println("\n🚨 Available Alert Levels:")
	alertLevels, err := client.GetAlertLevels()
	if err != nil {
		fmt.Printf("❌ Failed to get alert levels: %v\n", err)
	} else {
		for _, level := range alertLevels {
			fmt.Printf("Index %d: %s (Enabled: %t, Threshold: %s)\n", 
				level.Index, level.Level, level.Enabled, level.Threshold.Round(time.Minute))
		}
	}

	// Get hyperfocus status
	fmt.Println("\n🎯 Hyperfocus Status:")
	hyperfocusStatus, err := client.GetHyperfocusStatus()
	if err != nil {
		fmt.Printf("❌ Failed to get hyperfocus status: %v\n", err)
	} else {
		for key, value := range hyperfocusStatus {
			fmt.Printf("%s: %v\n", key, value)
		}
	}

	// Example: Trigger alert level 0 (if available)
	if len(alertLevels) > 0 {
		fmt.Printf("\n🔔 Triggering alert level %d (%s)...\n", alertLevels[0].Index, alertLevels[0].Level)
		if err := client.TriggerAlert(alertLevels[0].Index); err != nil {
			fmt.Printf("❌ Failed to trigger alert: %v\n", err)
		} else {
			fmt.Println("✅ Alert triggered successfully")
		}
	}

	fmt.Println("\n✅ MCP client demo completed")
}
