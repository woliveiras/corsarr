package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	goruntime "runtime"
	"runtime/debug"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/woliveiras/corsarr/internal/application"
	"github.com/woliveiras/corsarr/internal/autostart"
	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/credentials"
	"github.com/woliveiras/corsarr/internal/diagnostics"
	"github.com/woliveiras/corsarr/internal/legal"
	"github.com/woliveiras/corsarr/internal/onboarding"
	"github.com/woliveiras/corsarr/internal/orchestrator"
	"github.com/woliveiras/corsarr/internal/provisioning"
	runtimeenv "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
	statefile "github.com/woliveiras/corsarr/internal/state"
	"github.com/woliveiras/corsarr/internal/storage"
)

type directoryPicker interface {
	Choose(ctx context.Context) (string, error)
}

type storageInspector interface {
	Inspect(path string) storage.Status
}

type setupManager interface {
	Load() (application.SetupStatus, error)
	SaveStorage(path string) (application.SetupStatus, error)
	SaveApplications(applicationIDs []string) (application.SetupStatus, error)
	AcceptCurrentTerms() (application.SetupStatus, error)
	SetStartAtLogin(enabled bool) (application.SetupStatus, error)
	OpenStartAtLoginSettings() error
}

type storageLayoutPreparer interface {
	Prepare(basePath string, applicationIDs []string) (storage.LayoutStatus, error)
}

type installationManager interface {
	InstallSelected(
		ctx context.Context,
		options runtimecatalog.RuntimeOptions,
	) (application.InstallationResult, error)
}

type applicationManager interface {
	ListStatuses(ctx context.Context) []application.ManagedApplicationStatus
	Start(ctx context.Context, applicationID string) error
	Stop(ctx context.Context, applicationID string) error
	Restart(ctx context.Context, applicationID string) error
	Remove(ctx context.Context, applicationID string) error
}

type applicationUpdateManager interface {
	Update(
		ctx context.Context,
		applicationID string,
		options runtimecatalog.RuntimeOptions,
	) (application.ApplicationUpdateResult, error)
}

type applicationDataManager interface {
	ListStatuses() ([]storage.ApplicationDataStatus, error)
	Archive(ctx context.Context, applicationID string) (storage.ArchivedApplicationData, error)
}

type serviceAccessManager interface {
	JellyfinStatus(ctx context.Context) (application.ServiceAccessStatus, error)
	JellyfinPassword(ctx context.Context) (credentials.Secret, error)
	QBittorrentStatus(ctx context.Context) (application.ServiceAccessStatus, error)
	QBittorrentPassword(ctx context.Context) (credentials.Secret, error)
}

type clipboardWriter interface {
	SetText(ctx context.Context, value string) error
}

type runtimePreparer interface {
	Prepare(ctx context.Context) (onboarding.PreparationResult, error)
}

type diagnosticFilePicker interface {
	Choose(ctx context.Context, suggestedName string) (string, error)
}

type diagnosticReporter interface {
	Build(ctx context.Context) (diagnostics.Report, error)
}

type diagnosticWriter interface {
	Write(path string, report diagnostics.Report) error
}

type DiagnosticExportResult struct {
	Exported bool   `json:"exported"`
	Path     string `json:"path,omitempty"`
}

type wailsClipboard struct{}

func (wailsClipboard) SetText(ctx context.Context, value string) error {
	return wailsruntime.ClipboardSetText(ctx, value)
}

// App is the narrow bridge between the desktop UI and Corsarr's application layer.
type App struct {
	ctx                context.Context
	catalog            *application.Catalog
	legal              *legal.Catalog
	environment        *application.EnvironmentService
	directoryPicker    directoryPicker
	storageInspector   storageInspector
	setup              setupManager
	layoutPreparer     storageLayoutPreparer
	installation       installationManager
	management         applicationManager
	updates            applicationUpdateManager
	runtimeOnboarding  runtimePreparer
	applicationData    applicationDataManager
	serviceAccess      serviceAccessManager
	clipboard          clipboardWriter
	diagnosticPicker   diagnosticFilePicker
	diagnosticReporter diagnosticReporter
	diagnosticWriter   diagnosticWriter
}

