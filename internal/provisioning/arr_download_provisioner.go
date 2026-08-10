package provisioning

import (
	"context"
	"fmt"

	"github.com/woliveiras/corsarr/internal/credentials"
)

type DownloadClientConfigurator interface {
	EnsureQBittorrentDownloadClient(
		ctx context.Context,
		applicationID string,
		apiKey APIKey,
		username string,
		password credentials.Secret,
	) error
}

type ARRDownloadClientProvisioner struct {
	arrCredentials CredentialReader
	serviceSecrets credentials.Store
	client         DownloadClientConfigurator
}

func NewARRDownloadClientProvisioner(
	arrCredentials CredentialReader,
	serviceSecrets credentials.Store,
	client DownloadClientConfigurator,
) *ARRDownloadClientProvisioner {
	return &ARRDownloadClientProvisioner{
		arrCredentials: arrCredentials,
		serviceSecrets: serviceSecrets,
		client:         client,
	}
}

func (p *ARRDownloadClientProvisioner) Provision(
	ctx context.Context,
	rootPath string,
	applicationID string,
) error {
	if _, supported := arrCategoryFields[applicationID]; !supported {
		return nil
	}
	arrKey, err := p.arrCredentials.Read(rootPath, applicationID)
	if err != nil {
		return fmt.Errorf("read Arr API credential: %w", err)
	}
	qbittorrentPassword, err := p.serviceSecrets.Load(
		ctx,
		credentials.KeyQBitTorrentPassword,
	)
	if err != nil {
		return fmt.Errorf("load qBittorrent credential: %w", err)
	}
	if err := p.client.EnsureQBittorrentDownloadClient(
		ctx,
		applicationID,
		arrKey,
		qbittorrentUsername,
		qbittorrentPassword,
	); err != nil {
		return fmt.Errorf("ensure qBittorrent download client: %w", err)
	}
	return nil
}
