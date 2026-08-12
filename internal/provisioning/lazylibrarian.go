package provisioning

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/woliveiras/corsarr/internal/credentials"
)

const (
	lazyLibrarianApplicationID = "lazylibrarian"
	lazyLibrarianUsername      = "corsarr"
)

type LazyLibrarianSetupRequest struct {
	Username             string
	Password             credentials.Secret
	APIKey               APIKey
	ConfigureQBittorrent bool
	QBittorrentUsername  string
	QBittorrentPassword  credentials.Secret
}

type LazyLibrarianConfigurator interface {
	EnsureSetup(ctx context.Context, request LazyLibrarianSetupRequest) error
}

type LazyLibrarianProvisioner struct {
	secrets          credentials.Store
	client           LazyLibrarianConfigurator
	generatePassword func() (credentials.Secret, error)
	generateAPIKey   func() (APIKey, error)
}

func NewLazyLibrarianProvisioner(
	secrets credentials.Store,
	client LazyLibrarianConfigurator,
) *LazyLibrarianProvisioner {
	return &LazyLibrarianProvisioner{
		secrets:          secrets,
		client:           client,
		generatePassword: generateQBittorrentPassword,
		generateAPIKey:   generateLazyLibrarianAPIKey,
	}
}

func (p *LazyLibrarianProvisioner) Provision(
	ctx context.Context,
	_ string,
	applicationID string,
	selected []string,
) error {
	if applicationID != lazyLibrarianApplicationID {
		return nil
	}

	password, err := p.loadOrCreateSecret(
		ctx,
		credentials.KeyLazyLibrarianPassword,
		p.generatePassword,
	)
	if err != nil {
		return fmt.Errorf("prepare LazyLibrarian password: %w", err)
	}
	apiKeySecret, err := p.loadOrCreateSecret(
		ctx,
		credentials.KeyLazyLibrarianAPIKey,
		func() (credentials.Secret, error) {
			key, keyErr := p.generateAPIKey()
			return credentials.NewSecret(key.Reveal()), keyErr
		},
	)
	if err != nil {
		return fmt.Errorf("prepare LazyLibrarian API key: %w", err)
	}
	apiKey, err := newAPIKey(apiKeySecret.Reveal())
	if err != nil {
		return fmt.Errorf("load LazyLibrarian API key: %w", err)
	}

	request := LazyLibrarianSetupRequest{
		Username: lazyLibrarianUsername,
		Password: password,
		APIKey:   apiKey,
	}
	if selectedApplication(selected, qbittorrentApplicationID) {
		request.ConfigureQBittorrent = true
		request.QBittorrentUsername = qbittorrentUsername
		request.QBittorrentPassword, err = p.secrets.Load(
			ctx,
			credentials.KeyQBitTorrentPassword,
		)
		if err != nil {
			return fmt.Errorf("load qBittorrent credential for LazyLibrarian: %w", err)
		}
	}
	if err := p.client.EnsureSetup(ctx, request); err != nil {
		return fmt.Errorf("configure LazyLibrarian: %w", err)
	}
	return nil
}

func (p *LazyLibrarianProvisioner) loadOrCreateSecret(
	ctx context.Context,
	key credentials.Key,
	generate func() (credentials.Secret, error),
) (credentials.Secret, error) {
	secret, err := p.secrets.Load(ctx, key)
	if err == nil {
		return secret, nil
	}
	if !errors.Is(err, credentials.ErrCredentialNotFound) {
		return credentials.Secret{}, err
	}
	secret, err = generate()
	if err != nil {
		return credentials.Secret{}, err
	}
	if err := p.secrets.Save(ctx, key, secret); err != nil {
		return credentials.Secret{}, err
	}
	return secret, nil
}

func generateLazyLibrarianAPIKey() (APIKey, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return APIKey{}, err
	}
	return newAPIKey(hex.EncodeToString(random))
}

func newAPIKey(value string) (APIKey, error) {
	if !arrAPIKeyPattern.MatchString(value) {
		return APIKey{}, fmt.Errorf("API key is invalid")
	}
	return APIKey{value: value}, nil
}
