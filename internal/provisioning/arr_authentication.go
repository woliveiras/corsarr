package provisioning

import (
	"context"
	"errors"
	"fmt"

	"github.com/woliveiras/corsarr/internal/credentials"
)

type ARRAuthenticationConfigurator interface {
	InspectAuthentication(
		ctx context.Context,
		applicationID string,
		apiKey APIKey,
	) (ARRAuthenticationInspection, error)
	ActivateAuthentication(
		ctx context.Context,
		applicationID string,
		apiKey APIKey,
		username string,
		password credentials.Secret,
	) (ARRAuthenticationActivation, error)
}

type ARRAuthenticationProvisioner struct {
	apiCredentials   CredentialReader
	serviceSecrets   credentials.Store
	client           ARRAuthenticationConfigurator
	generatePassword func() (credentials.Secret, error)
}

func NewARRAuthenticationProvisioner(
	apiCredentials CredentialReader,
	serviceSecrets credentials.Store,
	client ARRAuthenticationConfigurator,
) *ARRAuthenticationProvisioner {
	return &ARRAuthenticationProvisioner{
		apiCredentials:   apiCredentials,
		serviceSecrets:   serviceSecrets,
		client:           client,
		generatePassword: generateQBittorrentPassword,
	}
}

func (p *ARRAuthenticationProvisioner) Provision(
	ctx context.Context,
	rootPath string,
	applicationID string,
	_ []string,
) error {
	if _, supported := arrAuthenticationAPIVersions[applicationID]; !supported {
		return nil
	}
	credentialKey, err := credentials.ARRPasswordKey(applicationID)
	if err != nil {
		return fmt.Errorf("resolve Arr credential key: %w", err)
	}
	apiKey, err := p.apiCredentials.Read(rootPath, applicationID)
	if err != nil {
		return fmt.Errorf("read Arr API credential for authentication: %w", err)
	}
	inspection, err := p.client.InspectAuthentication(ctx, applicationID, apiKey)
	if err != nil {
		return fmt.Errorf("inspect Arr authentication: %w", err)
	}
	if !inspection.BootstrapRequired {
		return nil
	}

	password, err := p.serviceSecrets.Load(ctx, credentialKey)
	created := false
	if errors.Is(err, credentials.ErrCredentialNotFound) {
		password, err = p.generatePassword()
		if err != nil {
			return fmt.Errorf("generate Arr credential: %w", err)
		}
		if err := p.serviceSecrets.Save(ctx, credentialKey, password); err != nil {
			return fmt.Errorf("store Arr credential before activation: %w", err)
		}
		created = true
	} else if err != nil {
		return fmt.Errorf("load Arr credential: %w", err)
	}

	result, activationErr := p.client.ActivateAuthentication(
		ctx,
		applicationID,
		apiKey,
		arrManagedUsername,
		password,
	)
	if activationErr == nil {
		return nil
	}
	if created && !result.CredentialAccepted {
		if cleanupErr := p.serviceSecrets.Delete(ctx, credentialKey); cleanupErr != nil {
			return errors.Join(
				fmt.Errorf("configure Arr authentication: %w", activationErr),
				fmt.Errorf("remove unused Arr credential: %w", cleanupErr),
			)
		}
	}
	return fmt.Errorf("configure Arr authentication: %w", activationErr)
}
