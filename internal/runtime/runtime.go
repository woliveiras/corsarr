package runtime

import "context"

type Provider string

const ProviderDocker Provider = "docker"

type State string

const (
	StateUnavailable State = "unavailable"
	StateStopped     State = "stopped"
	StateReady       State = "ready"
	StateError       State = "error"
)

// Status is the runtime-neutral diagnostic exposed to application layers.
type Status struct {
	Provider        Provider `json:"provider"`
	State           State    `json:"state"`
	Version         string   `json:"version,omitempty"`
	TechnicalDetail string   `json:"technicalDetail,omitempty"`
}

// Probe performs a read-only runtime health check.
type Probe interface {
	Check(ctx context.Context) Status
}

// CommandRunner is the narrow process boundary used by runtime probes.
type CommandRunner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, name string, args ...string) (string, error)
}
