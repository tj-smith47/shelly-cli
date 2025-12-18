// Package github provides GitHub API integration for downloading releases.
package github

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// CheckForUpdates checks GitHub releases for available updates.
// Returns an UpdateResult with version comparison and caches the latest version.
func CheckForUpdates(ctx context.Context, ios *iostreams.IOStreams, currentVersion string) (*version.UpdateResult, error) {
	return version.CheckForUpdates(ctx, currentVersion, ReleaseFetcher(ios), IsNewerVersion)
}
