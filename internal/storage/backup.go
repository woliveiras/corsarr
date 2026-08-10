package storage

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BackupResult struct {
	ApplicationID string    `json:"applicationId"`
	Path          string    `json:"path"`
	SHA256        string    `json:"sha256"`
	FileCount     int       `json:"fileCount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type BackupManager struct {
	now func() time.Time
}

func NewBackupManager() *BackupManager {
	return &BackupManager{now: time.Now}
}

// Backup creates a private, compressed archive of one application's config.
// It deliberately accepts a Corsarr root instead of an arbitrary source path.
func (m *BackupManager) Backup(rootPath, applicationID string) (result BackupResult, resultErr error) {
	if !safeApplicationIDPattern.MatchString(applicationID) {
		return result, fmt.Errorf("unsafe application ID: %q", applicationID)
	}
	if !filepath.IsAbs(rootPath) {
		return result, fmt.Errorf("Corsarr root must be an absolute path")
	}

	sourcePath := filepath.Join(rootPath, "config", applicationID)
	sourceInfo, err := os.Lstat(sourcePath)
	if err != nil {
		return result, fmt.Errorf("inspect application config: %w", err)
	}
	if sourceInfo.Mode()&os.ModeSymlink != 0 || !sourceInfo.IsDir() {
		return result, fmt.Errorf("application config must be a directory and not a symlink")
	}

	backupPath := filepath.Join(rootPath, "backups", "config", applicationID)
	if err := os.MkdirAll(backupPath, 0o700); err != nil {
		return result, fmt.Errorf("create backup directory: %w", err)
	}
	if err := os.Chmod(backupPath, 0o700); err != nil {
		return result, fmt.Errorf("protect backup directory: %w", err)
	}

	temporary, err := os.CreateTemp(backupPath, ".backup-*")
	if err != nil {
		return result, fmt.Errorf("create temporary backup: %w", err)
	}
	temporaryPath := temporary.Name()
	published := false
	defer func() {
		if !published {
			_ = temporary.Close()
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return result, fmt.Errorf("protect temporary backup: %w", err)
	}

	hash := sha256.New()
	gzipWriter := gzip.NewWriter(io.MultiWriter(temporary, hash))
	tarWriter := tar.NewWriter(gzipWriter)
	fileCount, err := writeConfigArchive(tarWriter, sourcePath)
	if err != nil {
		return result, err
	}
	if err := tarWriter.Close(); err != nil {
		return result, fmt.Errorf("finish backup archive: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return result, fmt.Errorf("compress backup archive: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return result, fmt.Errorf("persist backup archive: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return result, fmt.Errorf("close backup archive: %w", err)
	}

	createdAt := m.now().UTC()
	randomSuffix := strings.TrimPrefix(filepath.Base(temporaryPath), ".backup-")
	archiveName := createdAt.Format("20060102T150405.000000000Z") + "-" + randomSuffix + ".tar.gz"
	archivePath := filepath.Join(backupPath, archiveName)
	if err := os.Rename(temporaryPath, archivePath); err != nil {
		return result, fmt.Errorf("publish backup archive: %w", err)
	}
	published = true

	return BackupResult{
		ApplicationID: applicationID,
		Path:          archivePath,
		SHA256:        hex.EncodeToString(hash.Sum(nil)),
		FileCount:     fileCount,
		CreatedAt:     createdAt,
	}, nil
}

func writeConfigArchive(writer *tar.Writer, sourcePath string) (int, error) {
	fileCount := 0
	err := filepath.WalkDir(sourcePath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("config backup refuses symlink %q", path)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("config backup refuses special file %q", path)
		}

		relativePath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relativePath)
		if info.IsDir() {
			header.Name += "/"
		}
		if err := writer.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		fileCount++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("archive application config: %w", err)
	}
	return fileCount, nil
}
