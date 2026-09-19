// Package execution selects the Docker installation used by both Corsarr clients.
package execution

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

type Kind string

const (
	Desktop  Kind = "docker-desktop"
	Existing Kind = "existing"
	Engine   Kind = "docker-engine"
)

// Config contains no credentials. Context is an explicit, local Docker context.
type Config struct {
	Kind    Kind   `json:"kind"`
	Context string `json:"context,omitempty"`
}

var contextName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)

func (c Config) Validate(platform string) error {
	switch c.Kind {
	case Desktop:
		if c.Context != "" {
			return fmt.Errorf("docker Desktop does not accept a custom context")
		}
	case Existing:
		if !contextName.MatchString(c.Context) {
			return fmt.Errorf("select a valid local Docker context")
		}
	case Engine:
		if platform != "linux" {
			return fmt.Errorf("native Docker Engine installation requires Linux; use an existing local runtime on this platform")
		}
		if c.Context != "" {
			return fmt.Errorf("native Docker Engine uses the local system socket")
		}
	default:
		return fmt.Errorf("unsupported execution kind: %q", c.Kind)
	}
	return nil
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Corsarr", "execution.json"), nil
}

// Selection keeps all adapters in a process on the same configuration. Callers
// serialize Save with application operations; configuration changes are only
// allowed before applications have been selected.
type Selection struct {
	mu             sync.RWMutex
	path, platform string
	config         Config
}

func Open(path, platform string) (*Selection, error) {
	s := &Selection{path: path, platform: platform, config: Config{Kind: Desktop}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	s.config = Config{}
	if err := json.Unmarshal(data, &s.config); err != nil {
		return nil, fmt.Errorf("read execution configuration: %w", err)
	}
	if err := s.config.Validate(platform); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Selection) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Selection) Save(c Config) error {
	if err := c.Validate(s.platform); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".execution-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if err := f.Chmod(0600); err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	s.config = c
	return nil
}
