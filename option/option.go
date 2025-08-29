package option

import (
	"time"

	"github.com/kart-io/logger/core"
	"github.com/spf13/pflag"
)

// LogOption represents the complete logger configuration.
type LogOption struct {
	// Engine specifies which logging engine to use ("zap" or "slog")
	Engine string `json:"engine" mapstructure:"engine"`

	// Level sets the minimum logging level
	Level string `json:"level" mapstructure:"level"`

	// Format specifies output format ("json" or "console")
	Format string `json:"format" mapstructure:"format"`

	// OutputPaths specifies where logs should be written
	OutputPaths []string `json:"output_paths" mapstructure:"output_paths"`

	// OTLP configuration (flattened and nested)
	OTLPEndpoint string      `json:"otlp_endpoint" mapstructure:"otlp_endpoint"`
	OTLP         *OTLPOption `json:"otlp" mapstructure:"otlp"`

	// Development mode enables caller info and stacktraces
	Development bool `json:"development" mapstructure:"development"`

	// DisableCaller disables automatic caller detection
	DisableCaller bool `json:"disable_caller" mapstructure:"disable_caller"`

	// DisableStacktrace disables automatic stacktrace capture
	DisableStacktrace bool `json:"disable_stacktrace" mapstructure:"disable_stacktrace"`
}

// OTLPOption contains OTLP-specific configuration.
type OTLPOption struct {
	Enabled  *bool             `json:"enabled" mapstructure:"enabled"`
	Endpoint string            `json:"endpoint" mapstructure:"endpoint"`
	Protocol string            `json:"protocol" mapstructure:"protocol"`
	Timeout  time.Duration     `json:"timeout" mapstructure:"timeout"`
	Headers  map[string]string `json:"headers" mapstructure:"headers"`
	Insecure bool              `json:"insecure" mapstructure:"insecure"`
}

// DefaultLogOption returns a configuration with sensible defaults.
func DefaultLogOption() *LogOption {
	return &LogOption{
		Engine:            "slog",
		Level:             "INFO",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		OTLP: &OTLPOption{
			Protocol: "grpc",
			Timeout:  10 * time.Second,
			Insecure: true, // Default to insecure for development
		},
	}
}

// AddFlags adds configuration flags to the provided pflag.FlagSet.
func (opt *LogOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&opt.Engine, "engine", "slog", "Logging engine (zap|slog)")
	fs.StringVar(&opt.Level, "level", "INFO", "Log level (DEBUG|INFO|WARN|ERROR|FATAL)")
	fs.StringVar(&opt.Format, "format", "json", "Log format (json|console)")
	fs.StringSliceVar(&opt.OutputPaths, "output-paths", []string{"stdout"}, "Output paths for logs")
	fs.StringVar(&opt.OTLPEndpoint, "otlp-endpoint", "", "OTLP endpoint URL")
	fs.BoolVar(&opt.Development, "development", false, "Enable development mode")
	fs.BoolVar(&opt.DisableCaller, "disable-caller", false, "Disable caller detection")
	fs.BoolVar(&opt.DisableStacktrace, "disable-stacktrace", false, "Disable stacktrace capture")

	// OTLP nested options
	if opt.OTLP == nil {
		opt.OTLP = &OTLPOption{}
	}
	fs.StringVar(&opt.OTLP.Endpoint, "otlp.endpoint", "", "OTLP nested endpoint URL")
	fs.StringVar(&opt.OTLP.Protocol, "otlp.protocol", "grpc", "OTLP protocol (grpc|http)")
	fs.DurationVar(&opt.OTLP.Timeout, "otlp.timeout", 10*time.Second, "OTLP timeout duration")
}

// Validate checks the configuration for consistency and applies intelligent defaults.
func (opt *LogOption) Validate() error {
	// Parse and validate log level
	if _, err := core.ParseLevel(opt.Level); err != nil {
		return err
	}

	// Apply OTLP intelligent configuration resolution
	opt.resolveOTLPConfig()

	// Validate engine selection
	if opt.Engine != "zap" && opt.Engine != "slog" {
		opt.Engine = "slog" // Default fallback
	}

	return nil
}

// resolveOTLPConfig implements the intelligent OTLP configuration resolution
// as specified in the requirements document.
func (opt *LogOption) resolveOTLPConfig() {
	if opt.OTLP == nil {
		opt.OTLP = &OTLPOption{}
	}

	// Check for environment variable override (highest priority)
	// Note: Environment variables are read at startup and require reload for runtime changes

	// Apply flattened configuration logic
	if opt.OTLPEndpoint != "" {
		// If explicit enabled=false is set, respect user intent
		if opt.OTLP.Enabled != nil && !*opt.OTLP.Enabled {
			// User explicitly disabled OTLP, keep disabled
			return
		}

		// Auto-enable OTLP when endpoint is provided (intelligent detection)
		if opt.OTLP.Enabled == nil {
			enabled := true
			opt.OTLP.Enabled = &enabled
		}

		// Use flattened endpoint (priority over nested endpoint)
		opt.OTLP.Endpoint = opt.OTLPEndpoint
	} else {
		// No flattened endpoint, use nested configuration
		if opt.OTLP.Enabled == nil && opt.OTLP.Endpoint != "" {
			// Auto-enable if nested endpoint is provided
			enabled := true
			opt.OTLP.Enabled = &enabled
		}
	}

	// Apply defaults for enabled OTLP
	if opt.OTLP.Enabled != nil && *opt.OTLP.Enabled {
		if opt.OTLP.Protocol == "" {
			opt.OTLP.Protocol = "grpc"
		}
		if opt.OTLP.Timeout == 0 {
			opt.OTLP.Timeout = 10 * time.Second
		}
	}
}

// IsOTLPEnabled returns true if OTLP is enabled after configuration resolution.
func (opt *LogOption) IsOTLPEnabled() bool {
	return opt.OTLP != nil && opt.OTLP.Enabled != nil && *opt.OTLP.Enabled && opt.OTLP.Endpoint != ""
}

// IsEnabled returns true if OTLP is enabled.
func (opt *OTLPOption) IsEnabled() bool {
	return opt != nil && opt.Enabled != nil && *opt.Enabled && opt.Endpoint != ""
}