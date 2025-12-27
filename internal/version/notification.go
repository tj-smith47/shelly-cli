// Package version provides build-time version information for the CLI.
package version

import (
	"os"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// ShowUpdateNotification displays a cached update notification if available.
// This is non-blocking and only reads from the cache file.
// It skips notification for certain commands (version, update, completion, help)
// and respects SHELLY_NO_UPDATE_CHECK env var.
func ShowUpdateNotification() {
	// Skip if update check is disabled
	if os.Getenv("SHELLY_NO_UPDATE_CHECK") != "" {
		return
	}

	// Skip for certain commands (they handle their own update info)
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmd == "version" || cmd == "update" || cmd == "completion" || cmd == "help" {
			return
		}
	}

	// Use existing cache reader
	cachedVersion := ReadCachedVersion()
	if cachedVersion == "" {
		return
	}

	currentVersion := strings.TrimPrefix(Version, "v")
	latestVersion := strings.TrimPrefix(cachedVersion, "v")

	if currentVersion == DevVersion || currentVersion == "" {
		return
	}

	// Simple semver comparison - if latest is different and "newer"
	if latestVersion > currentVersion {
		iostreams.Warning("\nUpdate available: %s -> %s (run 'shelly update' to install)\n", Version, cachedVersion)
	}
}
