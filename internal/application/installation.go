package application

import (
	"context"
	"errors"
	"fmt"

	runtimecatalog "github.com/woliveiras/corsarr/internal/catalog"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

var ErrTermsNotAccepted = errors.New("current runtime terms have not been accepted")

type InstallationSetup interface {
	Load() (SetupStatus, error)
}

type InstallationLayout interface {
	Prepare(basePath string, applicationIDs []string) (storage.LayoutStatus, error)
}

type ApplicationInstaller interface {
	Install(
		ctx context.Context,
		applicationID string,
		rootPath string,
		options runtimecatalog.RuntimeOptions,
	) (containerruntime.ContainerStatus, error)
}

type ApplicationProvisioner interface {
	Provision(ctx context.Context, rootPath string, applicationID string) error
}

type InstallationItem struct {
	ApplicationID string                           `json:"applicationId"`
	Status        containerruntime.ContainerStatus `json:"status"`
	Error         string                           `json:"error,omitempty"`
}

type InstallationResult struct {
	Items    []InstallationItem `json:"items"`
	Complete bool               `json:"complete"`
}

type InstallationService struct {
	setup       InstallationSetup
	layout      InstallationLayout
	catalog     *Catalog
	installer   ApplicationInstaller
	provisioner ApplicationProvisioner
}

func NewInstallationService(
	setup InstallationSetup,
	layout InstallationLayout,
	catalog *Catalog,
	installer ApplicationInstaller,
	provisioner ApplicationProvisioner,
) *InstallationService {
	return &InstallationService{
		setup: setup, layout: layout, catalog: catalog,
		installer: installer, provisioner: provisioner,
	}
}

func (s *InstallationService) InstallSelected(
	ctx context.Context,
	options runtimecatalog.RuntimeOptions,
) (InstallationResult, error) {
	setup, err := s.setup.Load()
	if err != nil {
		return InstallationResult{}, fmt.Errorf("load reviewed setup: %w", err)
	}
	if !setup.CanInstall || !setup.TermsAccepted {
		return InstallationResult{}, ErrTermsNotAccepted
	}
	layout, err := s.layout.Prepare(setup.StoragePath, setup.Applications)
	if err != nil {
		return InstallationResult{}, fmt.Errorf("prepare reviewed storage: %w", err)
	}
	ordered, err := s.catalog.InstallationOrder(setup.Applications)
	if err != nil {
		return InstallationResult{}, fmt.Errorf("order selected applications: %w", err)
	}

	result := InstallationResult{Items: make([]InstallationItem, 0, len(ordered))}
	for _, applicationID := range ordered {
		status, installErr := s.installer.Install(ctx, applicationID, layout.RootPath, options)
		item := InstallationItem{ApplicationID: applicationID, Status: status}
		if installErr != nil {
			item.Error = installErr.Error()
			result.Items = append(result.Items, item)
			return result, nil
		}
		if provisionErr := s.provisioner.Provision(ctx, layout.RootPath, applicationID); provisionErr != nil {
			item.Error = provisionErr.Error()
			result.Items = append(result.Items, item)
			return result, nil
		}
		result.Items = append(result.Items, item)
	}
	result.Complete = true
	return result, nil
}
