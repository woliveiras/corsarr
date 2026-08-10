package application

import (
	"context"
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
