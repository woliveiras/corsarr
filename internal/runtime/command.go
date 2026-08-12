package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// OSCommandRunner executes fixed commands selected by a runtime adapter.
type OSCommandRunner struct{}

func (OSCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// RunWithEnvironment executes a fixed adapter command with additional
// process-scoped values. Callers pass secret names to the child command rather
// than placing secret values in its argument list.
func (OSCommandRunner) RunWithEnvironment(
	ctx context.Context,
	environment map[string]string,
	name string,
	args ...string,
) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Env = environmentWithOverrides(os.Environ(), environment)
	output, err := command.CombinedOutput()
	cleanOutput := strings.TrimSpace(string(output))
	if err == nil {
		return cleanOutput, nil
	}
	if cleanOutput == "" {
		return "", err
	}
	return "", fmt.Errorf("%s: %w", cleanOutput, err)
}

func environmentWithOverrides(base []string, overrides map[string]string) []string {
	merged := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		name, _, found := strings.Cut(entry, "=")
		if found {
			if _, overridden := overrides[name]; overridden {
				continue
			}
		}
		merged = append(merged, entry)
	}
	for key, value := range overrides {
		merged = append(merged, key+"="+value)
	}
	return merged
}

func (OSCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	cleanOutput := strings.TrimSpace(string(output))
	if err == nil {
		return cleanOutput, nil
	}
	if cleanOutput == "" {
		return "", err
	}
	return "", fmt.Errorf("%s: %w", cleanOutput, err)
}
