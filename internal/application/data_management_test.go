package application

import (
	"context"
	"errors"
	"testing"

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

func TestDataManagementServiceListsConfigurationPresence(t *testing.T) {
	registry, err := services.NewRegistry()
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	archiver := &dataArchiver{present: map[string]bool{"radarr": true}}
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
	if !findDataStatus(statuses, "radarr").Present {
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
