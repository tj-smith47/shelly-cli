package plugins

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

const testMigrateDir = "/test/migrate"

func setupMigrateTestFs(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })
	return fs
}

// TestErrAlreadyMigrated tests ErrAlreadyMigrated error.
func TestErrAlreadyMigrated(t *testing.T) {
	t.Parallel()

	if ErrAlreadyMigrated == nil {
		t.Fatal("ErrAlreadyMigrated should not be nil")
	}
	if ErrAlreadyMigrated.Error() != "plugins already migrated" {
		t.Errorf("ErrAlreadyMigrated.Error() = %q, want %q", ErrAlreadyMigrated.Error(), "plugins already migrated")
	}
}

// TestMigrationResult tests MigrationResult struct.
func TestMigrationResult(t *testing.T) {
	t.Parallel()

	result := MigrationResult{
		Migrated: []string{"plugin1", "plugin2"},
		Skipped:  []string{"plugin3"},
		Errors:   []string{"plugin4: failed"},
	}

	if len(result.Migrated) != 2 {
		t.Errorf("Migrated len = %d, want 2", len(result.Migrated))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("Skipped len = %d, want 1", len(result.Skipped))
	}
	if len(result.Errors) != 1 {
		t.Errorf("Errors len = %d, want 1", len(result.Errors))
	}
}

// TestMigrationResult_Empty tests empty MigrationResult.
func TestMigrationResult_Empty(t *testing.T) {
	t.Parallel()

	result := MigrationResult{}

	if len(result.Migrated) != 0 {
		t.Errorf("Migrated len = %d, want 0", len(result.Migrated))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("Skipped len = %d, want 0", len(result.Skipped))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors len = %d, want 0", len(result.Errors))
	}
}

// TestHasOldFormatPlugins tests hasOldFormatPlugins function.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestHasOldFormatPlugins(t *testing.T) {
	fs := setupMigrateTestFs(t)
	pluginsDir := testMigrateDir + "/hasold"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Empty directory - no old format plugins
	has, err := hasOldFormatPlugins(fs, pluginsDir)
	if err != nil {
		t.Fatalf("hasOldFormatPlugins() error: %v", err)
	}
	if has {
		t.Error("hasOldFormatPlugins() = true for empty dir, want false")
	}

	// Add a new format plugin (directory)
	newPluginDir := filepath.Join(pluginsDir, "shelly-new")
	if err := fs.MkdirAll(newPluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	has, err = hasOldFormatPlugins(fs, pluginsDir)
	if err != nil {
		t.Fatalf("hasOldFormatPlugins() error: %v", err)
	}
	if has {
		t.Error("hasOldFormatPlugins() = true for new format plugin, want false")
	}

	// Add an old format plugin (bare binary)
	oldPluginPath := filepath.Join(pluginsDir, "shelly-old")
	if err := afero.WriteFile(fs, oldPluginPath, []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create old plugin: %v", err)
	}

	has, err = hasOldFormatPlugins(fs, pluginsDir)
	if err != nil {
		t.Fatalf("hasOldFormatPlugins() error: %v", err)
	}
	if !has {
		t.Error("hasOldFormatPlugins() = false for old format plugin, want true")
	}
}

// TestHasOldFormatPlugins_NonexistentDir tests hasOldFormatPlugins with non-existent dir.
func TestHasOldFormatPlugins_NonexistentDir(t *testing.T) {
	t.Parallel()

	fs := afero.NewOsFs()

	has, err := hasOldFormatPlugins(fs, "/nonexistent/path")
	if err != nil {
		t.Fatalf("hasOldFormatPlugins() error: %v", err)
	}
	if has {
		t.Error("hasOldFormatPlugins() = true for non-existent dir, want false")
	}
}

// TestMigrateEntry_SkipsDirectories tests migrateEntry skips directories.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestMigrateEntry_SkipsDirectories(t *testing.T) {
	fs := setupMigrateTestFs(t)
	pluginsDir := testMigrateDir + "/skipsdirs"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create a directory (should be skipped)
	pluginDir := filepath.Join(pluginsDir, "shelly-existing")
	if err := fs.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	entries, err := afero.ReadDir(fs, pluginsDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	result := &MigrationResult{}
	for _, entry := range entries {
		migrateEntry(fs, pluginsDir, entry, result)
	}

	// Directories should be skipped, not migrated
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated len = %d, want 0", len(result.Migrated))
	}
}

