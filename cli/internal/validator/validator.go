package validator

import (
	"fmt"

	"github.com/woliveiras/corsarr/internal/services"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
	Severity Severity
}

// Severity indicates how critical a validation error is
type Severity int

const (
	SeverityWarning Severity = iota
	SeverityError
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", ve.Severity, ve.Field, ve.Message)
}

// ValidationResult holds the results of validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(field, message string, severity Severity) {
	err := ValidationError{
		Field:    field,
		Message:  message,
		Severity: severity,
	}

	if severity == SeverityWarning {
		vr.Warnings = append(vr.Warnings, err)
	} else {
		vr.Errors = append(vr.Errors, err)
		vr.Valid = false
	}
}

// HasErrors returns true if there are any errors (not warnings)
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// Validator interface for all validators
type Validator interface {
	Validate() *ValidationResult
}

// Config holds all validation configuration
type Config struct {
	Services       []*services.Service
	Registry       *services.Registry
	BasePath       string
	OutputDir      string
	VPNEnabled     bool
	SkipDockerCheck bool
}

// NewConfig creates a new validation config
func NewConfig(registry *services.Registry, serviceIDs []string, basePath, outputDir string, vpnEnabled bool) (*Config, error) {
	selectedServices := make([]*services.Service, 0, len(serviceIDs))
	
	for _, id := range serviceIDs {
		service, err := registry.GetService(id)
		if err != nil {
			return nil, fmt.Errorf("service %s not found: %w", id, err)
		}
		selectedServices = append(selectedServices, service)
	}

	return &Config{
		Services:   selectedServices,
		Registry:   registry,
		BasePath:   basePath,
		OutputDir:  outputDir,
		VPNEnabled: vpnEnabled,
	}, nil
}

// ValidateAll runs all validators and returns combined results
func ValidateAll(config *Config) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Run all validators
	validators := []Validator{
		NewPortValidator(config),
		NewDependencyValidator(config),
		NewPathValidator(config),
	}

	// Add Docker validator if not skipped
	if !config.SkipDockerCheck {
		validators = append(validators, NewDockerValidator())
	}

	// Run each validator and merge results
	for _, validator := range validators {
		vResult := validator.Validate()
		result.Errors = append(result.Errors, vResult.Errors...)
		result.Warnings = append(result.Warnings, vResult.Warnings...)
		if vResult.HasErrors() {
			result.Valid = false
		}
	}

	return result
}
