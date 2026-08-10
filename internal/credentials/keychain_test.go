package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSecretRedactsFormattingAndJSON(t *testing.T) {
	secret := NewSecret("do-not-print-this")
	rendered := fmt.Sprintf("%v %#v", secret, secret)
	if strings.Contains(rendered, secret.Reveal()) {
		t.Fatalf("expected formatted secret redacted, got %q", rendered)
	}
	encoded, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("encode secret: %v", err)
	}
	if strings.Contains(string(encoded), secret.Reveal()) {
		t.Fatalf("expected JSON secret redacted, got %s", encoded)
	}
}

func TestDarwinKeychainStoresAndLoadsAllowlistedCredential(t *testing.T) {
	runner := &recordingKeychainRunner{results: []keychainResult{
		{},
		{output: "generated-password\n"},
	}}
	store := newDarwinKeychain(runner)
	secret := NewSecret("generated-password")

	if err := store.Save(context.Background(), KeyQBitTorrentPassword, secret); err != nil {
		t.Fatalf("save keychain secret: %v", err)
	}
	loaded, err := store.Load(context.Background(), KeyQBitTorrentPassword)
	if err != nil {
		t.Fatalf("load keychain secret: %v", err)
	}
	if loaded.Reveal() != secret.Reveal() {
		t.Fatal("expected exact stored secret")
	}
	want := []keychainCall{
		{path: darwinSecurityPath, args: []string{
			"add-generic-password", "-a", "qbittorrent", "-s", keychainService,
			"-w", "generated-password", "-U",
		}},
		{path: darwinSecurityPath, args: []string{
			"find-generic-password", "-a", "qbittorrent", "-s", keychainService, "-w",
		}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected keychain calls\nwant: %#v\n got: %#v", want, runner.calls)
	}
}

func TestDarwinKeychainMapsMissingCredential(t *testing.T) {
	runner := &recordingKeychainRunner{results: []keychainResult{{
		output: "security: SecKeychainSearchCopyNext: The specified item could not be found in the keychain.",
		err:    errors.New("exit status 44"),
	}}}
	store := newDarwinKeychain(runner)

	_, err := store.Load(context.Background(), KeyQBitTorrentPassword)
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("expected missing credential error, got %v", err)
	}
}

func TestDarwinKeychainUsesDedicatedJellyfinAccount(t *testing.T) {
	runner := &recordingKeychainRunner{}
	store := newDarwinKeychain(runner)

	if err := store.Save(
		context.Background(),
		KeyJellyfinPassword,
		NewSecret("generated-password"),
	); err != nil {
		t.Fatalf("save Jellyfin credential: %v", err)
	}
	if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0].args, []string{
		"add-generic-password", "-a", "jellyfin", "-s", keychainService,
		"-w", "generated-password", "-U",
	}) {
		t.Fatalf("unexpected Jellyfin keychain call %#v", runner.calls)
	}
}

func TestDarwinKeychainRejectsUnknownKeyWithoutCommand(t *testing.T) {
	runner := &recordingKeychainRunner{}
	store := newDarwinKeychain(runner)

	if err := store.Save(context.Background(), Key("../foreign"), NewSecret("secret")); err == nil {
		t.Fatal("expected unknown key to be rejected")
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no keychain command, got %v", runner.calls)
	}
}

type keychainCall struct {
	path string
	args []string
}

type keychainResult struct {
	output string
	err    error
}

type recordingKeychainRunner struct {
	calls   []keychainCall
	results []keychainResult
}

func (r *recordingKeychainRunner) Run(_ context.Context, path string, args ...string) (string, error) {
	r.calls = append(r.calls, keychainCall{path: path, args: append([]string(nil), args...)})
	if len(r.results) == 0 {
		return "", nil
	}
	result := r.results[0]
	r.results = r.results[1:]
	return result.output, result.err
}
