package variables

import (
	"testing"
)

func TestNewProcessor(t *testing.T) {
	processor := NewProcessor()
	
	if processor == nil {
		t.Fatal("NewProcessor() returned nil")
	}
	
	if processor.resolvers == nil {
		t.Error("resolvers map not initialized")
	}
	
	if processor.regex == nil {
		t.Error("regex not initialized")
	}
	
	if len(processor.resolvers) != 0 {
		t.Error("resolvers map should be empty initially")
	}
}

func TestRegisterHandler(t *testing.T) {
	processor := NewProcessor()
	
	handler := func(context ...string) string {
		return "test_value"
	}
	
	processor.RegisterHandler("test_key", handler)
	
	if len(processor.resolvers) != 1 {
		t.Errorf("Expected 1 resolver, got %d", len(processor.resolvers))
	}
	
	if processor.resolvers["test_key"] == nil {
		t.Error("Handler not registered")
	}
	
	result := processor.resolvers["test_key"]()
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}
}

func TestProcess_NoVariables(t *testing.T) {
	processor := NewProcessor()
	
	template := "This is a simple string without variables"
	result := processor.Process(template, "test_persona")
	
	if result != template {
		t.Errorf("Expected '%s', got '%s'", template, result)
	}
}

func TestProcess_WithVariables(t *testing.T) {
	processor := NewProcessor()
	
	processor.RegisterHandler("username", func(context ...string) string {
		return "Alex"
	})
	
	processor.RegisterHandler("level", func(context ...string) string {
		return "87%"
	})
	
	processor.RegisterHandler("person", func(context ...string) string {
		if len(context) > 0 {
			return "Test Person"
		}
		return "System"
	})
	
	tests := []struct {
		name     string
		template string
		persona  string
		expected string
	}{
		{
			name:     "single variable",
			template: "Hello %username%!",
			persona:  "test",
			expected: "Hello Alex!",
		},
		{
			name:     "multiple variables",
			template: "User %username% has level %level%",
			persona:  "test",
			expected: "User Alex has level 87%",
		},
		{
			name:     "person variable with context",
			template: "Hello %person%!",
			persona:  "test_persona",
			expected: "Hello Test Person!",
		},
		{
			name:     "mixed variables",
			template: "User %username% (%person%) has level %level%",
			persona:  "test_persona",
			expected: "User Alex (Test Person) has level 87%",
		},
		{
			name:     "unknown variable",
			template: "Hello %unknown%!",
			persona:  "test",
			expected: "Hello %unknown%!",
		},
		{
			name:     "empty template",
			template: "",
			persona:  "test",
			expected: "",
		},
		{
			name:     "template with no variables",
			template: "No variables here",
			persona:  "test",
			expected: "No variables here",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.Process(tt.template, tt.persona)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestProcess_EdgeCases(t *testing.T) {
	processor := NewProcessor()
	
	processor.RegisterHandler("test", func(context ...string) string {
		return "value"
	})
	
	tests := []struct {
		name     string
		template string
		persona  string
		expected string
	}{
		{
			name:     "variable at start",
			template: "%test% at start",
			persona:  "test",
			expected: "value at start",
		},
		{
			name:     "variable at end",
			template: "at end %test%",
			persona:  "test",
			expected: "at end value",
		},
		{
			name:     "consecutive variables",
			template: "%test%%test%",
			persona:  "test",
			expected: "valuevalue",
		},
		{
			name:     "variable with special characters",
			template: "Value: %test%!@#$%",
			persona:  "test",
			expected: "Value: value!@#$%",
		},
		{
			name:     "malformed variable",
			template: "%test",
			persona:  "test",
			expected: "%test",
		},
		{
			name:     "empty variable name",
			template: "%%",
			persona:  "test",
			expected: "%%",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.Process(tt.template, tt.persona)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestProcess_ContextHandling(t *testing.T) {
	processor := NewProcessor()
	
	processor.RegisterHandler("context_test", func(context ...string) string {
		if len(context) > 0 {
			return "context: " + context[0]
		}
		return "no context"
	})
	
	processor.RegisterHandler("person", func(context ...string) string {
		if len(context) > 0 {
			return "Person: " + context[0]
		}
		return "System"
	})
	
	tests := []struct {
		name     string
		template string
		persona  string
		expected string
	}{
		{
			name:     "person variable gets context",
			template: "Hello %person%!",
			persona:  "test_persona",
			expected: "Hello Person: test_persona!",
		},
		{
			name:     "regular variable no context",
			template: "Test %context_test%",
			persona:  "test_persona",
			expected: "Test no context",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.Process(tt.template, tt.persona)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestProcess_ConcurrentAccess(t *testing.T) {
	processor := NewProcessor()
	
	processor.RegisterHandler("concurrent", func(context ...string) string {
		return "safe"
	})
	
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				result := processor.Process("Test %concurrent%", "test")
				if result != "Test safe" {
					t.Errorf("Concurrent access failed: expected 'Test safe', got '%s'", result)
				}
			}
			done <- true
		}()
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRegexPattern(t *testing.T) {
	processor := NewProcessor()
	
	processor.RegisterHandler("test", func(context ...string) string {
		return "matched"
	})
	
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "word characters only",
			template: "%test%",
			expected: "matched",
		},
		{
			name:     "alphanumeric",
			template: "%test123%",
			expected: "%test123%",
		},
		{
			name:     "underscore",
			template: "%test_var%",
			expected: "%test_var%",
		},
		{
			name:     "numbers only",
			template: "%123%",
			expected: "%123%",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.Process(tt.template, "test")
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
