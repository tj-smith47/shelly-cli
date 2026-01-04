package version

import (
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

func TestShowUpdateNotification_EnvDisabled(t *testing.T) {
	// Set the disable env var
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "1")

	// This should return early without doing anything
	// We just verify it doesn't panic
	ShowUpdateNotification()
}
func TestShowUpdateNotification_NoCache(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	// Create a virtual home with no cache
	t.Setenv("HOME", "/test/home")
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")

	// Should not panic with no cache
	ShowUpdateNotification()
}
func TestShowUpdateNotification_DevBuild(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	const testHome = "/test/home"
	t.Setenv("HOME", testHome)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")

	// Create cache with a version
	cachePath := testHome + "/.config/shelly/cache"
	if err := config.Fs().MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(config.Fs(), cachePath+"/latest-version", []byte("2.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Save original version and set to dev
	originalVersion := Version
	Version = "dev"
	defer func() { Version = originalVersion }()

	// Should return early for dev builds without panic
	ShowUpdateNotification()
}
func TestShowUpdateNotification_SkippedCommands(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	const testHome = "/test/home"
	t.Setenv("HOME", testHome)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")

	// Create cache with a newer version
	cachePath := testHome + "/.config/shelly/cache"
	if err := config.Fs().MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(config.Fs(), cachePath+"/latest-version", []byte("2.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Save original args
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })

	// Test skipped commands
	skippedCommands := []string{"version", "update", "completion", "help"}
	//nolint:paralleltest // Subtests modify global os.Args and can't run in parallel
	for _, cmd := range skippedCommands {
		t.Run("skip_"+cmd, func(t *testing.T) {
			os.Args = []string{"shelly", cmd}
			// Should return early without panic
			ShowUpdateNotification()
		})
	}
}
func TestShowUpdateNotification_NoArgs(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	t.Setenv("HOME", "/test/home")
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with no args (just program name)
	os.Args = []string{"shelly"}

	// Should not panic with no args
	ShowUpdateNotification()
}
func TestShowUpdateNotification_UpdateAvailable(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	const testHome = "/test/home"
	t.Setenv("HOME", testHome)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")

	// Create cache with a newer version
	cachePath := testHome + "/.config/shelly/cache"
	if err := config.Fs().MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(config.Fs(), cachePath+"/latest-version", []byte("99.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Save original version and set to a release version
	originalVersion := Version
	Version = "v1.0.0"
	defer func() { Version = originalVersion }()

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Use a command that's not skipped
	os.Args = []string{"shelly", "device", "list"}

	// Should execute the full path including showing notification
	// (we can't easily verify the output, but we ensure it doesn't panic)
	ShowUpdateNotification()
}
func TestShowUpdateNotification_EmptyVersion(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	const testHome = "/test/home"
	t.Setenv("HOME", testHome)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")

	// Create cache with a version
	cachePath := testHome + "/.config/shelly/cache"
	if err := config.Fs().MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(config.Fs(), cachePath+"/latest-version", []byte("2.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Save original version and set to empty
	originalVersion := Version
	Version = ""
	defer func() { Version = originalVersion }()

	// Should return early for empty version
	ShowUpdateNotification()
}
func TestShowUpdateNotification_NoUpdate(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	const testHome = "/test/home"
	t.Setenv("HOME", testHome)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")

	// Create cache with same version (no update)
	cachePath := testHome + "/.config/shelly/cache"
	if err := config.Fs().MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(config.Fs(), cachePath+"/latest-version", []byte("1.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Save original version
	originalVersion := Version
	Version = "v1.0.0"
	defer func() { Version = originalVersion }()

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"shelly", "device", "list"}

	// Should not show notification when versions are equal
	ShowUpdateNotification()
}
