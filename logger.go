package logger

import (
	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/factory"
	"github.com/kart-io/logger/option"
)

// Global logger instance
var global core.Logger

// New creates a new logger with the provided configuration.
func New(opt *option.LogOption) (core.Logger, error) {
	f := factory.NewLoggerFactory(opt)
	return f.CreateLogger()
}

// NewWithDefaults creates a new logger with default configuration.
func NewWithDefaults() (core.Logger, error) {
	return New(option.DefaultLogOption())
}

// SetGlobal sets the global logger instance.
func SetGlobal(logger core.Logger) {
	global = logger
}

// Global returns the global logger instance.
// If no global logger is set, it returns a logger with default configuration.
func Global() core.Logger {
	if global == nil {
		// Fallback to default logger if none is set
		logger, err := NewWithDefaults()
		if err != nil {
			// This should not happen with valid default config
			panic("failed to create default logger: " + err.Error())
		}
		global = logger
	}
	return global
}

// Package-level convenience functions using the global logger

// Debug logs a debug message using the global logger.
func Debug(args ...interface{}) {
	Global().Debug(args...)
}

// Info logs an info message using the global logger.
func Info(args ...interface{}) {
	Global().Info(args...)
}

// Warn logs a warning message using the global logger.
func Warn(args ...interface{}) {
	Global().Warn(args...)
}

// Error logs an error message using the global logger.
func Error(args ...interface{}) {
	Global().Error(args...)
}

// Fatal logs a fatal message using the global logger.
func Fatal(args ...interface{}) {
	Global().Fatal(args...)
}

// Debugf logs a debug message with formatting using the global logger.
func Debugf(template string, args ...interface{}) {
	Global().Debugf(template, args...)
}

// Infof logs an info message with formatting using the global logger.
func Infof(template string, args ...interface{}) {
	Global().Infof(template, args...)
}

// Warnf logs a warning message with formatting using the global logger.
func Warnf(template string, args ...interface{}) {
	Global().Warnf(template, args...)
}

// Errorf logs an error message with formatting using the global logger.
func Errorf(template string, args ...interface{}) {
	Global().Errorf(template, args...)
}

// Fatalf logs a fatal message with formatting using the global logger.
func Fatalf(template string, args ...interface{}) {
	Global().Fatalf(template, args...)
}

// Debugw logs a debug message with structured fields using the global logger.
func Debugw(msg string, keysAndValues ...interface{}) {
	Global().Debugw(msg, keysAndValues...)
}

// Infow logs an info message with structured fields using the global logger.
func Infow(msg string, keysAndValues ...interface{}) {
	Global().Infow(msg, keysAndValues...)
}

// Warnw logs a warning message with structured fields using the global logger.
func Warnw(msg string, keysAndValues ...interface{}) {
	Global().Warnw(msg, keysAndValues...)
}

// Errorw logs an error message with structured fields using the global logger.
func Errorw(msg string, keysAndValues ...interface{}) {
	Global().Errorw(msg, keysAndValues...)
}

// Fatalw logs a fatal message with structured fields using the global logger.
func Fatalw(msg string, keysAndValues ...interface{}) {
	Global().Fatalw(msg, keysAndValues...)
}

// With creates a child logger with the specified key-value pairs using the global logger.
func With(keysAndValues ...interface{}) core.Logger {
	return Global().With(keysAndValues...)
}