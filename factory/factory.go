package factory

import (
	"fmt"

	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/engines/slog"
	"github.com/kart-io/logger/engines/zap"
	"github.com/kart-io/logger/option"
)

// LoggerFactory creates logger instances based on configuration.
type LoggerFactory struct {
	option *option.LogOption
}

// NewLoggerFactory creates a new logger factory with the provided configuration.
func NewLoggerFactory(opt *option.LogOption) *LoggerFactory {
	return &LoggerFactory{
		option: opt,
	}
}

// CreateLogger creates a logger instance based on the configured engine.
func (f *LoggerFactory) CreateLogger() (core.Logger, error) {
	if err := f.option.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Validate engine before attempting creation
	if f.option.Engine != "zap" && f.option.Engine != "slog" {
		return nil, fmt.Errorf("unsupported logger engine: %s", f.option.Engine)
	}

	// Try engines in fallback order: requested -> alternative -> error
	switch f.option.Engine {
	case "zap":
		if logger, err := f.createZapLogger(); err == nil {
			return logger, nil
		}
		// Fallback to slog
		return f.createSlogLogger()
	case "slog":
		if logger, err := f.createSlogLogger(); err == nil {
			return logger, nil
		}
		// Fallback to zap
		return f.createZapLogger()
	default:
		// This should never be reached due to validation above
		return nil, fmt.Errorf("unsupported logger engine: %s", f.option.Engine)
	}
}

// createZapLogger creates a Zap-based logger implementation.
func (f *LoggerFactory) createZapLogger() (core.Logger, error) {
	return zap.NewZapLogger(f.option)
}

// createSlogLogger creates a Slog-based logger implementation.
func (f *LoggerFactory) createSlogLogger() (core.Logger, error) {
	return slog.NewSlogLogger(f.option)
}

// GetOption returns the current configuration.
func (f *LoggerFactory) GetOption() *option.LogOption {
	return f.option
}

// UpdateOption updates the factory configuration and can be used for dynamic reconfiguration.
func (f *LoggerFactory) UpdateOption(opt *option.LogOption) error {
	if err := opt.Validate(); err != nil {
		return fmt.Errorf("invalid configuration update: %w", err)
	}
	f.option = opt
	return nil
}