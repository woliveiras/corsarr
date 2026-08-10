package main

import (
	"context"
	"testing"

	"github.com/woliveiras/corsarr/internal/application"
	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/credentials"
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
		CanPrepare:   true,
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

func TestInstallSelectedApplicationsUsesBoundedApplicationService(t *testing.T) {
	installation := &desktopInstallationManager{result: application.InstallationResult{Complete: true}}
	app := &App{installation: installation}

	result, err := app.InstallSelectedApplications()
	if err != nil {
		t.Fatalf("install selected applications: %v", err)
	}
	if !result.Complete || installation.calls != 1 {
		t.Fatalf("expected one completed installation call, got result=%#v calls=%d", result, installation.calls)
	}
}

func TestArchiveApplicationDataUsesBoundedApplicationService(t *testing.T) {
	data := &desktopApplicationDataManager{result: storage.ArchivedApplicationData{
		ApplicationID: "radarr",
		Archived:      true,
	}}
	app := &App{applicationData: data}

	result, err := app.ArchiveApplicationData("radarr")
	if err != nil {
		t.Fatalf("archive application data: %v", err)
	}
	if !result.Archived || data.applicationID != "radarr" || data.calls != 1 {
		t.Fatalf("expected one bounded archive call, result=%#v manager=%#v", result, data)
	}
}

func TestCopyQBittorrentPasswordWritesOnlyToNativeClipboard(t *testing.T) {
	access := &desktopServiceAccess{password: credentials.NewSecret("private-password")}
	clipboard := &desktopClipboard{}
	app := &App{serviceAccess: access, clipboard: clipboard}

	if err := app.CopyQBittorrentPassword(); err != nil {
		t.Fatalf("copy qBittorrent password: %v", err)
	}
	if clipboard.value != "private-password" || clipboard.calls != 1 {
		t.Fatalf("expected one native clipboard write, got %#v", clipboard)
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

func (f *desktopSetupManager) AcceptCurrentTerms() (application.SetupStatus, error) {
	f.status.TermsAccepted = true
	f.status.CanInstall = f.status.CanPrepare
	return f.status, nil
}

type desktopLayoutPreparer struct {
	status         storage.LayoutStatus
	basePath       string
	applicationIDs []string
	prepareCalls   int
}

type desktopInstallationManager struct {
	result application.InstallationResult
	calls  int
}

type desktopApplicationDataManager struct {
	result        storage.ArchivedApplicationData
	applicationID string
	calls         int
}

type desktopServiceAccess struct {
	password credentials.Secret
}

func (a *desktopServiceAccess) QBittorrentStatus(context.Context) (application.ServiceAccessStatus, error) {
	return application.ServiceAccessStatus{
		ApplicationID: "qbittorrent",
		Username:      "corsarr",
		Available:     true,
	}, nil
}

func (a *desktopServiceAccess) QBittorrentPassword(context.Context) (credentials.Secret, error) {
	return a.password, nil
}

type desktopClipboard struct {
	value string
	calls int
}

func (c *desktopClipboard) SetText(_ context.Context, value string) error {
	c.calls++
	c.value = value
	return nil
}

func (m *desktopApplicationDataManager) ListStatuses() ([]storage.ApplicationDataStatus, error) {
	return nil, nil
}

func (m *desktopApplicationDataManager) Archive(
	_ context.Context,
	applicationID string,
) (storage.ArchivedApplicationData, error) {
	m.calls++
	m.applicationID = applicationID
	return m.result, nil
}

func (m *desktopInstallationManager) InstallSelected(
	context.Context,
	runtimecatalog.RuntimeOptions,
) (application.InstallationResult, error) {
	m.calls++
	return m.result, nil
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
