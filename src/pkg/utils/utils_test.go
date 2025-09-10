package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		hUnit    string
		mUnit    string
		sUnit    string
		expected string
	}{
		{
			name:     "hours, minutes, seconds",
			duration: 2*time.Hour + 30*time.Minute + 45*time.Second,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "2h 30m 45s",
		},
		{
			name:     "minutes and seconds only",
			duration: 30*time.Minute + 45*time.Second,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "30m 45s",
		},
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "45s",
		},
		{
			name:     "zero duration",
			duration: 0,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "0s",
		},
		{
			name:     "custom units",
			duration: 1*time.Hour + 15*time.Minute + 30*time.Second,
			hUnit:    " hours",
			mUnit:    " minutes",
			sUnit:    " seconds",
			expected: "1 hours 15 minutes 30 seconds",
		},
		{
			name:     "exact hour",
			duration: 2 * time.Hour,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "2h 0m 0s",
		},
		{
			name:     "exact minute",
			duration: 5 * time.Minute,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "5m 0s",
		},
		{
			name:     "large duration",
			duration: 25*time.Hour + 30*time.Minute + 15*time.Second,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "25h 30m 15s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration, tt.hUnit, tt.mUnit, tt.sUnit)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{
			name:     "zero seconds",
			seconds:  0,
			expected: "0:00",
		},
		{
			name:     "single digit seconds",
			seconds:  5,
			expected: "0:05",
		},
		{
			name:     "double digit seconds",
			seconds:  45,
			expected: "0:45",
		},
		{
			name:     "single digit minutes",
			seconds:  65,
			expected: "1:05",
		},
		{
			name:     "double digit minutes",
			seconds:  125,
			expected: "2:05",
		},
		{
			name:     "exact minute",
			seconds:  60,
			expected: "1:00",
		},
		{
			name:     "exact minutes",
			seconds:  300,
			expected: "5:00",
		},
		{
			name:     "large number",
			seconds:  3661,
			expected: "61:01",
		},
		{
			name:     "negative seconds",
			seconds:  -30,
			expected: "-0:30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(tt.seconds)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerateRandomOdd(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{
			name: "positive range",
			min:  1,
			max:  10,
		},
		{
			name: "single number range",
			min:  5,
			max:  5,
		},
		{
			name: "swapped min max",
			min:  10,
			max:  1,
		},
		{
			name: "negative range",
			min:  -10,
			max:  -1,
		},
		{
			name: "zero to positive",
			min:  0,
			max:  5,
		},
		{
			name: "large range",
			min:  1,
			max:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				result := GenerateRandomOdd(tt.min, tt.max)

				expectedMin := tt.min
				expectedMax := tt.max
				if tt.min > tt.max {
					expectedMin = tt.max
					expectedMax = tt.min
				}

				if result < expectedMin || result > expectedMax {
					t.Errorf("Result %d is outside range [%d, %d]", result, expectedMin, expectedMax)
				}

				if result%2 == 0 {
					t.Errorf("Result %d is not odd", result)
				}
			}
		})
	}
}

func TestGenerateRandomSquareIntervalVaried(t *testing.T) {
	tests := []struct {
		name   string
		minSec int
		maxSec int
	}{
		{
			name:   "normal range",
			minSec: 0,
			maxSec: 100,
		},
		{
			name:   "small range",
			minSec: 5,
			maxSec: 15,
		},
		{
			name:   "single second range",
			minSec: 10,
			maxSec: 10,
		},
		{
			name:   "swapped range",
			minSec: 100,
			maxSec: 0,
		},
		{
			name:   "negative range",
			minSec: -10,
			maxSec: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				start, end := GenerateRandomSquareIntervalVaried(tt.minSec, tt.maxSec)

				expectedMin := tt.minSec
				expectedMax := tt.maxSec
				if tt.minSec > tt.maxSec {
					expectedMin = tt.maxSec
					expectedMax = tt.minSec
				}

				if start < expectedMin || start > expectedMax {
					t.Errorf("Start %d is outside range [%d, %d]", start, expectedMin, expectedMax)
				}

				if end < start {
					t.Errorf("End %d is less than start %d", end, start)
				}

				if end > expectedMax {
					t.Errorf("End %d exceeds max %d", end, expectedMax)
				}

				intervalLength := end - start
				validLengths := []int{10, 15, 20, 30}
				validLength := false
				for _, validLen := range validLengths {
					if intervalLength == validLen {
						validLength = true
						break
					}
				}

				if !validLength && intervalLength > 0 {
					t.Errorf("Interval length %d is not valid (should be 10, 15, 20, or 30)", intervalLength)
				}
			}
		})
	}
}

