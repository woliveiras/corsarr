package quality

import (
	"context"
	"fmt"
	"strings"
	"time"

	runtimeenv "github.com/woliveiras/corsarr/internal/runtime"
)

type environmentCommandRunner interface {
	LookPath(file string) (string, error)
	RunWithEnvironment(ctx context.Context, environment map[string]string, name string, args ...string) (string, error)
}

type DockerRunner struct {
	runner  environmentCommandRunner
	timeout time.Duration
}

func NewDockerRunner(runner environmentCommandRunner, timeout time.Duration) *DockerRunner {
	return &DockerRunner{runner: runner, timeout: timeout}
}

func NewPlatformDockerRunner(timeout time.Duration) *DockerRunner {
	return NewDockerRunner(runtimeenv.OSCommandRunner{}, timeout)
}

func (r *DockerRunner) Run(ctx context.Context, environment map[string]string, arguments ...string) error {
	dockerPath, err := r.runner.LookPath("docker")
	if err != nil {
		return fmt.Errorf("find Docker client: %w", err)
	}
	operationContext, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	if _, err := r.runner.RunWithEnvironment(operationContext, environment, dockerPath, arguments...); err != nil {
		return fmt.Errorf("%w: %s", ErrSyncFailed, redactEnvironmentValues(err.Error(), environment))
	}
	return nil
}

func redactEnvironmentValues(detail string, environment map[string]string) string {
	for _, value := range environment {
		if strings.TrimSpace(value) != "" {
			detail = strings.ReplaceAll(detail, value, "[REDACTED]")
		}
	}
	return detail
}
