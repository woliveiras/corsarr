package provisioning

import (
	"context"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestSeerrProvisionerLoadsOnlyRequiredBackendCredentials(t *testing.T) {
	store := &recordingCredentialStore{loaded: credentials.NewSecret("jellyfin-password")}
	reader := &multiCredentialReader{keys: map[string]APIKey{
		"radarr": {value: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		"sonarr": {value: "cccccccccccccccccccccccccccccccc"},
	}}
	client := &recordingSeerrConfigurator{}
	provisioner := NewSeerrProvisioner(store, reader, client)

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"jellyseerr",
		[]string{"jellyfin", "jellyseerr", "radarr", "sonarr"},
	); err != nil {
		t.Fatalf("provision Seerr: %v", err)
	}
	if client.password.Reveal() == "" || client.radarrKey.Reveal() == "" || client.sonarrKey.Reveal() == "" {
		t.Fatal("expected all Seerr backend credentials")
	}
	if len(reader.applications) != 2 || reader.applications[0] != "radarr" || reader.applications[1] != "sonarr" {
		t.Fatalf("unexpected Arr credential reads %v", reader.applications)
	}
}

func TestSeerrProvisionerSkipsSetupWithoutManagedBackends(t *testing.T) {
	store := &recordingCredentialStore{}
	reader := &multiCredentialReader{}
	client := &recordingSeerrConfigurator{}
	provisioner := NewSeerrProvisioner(store, reader, client)

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"jellyseerr",
		[]string{"jellyseerr"},
	); err != nil {
		t.Fatalf("skip unselected backends: %v", err)
	}
	if store.loadCalls != 0 || len(reader.applications) != 0 || client.password.Reveal() != "" {
		t.Fatalf("expected external backends to remain untouched")
	}
}

type recordingSeerrConfigurator struct {
	password  credentials.Secret
	radarrKey APIKey
	sonarrKey APIKey
}

func (c *recordingSeerrConfigurator) EnsureSetup(
	_ context.Context,
	password credentials.Secret,
	radarrKey APIKey,
	sonarrKey APIKey,
) error {
	c.password, c.radarrKey, c.sonarrKey = password, radarrKey, sonarrKey
	return nil
}
