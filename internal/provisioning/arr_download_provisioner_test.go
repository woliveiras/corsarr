package provisioning

import (
	"context"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestARRDownloadClientProvisionerUsesStoredQBittorrentCredential(t *testing.T) {
	reader := &recordingCredentialReader{credential: APIKey{value: "0123456789abcdef0123456789abcdef"}}
	store := &recordingCredentialStore{loaded: credentials.NewSecret("qbit-password")}
	client := &recordingDownloadClientConfigurator{}
	provisioner := NewARRDownloadClientProvisioner(reader, store, client)

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"radarr",
		[]string{"radarr", "qbittorrent"},
	); err != nil {
		t.Fatalf("provision Radarr download client: %v", err)
	}
	if client.applicationID != "radarr" || client.username != "corsarr" ||
		client.password.Reveal() != "qbit-password" {
		t.Fatalf("unexpected download client configuration %#v", client)
	}
	if reader.rootPath != "/host/Corsarr" || store.loadCalls != 1 {
		t.Fatalf("unexpected credential access reader=%#v store=%#v", reader, store)
	}
}

func TestARRDownloadClientProvisionerSkipsUnsupportedApplication(t *testing.T) {
	reader := &recordingCredentialReader{}
	store := &recordingCredentialStore{}
	client := &recordingDownloadClientConfigurator{}
	provisioner := NewARRDownloadClientProvisioner(reader, store, client)

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "jellyfin", nil); err != nil {
		t.Fatalf("skip Jellyfin download client: %v", err)
	}
	if reader.calls != 0 || store.loadCalls != 0 || client.calls != 0 {
		t.Fatalf("expected unsupported application unchanged")
	}
}

func TestARRDownloadClientProvisionerSkipsUnselectedQBittorrent(t *testing.T) {
	reader := &recordingCredentialReader{}
	store := &recordingCredentialStore{}
	client := &recordingDownloadClientConfigurator{}
	provisioner := NewARRDownloadClientProvisioner(reader, store, client)

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"radarr",
		[]string{"radarr"},
	); err != nil {
		t.Fatalf("skip unselected qBittorrent: %v", err)
	}
	if reader.calls != 0 || store.loadCalls != 0 || client.calls != 0 {
		t.Fatalf("expected external download client to remain untouched")
	}
}

type recordingDownloadClientConfigurator struct {
	applicationID string
	username      string
	password      credentials.Secret
	calls         int
}

func (c *recordingDownloadClientConfigurator) EnsureQBittorrentDownloadClient(
	_ context.Context,
	applicationID string,
	_ APIKey,
	username string,
	password credentials.Secret,
) error {
	c.calls++
	c.applicationID = applicationID
	c.username = username
	c.password = password
	return nil
}
