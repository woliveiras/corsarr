package diagnostics

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Writer interface {
	Write(path string, report Report) error
}

type FileWriter struct{}

func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

func (w *FileWriter) Write(destination string, report Report) error {
	if !filepath.IsAbs(destination) {
		return fmt.Errorf("diagnostic destination must be absolute")
	}
	if info, err := os.Lstat(destination); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return fmt.Errorf("diagnostic destination must be a regular file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect diagnostic destination: %w", err)
	}

	directory := filepath.Dir(destination)
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("inspect diagnostic directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("diagnostic parent is not a directory")
	}

	temporaryFile, err := os.CreateTemp(directory, ".corsarr-diagnostics-*")
	if err != nil {
		return fmt.Errorf("create temporary diagnostics: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporaryFile.Chmod(0o600); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("protect temporary diagnostics: %w", err)
	}

	encoder := json.NewEncoder(temporaryFile)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(report); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("encode diagnostics: %w", err)
	}
	if err := temporaryFile.Sync(); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("sync diagnostics: %w", err)
	}
	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("close diagnostics: %w", err)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return fmt.Errorf("replace diagnostics: %w", err)
	}

	directoryHandle, err := os.Open(directory)
	if err == nil {
		_ = directoryHandle.Sync()
		_ = directoryHandle.Close()
	}
	return nil
}
