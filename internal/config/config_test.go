package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Define the expected config filename
	configFilename := getConfigFilename()
	configPath := filepath.Join(tempDir, configFilename)

	// Create a sample TOML config file
	sampleConfig := `
[application]
version = "0.0.1"
name = "example-service"

[server]
addr = ":8080"
timeout = 10 

[logging]
level = "info"

[cors]
allowed_origins = ["https://example.com", "http://localhost"]
allowed_methods = ["GET", "POST", "OPTIONS"]
allow_credentials = true
`
	err = os.WriteFile(configPath, []byte(sampleConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write sample config file: %v", err)
	}

	// Call LoadConfig
	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig returned an unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("LoadConfig returned nil config without an error")
	}

	// Validate ApplicationConfig
	if config.Application.Version != "0.0.1" {
		t.Errorf("Expected Application.Version '0.0.1', got '%s'", config.Application.Version)
	}
	if config.Application.Name != "example-service" {
		t.Errorf("Expected Application.Name 'example-service', got '%s'", config.Application.Name)
	}

	// Validate ServerConfig
	if config.Server.Addr != ":8080" {
		t.Errorf("Expected Server.Addr ':8080', got '%s'", config.Server.Addr)
	}
	if config.Server.Timeout != 10 {
		t.Errorf("Expected Server.Timeout 10, got %d", config.Server.Timeout)
	}

	// Validate LoggingConfig
	if config.Logging.Level != "info" {
		t.Errorf("Expected Logging.Level 'info', got '%s'", config.Logging.Level)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Create a temporary directory without any config files
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Call LoadConfig, expecting an error
	config, err := LoadConfig(tempDir)
	if err == nil {
		t.Fatal("Expected an error when config file does not exist, but got none")
	}
	if config != nil {
		t.Errorf("Expected config to be nil when file does not exist, but got: %+v", config)
	}

	// Check if the error is about file not found
	if !errors.Is(err, os.ErrNotExist) && !os.IsNotExist(err) {
		t.Errorf("Expected file not exist error, but got: %v", err)
	}
}

func TestLoadConfig_MalformedTOML(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Define the expected config filename
	configFilename := "malformed.toml"
	configPath := filepath.Join(tempDir, configFilename)

	// Create a malformed TOML config file
	malformedConfig := `
[application]
version = "0.0.1
name = "example-service"

[server]
addr = ":8080"
timeout = 10 

[logging]
level = "info"
`
	err = os.WriteFile(configPath, []byte(malformedConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write malformed config file: %v", err)
	}

	// Call LoadConfig, expecting a TOML decoding error
	config, err := LoadConfig(tempDir)
	if err == nil {
		t.Fatal("Expected an error when parsing malformed TOML, but got none")
	}
	if config != nil {
		t.Errorf("Expected config to be nil when TOML is malformed, but got: %+v", config)
	}
}

func TestLoadConfig_InvalidPath(t *testing.T) {
	// Use a non-existent directory path
	invalidPath := "/path/does/not/exist"

	// Call LoadConfig, expecting an error
	config, err := LoadConfig(invalidPath)
	if err == nil {
		t.Fatal("Expected an error when providing an invalid path, but got none")
	}
	if config != nil {
		t.Errorf("Expected config to be nil when path is invalid, but got: %+v", config)
	}
}
