package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"sort"
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
	Init          bool
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

// ContractFingerprint identifies the runtime contract that must remain stable
// across an image-only update. The image itself is deliberately excluded.
func (s ContainerSpec) ContractFingerprint() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	ports := append([]PortBinding(nil), s.Ports...)
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].HostPort != ports[j].HostPort {
			return ports[i].HostPort < ports[j].HostPort
		}
		if ports[i].ContainerPort != ports[j].ContainerPort {
			return ports[i].ContainerPort < ports[j].ContainerPort
		}
		if ports[i].Protocol != ports[j].Protocol {
			return ports[i].Protocol < ports[j].Protocol
		}
		return ports[i].Exposure < ports[j].Exposure
	})
	mounts := append([]BindMount(nil), s.Mounts...)
	for index := range mounts {
		mounts[index].HostPath = filepath.Clean(mounts[index].HostPath)
		mounts[index].ContainerPath = path.Clean(mounts[index].ContainerPath)
	}
	sort.Slice(mounts, func(i, j int) bool {
		if mounts[i].ContainerPath != mounts[j].ContainerPath {
			return mounts[i].ContainerPath < mounts[j].ContainerPath
		}
		return mounts[i].HostPath < mounts[j].HostPath
	})
	payload, err := json.Marshal(struct {
		ApplicationID string            `json:"applicationId"`
		Init          bool              `json:"init"`
		Ports         []PortBinding     `json:"ports"`
		Mounts        []BindMount       `json:"mounts"`
		Environment   map[string]string `json:"environment"`
	}{
		ApplicationID: s.ApplicationID,
		Init:          s.Init,
		Ports:         ports,
		Mounts:        mounts,
		Environment:   s.Environment,
	})
	if err != nil {
		return "", fmt.Errorf("encode container contract: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
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
