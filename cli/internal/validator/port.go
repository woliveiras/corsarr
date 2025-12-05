package validator

import (
	"fmt"
	"net"
)

// PortValidator checks for port conflicts
type PortValidator struct {
	config *Config
}

// NewPortValidator creates a new port validator
func NewPortValidator(config *Config) *PortValidator {
	return &PortValidator{config: config}
}

// Validate checks for port conflicts
func (pv *PortValidator) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Collect all ports from selected services
	portMap := make(map[string][]string) // port -> service names

	for _, service := range pv.config.Services {
		// Get ports based on VPN mode
		var ports []string
		if pv.config.VPNEnabled {
			// In VPN mode, only Gluetun exposes ports
			if service.ID == "gluetun" {
				// Gluetun will expose all other services' ports
				for _, s := range pv.config.Services {
					if s.ID != "gluetun" && len(s.Ports) > 0 {
						for _, portMapping := range s.Ports {
							ports = append(ports, portMapping.Host)
						}
					}
				}
			}
		} else {
			// Bridge mode - each service exposes its own ports
			for _, portMapping := range service.Ports {
				ports = append(ports, portMapping.Host)
			}
		}

		// Check each port
		for _, port := range ports {
			portMap[port] = append(portMap[port], service.Name)
		}
	}

	// Check for conflicts (same port used by multiple services)
	for port, serviceNames := range portMap {
		if len(serviceNames) > 1 {
			result.AddError(
				"ports",
				fmt.Sprintf("Port %s is used by multiple services: %v", port, serviceNames),
				SeverityError,
			)
		}

		// Check if port is already in use on the system
		if pv.isPortInUse(port) {
			result.AddError(
				"ports",
				fmt.Sprintf("Port %s is already in use on the system", port),
				SeverityWarning,
			)
		}
	}

	return result
}

// isPortInUse checks if a port is currently in use on localhost
func (pv *PortValidator) isPortInUse(port string) bool {
	address := fmt.Sprintf("localhost:%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		// Port is in use
		return true
	}
	defer listener.Close()
	return false
}

// GetPortConflicts returns a map of ports to conflicting services
func GetPortConflicts(config *Config) map[string][]string {
	conflicts := make(map[string][]string)
	portMap := make(map[string][]string)

	for _, service := range config.Services {
		var ports []string
		if config.VPNEnabled && service.ID != "gluetun" {
			// In VPN mode, services don't expose ports directly
			continue
		}

		for _, portMapping := range service.Ports {
			ports = append(ports, portMapping.Host)
		}

		for _, port := range ports {
			portMap[port] = append(portMap[port], service.Name)
		}
	}

	// Find conflicts
	for port, serviceNames := range portMap {
		if len(serviceNames) > 1 {
			conflicts[port] = serviceNames
		}
	}

	return conflicts
}
