// Package version provides build-time version information for the CLI.
package version

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// CachePath returns the path to the version cache file.
func CachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "shelly", "cache", "latest-version")
}

// ReadCachedVersion reads the cached latest version if available and recent (within 24 hours).
// Returns empty string if cache is missing, stale, or unreadable.
func ReadCachedVersion() string {
	cachePath := CachePath()
	if cachePath == "" {
		return ""
	}

	info, err := os.Stat(cachePath)
	if err != nil {
		return ""
	}

	if time.Since(info.ModTime()) > 24*time.Hour {
		return ""
	}

	data, err := os.ReadFile(cachePath) //nolint:gosec // G304: cachePath is from known config directory
	if err != nil {
		return ""
	}

	return string(data)
}

// WriteCache writes the latest version to the cache file.
// Errors are returned for the caller to handle (e.g., log via ios.DebugErr).
func WriteCache(latestVersion string) error {
	cachePath := CachePath()
	if cachePath == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o750); err != nil {
		return err
	}

	return os.WriteFile(cachePath, []byte(latestVersion), 0o600)
}

// UpdateResult holds the result of an update check.
type UpdateResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	SkippedDevBuild bool
	CacheWriteErr   error // Error from writing cache, if any
}

// ReleaseFetcher fetches the latest release version.
type ReleaseFetcher func(ctx context.Context) (version string, err error)

// CheckForUpdates checks if an update is available using the provided fetcher.
// Returns nil result if the check was skipped (dev build).
func CheckForUpdates(ctx context.Context, currentVersion string, fetch ReleaseFetcher, isNewer func(current, latest string) bool) (*UpdateResult, error) {
	// Skip for dev versions
	if currentVersion == "" || currentVersion == "dev" {
		return &UpdateResult{SkippedDevBuild: true}, nil
	}

	latestVersion, err := fetch(ctx)
	if err != nil {
		return nil, err
	}

	result := &UpdateResult{
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: isNewer(currentVersion, latestVersion),
		CacheWriteErr:   WriteCache(latestVersion),
	}

	return result, nil
}
