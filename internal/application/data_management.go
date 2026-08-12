package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/woliveiras/corsarr/internal/credentials"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

var ErrApplicationStillInstalled = errors.New("remove the application before removing its data")

type ApplicationDataArchiver interface {
	Archive(basePath string, applicationID string) (storage.ArchivedApplicationData, error)
	Inspect(basePath string, applicationID string) (storage.ApplicationDataStatus, error)
	Restore(basePath, applicationID, archivePath string) error
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
	secrets  credentials.Store
}

func NewDataManagementService(
	catalog *Catalog,
	setup InstallationSetup,
	runtime containerruntime.Manager,
	archiver ApplicationDataArchiver,
	credentialStores ...credentials.Store,
) *DataManagementService {
	var secrets credentials.Store
	if len(credentialStores) > 0 {
		secrets = credentialStores[0]
	}
	return &DataManagementService{
		catalog: catalog, setup: setup, runtime: runtime, archiver: archiver, secrets: secrets,
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
	if err := s.deleteApplicationCredentials(ctx, applicationID); err != nil {
		if !archived.Archived {
			return archived, err
		}
		restoreErr := s.archiver.Restore(
			setup.StoragePath,
			applicationID,
			archived.ArchivePath,
		)
		if restoreErr == nil {
			archived.Archived = false
			archived.ArchivePath = ""
		}
		return archived, errors.Join(err, restoreErr)
	}
	return archived, nil
}

func (s *DataManagementService) deleteApplicationCredentials(
	ctx context.Context,
	applicationID string,
) error {
	if s.secrets == nil {
		return nil
	}

	var keys []credentials.Key
	switch applicationID {
	case "jellyfin":
		keys = []credentials.Key{credentials.KeyJellyfinPassword}
	case "lazylibrarian":
		keys = []credentials.Key{
			credentials.KeyLazyLibrarianPassword,
			credentials.KeyLazyLibrarianAPIKey,
		}
	case "qbittorrent":
		keys = []credentials.Key{credentials.KeyQBitTorrentPassword}
	default:
		if key, err := credentials.ARRPasswordKey(applicationID); err == nil {
			keys = []credentials.Key{key}
		}
	}
	stored := make(map[credentials.Key]credentials.Secret, len(keys))
	for _, key := range keys {
		secret, err := s.secrets.Load(ctx, key)
		if errors.Is(err, credentials.ErrCredentialNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("load archived %s credential before removal: %w", applicationID, err)
		}
		stored[key] = secret
	}
	deleted := make([]credentials.Key, 0, len(keys))
	for _, key := range keys {
		if err := s.secrets.Delete(ctx, key); err != nil {
			rollbackErrors := make([]error, 0, len(deleted))
			for _, deletedKey := range deleted {
				secret, existed := stored[deletedKey]
				if !existed {
					continue
				}
				if saveErr := s.secrets.Save(ctx, deletedKey, secret); saveErr != nil {
					rollbackErrors = append(
						rollbackErrors,
						fmt.Errorf("restore %s credential: %w", applicationID, saveErr),
					)
				}
			}
			return errors.Join(
				fmt.Errorf("remove archived %s credential: %w", applicationID, err),
				errors.Join(rollbackErrors...),
			)
		}
		deleted = append(deleted, key)
	}
	return nil
}
