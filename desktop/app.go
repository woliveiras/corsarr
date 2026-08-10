package main

import (
	"context"
	"fmt"
	goruntime "runtime"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/woliveiras/corsarr/internal/application"
	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	"github.com/woliveiras/corsarr/internal/credentials"
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

type wailsClipboard struct{}

func (wailsClipboard) SetText(ctx context.Context, value string) error {
	return wailsruntime.ClipboardSetText(ctx, value)
}

// App is the narrow bridge between the desktop UI and Corsarr's application layer.
type App struct {
	ctx              context.Context
	catalog          *application.Catalog
	environment      *application.EnvironmentService
	directoryPicker  directoryPicker
	storageInspector storageInspector
	setup            setupManager
	layoutPreparer   storageLayoutPreparer
	installation     installationManager
	management       applicationManager
	applicationData  applicationDataManager
	serviceAccess    serviceAccessManager
	clipboard        clipboardWriter
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
	setup := application.NewSetupService(catalog, statefile.NewFileStore(statePath))
	approvedCatalog, err := runtimecatalog.NewRuntimeCatalog(registry)
	if err != nil {
		return nil, fmt.Errorf("create approved runtime catalog: %w", err)
	}
	dockerManager := runtimeenv.NewDockerManager(runtimeenv.OSCommandRunner{}, 10*time.Minute)
	readiness := provisioning.NewHTTPReadiness(catalog, 2*time.Minute, time.Second)
	installer := orchestrator.NewInstaller(dockerManager, approvedCatalog, readiness)
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
	management := application.NewManagementService(catalog, dockerManager)
	applicationData := application.NewDataManagementService(
		catalog,
		setup,
		dockerManager,
		storage.NewApplicationDataManager(),
	)

	return &App{
		catalog:          catalog,
		environment:      environment,
		directoryPicker:  wailsDirectoryPicker{},
		storageInspector: storage.NewInspector(),
		setup:            setup,
		layoutPreparer:   storage.NewLayoutPreparer(),
		installation:     installation,
		management:       management,
		applicationData:  applicationData,
		serviceAccess:    application.NewServiceAccess(credentialStore),
		clipboard:        wailsClipboard{},
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ListApplications returns the user-facing applications known by Corsarr.
func (a *App) ListApplications() []application.ApplicationSummary {
	return a.catalog.ListApplications()
}

// GetEnvironmentStatus performs a bounded, read-only host and runtime check.
func (a *App) GetEnvironmentStatus() application.EnvironmentStatus {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.environment.Status(ctx)
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
