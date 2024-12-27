package config

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/joho/godotenv"
)

// Config represents the configuration structure
type Config struct {
    Lotus struct {
        APIURL    string        `yaml:"api_url"`
        AuthToken string        `yaml:"auth_token"`
        Timeout   time.Duration `yaml:"timeout"`
    } `yaml:"lotus"`
    THCloud struct {
        APIKey string `yaml:"api_key"`
    } `yaml:"thcloud"`
}

var (
    configDir  string
    config     *Config
    configOnce sync.Once
)

func init() {
    if dir := os.Getenv("THCTL_CONFIG_DIR"); dir != "" {
        configDir = dir
    } else {
        // Default to user's home directory
        home, err := os.UserHomeDir()
        if err == nil {
            configDir = filepath.Join(home, ".thctl")
        }
    }
}

// GetConfigDir returns the configuration directory
func GetConfigDir() string {
    return configDir
}

// SetConfigDir sets the configuration directory
func SetConfigDir(dir string) {
    configDir = dir
}

// Load loads the configuration from environment variables and .thctl.env file
func Load() (cfg *Config, err error) {
    configOnce.Do(func() {
        config = &Config{}

        // Try to load from .thctl.env file in the current directory
        fmt.Printf("Trying to load .thctl.env from current directory...\n")
        err = godotenv.Load(".thctl.env")
        if err != nil && !os.IsNotExist(err) {
            err = fmt.Errorf("failed to load .thctl.env file: %v", err)
            return
        }

        // If not found in current directory, try home directory
        if os.IsNotExist(err) {
            home, homeErr := os.UserHomeDir()
            if homeErr == nil {
                envFile := filepath.Join(home, ".thctl.env")
                fmt.Printf("Trying to load .thctl.env from home directory: %s\n", envFile)
                err = godotenv.Load(envFile)
                if err != nil && !os.IsNotExist(err) {
                    err = fmt.Errorf("failed to load home directory .thctl.env file: %v", err)
                    return
                }
            }
        }

        // Set values from environment variables
        config.Lotus.APIURL = getEnvWithDefault("LOTUS_API_URL", "/ip4/127.0.0.1/tcp/1234")
        config.Lotus.AuthToken = getEnvWithDefault("LOTUS_API_TOKEN", "")
        config.Lotus.Timeout = getDurationEnvWithDefault("LOTUS_API_TIMEOUT", 30*time.Second)
        config.THCloud.APIKey = getEnvWithDefault("THCLOUD_API_KEY", "")

        fmt.Printf("Loaded config: LOTUS_API_URL=%s\n", config.Lotus.APIURL)

        // Clear error if we successfully loaded the config
        err = nil
    })

    if err != nil {
        return nil, err
    }

    return config, nil
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

// getDurationEnvWithDefault returns the duration value of an environment variable or a default value
func getDurationEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    duration, err := time.ParseDuration(value)
    if err != nil {
        return defaultValue
    }
    return duration
}
