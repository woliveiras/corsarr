package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/woliveiras/corsarr/internal/application"
	"github.com/woliveiras/corsarr/internal/services"
)

// App is the narrow bridge between the desktop UI and Corsarr's application layer.
type App struct {
	ctx     context.Context
	catalog *application.Catalog
}

func NewApp() (*App, error) {
	registry, err := services.NewRegistry()
	if err != nil {
		return nil, fmt.Errorf("create service registry: %w", err)
	}

	return &App{catalog: application.NewCatalog(registry)}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ListApplications returns the user-facing applications known by Corsarr.
func (a *App) ListApplications() []application.ApplicationSummary {
	return a.catalog.ListApplications()
}

// OpenApplication opens only a local URL resolved from Corsarr's catalog.
func (a *App) OpenApplication(id string) error {
	applicationURL, err := a.catalog.ResolveApplicationURL(id)
	if err != nil {
		return err
	}

	runtime.BrowserOpenURL(a.ctx, applicationURL)
	return nil
}
