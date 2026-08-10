package runtime

import (
	"strings"
	"testing"
)

const testImageDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestContainerSpecAcceptsPinnedLocalOnlyApplication(t *testing.T) {
	spec := ContainerSpec{
		ApplicationID: "radarr",
		Image:         "lscr.io/linuxserver/radarr@" + testImageDigest,
		Ports: []PortBinding{{
			HostPort:      7878,
			ContainerPort: 7878,
			Protocol:      ProtocolTCP,
			Exposure:      ExposureLoopback,
		}},
		Mounts: []BindMount{
			{HostPath: "/Users/test/Media/Corsarr/config/radarr", ContainerPath: "/config"},
			{HostPath: "/Users/test/Media/Corsarr/media", ContainerPath: "/data"},
		},
		Environment: map[string]string{"TZ": "Europe/Madrid"},
	}

	if err := spec.Validate(); err != nil {
		t.Fatalf("expected valid container spec, got %v", err)
	}
}

func TestContainerContractFingerprintIsOrderStableAndExcludesImage(t *testing.T) {
	first := ContainerSpec{
		ApplicationID: "radarr",
		Image:         "example.invalid/app@" + testImageDigest,
		Ports: []PortBinding{
			{HostPort: 7878, ContainerPort: 7878, Protocol: ProtocolTCP, Exposure: ExposureLoopback},
			{HostPort: 9898, ContainerPort: 9898, Protocol: ProtocolUDP, Exposure: ExposureLAN},
		},
		Mounts: []BindMount{
			{HostPath: "/tmp/Corsarr/media", ContainerPath: "/data"},
			{HostPath: "/tmp/Corsarr/config/radarr", ContainerPath: "/config"},
		},
		Environment: map[string]string{"TZ": "Europe/Madrid", "UMASK": "002"},
	}
	second := first
	second.Image = "example.invalid/app@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	second.Ports = []PortBinding{first.Ports[1], first.Ports[0]}
	second.Mounts = []BindMount{first.Mounts[1], first.Mounts[0]}

	firstFingerprint, err := first.ContractFingerprint()
	if err != nil {
		t.Fatalf("fingerprint first contract: %v", err)
	}
	secondFingerprint, err := second.ContractFingerprint()
	if err != nil {
		t.Fatalf("fingerprint reordered contract: %v", err)
	}
	if firstFingerprint != secondFingerprint || len(firstFingerprint) != 64 {
		t.Fatalf("expected stable SHA-256 fingerprint, got %q and %q", firstFingerprint, secondFingerprint)
	}

	changed := second
	changed.Ports = append([]PortBinding(nil), second.Ports...)
	changed.Ports[0].Exposure = ExposureLoopback
	changedFingerprint, err := changed.ContractFingerprint()
	if err != nil {
		t.Fatalf("fingerprint changed contract: %v", err)
	}
	if firstFingerprint == changedFingerprint {
		t.Fatal("port exposure change did not change container contract fingerprint")
	}
}

func TestContainerSpecRejectsMutableImageReference(t *testing.T) {
	spec := ContainerSpec{
		ApplicationID: "radarr",
		Image:         "lscr.io/linuxserver/radarr:latest",
	}

	err := spec.Validate()
	if err == nil || !strings.Contains(err.Error(), "immutable digest") {
		t.Fatalf("expected immutable digest error, got %v", err)
	}
}

func TestContainerSpecRejectsUnsafeApplicationIDAndMounts(t *testing.T) {
	tests := []struct {
		name string
		spec ContainerSpec
	}{
		{
			name: "path traversal application ID",
			spec: ContainerSpec{ApplicationID: "../radarr", Image: "example.invalid/app@" + testImageDigest},
		},
		{
			name: "relative host mount",
			spec: ContainerSpec{
				ApplicationID: "radarr",
				Image:         "example.invalid/app@" + testImageDigest,
				Mounts:        []BindMount{{HostPath: "Corsarr/config/radarr", ContainerPath: "/config"}},
			},
		},
		{
			name: "relative container mount",
			spec: ContainerSpec{
				ApplicationID: "radarr",
				Image:         "example.invalid/app@" + testImageDigest,
				Mounts:        []BindMount{{HostPath: "/tmp/radarr", ContainerPath: "config"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.spec.Validate(); err == nil {
				t.Fatal("expected unsafe spec to be rejected")
			}
		})
	}
}

func TestContainerSpecRejectsImplicitLANExposure(t *testing.T) {
	spec := ContainerSpec{
		ApplicationID: "radarr",
		Image:         "example.invalid/app@" + testImageDigest,
		Ports: []PortBinding{{
			HostPort:      7878,
			ContainerPort: 7878,
			Protocol:      ProtocolTCP,
		}},
	}

	err := spec.Validate()
	if err == nil || !strings.Contains(err.Error(), "exposure") {
		t.Fatalf("expected explicit exposure error, got %v", err)
	}
}
