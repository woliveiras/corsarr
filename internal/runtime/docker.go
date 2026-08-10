package runtime

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

const maxTechnicalDetailLength = 500

var dockerClientVersionPattern = regexp.MustCompile(`(?i)docker version ([^,\s]+)`)

type DockerDetector struct {
	runner  CommandRunner
	timeout time.Duration
}

func NewDockerDetector(runner CommandRunner, timeout time.Duration) *DockerDetector {
	return &DockerDetector{runner: runner, timeout: timeout}
}

func (d *DockerDetector) Check(ctx context.Context) Status {
	status := Status{Provider: ProviderDocker}
	dockerPath, err := d.runner.LookPath("docker")
	if err != nil {
		status.State = StateUnavailable
		return status
	}

	checkContext, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	clientOutput, err := d.runner.Run(checkContext, dockerPath, "--version")
	if err != nil {
		status.State = StateError
		status.TechnicalDetail = boundedDetail(err)
		return status
	}
	status.Version = dockerClientVersion(clientOutput)

	serverVersion, err := d.runner.Run(checkContext, dockerPath, "info", "--format", "{{.ServerVersion}}")
	if err == nil {
		status.State = StateReady
		if version := strings.TrimSpace(serverVersion); version != "" {
			status.Version = version
		}
		return status
	}

	status.TechnicalDetail = boundedDetail(err)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(checkContext.Err(), context.DeadlineExceeded) {
		status.State = StateError
		return status
	}
	if indicatesStoppedDocker(err.Error()) {
		status.State = StateStopped
		return status
	}

	status.State = StateError
	return status
}

func dockerClientVersion(output string) string {
	matches := dockerClientVersionPattern.FindStringSubmatch(output)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func indicatesStoppedDocker(detail string) bool {
	normalized := strings.ToLower(detail)
	markers := []string{
		"cannot connect to the docker daemon",
		"is the docker daemon running",
		"docker desktop is not running",
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

func boundedDetail(err error) string {
	detail := strings.TrimSpace(err.Error())
	if len(detail) <= maxTechnicalDetailLength {
		return detail
	}
	return detail[:maxTechnicalDetailLength] + "…"
}
