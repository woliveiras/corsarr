package provisioning

import (
	"context"
	"fmt"
)

type CredentialReader interface {
	Read(rootPath string, applicationID string) (APIKey, error)
}

type RootFolderClient interface {
	EnsureRootFolder(
		ctx context.Context,
		applicationID string,
		apiKey APIKey,
		rootPath string,
	) error
}

type ARRProvisioner struct {
	credentials CredentialReader
	client      RootFolderClient
}

func NewARRProvisioner(credentials CredentialReader, client RootFolderClient) *ARRProvisioner {
	return &ARRProvisioner{credentials: credentials, client: client}
}

func (p *ARRProvisioner) Provision(
	ctx context.Context,
	rootPath string,
	applicationID string,
	_ []string,
) error {
	applicationRoot, supported := approvedARRRootFolders[applicationID]
	if !supported {
		return nil
	}
	credential, err := p.credentials.Read(rootPath, applicationID)
	if err != nil {
		return fmt.Errorf("read application API credential: %w", err)
	}
	if err := p.client.EnsureRootFolder(ctx, applicationID, credential, applicationRoot); err != nil {
		return fmt.Errorf("ensure application root folder: %w", err)
	}
	return nil
}
