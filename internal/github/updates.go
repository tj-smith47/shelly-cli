// Package github provides GitHub API integration for downloading releases.
package github

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// CheckForUpdates checks GitHub releases for available updates.
// Returns an UpdateResult with version comparison and caches the latest version.
func CheckForUpdates(ctx context.Context, ios *iostreams.IOStreams, currentVersion string) (*version.UpdateResult, error) {
	return version.CheckForUpdates(ctx, currentVersion, ReleaseFetcher(ios), IsNewerVersion)
}

// FetchSpecificVersion fetches a specific release by version tag.
func (c *Client) FetchSpecificVersion(ctx context.Context, ver string) (*Release, error) {
	tag := ver
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}
	return c.GetReleaseByTag(ctx, DefaultOwner, DefaultRepo, tag)
}

// FetchLatestVersion fetches the latest release, optionally including prereleases.
func (c *Client) FetchLatestVersion(ctx context.Context, includePre bool) (*Release, error) {
	if !includePre {
		return c.GetLatestRelease(ctx, DefaultOwner, DefaultRepo)
	}

	// Get all releases and find the latest (including pre)
	releases, err := c.ListReleases(ctx, DefaultOwner, DefaultRepo, true)
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, ErrNoReleases
	}

	SortReleasesByVersion(releases)
	return &releases[0], nil
}

// GetTargetRelease fetches either a specific version or the latest release.
func (c *Client) GetTargetRelease(ctx context.Context, ver string, includePre bool) (*Release, error) {
	if ver != "" {
		return c.FetchSpecificVersion(ctx, ver)
	}
	return c.FetchLatestVersion(ctx, includePre)
}

// FindPreviousRelease finds the release before the current version for rollback.
func (c *Client) FindPreviousRelease(ctx context.Context, currentVersion string, includePre bool) (*Release, error) {
	releases, err := c.ListReleases(ctx, DefaultOwner, DefaultRepo, includePre)
	if err != nil {
		return nil, err
	}

	if len(releases) < 2 {
		return nil, ErrNoReleases
	}

	SortReleasesByVersion(releases)

	// Find current version index and get previous
	for i, r := range releases {
		if r.Version() == strings.TrimPrefix(currentVersion, "v") && i+1 < len(releases) {
			return &releases[i+1], nil
		}
	}

	// Fallback to second release
	return &releases[1], nil
}

// CheckForUpdatesCached checks for updates using a cached result.
// Returns nil if no update is available or check is disabled.
func CheckForUpdatesCached(ctx context.Context, ios *iostreams.IOStreams, cachePath, currentVersion string) *Release {
	// Check if updates are disabled
	if os.Getenv("SHELLY_NO_UPDATE_CHECK") != "" {
		return nil
	}

	// Check cache
	cacheValid := false
	if info, err := os.Stat(cachePath); err == nil {
		// Cache is valid for 24 hours
		cacheValid = time.Since(info.ModTime()) < 24*time.Hour
	}

	if cacheValid {
		// Read cached version
		data, err := os.ReadFile(cachePath) //nolint:gosec // G304: cachePath is from known config directory
		if err == nil {
			cachedVersion := strings.TrimSpace(string(data))
			if cachedVersion != "" && IsNewerVersion(currentVersion, cachedVersion) {
				// Return a minimal release with just the version
				return &Release{TagName: "v" + cachedVersion}
			}
			return nil
		}
	}

	// Perform check in background with short timeout
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ghClient := NewClient(ios)
	release, err := ghClient.GetLatestRelease(checkCtx, DefaultOwner, DefaultRepo)
	if err != nil {
		return nil
	}

	// Update cache
	if merr := os.MkdirAll(filepath.Dir(cachePath), 0o750); merr != nil {
		ios.DebugErr("creating cache directory", merr)
	}
	if werr := os.WriteFile(cachePath, []byte(release.Version()), 0o600); werr != nil {
		ios.DebugErr("writing update cache", werr)
	}

	if IsNewerVersion(currentVersion, release.Version()) {
		return release
	}

	return nil
}
