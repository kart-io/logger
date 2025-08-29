package option

import (
	"reflect"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func TestDefaultLogOption(t *testing.T) {
	opt := DefaultLogOption()

	if opt.Engine != "slog" {
		t.Errorf("Expected Engine to be 'slog', got %s", opt.Engine)
	}

	if opt.Level != "INFO" {
		t.Errorf("Expected Level to be 'INFO', got %s", opt.Level)
	}

	if opt.Format != "json" {
		t.Errorf("Expected Format to be 'json', got %s", opt.Format)
	}

	expectedPaths := []string{"stdout"}
	if !reflect.DeepEqual(opt.OutputPaths, expectedPaths) {
		t.Errorf("Expected OutputPaths to be %v, got %v", expectedPaths, opt.OutputPaths)
	}

	if opt.Development != false {
		t.Errorf("Expected Development to be false, got %t", opt.Development)
	}

	if opt.OTLP == nil {
		t.Fatal("Expected OTLP to be initialized")
	}

	if opt.OTLP.Protocol != "grpc" {
		t.Errorf("Expected OTLP Protocol to be 'grpc', got %s", opt.OTLP.Protocol)
	}

	if opt.OTLP.Timeout != 10*time.Second {
		t.Errorf("Expected OTLP Timeout to be 10s, got %v", opt.OTLP.Timeout)
	}
}

func TestLogOption_AddFlags(t *testing.T) {
	opt := DefaultLogOption()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

	opt.AddFlags(fs)

	// Test that flags are registered
	expectedFlags := []string{
		"engine", "level", "format", "output-paths", "otlp-endpoint",
		"development", "disable-caller", "disable-stacktrace",
		"otlp.endpoint", "otlp.protocol", "otlp.timeout",
	}

	for _, flagName := range expectedFlags {
		if fs.Lookup(flagName) == nil {
			t.Errorf("Flag %s was not registered", flagName)
		}
	}

	// Test flag default values
	if flag := fs.Lookup("engine"); flag.DefValue != "slog" {
		t.Errorf("Expected engine default to be 'slog', got %s", flag.DefValue)
	}

	if flag := fs.Lookup("level"); flag.DefValue != "INFO" {
		t.Errorf("Expected level default to be 'INFO', got %s", flag.DefValue)
	}

	if flag := fs.Lookup("otlp.protocol"); flag.DefValue != "grpc" {
		t.Errorf("Expected otlp.protocol default to be 'grpc', got %s", flag.DefValue)
	}
}

func TestLogOption_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opt     *LogOption
		wantErr bool
	}{
		{
			name:    "valid default option",
			opt:     DefaultLogOption(),
			wantErr: false,
		},
		{
			name: "invalid log level",
			opt: &LogOption{
				Engine: "slog",
				Level:  "INVALID",
				Format: "json",
			},
			wantErr: true,
		},
		{
			name: "invalid engine gets corrected",
			opt: &LogOption{
				Engine: "invalid",
				Level:  "INFO",
				Format: "json",
				OTLP:   &OTLPOption{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opt.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check that invalid engine was corrected to slog
			if tt.name == "invalid engine gets corrected" && tt.opt.Engine != "slog" {
				t.Errorf("Expected invalid engine to be corrected to 'slog', got %s", tt.opt.Engine)
			}
		})
	}
}

func TestLogOption_resolveOTLPConfig(t *testing.T) {
	tests := []struct {
		name     string
		opt      *LogOption
		expected bool
	}{
		{
			name: "OTLP enabled with flattened endpoint",
			opt: &LogOption{
				OTLPEndpoint: "http://localhost:4317",
				OTLP:         &OTLPOption{},
			},
			expected: true,
		},
		{
			name: "OTLP enabled with nested endpoint",
			opt: &LogOption{
				OTLP: &OTLPOption{
					Endpoint: "http://localhost:4317",
				},
			},
			expected: true,
		},
		{
			name: "OTLP explicitly disabled",
			opt: &LogOption{
				OTLPEndpoint: "http://localhost:4317",
				OTLP: &OTLPOption{
					Enabled: boolPtr(false),
				},
			},
			expected: false,
		},
		{
			name: "No endpoint provided",
			opt: &LogOption{
				OTLP: &OTLPOption{},
			},
			expected: false,
		},
		{
			name: "Flattened endpoint priority over nested",
			opt: &LogOption{
				OTLPEndpoint: "http://flattened:4317",
				OTLP: &OTLPOption{
					Endpoint: "http://nested:4317",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opt.resolveOTLPConfig()
			
			if got := tt.opt.IsOTLPEnabled(); got != tt.expected {
				t.Errorf("IsOTLPEnabled() = %v, expected %v", got, tt.expected)
			}

			// Check that defaults are applied when enabled
			if tt.opt.IsOTLPEnabled() {
				if tt.opt.OTLP.Protocol == "" {
					t.Error("Expected OTLP protocol to be set when enabled")
				}
				if tt.opt.OTLP.Timeout == 0 {
					t.Error("Expected OTLP timeout to be set when enabled")
				}
			}

			// Test flattened endpoint priority
			if tt.name == "Flattened endpoint priority over nested" {
				if tt.opt.OTLP.Endpoint != "http://flattened:4317" {
					t.Errorf("Expected nested endpoint to use flattened value, got %s", tt.opt.OTLP.Endpoint)
				}
			}
		})
	}
}

func TestLogOption_IsOTLPEnabled(t *testing.T) {
	tests := []struct {
		name     string
		opt      *LogOption
		expected bool
	}{
		{
			name: "enabled with endpoint",
			opt: &LogOption{
				OTLP: &OTLPOption{
					Enabled:  boolPtr(true),
					Endpoint: "http://localhost:4317",
				},
			},
			expected: true,
		},
		{
			name: "disabled",
			opt: &LogOption{
				OTLP: &OTLPOption{
					Enabled:  boolPtr(false),
					Endpoint: "http://localhost:4317",
				},
			},
			expected: false,
		},
		{
			name: "enabled without endpoint",
			opt: &LogOption{
				OTLP: &OTLPOption{
					Enabled: boolPtr(true),
				},
			},
			expected: false,
		},
		{
			name: "nil OTLP",
			opt:  &LogOption{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opt.IsOTLPEnabled(); got != tt.expected {
				t.Errorf("IsOTLPEnabled() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestLogOption_StructTags(t *testing.T) {
	// Test that struct fields have correct tags
	optType := reflect.TypeOf(LogOption{})
	
	tests := []struct {
		fieldName string
		jsonTag   string
		mapTag    string
	}{
		{"Engine", "engine", "engine"},
		{"Level", "level", "level"},
		{"Format", "format", "format"},
		{"OutputPaths", "output_paths", "output_paths"},
		{"OTLPEndpoint", "otlp_endpoint", "otlp_endpoint"},
		{"OTLP", "otlp", "otlp"},
		{"Development", "development", "development"},
		{"DisableCaller", "disable_caller", "disable_caller"},
		{"DisableStacktrace", "disable_stacktrace", "disable_stacktrace"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field, found := optType.FieldByName(tt.fieldName)
			if !found {
				t.Fatalf("Field %s not found", tt.fieldName)
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag != tt.jsonTag {
				t.Errorf("Field %s: expected json tag %s, got %s", tt.fieldName, tt.jsonTag, jsonTag)
			}

			mapTag := field.Tag.Get("mapstructure")
			if mapTag != tt.mapTag {
				t.Errorf("Field %s: expected mapstructure tag %s, got %s", tt.fieldName, tt.mapTag, mapTag)
			}
		})
	}
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}