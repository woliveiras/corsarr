package application

import (
	"context"
	"errors"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/services"
	"github.com/woliveiras/corsarr/internal/storage"
)

func TestDataManagementServiceRequiresContainerRemovalBeforeArchive(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	runtime := &managementRuntime{statuses: map[string]containerruntime.ContainerStatus{
		"radarr": {ApplicationID: "radarr", State: containerruntime.ContainerStateStopped},
	}}
	archiver := &dataArchiver{}
	service := NewDataManagementService(
		NewCatalog(registry),
		&dataSetup{status: SetupStatus{StoragePath: "/media"}},
		runtime,
		archiver,
	)

	_, err = service.Archive(context.Background(), "radarr")
	if !errors.Is(err, ErrApplicationStillInstalled) {
		t.Fatalf("expected installed application rejection, got %v", err)
	}
	if archiver.calls != 0 {
		t.Fatalf("expected no archive write, got %d", archiver.calls)
	}
}

func TestDataManagementServiceArchivesKnownRemovedApplication(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	archiver := &dataArchiver{result: storage.ArchivedApplicationData{
		ApplicationID: "radarr",
		Archived:      true,
		ArchivePath:   "/media/Corsarr/trash/config/radarr/backup",
	}}
	service := NewDataManagementService(
		NewCatalog(registry),
		&dataSetup{status: SetupStatus{StoragePath: "/media"}},
		&managementRuntime{},
		archiver,
	)

	result, err := service.Archive(context.Background(), "radarr")
	if err != nil {
		t.Fatalf("archive removed application config: %v", err)
	}
	if !result.Archived || archiver.basePath != "/media" || archiver.applicationID != "radarr" {
		t.Fatalf("unexpected archive call/result: result=%#v archiver=%#v", result, archiver)
	}
}

func TestDataManagementServiceRemovesCredentialAfterArchivingConfiguration(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	archiver := &dataArchiver{result: storage.ArchivedApplicationData{
		ApplicationID: "qbittorrent", Archived: true,
	}}
	secrets := &dataCredentialStore{archiver: archiver}
	service := NewDataManagementService(
		NewCatalog(registry),
		&dataSetup{status: SetupStatus{StoragePath: "/media"}},
		&managementRuntime{},
		archiver,
		secrets,
	)

	if _, err := service.Archive(context.Background(), "qbittorrent"); err != nil {
		t.Fatalf("archive qBittorrent data: %v", err)
	}
	if len(secrets.deleted) != 1 || secrets.deleted[0] != credentials.KeyQBitTorrentPassword {
		t.Fatalf("expected only qBittorrent credential removal, got %v", secrets.deleted)
	}
	if secrets.deletedBeforeArchive {
		t.Fatal("credential was removed before configuration became recoverable")
	}
}

func TestDataManagementServiceRemovesArrCredentialAfterArchivingConfiguration(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	archiver := &dataArchiver{result: storage.ArchivedApplicationData{
		ApplicationID: "radarr", Archived: true,
	}}
	secrets := &dataCredentialStore{archiver: archiver}
	service := NewDataManagementService(
		NewCatalog(registry),
		&dataSetup{status: SetupStatus{StoragePath: "/media"}},
		&managementRuntime{},
		archiver,
		secrets,
	)

	if _, err := service.Archive(context.Background(), "radarr"); err != nil {
		t.Fatalf("archive Radarr data: %v", err)
	}
	if len(secrets.deleted) != 1 || secrets.deleted[0] != credentials.KeyRadarrPassword {
		t.Fatalf("expected only Radarr credential removal, got %v", secrets.deleted)
	}
}

func TestDataManagementServicePreservesCredentialWhenArchiveFails(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	archiver := &dataArchiver{err: errors.New("archive failed")}
	secrets := &dataCredentialStore{archiver: archiver}
	service := NewDataManagementService(
		NewCatalog(registry),
		&dataSetup{status: SetupStatus{StoragePath: "/media"}},
		&managementRuntime{},
		archiver,
		secrets,
	)

	if _, err := service.Archive(context.Background(), "jellyfin"); err == nil {
		t.Fatal("expected archive failure")
	}
	if len(secrets.deleted) != 0 {
		t.Fatalf("archive failure removed credentials: %v", secrets.deleted)
	}
}

