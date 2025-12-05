package validator

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},
		{"20.10.0", "19.03.0", 1},
		{"2.23.0", "2.0.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version  string
		expected [3]int
	}{
		{"1.0.0", [3]int{1, 0, 0}},
		{"20.10.5", [3]int{20, 10, 5}},
		{"2.23.0-rc1", [3]int{2, 23, 0}},
		{"1.2", [3]int{1, 2, 0}},
		{"5", [3]int{5, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := parseVersion(tt.version)
			if result != tt.expected {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestDockerValidator_isDockerVersionValid(t *testing.T) {
	dv := NewDockerValidator()

	tests := []struct {
		version string
		valid   bool
	}{
		{"20.10.0", true},
		{"20.10.5", true},
		{"24.0.7", true},
		{"19.03.0", false},
		{"18.09.0", false},
		{"unknown", true}, // Assume valid if unknown
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := dv.isDockerVersionValid(tt.version)
			if result != tt.valid {
				t.Errorf("isDockerVersionValid(%q) = %v, want %v", tt.version, result, tt.valid)
			}
		})
	}
}

func TestDockerValidator_isComposeVersionValid(t *testing.T) {
	dv := NewDockerValidator()

	tests := []struct {
		version string
		valid   bool
	}{
		{"2.0.0", true},
		{"2.23.0", true},
		{"1.29.2", false},
		{"1.27.0", false},
		{"unknown", true}, // Assume valid if unknown
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := dv.isComposeVersionValid(tt.version)
			if result != tt.valid {
				t.Errorf("isComposeVersionValid(%q) = %v, want %v", tt.version, result, tt.valid)
			}
		})
	}
}

func TestGetDockerInfo(t *testing.T) {
	info := GetDockerInfo()

	requiredKeys := []string{
		"docker_installed",
		"docker_version",
		"docker_running",
		"compose_installed",
		"compose_version",
	}

	for _, key := range requiredKeys {
		if _, ok := info[key]; !ok {
			t.Errorf("Missing key in Docker info: %s", key)
		}
	}
}
