package validator

import (
	"testing"

	"github.com/woliveiras/corsarr/internal/services"
)

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:    "test_field",
		Message:  "test message",
		Severity: SeverityError,
	}

	expected := "[ERROR] test_field: test message"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{Valid: true}

	// Add a warning - should not invalidate
	result.AddError("field1", "warning message", SeverityWarning)
	if !result.Valid {
		t.Error("Adding warning should not invalidate result")
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	// Add an error - should invalidate
	result.AddError("field2", "error message", SeverityError)
	if result.Valid {
		t.Error("Adding error should invalidate result")
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	result := &ValidationResult{Valid: true}

	if result.HasErrors() {
		t.Error("New result should not have errors")
	}

	result.AddError("field", "message", SeverityWarning)
	if result.HasErrors() {
		t.Error("Warnings should not count as errors")
	}

	result.AddError("field", "message", SeverityError)
	if !result.HasErrors() {
		t.Error("Result should have errors")
	}
}

func TestNewConfig(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	tests := []struct {
		name        string
		serviceIDs  []string
		basePath    string
		outputDir   string
		vpnEnabled  bool
		expectError bool
	}{
		{
			name:        "Valid config",
			serviceIDs:  []string{"sonarr", "radarr"},
			basePath:    "/data",
			outputDir:   "/output",
			vpnEnabled:  false,
			expectError: false,
		},
		{
			name:        "Invalid service ID",
			serviceIDs:  []string{"nonexistent"},
			basePath:    "/data",
			outputDir:   "/output",
			vpnEnabled:  false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(registry, tt.serviceIDs, tt.basePath, tt.outputDir, tt.vpnEnabled)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if config == nil {
					t.Error("Config should not be nil")
				}
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.severity.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.severity.String())
			}
		})
	}
}
