package provisioning

import (
	"context"
	"errors"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestARRAuthenticationProvisionerBootstrapsPristineApplication(t *testing.T) {
	store := &arrAuthenticationStore{loadErr: credentials.ErrCredentialNotFound}
	client := &recordingARRAuthenticationClient{
		inspection: ARRAuthenticationInspection{BootstrapRequired: true},
		activation: ARRAuthenticationActivation{CredentialAccepted: true},
	}
	provisioner := NewARRAuthenticationProvisioner(arrCredentialReader{}, store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("generated-password"), nil
	}

	err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"radarr",
		[]string{"radarr"},
	)
	if err != nil {
		t.Fatalf("bootstrap Radarr authentication: %v", err)
	}
	if store.savedKey != credentials.KeyRadarrPassword ||
		store.saved.Reveal() != "generated-password" || store.deleteCalls != 0 {
		t.Fatalf("unexpected credential lifecycle %#v", store)
	}
	if client.applicationID != "radarr" || client.username != "corsarr" ||
		client.password.Reveal() != "generated-password" {
		t.Fatalf("unexpected activation %#v", client)
	}
}

func TestARRAuthenticationProvisionerPreservesManualAuthentication(t *testing.T) {
	store := &arrAuthenticationStore{}
	client := &recordingARRAuthenticationClient{
		inspection: ARRAuthenticationInspection{BootstrapRequired: false, CorsarrManaged: false},
	}
	provisioner := NewARRAuthenticationProvisioner(arrCredentialReader{}, store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		t.Fatal("manual authentication must not generate a credential")
		return credentials.Secret{}, nil
	}

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"sonarr",
		[]string{"sonarr"},
	); err != nil {
		t.Fatalf("preserve manual authentication: %v", err)
	}
	if store.loadCalls != 0 || store.saveCalls != 0 || client.activationCalls != 0 {
		t.Fatalf("manual authentication was mutated: store=%#v client=%#v", store, client)
	}
}

func TestARRAuthenticationProvisionerDeletesCredentialRejectedBeforeActivation(t *testing.T) {
	store := &arrAuthenticationStore{loadErr: credentials.ErrCredentialNotFound}
	client := &recordingARRAuthenticationClient{
		inspection:  ARRAuthenticationInspection{BootstrapRequired: true},
		activation:  ARRAuthenticationActivation{CredentialAccepted: false},
		activateErr: errors.New("request rejected"),
	}
	provisioner := NewARRAuthenticationProvisioner(arrCredentialReader{}, store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("generated-password"), nil
	}

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"prowlarr",
		[]string{"prowlarr"},
	); err == nil {
		t.Fatal("expected rejected authentication to fail")
	}
	if store.deleteCalls != 1 || store.deletedKey != credentials.KeyProwlarrPassword {
		t.Fatalf("expected unused credential cleanup, got %#v", store)
	}
}

func TestARRAuthenticationProvisionerKeepsCredentialAfterAcceptedChange(t *testing.T) {
	store := &arrAuthenticationStore{loadErr: credentials.ErrCredentialNotFound}
	client := &recordingARRAuthenticationClient{
		inspection:  ARRAuthenticationInspection{BootstrapRequired: true},
		activation:  ARRAuthenticationActivation{CredentialAccepted: true},
		activateErr: errors.New("verification interrupted"),
	}
	provisioner := NewARRAuthenticationProvisioner(arrCredentialReader{}, store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("generated-password"), nil
	}

	if err := provisioner.Provision(
		context.Background(),
		"/host/Corsarr",
		"radarr",
		[]string{"radarr"},
	); err == nil {
		t.Fatal("expected interrupted verification to fail")
	}
	if store.deleteCalls != 0 {
		t.Fatal("accepted credential must be retained for recovery")
	}
}

type arrCredentialReader struct{}

func (arrCredentialReader) Read(string, string) (APIKey, error) {
	return APIKey{value: "0123456789abcdef0123456789abcdef"}, nil
}

type arrAuthenticationStore struct {
	loaded      credentials.Secret
	loadErr     error
	saved       credentials.Secret
	savedKey    credentials.Key
	deletedKey  credentials.Key
	loadCalls   int
	saveCalls   int
	deleteCalls int
}

func (s *arrAuthenticationStore) Save(
	_ context.Context,
	key credentials.Key,
	secret credentials.Secret,
) error {
	s.saveCalls++
	s.savedKey = key
	s.saved = secret
	return nil
}

func (s *arrAuthenticationStore) Load(
	context.Context,
	credentials.Key,
) (credentials.Secret, error) {
	s.loadCalls++
	return s.loaded, s.loadErr
}

func (s *arrAuthenticationStore) Delete(_ context.Context, key credentials.Key) error {
	s.deleteCalls++
	s.deletedKey = key
	return nil
}

type recordingARRAuthenticationClient struct {
	inspection      ARRAuthenticationInspection
	inspectErr      error
	activation      ARRAuthenticationActivation
	activateErr     error
	applicationID   string
	username        string
	password        credentials.Secret
	activationCalls int
}

func (c *recordingARRAuthenticationClient) InspectAuthentication(
	context.Context,
	string,
	APIKey,
) (ARRAuthenticationInspection, error) {
	return c.inspection, c.inspectErr
}

func (c *recordingARRAuthenticationClient) ActivateAuthentication(
	_ context.Context,
	applicationID string,
	_ APIKey,
	username string,
	password credentials.Secret,
) (ARRAuthenticationActivation, error) {
	c.activationCalls++
	c.applicationID = applicationID
	c.username = username
	c.password = password
	return c.activation, c.activateErr
}
