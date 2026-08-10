package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var safeApplicationIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type LayoutStatus struct {
	RootPath    string   `json:"rootPath"`
	Directories []string `json:"directories"`
}

type LayoutPreparer struct{}

func NewLayoutPreparer() *LayoutPreparer {
	return &LayoutPreparer{}
}

// Prepare creates only the Corsarr-owned tree below an existing selected directory.
func (p *LayoutPreparer) Prepare(basePath string, applicationIDs []string) (LayoutStatus, error) {
	baseInfo, err := os.Stat(basePath)
	if err != nil {
		return LayoutStatus{}, fmt.Errorf("inspect storage base: %w", err)
	}
	if !baseInfo.IsDir() {
		return LayoutStatus{}, fmt.Errorf("storage base is not a directory")
	}

	applications, err := validatedLayoutApplications(applicationIDs)
	if err != nil {
		return LayoutStatus{}, err
	}

	rootPath := filepath.Join(basePath, "Corsarr")
	configPath := filepath.Join(rootPath, "config")
	if err := os.MkdirAll(configPath, 0o700); err != nil {
		return LayoutStatus{}, fmt.Errorf("create private config directory: %w", err)
	}
	if err := os.Chmod(configPath, 0o700); err != nil {
		return LayoutStatus{}, fmt.Errorf("protect config directory: %w", err)
	}

	directories := []string{
		"config",
		filepath.Join("media", "downloads", "incomplete"),
		filepath.Join("media", "downloads", "complete"),
		filepath.Join("media", "downloads", "complete", "lidarr"),
		filepath.Join("media", "downloads", "complete", "radarr"),
		filepath.Join("media", "downloads", "complete", "sonarr"),
		filepath.Join("media", "library", "movies"),
		filepath.Join("media", "library", "tv"),
		filepath.Join("media", "library", "music"),
		filepath.Join("media", "library", "books"),
	}
	for _, applicationID := range applications {
		relativePath := filepath.Join("config", applicationID)
		applicationPath := filepath.Join(rootPath, relativePath)
		if err := os.MkdirAll(applicationPath, 0o700); err != nil {
			return LayoutStatus{}, fmt.Errorf("create application config directory: %w", err)
		}
		if err := os.Chmod(applicationPath, 0o700); err != nil {
			return LayoutStatus{}, fmt.Errorf("protect application config directory: %w", err)
		}
		directories = append(directories, relativePath)
	}

	for _, relativePath := range directories {
		if relativePath == "config" || filepath.Dir(relativePath) == "config" {
			continue
		}
		if err := os.MkdirAll(filepath.Join(rootPath, relativePath), 0o755); err != nil {
			return LayoutStatus{}, fmt.Errorf("create media directory: %w", err)
		}
	}

	sort.Strings(directories)
	return LayoutStatus{RootPath: rootPath, Directories: directories}, nil
}

func validatedLayoutApplications(applicationIDs []string) ([]string, error) {
	unique := make(map[string]struct{}, len(applicationIDs))
	for _, applicationID := range applicationIDs {
		if !safeApplicationIDPattern.MatchString(applicationID) {
			return nil, fmt.Errorf("unsafe application ID: %q", applicationID)
		}
		unique[applicationID] = struct{}{}
	}
	applications := make([]string, 0, len(unique))
	for applicationID := range unique {
		applications = append(applications, applicationID)
	}
	sort.Strings(applications)
	return applications, nil
}
