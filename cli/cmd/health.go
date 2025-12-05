package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/woliveiras/corsarr/internal/i18n"
)

var (
	healthOutputDir string
	healthDetailed  bool
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health of running containers",
	Long: `Check the health status of all configured services.

This command will:
- Verify if containers are running
- Check their health status
- Show uptime and resource usage
- Detect any issues

Example:
  corsarr health
  corsarr health --output /path/to/compose
  corsarr health --detailed`,
	Run: func(cmd *cobra.Command, args []string) {
		t := GetTranslator()

		if err := runHealthCheck(t); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: %v\n", t.T("errors.health_check_failed"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.Flags().StringVarP(&healthOutputDir, "output", "o", ".", "Directory with docker-compose.yml")
	healthCmd.Flags().BoolVarP(&healthDetailed, "detailed", "d", false, "Show detailed container information")
}

func runHealthCheck(t *i18n.I18n) error {
	// Check if Docker is available
	if !isDockerAvailable() {
		return fmt.Errorf("%s", t.T("errors.docker_not_found"))
	}

	// Check if docker-compose.yml exists
	composePath := healthOutputDir + "/docker-compose.yml"
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("%s: %s", t.T("errors.compose_not_found"), composePath)
	}

	fmt.Printf("🏥 %s\n", t.T("health.checking_services"))
	fmt.Printf("📂 %s: %s\n\n", t.T("health.directory"), healthOutputDir)

	// Get container information
	containers, err := getContainerStatus(healthOutputDir)
	if err != nil {
		return fmt.Errorf("%s: %w", t.T("errors.failed_to_get_status"), err)
	}

	if len(containers) == 0 {
		fmt.Printf("ℹ️  %s\n", t.T("health.no_containers"))
		fmt.Printf("💡 %s: docker compose up -d\n", t.T("health.start_hint"))
		return nil
	}

	// Display results
	displayHealthStatus(t, containers)

	// Summary
	running := 0
	unhealthy := 0
	stopped := 0

	for _, c := range containers {
		switch c.Status {
		case "running":
			running++
		case "exited", "dead":
			stopped++
		}
		if c.Health == "unhealthy" {
			unhealthy++
		}
	}

	fmt.Println()
	fmt.Println("═════════════════════════════════════════")
	fmt.Printf("📊 %s\n", t.T("health.summary"))
	fmt.Println("═════════════════════════════════════════")
	fmt.Printf("✅ %s: %d\n", t.T("health.running"), running)
	if stopped > 0 {
		fmt.Printf("⏹️  %s: %d\n", t.T("health.stopped"), stopped)
	}
	if unhealthy > 0 {
		fmt.Printf("❌ %s: %d\n", t.T("health.unhealthy"), unhealthy)
	}
	fmt.Printf("📦 %s: %d\n", t.T("health.total"), len(containers))

	// Show commands for stopped/unhealthy containers
	if stopped > 0 || unhealthy > 0 {
		fmt.Println()
		fmt.Printf("💡 %s:\n", t.T("health.suggested_actions"))
		if stopped > 0 {
			fmt.Printf("   • %s: cd %s && docker compose up -d\n", t.T("health.start_containers"), healthOutputDir)
		}
		if unhealthy > 0 {
			fmt.Printf("   • %s: docker compose logs -f\n", t.T("health.check_logs"))
		}
	}

	return nil
}

type ContainerInfo struct {
	Name    string
	Status  string
	Health  string
	Uptime  string
	CPU     string
	Memory  string
	Ports   []string
}

func isDockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info")
	return cmd.Run() == nil
}

func getContainerStatus(dir string) ([]ContainerInfo, error) {
	ctx := context.Background()

	// List containers for this project
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", dir+"/docker-compose.yml", "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		// Try fallback without json format
		return getContainerStatusFallback(dir)
	}

	var containers []ContainerInfo

	// Parse JSON output (Docker Compose v2 format)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Simple parsing (Docker Compose ps format may vary)
		info := parseContainerLine(line)
		if info != nil {
			containers = append(containers, *info)
		}
	}

	// Get additional details for each container if detailed mode
	if healthDetailed {
		for i := range containers {
			enrichContainerInfo(&containers[i])
		}
	}

	return containers, nil
}

