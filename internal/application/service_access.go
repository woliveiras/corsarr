package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/woliveiras/corsarr/internal/credentials"
)

type ServiceAccessStatus struct {
	ApplicationID string `json:"applicationId"`
	Username      string `json:"username"`
	Available     bool   `json:"available"`
}

type ServiceAccess struct {
	store credentials.Store
}

func NewServiceAccess(store credentials.Store) *ServiceAccess {
	return &ServiceAccess{store: store}
}

func (s *ServiceAccess) QBittorrentStatus(ctx context.Context) (ServiceAccessStatus, error) {
	status := ServiceAccessStatus{ApplicationID: "qbittorrent", Username: "corsarr"}
	_, err := s.store.Load(ctx, credentials.KeyQBitTorrentPassword)
	if errors.Is(err, credentials.ErrCredentialNotFound) {
		return status, nil
	}
	if err != nil {
		return status, fmt.Errorf("load qBittorrent access status: %w", err)
	}
	status.Available = true
	return status, nil
}

func (s *ServiceAccess) QBittorrentPassword(ctx context.Context) (credentials.Secret, error) {
	secret, err := s.store.Load(ctx, credentials.KeyQBitTorrentPassword)
	if err != nil {
		return credentials.Secret{}, fmt.Errorf("load qBittorrent credential: %w", err)
	}
	return secret, nil
}

func (s *ServiceAccess) JellyfinStatus(ctx context.Context) (ServiceAccessStatus, error) {
	status := ServiceAccessStatus{ApplicationID: "jellyfin", Username: "corsarr"}
	_, err := s.store.Load(ctx, credentials.KeyJellyfinPassword)
	if errors.Is(err, credentials.ErrCredentialNotFound) {
		return status, nil
	}
	if err != nil {
		return status, fmt.Errorf("load Jellyfin access status: %w", err)
	}
	status.Available = true
	return status, nil
}

func (s *ServiceAccess) JellyfinPassword(ctx context.Context) (credentials.Secret, error) {
	secret, err := s.store.Load(ctx, credentials.KeyJellyfinPassword)
	if err != nil {
		return credentials.Secret{}, fmt.Errorf("load Jellyfin credential: %w", err)
	}
	return secret, nil
}
