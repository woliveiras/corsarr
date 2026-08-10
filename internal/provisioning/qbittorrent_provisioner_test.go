package provisioning

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestQBittorrentProvisionerBootstrapsTemporaryCredential(t *testing.T) {
	store := &recordingCredentialStore{loadErr: credentials.ErrCredentialNotFound}
	logs := &recordingLogReader{contents: "The WebUI administrator username is: admin\nThe WebUI administrator password was not set. A temporary password is provided for this session: temp-pass\n"}
	client := &recordingQBittorrentAPI{}
	provisioner := NewQBittorrentProvisioner(logs, store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("permanent-password"), nil
	}

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "qbittorrent"); err != nil {
		t.Fatalf("bootstrap qBittorrent: %v", err)
	}
	wantOperations := []string{
		"login:admin:temp-pass",
		"set:corsarr:permanent-password",
		"login:corsarr:permanent-password",
		"paths",
		"categories",
	}
	if !reflect.DeepEqual(client.operations, wantOperations) {
		t.Fatalf("unexpected qBittorrent operations\nwant: %v\n got: %v", wantOperations, client.operations)
	}
	if store.saved.Reveal() != "permanent-password" || store.saveCalls != 1 {
		t.Fatalf("expected permanent credential saved before use, store=%#v", store)
	}
	if logs.calls != 1 || logs.tail != 200 {
		t.Fatalf("expected one bounded log read, logs=%#v", logs)
	}
}

func TestQBittorrentProvisionerReusesStoredCredential(t *testing.T) {
	store := &recordingCredentialStore{loaded: credentials.NewSecret("stored-password")}
	logs := &recordingLogReader{}
	client := &recordingQBittorrentAPI{}
	provisioner := NewQBittorrentProvisioner(logs, store, client)

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "qbittorrent"); err != nil {
		t.Fatalf("reconcile qBittorrent: %v", err)
	}
	want := []string{"login:corsarr:stored-password", "paths", "categories"}
	if !reflect.DeepEqual(client.operations, want) {
		t.Fatalf("unexpected reconcile operations\nwant: %v\n got: %v", want, client.operations)
	}
	if logs.calls != 0 || store.saveCalls != 0 {
		t.Fatalf("expected stored credential without bootstrap, logs=%#v store=%#v", logs, store)
	}
}

func TestQBittorrentProvisionerDeletesPreparedCredentialWhenChangeFails(t *testing.T) {
	store := &recordingCredentialStore{loadErr: credentials.ErrCredentialNotFound}
	logs := &recordingLogReader{contents: "temporary password is provided for this session: temp-pass"}
	client := &recordingQBittorrentAPI{setErr: errors.New("rejected")}
	provisioner := NewQBittorrentProvisioner(logs, store, client)
	provisioner.generatePassword = func() (credentials.Secret, error) {
		return credentials.NewSecret("permanent-password"), nil
	}

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "qbittorrent"); err == nil {
		t.Fatal("expected credential change failure")
	}
	if store.deleteCalls != 1 {
		t.Fatalf("expected prepared credential cleanup, got %#v", store)
	}
}

func TestQBittorrentProvisionerIgnoresOtherApplications(t *testing.T) {
	store := &recordingCredentialStore{}
	logs := &recordingLogReader{}
	client := &recordingQBittorrentAPI{}
	provisioner := NewQBittorrentProvisioner(logs, store, client)

	if err := provisioner.Provision(context.Background(), "/host/Corsarr", "radarr"); err != nil {
		t.Fatalf("skip non-qBittorrent app: %v", err)
	}
	if store.loadCalls != 0 || logs.calls != 0 || len(client.operations) != 0 {
		t.Fatalf("expected no qBittorrent work, store=%#v logs=%#v client=%#v", store, logs, client)
	}
}

type recordingCredentialStore struct {
	loaded      credentials.Secret
	loadErr     error
	saved       credentials.Secret
	loadCalls   int
	saveCalls   int
	deleteCalls int
}

func (s *recordingCredentialStore) Save(_ context.Context, _ credentials.Key, secret credentials.Secret) error {
	s.saveCalls++
	s.saved = secret
	return nil
}

func (s *recordingCredentialStore) Load(context.Context, credentials.Key) (credentials.Secret, error) {
	s.loadCalls++
	return s.loaded, s.loadErr
}

func (s *recordingCredentialStore) Delete(context.Context, credentials.Key) error {
	s.deleteCalls++
	return nil
}

type recordingLogReader struct {
	contents string
	err      error
	calls    int
	tail     int
}

func (r *recordingLogReader) Logs(_ context.Context, _ string, tail int) (string, error) {
	r.calls++
	r.tail = tail
	return r.contents, r.err
}

type recordingQBittorrentAPI struct {
	operations []string
	setErr     error
}

func (c *recordingQBittorrentAPI) Login(
	_ context.Context,
	username string,
	password credentials.Secret,
) (*QBittorrentSession, error) {
	c.operations = append(c.operations, "login:"+username+":"+password.Reveal())
	return &QBittorrentSession{}, nil
}

func (c *recordingQBittorrentAPI) SetCredentials(
	_ context.Context,
	_ *QBittorrentSession,
	username string,
	password credentials.Secret,
) error {
	c.operations = append(c.operations, "set:"+username+":"+password.Reveal())
	return c.setErr
}

func (c *recordingQBittorrentAPI) EnsureCategories(
	context.Context,
	*QBittorrentSession,
) error {
	c.operations = append(c.operations, "categories")
	return nil
}

func (c *recordingQBittorrentAPI) EnsureDownloadPaths(
	context.Context,
	*QBittorrentSession,
) error {
	c.operations = append(c.operations, "paths")
	return nil
}
