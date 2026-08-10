package runtime

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDockerDetectorReportsUnavailableWhenClientIsMissing(t *testing.T) {
	runner := &fakeCommandRunner{lookPathErr: errors.New("executable not found")}

	status := NewDockerDetector(runner, time.Second).Check(context.Background())

	if status.State != StateUnavailable {
		t.Fatalf("expected unavailable state, got %q", status.State)
	}
	if status.Provider != ProviderDocker {
		t.Fatalf("expected Docker provider, got %q", status.Provider)
	}
	if runner.runCalls != 0 {
		t.Fatalf("expected no commands when Docker is missing, got %d", runner.runCalls)
	}
}

func TestDockerDetectorDistinguishesStoppedDaemonFromMissingClient(t *testing.T) {
	runner := &fakeCommandRunner{
		outputs: map[string]commandResult{
			"docker --version": {output: "Docker version 28.3.2, build 578ccf6"},
			"docker info --format {{.ServerVersion}}": {
				err: errors.New("Cannot connect to the Docker daemon. Is the docker daemon running?"),
			},
		},
	}

	status := NewDockerDetector(runner, time.Second).Check(context.Background())

	if status.State != StateStopped {
		t.Fatalf("expected stopped state, got %q", status.State)
	}
	if status.Version != "28.3.2" {
		t.Fatalf("expected client version 28.3.2, got %q", status.Version)
	}
}

func TestDockerDetectorReportsReadyServer(t *testing.T) {
	runner := &fakeCommandRunner{
		outputs: map[string]commandResult{
			"docker --version":                        {output: "Docker version 28.3.2, build 578ccf6"},
			"docker info --format {{.ServerVersion}}": {output: "28.3.2\n"},
		},
	}

	status := NewDockerDetector(runner, time.Second).Check(context.Background())

	if status.State != StateReady {
		t.Fatalf("expected ready state, got %q", status.State)
	}
	if status.Version != "28.3.2" {
		t.Fatalf("expected server version 28.3.2, got %q", status.Version)
	}
}

func TestDockerDetectorTimesOutWithoutHangingDesktop(t *testing.T) {
	runner := &fakeCommandRunner{
		outputs: map[string]commandResult{
			"docker --version": {output: "Docker version 28.3.2, build 578ccf6"},
		},
		blockCommand: "docker info --format {{.ServerVersion}}",
	}

	started := time.Now()
	status := NewDockerDetector(runner, 20*time.Millisecond).Check(context.Background())

	if status.State != StateError {
		t.Fatalf("expected error state after timeout, got %q", status.State)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("detector exceeded bounded timeout: %s", elapsed)
	}
}

type commandResult struct {
	output string
	err    error
}

type fakeCommandRunner struct {
	lookPathErr  error
	outputs      map[string]commandResult
	blockCommand string
	runCalls     int
}

func (f *fakeCommandRunner) LookPath(string) (string, error) {
	if f.lookPathErr != nil {
		return "", f.lookPathErr
	}
	return "docker", nil
}

func (f *fakeCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	f.runCalls++
	key := name
	for _, arg := range args {
		key += " " + arg
	}

	if key == f.blockCommand {
		<-ctx.Done()
		return "", ctx.Err()
	}

	result, ok := f.outputs[key]
	if !ok {
		return "", errors.New("unexpected command: " + key)
	}
	return result.output, result.err
}
