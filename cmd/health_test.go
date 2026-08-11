package cmd

import (
	"context"
	"errors"
	"testing"
)

func TestInspectDockerAvailabilityDistinguishesPermissionErrorsFromMissingCLI(t *testing.T) {
	t.Parallel()

	status := inspectDockerAvailabilityWith(
		func(string) (string, error) { return "/usr/bin/docker", nil },
		func(context.Context, string) (string, error) {
			return "permission denied while trying to connect to the Docker daemon socket", errors.New("exit status 1")
		},
	)

	if !status.installed {
		t.Fatal("Docker CLI was incorrectly reported as not installed")
	}
	if status.available {
		t.Fatal("Docker daemon was incorrectly reported as available")
	}
	if status.detail != "permission denied while trying to connect to the Docker daemon socket" {
		t.Fatalf("unexpected availability detail %q", status.detail)
	}
}

func TestInspectDockerAvailabilityReportsMissingCLI(t *testing.T) {
	t.Parallel()

	status := inspectDockerAvailabilityWith(
		func(string) (string, error) { return "", errors.New("executable file not found") },
		func(context.Context, string) (string, error) {
			t.Fatal("docker info must not run without the Docker CLI")
			return "", nil
		},
	)

	if status.installed || status.available {
		t.Fatalf("unexpected availability status %#v", status)
	}
}

func TestInspectDockerAvailabilityReportsAccessibleDaemon(t *testing.T) {
	t.Parallel()

	status := inspectDockerAvailabilityWith(
		func(string) (string, error) { return "/usr/bin/docker", nil },
		func(context.Context, string) (string, error) { return "", nil },
	)

	if !status.installed || !status.available || status.detail != "" {
		t.Fatalf("unexpected availability status %#v", status)
	}
}
