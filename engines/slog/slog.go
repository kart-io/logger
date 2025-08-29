package slog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/fields"
	"github.com/kart-io/logger/option"
	"github.com/kart-io/logger/otlp"
)

// SlogLogger implements the core.Logger interface using Go's standard slog library.
type SlogLogger struct {
	logger            *slog.Logger
	level             core.Level
	mapper            *fields.FieldMapper
	callerSkip        int
	disableStacktrace bool
	otlpProvider      *otlp.LoggerProvider
}

// NewSlogLogger creates a new Slog-based logger with the provided configuration.
func NewSlogLogger(opt *option.LogOption) (core.Logger, error) {
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

	// Create output writers
	writers, err := createOutputWriters(opt.OutputPaths)
	if err != nil {
		return nil, err
	}

	// Create handler options - we handle caller manually for consistent formatting
	handlerOpts := &slog.HandlerOptions{
		Level:     mapToSlogLevel(level),
		AddSource: false, // We'll add standardized caller field ourselves
	}

	// Create handler based on format
	var handler slog.Handler
	switch strings.ToLower(opt.Format) {
	case "json":
		handler = slog.NewJSONHandler(writers, handlerOpts)
	case "console", "text":
		handler = slog.NewTextHandler(writers, handlerOpts)
	default:
		handler = slog.NewJSONHandler(writers, handlerOpts)
	}

	// Create standardized handler wrapper for field consistency
	standardHandler := &standardizedHandler{
		handler:           handler,
		mapper:            fields.NewFieldMapper(),
		disableCaller:     opt.DisableCaller,
		disableStacktrace: opt.DisableStacktrace,
	}

	logger := slog.New(standardHandler)

	return &SlogLogger{
		logger:            logger,
		level:             level,
		mapper:            fields.NewFieldMapper(),
		callerSkip:        0,
		disableStacktrace: opt.DisableStacktrace,
		otlpProvider:      otlpProvider,
	}, nil
}

// Debug logs a debug message.
func (l *SlogLogger) Debug(args ...interface{}) {
	if caller := l.getCaller(); caller != "" {
		l.logger.Debug(formatArgs(args...), slog.String(fields.CallerField, caller))
	} else {
		l.logger.Debug(formatArgs(args...))
	}
}

// Info logs an info message.
func (l *SlogLogger) Info(args ...interface{}) {
	if caller := l.getCaller(); caller != "" {
		l.logger.Info(formatArgs(args...), slog.String(fields.CallerField, caller))
	} else {
		l.logger.Info(formatArgs(args...))
	}
}

// Warn logs a warning message.
func (l *SlogLogger) Warn(args ...interface{}) {
	if caller := l.getCaller(); caller != "" {
		l.logger.Warn(formatArgs(args...), slog.String(fields.CallerField, caller))
	} else {
		l.logger.Warn(formatArgs(args...))
	}
}

// Error logs an error message.
func (l *SlogLogger) Error(args ...interface{}) {
	attrs := []any{}
	
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	
	// Add stacktrace for error level
	if stacktrace := l.getStacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String(fields.StacktraceField, stacktrace))
	}
	
	l.logger.Error(formatArgs(args...), attrs...)
}

// Fatal logs a fatal message and exits.
func (l *SlogLogger) Fatal(args ...interface{}) {
	attrs := []any{}
	
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	
	// Add stacktrace for fatal level
	if stacktrace := l.getStacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String(fields.StacktraceField, stacktrace))
	}
	
	l.logger.Error(formatArgs(args...), attrs...)
	os.Exit(1)
}

// Debugf logs a formatted debug message.
func (l *SlogLogger) Debugf(template string, args ...interface{}) {
	if caller := l.getCaller(); caller != "" {
		l.logger.Debug(fmt.Sprintf(template, args...), slog.String(fields.CallerField, caller))
	} else {
		l.logger.Debug(fmt.Sprintf(template, args...))
	}
}

// Infof logs a formatted info message.
func (l *SlogLogger) Infof(template string, args ...interface{}) {
	if caller := l.getCaller(); caller != "" {
		l.logger.Info(fmt.Sprintf(template, args...), slog.String(fields.CallerField, caller))
	} else {
		l.logger.Info(fmt.Sprintf(template, args...))
	}
}

