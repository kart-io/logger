package zap

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/fields"
	"github.com/kart-io/logger/option"
	"github.com/kart-io/logger/otlp"
)

// ZapLogger implements the core.Logger interface using Uber's Zap library.
type ZapLogger struct {
	logger       *zap.Logger
	sugar        *zap.SugaredLogger
	level        core.Level
	mapper       *fields.FieldMapper
	callerSkip   int
	otlpProvider *otlp.LoggerProvider
}

// NewZapLogger creates a new Zap-based logger with the provided configuration.
func NewZapLogger(opt *option.LogOption) (core.Logger, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	// Parse the log level
	level, err := core.ParseLevel(opt.Level)
	if err != nil {
		return nil, err
	}

	// Initialize OTLP provider if enabled
	var otlpProvider *otlp.LoggerProvider
	if opt.IsOTLPEnabled() {
		provider, err := otlp.NewLoggerProvider(context.Background(), opt.OTLP)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP provider: %w", err)
		}
		otlpProvider = provider
	}

	// Create Zap config
	config := createZapConfig(opt, level)

	// Create Zap logger
	zapLogger, err := config.Build(
		zap.AddCallerSkip(1), // Base skip for our wrapper methods
	)
	if err != nil {
		return nil, err
	}

	// Create standardized field mapper wrapper
	standardizedLogger := newStandardizedZapLogger(zapLogger, fields.NewFieldMapper())

	// Add engine identifier as a persistent field
	standardizedLogger = standardizedLogger.With(zap.String("engine", "zap"))

	return &ZapLogger{
		logger:       standardizedLogger,
		sugar:        standardizedLogger.Sugar(),
		level:        level,
		mapper:       fields.NewFieldMapper(),
		callerSkip:   0,
		otlpProvider: otlpProvider,
	}, nil
}

// Debug logs a debug message.
func (l *ZapLogger) Debug(args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Debug(args...)
}

// Info logs an info message.
func (l *ZapLogger) Info(args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Info(args...)
}

// Warn logs a warning message.
func (l *ZapLogger) Warn(args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Warn(args...)
}

// Error logs an error message.
func (l *ZapLogger) Error(args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Error(args...)
}

// Fatal logs a fatal message and exits.
func (l *ZapLogger) Fatal(args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Fatal(args...)
}

// Debugf logs a formatted debug message.
func (l *ZapLogger) Debugf(template string, args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Debugf(template, args...)
}

// Infof logs a formatted info message.
func (l *ZapLogger) Infof(template string, args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Infof(template, args...)
}

// Warnf logs a formatted warning message.
func (l *ZapLogger) Warnf(template string, args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Warnf(template, args...)
}

// Errorf logs a formatted error message.
func (l *ZapLogger) Errorf(template string, args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Errorf(template, args...)
}

// Fatalf logs a formatted fatal message and exits.
func (l *ZapLogger) Fatalf(template string, args ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Fatalf(template, args...)
}

// Debugw logs a debug message with structured fields.
func (l *ZapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Debugw(msg, logger.standardizeFields(keysAndValues...)...)
	l.sendToOTLP(core.DebugLevel, msg, keysAndValues...)
}

// Infow logs an info message with structured fields.
func (l *ZapLogger) Infow(msg string, keysAndValues ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Infow(msg, logger.standardizeFields(keysAndValues...)...)
	l.sendToOTLP(core.InfoLevel, msg, keysAndValues...)
}

// Warnw logs a warning message with structured fields.
func (l *ZapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Warnw(msg, logger.standardizeFields(keysAndValues...)...)
	l.sendToOTLP(core.WarnLevel, msg, keysAndValues...)
}

// Errorw logs an error message with structured fields.
func (l *ZapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Errorw(msg, logger.standardizeFields(keysAndValues...)...)
	l.sendToOTLP(core.ErrorLevel, msg, keysAndValues...)
}

// Fatalw logs a fatal message with structured fields and exits.
func (l *ZapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	logger := l.withDynamicCallerSkip().(*ZapLogger)
	logger.sugar.Fatalw(msg, logger.standardizeFields(keysAndValues...)...)
	l.sendToOTLP(core.FatalLevel, msg, keysAndValues...)
}

// With creates a child logger with the specified key-value pairs.
func (l *ZapLogger) With(keysAndValues ...interface{}) core.Logger {
	standardizedFields := l.standardizeFields(keysAndValues...)
	newSugar := l.sugar.With(standardizedFields...)
	
	return &ZapLogger{
		logger:       newSugar.Desugar(),
		sugar:        newSugar,
		level:        l.level,
		mapper:       l.mapper,
		callerSkip:   l.callerSkip,
		otlpProvider: l.otlpProvider, // Preserve OTLP provider
	}
}

// WithCtx creates a child logger with context and key-value pairs.
func (l *ZapLogger) WithCtx(ctx context.Context, keysAndValues ...interface{}) core.Logger {
	// Zap doesn't have direct context support, so we'll just add the fields
	return l.With(keysAndValues...)
}

// WithCallerSkip creates a child logger that skips additional stack frames.
func (l *ZapLogger) WithCallerSkip(skip int) core.Logger {
	newLogger := l.logger.WithOptions(zap.AddCallerSkip(skip))
	
	return &ZapLogger{
		logger:       newLogger,
		sugar:        newLogger.Sugar(),
		level:        l.level,
		mapper:       l.mapper,
		callerSkip:   l.callerSkip + skip,
		otlpProvider: l.otlpProvider, // Preserve OTLP provider
	}
}

