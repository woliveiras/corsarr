package execution

import (
	"context"
	"strings"
	"time"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

type linuxProbe struct {
	runner *Runner
	docker containerruntime.Probe
}

func (p linuxProbe) Check(ctx context.Context) containerruntime.Status {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	status := p.docker.Check(ctx)
	if status.State != containerruntime.StateReady {
		return status
	}
	name, err := p.runner.LookPath("docker")
	if err != nil {
		status.State = containerruntime.StateError
		status.TechnicalDetail = err.Error()
		return status
	}
	output, err := p.runner.Run(ctx, name, "info", "--format", "{{.OSType}}")
	if err != nil {
		status.State = containerruntime.StateError
		status.TechnicalDetail = err.Error()
		return status
	}
	if strings.TrimSpace(output) != "linux" {
		status.State = containerruntime.StateError
		status.TechnicalDetail = "Corsarr requires a Docker daemon running Linux containers"
	}
	return status
}
