package config

import (
	"time"

	"github.com/kart-io/logger/core"
)

// Config represents the complete logger configuration.
type Config struct {
	// Engine specifies which logging engine to use ("zap" or "slog")
	Engine string `yaml:"engine" json:"engine" env:"LOG_ENGINE"`

	// Level sets the minimum logging level
	Level string `yaml:"level" json:"level" env:"LOG_LEVEL"`

	// Format specifies output format ("json" or "console")
	Format string `yaml:"format" json:"format" env:"LOG_FORMAT"`

	// OutputPaths specifies where logs should be written
	OutputPaths []string `yaml:"output-paths" json:"output_paths" env:"LOG_OUTPUT_PATHS"`

	// OTLP configuration (flattened and nested)
	OTLPEndpoint string      `yaml:"otlp-endpoint" json:"otlp_endpoint" env:"LOG_OTLP_ENDPOINT"`
	OTLP         *OTLPConfig `yaml:"otlp" json:"otlp"`

	// Development mode enables caller info and stacktraces
	Development bool `yaml:"development" json:"development" env:"LOG_DEVELOPMENT"`

	// DisableCaller disables automatic caller detection
	DisableCaller bool `yaml:"disable-caller" json:"disable_caller" env:"LOG_DISABLE_CALLER"`

	// DisableStacktrace disables automatic stacktrace capture
	DisableStacktrace bool `yaml:"disable-stacktrace" json:"disable_stacktrace" env:"LOG_DISABLE_STACKTRACE"`
}

// OTLPConfig contains OTLP-specific configuration.
type OTLPConfig struct {
	Enabled  *bool             `yaml:"enabled" json:"enabled" env:"LOG_OTLP_ENABLED"`
	Endpoint string            `yaml:"endpoint" json:"endpoint" env:"LOG_OTLP_ENDPOINT"`
	Protocol string            `yaml:"protocol" json:"protocol" env:"LOG_OTLP_PROTOCOL"`
	Timeout  time.Duration     `yaml:"timeout" json:"timeout" env:"LOG_OTLP_TIMEOUT"`
	Headers  map[string]string `yaml:"headers" json:"headers"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Engine:            "slog",
		Level:             "INFO",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		OTLP: &OTLPConfig{
			Protocol: "grpc",
			Timeout:  10 * time.Second,
		},
	}
}

// Validate checks the configuration for consistency and applies intelligent defaults.
func (c *Config) Validate() error {
	// Parse and validate log level
	if _, err := core.ParseLevel(c.Level); err != nil {
		return err
	}

	// Apply OTLP intelligent configuration resolution
	c.resolveOTLPConfig()

	// Validate engine selection
	if c.Engine != "zap" && c.Engine != "slog" {
		c.Engine = "slog" // Default fallback
	}

	return nil
}

// resolveOTLPConfig implements the intelligent OTLP configuration resolution
// as specified in the requirements document.
func (c *Config) resolveOTLPConfig() {
	if c.OTLP == nil {
		c.OTLP = &OTLPConfig{}
	}

	// Check for environment variable override (highest priority)
	// Note: Environment variables are read at startup and require reload for runtime changes

	// Apply flattened configuration logic
	if c.OTLPEndpoint != "" {
		// If explicit enabled=false is set, respect user intent
		if c.OTLP.Enabled != nil && !*c.OTLP.Enabled {
			// User explicitly disabled OTLP, keep disabled
			return
		}

		// Auto-enable OTLP when endpoint is provided (intelligent detection)
		if c.OTLP.Enabled == nil {
			enabled := true
			c.OTLP.Enabled = &enabled
		}

		// Use flattened endpoint if nested endpoint is not set
		if c.OTLP.Endpoint == "" {
			c.OTLP.Endpoint = c.OTLPEndpoint
		}
	} else {
		// No flattened endpoint, use nested configuration
		if c.OTLP.Enabled == nil && c.OTLP.Endpoint != "" {
			// Auto-enable if nested endpoint is provided
			enabled := true
			c.OTLP.Enabled = &enabled
		}
	}

	// Apply defaults for enabled OTLP
	if c.OTLP.Enabled != nil && *c.OTLP.Enabled {
		if c.OTLP.Protocol == "" {
			c.OTLP.Protocol = "grpc"
		}
		if c.OTLP.Timeout == 0 {
			c.OTLP.Timeout = 10 * time.Second
		}
	}
}

// IsOTLPEnabled returns true if OTLP is enabled after configuration resolution.
func (c *Config) IsOTLPEnabled() bool {
	return c.OTLP != nil && c.OTLP.Enabled != nil && *c.OTLP.Enabled && c.OTLP.Endpoint != ""
}