func getContainerStatusFallback(dir string) ([]ContainerInfo, error) {
	ctx := context.Background()

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", dir+"/docker-compose.yml", "ps")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var containers []ContainerInfo
	lines := strings.Split(string(output), "\n")

	// Skip header
	for i, line := range lines {
		if i < 1 || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			containers = append(containers, ContainerInfo{
				Name:   fields[0],
				Status: fields[1],
			})
		}
	}

	return containers, nil
}

func parseContainerLine(line string) *ContainerInfo {
	// Basic parsing - Docker Compose output format
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	return &ContainerInfo{
		Name:   fields[0],
		Status: fields[1],
	}
}

func enrichContainerInfo(info *ContainerInfo) {
	ctx := context.Background()

	// Get container stats
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "{{.CPUPerc}}\t{{.MemUsage}}", info.Name)
	output, err := cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), "\t")
		if len(parts) >= 2 {
			info.CPU = parts[0]
			info.Memory = parts[1]
		}
	}

	// Get container inspect for more details
	cmd = exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Status}}\t{{.State.Health.Status}}", info.Name)
	output, err = cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), "\t")
		if len(parts) >= 1 {
			info.Status = parts[0]
		}
		if len(parts) >= 2 && parts[1] != "<no value>" {
			info.Health = parts[1]
		}
	}
}

func getProjectName(dir string) string {
	// Try to extract project name from .env or use directory name
	envPath := dir + "/.env"
	if data, err := os.ReadFile(envPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "COMPOSE_PROJECT_NAME=") {
				return strings.TrimPrefix(line, "COMPOSE_PROJECT_NAME=")
			}
		}
	}

	// Fallback to directory name
	parts := strings.Split(strings.TrimRight(dir, "/"), "/")
	return parts[len(parts)-1]
}

func displayHealthStatus(t *i18n.I18n, containers []ContainerInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Header
	if healthDetailed {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.T("health.container"),
			t.T("health.status"),
			t.T("health.health"),
			t.T("health.cpu"),
			t.T("health.memory"))
	} else {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			t.T("health.container"),
			t.T("health.status"),
			t.T("health.health"))
	}

	fmt.Fprintf(w, "%s\t%s\t%s",
		strings.Repeat("-", 20),
		strings.Repeat("-", 15),
		strings.Repeat("-", 15))

	if healthDetailed {
		fmt.Fprintf(w, "\t%s\t%s", strings.Repeat("-", 10), strings.Repeat("-", 15))
	}
	fmt.Fprintf(w, "\n")

	// Rows
	for _, c := range containers {
		statusIcon := getStatusIcon(c.Status)
		healthIcon := getHealthIcon(c.Health)

		if healthDetailed {
			cpu := c.CPU
			if cpu == "" {
				cpu = "N/A"
			}
			mem := c.Memory
			if mem == "" {
				mem = "N/A"
			}

			fmt.Fprintf(w, "%s\t%s %s\t%s %s\t%s\t%s\n",
				c.Name,
				statusIcon, c.Status,
				healthIcon, getHealthText(c.Health),
				cpu,
				mem)
		} else {
			fmt.Fprintf(w, "%s\t%s %s\t%s %s\n",
				c.Name,
				statusIcon, c.Status,
				healthIcon, getHealthText(c.Health))
		}
	}

	w.Flush()
}

func getStatusIcon(status string) string {
	switch status {
	case "running":
		return "✅"
	case "exited", "dead":
		return "❌"
	case "paused":
		return "⏸️"
	case "restarting":
		return "🔄"
	default:
		return "❓"
	}
}

func getHealthIcon(health string) string {
	switch health {
	case "healthy":
		return "💚"
	case "unhealthy":
		return "❤️"
	case "starting":
		return "🟡"
	default:
		return "⚪"
	}
}

func getHealthText(health string) string {
	if health == "" {
		return "N/A"
	}
	return health
}