// withDynamicCallerSkip creates a logger with caller skip based on call stack
func (l *ZapLogger) withDynamicCallerSkip() core.Logger {
	// Check if this is a call through global logger function
	var pcs [10]uintptr
	n := runtime.Callers(1, pcs[:])
	hasGlobalCall := false
	
	if n > 0 {
		fs := runtime.CallersFrames(pcs[:n])
		for i := 0; i < n; i++ {
			if f, more := fs.Next(); more || i == n-1 {
				if strings.Contains(f.File, "github.com/kart-io/logger/logger.go") {
					hasGlobalCall = true
					break
				}
			}
		}
	}
	
	// Add extra skip for global calls
	extraSkip := 0
	if hasGlobalCall {
		extraSkip = 1
	}
	
	if extraSkip > 0 {
		return l.WithCallerSkip(extraSkip)
	}
	
	return l
}

// SetLevel sets the minimum logging level.
func (l *ZapLogger) SetLevel(level core.Level) {
	l.level = level
	// Note: Zap doesn't support dynamic level changes easily
	// This would require creating a new logger with different config
}

// Helper functions

func (l *ZapLogger) standardizeFields(keysAndValues ...interface{}) []interface{} {
	standardized := make([]interface{}, 0, len(keysAndValues))
	
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			// Odd number of arguments, use empty value for last key
			key := l.getStandardFieldName(anyToString(keysAndValues[i]))
			standardized = append(standardized, key, nil)
			break
		}
		
		key := anyToString(keysAndValues[i])
		value := keysAndValues[i+1]
		
		// Apply field mapping for consistency
		standardKey := l.getStandardFieldName(key)
		standardized = append(standardized, standardKey, value)
	}
	
	return standardized
}

func (l *ZapLogger) getStandardFieldName(fieldName string) string {
	coreMapping := l.mapper.MapCoreFields()
	if mapped, exists := coreMapping[fieldName]; exists {
		return mapped
	}
	
	tracingMapping := l.mapper.MapTracingFields()
	if mapped, exists := tracingMapping[fieldName]; exists {
		return mapped
	}
	
	return fieldName // Return original if no mapping found
}

func anyToString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	if s, ok := v.(string); ok {
		return s
	}
	// Use fmt.Sprintf for simple string conversion
	return fmt.Sprintf("%v", v)
}

func createZapConfig(opt *option.LogOption, level core.Level) zap.Config {
	// Start with appropriate preset
	var config zap.Config
	if opt.Development {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// Override with our specific settings
	config.Level = zap.NewAtomicLevelAt(mapToZapLevel(level))
	config.DisableCaller = opt.DisableCaller
	config.DisableStacktrace = opt.DisableStacktrace

	// Set encoding format
	switch strings.ToLower(opt.Format) {
	case "console", "text":
		config.Encoding = "console"
	case "json":
		config.Encoding = "json"
	default:
		config.Encoding = "json"
	}

	// Configure output paths
	if len(opt.OutputPaths) > 0 {
		config.OutputPaths = normalizeOutputPaths(opt.OutputPaths)
		config.ErrorOutputPaths = normalizeOutputPaths(opt.OutputPaths) // Use same for errors
	}

	// Configure encoder with standardized field names
	config.EncoderConfig = createStandardizedEncoderConfig()

	return config
}

func createStandardizedEncoderConfig() zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()
	
	// Use our standardized field names
	config.TimeKey = fields.TimestampField
	config.LevelKey = fields.LevelField
	config.MessageKey = fields.MessageField
	config.CallerKey = fields.CallerField
	config.StacktraceKey = fields.StacktraceField
	
	// Configure time format
	config.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder
	
	return config
}

func mapToZapLevel(level core.Level) zapcore.Level {
	switch level {
	case core.DebugLevel:
		return zapcore.DebugLevel
	case core.InfoLevel:
		return zapcore.InfoLevel
	case core.WarnLevel:
		return zapcore.WarnLevel
	case core.ErrorLevel:
		return zapcore.ErrorLevel
	case core.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func normalizeOutputPaths(paths []string) []string {
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		switch strings.ToLower(path) {
		case "stdout", "":
			normalized = append(normalized, "stdout")
		case "stderr":
			normalized = append(normalized, "stderr")
		default:
			normalized = append(normalized, path)
		}
	}
	return normalized
}

// standardizedZapLogger wraps zap.Logger to ensure field standardization
type standardizedZapLogger struct {
	*zap.Logger
	mapper *fields.FieldMapper
}

func newStandardizedZapLogger(logger *zap.Logger, mapper *fields.FieldMapper) *zap.Logger {
	// For now, return the original logger as Zap's field standardization
	// is handled through the EncoderConfig and our With() method
	// The real standardization happens in the ZapLogger.standardizeFields method
	_ = mapper // Silence unused warning
	return logger
}

// sendToOTLP sends log data to OTLP as a log record.
func (l *ZapLogger) sendToOTLP(level core.Level, msg string, keysAndValues ...interface{}) {
	if l.otlpProvider == nil {
		return
	}

	// Convert keysAndValues to map
	attributes := make(map[string]interface{})
	
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			break
		}
		
		key := anyToString(keysAndValues[i])
		value := keysAndValues[i+1]
		
		// Apply field mapping
		standardKey := l.getStandardFieldName(key)
		attributes[standardKey] = value
	}

	// Send log record to OTLP
	if err := l.otlpProvider.SendLogRecord(level, msg, attributes); err != nil {
		// Log the error to stderr without causing recursion
		fmt.Printf("OTLP export error: %v\n", err)
	}
}