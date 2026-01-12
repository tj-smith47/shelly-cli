// Package version provides build-time version information for the CLI.
package version

import (
	"fmt"
	"runtime"
)

// DevVersion is the version string used for development builds.
const DevVersion = "dev"

// Build-time variables set via ldflags.
var (
	Version = DevVersion
	Commit  = "unknown"
	Date    = "unknown"
	BuiltBy = "unknown"
)

// Info contains structured version information.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	BuiltBy   string `json:"built_by"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the current version info.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		BuiltBy:   BuiltBy,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// Short returns a short version string (just the version number).
func Short() string {
	return Version
}

// Long returns a detailed version string with all build info.
func Long() string {
	info := Get()
	return fmt.Sprintf(
		"shelly %s\n"+
			"  commit:  %s\n"+
			"  built:   %s\n"+
			"  by:      %s\n"+
			"  go:      %s\n"+
			"  os/arch: %s/%s",
		info.Version,
		info.Commit,
		info.Date,
		info.BuiltBy,
		info.GoVersion,
		info.OS,
		info.Arch,
	)
}

// String returns the version string (alias for Short).
func String() string {
	return Short()
}

// IsDevelopment returns true if running a development build.
func IsDevelopment() bool {
	return Version == "" || Version == DevVersion
}

// CompareVersions compares two semantic versions.
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func CompareVersions(v1, v2 string) int {
	v1 = trimPrefix(v1, "v")
	v2 = trimPrefix(v2, "v")

	// Split into parts
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := range maxLen {
		var p1, p2 int
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	return 0
}

// parseVersion parses a version string into numeric parts.
func parseVersion(v string) []int {
	// Remove prerelease suffix (e.g., -beta.1)
	if idx := indexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}

	parts := split(v, ".")
	result := make([]int, 0, len(parts))

	for _, p := range parts {
		var num int
		_, err := fmt.Sscanf(p, "%d", &num)
		if err != nil {
			num = 0
		}
		result = append(result, num)
	}

	return result
}

// IsNewerVersion returns true if available is newer than current.
func IsNewerVersion(current, available string) bool {
	return CompareVersions(available, current) > 0
}

// trimPrefix removes the prefix from s if present.
func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

// indexAny returns the index of the first occurrence of any character in chars.
func indexAny(s, chars string) int {
	for i, c := range s {
		for _, ch := range chars {
			if c == ch {
				return i
			}
		}
	}
	return -1
}

// split splits s by sep.
func split(s, sep string) []string {
	if sep == "" {
		return []string{s}
	}
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}
