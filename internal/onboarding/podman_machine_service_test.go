package onboarding

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

func TestPodmanMachineServiceReturnsWhenRuntimeIsAlreadyReady(t *testing.T) {
	probe := &sequenceProbe{statuses: []containerruntime.Status{{
		Provider: containerruntime.ProviderPodman,
		State:    containerruntime.StateReady,
		Version:  "5.6.2",
	}}}
	runner := &podmanMachineRunner{path: "/opt/homebrew/bin/podman"}
	service := newTestPodmanMachineService(probe, runner, "darwin")

	result, err := service.Prepare(context.Background())
	if err != nil {
		t.Fatalf("prepare ready runtime: %v", err)
	}
	want := PodmanMachinePreparationResult{Ready: true, Version: "5.6.2"}
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("unexpected result\nwant: %#v\n got: %#v", want, result)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no machine access, got %v", runner.calls)
	}
}

func TestPodmanMachineServiceInitializesMissingDefaultMachine(t *testing.T) {
	probe := &sequenceProbe{statuses: []containerruntime.Status{
		{Provider: containerruntime.ProviderPodman, State: containerruntime.StateStopped},
		{Provider: containerruntime.ProviderPodman, State: containerruntime.StateStopped},
		{Provider: containerruntime.ProviderPodman, State: containerruntime.StateReady, Version: "5.6.2"},
	}}
	runner := &podmanMachineRunner{
		path: "/opt/homebrew/bin/podman",
		results: []podmanMachineCommandResult{
			{err: errors.New("Error: podman-machine-default: VM does not exist")},
			{output: "Machine init complete"},
			{output: "Machine started"},
		},
	}
	service := newTestPodmanMachineService(probe, runner, "darwin")

	result, err := service.Prepare(context.Background())
	if err != nil {
		t.Fatalf("prepare missing machine: %v", err)
	}
	wantResult := PodmanMachinePreparationResult{
		Ready: true, Initialized: true, Started: true, Version: "5.6.2",
	}
	if !reflect.DeepEqual(result, wantResult) {
		t.Fatalf("unexpected result\nwant: %#v\n got: %#v", wantResult, result)
	}
	wantCalls := []podmanMachineCall{
		{name: runner.path, args: []string{"machine", "inspect", "podman-machine-default", "--format", "{{.State}}"}},
		{name: runner.path, args: []string{"machine", "init", "--cpus", "4", "--memory", "4096", "--disk-size", "100", "--update-connection=true", "podman-machine-default"}},
		{name: runner.path, args: []string{"machine", "start", "--quiet", "--update-connection=true", "podman-machine-default"}},
	}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("unexpected commands\nwant: %#v\n got: %#v", wantCalls, runner.calls)
	}
}

func TestPodmanMachineServiceStartsExistingStoppedMachine(t *testing.T) {
	probe := &sequenceProbe{statuses: []containerruntime.Status{
		{Provider: containerruntime.ProviderPodman, State: containerruntime.StateStopped},
		{Provider: containerruntime.ProviderPodman, State: containerruntime.StateReady},
	}}
	runner := &podmanMachineRunner{
		path: "/opt/homebrew/bin/podman",
		results: []podmanMachineCommandResult{
			{output: "stopped\n"},
			{output: "Machine started"},
		},
	}
	service := newTestPodmanMachineService(probe, runner, "windows")

	result, err := service.Prepare(context.Background())
	if err != nil {
		t.Fatalf("prepare stopped machine: %v", err)
	}
	if !result.Ready || !result.Started || result.Initialized {
		t.Fatalf("unexpected result %#v", result)
	}
	if len(runner.calls) != 2 || runner.calls[1].args[1] != "start" {
		t.Fatalf("expected inspect then start, got %v", runner.calls)
	}
}

func TestPodmanMachineServiceDoesNotUseMachineOnLinux(t *testing.T) {
	probe := &sequenceProbe{statuses: []containerruntime.Status{{
		Provider: containerruntime.ProviderPodman,
		State:    containerruntime.StateStopped,
	}}}
	runner := &podmanMachineRunner{path: "/usr/bin/podman"}
	service := newTestPodmanMachineService(probe, runner, "linux")

	_, err := service.Prepare(context.Background())
	if err == nil {
		t.Fatal("expected native Linux runtime error")
	}
	if len(runner.calls) != 0 || runner.lookPathCalls != 0 {
		t.Fatalf("Linux preparation must not access Podman Machine, got %v", runner.calls)
	}
}

func TestPodmanMachineServiceRejectsUnexpectedInspectFailure(t *testing.T) {
	probe := &sequenceProbe{statuses: []containerruntime.Status{{
		Provider: containerruntime.ProviderPodman,
		State:    containerruntime.StateStopped,
	}}}
	runner := &podmanMachineRunner{
		path:    "/opt/homebrew/bin/podman",
		results: []podmanMachineCommandResult{{err: errors.New("permission denied")}},
	}
	service := newTestPodmanMachineService(probe, runner, "darwin")

	_, err := service.Prepare(context.Background())
	if err == nil {
		t.Fatal("expected inspect failure")
	}
	if len(runner.calls) != 1 {
		t.Fatalf("unexpected mutation after inspect failure: %v", runner.calls)
	}
}

func newTestPodmanMachineService(
	probe containerruntime.Probe,
	runner containerruntime.CommandRunner,
	platform string,
) *PodmanMachineService {
	service := NewPodmanMachineService(probe, runner, platform)
	service.pollInterval = time.Millisecond
	service.timeout = 100 * time.Millisecond
	return service
}

type sequenceProbe struct {
	statuses []containerruntime.Status
}

func (p *sequenceProbe) Check(context.Context) containerruntime.Status {
	if len(p.statuses) == 0 {
		return containerruntime.Status{Provider: containerruntime.ProviderPodman, State: containerruntime.StateError}
	}
	status := p.statuses[0]
	if len(p.statuses) > 1 {
		p.statuses = p.statuses[1:]
	}
	return status
}

type podmanMachineCall struct {
	name string
	args []string
}

type podmanMachineCommandResult struct {
	output string
	err    error
}

type podmanMachineRunner struct {
	path          string
	lookPathErr   error
	lookPathCalls int
	calls         []podmanMachineCall
	results       []podmanMachineCommandResult
}

func (r *podmanMachineRunner) LookPath(string) (string, error) {
	r.lookPathCalls++
	return r.path, r.lookPathErr
}

func (r *podmanMachineRunner) Run(
	_ context.Context,
	name string,
	args ...string,
) (string, error) {
	r.calls = append(r.calls, podmanMachineCall{name: name, args: append([]string(nil), args...)})
	if len(r.results) == 0 {
		return "", nil
	}
	result := r.results[0]
	r.results = r.results[1:]
	return result.output, result.err
}
