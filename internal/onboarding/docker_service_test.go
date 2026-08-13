package onboarding

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

func TestDockerServiceInstallsStartsAndWaitsForReady(t *testing.T) {
	probe := &preparationProbe{statuses: []containerruntime.Status{
		{State: containerruntime.StateUnavailable},
		{State: containerruntime.StateStopped},
		{State: containerruntime.StateReady, Version: "29.7.2"},
	}}
	runner := &preparationRunner{dockerPath: "/usr/local/bin/docker"}
	installer := &preparationInstaller{}
	service := NewDockerService(probe, runner, installer, "darwin")
	service.exists = func(string) bool { return false }
	service.pollInterval = time.Millisecond
	service.timeout = time.Second

	result, err := service.Prepare(context.Background())
	if err != nil {
		t.Fatalf("prepare Docker Desktop: %v", err)
	}
	if !result.Ready || !result.Installed || !result.Started || installer.calls != 1 {
		t.Fatalf("unexpected preparation result %#v, installer calls %d", result, installer.calls)
	}
	if !reflect.DeepEqual(runner.operations, []string{"/usr/local/bin/docker desktop start --detach"}) {
		t.Fatalf("unexpected start operations %v", runner.operations)
	}
}

func TestDockerServiceFallsBackToOpeningInstalledApplication(t *testing.T) {
	probe := &preparationProbe{statuses: []containerruntime.Status{
		{State: containerruntime.StateStopped}, {State: containerruntime.StateReady},
	}}
	runner := &preparationRunner{lookPathErr: errors.New("not found")}
	service := NewDockerService(probe, runner, nil, "darwin")
	service.exists = func(string) bool { return true }
	service.pollInterval = time.Millisecond

	result, err := service.Prepare(context.Background())
	if err != nil || !result.Ready || result.Installed {
		t.Fatalf("unexpected result %#v, %v", result, err)
	}
	if !reflect.DeepEqual(runner.operations, []string{"/usr/bin/open -a /Applications/Docker.app"}) {
		t.Fatalf("unexpected fallback operations %v", runner.operations)
	}
}

func TestDockerServiceRecoveryStartsExistingRuntimeWithoutInstalling(t *testing.T) {
	probe := &preparationProbe{statuses: []containerruntime.Status{
		{State: containerruntime.StateStopped}, {State: containerruntime.StateReady, Version: "29.7.2"},
	}}
	runner := &preparationRunner{dockerPath: "/usr/local/bin/docker"}
	installer := &preparationInstaller{}
	service := NewDockerService(probe, runner, installer, "darwin")
	service.exists = func(string) bool { return true }
	service.pollInterval = time.Millisecond

	result, err := service.Recover(context.Background())
	if err != nil || !result.Ready || !result.Started || result.Installed {
		t.Fatalf("unexpected recovery result %#v, %v", result, err)
	}
	if installer.calls != 0 {
		t.Fatalf("recovery must not invoke installer, got %d calls", installer.calls)
	}
	if !reflect.DeepEqual(runner.operations, []string{"/usr/local/bin/docker desktop start --detach"}) {
		t.Fatalf("unexpected recovery start operations %v", runner.operations)
	}
}

func TestDockerServiceRecoveryRefusesToInstallMissingRuntime(t *testing.T) {
	probe := &preparationProbe{statuses: []containerruntime.Status{{
		State: containerruntime.StateUnavailable,
	}}}
	installer := &preparationInstaller{}
	service := NewDockerService(probe, &preparationRunner{}, installer, "darwin")
	service.exists = func(string) bool { return false }

	if _, err := service.Recover(context.Background()); err == nil {
		t.Fatal("expected missing runtime recovery error")
	}
	if installer.calls != 0 {
		t.Fatalf("recovery must not install missing runtime, got %d calls", installer.calls)
	}
}

type preparationProbe struct {
	statuses []containerruntime.Status
	calls    int
}

func (p *preparationProbe) Check(context.Context) containerruntime.Status {
	index := p.calls
	if index >= len(p.statuses) {
		index = len(p.statuses) - 1
	}
	p.calls++
	return p.statuses[index]
}

type preparationRunner struct {
	dockerPath  string
	lookPathErr error
	operations  []string
}

func (r *preparationRunner) LookPath(string) (string, error) { return r.dockerPath, r.lookPathErr }
func (r *preparationRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	operation := name
	for _, argument := range args {
		operation += " " + argument
	}
	r.operations = append(r.operations, operation)
	return "", nil
}

type preparationInstaller struct{ calls int }

func (i *preparationInstaller) Install(context.Context) error {
	i.calls++
	return nil
}
