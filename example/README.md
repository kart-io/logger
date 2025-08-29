# Logger Usage Examples

This directory contains comprehensive usage examples for the unified logger library. All examples demonstrate the complete functionality of both Slog and Zap engines with consistent APIs and field standardization.

## 📁 Available Examples

### [comprehensive/](comprehensive/)

**Complete feature demonstration** - Shows all logger methods and capabilities

- ✅ All 15 core logger methods (Debug, Info, Warn, Error + Printf + Structured variants)
- ✅ Logger enhancement methods (With, WithCtx, WithCallerSkip)
- ✅ Global logger usage patterns
- ✅ Configuration examples for different environments
- ✅ Error handling with automatic stacktraces
- ✅ Context and distributed tracing integration
- ✅ Field standardization examples

### [zap/](zap/)

**Zap engine focused examples** - Deep dive into Zap-specific features

- ✅ Production vs development configurations
- ✅ High-performance logging patterns
- ✅ Advanced structured logging with rich context
- ✅ Error handling with stacktraces
- ✅ Batch processing and performance optimizations
- ✅ Zero-allocation logging techniques

### [performance/](performance/)

**Performance benchmarking** - Compare engines and optimize usage

- ✅ Single-threaded performance comparison
- ✅ Multi-threaded/concurrent logging benchmarks
- ✅ Memory allocation analysis
- ✅ Different logging pattern performance characteristics
- ✅ Best practices for high-throughput scenarios

### [configuration/](configuration/)

**Configuration management** - All configuration options and integrations

- ✅ Basic to advanced configuration examples
- ✅ Command-line flags integration (pflag)
- ✅ Environment-specific configurations
- ✅ Multiple output paths (stdout, stderr, files)
- ✅ Dynamic level configuration
- ✅ Development vs production settings

### [otlp/](otlp/)

**OTLP Integration** - OpenTelemetry Protocol integration and testing

- ✅ OTLP gRPC and HTTP protocol support
- ✅ Endpoint configuration (127.0.0.1:4327)
- ✅ Distributed tracing context integration
- ✅ Error handling and fallback behaviors
- ✅ Timeout configuration and connection testing

### [slog/slog-demo/](slog/slog-demo/)

**Simple Slog example** - Quick start with Slog engine

- ✅ Basic Slog engine usage
- ✅ Error logging with stacktraces
- ✅ Structured logging example

## 🚀 Quick Start

### Run All Examples

```bash
# Comprehensive examples (recommended starting point)
cd example/comprehensive && go run main.go

# Zap engine deep dive
cd example/zap && go run main.go

# Performance comparison
cd example/performance && go run main.go

# Configuration examples
cd example/configuration && go run main.go

# OTLP integration testing
cd example/otlp && go run main.go

# Simple Slog example
cd example/slog/slog-demo && go run main.go
```

### Key Logger Methods Demonstrated

#### Basic Logging Methods

```go
logger.Debug("Debug message")
logger.Info("Info message")
logger.Warn("Warning message")
logger.Error("Error message")
logger.Fatal("Fatal message") // Exits program
```

#### Printf-style Methods

```go
logger.Debugf("Debug: %s", value)
logger.Infof("User %s logged in at %s", user, time)
logger.Warnf("Memory usage: %d%%", percent)
logger.Errorf("Failed to process %s: %v", item, err)
logger.Fatalf("Critical error: %v", err) // Exits program
```

#### Structured Logging Methods

```go
logger.Debugw("Debug with context", "key", "value")
logger.Infow("User activity", "user_id", 123, "action", "login")
logger.Warnw("High load", "cpu", 90.5, "memory", 85.2)
logger.Errorw("Database error", "error", err, "query", sql)
logger.Fatalw("System failure", "component", "db") // Exits program
```

#### Logger Enhancement Methods

```go
// Create child logger with persistent fields
userLogger := logger.With("user_id", 123, "service", "auth")

// Add context and fields
ctxLogger := logger.WithCtx(ctx, "request_id", reqID)

// Adjust caller reporting for wrapper functions
skipLogger := logger.WithCallerSkip(1)
```

