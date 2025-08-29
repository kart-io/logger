package core

import "context"

// Logger defines the standard logging interface used throughout the application.
// It provides structured logging capabilities with context support and multiple output formats.
type Logger interface {
	// Basic logging methods with variadic arguments
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	// Printf-style logging methods with format templates
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	// Structured logging methods with key-value pairs
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	// Logger enhancement methods
	With(keyValues ...interface{}) Logger
	WithCtx(ctx context.Context, keyValues ...interface{}) Logger
	WithCallerSkip(skip int) Logger

	// Configuration methods
	SetLevel(level Level)
}