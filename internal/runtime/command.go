package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// OSCommandRunner executes fixed commands selected by a runtime adapter.
type OSCommandRunner struct{}

func (OSCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
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
