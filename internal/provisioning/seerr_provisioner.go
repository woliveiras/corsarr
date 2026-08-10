package provisioning

import (
	"context"
	"fmt"

	"github.com/woliveiras/corsarr/internal/credentials"
)

type SeerrConfigurator interface {
	EnsureSetup(context.Context, credentials.Secret, APIKey, APIKey) error
}

type SeerrProvisioner struct {
	secrets        credentials.Store
	arrCredentials CredentialReader
	client         SeerrConfigurator
}

func NewSeerrProvisioner(
	secrets credentials.Store,
	arrCredentials CredentialReader,
	client SeerrConfigurator,
) *SeerrProvisioner {
	return &SeerrProvisioner{secrets: secrets, arrCredentials: arrCredentials, client: client}
}

func (p *SeerrProvisioner) Provision(ctx context.Context, rootPath, applicationID string) error {
	if applicationID != "jellyseerr" {
		return nil
	}
	jellyfinPassword, err := p.secrets.Load(ctx, credentials.KeyJellyfinPassword)
	if err != nil {
		return fmt.Errorf("load Jellyfin credential for Seerr: %w", err)
	}
	radarrKey, err := p.arrCredentials.Read(rootPath, "radarr")
	if err != nil {
		return fmt.Errorf("read Radarr credential for Seerr: %w", err)
	}
	sonarrKey, err := p.arrCredentials.Read(rootPath, "sonarr")
	if err != nil {
		return fmt.Errorf("read Sonarr credential for Seerr: %w", err)
	}
	if err := p.client.EnsureSetup(ctx, jellyfinPassword, radarrKey, sonarrKey); err != nil {
		return fmt.Errorf("ensure Seerr setup: %w", err)
	}
	return nil
}
