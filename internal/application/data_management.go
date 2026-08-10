package application

import (
	"context"
	"errors"
	"fmt"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

var ErrApplicationStillInstalled = errors.New("remove the application before removing its data")

type ApplicationDataArchiver interface {
	Archive(basePath string, applicationID string) (storage.ArchivedApplicationData, error)
	Inspect(basePath string, applicationID string) (storage.ApplicationDataStatus, error)
}

func (s *DataManagementService) ListStatuses() ([]storage.ApplicationDataStatus, error) {
	setup, err := s.setup.Load()
	if err != nil {
		return nil, fmt.Errorf("load reviewed setup: %w", err)
	}
	applications := s.catalog.ListApplications()
	statuses := make([]storage.ApplicationDataStatus, 0, len(applications))
	for _, application := range applications {
		status := storage.ApplicationDataStatus{ApplicationID: application.ID}
		if setup.StoragePath != "" {
			status, err = s.archiver.Inspect(setup.StoragePath, application.ID)
			if err != nil {
				return nil, fmt.Errorf("inspect %s data: %w", application.ID, err)
			}
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

type DataManagementService struct {
	catalog  *Catalog
	setup    InstallationSetup
	runtime  containerruntime.Manager
	archiver ApplicationDataArchiver
}

func NewDataManagementService(
	catalog *Catalog,
	setup InstallationSetup,
	runtime containerruntime.Manager,
	archiver ApplicationDataArchiver,
) *DataManagementService {
	return &DataManagementService{
		catalog:  catalog,
		setup:    setup,
		runtime:  runtime,
		archiver: archiver,
	}
}

// Archive moves one removed application's config to Corsarr's recoverable trash.
// It refuses to write while the corresponding runtime container still exists.
func (s *DataManagementService) Archive(
	ctx context.Context,
	applicationID string,
) (storage.ArchivedApplicationData, error) {
	if _, exists := s.catalog.byID[applicationID]; !exists {
		return storage.ArchivedApplicationData{}, fmt.Errorf(
			"application is not available in the desktop catalog: %s",
			applicationID,
		)
	}

	if _, err := s.runtime.Inspect(ctx, applicationID); err == nil {
		return storage.ArchivedApplicationData{}, ErrApplicationStillInstalled
	} else if !errors.Is(err, containerruntime.ErrResourceNotFound) {
		return storage.ArchivedApplicationData{}, fmt.Errorf("verify application removal: %w", err)
	}

	setup, err := s.setup.Load()
	if err != nil {
		return storage.ArchivedApplicationData{}, fmt.Errorf("load reviewed setup: %w", err)
	}
	if setup.StoragePath == "" {
		return storage.ArchivedApplicationData{}, fmt.Errorf("storage location has not been selected")
	}

	archived, err := s.archiver.Archive(setup.StoragePath, applicationID)
	if err != nil {
		return storage.ArchivedApplicationData{}, fmt.Errorf("archive application data: %w", err)
	}
	return archived, nil
}