// TestMigrateEntry_SkipsNonPluginFiles tests migrateEntry skips non-plugin files.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestMigrateEntry_SkipsNonPluginFiles(t *testing.T) {
	fs := setupMigrateTestFs(t)
	pluginsDir := testMigrateDir + "/skipsnonplugin"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create a file without shelly- prefix (should be skipped)
	if err := afero.WriteFile(fs, filepath.Join(pluginsDir, "other-file"), []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	entries, err := afero.ReadDir(fs, pluginsDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	result := &MigrationResult{}
	for _, entry := range entries {
		migrateEntry(fs, pluginsDir, entry, result)
	}

	// Non-plugin files should be skipped
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated len = %d, want 0", len(result.Migrated))
	}
}

// TestMigrateEntry_SkipsHiddenFiles tests migrateEntry skips hidden files.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestMigrateEntry_SkipsHiddenFiles(t *testing.T) {
	fs := setupMigrateTestFs(t)
	pluginsDir := testMigrateDir + "/skipshidden"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create a hidden file (should be skipped)
	if err := afero.WriteFile(fs, filepath.Join(pluginsDir, ".hidden"), []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	entries, err := afero.ReadDir(fs, pluginsDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	result := &MigrationResult{}
	for _, entry := range entries {
		migrateEntry(fs, pluginsDir, entry, result)
	}

	// Hidden files should be skipped
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated len = %d, want 0", len(result.Migrated))
	}
}

// TestMigratePlugin_Success tests migratePlugin successful migration.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestMigratePlugin_Success(t *testing.T) {
	fs := setupMigrateTestFs(t)
	pluginsDir := testMigrateDir + "/success"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create an old-format plugin
	pluginName := "shelly-testplugin"
	oldPath := filepath.Join(pluginsDir, pluginName)
	if err := afero.WriteFile(fs, oldPath, []byte("#!/bin/bash\necho test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	result := &MigrationResult{}
	err := migratePlugin(fs, pluginsDir, pluginName, oldPath, result)
	if err != nil {
		t.Fatalf("migratePlugin() error: %v", err)
	}

	if len(result.Migrated) != 1 {
		t.Errorf("Migrated len = %d, want 1", len(result.Migrated))
	}

	// Verify new directory structure
	newDir := filepath.Join(pluginsDir, pluginName)
	if info, err := fs.Stat(newDir); err != nil || !info.IsDir() {
		t.Error("new plugin directory not created")
	}

	// Verify binary was moved
	newBinaryPath := filepath.Join(newDir, pluginName)
	if _, err := fs.Stat(newBinaryPath); err != nil {
		t.Error("binary not moved to new directory")
	}

	// Verify manifest was created
	manifestPath := filepath.Join(newDir, ManifestFileName)
	if _, err := fs.Stat(manifestPath); err != nil {
		t.Error("manifest not created")
	}
}

// TestCleanupMigrationFailure tests cleanupMigrationFailure function.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestCleanupMigrationFailure(t *testing.T) {
	fs := setupMigrateTestFs(t)
	pluginsDir := testMigrateDir + "/cleanup"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create temp file simulating migration in progress
	tempPath := filepath.Join(pluginsDir, "plugin.migrating")
	oldPath := filepath.Join(pluginsDir, "plugin")
	newDir := filepath.Join(pluginsDir, "plugin-dir")

	if err := afero.WriteFile(fs, tempPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := fs.MkdirAll(newDir, 0o750); err != nil {
		t.Fatalf("failed to create new dir: %v", err)
	}

	err := cleanupMigrationFailure(fs, tempPath, oldPath, newDir)
	if err != nil {
		t.Fatalf("cleanupMigrationFailure() error: %v", err)
	}

	// Verify temp file was renamed back
	if _, err := fs.Stat(oldPath); err != nil {
		t.Error("temp file not restored to original path")
	}

	// Verify new directory was removed
	if _, statErr := fs.Stat(newDir); statErr == nil {
		t.Error("new directory should have been removed")
	}
}

// TestCleanupMigrationFailure_PartialFailure tests cleanup with some failures.
func TestCleanupMigrationFailure_PartialFailure(t *testing.T) {
	t.Parallel()

	fs := afero.NewOsFs()

	// Using paths that don't exist should cause errors
	err := cleanupMigrationFailure(fs, "/nonexistent/temp", "/nonexistent/old", "/nonexistent/dir")
	// This may or may not error depending on whether the dir exists
	// The function should handle missing files gracefully - just verify it doesn't panic
	if err != nil {
		t.Logf("cleanup error (expected for nonexistent paths): %v", err)
	}
}
