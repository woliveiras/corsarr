package hostreadiness

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const (
	MinimumMemoryBytes   = 4 * 1024 * 1024 * 1024
	MinimumFreeDiskBytes = 4 * 1024 * 1024 * 1024
	minimumMacOSMajor    = 14
)

type Status struct {
	Ready         bool     `json:"ready"`
	OSVersion     string   `json:"osVersion,omitempty"`
	MemoryBytes   uint64   `json:"memoryBytes,omitempty"`
	FreeDiskBytes uint64   `json:"freeDiskBytes,omitempty"`
	Issues        []string `json:"issues"`
}

type Checker interface {
	Check(ctx context.Context) Status
}

type Facts struct {
	Platform      string
	Architecture  string
	OSVersion     string
	MemoryBytes   uint64
	FreeDiskBytes uint64
}

func Evaluate(facts Facts) Status {
	status := Status{
		OSVersion: facts.OSVersion, MemoryBytes: facts.MemoryBytes,
		FreeDiskBytes: facts.FreeDiskBytes, Issues: []string{},
	}
	if facts.Platform != "darwin" {
		status.Issues = append(status.Issues, "a preparação automática ainda não é suportada neste sistema")
	}
	if facts.Architecture != "arm64" && facts.Architecture != "amd64" {
		status.Issues = append(status.Issues, "a arquitetura deste computador não é suportada")
	}
	major, err := macOSMajor(facts.OSVersion)
	if facts.Platform == "darwin" && (err != nil || major < minimumMacOSMajor) {
		status.Issues = append(status.Issues, "é necessário macOS 14 ou mais recente")
	}
	if facts.MemoryBytes < MinimumMemoryBytes {
		status.Issues = append(status.Issues, "são necessários pelo menos 4 GiB de memória")
	}
	if facts.FreeDiskBytes < MinimumFreeDiskBytes {
		status.Issues = append(status.Issues, "são necessários pelo menos 4 GiB livres para preparar o ambiente")
	}
	status.Ready = len(status.Issues) == 0
	return status
}

func macOSMajor(version string) (int, error) {
	first, _, _ := strings.Cut(strings.TrimSpace(version), ".")
	major, err := strconv.Atoi(first)
	if err != nil || major < 1 {
		return 0, fmt.Errorf("invalid macOS version")
	}
	return major, nil
}
