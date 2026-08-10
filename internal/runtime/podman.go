package runtime

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

var podmanClientVersionPattern = regexp.MustCompile(`(?i)podman version ([^,\s]+)`)

// PodmanDetector performs only read-only, bounded client and server checks. On
// macOS and Windows, a stopped Podman Machine is reported as a stopped runtime.
type PodmanDetector struct {
	runner  CommandRunner
	timeout time.Duration
}

var _ Probe = (*PodmanDetector)(nil)

func NewPodmanDetector(runner CommandRunner, timeout time.Duration) *PodmanDetector {
	return &PodmanDetector{runner: runner, timeout: timeout}
}

func (d *PodmanDetector) Check(ctx context.Context) Status {
	status := Status{Provider: ProviderPodman}
	podmanPath, err := d.runner.LookPath("podman")
	if err != nil {
		status.State = StateUnavailable
		return status
	}

	checkContext, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	clientOutput, err := d.runner.Run(checkContext, podmanPath, "--version")
	if err != nil {
		status.State = StateError
		status.TechnicalDetail = boundedDetail(err)
		return status
	}
	status.Version = podmanClientVersion(clientOutput)

	serverVersion, err := d.runner.Run(
		checkContext,
		podmanPath,
		"info", "--format", "{{.Version.Version}}",
	)
	if err == nil {
		status.State = StateReady
		if version := strings.TrimSpace(serverVersion); version != "" {
			status.Version = version
		}
		return status
	}

	status.TechnicalDetail = boundedDetail(err)
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(checkContext.Err(), context.DeadlineExceeded) {
		status.State = StateError
		return status
	}
	if indicatesStoppedPodman(err.Error()) {
		status.State = StateStopped
		return status
	}

	status.State = StateError
	return status
}

func podmanClientVersion(output string) string {
	matches := podmanClientVersionPattern.FindStringSubmatch(output)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func indicatesStoppedPodman(detail string) bool {
	normalized := strings.ToLower(detail)
	markers := []string{
		"cannot connect to podman",
		"unable to connect to podman socket",
		"podman machine is not running",
		"connection refused",
		"no such file or directory",
	}
	for _, marker := range markers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
