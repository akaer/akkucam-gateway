// Package version exposes the gateway version and the commit it was built from.
package version

import (
	"runtime/debug"
	"strings"
)

// Version is the semantic version of the gateway. Bump it in the release pull
// request; the tag on main uses the same value with a "v" prefix.
var Version = "0.2.1"

// commit is injected at link time:
// -ldflags "-X akkucam-gateway/internal/version.commit=<hash>"
var commit = ""

// Commit returns the short commit hash. Builds inside a checkout fall back to
// the VCS stamp Go records in the binary, so local runs show it as well.
func Commit() string {
	if hash := short(strings.TrimSpace(commit)); hash != "" {
		return hash
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return short(strings.TrimSpace(setting.Value))
		}
	}
	return ""
}

// Display formats the version for logs and the dashboard.
func Display() string {
	if hash := Commit(); hash != "" {
		return Version + " (" + hash + ")"
	}
	return Version
}

func short(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
