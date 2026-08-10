package provisioning

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestARRCredentialReaderLoadsAPIKeyFromFixedConfig(t *testing.T) {
	rootPath := t.TempDir()
	configDirectory := filepath.Join(rootPath, "config", "radarr")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	const key = "0123456789abcdef0123456789abcdef"
	config := "<Config><BindAddress>*</BindAddress><ApiKey>" + key + "</ApiKey></Config>"
	if err := os.WriteFile(filepath.Join(configDirectory, "config.xml"), []byte(config), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	credential, err := NewARRCredentialReader().Read(rootPath, "radarr")
	if err != nil {
		t.Fatalf("read API credential: %v", err)
	}
	if credential.Reveal() != key {
		t.Fatal("expected exact API key")
	}
	if rendered := fmt.Sprintf("%v %#v", credential, credential); strings.Contains(rendered, key) {
		t.Fatalf("expected formatted credential to be redacted, got %q", rendered)
	}
	encoded, err := json.Marshal(credential)
	if err != nil {
		t.Fatalf("encode credential: %v", err)
	}
	if strings.Contains(string(encoded), key) {
		t.Fatalf("expected JSON credential to be redacted, got %s", encoded)
	}
}

func TestARRCredentialReaderRejectsUnknownOrUnsafeApplication(t *testing.T) {
	reader := NewARRCredentialReader()
	for _, applicationID := range []string{"../radarr", "jellyfin"} {
		if _, err := reader.Read(t.TempDir(), applicationID); err == nil {
			t.Fatalf("expected application %q to be rejected", applicationID)
		}
	}
}

func TestARRCredentialReaderRejectsSymlinkedConfig(t *testing.T) {
	rootPath := t.TempDir()
	configDirectory := filepath.Join(rootPath, "config", "sonarr")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "secret.xml")
	if err := os.WriteFile(outside, []byte("<Config><ApiKey>0123456789abcdef0123456789abcdef</ApiKey></Config>"), 0o600); err != nil {
		t.Fatalf("write external config: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(configDirectory, "config.xml")); err != nil {
		t.Fatalf("create config symlink: %v", err)
	}

	if _, err := NewARRCredentialReader().Read(rootPath, "sonarr"); err == nil {
		t.Fatal("expected symlinked config to be rejected")
	}
}

func TestARRCredentialReaderRejectsMalformedKey(t *testing.T) {
	rootPath := t.TempDir()
	configDirectory := filepath.Join(rootPath, "config", "prowlarr")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(configDirectory, "config.xml"),
		[]byte("<Config><ApiKey>not-a-key</ApiKey></Config>"),
		0o600,
	); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	if _, err := NewARRCredentialReader().Read(rootPath, "prowlarr"); err == nil {
		t.Fatal("expected malformed API key to be rejected")
	}
}
