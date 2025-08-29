package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/option"
)

func main() {
	fmt.Println("=== Comprehensive Logger Usage Examples ===\n")

	// Example 1: Basic Logging Methods
	demonstrateBasicMethods()

	// Example 2: Printf-style Logging Methods  
	demonstratePrintfMethods()

	// Example 3: Structured Logging Methods
	demonstrateStructuredMethods()

	// Example 4: Logger Enhancement Methods
	demonstrateEnhancementMethods()

	// Example 5: Global Logger Usage
	demonstrateGlobalLogger()

	// Example 6: Configuration Examples
	demonstrateConfiguration()

	// Example 7: Error Handling and Stacktraces
	demonstrateErrorHandling()

	// Example 8: Context and Tracing
	demonstrateContextAndTracing()

	fmt.Println("\n=== All Examples Complete ===")
}

// demonstrateBasicMethods shows basic logging methods: Debug, Info, Warn, Error
func demonstrateBasicMethods() {
	fmt.Println("1. Basic Logging Methods")
	fmt.Println("========================")
	
	// Create a logger instance
	opt := &option.LogOption{
		Engine:      "slog",
		Level:       "DEBUG", // Set to DEBUG to show all levels
		Format:      "json",
		OutputPaths: []string{"stdout"},
		OTLP:        &option.OTLPOption{},
	}

	logger, err := logger.New(opt)
	if err != nil {
		panic(err)
	}

	// Basic logging methods
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
	// Note: Fatal would exit the program, so we skip it in examples
	
	fmt.Println()
}

// demonstratePrintfMethods shows printf-style logging methods
func demonstratePrintfMethods() {
	fmt.Println("2. Printf-style Logging Methods")
	fmt.Println("===============================")
	
	opt := &option.LogOption{
		Engine:      "zap",
		Level:       "DEBUG",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		OTLP:        &option.OTLPOption{},
	}

	logger, err := logger.New(opt)
	if err != nil {
		panic(err)
	}

	// Printf-style logging methods
	userName := "Alice"
	userID := 12345
	loginTime := time.Now()

	logger.Debugf("User %s (ID: %d) debugging at %s", userName, userID, loginTime.Format(time.RFC3339))
	logger.Infof("User %s (ID: %d) logged in at %s", userName, userID, loginTime.Format(time.RFC3339))
	logger.Warnf("User %s (ID: %d) has %d failed login attempts", userName, userID, 3)
	logger.Errorf("User %s (ID: %d) login failed: %s", userName, userID, "invalid password")
	// Note: Fatalf would exit the program, so we skip it in examples
	
	fmt.Println()
}

// demonstrateStructuredMethods shows structured logging methods with key-value pairs
func demonstrateStructuredMethods() {
	fmt.Println("3. Structured Logging Methods")
	fmt.Println("=============================")
	
	opt := &option.LogOption{
		Engine:      "slog",
		Level:       "DEBUG",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		OTLP:        &option.OTLPOption{},
	}

	logger, err := logger.New(opt)
	if err != nil {
		panic(err)
	}

	// Structured logging methods with key-value pairs
	logger.Debugw("User session debug info",
		"user_id", 12345,
		"session_id", "sess_abc123",
		"debug_level", "verbose",
		"timestamp", time.Now(),
	)

	logger.Infow("User activity recorded",
		"user_id", 12345,
		"action", "file_upload",
		"file_name", "document.pdf",
		"file_size", 2048576,
		"duration_ms", 1500,
	)

	logger.Warnw("Rate limit approaching",
		"user_id", 12345,
		"current_requests", 95,
		"limit", 100,
		"window", "1m",
		"remaining_time", "45s",
	)

	logger.Errorw("Database connection failed",
		"user_id", 12345,
		"database", "users_db",
		"error", "connection timeout",
		"retry_count", 3,
		"max_retries", 5,
	)

	// Note: Fatalw would exit the program, so we skip it in examples
	
	fmt.Println()
}

