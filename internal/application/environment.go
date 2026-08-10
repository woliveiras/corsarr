package application

import (
	"context"

	runtimeenv "github.com/woliveiras/corsarr/internal/runtime"
)

type EnvironmentStatus struct {
	Platform     string            `json:"platform"`
	Architecture string            `json:"architecture"`
	Runtime      runtimeenv.Status `json:"runtime"`
}

type EnvironmentService struct {
	runtimeProbe runtimeenv.Probe
	platform     string
	architecture string
}

func NewEnvironmentService(
	runtimeProbe runtimeenv.Probe,
	platform string,
	architecture string,
) *EnvironmentService {
	return &EnvironmentService{
		runtimeProbe: runtimeProbe,
		platform:     platform,
		architecture: architecture,
	}
}

// Status performs read-only environment diagnostics.
func (s *EnvironmentService) Status(ctx context.Context) EnvironmentStatus {
	return EnvironmentStatus{
		Platform:     s.platform,
		Architecture: s.architecture,
		Runtime:      s.runtimeProbe.Check(ctx),
	}
}
