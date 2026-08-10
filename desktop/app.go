package main

import (
	"context"
	"fmt"
	goruntime "runtime"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/woliveiras/corsarr/internal/application"
	runtimeenv "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
	"github.com/woliveiras/corsarr/internal/storage"
)

type directoryPicker interface {
	Choose(ctx context.Context) (string, error)
}

type storageInspector interface {
	Inspect(path string) storage.Status
}

// App is the narrow bridge between the desktop UI and Corsarr's application layer.
type App struct {
	ctx              context.Context
	catalog          *application.Catalog
	environment      *application.EnvironmentService
	directoryPicker  directoryPicker
	storageInspector storageInspector
}

func NewApp() (*App, error) {
	registry, err := services.NewRegistry()
	if err != nil {
		return nil, fmt.Errorf("create service registry: %w", err)
	}

	dockerDetector := runtimeenv.NewDockerDetector(runtimeenv.OSCommandRunner{}, 5*time.Second)
	environment := application.NewEnvironmentService(
		dockerDetector,
		goruntime.GOOS,
		goruntime.GOARCH,
	)

	return &App{
		catalog:          application.NewCatalog(registry),
		environment:      environment,
		directoryPicker:  wailsDirectoryPicker{},
		storageInspector: storage.NewInspector(),
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

	return a.storageInspector.Inspect(selectedPath), nil
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