// Warnf logs a formatted warning message.
func (l *SlogLogger) Warnf(template string, args ...interface{}) {
	if caller := l.getCaller(); caller != "" {
		l.logger.Warn(fmt.Sprintf(template, args...), slog.String(fields.CallerField, caller))
	} else {
		l.logger.Warn(fmt.Sprintf(template, args...))
	}
}

// Errorf logs a formatted error message.
func (l *SlogLogger) Errorf(template string, args ...interface{}) {
	attrs := []any{}
	
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	
	// Add stacktrace for error level
	if stacktrace := l.getStacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String(fields.StacktraceField, stacktrace))
	}
	
	l.logger.Error(fmt.Sprintf(template, args...), attrs...)
}

// Fatalf logs a formatted fatal message and exits.
func (l *SlogLogger) Fatalf(template string, args ...interface{}) {
	attrs := []any{}
	
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	
	// Add stacktrace for fatal level
	if stacktrace := l.getStacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String(fields.StacktraceField, stacktrace))
	}
	
	l.logger.Error(fmt.Sprintf(template, args...), attrs...)
	os.Exit(1)
}

// Debugw logs a debug message with structured fields.
func (l *SlogLogger) Debugw(msg string, keysAndValues ...interface{}) {
	attrs := l.convertToSlogAttrs(keysAndValues...)
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	l.logger.DebugContext(context.Background(), msg, attrs...)
	l.sendToOTLP(core.DebugLevel, msg, keysAndValues...)
}

// Infow logs an info message with structured fields.
func (l *SlogLogger) Infow(msg string, keysAndValues ...interface{}) {
	attrs := l.convertToSlogAttrs(keysAndValues...)
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	l.logger.InfoContext(context.Background(), msg, attrs...)
	l.sendToOTLP(core.InfoLevel, msg, keysAndValues...)
}

// Warnw logs a warning message with structured fields.
func (l *SlogLogger) Warnw(msg string, keysAndValues ...interface{}) {
	attrs := l.convertToSlogAttrs(keysAndValues...)
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	l.logger.WarnContext(context.Background(), msg, attrs...)
	l.sendToOTLP(core.WarnLevel, msg, keysAndValues...)
}

// Errorw logs an error message with structured fields.
func (l *SlogLogger) Errorw(msg string, keysAndValues ...interface{}) {
	attrs := l.convertToSlogAttrs(keysAndValues...)
	
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	
	// Add stacktrace for error level
	if stacktrace := l.getStacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String(fields.StacktraceField, stacktrace))
	}
	
	l.logger.ErrorContext(context.Background(), msg, attrs...)
	l.sendToOTLP(core.ErrorLevel, msg, keysAndValues...)
}

// Fatalw logs a fatal message with structured fields and exits.
func (l *SlogLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	attrs := l.convertToSlogAttrs(keysAndValues...)
	
	if caller := l.getCaller(); caller != "" {
		attrs = append(attrs, slog.String(fields.CallerField, caller))
	}
	
	// Add stacktrace for fatal level
	if stacktrace := l.getStacktrace(); stacktrace != "" {
		attrs = append(attrs, slog.String(fields.StacktraceField, stacktrace))
	}
	
	l.logger.ErrorContext(context.Background(), msg, attrs...)
	l.sendToOTLP(core.FatalLevel, msg, keysAndValues...)
	os.Exit(1)
}

// With creates a child logger with the specified key-value pairs.
func (l *SlogLogger) With(keysAndValues ...interface{}) core.Logger {
	newLogger := l.logger.With(l.convertToSlogAttrs(keysAndValues...)...)
	return &SlogLogger{
		logger:            newLogger,
		level:             l.level,
		mapper:            l.mapper,
		callerSkip:        l.callerSkip,
		disableStacktrace: l.disableStacktrace,
		otlpProvider:      l.otlpProvider,
	}
}

// WithCtx creates a child logger with context and key-value pairs.
func (l *SlogLogger) WithCtx(ctx context.Context, keysAndValues ...interface{}) core.Logger {
	// Slog doesn't have a direct equivalent, so we'll create a logger with the fields
	newLogger := l.logger.With(l.convertToSlogAttrs(keysAndValues...)...)
	return &SlogLogger{
		logger:            newLogger,
		level:             l.level,
		mapper:            l.mapper,
		callerSkip:        l.callerSkip,
		disableStacktrace: l.disableStacktrace,
		otlpProvider:      l.otlpProvider,
	}
}

