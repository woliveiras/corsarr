package provisioning

import (
	"context"
	"errors"
	"fmt"

	"github.com/woliveiras/corsarr/internal/credentials"
)

type JellyfinConfigurator interface {
	EnsureSetup(ctx context.Context, password credentials.Secret) (JellyfinSetupResult, error)
}

type JellyfinProvisioner struct {
	credentials      credentials.Store
	client           JellyfinConfigurator
	generatePassword func() (credentials.Secret, error)
}

func NewJellyfinProvisioner(
	store credentials.Store,
	client JellyfinConfigurator,
) *JellyfinProvisioner {
	return &JellyfinProvisioner{
		credentials:      store,
		client:           client,
		generatePassword: generateQBittorrentPassword,
	}
}

func (p *JellyfinProvisioner) Provision(
	ctx context.Context,
	_ string,
	applicationID string,
) error {
	if applicationID != "jellyfin" {
		return nil
	}
	password, err := p.credentials.Load(ctx, credentials.KeyJellyfinPassword)
	created := false
	if errors.Is(err, credentials.ErrCredentialNotFound) {
		password, err = p.generatePassword()
		if err != nil {
			return fmt.Errorf("generate Jellyfin credential: %w", err)
		}
		if err := p.credentials.Save(ctx, credentials.KeyJellyfinPassword, password); err != nil {
			return fmt.Errorf("store Jellyfin credential before activation: %w", err)
		}
		created = true
	} else if err != nil {
		return fmt.Errorf("load Jellyfin credential: %w", err)
	}

	result, setupErr := p.client.EnsureSetup(ctx, password)
	if setupErr == nil {
		return nil
	}
	if created && !result.CredentialAccepted {
		if cleanupErr := p.credentials.Delete(ctx, credentials.KeyJellyfinPassword); cleanupErr != nil {
			return errors.Join(
				fmt.Errorf("configure Jellyfin: %w", setupErr),
				fmt.Errorf("remove unused Jellyfin credential: %w", cleanupErr),
			)
		}
	}
	return fmt.Errorf("configure Jellyfin: %w", setupErr)
}
