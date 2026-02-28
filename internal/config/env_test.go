package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback string
		expected string
		setEnv   bool
	}{
		{
			name:     "Environment variable is set",
			key:      "TEST_KEY",
			value:    "test_value",
			fallback: "fallback_value",
			expected: "test_value",
			setEnv:   true,
		},
		{
			name:     "Environment variable is not set",
			key:      "TEST_KEY",
			fallback: "fallback_value",
			expected: "fallback_value",
			setEnv:   false,
		},
		{
			name:     "Environment variable is empty string",
			key:      "TEST_KEY",
			value:    "",
			fallback: "fallback_value",
			expected: "",
			setEnv:   true,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			result := GetEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("GetEnv(%q, %q) = %q; want %q", tt.key, tt.fallback, result, tt.expected)
			}
		})
	}
}

func TestGetCurrentRuntimeEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "Runtime environment is set to dev",
			envValue: string(Dev),
			expected: string(Dev),
		},
		{
			name:     "Runtime environment is set to staging",
			envValue: string(Staging),
			expected: string(Staging),
		},
		{
			name:     "Runtime environment is set to production",
			envValue: string(Production),
			expected: string(Production),
		},
		{
			name:     "Runtime environment is not set",
			envValue: "",
			expected: string(Dev), // Default fallback
		},
		{
			name:     "Runtime environment is set to an unexpected value",
			envValue: "unexpected",
			expected: "unexpected",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(RuntimeEnvVar, tt.envValue)
			} else {
				os.Unsetenv(RuntimeEnvVar)
			}

			result := GetCurrentRuntimeEnvironment()
			if result != tt.expected {
				t.Errorf("GetCurrentRuntimeEnvironment() = %q; want %q", result, tt.expected)
			}
		})
	}
}