// WithCallerSkip creates a child logger that skips additional stack frames.
func (l *SlogLogger) WithCallerSkip(skip int) core.Logger {
	return &SlogLogger{
		logger:            l.logger,
		level:             l.level,
		mapper:            l.mapper,
		callerSkip:        l.callerSkip + skip,
		disableStacktrace: l.disableStacktrace,
		otlpProvider:      l.otlpProvider,
	}
}

// SetLevel sets the minimum logging level.
func (l *SlogLogger) SetLevel(level core.Level) {
	l.level = level
	// Note: slog doesn't support dynamic level changes easily
	// This would require recreating the handler with new options
}

// Helper functions

func formatArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}
	if len(args) == 1 {
		return anyToString(args[0])
	}
	
	var parts []string
	for _, arg := range args {
		parts = append(parts, anyToString(arg))
	}
	return strings.Join(parts, " ")
}

func formatToSlogArgs(args ...interface{}) []interface{} {
	// For printf-style formatting, we don't need to convert to slog.Attr
	return args
}

func anyToString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	if s, ok := v.(string); ok {
		return s
	}
	return slog.AnyValue(v).String()
}

func (l *SlogLogger) convertToSlogAttrs(keysAndValues ...interface{}) []interface{} {
	attrs := make([]interface{}, 0, len(keysAndValues))
	
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			// Odd number of arguments, use empty value for last key
			attrs = append(attrs, slog.Any(anyToString(keysAndValues[i]), nil))
			break
		}
		
		key := anyToString(keysAndValues[i])
		value := keysAndValues[i+1]
		
		// Apply field mapping for consistency
		if mappedKey := l.getStandardFieldName(key); mappedKey != "" {
			key = mappedKey
		}
		
		attrs = append(attrs, slog.Any(key, value))
	}
	
	return attrs
}

func (l *SlogLogger) getStandardFieldName(fieldName string) string {
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

func mapToSlogLevel(level core.Level) slog.Level {
	switch level {
	case core.DebugLevel:
		return slog.LevelDebug
	case core.InfoLevel:
		return slog.LevelInfo
	case core.WarnLevel:
		return slog.LevelWarn
	case core.ErrorLevel:
		return slog.LevelError
	case core.FatalLevel:
		return slog.LevelError // slog doesn't have Fatal level
	default:
		return slog.LevelInfo
	}
}

func createOutputWriters(paths []string) (io.Writer, error) {
	if len(paths) == 0 {
		return os.Stdout, nil
	}
	
	var writers []io.Writer
	for _, path := range paths {
		switch strings.ToLower(path) {
		case "stdout", "":
			writers = append(writers, os.Stdout)
		case "stderr":
			writers = append(writers, os.Stderr)
		default:
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, err
			}
			writers = append(writers, file)
		}
	}
	
	if len(writers) == 1 {
		return writers[0], nil
	}
	
	return io.MultiWriter(writers...), nil
}

// standardizedHandler wraps slog.Handler to ensure field standardization
type standardizedHandler struct {
	handler            slog.Handler
	mapper             *fields.FieldMapper
	disableCaller      bool
	disableStacktrace  bool
}

func (h *standardizedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *standardizedHandler) Handle(ctx context.Context, record slog.Record) error {
	// Create a new record with standardized field names
	newRecord := slog.Record{
		Time:    record.Time,
		Level:   record.Level,
		Message: record.Message,
		PC:      record.PC,
	}
	
	// Add standardized engine identifier
	newRecord.AddAttrs(slog.Attr{
		Key:   "engine",
		Value: slog.StringValue("slog"),
	})
	
	
	// Map user-defined fields using our field standardization system
	record.Attrs(func(attr slog.Attr) bool {
		standardKey := h.getStandardFieldName(attr.Key)
		newRecord.AddAttrs(slog.Attr{
			Key:   standardKey,
			Value: attr.Value,
		})
		return true
	})
	
	return h.handler.Handle(ctx, newRecord)
}

