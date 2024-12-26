package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the configuration
type Config struct {
	LotusAPI struct {
		URL     string
		Token   string
		Timeout time.Duration
	}
}

// LoadFilConfig loads the Filecoin configuration
// Configuration priority (from highest to lowest):
// 1. .thctl.env file in current directory
// 2. .thctl.env file in user's home directory
// 3. Environment variables
// 4. Default values
func LoadFilConfig() *Config {
	cfg := &Config{}
	
	// Set default values first (lowest priority)
	cfg.LotusAPI.URL = "http://127.0.0.1:1234/rpc/v0"
	cfg.LotusAPI.Timeout = 30 * time.Second

	// Try to load from .thctl.env files (highest priority)
	envFiles := []string{
		".thctl.env", // Current directory first
	}

	// Add home directory .thctl.env if available
	if homeDir, err := os.UserHomeDir(); err == nil {
		envFiles = append(envFiles, filepath.Join(homeDir, ".thctl.env"))
	}

	// Load env files in order (first file found takes precedence)
	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			godotenv.Load(envFile)
			break // Stop after first file found
		}
	}

	// Now load from environment variables (lower priority than .env file)
	// Only override if the value wasn't already set by .env file
	if os.Getenv("LOTUS_API_URL") != "" {
		cfg.LotusAPI.URL = os.Getenv("LOTUS_API_URL")
	}
	if os.Getenv("LOTUS_API_TOKEN") != "" {
		cfg.LotusAPI.Token = os.Getenv("LOTUS_API_TOKEN")
	}
	if timeoutStr := os.Getenv("LOTUS_API_TIMEOUT"); timeoutStr != "" {
		if duration, err := time.ParseDuration(timeoutStr); err == nil {
			cfg.LotusAPI.Timeout = duration
		}
	}
	
	return cfg
}

// GetString gets a string value from config
func (c *Config) GetString(key string) string {
	switch key {
	case "lotus.api_url":
		return c.LotusAPI.URL
	case "lotus.token":
		return c.LotusAPI.Token
	default:
		return ""
	}
}

// GetDuration gets a duration value from config
func (c *Config) GetDuration(key string) time.Duration {
	switch key {
	case "lotus.timeout":
		return c.LotusAPI.Timeout
	default:
		return 0
	}
}

// Set sets a value in config
func (c *Config) Set(key string, value interface{}) {
	switch key {
	case "lotus.api_url":
		if v, ok := value.(string); ok {
			c.LotusAPI.URL = v
		}
	case "lotus.token":
		if v, ok := value.(string); ok {
			c.LotusAPI.Token = v
		}
	}
}

// Helper functions

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
