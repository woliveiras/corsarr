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

var managedARRApplications = []string{"lidarr", "prowlarr", "radarr", "sonarr"}

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

func (s *ServiceAccess) LazyLibrarianStatus(
	ctx context.Context,
) (ServiceAccessStatus, error) {
	status := ServiceAccessStatus{ApplicationID: "lazylibrarian", Username: "corsarr"}
	_, err := s.store.Load(ctx, credentials.KeyLazyLibrarianPassword)
	if errors.Is(err, credentials.ErrCredentialNotFound) {
		return status, nil
	}
	if err != nil {
		return status, fmt.Errorf("load LazyLibrarian access status: %w", err)
	}
	status.Available = true
	return status, nil
}

func (s *ServiceAccess) LazyLibrarianPassword(
	ctx context.Context,
) (credentials.Secret, error) {
	secret, err := s.store.Load(ctx, credentials.KeyLazyLibrarianPassword)
	if err != nil {
		return credentials.Secret{}, fmt.Errorf("load LazyLibrarian credential: %w", err)
	}
	return secret, nil
}

func (s *ServiceAccess) ARRStatuses(
	ctx context.Context,
) ([]ServiceAccessStatus, error) {
	statuses := make([]ServiceAccessStatus, 0, len(managedARRApplications))
	for _, applicationID := range managedARRApplications {
		status := ServiceAccessStatus{ApplicationID: applicationID, Username: "corsarr"}
		key, err := credentials.ARRPasswordKey(applicationID)
		if err != nil {
			return nil, fmt.Errorf("resolve %s access status: %w", applicationID, err)
		}
		_, err = s.store.Load(ctx, key)
		if errors.Is(err, credentials.ErrCredentialNotFound) {
			statuses = append(statuses, status)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("load %s access status: %w", applicationID, err)
		}
		status.Available = true
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func (s *ServiceAccess) ARRPassword(
	ctx context.Context,
	applicationID string,
) (credentials.Secret, error) {
	key, err := credentials.ARRPasswordKey(applicationID)
	if err != nil {
		return credentials.Secret{}, err
	}
	secret, err := s.store.Load(ctx, key)
	if err != nil {
		return credentials.Secret{}, fmt.Errorf("load %s credential: %w", applicationID, err)
	}
	return secret, nil
}