func (h *standardizedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	standardizedAttrs := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		standardizedAttrs[i] = slog.Attr{
			Key:   h.getStandardFieldName(attr.Key),
			Value: attr.Value,
		}
	}
	return &standardizedHandler{
		handler:           h.handler.WithAttrs(standardizedAttrs),
		mapper:            h.mapper,
		disableCaller:     h.disableCaller,
		disableStacktrace: h.disableStacktrace,
	}
}

func (h *standardizedHandler) WithGroup(name string) slog.Handler {
	return &standardizedHandler{
		handler:           h.handler.WithGroup(name),
		mapper:            h.mapper,
		disableCaller:     h.disableCaller,
		disableStacktrace: h.disableStacktrace,
	}
}


func (h *standardizedHandler) getStandardFieldName(fieldName string) string {
	coreMapping := h.mapper.MapCoreFields()
	if mapped, exists := coreMapping[fieldName]; exists {
		return mapped
	}
	
	tracingMapping := h.mapper.MapTracingFields()
	if mapped, exists := tracingMapping[fieldName]; exists {
		return mapped
	}
	
	return fieldName // Return original if no mapping found
}

// getCaller returns the caller information for the SlogLogger
func (l *SlogLogger) getCaller() string {
	if l == nil {
		return ""
	}
	
	// Check if this is a call through global logger function
	// by looking at the call stack
	var pcs [10]uintptr
	n := runtime.Callers(1, pcs[:])
	if n == 0 {
		return ""
	}
	
	fs := runtime.CallersFrames(pcs[:n])
	hasGlobalCall := false
	
	// Check if there's a global logger function in the call stack
	for i := 0; i < n; i++ {
		if f, more := fs.Next(); more || i == n-1 {
			if strings.Contains(f.File, "github.com/kart-io/logger/logger.go") {
				hasGlobalCall = true
				break
			}
		}
	}
	
	// Determine skip based on call type
	var skip int
	if hasGlobalCall {
		skip = 4 + l.callerSkip // getCaller -> SlogLogger method -> global function -> actual caller
	} else {
		skip = 3 + l.callerSkip // getCaller -> SlogLogger method -> actual caller
	}
	
	var pcs2 [1]uintptr
	if runtime.Callers(skip, pcs2[:]) > 0 {
		fs2 := runtime.CallersFrames(pcs2[:1])
		if f, _ := fs2.Next(); f.File != "" {
			// Extract just the filename from the full path
			file := f.File
			if idx := strings.LastIndex(file, "/"); idx >= 0 {
				if idx2 := strings.LastIndex(file[:idx], "/"); idx2 >= 0 {
					file = file[idx2+1:] // Keep last two path segments
				}
			}
			
			return fmt.Sprintf("%s:%d", file, f.Line)
		}
	}
	
	return ""
}

// getStacktrace returns the stack trace for error/fatal level logs
func (l *SlogLogger) getStacktrace() string {
	if l == nil || l.disableStacktrace {
		return ""
	}
	
	// Skip frames: getStacktrace -> SlogLogger method -> actual caller
	const baseSkip = 3
	skip := baseSkip + l.callerSkip
	
	var pcs [10]uintptr
	n := runtime.Callers(skip, pcs[:])
	if n == 0 {
		return ""
	}
	
	fs := runtime.CallersFrames(pcs[:n])
	var stackTrace strings.Builder
	
	for {
		f, more := fs.Next()
		
		// Extract function name and location
		funcName := f.Function
		if idx := strings.LastIndex(funcName, "/"); idx >= 0 {
			funcName = funcName[idx+1:]
		}
		
		file := f.File
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			if idx2 := strings.LastIndex(file[:idx], "/"); idx2 >= 0 {
				file = file[idx2+1:] // Keep last two path segments
			}
		}
		
		if stackTrace.Len() > 0 {
			stackTrace.WriteString("\\n")
		}
		stackTrace.WriteString(fmt.Sprintf("%s\\n\\t%s:%d", funcName, file, f.Line))
		
		if !more {
			break
		}
	}
	
	return stackTrace.String()
}

// sendToOTLP sends log data to OTLP as a log record.
func (l *SlogLogger) sendToOTLP(level core.Level, msg string, keysAndValues ...interface{}) {
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