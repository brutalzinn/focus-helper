package config

import (
	"os"
	"testing"
	"time"
)

func TestProcessEnvVars_BooleanFields(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "boolean true with default",
			input:    `"ia_detector_enabled": "${FOCUSHELPER_IA_DETECTOR_ENABLED:true}"`,
			envVars:  map[string]string{},
			expected: `"ia_detector_enabled": "true"`,
		},
		{
			name:     "boolean false with default",
			input:    `"docker_mode": "${FOCUSHELPER_DOCKER_MODE:false}"`,
			envVars:  map[string]string{},
			expected: `"docker_mode": "false"`,
		},
		{
			name:     "boolean true from env var",
			input:    `"ia_detector_enabled": "${FOCUSHELPER_IA_DETECTOR_ENABLED:false}"`,
			envVars:  map[string]string{"FOCUSHELPER_IA_DETECTOR_ENABLED": "true"},
			expected: `"ia_detector_enabled": "true"`,
		},
		{
			name:     "boolean false from env var",
			input:    `"docker_mode": "${FOCUSHELPER_DOCKER_MODE:true}"`,
			envVars:  map[string]string{"FOCUSHELPER_DOCKER_MODE": "false"},
			expected: `"docker_mode": "false"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := processEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestProcessEnvVars_DurationFields(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "duration with default",
			input:    `"hyperfocus_min_duration": "${FOCUSHELPER_HYPERFOCUS_MIN_DURATION:20m0s}"`,
			envVars:  map[string]string{},
			expected: `"hyperfocus_min_duration": "20m0s"`,
		},
		{
			name:     "duration from env var",
			input:    `"hyperfocus_min_duration": "${FOCUSHELPER_HYPERFOCUS_MIN_DURATION:10m0s}"`,
			envVars:  map[string]string{"FOCUSHELPER_HYPERFOCUS_MIN_DURATION": "30m0s"},
			expected: `"hyperfocus_min_duration": "30m0s"`,
		},
		{
			name:     "idle timeout with default",
			input:    `"idle_timeout": "${FOCUSHELPER_IDLE_TIMEOUT:3m0s}"`,
			envVars:  map[string]string{},
			expected: `"idle_timeout": "3m0s"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := processEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestLoadConfigWithEnvVars_Integration(t *testing.T) {

	tempFile := createTempProfilesFile(t, `[
		{
			"name": "test",
			"ia_detector_enabled": "${FOCUSHELPER_IA_DETECTOR_ENABLED:true}",
			"docker_mode": "${FOCUSHELPER_DOCKER_MODE:false}",
			"hyperfocus_min_duration": "${FOCUSHELPER_HYPERFOCUS_MIN_DURATION:20m0s}",
			"database_file": "${FOCUSHELPER_DATABASE_FILE:./test.db}"
		}
	]`)
	defer os.Remove(tempFile)


	os.Setenv("FOCUSHELPER_IA_DETECTOR_ENABLED", "false")
	os.Setenv("FOCUSHELPER_DOCKER_MODE", "true")
	os.Setenv("FOCUSHELPER_HYPERFOCUS_MIN_DURATION", "30m0s")
	os.Setenv("FOCUSHELPER_DATABASE_FILE", "/custom/path.db")
	defer func() {
		os.Unsetenv("FOCUSHELPER_IA_DETECTOR_ENABLED")
		os.Unsetenv("FOCUSHELPER_DOCKER_MODE")
		os.Unsetenv("FOCUSHELPER_HYPERFOCUS_MIN_DURATION")
		os.Unsetenv("FOCUSHELPER_DATABASE_FILE")
	}()

	profiles, err := LoadProfiles(tempFile)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	profile := profiles[0]
	if profile.IADetectorEnabled != false {
		t.Errorf("Expected IADetectorEnabled to be false, got %v", profile.IADetectorEnabled)
	}
	if profile.DockerMode != true {
		t.Errorf("Expected DockerMode to be true, got %v", profile.DockerMode)
	}
	if profile.HyperfocusMinDuration.Duration != 30*time.Minute {
		t.Errorf("Expected HyperfocusMinDuration to be 30m, got %v", profile.HyperfocusMinDuration.Duration)
	}
	if profile.DatabaseFile != "/custom/path.db" {
		t.Errorf("Expected DatabaseFile to be '/custom/path.db', got %s", profile.DatabaseFile)
	}
}

func TestLoadConfigWithDefaultEnvVars(t *testing.T) {

	tempFile := createTempProfilesFile(t, `[
		{
			"name": "test",
			"ia_detector_enabled": "${FOCUSHELPER_IA_DETECTOR_ENABLED:true}",
			"docker_mode": "${FOCUSHELPER_DOCKER_MODE:false}",
			"hyperfocus_min_duration": "${FOCUSHELPER_HYPERFOCUS_MIN_DURATION:20m0s}",
			"database_file": "${FOCUSHELPER_DATABASE_FILE:./default.db}"
		}
	]`)
	defer os.Remove(tempFile)


	profiles, err := LoadProfiles(tempFile)
	if err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	profile := profiles[0]
	if profile.IADetectorEnabled != true {
		t.Errorf("Expected IADetectorEnabled to be true (default), got %v", profile.IADetectorEnabled)
	}
	if profile.DockerMode != false {
		t.Errorf("Expected DockerMode to be false (default), got %v", profile.DockerMode)
	}
	if profile.HyperfocusMinDuration.Duration != 20*time.Minute {
		t.Errorf("Expected HyperfocusMinDuration to be 20m (default), got %v", profile.HyperfocusMinDuration.Duration)
	}
	if profile.DatabaseFile != "./default.db" {
		t.Errorf("Expected DatabaseFile to be './default.db' (default), got %s", profile.DatabaseFile)
	}
}

func createTempProfilesFile(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "test_profiles_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tempFile.Name()
}

func TestProcessStringEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "simple env var",
			input:    "${TEST_VAR}",
			envVars:  map[string]string{"TEST_VAR": "test_value"},
			expected: "test_value",
		},
		{
			name:     "env var with default",
			input:    "${TEST_VAR:default_value}",
			envVars:  map[string]string{},
			expected: "default_value",
		},
		{
			name:     "env var with default when env exists",
			input:    "${TEST_VAR:default_value}",
			envVars:  map[string]string{"TEST_VAR": "env_value"},
			expected: "env_value",
		},
		{
			name:     "multiple env vars",
			input:    "${VAR1} and ${VAR2:default2}",
			envVars:  map[string]string{"VAR1": "value1"},
			expected: "value1 and default2",
		},
		{
			name:     "no env vars",
			input:    "plain text",
			envVars:  map[string]string{},
			expected: "plain text",
		},
		{
			name:     "empty env var",
			input:    "${EMPTY_VAR:fallback}",
			envVars:  map[string]string{"EMPTY_VAR": ""},
			expected: "fallback",
		},
		{
			name:     "nested env vars",
			input:    "${VAR1}${VAR2:default2}",
			envVars:  map[string]string{"VAR1": "prefix_"},
			expected: "prefix_default2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := processEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("processEnvVars() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadConfigWithoutEnvVars(t *testing.T) {
	tempFile := createTempProfilesFile(t, `[
		{
			"name": "test_profile",
			"username": "TestUser",
			"database_file": "./test.db",
			"log_file": "./test.log"
		}
	]`)
	defer os.Remove(tempFile)

	profiles, err := LoadProfiles(tempFile)
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	profile := profiles[0]
	if profile.ProfileName != "test_profile" {
		t.Errorf("Expected ProfileName 'test_profile', got '%s'", profile.ProfileName)
	}
	if profile.Username != "TestUser" {
		t.Errorf("Expected Username 'TestUser', got '%s'", profile.Username)
	}
	if profile.DatabaseFile != "./test.db" {
		t.Errorf("Expected DatabaseFile './test.db', got '%s'", profile.DatabaseFile)
	}
	if profile.LogFile != "./test.log" {
		t.Errorf("Expected LogFile './test.log', got '%s'", profile.LogFile)
	}
}

func TestLoadConfigWithMissingEnvVars(t *testing.T) {
	tempFile := createTempProfilesFile(t, `[
		{
			"name": "test_profile",
			"username": "${MISSING_VAR:default_user}",
			"database_file": "${MISSING_DB:./default.db}"
		}
	]`)
	defer os.Remove(tempFile)

	profiles, err := LoadProfiles(tempFile)
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	profile := profiles[0]
	if profile.Username != "default_user" {
		t.Errorf("Expected Username 'default_user', got '%s'", profile.Username)
	}
	if profile.DatabaseFile != "./default.db" {
		t.Errorf("Expected DatabaseFile './default.db', got '%s'", profile.DatabaseFile)
	}
}
