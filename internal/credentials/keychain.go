package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	keychainService    = "io.github.woliveiras.corsarr"
	darwinSecurityPath = "/usr/bin/security"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrStoreUnsupported   = errors.New("credential store is not supported on this platform")
)

type Key string

const (
	KeyJellyfinPassword    Key = "jellyfin-password"
	KeyQBitTorrentPassword Key = "qbittorrent-password"
)

var keychainAccounts = map[Key]string{
	KeyJellyfinPassword:    "jellyfin",
	KeyQBitTorrentPassword: "qbittorrent",
}

type Secret struct {
	value string
}

func NewSecret(value string) Secret { return Secret{value: value} }
func (s Secret) Reveal() string     { return s.value }
func (Secret) String() string       { return "[REDACTED]" }
func (Secret) GoString() string     { return "credentials.Secret{[REDACTED]}" }
func (Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

type Store interface {
	Save(ctx context.Context, key Key, secret Secret) error
	Load(ctx context.Context, key Key) (Secret, error)
	Delete(ctx context.Context, key Key) error
}

type keychainCommandRunner interface {
	Run(ctx context.Context, path string, args ...string) (string, error)
}

type osKeychainCommandRunner struct{}

func (osKeychainCommandRunner) Run(
	ctx context.Context,
	path string,
	args ...string,
) (string, error) {
	output, err := exec.CommandContext(ctx, path, args...).CombinedOutput()
	return string(output), err
}

type darwinKeychain struct {
	runner keychainCommandRunner
}

func newDarwinKeychain(runner keychainCommandRunner) *darwinKeychain {
	return &darwinKeychain{runner: runner}
}

func (s *darwinKeychain) Save(ctx context.Context, key Key, secret Secret) error {
	account, err := keychainAccount(key)
	if err != nil {
		return err
	}
	if secret.Reveal() == "" || len(secret.Reveal()) > 256 || strings.ContainsRune(secret.Reveal(), '\x00') {
		return fmt.Errorf("credential secret is invalid")
	}
	_, err = s.runner.Run(
		ctx,
		darwinSecurityPath,
		"add-generic-password",
		"-a", account,
		"-s", keychainService,
		"-w", secret.Reveal(),
		"-U",
	)
	if err != nil {
		return fmt.Errorf("save credential in macOS Keychain: %w", err)
	}
	return nil
}

func (s *darwinKeychain) Load(ctx context.Context, key Key) (Secret, error) {
	account, err := keychainAccount(key)
	if err != nil {
		return Secret{}, err
	}
	output, err := s.runner.Run(
		ctx,
		darwinSecurityPath,
		"find-generic-password",
		"-a", account,
		"-s", keychainService,
		"-w",
	)
	if err != nil {
		detail := strings.ToLower(output + " " + err.Error())
		if strings.Contains(detail, "could not be found") {
			return Secret{}, ErrCredentialNotFound
		}
		return Secret{}, fmt.Errorf("load credential from macOS Keychain: %w", err)
	}
	value := strings.TrimSuffix(strings.TrimSuffix(output, "\n"), "\r")
	if value == "" {
		return Secret{}, ErrCredentialNotFound
	}
	return NewSecret(value), nil
}

func (s *darwinKeychain) Delete(ctx context.Context, key Key) error {
	account, err := keychainAccount(key)
	if err != nil {
		return err
	}
	output, err := s.runner.Run(
		ctx,
		darwinSecurityPath,
		"delete-generic-password",
		"-a", account,
		"-s", keychainService,
	)
	if err != nil && !strings.Contains(strings.ToLower(output+" "+err.Error()), "could not be found") {
		return fmt.Errorf("delete credential from macOS Keychain: %w", err)
	}
	return nil
}

func keychainAccount(key Key) (string, error) {
	account, allowed := keychainAccounts[key]
	if !allowed {
		return "", fmt.Errorf("credential key is not allowlisted: %s", key)
	}
	return account, nil
}

type unsupportedStore struct{}

func (unsupportedStore) Save(context.Context, Key, Secret) error { return ErrStoreUnsupported }
func (unsupportedStore) Load(context.Context, Key) (Secret, error) {
	return Secret{}, ErrStoreUnsupported
}
func (unsupportedStore) Delete(context.Context, Key) error { return ErrStoreUnsupported }

func NewPlatformStore() Store {
	if runtime.GOOS == "darwin" {
		return newDarwinKeychain(osKeychainCommandRunner{})
	}
	return unsupportedStore{}
}
