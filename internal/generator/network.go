package generator

import (
	"fmt"

	"github.com/woliveiras/corsarr/internal/services"
)

// NetworkConfig holds network configuration for docker-compose
type NetworkConfig struct {
	Name   string
	Driver string
}

// GetNetworkConfig returns the network configuration based on VPN mode
func GetNetworkConfig(vpnMode bool) *NetworkConfig {
	if vpnMode {
		// VPN mode doesn't need custom network (uses service:gluetun)
		return nil
	}

	return &NetworkConfig{
		Name:   "media",
		Driver: "bridge",
	}
}

// ConfigureServiceNetworking adjusts service networking based on VPN mode
func ConfigureServiceNetworking(service *services.Service, vpnMode bool) error {
	// If service requires VPN but VPN is not enabled, error
	if service.RequiresVPN && !vpnMode {
		return fmt.Errorf("service %s requires VPN but VPN mode is disabled", service.ID)
	}

	if vpnMode {
		// For VPN mode, services (except Gluetun) use network_mode: "service:gluetun"
		// All services can work with VPN except those that require NOT having VPN
		if service.Category != services.CategoryVPN {
			// Service is compatible if it's not marked as incompatible
			if !service.IsCompatibleWithVPN(true) {
				return fmt.Errorf("service %s is not compatible with VPN mode", service.ID)
			}
		}
	} else {
		// For bridge mode, ensure service has proper network configuration
		if len(service.Network.BridgeMode.Networks) == 0 {
			return fmt.Errorf("service %s has no bridge network configuration", service.ID)
		}
	}

	return nil
}

// ValidateNetworkConfiguration validates network configuration for all services
func ValidateNetworkConfiguration(selectedServices []*services.Service, vpnMode bool) error {
	for _, service := range selectedServices {
		if err := ConfigureServiceNetworking(service, vpnMode); err != nil {
			return err
		}
	}
	return nil
}

// GetExposedPorts returns all exposed ports for VPN mode
// In VPN mode, all service ports must be exposed through Gluetun
func GetExposedPorts(selectedServices []*services.Service, vpnMode bool) []services.PortMapping {
	if !vpnMode {
		return nil // In bridge mode, each service exposes its own ports
	}

	var ports []services.PortMapping
	for _, service := range selectedServices {
		// Skip Gluetun itself
		if service.Category == services.CategoryVPN {
			continue
		}

		// Add all service ports
		if service.Ports != nil {
			ports = append(ports, service.Ports...)
		}
	}

	return ports
}
