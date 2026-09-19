package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/woliveiras/corsarr/internal/execution"
	"github.com/woliveiras/corsarr/internal/onboarding"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/state"
)

func newRuntimeCommand() *cobra.Command {
	command := &cobra.Command{Use: "runtime", Short: "Select, inspect or prepare a local Docker environment",
		PersistentPreRun: func(*cobra.Command, []string) {}}
	command.AddCommand(&cobra.Command{Use: "status", Short: "Check the selected Docker daemon", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		env, err := loadExecutionEnvironment()
		if err != nil {
			return err
		}
		status := env.Probe.Check(cmd.Context())
		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
			Configuration execution.Config        `json:"configuration"`
			Runtime       containerruntime.Status `json:"runtime"`
		}{env.Selection.Config(), status}); err != nil {
			return err
		}
		if status.State != containerruntime.StateReady {
			return fmt.Errorf("selected runtime is not ready")
		}
		return nil
	}})
	var contextName string
	use := &cobra.Command{Use: "use <existing|docker-engine|docker-desktop>", Short: "Save the environment for CLI and Desktop", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		env, err := loadExecutionEnvironment()
		if err != nil {
			return err
		}
		config := execution.Config{Kind: execution.Kind(args[0]), Context: contextName}
		if err := config.Validate(runtime.GOOS); err != nil {
			return err
		}
		statePath, err := state.DefaultPath()
		if err != nil {
			return err
		}
		if err := prepareExecutionChange(statePath, env.Selection.Config(), config); err != nil {
			return err
		}
		if err := env.Selection.Save(config); err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Execution environment saved. Restart Corsarr Desktop if it is open. No software was installed.")
		return err
	}}
	use.Flags().StringVar(&contextName, "context", "", "Existing local Docker context (required for existing)")
	command.AddCommand(use)
	var yes bool
	install := &cobra.Command{Use: "install", Short: "Install native Docker Engine and Compose on supported Linux systems", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if !yes {
			return fmt.Errorf("review docs/DOCKER_ENGINE.md, then pass --yes to install official Docker APT packages and enable the docker service; existing installations are never replaced")
		}
		installer := onboarding.LinuxEngineInstaller{Runner: containerruntime.OSCommandRunner{}, Platform: runtime.GOOS}
		if err := installer.Install(cmd.Context()); err != nil {
			return err
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Docker Engine and Compose installed. No account permissions were changed. As your regular user, select docker-engine and run runtime status. Your administrator may need to grant Docker socket access.")
		return err
	}}
	install.Flags().BoolVar(&yes, "yes", false, "Authorize repository setup, package installation and Docker service startup")
	command.AddCommand(install)
	prepare := &cobra.Command{Use: "prepare", Short: "Prepare Engine or check an externally managed environment", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		env, err := loadExecutionEnvironment()
		if err != nil {
			return err
		}
		if env.Selection.Config().Kind == execution.Engine && !yes {
			return fmt.Errorf("pass --yes to authorize Docker Engine installation or service startup")
		}
		result, err := env.Prepare(cmd.Context())
		if err != nil {
			return err
		}
		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
			return err
		}
		if !result.Ready {
			return fmt.Errorf("%s", result.Message)
		}
		return nil
	}}
	prepare.Flags().BoolVar(&yes, "yes", false, "Authorize preparing the selected environment")
	command.AddCommand(prepare)
	command.AddCommand(&cobra.Command{Use: "compose [arguments...]", Short: "Run Compose against the selected local Docker environment", DisableFlagParsing: true, RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
		defer cancel()
		output, err := runSelectedDocker(ctx, append([]string{"compose"}, args...)...)
		if output != "" {
			if _, writeErr := fmt.Fprintln(cmd.OutOrStdout(), output); writeErr != nil {
				return writeErr
			}
		}
		return err
	}})
	return command
}

func runSelectedDocker(ctx context.Context, args ...string) (string, error) {
	env, err := loadExecutionEnvironment()
	if err != nil {
		return "", err
	}
	path, err := env.Runner.LookPath("docker")
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	return env.Runner.Run(ctx, path, args...)
}

func loadExecutionEnvironment() (*execution.Environment, error) {
	path, err := execution.DefaultPath()
	if err != nil {
		return nil, err
	}
	selection, err := execution.Open(path, runtime.GOOS)
	if err != nil {
		return nil, err
	}
	return execution.New(selection, containerruntime.OSCommandRunner{}, runtime.GOOS, runtime.GOARCH), nil
}

func prepareExecutionChange(path string, current, next execution.Config) error {
	if current == next {
		return nil
	}
	desktop, err := state.NewFileStore(path).Load()
	if err != nil {
		return err
	}
	if len(desktop.Applications) > 0 || desktop.StoragePath != "" || desktop.OnboardingCompleted {
		return fmt.Errorf("execution environment is locked after storage or applications have been selected; migration of an existing installation is not supported")
	}
	desktop.RuntimeConsentVersion = ""
	desktop.RuntimeConsentAcceptedAt = ""
	return state.NewFileStore(path).Save(desktop)
}

func init() { rootCmd.AddCommand(newRuntimeCommand()) }
