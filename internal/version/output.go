// Package version provides build-time version information for the CLI.
package version

import (
	"context"
	"encoding/json"
	"io"
)

// Output extends Info with optional update availability information for JSON output.
type Output struct {
	Version       string  `json:"version"`
	Commit        string  `json:"commit"`
	Date          string  `json:"date"`
	BuiltBy       string  `json:"built_by"`
	GoVersion     string  `json:"go_version"`
	OS            string  `json:"os"`
	Arch          string  `json:"arch"`
	UpdateAvail   *string `json:"update_available,omitempty"`
	LatestVersion *string `json:"latest_version,omitempty"`
}

// NewOutput creates a new Output from Info.
func NewOutput(info Info) *Output {
	return &Output{
		Version:   info.Version,
		Commit:    info.Commit,
		Date:      info.Date,
		BuiltBy:   info.BuiltBy,
		GoVersion: info.GoVersion,
		OS:        info.OS,
		Arch:      info.Arch,
	}
}

// SetUpdateInfo sets the update availability information.
func (o *Output) SetUpdateInfo(latestVersion string, updateAvailable bool) {
	o.LatestVersion = &latestVersion
	if updateAvailable {
		avail := "yes"
		o.UpdateAvail = &avail
	} else {
		avail := "no"
		o.UpdateAvail = &avail
	}
}

// WriteJSON writes the version output as indented JSON.
func (o *Output) WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(o)
}

// WriteJSONOutput writes version info as JSON, optionally including update check results.
// The isNewer function compares current and latest versions to determine if an update is available.
func WriteJSONOutput(ctx context.Context, w io.Writer, info Info, checkUpdate bool, fetcher ReleaseFetcher, isNewer func(current, latest string) bool) error {
	output := NewOutput(info)
	if checkUpdate {
		if result, err := CheckForUpdates(ctx, info.Version, fetcher, isNewer); err == nil && !result.SkippedDevBuild {
			output.SetUpdateInfo(result.LatestVersion, result.UpdateAvailable)
		}
	}
	return output.WriteJSON(w)
}
