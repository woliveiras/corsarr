package main

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/woliveiras/corsarr/internal/application"
	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/credentials"
	"github.com/woliveiras/corsarr/internal/diagnostics"
	"github.com/woliveiras/corsarr/internal/hostreadiness"
	"github.com/woliveiras/corsarr/internal/onboarding"
	runtimeenv "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
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
	if len(app.ListLegalNotices()) < len(app.ListApplications()) {
		t.Fatal("expected legal notices for the application catalog and runtimes")
	}
	if err := app.OpenLegalLink("radarr", "https://attacker.example"); err == nil {
		t.Fatal("expected arbitrary legal URL to be rejected before reaching Wails runtime")
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

func TestSelectRecommendedApplicationsUsesBackendCatalogPreset(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create service registry: %v", err)
	}
	setup := &desktopSetupManager{}
	app := &App{catalog: application.NewCatalog(registry), setup: setup}

	status, err := app.SelectRecommendedApplications()
	if err != nil {
		t.Fatalf("select recommended applications: %v", err)
	}
	want := []string{
		"bazarr",
		"jellyfin",
		"jellyseerr",
		"prowlarr",
		"qbittorrent",
		"radarr",
		"sonarr",
	}
	if !reflect.DeepEqual(status.Applications, want) {
		t.Fatalf("expected backend preset %v, got %v", want, status.Applications)
	}
	if setup.saveApplicationsCalls != 1 {
		t.Fatalf("expected one persisted selection, got %d", setup.saveApplicationsCalls)
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
	runtime := &desktopRuntimePreparer{result: onboarding.PreparationResult{Ready: true}}
	app := &App{
		setup: &desktopSetupManager{status: application.SetupStatus{
			TermsAccepted: true, JellyfinLANEnabled: true,
		}},
		installation: installation, runtimeOnboarding: runtime,
	}

	result, err := app.InstallSelectedApplications()
	if err != nil {
		t.Fatalf("install selected applications: %v", err)
	}
	if !result.Complete || installation.calls != 1 || runtime.calls != 1 ||
		!installation.options.AllowJellyfinLAN {
		t.Fatalf("expected one completed installation call, got result=%#v calls=%d", result, installation.calls)
	}
}

func TestInstallSelectedApplicationsDoesNotPrepareWithoutConsent(t *testing.T) {
	runtime := &desktopRuntimePreparer{}
	installation := &desktopInstallationManager{}
	app := &App{
		setup: &desktopSetupManager{}, runtimeOnboarding: runtime, installation: installation,
	}

	if _, err := app.InstallSelectedApplications(); err == nil {
		t.Fatal("expected missing consent error")
	}
	if runtime.calls != 0 || installation.calls != 0 {
		t.Fatalf("unauthorized install accessed runtime=%d installer=%d", runtime.calls, installation.calls)
	}
}

func TestInstallSelectedApplicationsStopsWhenRuntimePreparationFails(t *testing.T) {
	runtime := &desktopRuntimePreparer{prepareErr: errors.New("runtime unavailable")}
	installation := &desktopInstallationManager{}
	app := &App{
		setup:             &desktopSetupManager{status: application.SetupStatus{TermsAccepted: true}},
		runtimeOnboarding: runtime, installation: installation,
	}

	if _, err := app.InstallSelectedApplications(); err == nil {
		t.Fatal("expected runtime preparation failure")
	}
	if runtime.calls != 1 || installation.calls != 0 {
		t.Fatalf("unexpected calls runtime=%d installer=%d", runtime.calls, installation.calls)
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

func TestUpdateApplicationUsesBoundedApplicationService(t *testing.T) {
	updates := &desktopUpdateManager{result: application.ApplicationUpdateResult{
		ApplicationID: "radarr", Updated: true,
	}}
	app := &App{setup: &desktopSetupManager{}, updates: updates}

	result, err := app.UpdateApplication("radarr")
	if err != nil {
		t.Fatalf("update application: %v", err)
	}
	if !result.Updated || updates.applicationID != "radarr" || updates.calls != 1 {
		t.Fatalf("expected one bounded update call, result=%#v manager=%#v", result, updates)
	}
}

func TestSetJellyfinLANRejectsInstalledContainer(t *testing.T) {
	setup := &desktopSetupManager{}
	management := &desktopApplicationManager{statuses: []application.ManagedApplicationStatus{{
		ApplicationID: "jellyfin", State: application.ManagedStateRunning,
	}}}
	app := &App{setup: setup, management: management}

	if _, err := app.SetJellyfinLAN(true); err == nil {
		t.Fatal("expected installed Jellyfin setting change to be rejected")
	}
	if setup.jellyfinLANCalls != 0 {
		t.Fatalf("rejected setting reached setup %d times", setup.jellyfinLANCalls)
	}
}

func TestSetJellyfinLANPersistsBeforeInstallation(t *testing.T) {
	setup := &desktopSetupManager{}
	management := &desktopApplicationManager{statuses: []application.ManagedApplicationStatus{{
		ApplicationID: "jellyfin", State: application.ManagedStateNotInstalled,
	}}}
	app := &App{setup: setup, management: management}

	status, err := app.SetJellyfinLAN(true)
	if err != nil || !status.JellyfinLANEnabled || setup.jellyfinLANCalls != 1 {
		t.Fatalf("unexpected LAN setting %#v, %v, calls=%d", status, err, setup.jellyfinLANCalls)
	}
}

func TestRuntimeOptionsPreserveHostProfileAndApplyReviewedNetworkChoice(t *testing.T) {
	options := runtimeOptions(runtimecatalog.RuntimeOptions{
		Timezone: "Europe/Madrid", PUID: 1001, PGID: 1002,
	}, application.SetupStatus{JellyfinLANEnabled: true})

	if options.Timezone != "Europe/Madrid" || options.PUID != 1001 || options.PGID != 1002 ||
		!options.AllowJellyfinLAN {
		t.Fatalf("unexpected runtime options %#v", options)
	}
}

func TestGetJellyfinNetworkStatusRequiresReviewedLANChoice(t *testing.T) {
	setup := &desktopSetupManager{status: application.SetupStatus{JellyfinLANEnabled: false}}
	provider := &desktopLocalNetwork{urls: []string{"http://192.168.1.42:8096"}}
	app := &App{setup: setup, localNetwork: provider}

	status, err := app.GetJellyfinNetworkStatus()
	if err != nil {
		t.Fatalf("get disabled Jellyfin network status: %v", err)
	}
	if status.Enabled || len(status.URLs) != 0 || provider.calls != 0 {
		t.Fatalf("unexpected disabled status %#v, calls=%d", status, provider.calls)
	}

	setup.status.JellyfinLANEnabled = true
	status, err = app.GetJellyfinNetworkStatus()
	if err != nil {
		t.Fatalf("get enabled Jellyfin network status: %v", err)
	}
	if !status.Enabled || len(status.URLs) != 1 || provider.calls != 1 {
		t.Fatalf("unexpected enabled status %#v, calls=%d", status, provider.calls)
	}
}

func TestCopyJellyfinNetworkURLAllowsOnlyCurrentDiscoveredAddress(t *testing.T) {
	clipboard := &desktopClipboard{}
	app := &App{
		setup:        &desktopSetupManager{status: application.SetupStatus{JellyfinLANEnabled: true}},
		localNetwork: &desktopLocalNetwork{urls: []string{"http://192.168.1.42:8096"}},
		clipboard:    clipboard,
	}

	if err := app.CopyJellyfinNetworkURL("https://attacker.example"); err == nil {
		t.Fatal("expected arbitrary network URL to be rejected")
	}
	if clipboard.calls != 0 {
		t.Fatal("clipboard received an untrusted URL")
	}
	if err := app.CopyJellyfinNetworkURL("http://192.168.1.42:8096"); err != nil {
		t.Fatalf("copy discovered URL: %v", err)
	}
	if clipboard.value != "http://192.168.1.42:8096" || clipboard.calls != 1 {
		t.Fatalf("unexpected clipboard state %#v", clipboard)
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

func TestCopyJellyfinPasswordWritesOnlyToNativeClipboard(t *testing.T) {
	access := &desktopServiceAccess{password: credentials.NewSecret("private-password")}
	clipboard := &desktopClipboard{}
	app := &App{serviceAccess: access, clipboard: clipboard}

	if err := app.CopyJellyfinPassword(); err != nil {
		t.Fatalf("copy Jellyfin password: %v", err)
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
		environment: application.NewEnvironmentService(probe, "darwin", "arm64", nil),
	}

	status := app.GetEnvironmentStatus()

	if status.Runtime.State != runtimeenv.StateStopped {
		t.Fatalf("expected stopped runtime, got %q", status.Runtime.State)
	}
	if probe.calls != 1 {
		t.Fatalf("expected one read-only probe, got %d", probe.calls)
	}
}

func TestPrepareRuntimeRequiresCurrentConsent(t *testing.T) {
	preparer := &desktopRuntimePreparer{}
	app := &App{setup: &desktopSetupManager{}, runtimeOnboarding: preparer}

	if _, err := app.PrepareRuntime(); err == nil {
		t.Fatal("expected runtime preparation without consent to be rejected")
	}
	if preparer.calls != 0 {
		t.Fatalf("expected no preparation, got %d calls", preparer.calls)
	}
}

func TestPrepareRuntimeRejectsUnsupportedHostBeforeMutation(t *testing.T) {
	preparer := &desktopRuntimePreparer{}
	app := &App{
		setup:             &desktopSetupManager{status: application.SetupStatus{TermsAccepted: true}},
		runtimeOnboarding: preparer,
		hostReadiness: &desktopHostReadiness{status: hostreadiness.Status{
			Ready: false, Issues: []string{"são necessários pelo menos 4 GiB de memória"},
		}},
	}

	if _, err := app.PrepareRuntime(); err == nil {
		t.Fatal("expected unsupported host to be rejected")
	}
	if preparer.calls != 0 {
		t.Fatalf("runtime preparation mutated an unsupported host, calls=%d", preparer.calls)
	}
}

func TestPrepareRuntimeUsesBoundedOnboardingAfterConsent(t *testing.T) {
	preparer := &desktopRuntimePreparer{result: onboarding.PreparationResult{Ready: true, Started: true}}
	app := &App{
		setup:             &desktopSetupManager{status: application.SetupStatus{TermsAccepted: true}},
		runtimeOnboarding: preparer,
	}

	result, err := app.PrepareRuntime()
	if err != nil || !result.Ready || preparer.calls != 1 {
		t.Fatalf("unexpected runtime preparation %#v, %v, calls=%d", result, err, preparer.calls)
	}
}

func TestSetStartAtLoginUsesSetupBoundary(t *testing.T) {
	setup := &desktopSetupManager{}
	app := &App{setup: setup}

	status, err := app.SetStartAtLogin(true)
	if err != nil {
		t.Fatalf("enable start at login: %v", err)
	}
	if !status.StartAtLogin || setup.startAtLoginCalls != 1 {
		t.Fatalf("unexpected start-at-login result %#v setup=%#v", status, setup)
	}
}

func TestBackgroundRecoveryStartsRuntimeBeforeExistingApplications(t *testing.T) {
	setup := &desktopSetupManager{status: application.SetupStatus{
		StartAtLogin: true, StartAtLoginSupported: true,
	}}
	runtime := &desktopRuntimePreparer{recoveryResult: onboarding.PreparationResult{Ready: true}}
	recovery := &desktopBackgroundRecovery{result: application.RecoveryResult{Complete: true}}
	events := &desktopEventPublisher{}
	app := &App{
		setup: setup, runtimeOnboarding: runtime, backgroundRecovery: recovery, events: events,
	}

	app.runBackgroundRecovery(context.Background())

	if runtime.recoverCalls != 1 || recovery.calls != 1 {
		t.Fatalf("unexpected recovery calls runtime=%d applications=%d", runtime.recoverCalls, recovery.calls)
	}
	if events.calls != 1 || events.name != backgroundRecoveryCompletedEvent {
		t.Fatalf("unexpected recovery event %#v", events)
	}
	event, ok := events.data[0].(BackgroundRecoveryEvent)
	if !ok || !event.Complete {
		t.Fatalf("unexpected recovery payload %#v", events.data)
	}
}

func TestBackgroundRecoveryDoesNothingWithoutEnabledNativeSetting(t *testing.T) {
	for _, status := range []application.SetupStatus{
		{},
		{StartAtLogin: true, StartAtLoginSupported: true, StartAtLoginRequiresApproval: true},
	} {
		runtime := &desktopRuntimePreparer{}
		recovery := &desktopBackgroundRecovery{}
		events := &desktopEventPublisher{}
		app := &App{
			setup:             &desktopSetupManager{status: status},
			runtimeOnboarding: runtime, backgroundRecovery: recovery, events: events,
		}

		app.runBackgroundRecovery(context.Background())

		if runtime.recoverCalls != 0 || recovery.calls != 0 || events.calls != 0 {
			t.Fatalf("disabled setting caused recovery for %#v", status)
		}
	}
}

func TestExportDiagnosticsWritesOnlyAfterNativeDestinationSelection(t *testing.T) {
	picker := &desktopDiagnosticPicker{path: "/Users/test/corsarr-diagnostics.json"}
	reporter := &desktopDiagnosticReporter{report: diagnostics.Report{
		SchemaVersion: diagnostics.CurrentSchemaVersion,
	}}
	writer := &desktopDiagnosticWriter{}
	app := &App{
		diagnosticPicker: picker, diagnosticReporter: reporter, diagnosticWriter: writer,
	}

	result, err := app.ExportDiagnostics()
	if err != nil {
		t.Fatalf("export diagnostics: %v", err)
	}
	if !result.Exported || result.Path != picker.path {
		t.Fatalf("unexpected export result %#v", result)
	}
	if reporter.calls != 1 || writer.calls != 1 || writer.path != picker.path {
		t.Fatalf("unexpected diagnostic flow reporter=%#v writer=%#v", reporter, writer)
	}
}

func TestExportDiagnosticsCancelDoesNotCollectOrWrite(t *testing.T) {
	reporter := &desktopDiagnosticReporter{}
	writer := &desktopDiagnosticWriter{}
	app := &App{
		diagnosticPicker:   &desktopDiagnosticPicker{},
		diagnosticReporter: reporter,
		diagnosticWriter:   writer,
	}

	result, err := app.ExportDiagnostics()
	if err != nil {
		t.Fatalf("cancel diagnostics: %v", err)
	}
	if result.Exported || reporter.calls != 0 || writer.calls != 0 {
		t.Fatalf("expected canceled export without collection, result=%#v", result)
	}
}

type desktopRuntimeProbe struct {
	status runtimeenv.Status
	calls  int
}

type desktopHostReadiness struct {
	status hostreadiness.Status
	calls  int
}

func (h *desktopHostReadiness) Check(context.Context) hostreadiness.Status {
	h.calls++
	return h.status
}

type desktopRuntimePreparer struct {
	result         onboarding.PreparationResult
	recoveryResult onboarding.PreparationResult
	recoveryErr    error
	prepareErr     error
	calls          int
	recoverCalls   int
}

func (p *desktopRuntimePreparer) Prepare(context.Context) (onboarding.PreparationResult, error) {
	p.calls++
	return p.result, p.prepareErr
}

func (p *desktopRuntimePreparer) Recover(context.Context) (onboarding.PreparationResult, error) {
	p.recoverCalls++
	return p.recoveryResult, p.recoveryErr
}

func (f *desktopRuntimeProbe) Check(context.Context) runtimeenv.Status {
	f.calls++
	return f.status
}

type desktopDirectoryPicker struct {
	path string
	err  error
}

type desktopDiagnosticPicker struct {
	path string
	err  error
}

func (p *desktopDiagnosticPicker) Choose(context.Context, string) (string, error) {
	return p.path, p.err
}

type desktopDiagnosticReporter struct {
	report diagnostics.Report
	err    error
	calls  int
}

func (r *desktopDiagnosticReporter) Build(context.Context) (diagnostics.Report, error) {
	r.calls++
	return r.report, r.err
}

type desktopDiagnosticWriter struct {
	path   string
	report diagnostics.Report
	err    error
	calls  int
}

type desktopBackgroundRecovery struct {
	result application.RecoveryResult
	err    error
	calls  int
}

func (r *desktopBackgroundRecovery) Recover(context.Context) (application.RecoveryResult, error) {
	r.calls++
	return r.result, r.err
}

type desktopEventPublisher struct {
	name  string
	data  []interface{}
	calls int
}

func (p *desktopEventPublisher) Emit(_ context.Context, name string, data ...interface{}) {
	p.calls++
	p.name = name
	p.data = data
}

func (w *desktopDiagnosticWriter) Write(path string, report diagnostics.Report) error {
	w.calls++
	w.path = path
	w.report = report
	return w.err
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
	status                application.SetupStatus
	savedStorage          string
	saveApplicationsCalls int
	startAtLoginCalls     int
	jellyfinLANCalls      int
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
	f.saveApplicationsCalls++
	f.status.Applications = applicationIDs
	return f.status, nil
}

func (f *desktopSetupManager) AcceptCurrentTerms() (application.SetupStatus, error) {
	f.status.TermsAccepted = true
	f.status.CanInstall = f.status.CanPrepare
	return f.status, nil
}

func (f *desktopSetupManager) SetStartAtLogin(enabled bool) (application.SetupStatus, error) {
	f.startAtLoginCalls++
	f.status.StartAtLogin = enabled
	return f.status, nil
}

func (f *desktopSetupManager) OpenStartAtLoginSettings() error { return nil }

func (f *desktopSetupManager) SetJellyfinLAN(enabled bool) (application.SetupStatus, error) {
	f.jellyfinLANCalls++
	f.status.JellyfinLANEnabled = enabled
	return f.status, nil
}

type desktopLayoutPreparer struct {
	status         storage.LayoutStatus
	basePath       string
	applicationIDs []string
	prepareCalls   int
}

type desktopInstallationManager struct {
	result  application.InstallationResult
	calls   int
	options runtimecatalog.RuntimeOptions
}

type desktopApplicationDataManager struct {
	result        storage.ArchivedApplicationData
	applicationID string
	calls         int
}

type desktopApplicationManager struct {
	statuses []application.ManagedApplicationStatus
}

func (m *desktopApplicationManager) ListStatuses(context.Context) []application.ManagedApplicationStatus {
	return m.statuses
}

func (m *desktopApplicationManager) Start(context.Context, string) error   { return nil }
func (m *desktopApplicationManager) Stop(context.Context, string) error    { return nil }
func (m *desktopApplicationManager) Restart(context.Context, string) error { return nil }
func (m *desktopApplicationManager) Remove(context.Context, string) error  { return nil }

type desktopUpdateManager struct {
	result        application.ApplicationUpdateResult
	applicationID string
	calls         int
}

func (m *desktopUpdateManager) Update(
	_ context.Context,
	applicationID string,
	_ runtimecatalog.RuntimeOptions,
) (application.ApplicationUpdateResult, error) {
	m.calls++
	m.applicationID = applicationID
	return m.result, nil
}

type desktopServiceAccess struct {
	password credentials.Secret
}

type desktopLocalNetwork struct {
	urls  []string
	calls int
}

func (n *desktopLocalNetwork) HTTPURLs(int) []string {
	n.calls++
	return n.urls
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

func (a *desktopServiceAccess) JellyfinStatus(context.Context) (application.ServiceAccessStatus, error) {
	return application.ServiceAccessStatus{
		ApplicationID: "jellyfin",
		Username:      "corsarr",
		Available:     true,
	}, nil
}

func (a *desktopServiceAccess) JellyfinPassword(context.Context) (credentials.Secret, error) {
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
	_ context.Context,
	options runtimecatalog.RuntimeOptions,
) (application.InstallationResult, error) {
	m.calls++
	m.options = options
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
