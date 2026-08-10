package provisioning

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const maxARRConfigSize = 1024 * 1024

var (
	arrAPIKeyPattern = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	arrApplications  = map[string]struct{}{
		"lidarr":   {},
		"prowlarr": {},
		"radarr":   {},
		"sonarr":   {},
	}
)

// APIKey keeps secret material out of default formatting and JSON payloads.
// Reveal is intentionally explicit and should only be used for a local API request.
type APIKey struct {
	value string
}

func (k APIKey) Reveal() string { return k.value }
func (APIKey) String() string   { return "[REDACTED]" }
func (APIKey) GoString() string { return "provisioning.APIKey{[REDACTED]}" }
func (APIKey) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

type ARRCredentialReader struct{}

func NewARRCredentialReader() *ARRCredentialReader {
	return &ARRCredentialReader{}
}

// Read opens only <root>/config/<known-arr>/config.xml and extracts ApiKey.
func (r *ARRCredentialReader) Read(rootPath string, applicationID string) (APIKey, error) {
	if _, allowed := arrApplications[applicationID]; !allowed {
		return APIKey{}, fmt.Errorf("application does not expose a supported Arr API: %s", applicationID)
	}
	if !filepath.IsAbs(rootPath) {
		return APIKey{}, fmt.Errorf("Corsarr root path must be absolute")
	}

	configDirectory := filepath.Join(rootPath, "config", applicationID)
	directoryInfo, err := os.Lstat(configDirectory)
	if err != nil {
		return APIKey{}, fmt.Errorf("inspect application config directory: %w", err)
	}
	if !directoryInfo.IsDir() || directoryInfo.Mode()&os.ModeSymlink != 0 {
		return APIKey{}, fmt.Errorf("application config path is not a regular directory")
	}

	configPath := filepath.Join(configDirectory, "config.xml")
	configInfo, err := os.Lstat(configPath)
	if err != nil {
		return APIKey{}, fmt.Errorf("inspect application config: %w", err)
	}
	if !configInfo.Mode().IsRegular() || configInfo.Mode()&os.ModeSymlink != 0 {
		return APIKey{}, fmt.Errorf("application config is not a regular file")
	}
	if configInfo.Size() > maxARRConfigSize {
		return APIKey{}, fmt.Errorf("application config exceeds size limit")
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return APIKey{}, fmt.Errorf("open application config: %w", err)
	}
	defer configFile.Close()
	contents, err := io.ReadAll(io.LimitReader(configFile, maxARRConfigSize+1))
	if err != nil {
		return APIKey{}, fmt.Errorf("read application config: %w", err)
	}
	if len(contents) > maxARRConfigSize {
		return APIKey{}, fmt.Errorf("application config exceeds size limit")
	}

	var config struct {
		APIKey string `xml:"ApiKey"`
	}
	if err := xml.Unmarshal(contents, &config); err != nil {
		return APIKey{}, fmt.Errorf("decode application config: %w", err)
	}
	key := strings.TrimSpace(config.APIKey)
	if !arrAPIKeyPattern.MatchString(key) {
		return APIKey{}, fmt.Errorf("application config contains an invalid API key")
	}
	return APIKey{value: key}, nil
}
