package runtime

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Protocol string

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolUDP Protocol = "udp"
)

type Exposure string

const (
	ExposureLoopback Exposure = "loopback"
	ExposureLAN      Exposure = "lan"
)

type PortBinding struct {
	HostPort      int
	ContainerPort int
	Protocol      Protocol
	Exposure      Exposure
}

type BindMount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

// ContainerSpec is the runtime-neutral, fully resolved contract accepted by a
// container runtime adapter. It never contains templates or arbitrary commands.
type ContainerSpec struct {
	ApplicationID string
	Image         string
	Ports         []PortBinding
	Mounts        []BindMount
	Environment   map[string]string
}

var (
	runtimeApplicationIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	imageDigestPattern          = regexp.MustCompile(`^[a-f0-9]{64}$`)
	environmentNamePattern      = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

func (s ContainerSpec) Validate() error {
	if !runtimeApplicationIDPattern.MatchString(s.ApplicationID) {
		return fmt.Errorf("unsafe application ID: %q", s.ApplicationID)
	}
	if err := validateImageReference(s.Image); err != nil {
		return err
	}
	if err := validatePortBindings(s.Ports); err != nil {
		return err
	}
	if err := validateBindMounts(s.Mounts); err != nil {
		return err
	}
	for name, value := range s.Environment {
		if !environmentNamePattern.MatchString(name) {
			return fmt.Errorf("invalid environment variable name: %q", name)
		}
		if strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("environment variable %s contains a null byte", name)
		}
	}
	return nil
}

func validateImageReference(reference string) error {
	const digestSeparator = "@sha256:"
	parts := strings.Split(reference, digestSeparator)
	if len(parts) != 2 || parts[0] == "" || !imageDigestPattern.MatchString(parts[1]) {
		return fmt.Errorf("image reference must use an immutable digest: %q", reference)
	}
	if strings.ContainsAny(parts[0], " \t\r\n@") {
		return fmt.Errorf("invalid image reference: %q", reference)
	}
	return nil
}

func validatePortBindings(bindings []PortBinding) error {
	seenHostPorts := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding.HostPort < 1 || binding.HostPort > 65535 {
			return fmt.Errorf("invalid host port: %d", binding.HostPort)
		}
		if binding.ContainerPort < 1 || binding.ContainerPort > 65535 {
			return fmt.Errorf("invalid container port: %d", binding.ContainerPort)
		}
		if binding.Protocol != ProtocolTCP && binding.Protocol != ProtocolUDP {
			return fmt.Errorf("unsupported port protocol: %q", binding.Protocol)
		}
		if binding.Exposure != ExposureLoopback && binding.Exposure != ExposureLAN {
			return fmt.Errorf("port exposure must be explicit for %d", binding.HostPort)
		}
		key := fmt.Sprintf("%d/%s", binding.HostPort, binding.Protocol)
		if _, duplicate := seenHostPorts[key]; duplicate {
			return fmt.Errorf("duplicate host port: %s", key)
		}
		seenHostPorts[key] = struct{}{}
	}
	return nil
}

func validateBindMounts(mounts []BindMount) error {
	seenContainerPaths := make(map[string]struct{}, len(mounts))
	for _, mount := range mounts {
		if !filepath.IsAbs(mount.HostPath) {
			return fmt.Errorf("host mount path must be absolute: %q", mount.HostPath)
		}
		if !path.IsAbs(mount.ContainerPath) {
			return fmt.Errorf("container mount path must be absolute: %q", mount.ContainerPath)
		}
		containerPath := path.Clean(mount.ContainerPath)
		if containerPath == "/" {
			return fmt.Errorf("container root cannot be a bind mount target")
		}
		if _, duplicate := seenContainerPaths[containerPath]; duplicate {
			return fmt.Errorf("duplicate container mount path: %q", containerPath)
		}
		seenContainerPaths[containerPath] = struct{}{}
	}
	return nil
}