func TestClearTempAudioOnExit(t *testing.T) {
	tempDir := t.TempDir()

	originalConfigPath := os.Getenv("FOCUSHELPER_CONFIG_PATH")
	defer func() {
		if originalConfigPath != "" {
			os.Setenv("FOCUSHELPER_CONFIG_PATH", originalConfigPath)
		} else {
			os.Unsetenv("FOCUSHELPER_CONFIG_PATH")
		}
	}()

	os.Setenv("FOCUSHELPER_CONFIG_PATH", tempDir)

	tempAudioDir := filepath.Join(tempDir, "temp_audio")
	err := os.MkdirAll(tempAudioDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp audio dir: %v", err)
	}

	testFile := filepath.Join(tempAudioDir, "test.wav")
	err = os.WriteFile(testFile, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Test file should exist before clearing")
	}

	ClearTempAudioOnExit()

	if _, err := os.Stat(tempAudioDir); !os.IsNotExist(err) {
		t.Error("Temp audio directory should be removed after clearing")
	}
}

func TestClearTempAudioOnExit_NonExistentDir(t *testing.T) {
	tempDir := t.TempDir()

	originalConfigPath := os.Getenv("FOCUSHELPER_CONFIG_PATH")
	defer func() {
		if originalConfigPath != "" {
			os.Setenv("FOCUSHELPER_CONFIG_PATH", originalConfigPath)
		} else {
			os.Unsetenv("FOCUSHELPER_CONFIG_PATH")
		}
	}()

	os.Setenv("FOCUSHELPER_CONFIG_PATH", tempDir)

	ClearTempAudioOnExit()

	tempAudioDir := filepath.Join(tempDir, "temp_audio")
	if _, err := os.Stat(tempAudioDir); !os.IsNotExist(err) {
		t.Error("Non-existent directory should not cause errors")
	}
}

func TestFormatDuration_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		hUnit    string
		mUnit    string
		sUnit    string
		expected string
	}{
		{
			name:     "microseconds",
			duration: 500 * time.Microsecond,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "0s",
		},
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "0s",
		},
		{
			name:     "nanoseconds",
			duration: 500 * time.Nanosecond,
			hUnit:    "h",
			mUnit:    "m",
			sUnit:    "s",
			expected: "0s",
		},
		{
			name:     "empty units",
			duration: 2*time.Hour + 30*time.Minute + 45*time.Second,
			hUnit:    "",
			mUnit:    "",
			sUnit:    "",
			expected: "2 30 45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration, tt.hUnit, tt.mUnit, tt.sUnit)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerateRandomOdd_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{
			name: "same min and max odd",
			min:  5,
			max:  5,
		},
		{
			name: "same min and max even",
			min:  4,
			max:  4,
		},
		{
			name: "zero range",
			min:  0,
			max:  0,
		},
		{
			name: "negative same",
			min:  -5,
			max:  -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 20; i++ {
				result := GenerateRandomOdd(tt.min, tt.max)

				if tt.min == tt.max && tt.min%2 == 0 {
					if result != tt.min {
						t.Errorf("Expected %d for even same min/max, got %d", tt.min, result)
					}
				} else {
					if result%2 != 1 {
						t.Errorf("Result %d is not odd", result)
					}
				}
			}
		})
	}
}
