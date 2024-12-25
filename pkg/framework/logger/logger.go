package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.SugaredLogger

// Config represents logger configuration
type Config struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Filename   string `mapstructure:"filename"`    // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 每个日志文件的最大大小（MB）
	MaxBackups int    `mapstructure:"max_backups"` // 保留的旧日志文件最大数量
	MaxAge     int    `mapstructure:"max_age"`     // 保留的旧日志文件的最大天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩旧日志文件
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Filename:   "thctl.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
}

// Init initializes the global logger
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// 配置日志输出
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writer,
		level,
	)

	logger := zap.New(core, zap.AddCaller())
	globalLogger = logger.Sugar()
	return nil
}

// WithModule returns a logger with module field
func WithModule(module string) *zap.SugaredLogger {
	return globalLogger.With("module", module)
}

// Debug logs a message at debug level
func Debug(args ...interface{}) {
	globalLogger.Debug(args...)
}

// Info logs a message at info level
func Info(args ...interface{}) {
	globalLogger.Info(args...)
}

// Warn logs a message at warn level
func Warn(args ...interface{}) {
	globalLogger.Warn(args...)
}

// Error logs a message at error level
func Error(args ...interface{}) {
	globalLogger.Error(args...)
}

// Debugf logs a formatted message at debug level
func Debugf(template string, args ...interface{}) {
	globalLogger.Debugf(template, args...)
}

// Infof logs a formatted message at info level
func Infof(template string, args ...interface{}) {
	globalLogger.Infof(template, args...)
}

// Warnf logs a formatted message at warn level
func Warnf(template string, args ...interface{}) {
	globalLogger.Warnf(template, args...)
}

// Errorf logs a formatted message at error level
func Errorf(template string, args ...interface{}) {
	globalLogger.Errorf(template, args...)
}
