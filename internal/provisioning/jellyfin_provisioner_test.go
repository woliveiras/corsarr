package provisioning

import (
	"context"
	"errors"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestJellyfinProvisionerStoresCredentialBeforeSetup(t *testing.T) {
	store := &recordingCredentialStore{loadErr: credentials.ErrCredentialNotFound}
	client := &recordingJellyfinConfigurator{}
	provisioner := NewJellyfinProvisioner(store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("generated-password"), nil
	}

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "jellyfin", nil); err != nil {
		t.Fatalf("provision Jellyfin: %v", err)
	}
	if store.saved.Reveal() != "generated-password" || client.password.Reveal() != "generated-password" {
		t.Fatal("expected generated password to be stored and delivered only to Jellyfin client")
	}
}

func TestJellyfinProvisionerRemovesUnusedGeneratedCredential(t *testing.T) {
	store := &recordingCredentialStore{loadErr: credentials.ErrCredentialNotFound}
	client := &recordingJellyfinConfigurator{err: errors.New("user already configured")}
	provisioner := NewJellyfinProvisioner(store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("generated-password"), nil
	}

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "jellyfin", nil); err == nil {
		t.Fatal("expected setup failure")
	}
	if store.deleteCalls != 1 {
		t.Fatalf("expected unused credential cleanup, got %d", store.deleteCalls)
	}
}

func TestJellyfinProvisionerKeepsAcceptedCredentialAfterLaterFailure(t *testing.T) {
	store := &recordingCredentialStore{loadErr: credentials.ErrCredentialNotFound}
	client := &recordingJellyfinConfigurator{
		result: JellyfinSetupResult{CredentialAccepted: true},
		err:    errors.New("library setup failed"),
	}
	provisioner := NewJellyfinProvisioner(store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("generated-password"), nil
	}

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "jellyfin", nil); err == nil {
		t.Fatal("expected setup failure")
	}
	if store.deleteCalls != 0 {
		t.Fatal("expected activated Jellyfin credential to be preserved")
	}
}

type recordingJellyfinConfigurator struct {
	password credentials.Secret
	result   JellyfinSetupResult
	err      error
}

func (c *recordingJellyfinConfigurator) EnsureSetup(
	_ context.Context,
	password credentials.Secret,
) (JellyfinSetupResult, error) {
	c.password = password
	return c.result, c.err
}
