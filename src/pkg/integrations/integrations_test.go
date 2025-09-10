package integrations

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTriggerWebhook_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode JSON body: %v", err)
		}
		
		if body["title"] != "test" {
			t.Errorf("Expected title 'test', got %v", body["title"])
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	payload := `{"title":"test","level":"high"}`
	err := TriggerWebhook(server.URL, payload)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestTriggerWebhook_EmptyURL(t *testing.T) {
	payload := `{"title":"test","level":"high"}`
	err := TriggerWebhook("", payload)
	
	if err != nil {
		t.Errorf("Expected no error for empty URL, got %v", err)
	}
}

func TestTriggerWebhook_InvalidURL(t *testing.T) {
	payload := `{"title":"test","level":"high"}`
	err := TriggerWebhook("invalid-url", payload)
	
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestTriggerWebhook_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	
	payload := `{"title":"test","level":"high"}`
	err := TriggerWebhook(server.URL, payload)
	
	if err != nil {
		t.Errorf("Expected no error for server error response, got %v", err)
	}
}

func TestTriggerWebhook_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	payload := `{"title":"test","level":"high"}`
	err := TriggerWebhook(server.URL, payload)
	
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestTriggerWebhook_RequestBody(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		receivedBody = buf.Bytes()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	payload := `{"title":"hyperfocus","level":"critical","duration":"2h30m"}`
	err := TriggerWebhook(server.URL, payload)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if string(receivedBody) != payload {
		t.Errorf("Expected body %s, got %s", payload, string(receivedBody))
	}
}

func TestTriggerWebhook_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
		
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			t.Error("Expected User-Agent header to be set")
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	payload := `{"title":"test"}`
	err := TriggerWebhook(server.URL, payload)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestTriggerWebhook_JSONPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode JSON: %v", err)
			return
		}
		
		expectedFields := []string{"title", "level"}
		for _, field := range expectedFields {
			if _, exists := body[field]; !exists {
				t.Errorf("Expected field '%s' in JSON body", field)
			}
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	payload := `{"title":"Hyperfocus Level","level":"medium"}`
	err := TriggerWebhook(server.URL, payload)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestTriggerWebhook_NetworkError(t *testing.T) {
	payload := `{"title":"test","level":"high"}`
	err := TriggerWebhook("http://localhost:99999", payload)
	
	if err == nil {
		t.Error("Expected network error, got nil")
	}
	
	if !strings.Contains(err.Error(), "connection refused") && !strings.Contains(err.Error(), "no such host") {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

func TestTriggerWebhook_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	payload := `{"title":"test","level":invalid}`
	err := TriggerWebhook(server.URL, payload)
	
	if err != nil {
		t.Errorf("Expected no error for invalid JSON (server should handle it), got %v", err)
	}
}

func TestTriggerWebhook_LargePayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	largeData := make(map[string]string)
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	
	payload, _ := json.Marshal(largeData)
	err := TriggerWebhook(server.URL, string(payload))
	
	if err != nil {
		t.Errorf("Expected no error for large payload, got %v", err)
	}
}
