package config

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all the configuration data for the API server
type Config struct {
	Application ApplicationConfig `toml:"application"`
	Server      ServerConfig      `toml:"server"`
	Logging     LoggingConfig     `toml:"logging"`
	CORS        CORSConfig        `toml:"cors"`
}

// ApplicationConfig holds application-specific settings
type ApplicationConfig struct {
	Version string `toml:"version"`
	Name    string `toml:"name"`
}

// ServerConfig holds the server-related settings
type ServerConfig struct {
	Addr    string `toml:"addr"`
	Timeout uint32 `toml:"timeout"`
}

// LoggingConfig holds the logging-related settings
type LoggingConfig struct {
	Level string `toml:"level"`
}

// CORSConfig holds CORS settings for HTTP handlers
type CORSConfig struct {
	AllowedOrigins   []string `toml:"allowed_origins"`
	AllowedMethods   []string `toml:"allowed_methods"`
	AllowCredentials *bool    `toml:"allow_credentials"`
}

func GetAbsPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

func getConfigFilename() string {
	runtimeEnv := GetCurrentRuntimeEnvironment()
	return fmt.Sprintf("%s.toml", runtimeEnv)
}

// LoadConfig loads the TOML config file into the Config struct
func LoadConfig(path string) (*Config, error) {
	var config Config
	fpath := filepath.Join(path, getConfigFilename())
	if _, err := toml.DecodeFile(fpath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
