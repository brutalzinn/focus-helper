package integrations

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TestTriggerWebhook_N8NIntegration tests webhook integration with n8n workflow
func TestTriggerWebhook_N8NIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping n8n integration test in short mode")
	}

	// n8n webhook URL for focus-helper integration
	webhookURL := "http://192.168.0.47:5678/webhook-test/4b89be2a-29bf-4fd8-b716-11171b8c60f0"

	tests := []struct {
		name        string
		payload     string
		wantErr     bool
		description string
	}{
		{
			name:        "n8n_hyperfocus_low_level",
			payload:     `{"event":"hyperfocus_level","level":"low","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for low hyperfocus level",
		},
		{
			name:        "n8n_hyperfocus_medium_level",
			payload:     `{"event":"hyperfocus_level","level":"medium","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for medium hyperfocus level",
		},
		{
			name:        "n8n_hyperfocus_high_level",
			payload:     `{"event":"hyperfocus_level","level":"high","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for high hyperfocus level",
		},
		{
			name:        "n8n_hyperfocus_critical_level",
			payload:     `{"event":"hyperfocus_level","level":"critical","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for critical hyperfocus level",
		},
		{
			name:        "n8n_session_start",
			payload:     `{"event":"session_start","session_id":"test-session-123","user":"Piloto-Alfa-Um","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for session start",
		},
		{
			name:        "n8n_session_end",
			payload:     `{"event":"session_end","session_id":"test-session-123","duration":"2h30m","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for session end",
		},
		{
			name:        "n8n_alert_triggered",
			payload:     `{"event":"alert_triggered","alert_level":"high","duration":"1h45m","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for alert triggered",
		},
		{
			name:        "n8n_voice_command",
			payload:     `{"event":"voice_command","command":"mayday","timestamp":"` + time.Now().Format(time.RFC3339) + `","source":"focus-helper"}`,
			wantErr:     false,
			description: "Triggers n8n workflow for voice command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TriggerWebhook(webhookURL, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("TriggerWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				t.Logf("Successfully sent webhook: %s", tt.payload)
			}
		})
	}
}

func TestTriggerWebhook_Integration_RealTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	webhookURL := "http://192.168.0.47:5678/webhook-test/4b89be2a-29bf-4fd8-b716-11171b8c60f0"

	// Test real-time hyperfocus level changes
	levels := []string{"low", "medium", "high", "critical"}

	for i, level := range levels {
		payload := map[string]any{
			"title":     "Hyperfocus Level Change",
			"level":     level,
			"sequence":  i + 1,
			"timestamp": time.Now().Format(time.RFC3339),
			"message":   fmt.Sprintf("Hyperfocus level changed to %s", level),
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}

		err = TriggerWebhook(webhookURL, string(jsonPayload))
		if err != nil {
			t.Errorf("Failed to send webhook for level %s: %v", level, err)
		} else {
			t.Logf("Successfully sent webhook for level: %s", level)
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}
}

func TestTriggerWebhook_Integration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with invalid URL
	invalidURL := "http://192.168.0.47:9999/invalid-webhook"
	payload := `{"title":"Test","level":"low"}`

	err := TriggerWebhook(invalidURL, payload)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	} else {
		t.Logf("Got expected error for invalid URL: %v", err)
	}

	// Test with empty URL (should not error)
	err = TriggerWebhook("", payload)
	if err != nil {
		t.Errorf("Expected no error for empty URL, got %v", err)
	}
}

func TestTriggerWebhook_Integration_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	webhookURL := "http://192.168.0.47:5678/webhook-test/4b89be2a-29bf-4fd8-b716-11171b8c60f0"
	payload := `{"title":"Performance Test","level":"medium","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`

	// Test multiple rapid requests
	start := time.Now()
	successCount := 0
	errorCount := 0

	for i := 0; i < 10; i++ {
		err := TriggerWebhook(webhookURL, payload)
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(start)

	t.Logf("Performance test results:")
	t.Logf("  Total requests: 10")
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Errors: %d", errorCount)
	t.Logf("  Total duration: %v", duration)
	t.Logf("  Average per request: %v", duration/10)

	if errorCount > 5 {
		t.Errorf("Too many errors: %d out of 10 requests failed", errorCount)
	}
}

func TestTriggerWebhook_Integration_DataValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	webhookURL := "http://192.168.0.47:5678/webhook-test/4b89be2a-29bf-4fd8-b716-11171b8c60f0"

	// Test with various data types and structures
	testCases := []struct {
		name    string
		payload map[string]any
	}{
		{
			name: "Basic hyperfocus data",
			payload: map[string]any{
				"title": "Hyperfocus Level",
				"level": "high",
			},
		},
		{
			name: "Extended hyperfocus data",
			payload: map[string]any{
				"title":      "Hyperfocus Level",
				"level":      "critical",
				"duration":   "3h45m",
				"session_id": "session-123",
				"user":       "Piloto-Alfa-Um",
				"timestamp":  time.Now().Format(time.RFC3339),
				"metadata": map[string]any{
					"app_version": "1.0.0",
					"os":          "linux",
					"arch":        "amd64",
				},
			},
		},
		{
			name: "Numeric data",
			payload: map[string]any{
				"title":        "Hyperfocus Level",
				"level":        "medium",
				"duration_min": 120,
				"alert_count":  5,
				"is_active":    true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			err = TriggerWebhook(webhookURL, string(jsonPayload))
			if err != nil {
				t.Errorf("Failed to send webhook: %v", err)
			} else {
				t.Logf("Successfully sent webhook: %s", tc.name)
			}
		})
	}
}
