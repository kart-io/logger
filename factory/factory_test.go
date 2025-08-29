package factory

import (
	"strings"
	"testing"

	"github.com/kart-io/logger/option"
)

func TestNewLoggerFactory(t *testing.T) {
	opt := option.DefaultLogOption()
	factory := NewLoggerFactory(opt)

	if factory == nil {
		t.Error("NewLoggerFactory() returned nil")
	}

	if factory.option != opt {
		t.Error("Factory option was not set correctly")
	}
}

func TestLoggerFactory_GetOption(t *testing.T) {
	opt := option.DefaultLogOption()
	factory := NewLoggerFactory(opt)

	if got := factory.GetOption(); got != opt {
		t.Errorf("GetOption() = %v, want %v", got, opt)
	}
}

func TestLoggerFactory_UpdateOption(t *testing.T) {
	factory := NewLoggerFactory(option.DefaultLogOption())
	
	newOpt := &option.LogOption{
		Engine: "zap",
		Level:  "DEBUG",
		Format: "console",
		OTLP:   &option.OTLPOption{},
	}

	err := factory.UpdateOption(newOpt)
	if err != nil {
		t.Errorf("UpdateOption() error = %v", err)
	}

	if factory.GetOption().Engine != "zap" {
		t.Errorf("Expected engine to be updated to 'zap', got %s", factory.GetOption().Engine)
	}

	if factory.GetOption().Level != "DEBUG" {
		t.Errorf("Expected level to be updated to 'DEBUG', got %s", factory.GetOption().Level)
	}
}

func TestLoggerFactory_UpdateOption_InvalidConfig(t *testing.T) {
	factory := NewLoggerFactory(option.DefaultLogOption())
	
	invalidOpt := &option.LogOption{
		Engine: "slog",
		Level:  "INVALID_LEVEL", // This should cause validation to fail
		OTLP:   &option.OTLPOption{},
	}

	err := factory.UpdateOption(invalidOpt)
	if err == nil {
		t.Error("Expected UpdateOption() to return error for invalid config")
	}

	// Original configuration should remain unchanged
	if factory.GetOption().Level != "INFO" {
		t.Errorf("Expected original level to be preserved, got %s", factory.GetOption().Level)
	}
}

func TestLoggerFactory_CreateLogger_UnsupportedEngine(t *testing.T) {
	// Note: Since option validation automatically converts invalid engines to "slog",
	// this test verifies the factory behavior when engines are not implemented yet.
	// In the future, we could test with truly unsupported engines after validation is updated.
	
	opt := &option.LogOption{
		Engine: "unsupported-engine", // This will be converted to "slog" during validation
		Level:  "INFO",
		OTLP:   &option.OTLPOption{},
	}
	
	factory := NewLoggerFactory(opt)
	logger, err := factory.CreateLogger()

	// With Slog now implemented, this should actually succeed
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if logger == nil {
		t.Error("Expected logger to be created successfully")
	}

	// Verify that the engine was normalized during validation and logger was created
	if factory.GetOption().Engine != "slog" {
		t.Errorf("Expected engine to be normalized to 'slog' during validation, got %s", factory.GetOption().Engine)
	}
}

func TestLoggerFactory_CreateLogger_InvalidConfig(t *testing.T) {
	opt := &option.LogOption{
		Engine: "slog",
		Level:  "INVALID_LEVEL",
		OTLP:   &option.OTLPOption{},
	}
	
	factory := NewLoggerFactory(opt)
	logger, err := factory.CreateLogger()

	if err == nil {
		t.Error("Expected error for invalid configuration")
	}

	if logger != nil {
		t.Error("Expected logger to be nil for invalid configuration")
	}

	if !strings.Contains(err.Error(), "invalid configuration") {
		t.Errorf("Expected error message to contain 'invalid configuration', got %s", err.Error())
	}
}

func TestLoggerFactory_CreateLogger_EngineImplementationStatus(t *testing.T) {
	// Test the current implementation status of engines
	tests := []struct {
		name        string
		engine      string
		expectError bool
		description string
	}{
		{"zap engine", "zap", false, "Zap engine is now implemented"},
		{"slog engine", "slog", false, "Slog engine is implemented"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &option.LogOption{
				Engine: tt.engine,
				Level:  "INFO",
				OTLP:   &option.OTLPOption{},
			}
			
			factory := NewLoggerFactory(opt)
			logger, err := factory.CreateLogger()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s engine (%s)", tt.engine, tt.description)
				}
				if logger != nil {
					t.Errorf("Expected logger to be nil for %s engine (%s)", tt.engine, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s engine (%s): %v", tt.engine, tt.description, err)
				}
				if logger == nil {
					t.Errorf("Expected logger to be created for %s engine (%s)", tt.engine, tt.description)
				}
			}
		})
	}
}

func TestLoggerFactory_FallbackBehavior(t *testing.T) {
	// Test the fallback behavior described in the CreateLogger method
	tests := []struct {
		name        string
		engine      string
		expectError bool
		description string
	}{
		{
			name:        "zap works directly",
			engine:      "zap",
			expectError: false, // Should succeed directly with Zap
			description: "Should create zap logger directly",
		},
		{
			name:        "slog works directly",
			engine:      "slog",
			expectError: false, // Should succeed directly
			description: "Should create slog logger directly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &option.LogOption{
				Engine: tt.engine,
				Level:  "INFO",
				OTLP:   &option.OTLPOption{},
			}
			
			factory := NewLoggerFactory(opt)
			logger, err := factory.CreateLogger()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if logger != nil {
					t.Errorf("Expected nil logger but got one")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if logger == nil {
					t.Errorf("Expected logger but got nil")
				}
			}
		})
	}
}

func TestLoggerFactory_ConfigurationIntegrity(t *testing.T) {
	// Test that the factory maintains configuration integrity
	opt := &option.LogOption{
		Engine: "slog",
		Level:  "DEBUG",
		Format: "json",
		OutputPaths: []string{"stdout", "file.log"},
		Development: true,
		OTLP: &option.OTLPOption{
			Protocol: "grpc",
		},
	}

	factory := NewLoggerFactory(opt)

	// Verify that the factory maintains the original configuration
	retrieved := factory.GetOption()
	if retrieved.Engine != "slog" {
		t.Errorf("Expected engine 'slog', got %s", retrieved.Engine)
	}
	if retrieved.Level != "DEBUG" {
		t.Errorf("Expected level 'DEBUG', got %s", retrieved.Level)
	}
	if !retrieved.Development {
		t.Error("Expected development mode to be true")
	}
	if retrieved.OTLP.Protocol != "grpc" {
		t.Errorf("Expected OTLP protocol 'grpc', got %s", retrieved.OTLP.Protocol)
	}
}