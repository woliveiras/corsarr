package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ArchivedApplicationData struct {
	ApplicationID string `json:"applicationId"`
	Archived      bool   `json:"archived"`
	ArchivePath   string `json:"archivePath,omitempty"`
}

type ApplicationDataStatus struct {
	ApplicationID string `json:"applicationId"`
	Present       bool   `json:"present"`
}

// ApplicationDataManager moves a single Corsarr-owned application config directory
// to Corsarr's private trash tree. Shared media and download directories are never targeted.
type ApplicationDataManager struct {
	now func() time.Time
}

func NewApplicationDataManager() *ApplicationDataManager {
	return &ApplicationDataManager{now: time.Now}
}

func (m *ApplicationDataManager) Inspect(
	basePath string,
	applicationID string,
) (ApplicationDataStatus, error) {
	status := ApplicationDataStatus{ApplicationID: applicationID}
	if !safeApplicationIDPattern.MatchString(applicationID) {
		return status, fmt.Errorf("unsafe application ID: %q", applicationID)
	}

	configPath := filepath.Join(basePath, "Corsarr", "config", applicationID)
	configInfo, err := os.Lstat(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return status, nil
	}
	if err != nil {
		return status, fmt.Errorf("inspect application config: %w", err)
	}
	if !configInfo.IsDir() || configInfo.Mode()&os.ModeSymlink != 0 {
		return status, fmt.Errorf("application config is not a regular directory")
	}
	entries, err := os.ReadDir(configPath)
	if err != nil {
		return status, fmt.Errorf("read application config: %w", err)
	}
	status.Present = len(entries) > 0
	return status, nil
}

func (m *ApplicationDataManager) Archive(
	basePath string,
	applicationID string,
) (ArchivedApplicationData, error) {
	result := ArchivedApplicationData{ApplicationID: applicationID}
	if !safeApplicationIDPattern.MatchString(applicationID) {
		return result, fmt.Errorf("unsafe application ID: %q", applicationID)
	}

	rootPath := filepath.Join(basePath, "Corsarr")
	sourcePath := filepath.Join(rootPath, "config", applicationID)
	sourceInfo, err := os.Lstat(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("inspect application config: %w", err)
	}
	if !sourceInfo.IsDir() || sourceInfo.Mode()&os.ModeSymlink != 0 {
		return result, fmt.Errorf("application config is not a regular directory")
	}

	archiveParent := filepath.Join(rootPath, "trash", "config", applicationID)
	if err := os.MkdirAll(archiveParent, 0o700); err != nil {
		return result, fmt.Errorf("create application archive directory: %w", err)
	}
	if err := os.Chmod(filepath.Join(rootPath, "trash"), 0o700); err != nil {
		return result, fmt.Errorf("protect trash directory: %w", err)
	}
	if err := os.Chmod(filepath.Join(rootPath, "trash", "config"), 0o700); err != nil {
		return result, fmt.Errorf("protect config archive directory: %w", err)
	}
	if err := os.Chmod(archiveParent, 0o700); err != nil {
		return result, fmt.Errorf("protect application archive directory: %w", err)
	}

	archiveName := m.now().UTC().Format("20060102T150405.000000000Z")
	archivePath := filepath.Join(archiveParent, archiveName)
	if err := os.Rename(sourcePath, archivePath); err != nil {
		return result, fmt.Errorf("archive application config: %w", err)
	}
	result.Archived = true
	result.ArchivePath = archivePath
	return result, nil
}