## 🔧 Configuration Options

### Basic Configuration

```go
opt := &option.LogOption{
    Engine:            "slog",           // or "zap"
    Level:             "INFO",           // DEBUG, INFO, WARN, ERROR, FATAL
    Format:            "json",           // or "console"
    OutputPaths:       []string{"stdout"}, // stdout, stderr, file paths
    Development:       false,            // true for development mode
    DisableCaller:     false,            // disable caller information
    DisableStacktrace: false,            // disable automatic stacktraces
    OTLP: &option.OTLPOption{           // OpenTelemetry configuration
        Endpoint: "",
        Protocol: "grpc",
        Timeout:  10 * time.Second,
    },
}

logger, err := logger.New(opt)
```

### Command Line Flags (pflag integration)

```go
fs := pflag.NewFlagSet("myapp", pflag.ContinueOnError)
opt := option.DefaultLogOption()
opt.AddFlags(fs)
fs.Parse(os.Args[1:])

logger, err := logger.New(opt)
```

## 🎯 Key Features Demonstrated

### ✅ Engine Transparency

- **Slog Engine**: Go's standard library structured logging
- **Zap Engine**: High-performance structured logging with zero allocations
- **Unified API**: Same methods work with both engines seamlessly

### ✅ Field Standardization

- Automatic field name mapping: `ts` → `timestamp`, `msg` → `message`
- Tracing field support: `trace.id` → `trace_id`, `span.id` → `span_id`
- Consistent output format across engines

### ✅ Advanced Features

- **Caller Detection**: Shows exact code location (`caller` field)
- **Engine Identification**: `engine` field shows which engine produced log
- **Automatic Stacktraces**: Complete call stack for Error/Fatal levels
- **Context Integration**: Rich context and metadata support

### ✅ Production Ready

- **Performance Optimized**: Zap engine for high-throughput scenarios
- **Configurable**: Environment-specific configurations
- **Observable**: Integration with OpenTelemetry and distributed tracing
- **Flexible Output**: Multiple output destinations (stdout, files, etc.)

## 📊 Sample Output

### Standard Log Entry

```json
{
  "time": "2025-08-29T15:30:00.123456789+08:00",
  "level": "INFO",
  "msg": "User logged in",
  "engine": "slog",
  "caller": "main.go:45",
  "user_id": 12345,
  "action": "login"
}
```

### Error Log with Stacktrace

```json
{
  "time": "2025-08-29T15:30:00.123456789+08:00",
  "level": "ERROR",
  "msg": "Database connection failed",
  "engine": "slog",
  "caller": "main.go:78",
  "error": "connection timeout",
  "stacktrace": "main.connectDB\\n\\tmain.go:78\\nmain.main\\n\\tmain.go:45"
}
```

### Structured Log with Context

```json
{
  "level": "INFO",
  "timestamp": "2025-08-29T15:30:00.123456789+08:00",
  "caller": "main.go:123",
  "message": "Request processed",
  "engine": "zap",
  "trace_id": "abc123def456",
  "span_id": "789xyz012",
  "user_id": 67890,
  "method": "POST",
  "path": "/api/users",
  "duration_ms": 145
}
```

## 🏆 Best Practices Shown

1. **Use structured logging** (`*w` methods) for rich context and searchability
2. **Create child loggers** with `With()` for related operations and persistent fields
3. **Include tracing information** for distributed systems and request correlation
4. **Configure appropriately** for different environments (dev/staging/prod)
5. **Handle errors properly** with contextual information and automatic stacktraces
6. **Use appropriate log levels** to control verbosity and noise
7. **Leverage performance features** like zero-allocation logging for high-throughput scenarios

## 🔗 Related Documentation

- [Core Logger Interface](../core/logger.go)
- [Configuration Options](../option/option.go)
- [Field Standardization](../fields/fields.go)
- [Slog Engine Implementation](../engines/slog/)
- [Zap Engine Implementation](../engines/zap/)
