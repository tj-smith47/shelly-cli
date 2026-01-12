package term

import (
	"context"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// DisplayUpdateAvailable prints an update notification with current and available versions.
func DisplayUpdateAvailable(ios *iostreams.IOStreams, currentVersion, availableVersion string) {
	ios.Printf("\n")
	ios.Warning("Update available: %s -> %s", currentVersion, availableVersion)
	ios.Printf("  Run 'shelly update' to install the latest version\n")
}

// DisplayUpToDate prints a success message indicating the CLI is up to date.
func DisplayUpToDate(ios *iostreams.IOStreams) {
	ios.Printf("\n")
	ios.Success("You are using the latest version")
}

// DisplayUpdateResult displays the result of an update check.
func DisplayUpdateResult(ios *iostreams.IOStreams, currentVersion, latestVersion string, updateAvailable bool, cacheErr error) {
	if updateAvailable {
		DisplayUpdateAvailable(ios, currentVersion, latestVersion)
	} else {
		DisplayUpToDate(ios)
	}
	if cacheErr != nil {
		ios.DebugErr("writing version cache", cacheErr)
	}
}

// DisplayVersionInfo prints version information to the console.
func DisplayVersionInfo(ios *iostreams.IOStreams, ver, commit, date, builtBy, goVer, osName, arch string) {
	const unknownValue = "unknown"
	ios.Printf("shelly version %s\n", ver)
	if commit != "" && commit != unknownValue {
		ios.Printf("  commit: %s\n", commit)
	}
	if date != "" && date != unknownValue {
		ios.Printf("  built: %s\n", date)
	}
	if builtBy != "" && builtBy != unknownValue {
		ios.Printf("  by: %s\n", builtBy)
	}
	ios.Printf("  go: %s\n", goVer)
	ios.Printf("  os/arch: %s/%s\n", osName, arch)
}

// UpdateChecker is a function that checks for updates and returns the result.
type UpdateChecker func(ctx context.Context) (*version.UpdateResult, error)

// RunUpdateCheck runs an update check with spinner and displays results.
func RunUpdateCheck(ctx context.Context, ios *iostreams.IOStreams, checker UpdateChecker) {
	ios.StartProgress("Checking for updates...")
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	result, err := checker(checkCtx)
	cancel()
	ios.StopProgress()
	if err != nil {
		ios.DebugErr("checking for updates", err)
	} else if !result.SkippedDevBuild {
		DisplayUpdateResult(ios, result.CurrentVersion, result.LatestVersion, result.UpdateAvailable, result.CacheWriteErr)
	}
}

// DisplayUpdateStatus shows the update check status with optional release URL.
func DisplayUpdateStatus(ios *iostreams.IOStreams, currentVersion, availableVersion string, hasUpdate bool, releaseURL string) {
	if hasUpdate {
		ios.Warning("Update available: %s -> %s", currentVersion, availableVersion)
		ios.Printf("  Run 'shelly update' to install\n")
		if releaseURL != "" {
			ios.Printf("  Release page: %s\n", releaseURL)
		}
	} else {
		ios.Success("Already at latest version (%s)", currentVersion)
	}
}

// CacheInfo holds cache statistics for display.
type CacheInfo struct {
	Location    string
	Entries     int
	Size        string
	DeviceCount int
}

// DisplayCacheInfo shows cache statistics in version output.
func DisplayCacheInfo(ios *iostreams.IOStreams, info CacheInfo) {
	ios.Printf("\nCache:\n")
	ios.Printf("  location: %s\n", info.Location)
	ios.Printf("  entries: %d\n", info.Entries)
	ios.Printf("  size: %s\n", info.Size)
	if info.DeviceCount > 0 {
		ios.Printf("  devices: %d\n", info.DeviceCount)
	}
}
