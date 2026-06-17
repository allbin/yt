// Package version reports the build version of the yt binary.
package version

import "runtime/debug"

// injected at build time via:
//
//	-ldflags "-X github.com/allbin/yt/internal/version.version=<value>"
var version string

// Version returns the build version. It prefers a value injected at build
// time; otherwise it derives one from the embedded VCS build info (short git
// sha, with a "-dirty" suffix when the working tree had uncommitted changes).
func Version() string {
	if version != "" {
		return version
	}
	return fromBuildInfo()
}

func fromBuildInfo() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	var revision string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}

	if revision == "" {
		// Built with `go install pkg@version` or VCS stamping disabled.
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
		return "unknown"
	}

	if len(revision) > 7 {
		revision = revision[:7]
	}
	if dirty {
		revision += "-dirty"
	}
	return revision
}
