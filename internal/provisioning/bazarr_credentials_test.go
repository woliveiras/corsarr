package provisioning

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBazarrCredentialReaderLoadsAPIKeyFromFixedConfig(t *testing.T) {
	rootPath := t.TempDir()
	configDirectory := filepath.Join(rootPath, "config", "bazarr", "config")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	const key = "0123456789abcdef0123456789abcdef"
	if err := os.WriteFile(
		filepath.Join(configDirectory, "config.yaml"),
		[]byte("auth:\n  apikey: "+key+"\n"),
		0o600,
	); err != nil {
		t.Fatalf("write config: %v", err)
	}

	credential, err := NewBazarrCredentialReader().Read(rootPath)
	if err != nil {
		t.Fatalf("read Bazarr API credential: %v", err)
	}
	if credential.Reveal() != key {
		t.Fatal("expected exact API key")
	}
}

func TestBazarrCredentialReaderRejectsSymlinkedConfig(t *testing.T) {
	rootPath := t.TempDir()
	configDirectory := filepath.Join(rootPath, "config", "bazarr", "config")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(outside, []byte("auth:\n  apikey: 0123456789abcdef0123456789abcdef\n"), 0o600); err != nil {
		t.Fatalf("write external config: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(configDirectory, "config.yaml")); err != nil {
		t.Fatalf("create config symlink: %v", err)
	}

	if _, err := NewBazarrCredentialReader().Read(rootPath); err == nil {
		t.Fatal("expected symlinked Bazarr config to be rejected")
	}
}

func TestBazarrCredentialReaderRejectsMalformedKey(t *testing.T) {
	rootPath := t.TempDir()
	configDirectory := filepath.Join(rootPath, "config", "bazarr", "config")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(configDirectory, "config.yaml"),
		[]byte("auth:\n  apikey: not-a-key\n"),
		0o600,
	); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	if _, err := NewBazarrCredentialReader().Read(rootPath); err == nil {
		t.Fatal("expected malformed Bazarr API key to be rejected")
	}
}
