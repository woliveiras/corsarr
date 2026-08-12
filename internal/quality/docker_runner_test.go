package quality

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type failingEnvironmentCommandRunner struct {
	errorText string
}

func (f failingEnvironmentCommandRunner) LookPath(string) (string, error) {
	return "/usr/local/bin/docker", nil
}

func (f failingEnvironmentCommandRunner) RunWithEnvironment(
	context.Context,
	map[string]string,
	string,
	...string,
) (string, error) {
	return "", fmt.Errorf("%s", f.errorText)
}

func TestDockerRunnerPreservesFailureDetailWithoutEnvironmentSecrets(t *testing.T) {
	const secret = "radarr-secret-api-key"
	runner := NewDockerRunner(
		failingEnvironmentCommandRunner{
			errorText: "Recyclarr rejected the configuration using " + secret,
		},
		time.Second,
	)

	err := runner.Run(context.Background(), map[string]string{"RADARR_API_KEY": secret}, "sync")
	if !errors.Is(err, ErrSyncFailed) {
		t.Fatalf("expected sync failure, got %v", err)
	}
	if !strings.Contains(err.Error(), "Recyclarr rejected the configuration") {
		t.Fatalf("expected diagnostic detail, got %v", err)
	}
	if strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "[REDACTED]") {
		t.Fatalf("environment secret was not redacted: %v", err)
	}
}
