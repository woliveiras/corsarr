package storage

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ArchivedApplicationData struct {
	ApplicationID string `json:"applicationId"`
	Archived      bool   `json:"archived"`
	ArchivePath   string `json:"-"`
}

type ApplicationDataStatus struct {
	ApplicationID string `json:"applicationId"`
	Present       bool   `json:"present"`
	SizeBytes     uint64 `json:"sizeBytes"`
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
	err = filepath.WalkDir(configPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == configPath {
			return nil
		}
		status.Present = true
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("application config contains a symbolic link")
		}
		if entry.IsDir() {
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("application config contains a special file")
		}
		fileSize := info.Size()
		if fileSize < 0 || uint64(fileSize) > math.MaxUint64-status.SizeBytes {
			return fmt.Errorf("application config size cannot be represented")
		}
		status.SizeBytes += uint64(fileSize)
		return nil
	})
	if err != nil {
		return ApplicationDataStatus{ApplicationID: applicationID}, fmt.Errorf(
			"measure application config: %w",
			err,
		)
	}
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

// Restore moves an archive produced by Archive back to the application's
// configuration path. It is used only to roll back a failed credential cleanup.
func (m *ApplicationDataManager) Restore(basePath, applicationID, archivePath string) error {
	if !safeApplicationIDPattern.MatchString(applicationID) {
		return fmt.Errorf("unsafe application ID: %q", applicationID)
	}
	rootPath := filepath.Join(basePath, "Corsarr")
	archiveParent := filepath.Join(rootPath, "trash", "config", applicationID)
	relativeArchive, err := filepath.Rel(archiveParent, archivePath)
	if err != nil || relativeArchive == "." || relativeArchive == ".." ||
		strings.HasPrefix(relativeArchive, ".."+string(filepath.Separator)) {
		return fmt.Errorf("application archive is outside the Corsarr trash")
	}
	archiveInfo, err := os.Lstat(archivePath)
	if err != nil {
		return fmt.Errorf("inspect application archive: %w", err)
	}
	if !archiveInfo.IsDir() || archiveInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("application archive is not a regular directory")
	}

	destinationPath := filepath.Join(rootPath, "config", applicationID)
	if _, err := os.Lstat(destinationPath); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return fmt.Errorf("application config already exists")
		}
		return fmt.Errorf("inspect application config destination: %w", err)
	}
	if err := os.Rename(archivePath, destinationPath); err != nil {
		return fmt.Errorf("restore application config: %w", err)
	}
	return nil
}