func NewApp() (*App, error) {
	registry, err := services.NewRegistry()
	if err != nil {
		return nil, fmt.Errorf("create service registry: %w", err)
	}

	catalog := application.NewCatalog(registry)
	dockerDetector := runtimeenv.NewDockerDetector(runtimeenv.OSCommandRunner{}, 5*time.Second)
	environment := application.NewEnvironmentService(
		dockerDetector,
		goruntime.GOOS,
		goruntime.GOARCH,
	)
	statePath, err := statefile.DefaultPath()
	if err != nil {
		return nil, err
	}
	setup := application.NewSetupService(
		catalog,
		statefile.NewFileStore(statePath),
		autostart.NewPlatformManager(goruntime.GOOS),
	)
	runtimeOnboarding, err := newRuntimeOnboarding(dockerDetector)
	if err != nil {
		return nil, fmt.Errorf("create runtime onboarding: %w", err)
	}
	approvedCatalog, err := runtimecatalog.NewRuntimeCatalog(registry)
	if err != nil {
		return nil, fmt.Errorf("create approved runtime catalog: %w", err)
	}
	legalCatalog, err := legal.NewCatalog(registry, approvedCatalog)
	if err != nil {
		return nil, fmt.Errorf("create legal catalog: %w", err)
	}
	dockerManager := runtimeenv.NewDockerManager(runtimeenv.OSCommandRunner{}, 10*time.Minute)
	readiness := provisioning.NewHTTPReadiness(catalog, 2*time.Minute, time.Second)
	installer := orchestrator.NewInstaller(dockerManager, approvedCatalog, readiness)
	updater := orchestrator.NewUpdater(
		dockerManager,
		approvedCatalog,
		readiness,
		storage.NewBackupManager(),
	)
	arrCredentials := provisioning.NewARRCredentialReader()
	arrClient := provisioning.NewARRClient(catalog)
	arrProvisioner := provisioning.NewARRProvisioner(arrCredentials, arrClient)
	credentialStore := credentials.NewPlatformStore()
	qbittorrentProvisioner := provisioning.NewQBittorrentProvisioner(
		dockerManager,
		credentialStore,
		provisioning.NewQBittorrentClient(catalog),
	)
	arrDownloadProvisioner := provisioning.NewARRDownloadClientProvisioner(
		arrCredentials,
		credentialStore,
		arrClient,
	)
	prowlarrProvisioner := provisioning.NewProwlarrProvisioner(
		arrCredentials,
		provisioning.NewProwlarrClient(catalog),
	)
	bazarrProvisioner := provisioning.NewBazarrProvisioner(
		provisioning.NewBazarrCredentialReader(),
		arrCredentials,
		provisioning.NewBazarrClient(catalog),
	)
	jellyfinProvisioner := provisioning.NewJellyfinProvisioner(
		credentialStore,
		provisioning.NewJellyfinClient(catalog),
	)
	seerrProvisioner := provisioning.NewSeerrProvisioner(
		credentialStore,
		arrCredentials,
		provisioning.NewSeerrClient(catalog),
	)
	provisioner := provisioning.NewChainProvisioner(
		arrProvisioner,
		qbittorrentProvisioner,
		arrDownloadProvisioner,
		prowlarrProvisioner,
		bazarrProvisioner,
		jellyfinProvisioner,
		seerrProvisioner,
	)
	installation := application.NewInstallationService(
		setup,
		storage.NewLayoutPreparer(),
		catalog,
		installer,
		provisioner,
	)
	management := application.NewManagementService(catalog, dockerManager, approvedCatalog)
	updates := application.NewUpdateService(setup, catalog, updater, provisioner)
	applicationData := application.NewDataManagementService(
		catalog,
		setup,
		dockerManager,
		storage.NewApplicationDataManager(),
	)
	storageInspector := storage.NewInspector()
	diagnosticReporter := diagnostics.NewReporter(
		environment,
		setup,
		management,
		storageInspector,
		corsarrBuildVersion(),
		runtimecatalog.RuntimeCatalogVerifiedAt,
	)

	return &App{
		catalog:            catalog,
		legal:              legalCatalog,
		environment:        environment,
		directoryPicker:    wailsDirectoryPicker{},
		storageInspector:   storageInspector,
		setup:              setup,
		layoutPreparer:     storage.NewLayoutPreparer(),
		installation:       installation,
		management:         management,
		updates:            updates,
		runtimeOnboarding:  runtimeOnboarding,
		applicationData:    applicationData,
		serviceAccess:      application.NewServiceAccess(credentialStore),
		clipboard:          wailsClipboard{},
		diagnosticPicker:   wailsDiagnosticFilePicker{},
		diagnosticReporter: diagnosticReporter,
		diagnosticWriter:   diagnostics.NewFileWriter(),
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ListApplications returns the user-facing applications known by Corsarr.
func (a *App) ListApplications() []application.ApplicationSummary {
	return a.catalog.ListApplications()
}

func (a *App) ListLegalNotices() []legal.Notice {
	return a.legal.ListNotices()
}

// OpenLegalLink resolves an allowlisted link by component and semantic kind.
// The frontend cannot submit a URL.
func (a *App) OpenLegalLink(componentID, kind string) error {
	link, err := a.legal.ResolveLink(componentID, kind)
	if err != nil {
		return err
	}
	wailsruntime.BrowserOpenURL(a.ctx, link)
	return nil
}

// GetEnvironmentStatus performs a bounded, read-only host and runtime check.
func (a *App) GetEnvironmentStatus() application.EnvironmentStatus {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.environment.Status(ctx)
}

func (a *App) PrepareRuntime() (onboarding.PreparationResult, error) {
	setup, err := a.setup.Load()
	if err != nil {
		return onboarding.PreparationResult{}, err
	}
	if !setup.TermsAccepted {
		return onboarding.PreparationResult{}, application.ErrTermsNotAccepted
	}
	return a.runtimeOnboarding.Prepare(a.appContext())
}

// ExportDiagnostics writes a redacted, log-free snapshot only to the path
// explicitly selected in the native save dialog.
func (a *App) ExportDiagnostics() (DiagnosticExportResult, error) {
	destination, err := a.diagnosticPicker.Choose(
		a.appContext(),
		"corsarr-diagnostics.json",
	)
	if err != nil {
		return DiagnosticExportResult{}, fmt.Errorf("choose diagnostic destination: %w", err)
	}
	if destination == "" {
		return DiagnosticExportResult{}, nil
	}
	report, err := a.diagnosticReporter.Build(a.appContext())
	if err != nil {
		return DiagnosticExportResult{}, err
	}
	if err := a.diagnosticWriter.Write(destination, report); err != nil {
		return DiagnosticExportResult{}, err
	}
	return DiagnosticExportResult{Exported: true, Path: destination}, nil
}

// ChooseStorageLocation opens the native picker and inspects only the selected directory.
func (a *App) ChooseStorageLocation() (storage.Status, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	selectedPath, err := a.directoryPicker.Choose(ctx)
	if err != nil {
		return storage.Status{}, fmt.Errorf("choose storage location: %w", err)
	}
	if selectedPath == "" {
		return storage.Status{State: storage.StateCanceled}, nil
	}

	storageStatus := a.storageInspector.Inspect(selectedPath)
	if storageStatus.State != storage.StateReady {
		return storageStatus, nil
	}
	if _, err := a.setup.SaveStorage(storageStatus.Path); err != nil {
		return storage.Status{}, err
	}
	return storageStatus, nil
}

func (a *App) GetSetupStatus() (application.SetupStatus, error) {
	return a.setup.Load()
}

func (a *App) SaveApplicationSelection(applicationIDs []string) (application.SetupStatus, error) {
	return a.setup.SaveApplications(applicationIDs)
}

func (a *App) AcceptCurrentTerms() (application.SetupStatus, error) {
	return a.setup.AcceptCurrentTerms()
}

func (a *App) SetStartAtLogin(enabled bool) (application.SetupStatus, error) {
	return a.setup.SetStartAtLogin(enabled)
}

func (a *App) OpenStartAtLoginSettings() error {
	return a.setup.OpenStartAtLoginSettings()
}

func (a *App) InstallSelectedApplications() (application.InstallationResult, error) {
	return a.installation.InstallSelected(
		a.appContext(),
		runtimecatalog.RuntimeOptions{PUID: 1000, PGID: 1000},
	)
}

func (a *App) GetApplicationStatuses() []application.ManagedApplicationStatus {
	return a.management.ListStatuses(a.appContext())
}

func (a *App) StartApplication(id string) error {
	return a.management.Start(a.appContext(), id)
}

func (a *App) StopApplication(id string) error {
	return a.management.Stop(a.appContext(), id)
}

func (a *App) RestartApplication(id string) error {
	return a.management.Restart(a.appContext(), id)
}

func (a *App) RemoveApplication(id string) error {
	return a.management.Remove(a.appContext(), id)
}

func (a *App) UpdateApplication(id string) (application.ApplicationUpdateResult, error) {
	return a.updates.Update(
		a.appContext(),
		id,
		runtimecatalog.RuntimeOptions{PUID: 1000, PGID: 1000},
	)
}

func (a *App) GetApplicationDataStatuses() ([]storage.ApplicationDataStatus, error) {
	return a.applicationData.ListStatuses()
}

func (a *App) ArchiveApplicationData(id string) (storage.ArchivedApplicationData, error) {
	return a.applicationData.Archive(a.appContext(), id)
}

func (a *App) GetQBittorrentAccessStatus() (application.ServiceAccessStatus, error) {
	return a.serviceAccess.QBittorrentStatus(a.appContext())
}

// CopyQBittorrentPassword reveals the secret only to the native clipboard after
// an explicit user action. The value is never returned through Wails.
func (a *App) CopyQBittorrentPassword() error {
	secret, err := a.serviceAccess.QBittorrentPassword(a.appContext())
	if err != nil {
		return err
	}
	return a.clipboard.SetText(a.appContext(), secret.Reveal())
}

func (a *App) GetJellyfinAccessStatus() (application.ServiceAccessStatus, error) {
	return a.serviceAccess.JellyfinStatus(a.appContext())
}

// CopyJellyfinPassword reveals the secret only to the native clipboard after
// an explicit user action. The value is never returned through Wails.
func (a *App) CopyJellyfinPassword() error {
	secret, err := a.serviceAccess.JellyfinPassword(a.appContext())
	if err != nil {
		return err
	}
	return a.clipboard.SetText(a.appContext(), secret.Reveal())
}

func (a *App) appContext() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

// PrepareStorageLayout creates only the reviewed Corsarr-owned directory tree.
func (a *App) PrepareStorageLayout() (storage.LayoutStatus, error) {
	setupStatus, err := a.setup.Load()
	if err != nil {
		return storage.LayoutStatus{}, err
	}
	if !setupStatus.CanPrepare {
		return storage.LayoutStatus{}, fmt.Errorf(
			"storage and at least one application must be selected before preparation",
		)
	}
	return a.layoutPreparer.Prepare(setupStatus.StoragePath, setupStatus.Applications)
}

// OpenApplication opens only a local URL resolved from Corsarr's catalog.
func (a *App) OpenApplication(id string) error {
	applicationURL, err := a.catalog.ResolveApplicationURL(id)
	if err != nil {
		return err
	}

	wailsruntime.BrowserOpenURL(a.ctx, applicationURL)
	return nil
}

func newRuntimeOnboarding(probe runtimeenv.Probe) (runtimePreparer, error) {
	runner := runtimeenv.OSCommandRunner{}
	if goruntime.GOOS != "darwin" {
		return onboarding.NewDockerService(probe, runner, nil, goruntime.GOOS), nil
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	installer, err := onboarding.NewMacDockerInstaller(
		runner,
		onboarding.NewHTTPDownloader(),
		goruntime.GOARCH,
		filepath.Join(cacheRoot, "Corsarr", "installers"),
		currentUser.Username,
	)
	if err != nil {
		return nil, err
	}
	return onboarding.NewDockerService(probe, runner, installer, goruntime.GOOS), nil
}

func corsarrBuildVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok || buildInfo.Main.Version == "" || buildInfo.Main.Version == "(devel)" {
		return "development"
	}
	return buildInfo.Main.Version
}
