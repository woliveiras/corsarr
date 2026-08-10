//go:build darwin

package hostreadiness

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

type systemChecker struct {
	platform     string
	architecture string
	diskPath     string
	run          func(context.Context, string, ...string) (string, error)
	freeDisk     func(string) (uint64, error)
}

func NewChecker(platform, architecture, diskPath string) Checker {
	return &systemChecker{
		platform: platform, architecture: architecture, diskPath: diskPath,
		run: func(ctx context.Context, name string, args ...string) (string, error) {
			output, err := exec.CommandContext(ctx, name, args...).Output()
			return strings.TrimSpace(string(output)), err
		},
		freeDisk: func(path string) (uint64, error) {
			var stat unix.Statfs_t
			if err := unix.Statfs(path, &stat); err != nil {
				return 0, err
			}
			return stat.Bavail * uint64(stat.Bsize), nil
		},
	}
}

func (c *systemChecker) Check(ctx context.Context) Status {
	checkContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	version, _ := c.run(checkContext, "/usr/bin/sw_vers", "-productVersion")
	memoryText, _ := c.run(checkContext, "/usr/sbin/sysctl", "-n", "hw.memsize")
	memory, _ := strconv.ParseUint(strings.TrimSpace(memoryText), 10, 64)
	freeDisk, _ := c.freeDisk(c.diskPath)
	return Evaluate(Facts{
		Platform: c.platform, Architecture: c.architecture, OSVersion: version,
		MemoryBytes: memory, FreeDiskBytes: freeDisk,
	})
}
