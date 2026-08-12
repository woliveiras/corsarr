package application

import (
	"context"
	"errors"
	"testing"

	"github.com/woliveiras/corsarr/internal/credentials"
)

func TestServiceAccessReportsStoredQBittorrentCredentialWithoutRevealingIt(t *testing.T) {
	store := &serviceAccessStore{secret: credentials.NewSecret("private-password")}
	service := NewServiceAccess(store)

	status, err := service.QBittorrentStatus(context.Background())
	if err != nil {
		t.Fatalf("get qBittorrent access status: %v", err)
	}
	if !status.Available || status.Username != "corsarr" || status.ApplicationID != "qbittorrent" {
		t.Fatalf("unexpected access status %#v", status)
	}
}

func TestServiceAccessReportsStoredJellyfinCredentialWithoutRevealingIt(t *testing.T) {
	store := &serviceAccessStore{secret: credentials.NewSecret("private-password")}
	service := NewServiceAccess(store)

	status, err := service.JellyfinStatus(context.Background())
	if err != nil {
		t.Fatalf("get Jellyfin access status: %v", err)
	}
	if !status.Available || status.Username != "corsarr" || status.ApplicationID != "jellyfin" {
		t.Fatalf("unexpected access status %#v", status)
	}
}

func TestServiceAccessReportsStoredLazyLibrarianCredentialWithoutRevealingIt(t *testing.T) {
	store := &serviceAccessStore{secret: credentials.NewSecret("private-password")}
	service := NewServiceAccess(store)

	status, err := service.LazyLibrarianStatus(context.Background())
	if err != nil {
		t.Fatalf("get LazyLibrarian access status: %v", err)
	}
	if !status.Available || status.Username != "corsarr" || status.ApplicationID != "lazylibrarian" {
		t.Fatalf("unexpected access status %#v", status)
	}
}

func TestServiceAccessReportsArrCredentialsWithoutRevealingThem(t *testing.T) {
	store := &arrServiceAccessStore{secrets: map[credentials.Key]credentials.Secret{
		credentials.KeyRadarrPassword:   credentials.NewSecret("radarr-private"),
		credentials.KeyProwlarrPassword: credentials.NewSecret("prowlarr-private"),
	}}
	service := NewServiceAccess(store)

	statuses, err := service.ARRStatuses(context.Background())
	if err != nil {
		t.Fatalf("get Arr access statuses: %v", err)
	}
	if len(statuses) != 4 {
		t.Fatalf("expected every supported Arr status, got %#v", statuses)
	}
	available := map[string]bool{}
	for _, status := range statuses {
		if status.Username != "corsarr" {
			t.Fatalf("unexpected username in %#v", status)
		}
		available[status.ApplicationID] = status.Available
	}
	if !available["radarr"] || !available["prowlarr"] || available["sonarr"] || available["lidarr"] {
		t.Fatalf("unexpected Arr availability %#v", available)
	}
}

func TestServiceAccessAllowsOnlyKnownArrPassword(t *testing.T) {
	store := &arrServiceAccessStore{secrets: map[credentials.Key]credentials.Secret{
		credentials.KeySonarrPassword: credentials.NewSecret("sonarr-private"),
	}}
	service := NewServiceAccess(store)

	secret, err := service.ARRPassword(context.Background(), "sonarr")
	if err != nil || secret.Reveal() != "sonarr-private" {
		t.Fatalf("load Sonarr password: %v", err)
	}
	if _, err := service.ARRPassword(context.Background(), "../../foreign"); err == nil {
		t.Fatal("expected arbitrary credential lookup to be rejected")
	}
}

type serviceAccessStore struct {
	secret credentials.Secret
	err    error
}

func (s *serviceAccessStore) Save(context.Context, credentials.Key, credentials.Secret) error {
	return nil
}
func (s *serviceAccessStore) Load(context.Context, credentials.Key) (credentials.Secret, error) {
	return s.secret, s.err
}
func (s *serviceAccessStore) Delete(context.Context, credentials.Key) error { return nil }

type arrServiceAccessStore struct {
	secrets map[credentials.Key]credentials.Secret
}

func (s *arrServiceAccessStore) Save(context.Context, credentials.Key, credentials.Secret) error {
	return nil
}

func (s *arrServiceAccessStore) Load(
	_ context.Context,
	key credentials.Key,
) (credentials.Secret, error) {
	secret, exists := s.secrets[key]
	if !exists {
		return credentials.Secret{}, credentials.ErrCredentialNotFound
	}
	return secret, nil
}

func (s *arrServiceAccessStore) Delete(context.Context, credentials.Key) error {
	return errors.New("not implemented")
}
