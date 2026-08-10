package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type State string

const MinimumAvailableBytes uint64 = 10 * 1024 * 1024 * 1024

const (
	StateReady    State = "ready"
	StateInvalid  State = "invalid"
	StateCanceled State = "canceled"
)

type Status struct {
	Path            string `json:"path"`
	State           State  `json:"state"`
	Writable        bool   `json:"writable"`
	Hardlinks       bool   `json:"hardlinks"`
	AvailableBytes  uint64 `json:"availableBytes,omitempty"`
	RequiredBytes   uint64 `json:"requiredBytes"`
	TechnicalDetail string `json:"technicalDetail,omitempty"`
}

type Inspector struct {
	diskBytes func(string) (uint64, error)
}

func NewInspector() *Inspector {
	return &Inspector{diskBytes: availableDiskBytes}
}

// Inspect validates a user-selected directory and removes all probe artifacts.
func (i *Inspector) Inspect(path string) Status {
	status := Status{Path: normalizedPath(path), RequiredBytes: MinimumAvailableBytes}
	if status.Path == "" {
		status.State = StateInvalid
		status.TechnicalDetail = "storage path is empty"
		return status
	}

	pathInfo, err := os.Stat(status.Path)
	if err != nil {
		status.State = StateInvalid
		status.TechnicalDetail = boundedStorageDetail(err)
		return status
	}
	if !pathInfo.IsDir() {
		status.State = StateInvalid
		status.TechnicalDetail = "selected storage path is not a directory"
		return status
	}

	probeFile, err := os.CreateTemp(status.Path, ".corsarr-storage-check-*")
	if err != nil {
		status.State = StateInvalid
		status.TechnicalDetail = boundedStorageDetail(err)
		return status
	}
	probePath := probeFile.Name()
	linkPath := probePath + ".hardlink"
	defer func() {
		_ = os.Remove(linkPath)
		_ = os.Remove(probePath)
	}()

	if err := probeFile.Close(); err != nil {
		status.State = StateInvalid
		status.TechnicalDetail = boundedStorageDetail(err)
		return status
	}
	status.Writable = true

	if err := os.Link(probePath, linkPath); err == nil {
		status.Hardlinks = true
	} else {
		status.TechnicalDetail = boundedStorageDetail(fmt.Errorf("hardlink check failed: %w", err))
	}

	diskBytes := i.diskBytes
	if diskBytes == nil {
		diskBytes = availableDiskBytes
	}
	if availableBytes, err := diskBytes(status.Path); err == nil {
		status.AvailableBytes = availableBytes
		if availableBytes < MinimumAvailableBytes {
			status.State = StateInvalid
			status.TechnicalDetail = fmt.Sprintf(
				"selected storage has %d bytes available; at least %d bytes are required",
				availableBytes,
				MinimumAvailableBytes,
			)
			return status
		}
	} else {
		status.State = StateInvalid
		status.TechnicalDetail = boundedStorageDetail(fmt.Errorf("disk space check failed: %w", err))
		return status
	}

	status.State = StateReady
	return status
}

func normalizedPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(absolutePath)
}

func boundedStorageDetail(err error) string {
	const maximumLength = 500
	detail := strings.TrimSpace(err.Error())
	if len(detail) <= maximumLength {
		return detail
	}
	return detail[:maximumLength] + "…"
}
