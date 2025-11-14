package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger wraps zap.Logger to provide application-specific logging
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new logger instance with the specified log level
// Supported levels: debug, info, warn, error
func NewLogger(level string) (*Logger, error) {
	var zapLevel zap.AtomicLevel
	
	switch level {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		return nil, fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", level)
	}

	// Configure structured logging format
	config := zap.Config{
		Level:            zapLevel,
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// Build the logger
	zapLogger, err := config.Build(
		zap.AddCallerSkip(1), // Skip one level to show correct caller
		zap.AddStacktrace(zap.ErrorLevel), // Add stack trace for errors
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return &Logger{Logger: zapLogger}, nil
}

// NewDevelopmentLogger creates a logger optimized for development
// with console encoding and debug level
func NewDevelopmentLogger() (*Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	
	zapLogger, err := config.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize development logger: %w", err)
	}

	return &Logger{Logger: zapLogger}, nil
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.With(fields...)}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
