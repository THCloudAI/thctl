package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	once         sync.Once
)

// Config wraps viper configuration
type Config struct {
	v           *viper.Viper
	configPaths []string
	configName  string
	configType  string
}

// Options defines configuration options
type Options struct {
	ConfigName  string
	ConfigType  string
	ConfigPaths []string
}

// DefaultOptions returns default configuration options
func DefaultOptions() *Options {
	home, _ := os.UserHomeDir()
	return &Options{
		ConfigName: "config",
		ConfigType: "yaml",
		ConfigPaths: []string{
			".",
			filepath.Join(home, ".thctl"),
			"/etc/thctl",
		},
	}
}

// New creates a new Config instance
func New(opts *Options) *Config {
	if opts == nil {
		opts = DefaultOptions()
	}

	v := viper.New()
	v.SetEnvPrefix("THCTL")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &Config{
		v:           v,
		configName:  opts.ConfigName,
		configType:  opts.ConfigType,
		configPaths: opts.ConfigPaths,
	}
}

// Global returns the global Config instance
func Global() *Config {
	once.Do(func() {
		globalConfig = New(nil)
		if err := globalConfig.Load(); err != nil {
			fmt.Printf("Warning: failed to load config: %v\n", err)
		}
	})
	return globalConfig
}

// Load loads the configuration from files and environment
func (c *Config) Load() error {
	c.v.SetConfigName(c.configName)
	c.v.SetConfigType(c.configType)

	for _, path := range c.configPaths {
		c.v.AddConfigPath(path)
	}

	if err := c.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %v", err)
		}
	}

	c.v.WatchConfig()
	c.v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
	})

	return nil
}

// Get retrieves any value from config given its key
func (c *Config) Get(key string) interface{} {
	return c.v.Get(key)
}

// GetString retrieves a string value from config
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt retrieves an integer value from config
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool retrieves a boolean value from config
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetStringMap retrieves a map of strings from config
func (c *Config) GetStringMap(key string) map[string]interface{} {
	return c.v.GetStringMap(key)
}

// UnmarshalKey takes a single key and unmarshals it into a struct
func (c *Config) UnmarshalKey(key string, rawVal interface{}) error {
	return c.v.UnmarshalKey(key, rawVal)
}

// Set sets a value in the configuration
func (c *Config) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

// BindEnv binds a configuration key to an environment variable
func (c *Config) BindEnv(input ...string) error {
	return c.v.BindEnv(input...)
}

// Viper returns the underlying viper instance
func (c *Config) Viper() *viper.Viper {
	return c.v
}
