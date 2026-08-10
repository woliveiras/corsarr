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
	setup := &desktopSetupManager{}
	app := &App{directoryPicker: picker, storageInspector: inspector, setup: setup}

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
	if setup.savedStorage != picker.path {
		t.Fatalf("expected selected path %q to be persisted, got %q", picker.path, setup.savedStorage)
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

func TestPrepareStorageLayoutUsesOnlyPersistedSetup(t *testing.T) {
	setup := &desktopSetupManager{status: application.SetupStatus{
		StoragePath:  "/Users/test/Media",
		Applications: []string{"prowlarr", "radarr"},
		CanInstall:   true,
	}}
	layout := &desktopLayoutPreparer{status: storage.LayoutStatus{
		RootPath: "/Users/test/Media/Corsarr",
	}}
	app := &App{setup: setup, layoutPreparer: layout}

	status, err := app.PrepareStorageLayout()
	if err != nil {
		t.Fatalf("prepare storage layout: %v", err)
	}
	if status.RootPath != layout.status.RootPath {
		t.Fatalf("expected root %q, got %q", layout.status.RootPath, status.RootPath)
	}
	if layout.basePath != setup.status.StoragePath {
		t.Fatalf("expected persisted base path %q, got %q", setup.status.StoragePath, layout.basePath)
	}
}

func TestPrepareStorageLayoutRejectsIncompleteSetupWithoutWriting(t *testing.T) {
	setup := &desktopSetupManager{status: application.SetupStatus{
		StoragePath: "/Users/test/Media",
		CanInstall:  false,
	}}
	layout := &desktopLayoutPreparer{}
	app := &App{setup: setup, layoutPreparer: layout}

	_, err := app.PrepareStorageLayout()
	if err == nil {
		t.Fatal("expected incomplete setup to be rejected")
	}
	if layout.prepareCalls != 0 {
		t.Fatalf("expected no layout write, got %d calls", layout.prepareCalls)
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

type desktopSetupManager struct {
	status       application.SetupStatus
	savedStorage string
}

func (f *desktopSetupManager) Load() (application.SetupStatus, error) {
	return f.status, nil
}

func (f *desktopSetupManager) SaveStorage(path string) (application.SetupStatus, error) {
	f.savedStorage = path
	f.status.StoragePath = path
	return f.status, nil
}

func (f *desktopSetupManager) SaveApplications(applicationIDs []string) (application.SetupStatus, error) {
	f.status.Applications = applicationIDs
	return f.status, nil
}

type desktopLayoutPreparer struct {
	status         storage.LayoutStatus
	basePath       string
	applicationIDs []string
	prepareCalls   int
}

func (f *desktopLayoutPreparer) Prepare(
	basePath string,
	applicationIDs []string,
) (storage.LayoutStatus, error) {
	f.prepareCalls++
	f.basePath = basePath
	f.applicationIDs = applicationIDs
	return f.status, nil
}
