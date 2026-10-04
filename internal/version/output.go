// Package version provides build-time version information for the CLI.
package version

import "context"

const availabilityYes = "yes"

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
		avail := availabilityYes
		o.UpdateAvail = &avail
	} else {
		avail := "no"
		o.UpdateAvail = &avail
	}
}

// BuildOutput returns info as an Output, with the update check result added
// when checkUpdate is set and the check succeeds. The isNewer function compares
// the current and latest versions to decide whether an update is available.
func BuildOutput(ctx context.Context, info Info, checkUpdate bool, fetcher ReleaseFetcher, isNewer func(current, latest string) bool) *Output {
	output := NewOutput(info)
	if checkUpdate {
		if result, err := CheckForUpdates(ctx, info.Version, fetcher, isNewer); err == nil && !result.SkippedDevBuild {
			output.SetUpdateInfo(result.LatestVersion, result.UpdateAvailable)
		}
	}
	return output
}
