package provisioning

import (
	"context"
	"errors"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestLazyLibrarianProvisionerConfiguresManagedAccessAndQBittorrent(t *testing.T) {
	store := &lazyLibrarianStore{secrets: map[credentials.Key]credentials.Secret{
		credentials.KeyQBitTorrentPassword: credentials.NewSecret("qbittorrent-private-password"),
	}}
	client := &recordingLazyLibrarianConfigurator{}
	provisioner := NewLazyLibrarianProvisioner(store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("lazylibrarian-private-password"), nil
	}
	provisioner.generateAPIKey = func() (APIKey, error) {
		return newAPIKey("0123456789abcdef0123456789abcdef")
	}

	err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"lazylibrarian",
		[]string{"lazylibrarian", "prowlarr", "qbittorrent"},
	)
	if err != nil {
		t.Fatalf("provision LazyLibrarian: %v", err)
	}
	request := client.request
	if request.Username != "corsarr" ||
		request.Password.Reveal() != "lazylibrarian-private-password" ||
		request.APIKey.Reveal() != "0123456789abcdef0123456789abcdef" ||
		!request.ConfigureQBittorrent || request.QBittorrentUsername != "corsarr" ||
		request.QBittorrentPassword.Reveal() != "qbittorrent-private-password" {
		t.Fatalf("unexpected LazyLibrarian setup request %#v", request)
	}
	for _, key := range []credentials.Key{
		credentials.KeyLazyLibrarianPassword,
		credentials.KeyLazyLibrarianAPIKey,
	} {
		if store.secrets[key].Reveal() == "" {
			t.Fatalf("expected %s to be stored before setup", key)
		}
	}
}

func TestProwlarrProvisionerLoadsLazyLibrarianCredentialsFromNativeStore(t *testing.T) {
	arr := &multiCredentialReader{keys: map[string]APIKey{
		"prowlarr": {value: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}}
	store := &lazyLibrarianStore{secrets: map[credentials.Key]credentials.Secret{
		credentials.KeyLazyLibrarianAPIKey: credentials.NewSecret(
			"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		),
		credentials.KeyLazyLibrarianPassword: credentials.NewSecret("lazy-private-password"),
	}}
	client := &recordingProwlarrConfigurator{}
	provisioner := NewProwlarrProvisioner(arr, client, store)

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"lazylibrarian",
		[]string{"lazylibrarian", "prowlarr"},
	); err != nil {
		t.Fatalf("provision LazyLibrarian in Prowlarr: %v", err)
	}
	if client.applicationID != "lazylibrarian" ||
		client.prowlarrKey.Reveal() != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" ||
		client.targetKey.Reveal() != "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" ||
		client.username != "corsarr" || client.password.Reveal() != "lazy-private-password" {
		t.Fatalf("unexpected Prowlarr LazyLibrarian configuration %#v", client)
	}
}

type recordingLazyLibrarianConfigurator struct {
	request LazyLibrarianSetupRequest
}

func (c *recordingLazyLibrarianConfigurator) EnsureSetup(
	_ context.Context,
	request LazyLibrarianSetupRequest,
) error {
	c.request = request
	return nil
}

type lazyLibrarianStore struct {
	secrets map[credentials.Key]credentials.Secret
}

func (s *lazyLibrarianStore) Save(
	_ context.Context,
	key credentials.Key,
	secret credentials.Secret,
) error {
	if s.secrets == nil {
		s.secrets = map[credentials.Key]credentials.Secret{}
	}
	s.secrets[key] = secret
	return nil
}

func (s *lazyLibrarianStore) Load(
	_ context.Context,
	key credentials.Key,
) (credentials.Secret, error) {
	secret, found := s.secrets[key]
	if !found {
		return credentials.Secret{}, credentials.ErrCredentialNotFound
	}
	return secret, nil
}

func (s *lazyLibrarianStore) Delete(context.Context, credentials.Key) error {
	return errors.New("not implemented")
}
