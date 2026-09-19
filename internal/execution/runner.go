package execution

import (
	"context"
	"fmt"
	"strings"

	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
)

type EnvironmentRunner interface {
	containerruntime.CommandRunner
	RunWithEnvironment(context.Context, map[string]string, string, ...string) (string, error)
}

// Runner scopes every Docker operation (including quality jobs) to the selected
// local endpoint. It never changes Docker's global current context.
type Runner struct {
	Selection *Selection
	Base      EnvironmentRunner
}

func (r *Runner) LookPath(name string) (string, error) { return r.Base.LookPath(name) }
func (r *Runner) Run(ctx context.Context, name string, args ...string) (string, error) {
	return r.RunWithEnvironment(ctx, nil, name, args...)
}

func (r *Runner) RunWithEnvironment(ctx context.Context, extra map[string]string, name string, args ...string) (string, error) {
	c := r.Selection.Config()
	if c.Kind == Desktop {
		return r.Base.RunWithEnvironment(ctx, extra, name, args...)
	}
	env := map[string]string{}
	for k, v := range extra {
		env[k] = v
	}
	for _, key := range []string{"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH"} {
		env[key] = ""
	}
	if c.Kind == Engine {
		args = append([]string{"--host", "unix:///var/run/docker.sock"}, args...)
	} else {
		// Resolve each time: an external edit to a named context must not redirect
		// file mounts or destructive operations to a remote machine.
		endpoint, err := r.Base.RunWithEnvironment(ctx, env, name, "context", "inspect", c.Context, "--format", "{{.Endpoints.docker.Host}}")
		if err != nil {
			return "", fmt.Errorf("inspect selected Docker context: %w", err)
		}
		endpoint = strings.TrimSpace(endpoint)
		if !strings.HasPrefix(endpoint, "unix:///") && !strings.HasPrefix(endpoint, "npipe:////./pipe/") {
			return "", fmt.Errorf("selected Docker context is not local; remote storage and service access are not supported in this release")
		}
		// Pin the resolved endpoint rather than looking up the context again.
		args = append([]string{"--host", endpoint}, args...)
	}
	return r.Base.RunWithEnvironment(ctx, env, name, args...)
}
