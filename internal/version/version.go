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
