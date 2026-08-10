package hostprofile

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const defaultContainerUserID = 1000

type Profile struct {
	Timezone string
	PUID     int
	PGID     int
}

type Profiler struct {
	getenv       func(string) string
	evalSymlinks func(string) (string, error)
	currentUser  func() (string, string, error)
}

func NewProfiler() *Profiler {
	return &Profiler{
		getenv:       os.Getenv,
		evalSymlinks: filepath.EvalSymlinks,
		currentUser: func() (string, string, error) {
			current, err := user.Current()
			if err != nil {
				return "", "", err
			}
			return current.Uid, current.Gid, nil
		},
	}
}

func (p *Profiler) Current(platform string) Profile {
	profile := Profile{
		Timezone: detectTimezone(p.getenv, p.evalSymlinks),
		PUID:     defaultContainerUserID,
		PGID:     defaultContainerUserID,
	}
	if platform != "linux" {
		return profile
	}
	uid, gid, err := p.currentUser()
	if err != nil {
		return profile
	}
	parsedUID, uidErr := strconv.Atoi(uid)
	parsedGID, gidErr := strconv.Atoi(gid)
	if uidErr == nil && gidErr == nil && parsedUID > 0 && parsedGID > 0 {
		profile.PUID = parsedUID
		profile.PGID = parsedGID
	}
	return profile
}

func detectTimezone(
	getenv func(string) string,
	evalSymlinks func(string) (string, error),
) string {
	if timezone := validatedTimezone(getenv("TZ")); timezone != "" {
		return timezone
	}
	if target, err := evalSymlinks("/etc/localtime"); err == nil {
		const marker = "/zoneinfo/"
		if markerIndex := strings.LastIndex(target, marker); markerIndex >= 0 {
			if timezone := validatedTimezone(target[markerIndex+len(marker):]); timezone != "" {
				return timezone
			}
		}
	}
	if localName := time.Now().Location().String(); localName != "Local" {
		if timezone := validatedTimezone(localName); timezone != "" {
			return timezone
		}
	}
	return "Etc/UTC"
}

func validatedTimezone(candidate string) string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" || filepath.IsAbs(candidate) || strings.Contains(candidate, "..") {
		return ""
	}
	if _, err := time.LoadLocation(candidate); err != nil {
		return ""
	}
	return candidate
}
