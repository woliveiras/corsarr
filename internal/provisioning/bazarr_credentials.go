package provisioning

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const maxBazarrConfigSize = 1024 * 1024

type BazarrCredentialReader struct{}

func NewBazarrCredentialReader() *BazarrCredentialReader {
	return &BazarrCredentialReader{}
}

// Read opens only <root>/config/bazarr/config/config.yaml and extracts auth.apikey.
func (r *BazarrCredentialReader) Read(rootPath string) (APIKey, error) {
	if !filepath.IsAbs(rootPath) {
		return APIKey{}, fmt.Errorf("corsarr root path must be absolute")
	}

	configDirectory := filepath.Join(rootPath, "config", "bazarr", "config")
	directoryInfo, err := os.Lstat(configDirectory)
	if err != nil {
		return APIKey{}, fmt.Errorf("inspect Bazarr config directory: %w", err)
	}
	if !directoryInfo.IsDir() || directoryInfo.Mode()&os.ModeSymlink != 0 {
		return APIKey{}, fmt.Errorf("bazarr config path is not a regular directory")
	}

	configPath := filepath.Join(configDirectory, "config.yaml")
	configInfo, err := os.Lstat(configPath)
	if err != nil {
		return APIKey{}, fmt.Errorf("inspect Bazarr config: %w", err)
	}
	if !configInfo.Mode().IsRegular() || configInfo.Mode()&os.ModeSymlink != 0 {
		return APIKey{}, fmt.Errorf("bazarr config is not a regular file")
	}
	if configInfo.Size() > maxBazarrConfigSize {
		return APIKey{}, fmt.Errorf("bazarr config exceeds size limit")
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return APIKey{}, fmt.Errorf("open Bazarr config: %w", err)
	}
	contents, readErr := io.ReadAll(io.LimitReader(configFile, maxBazarrConfigSize+1))
	closeErr := configFile.Close()
	if readErr != nil {
		return APIKey{}, fmt.Errorf("read Bazarr config: %w", readErr)
	}
	if closeErr != nil {
		return APIKey{}, fmt.Errorf("close Bazarr config: %w", closeErr)
	}
	if len(contents) > maxBazarrConfigSize {
		return APIKey{}, fmt.Errorf("bazarr config exceeds size limit")
	}

	var config struct {
		Auth struct {
			APIKey string `yaml:"apikey"`
		} `yaml:"auth"`
	}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return APIKey{}, fmt.Errorf("decode Bazarr config: %w", err)
	}
	key := strings.TrimSpace(config.Auth.APIKey)
	if !arrAPIKeyPattern.MatchString(key) {
		return APIKey{}, fmt.Errorf("bazarr config contains an invalid API key")
	}
	return APIKey{value: key}, nil
}
