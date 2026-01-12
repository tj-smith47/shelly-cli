package version

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

const (
	testVersion1 = "1.0.0"
	testVersion2 = "2.0.0"
)

func TestCachePath(t *testing.T) {
	t.Parallel()

	path := CachePath()
	// Should return a non-empty path (unless home dir is unavailable)
	if path == "" {
		t.Skip("CachePath returned empty (home dir unavailable)")
	}

	// Should contain expected path components
	if !contains(path, ".config") {
		t.Errorf("CachePath() = %q, expected to contain '.config'", path)
	}
	if !contains(path, "shelly") {
		t.Errorf("CachePath() = %q, expected to contain 'shelly'", path)
	}
	if !contains(path, "latest-version") {
		t.Errorf("CachePath() = %q, expected to contain 'latest-version'", path)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestWriteCache_ReadCachedVersion(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	// Write a version to cache
	testVersion := "1.2.3"
	err := WriteCache(testVersion)
	if err != nil {
		t.Fatalf("WriteCache() error = %v", err)
	}

	// Read it back
	cached := ReadCachedVersion()
	if cached != testVersion {
		t.Errorf("ReadCachedVersion() = %q, want %q", cached, testVersion)
	}
}

func TestReadCachedVersion_NoCache(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	// Should return empty string when no cache exists
	cached := ReadCachedVersion()
	if cached != "" {
		t.Errorf("ReadCachedVersion() = %q, want empty string when no cache", cached)
	}
}

func TestReadCachedVersion_StaleCache(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	const testHome = "/test/home"
	t.Setenv("HOME", testHome)

	// Manually create a cache file
	cachePath := testHome + "/.config/shelly/cache"
	if err := fs.MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	cacheFile := cachePath + "/latest-version"
	if err := afero.WriteFile(fs, cacheFile, []byte("1.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}

	// Set the file modification time to 25 hours ago (stale)
	staleTime := time.Now().Add(-25 * time.Hour)
	if err := fs.Chtimes(cacheFile, staleTime, staleTime); err != nil {
		t.Fatalf("failed to set mtime: %v", err)
	}

	// Should return empty string for stale cache
	cached := ReadCachedVersion()
	if cached != "" {
		t.Errorf("ReadCachedVersion() = %q, want empty string for stale cache", cached)
	}
}

func TestCheckForUpdates_DevBuild(t *testing.T) {
	t.Parallel()

	fetcher := func(_ context.Context) (string, error) {
		return testVersion1, nil
	}

	// Test with dev version
	result, err := CheckForUpdates(context.Background(), "dev", fetcher, nil)
	if err != nil {
		t.Fatalf("CheckForUpdates() error = %v", err)
	}

	if !result.SkippedDevBuild {
		t.Error("SkippedDevBuild should be true for dev builds")
	}
}

func TestCheckForUpdates_EmptyVersion(t *testing.T) {
	t.Parallel()

	fetcher := func(_ context.Context) (string, error) {
		return testVersion1, nil
	}

	// Test with empty version
	result, err := CheckForUpdates(context.Background(), "", fetcher, nil)
	if err != nil {
		t.Fatalf("CheckForUpdates() error = %v", err)
	}

	if !result.SkippedDevBuild {
		t.Error("SkippedDevBuild should be true for empty version")
	}
}

func TestCheckForUpdates_UpdateAvailable(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	fetcher := func(_ context.Context) (string, error) {
		return testVersion2, nil
	}

	isNewer := func(current, latest string) bool {
		return latest > current
	}

	result, err := CheckForUpdates(context.Background(), testVersion1, fetcher, isNewer)
	if err != nil {
		t.Fatalf("CheckForUpdates() error = %v", err)
	}

	if result.SkippedDevBuild {
		t.Error("SkippedDevBuild should be false for release builds")
	}
	if result.CurrentVersion != testVersion1 {
		t.Errorf("CurrentVersion = %q, want %q", result.CurrentVersion, testVersion1)
	}
	if result.LatestVersion != testVersion2 {
		t.Errorf("LatestVersion = %q, want %q", result.LatestVersion, testVersion2)
	}
	if !result.UpdateAvailable {
		t.Error("UpdateAvailable should be true when latest > current")
	}
}

func TestCheckForUpdates_NoUpdate(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	fetcher := func(_ context.Context) (string, error) {
		return testVersion1, nil
	}

	isNewer := func(current, latest string) bool {
		return latest > current
	}

	result, err := CheckForUpdates(context.Background(), testVersion1, fetcher, isNewer)
	if err != nil {
		t.Fatalf("CheckForUpdates() error = %v", err)
	}

	if result.UpdateAvailable {
		t.Error("UpdateAvailable should be false when current == latest")
	}
}

func TestCheckForUpdates_FetcherError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("network error")
	fetcher := func(_ context.Context) (string, error) {
		return "", expectedErr
	}

	_, err := CheckForUpdates(context.Background(), testVersion1, fetcher, nil)
	if err == nil {
		t.Error("CheckForUpdates() should return error when fetcher fails")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("CheckForUpdates() error = %v, want %v", err, expectedErr)
	}
}

func TestUpdateResult_Struct(t *testing.T) {
	t.Parallel()

	result := UpdateResult{
		CurrentVersion:  testVersion1,
		LatestVersion:   testVersion2,
		UpdateAvailable: true,
		SkippedDevBuild: false,
		CacheWriteErr:   nil,
	}

	if result.CurrentVersion != testVersion1 {
		t.Errorf("CurrentVersion = %q, want %q", result.CurrentVersion, testVersion1)
	}
	if result.LatestVersion != testVersion2 {
		t.Errorf("LatestVersion = %q, want %q", result.LatestVersion, testVersion2)
	}
	if !result.UpdateAvailable {
		t.Error("UpdateAvailable should be true")
	}
	if result.SkippedDevBuild {
		t.Error("SkippedDevBuild should be false")
	}
	if result.CacheWriteErr != nil {
		t.Errorf("CacheWriteErr = %v, want nil", result.CacheWriteErr)
	}
}

func TestWriteCache_EmptyCachePath(t *testing.T) {
	// Use memory filesystem to prevent writes to real filesystem
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	// Set HOME to invalid path to make CachePath return empty
	t.Setenv("HOME", "")

	// WriteCache should return nil when CachePath is empty (no-op)
	err := WriteCache("1.0.0")
	if err != nil {
		t.Errorf("WriteCache() error = %v, want nil when cache path is empty", err)
	}
}

func TestReadCachedVersion_EmptyCachePath(t *testing.T) {
	// Use memory filesystem to prevent reads from real filesystem
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	// Set HOME to invalid path to make CachePath return empty
	t.Setenv("HOME", "")

	// ReadCachedVersion should return empty when CachePath is empty
	cached := ReadCachedVersion()
	if cached != "" {
		t.Errorf("ReadCachedVersion() = %q, want empty when cache path is empty", cached)
	}
}

func TestCachePath_NoHomeDir(t *testing.T) {
	// Use memory filesystem for consistency
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	// Set HOME to empty to trigger the error path
	t.Setenv("HOME", "")

	path := CachePath()
	if path != "" {
		t.Errorf("CachePath() = %q, want empty when HOME is unset", path)
	}
}
