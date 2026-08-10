package runtime

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestPodmanDetectorReportsUnavailableWhenClientIsMissing(t *testing.T) {
	runner := &recordingCommandRunner{lookPathErr: errors.New("executable not found")}

	status := NewPodmanDetector(runner, time.Second).Check(context.Background())

	if status.State != StateUnavailable || status.Provider != ProviderPodman {
		t.Fatalf("unexpected status %#v", status)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no commands when Podman is missing, got %v", runner.calls)
	}
}

func TestPodmanDetectorDistinguishesStoppedMachine(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/opt/homebrew/bin/podman",
		results: []managerCommandResult{
			{output: "podman version 5.6.2"},
			{err: errors.New("Error: podman machine is not running")},
		},
	}

	status := NewPodmanDetector(runner, time.Second).Check(context.Background())

	if status.State != StateStopped || status.Version != "5.6.2" {
		t.Fatalf("unexpected stopped status %#v", status)
	}
	want := []commandCall{
		{name: runner.path, args: []string{"--version"}},
		{name: runner.path, args: []string{"info", "--format", "{{.Version.Version}}"}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected probe commands\nwant: %#v\n got: %#v", want, runner.calls)
	}
}

func TestPodmanDetectorReportsReadyServerVersion(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/usr/bin/podman",
		results: []managerCommandResult{
			{output: "podman version 5.6.2"},
			{output: "5.6.1\n"},
		},
	}

	status := NewPodmanDetector(runner, time.Second).Check(context.Background())

	want := Status{Provider: ProviderPodman, State: StateReady, Version: "5.6.1"}
	if !reflect.DeepEqual(status, want) {
		t.Fatalf("unexpected ready status\nwant: %#v\n got: %#v", want, status)
	}
}

func TestPodmanDetectorBoundsUnexpectedTechnicalDetail(t *testing.T) {
	runner := &recordingCommandRunner{
		path: "/usr/bin/podman",
		results: []managerCommandResult{
			{output: "podman version 5.6.2"},
			{err: errors.New(string(make([]byte, maxTechnicalDetailLength+100)))},
		},
	}

	status := NewPodmanDetector(runner, time.Second).Check(context.Background())

	if status.State != StateError {
		t.Fatalf("expected error status, got %#v", status)
	}
	if len(status.TechnicalDetail) > maxTechnicalDetailLength+len("…") {
		t.Fatalf("technical detail was not bounded: %d", len(status.TechnicalDetail))
	}
}
