package onboarding

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Execute the real installer control flow with an isolated filesystem and
// fake system commands. No package manager, network or privilege change runs.
func TestLinuxInstallerScript(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux shell test")
	}
	for _, tc := range []struct {
		name, distro, conflict, fingerprint string
		success                             bool
	}{
		{"ubuntu", "ID=ubuntu\nVERSION_CODENAME=noble\n", "", "9DC858229FC7DD38854AE2D88D81803C0EBFCD88", true},
		{"debian", "ID=debian\nVERSION_CODENAME=trixie\n", "", "9DC858229FC7DD38854AE2D88D81803C0EBFCD88", true},
		{"conflict", "ID=ubuntu\nVERSION_CODENAME=noble\n", "docker.io", "", false},
		{"bad-key", "ID=ubuntu\nVERSION_CODENAME=noble\n", "", "BAD", false},
		{"existing-repository", "ID=ubuntu\nVERSION_CODENAME=noble\n", "", "", false},
		{"unsupported", "ID=linuxmint\nVERSION_CODENAME=noble\n", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			bin := filepath.Join(root, "bin")
			if err := os.MkdirAll(bin, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(root, "apt", "sources.list.d"), 0700); err != nil {
				t.Fatal(err)
			}
			write := func(path, contents string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(contents), 0700); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"grep", "awk", "install", "mktemp", "rm", "cat", "chmod"} {
				path, err := exec.LookPath(name)
				if err != nil {
					t.Skipf("%s unavailable", name)
				}
				if err := os.Symlink(path, filepath.Join(bin, name)); err != nil {
					t.Fatal(err)
				}
			}
			write(filepath.Join(root, "os-release"), tc.distro)
			if tc.name == "existing-repository" {
				write(filepath.Join(root, "apt", "sources.list.d", "docker.list"), "deb https://download.docker.com/linux/ubuntu noble stable\n")
			}
			write(filepath.Join(bin, "id"), "#!/bin/sh\necho 0\n")
			write(filepath.Join(bin, "dpkg"), "#!/bin/sh\necho amd64\n")
			write(filepath.Join(bin, "dpkg-query"), "#!/bin/sh\nfor arg; do if [ \"$arg\" = \"$CONFLICT\" ]; then printf 'install ok installed'; exit 0; fi; done\nexit 1\n")
			write(filepath.Join(bin, "apt-get"), "#!/bin/sh\nprintf 'apt %s\\n' \"$*\" >> \"$INSTALL_LOG\"\n")
			write(filepath.Join(bin, "systemctl"), "#!/bin/sh\nprintf 'service %s\\n' \"$*\" >> \"$INSTALL_LOG\"\n")
			write(filepath.Join(bin, "curl"), "#!/bin/sh\nfor arg; do last=$arg; done\nprintf key > \"$last\"\n")
			write(filepath.Join(bin, "gpg"), "#!/bin/sh\nprintf 'fpr:::::::::%s:\\n' \"$FINGERPRINT\"\n")
			script := strings.NewReplacer(
				"export PATH=/usr/sbin:/usr/bin:/sbin:/bin", "export PATH="+bin,
				"/etc/os-release", filepath.Join(root, "os-release"),
				"/run/systemd/system", root,
				"/etc/apt", filepath.Join(root, "apt"),
				"/usr/local/bin/docker", filepath.Join(root, "missing-local-docker"),
				"/snap/bin/docker", filepath.Join(root, "missing-snap-docker"),
			).Replace(linuxEngineInstallScript)
			command := exec.Command("/bin/sh", "-c", script)
			command.Env = append(os.Environ(), "CONFLICT="+tc.conflict, "FINGERPRINT="+tc.fingerprint, "INSTALL_LOG="+filepath.Join(root, "log"))
			output, err := command.CombinedOutput()
			if (err == nil) != tc.success {
				t.Fatalf("unexpected installer result: %v\n%s", err, output)
			}
			log, _ := os.ReadFile(filepath.Join(root, "log"))
			installed := strings.Contains(string(log), "install -y --no-remove docker-ce docker-ce-cli")
			if installed != tc.success {
				t.Fatalf("unexpected package installation: %s", log)
			}
			if tc.success && !strings.Contains(string(log), "service enable --now docker") {
				t.Fatal("service was not enabled")
			}
			if (tc.name == "conflict" || tc.name == "unsupported" || tc.name == "existing-repository") && len(log) > 0 {
				t.Fatalf("mutated unsupported host: %s", log)
			}
		})
	}
}
