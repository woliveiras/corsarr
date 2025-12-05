package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/woliveiras/corsarr/internal/i18n"
	"github.com/woliveiras/corsarr/internal/services"
)

var (
	checkPortsOutputDir string
	checkPortsSuggest   bool
)

// checkPortsCmd represents the check-ports command
var checkPortsCmd = &cobra.Command{
	Use:   "check-ports",
	Short: "Check for port conflicts",
	Long: `Check which ports are in use and detect potential conflicts.

This command will:
- List all ports used by configured services
- Check if ports are available on the system
- Suggest alternative ports if conflicts are detected
- Show which process is using conflicting ports

Example:
  corsarr check-ports
  corsarr check-ports --output /path/to/compose
  corsarr check-ports --suggest`,
	Run: func(cmd *cobra.Command, args []string) {
		t := GetTranslator()

		if err := runPortCheck(t); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: %v\n", t.T("errors.port_check_failed"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkPortsCmd)

	checkPortsCmd.Flags().StringVarP(&checkPortsOutputDir, "output", "o", ".", "Directory with docker-compose.yml")
	checkPortsCmd.Flags().BoolVarP(&checkPortsSuggest, "suggest", "s", false, "Suggest alternative ports for conflicts")
}

type PortInfo struct {
	Service   string
	Port      int
	Protocol  string
	InUse     bool
	Available bool
	UsedBy    string
}

func runPortCheck(t *i18n.I18n) error {
	fmt.Printf("🔍 %s\n", t.T("ports.checking_ports"))
	fmt.Printf("📂 %s: %s\n\n", t.T("ports.directory"), checkPortsOutputDir)

	// Load service registry
	registry, err := services.NewRegistry()
	if err != nil {
		return fmt.Errorf("%s: %w", t.T("errors.failed_to_load_services"), err)
	}

	// Get configured services from docker-compose.yml
	composePath := checkPortsOutputDir + "/docker-compose.yml"
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("%s: %s", t.T("errors.compose_not_found"), composePath)
	}

	configuredServices, err := getConfiguredServices(composePath)
	if err != nil {
		return fmt.Errorf("%s: %w", t.T("errors.failed_to_parse_compose"), err)
	}

	// Collect all ports from configured services
	var ports []PortInfo

	for _, serviceName := range configuredServices {
		service, err := registry.GetService(serviceName)
		if err != nil || service == nil {
			continue
		}

		for _, portMapping := range service.Ports {
			hostPort, err := strconv.Atoi(portMapping.Host)
			if err != nil {
				continue
			}

			info := PortInfo{
				Service:  service.Name,
				Port:     hostPort,
				Protocol: portMapping.Protocol,
			}

			// Check if port is available
			info.Available = isPortAvailable(hostPort, portMapping.Protocol)
			info.InUse = !info.Available

			if info.InUse {
				info.UsedBy = getProcessUsingPort(hostPort)
			}

			ports = append(ports, info)
		}
	}

	if len(ports) == 0 {
		fmt.Printf("ℹ️  %s\n", t.T("ports.no_ports_configured"))
		return nil
	}

	// Sort by port number
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	// Display results
	displayPortStatus(t, ports)

	// Count conflicts
	conflicts := 0
	for _, p := range ports {
		if p.InUse {
			conflicts++
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("═════════════════════════════════════════")
	fmt.Printf("📊 %s\n", t.T("ports.summary"))
	fmt.Println("═════════════════════════════════════════")
	fmt.Printf("📝 %s: %d\n", t.T("ports.total_ports"), len(ports))
	fmt.Printf("✅ %s: %d\n", t.T("ports.available"), len(ports)-conflicts)

	if conflicts > 0 {
		fmt.Printf("❌ %s: %d\n", t.T("ports.in_use"), conflicts)

		if checkPortsSuggest {
			fmt.Println()
			fmt.Printf("💡 %s:\n", t.T("ports.suggestions"))
			suggestAlternativePorts(t, ports)
		} else {
			fmt.Println()
			fmt.Printf("💡 %s: corsarr check-ports --suggest\n", t.T("ports.suggest_hint"))
		}
	} else {
		fmt.Printf("✅ %s\n", t.T("ports.no_conflicts"))
	}

	return nil
}

func getConfiguredServices(composePath string) ([]string, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, err
	}

	var services []string
	lines := strings.Split(string(data), "\n")
	inServices := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "services:" {
			inServices = true
			continue
		}

		if inServices {
			// Service names are at indentation level 1
			if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
				serviceName := strings.TrimSpace(strings.Split(trimmed, ":")[0])
				if serviceName != "" && !strings.HasPrefix(serviceName, "#") {
					services = append(services, serviceName)
				}
			}

			// Stop when we reach another top-level key
			if len(line) > 0 && !strings.HasPrefix(line, " ") && trimmed != "services:" {
				break
			}
		}
	}

	return services, nil
}

func isPortAvailable(port int, protocol string) bool {
	address := fmt.Sprintf(":%d", port)

	if protocol == "udp" {
		conn, err := net.ListenPacket("udp", address)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}

	// TCP (default)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}

func getProcessUsingPort(port int) string {
	// Try to get process info using lsof (Linux/Mac)
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		pid := strings.TrimSpace(string(output))
		return fmt.Sprintf("PID %s", pid)
	}

	// Try netstat on Windows
	cmd = exec.Command("netstat", "-ano")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		portStr := fmt.Sprintf(":%d", port)
		for _, line := range lines {
			if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
				fields := strings.Fields(line)
				if len(fields) > 4 {
					return fmt.Sprintf("PID %s", fields[len(fields)-1])
				}
			}
		}
	}

	return "Unknown"
}

func displayPortStatus(t *i18n.I18n, ports []PortInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		t.T("ports.service"),
		t.T("ports.port"),
		t.T("ports.protocol"),
		t.T("ports.status"),
		t.T("ports.used_by"))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		strings.Repeat("-", 20),
		strings.Repeat("-", 8),
		strings.Repeat("-", 10),
		strings.Repeat("-", 12),
		strings.Repeat("-", 15))

	// Rows
	for _, p := range ports {
		statusIcon := "✅"
		statusText := t.T("ports.available")
		usedBy := "-"

		if p.InUse {
			statusIcon = "❌"
			statusText = t.T("ports.in_use")
			usedBy = p.UsedBy
		}

		fmt.Fprintf(w, "%s\t%d\t%s\t%s %s\t%s\n",
			p.Service,
			p.Port,
			strings.ToUpper(p.Protocol),
			statusIcon,
			statusText,
			usedBy)
	}

	_ = w.Flush()
}

func suggestAlternativePorts(t *i18n.I18n, ports []PortInfo) {
	for _, p := range ports {
		if !p.InUse {
			continue
		}

		// Find next available port
		alternativePort := findNextAvailablePort(p.Port, p.Protocol)
		if alternativePort > 0 {
			fmt.Printf("   • %s (%d) → %s %d\n",
				p.Service,
				p.Port,
				t.T("ports.use_port"),
				alternativePort)
		}
	}
}

func findNextAvailablePort(startPort int, protocol string) int {
	// Try ports in range [startPort+1, startPort+100]
	for port := startPort + 1; port <= startPort+100; port++ {
		if isPortAvailable(port, protocol) {
			return port
		}
	}
	return 0
}