func TestDataManagementServiceRestoresConfigurationWhenCredentialRemovalFails(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	archiver := &dataArchiver{result: storage.ArchivedApplicationData{
		ApplicationID: "jellyfin", Archived: true, ArchivePath: "/media/archive",
	}}
	secrets := &dataCredentialStore{archiver: archiver, deleteErr: errors.New("keychain locked")}
	service := NewDataManagementService(
		NewCatalog(registry),
		&dataSetup{status: SetupStatus{StoragePath: "/media"}},
		&managementRuntime{},
		archiver,
		secrets,
	)

	result, err := service.Archive(context.Background(), "jellyfin")
	if err == nil {
		t.Fatal("expected credential removal failure")
	}
	if result.Archived || result.ArchivePath != "" {
		t.Fatalf("expected rolled-back archive result, got %#v", result)
	}
	if archiver.restoreCalls != 1 || archiver.restoredPath != "/media/archive" {
		t.Fatalf("expected configuration restore, got %#v", archiver)
	}
}

func TestDataManagementServiceListsConfigurationPresence(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	archiver := &dataArchiver{
		present: map[string]bool{"radarr": true},
		sizes:   map[string]uint64{"radarr": 42_000},
	}
	service := NewDataManagementService(
		NewCatalog(registry),
		&dataSetup{status: SetupStatus{StoragePath: "/media"}},
		&managementRuntime{},
		archiver,
	)

	statuses, err := service.ListStatuses()
	if err != nil {
		t.Fatalf("list application data: %v", err)
	}
	if status := findDataStatus(statuses, "radarr"); !status.Present || status.SizeBytes != 42_000 {
		t.Fatalf("expected Radarr data present, got %#v", statuses)
	}
	if findDataStatus(statuses, "sonarr").Present {
		t.Fatalf("expected Sonarr data absent, got %#v", statuses)
	}
}

type dataSetup struct {
	status SetupStatus
	err    error
}

func (s *dataSetup) Load() (SetupStatus, error) { return s.status, s.err }

type dataArchiver struct {
	result        storage.ArchivedApplicationData
	err           error
	calls         int
	basePath      string
	applicationID string
	present       map[string]bool
	sizes         map[string]uint64
	restoreCalls  int
	restoredPath  string
	restoreErr    error
}

func (a *dataArchiver) Restore(_, _, archivePath string) error {
	a.restoreCalls++
	a.restoredPath = archivePath
	return a.restoreErr
}

type dataCredentialStore struct {
	archiver             *dataArchiver
	deleted              []credentials.Key
	deletedBeforeArchive bool
	deleteErr            error
}

func (s *dataCredentialStore) Save(context.Context, credentials.Key, credentials.Secret) error {
	return nil
}

func (s *dataCredentialStore) Load(context.Context, credentials.Key) (credentials.Secret, error) {
	return credentials.Secret{}, credentials.ErrCredentialNotFound
}

func (s *dataCredentialStore) Delete(_ context.Context, key credentials.Key) error {
	if s.archiver.calls == 0 {
		s.deletedBeforeArchive = true
	}
	s.deleted = append(s.deleted, key)
	return s.deleteErr
}

func (a *dataArchiver) Archive(basePath string, applicationID string) (storage.ArchivedApplicationData, error) {
	a.calls++
	a.basePath = basePath
	a.applicationID = applicationID
	return a.result, a.err
}

func (a *dataArchiver) Inspect(_ string, applicationID string) (storage.ApplicationDataStatus, error) {
	return storage.ApplicationDataStatus{
		ApplicationID: applicationID,
		Present:       a.present[applicationID],
		SizeBytes:     a.sizes[applicationID],
	}, nil
}

func findDataStatus(statuses []storage.ApplicationDataStatus, id string) storage.ApplicationDataStatus {
	for _, status := range statuses {
		if status.ApplicationID == id {
			return status
		}
	}
	return storage.ApplicationDataStatus{}
}
