package application

import (
	"context"
	"testing"

	"github.com/woliveiras/corsarr/internal/hostreadiness"
	runtimeenv "github.com/woliveiras/corsarr/internal/runtime"
)

func TestEnvironmentStatusCombinesHostAndRuntimeWithoutMutation(t *testing.T) {
	probe := &fakeRuntimeProbe{status: runtimeenv.Status{
		Provider: runtimeenv.ProviderDocker,
		State:    runtimeenv.StateReady,
		Version:  "28.3.2",
	}}

	host := &fakeHostChecker{status: hostreadiness.Status{Ready: true, MemoryBytes: 8}}
	environment := NewEnvironmentService(probe, "darwin", "arm64", host)
	status := environment.Status(context.Background())

	if status.Platform != "darwin" {
		t.Fatalf("expected darwin platform, got %q", status.Platform)
	}
	if status.Architecture != "arm64" {
		t.Fatalf("expected arm64 architecture, got %q", status.Architecture)
	}
	if status.Runtime.State != runtimeenv.StateReady {
		t.Fatalf("expected ready runtime, got %q", status.Runtime.State)
	}
	if !status.Host.Ready || host.calls != 1 {
		t.Fatalf("unexpected host readiness %#v, calls=%d", status.Host, host.calls)
	}
	if probe.calls != 1 {
		t.Fatalf("expected one read-only runtime probe, got %d", probe.calls)
	}
}

type fakeHostChecker struct {
	status hostreadiness.Status
	calls  int
}

func (f *fakeHostChecker) Check(context.Context) hostreadiness.Status {
	f.calls++
	return f.status
}

type fakeRuntimeProbe struct {
	status runtimeenv.Status
	calls  int
}

func (f *fakeRuntimeProbe) Check(context.Context) runtimeenv.Status {
	f.calls++
	return f.status
}
