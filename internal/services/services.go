package services

// PortMapping represents a port mapping between host and container
type PortMapping struct {
	Host      string `yaml:"host"`
	Container string `yaml:"container"`
	Protocol  string `yaml:"protocol"` // tcp, udp
}

// VolumeMapping represents a volume mapping between host and container
type VolumeMapping struct {
	Host      string `yaml:"host"`
	Container string `yaml:"container"`
	ReadOnly  bool   `yaml:"read_only,omitempty"`
}

// NetworkConfig represents network configuration for different modes
type NetworkConfig struct {
	VPNMode    VPNModeConfig    `yaml:"vpn_mode"`
	BridgeMode BridgeModeConfig `yaml:"bridge_mode"`
}

// VPNModeConfig represents network configuration for VPN mode
type VPNModeConfig struct {
	NetworkMode string `yaml:"network_mode"`
}

// BridgeModeConfig represents network configuration for bridge mode
type BridgeModeConfig struct {
	Hostname string   `yaml:"hostname"`
	Networks []string `yaml:"networks"`
}

// Service represents a Docker service configuration
type Service struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Category    ServiceCategory   `yaml:"category"`
	Description string            `yaml:"description"`
	Image       string            `yaml:"image"`
	ContainerName string          `yaml:"container_name"`
	Ports       []PortMapping     `yaml:"ports,omitempty"`
	Volumes     []VolumeMapping   `yaml:"volumes"`
	Environment []string          `yaml:"environment,omitempty"`
	Devices     []string          `yaml:"devices,omitempty"`
	CapAdd      []string          `yaml:"cap_add,omitempty"`
	Network     NetworkConfig     `yaml:"network"`
	Restart     string            `yaml:"restart"`
	SupportsVPN bool              `yaml:"supports_vpn"`
	RequiresVPN bool              `yaml:"requires_vpn"`
	Dependencies []string         `yaml:"dependencies,omitempty"`
	Optional    bool              `yaml:"optional"`
}

// GetTranslationKey returns the i18n key for the service
func (s *Service) GetTranslationKey() string {
	return "services." + s.ID
}

// GetNameKey returns the i18n key for the service name
func (s *Service) GetNameKey() string {
	return s.GetTranslationKey() + ".name"
}

// GetDescriptionKey returns the i18n key for the service description
func (s *Service) GetDescriptionKey() string {
	return "services_" + s.ID + "_description"
}

// IsCompatibleWithVPN checks if service can run with VPN
func (s *Service) IsCompatibleWithVPN(vpnEnabled bool) bool {
	if s.RequiresVPN {
		return vpnEnabled
	}
	return true
}

// HasDependencies checks if service has dependencies
func (s *Service) HasDependencies() bool {
	return len(s.Dependencies) > 0
}
