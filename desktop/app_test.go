package main

import (
	"context"
	"testing"

	"github.com/woliveiras/corsarr/internal/application"
	runtimeenv "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

func TestNewAppExposesCatalogWithoutArbitraryURLs(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("create desktop app: %v", err)
	}

	if applications := app.ListApplications(); len(applications) == 0 {
		t.Fatal("expected desktop application catalog")
	}

	if err := app.OpenApplication("https://attacker.example"); err == nil {
		t.Fatal("expected arbitrary URL to be rejected before reaching Wails runtime")
	}
}

func TestChooseStorageLocationInspectsOnlyUserSelectedDirectory(t *testing.T) {
	picker := &desktopDirectoryPicker{path: "/Users/test/Media"}
	inspector := &desktopStorageInspector{status: storage.Status{
		Path:      "/Users/test/Media",
		State:     storage.StateReady,
		Writable:  true,
		Hardlinks: true,
	}}
	app := &App{directoryPicker: picker, storageInspector: inspector}

	status, err := app.ChooseStorageLocation()
	if err != nil {
		t.Fatalf("choose storage location: %v", err)
	}
	if status.State != storage.StateReady {
		t.Fatalf("expected ready storage, got %q", status.State)
	}
	if inspector.inspectedPath != picker.path {
		t.Fatalf("expected selected path %q to be inspected, got %q", picker.path, inspector.inspectedPath)
	}
}

func TestChooseStorageLocationDoesNotInspectAfterCancel(t *testing.T) {
	inspector := &desktopStorageInspector{}
	app := &App{
		directoryPicker:  &desktopDirectoryPicker{},
		storageInspector: inspector,
	}

	status, err := app.ChooseStorageLocation()
	if err != nil {
		t.Fatalf("cancel storage selection: %v", err)
	}
	if status.State != storage.StateCanceled {
		t.Fatalf("expected canceled state, got %q", status.State)
	}
	if inspector.calls != 0 {
		t.Fatalf("expected no inspection after cancel, got %d", inspector.calls)
	}
}

func TestGetEnvironmentStatusUsesReadOnlyProbe(t *testing.T) {
	probe := &desktopRuntimeProbe{status: runtimeenv.Status{
		Provider: runtimeenv.ProviderDocker,
		State:    runtimeenv.StateStopped,
		Version:  "28.3.2",
	}}
	app := &App{
		environment: application.NewEnvironmentService(probe, "darwin", "arm64"),
	}

	status := app.GetEnvironmentStatus()

	if status.Runtime.State != runtimeenv.StateStopped {
		t.Fatalf("expected stopped runtime, got %q", status.Runtime.State)
	}
	if probe.calls != 1 {
		t.Fatalf("expected one read-only probe, got %d", probe.calls)
	}
}

type desktopRuntimeProbe struct {
	status runtimeenv.Status
	calls  int
}

func (f *desktopRuntimeProbe) Check(context.Context) runtimeenv.Status {
	f.calls++
	return f.status
}

type desktopDirectoryPicker struct {
	path string
	err  error
}

func (f *desktopDirectoryPicker) Choose(context.Context) (string, error) {
	return f.path, f.err
}

type desktopStorageInspector struct {
	status        storage.Status
	inspectedPath string
	calls         int
}

func (f *desktopStorageInspector) Inspect(path string) storage.Status {
	f.calls++
	f.inspectedPath = path
	return f.status
}