// demonstrateEnhancementMethods shows logger enhancement methods
func demonstrateEnhancementMethods() {
	fmt.Println("4. Logger Enhancement Methods")
	fmt.Println("=============================")
	
	opt := &option.LogOption{
		Engine:      "zap",
		Level:       "INFO",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		OTLP:        &option.OTLPOption{},
	}

	baseLogger, err := logger.New(opt)
	if err != nil {
		panic(err)
	}

	// 4.1 With method - add persistent fields
	fmt.Println("4.1 With method - persistent fields:")
	userLogger := baseLogger.With(
		"user_id", 67890,
		"service", "user-service",
		"version", "1.2.3",
	)
	userLogger.Info("User logged in")
	userLogger.Info("User accessed profile page")

	// 4.2 WithCtx method - add context and fields
	fmt.Println("\n4.2 WithCtx method - context and fields:")
	ctx := context.WithValue(context.Background(), "request_id", "req_xyz789")
	ctxLogger := userLogger.WithCtx(ctx,
		"request_id", ctx.Value("request_id"),
		"method", "GET",
		"path", "/api/user/profile",
	)
	ctxLogger.Info("Processing API request")

	// 4.3 WithCallerSkip method - adjust caller reporting
	fmt.Println("\n4.3 WithCallerSkip method - adjusted caller:")
	skippedLogger := userLogger.WithCallerSkip(1)
	logWithWrapper(skippedLogger)

	fmt.Println()
}

// Helper function to demonstrate WithCallerSkip
func logWithWrapper(l core.Logger) {
	l.Info("This log shows the caller of logWithWrapper, not logWithWrapper itself")
}

// demonstrateGlobalLogger shows global logger usage
func demonstrateGlobalLogger() {
	fmt.Println("5. Global Logger Usage")
	fmt.Println("======================")
	
	// Global logger uses default configuration (slog engine)
	logger.Debug("Global debug message")
	logger.Info("Global info message")
	logger.Warn("Global warning message")

	// Printf-style global methods
	logger.Debugf("Global debug: %s", "formatted message")
	logger.Infof("Global info: processing %d items", 42)
	logger.Warnf("Global warning: %d%% memory usage", 85)

	// Structured global methods
	logger.Debugw("Global structured debug",
		"module", "global_example",
		"debug_enabled", true,
	)

	logger.Infow("Global structured info",
		"event", "system_startup",
		"components", []string{"auth", "db", "cache"},
		"startup_time_ms", 2500,
	)

	logger.Warnw("Global structured warning",
		"alert", "high_cpu_usage",
		"cpu_percent", 92.5,
		"threshold", 90.0,
	)

	fmt.Println()
}

// demonstrateConfiguration shows different configuration options
func demonstrateConfiguration() {
	fmt.Println("6. Configuration Examples")
	fmt.Println("=========================")
	
	// 6.1 Slog engine with console format
	fmt.Println("6.1 Slog engine with console format:")
	slogConsoleOpt := &option.LogOption{
		Engine:      "slog",
		Level:       "INFO",
		Format:      "console",
		OutputPaths: []string{"stdout"},
		Development: true,
		OTLP:        &option.OTLPOption{},
	}

	slogConsole, err := logger.New(slogConsoleOpt)
	if err != nil {
		panic(err)
	}
	slogConsole.Info("Slog console format message")

	// 6.2 Zap engine with production settings
	fmt.Println("\n6.2 Zap engine with production settings:")
	zapProdOpt := &option.LogOption{
		Engine:            "zap",
		Level:             "WARN",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		OTLP:              &option.OTLPOption{},
	}

	zapProd, err := logger.New(zapProdOpt)
	if err != nil {
		panic(err)
	}
	zapProd.Info("This won't show - level is WARN")
	zapProd.Warn("Zap production warning message")

	// 6.3 Level configuration example
	fmt.Println("\n6.3 Dynamic level configuration:")
	levelLogger, err := logger.New(&option.LogOption{
		Engine:      "slog",
		Level:       "ERROR",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		OTLP:        &option.OTLPOption{},
	})
	if err != nil {
		panic(err)
	}

	levelLogger.Info("This won't show - level is ERROR")
	levelLogger.Warn("This won't show - level is ERROR")
	levelLogger.Error("This will show - matches ERROR level")

	fmt.Println()
}

