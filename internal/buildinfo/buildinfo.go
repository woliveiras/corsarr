package buildinfo

import (
	"runtime/debug"
	"strings"
)

var (
	Version = "development"
	Commit  = ""
	Date    = ""
)

// Current returns the release version injected by the build pipeline. Local
// builds remain explicitly identifiable as development builds.
func Current() string {
	version := strings.TrimSpace(Version)
	if version != "" && version != "development" && version != "(devel)" {
		return strings.TrimPrefix(version, "v")
	}

	build, ok := debug.ReadBuildInfo()
	if !ok || build.Main.Version == "" || build.Main.Version == "(devel)" {
		return "development"
	}

	return strings.TrimPrefix(build.Main.Version, "v")
}
