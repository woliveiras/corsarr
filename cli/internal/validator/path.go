package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// PathValidator checks file system paths
type PathValidator struct {
	config *Config
}

// NewPathValidator creates a new path validator
func NewPathValidator(config *Config) *PathValidator {
	return &PathValidator{config: config}
}

// Validate checks paths are valid and accessible
func (pv *PathValidator) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check base path (ARRPATH)
	if pv.config.BasePath == "" {
		result.AddError(
			"base_path",
			"Base path (ARRPATH) is required",
			SeverityError,
		)
	} else {
		// Check if path exists
		if !pathExists(pv.config.BasePath) {
			result.AddError(
				"base_path",
				fmt.Sprintf("Base path does not exist: %s", pv.config.BasePath),
				SeverityWarning,
			)
		} else {
			// Check if path is writable
			if !isWritable(pv.config.BasePath) {
				result.AddError(
					"base_path",
					fmt.Sprintf("Base path is not writable: %s", pv.config.BasePath),
					SeverityError,
				)
			}

			// Check disk space
			availableGB := getAvailableDiskSpaceGB(pv.config.BasePath)
			if availableGB < 10 {
				result.AddError(
					"disk_space",
					fmt.Sprintf("Low disk space: %.1f GB available (recommended: at least 10 GB)", availableGB),
					SeverityWarning,
				)
			}
		}
	}

	// Check output directory
	if pv.config.OutputDir == "" {
		result.AddError(
			"output_dir",
			"Output directory is required",
			SeverityError,
		)
	} else {
		// Create output directory if it doesn't exist
		if !pathExists(pv.config.OutputDir) {
			if err := os.MkdirAll(pv.config.OutputDir, 0755); err != nil {
				result.AddError(
					"output_dir",
					fmt.Sprintf("Cannot create output directory: %s", err),
					SeverityError,
				)
			}
		} else {
			// Check if output directory is writable
			if !isWritable(pv.config.OutputDir) {
				result.AddError(
					"output_dir",
					fmt.Sprintf("Output directory is not writable: %s", pv.config.OutputDir),
					SeverityError,
				)
			}
		}
	}

	return result
}

// pathExists checks if a path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isWritable checks if a path is writable
func isWritable(path string) bool {
	// Try to create a temporary file
	testFile := filepath.Join(path, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	_ = file.Close()
	_ = os.Remove(testFile)
	return true
}

// getAvailableDiskSpaceGB returns available disk space in GB
func getAvailableDiskSpaceGB(path string) float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0
	}

	// Available blocks * block size / GB
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(availableBytes) / (1024 * 1024 * 1024)
	
	return availableGB
}

// ValidatePath performs a quick validation of a single path
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	if !pathExists(path) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	if !isWritable(path) {
		return fmt.Errorf("path is not writable: %s", path)
	}

	return nil
}

// EnsurePathExists creates a path if it doesn't exist
func EnsurePathExists(path string) error {
	if pathExists(path) {
		return nil
	}

	return os.MkdirAll(path, 0755)
}
