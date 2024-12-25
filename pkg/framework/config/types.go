package config

import "time"

// GlobalConfig represents the global configuration structure
type GlobalConfig struct {
	Environment string         `mapstructure:"env"`
	LogLevel    string         `mapstructure:"log_level"`
	Metrics     MetricsConfig  `mapstructure:"metrics"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

// FilConfig represents Filecoin related configuration
type FilConfig struct {
	Lotus    LotusConfig            `mapstructure:"lotus"`
	Services FilServicesConfig      `mapstructure:"services"`
}

// LotusConfig represents Lotus API configuration
type LotusConfig struct {
	APIURL  string        `mapstructure:"api_url"`
	Token   string        `mapstructure:"token"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// FilServicesConfig represents Filecoin services configuration
type FilServicesConfig struct {
	SectorsPenalty SectorsPenaltyConfig `mapstructure:"sectors_penalty"`
}

// SectorsPenaltyConfig represents sectors penalty service configuration
type SectorsPenaltyConfig struct {
	Port      int           `mapstructure:"port"`
	RateLimit int           `mapstructure:"rate_limit"`
	CacheTTL  time.Duration `mapstructure:"cache_ttl"`
}

// CLIConfig represents CLI specific configuration
type CLIConfig struct {
	DefaultOutput string `mapstructure:"default_output"`
	ColorEnabled  bool   `mapstructure:"color_enabled"`
}