// demonstrateErrorHandling shows error handling and stacktraces
func demonstrateErrorHandling() {
	fmt.Println("7. Error Handling and Stacktraces")
	fmt.Println("==================================")
	
	// 7.1 Slog engine with stacktrace
	fmt.Println("7.1 Slog engine error with stacktrace:")
	slogOpt := &option.LogOption{
		Engine:            "slog",
		Level:             "INFO",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		DisableStacktrace: false,
		OTLP:              &option.OTLPOption{},
	}

	slogLogger, err := logger.New(slogOpt)
	if err != nil {
		panic(err)
	}

	err = simulateError()
	if err != nil {
		slogLogger.Errorw("Slog error with context",
			"error", err.Error(),
			"function", "simulateError",
			"retry_count", 0,
		)
	}

	// 7.2 Zap engine with stacktrace
	fmt.Println("\n7.2 Zap engine error with stacktrace:")
	zapOpt := &option.LogOption{
		Engine:            "zap",
		Level:             "INFO",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		DisableStacktrace: false,
		OTLP:              &option.OTLPOption{},
	}

	zapLogger, err := logger.New(zapOpt)
	if err != nil {
		panic(err)
	}

	err = simulateError()
	if err != nil {
		zapLogger.Errorw("Zap error with context",
			"error", err.Error(),
			"function", "simulateError",
			"retry_count", 0,
		)
	}

	fmt.Println()
}

// Helper function to simulate an error
func simulateError() error {
	return errors.New("simulated database connection failure")
}

// demonstrateContextAndTracing shows context usage and tracing fields
func demonstrateContextAndTracing() {
	fmt.Println("8. Context and Tracing")
	fmt.Println("======================")
	
	opt := &option.LogOption{
		Engine:      "slog",
		Level:       "INFO",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		OTLP:        &option.OTLPOption{},
	}

	baseLogger, err := logger.New(opt)
	if err != nil {
		panic(err)
	}

	// Simulate distributed tracing context
	ctx := context.Background()
	
	// Create a trace context logger
	traceLogger := baseLogger.WithCtx(ctx,
		"trace_id", "trace_abc123def456",
		"span_id", "span_789xyz012",
		"parent_span_id", "span_456parent",
		"service_name", "user-service",
		"operation", "process_request",
	)

	// Log various stages of request processing
	traceLogger.Infow("Request received",
		"method", "POST",
		"endpoint", "/api/users",
		"user_agent", "MyApp/1.0",
		"ip", "192.168.1.100",
	)

	traceLogger.Infow("Database query started",
		"query", "SELECT * FROM users WHERE active = ?",
		"query_params", []interface{}{true},
	)

	time.Sleep(50 * time.Millisecond) // Simulate processing time

	traceLogger.Infow("Database query completed",
		"duration_ms", 45,
		"rows_affected", 3,
	)

	traceLogger.Infow("Request completed",
		"status_code", 200,
		"response_size", 1024,
		"total_duration_ms", 125,
	)

	// Demonstrate field standardization with various formats
	fmt.Println("\n8.1 Field standardization demo:")
	standardLogger := baseLogger.With(
		"ts", time.Now().Unix(),        // Will be mapped to timestamp
		"msg", "custom message",        // Will be mapped to message  
		"trace.id", "trace_field_test", // Will be mapped to trace_id
		"span.id", "span_field_test",   // Will be mapped to span_id
	)

	standardLogger.Info("Field standardization example")

	fmt.Println()
}