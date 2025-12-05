package validator

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// DockerValidator checks Docker and Docker Compose installation
type DockerValidator struct{}

// NewDockerValidator creates a new Docker validator
func NewDockerValidator() *DockerValidator {
	return &DockerValidator{}
}

// Validate checks Docker installation and version
func (dv *DockerValidator) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check if Docker is installed
	dockerInstalled, dockerVersion := dv.checkDocker()
	if !dockerInstalled {
		result.AddError(
			"docker",
			"Docker is not installed or not in PATH",
			SeverityCritical,
		)
	} else {
		// Check Docker version (require at least 20.10)
		if !dv.isDockerVersionValid(dockerVersion) {
			result.AddError(
				"docker",
				fmt.Sprintf("Docker version %s is too old (require 20.10+)", dockerVersion),
				SeverityWarning,
			)
		}

		// Check if Docker daemon is running
		if !dv.isDockerRunning() {
			result.AddError(
				"docker",
				"Docker daemon is not running",
				SeverityError,
			)
		}
	}

	// Check if Docker Compose is installed
	composeInstalled, composeVersion := dv.checkDockerCompose()
	if !composeInstalled {
		result.AddError(
			"docker_compose",
			"Docker Compose is not installed or not in PATH",
			SeverityCritical,
		)
	} else {
		// Check Compose version (require at least 2.0)
		if !dv.isComposeVersionValid(composeVersion) {
			result.AddError(
				"docker_compose",
				fmt.Sprintf("Docker Compose version %s is too old (require 2.0+)", composeVersion),
				SeverityWarning,
			)
		}
	}

	return result
}

// checkDocker checks if Docker is installed and returns version
func (dv *DockerValidator) checkDocker() (bool, string) {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}

	// Parse version from output: "Docker version 24.0.7, build afdd53b"
	versionRegex := regexp.MustCompile(`Docker version ([\d.]+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return true, matches[1]
	}

	return true, "unknown"
}

// checkDockerCompose checks if Docker Compose is installed and returns version
func (dv *DockerValidator) checkDockerCompose() (bool, string) {
	// Try "docker compose" first (modern)
	cmd := exec.Command("docker", "compose", "version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		// Parse version: "Docker Compose version v2.23.0"
		versionRegex := regexp.MustCompile(`version v?([\d.]+)`)
		matches := versionRegex.FindStringSubmatch(string(output))
		if len(matches) > 1 {
			return true, matches[1]
		}
		return true, "unknown"
	}

	// Try "docker-compose" (legacy)
	cmd = exec.Command("docker-compose", "--version")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}

	// Parse version: "docker-compose version 1.29.2"
	versionRegex := regexp.MustCompile(`version ([\d.]+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return true, matches[1]
	}

	return true, "unknown"
}

// isDockerRunning checks if Docker daemon is running
func (dv *DockerValidator) isDockerRunning() bool {
	cmd := exec.Command("docker", "ps")
	err := cmd.Run()
	return err == nil
}

// isDockerVersionValid checks if Docker version meets minimum requirements
func (dv *DockerValidator) isDockerVersionValid(version string) bool {
	if version == "unknown" {
		return true // Assume valid if we can't parse
	}

	return compareVersions(version, "20.10.0") >= 0
}

// isComposeVersionValid checks if Docker Compose version meets minimum requirements
func (dv *DockerValidator) isComposeVersionValid(version string) bool {
	if version == "unknown" {
		return true // Assume valid if we can't parse
	}

	return compareVersions(version, "2.0.0") >= 0
}

// compareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	for i := 0; i < 3; i++ {
		if parts1[i] < parts2[i] {
			return -1
		}
		if parts1[i] > parts2[i] {
			return 1
		}
	}

	return 0
}

// parseVersion parses a version string into [major, minor, patch]
func parseVersion(version string) [3]int {
	parts := [3]int{0, 0, 0}
	components := strings.Split(version, ".")

	for i := 0; i < len(components) && i < 3; i++ {
		// Remove any non-numeric suffix (e.g., "2.23.0-rc1" -> "2.23.0")
		numPart := regexp.MustCompile(`^\d+`).FindString(components[i])
		if num, err := strconv.Atoi(numPart); err == nil {
			parts[i] = num
		}
	}

	return parts
}

// GetDockerInfo returns information about Docker installation
func GetDockerInfo() map[string]string {
	info := make(map[string]string)
	
	dv := NewDockerValidator()
	
	dockerInstalled, dockerVersion := dv.checkDocker()
	info["docker_installed"] = fmt.Sprintf("%v", dockerInstalled)
	info["docker_version"] = dockerVersion
	info["docker_running"] = fmt.Sprintf("%v", dv.isDockerRunning())
	
	composeInstalled, composeVersion := dv.checkDockerCompose()
	info["compose_installed"] = fmt.Sprintf("%v", composeInstalled)
	info["compose_version"] = composeVersion
	
	return info
}
