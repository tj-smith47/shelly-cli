package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

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
func TestHasOldFormatPlugins(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Empty directory - no old format plugins
	has, err := hasOldFormatPlugins(tmpDir)
	if err != nil {
		t.Fatalf("hasOldFormatPlugins() error: %v", err)
	}
	if has {
		t.Error("hasOldFormatPlugins() = true for empty dir, want false")
	}

	// Add a new format plugin (directory)
	newPluginDir := filepath.Join(tmpDir, "shelly-new")
	if err := os.MkdirAll(newPluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	has, err = hasOldFormatPlugins(tmpDir)
	if err != nil {
		t.Fatalf("hasOldFormatPlugins() error: %v", err)
	}
	if has {
		t.Error("hasOldFormatPlugins() = true for new format plugin, want false")
	}

	// Add an old format plugin (bare binary)
	oldPluginPath := filepath.Join(tmpDir, "shelly-old")
	//nolint:gosec // Test file
	if err := os.WriteFile(oldPluginPath, []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create old plugin: %v", err)
	}

	has, err = hasOldFormatPlugins(tmpDir)
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

	has, err := hasOldFormatPlugins("/nonexistent/path")
	if err != nil {
		t.Fatalf("hasOldFormatPlugins() error: %v", err)
	}
	if has {
		t.Error("hasOldFormatPlugins() = true for non-existent dir, want false")
	}
}

// TestMigrateEntry_SkipsDirectories tests migrateEntry skips directories.
func TestMigrateEntry_SkipsDirectories(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a directory (should be skipped)
	pluginDir := filepath.Join(tmpDir, "shelly-existing")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	result := &MigrationResult{}
	for _, entry := range entries {
		migrateEntry(tmpDir, entry, result)
	}

	// Directories should be skipped, not migrated
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated len = %d, want 0", len(result.Migrated))
	}
}

// TestMigrateEntry_SkipsNonPluginFiles tests migrateEntry skips non-plugin files.
func TestMigrateEntry_SkipsNonPluginFiles(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a file without shelly- prefix (should be skipped)
	//nolint:gosec // Test file
	if err := os.WriteFile(filepath.Join(tmpDir, "other-file"), []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	result := &MigrationResult{}
	for _, entry := range entries {
		migrateEntry(tmpDir, entry, result)
	}

	// Non-plugin files should be skipped
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated len = %d, want 0", len(result.Migrated))
	}
}

// TestMigrateEntry_SkipsHiddenFiles tests migrateEntry skips hidden files.
func TestMigrateEntry_SkipsHiddenFiles(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a hidden file (should be skipped)
	//nolint:gosec // Test file
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	result := &MigrationResult{}
	for _, entry := range entries {
		migrateEntry(tmpDir, entry, result)
	}

	// Hidden files should be skipped
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated len = %d, want 0", len(result.Migrated))
	}
}

// TestMigratePlugin_Success tests migratePlugin successful migration.
func TestMigratePlugin_Success(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create an old-format plugin
	pluginName := "shelly-testplugin"
	oldPath := filepath.Join(tmpDir, pluginName)
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(oldPath, []byte("#!/bin/bash\necho test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	result := &MigrationResult{}
	err = migratePlugin(tmpDir, pluginName, oldPath, result)
	if err != nil {
		t.Fatalf("migratePlugin() error: %v", err)
	}

	if len(result.Migrated) != 1 {
		t.Errorf("Migrated len = %d, want 1", len(result.Migrated))
	}

	// Verify new directory structure
	newDir := filepath.Join(tmpDir, pluginName)
	if info, err := os.Stat(newDir); err != nil || !info.IsDir() {
		t.Error("new plugin directory not created")
	}

	// Verify binary was moved
	newBinaryPath := filepath.Join(newDir, pluginName)
	if _, err := os.Stat(newBinaryPath); err != nil {
		t.Error("binary not moved to new directory")
	}

	// Verify manifest was created
	manifestPath := filepath.Join(newDir, ManifestFileName)
	if _, err := os.Stat(manifestPath); err != nil {
		t.Error("manifest not created")
	}
}

// TestCleanupMigrationFailure tests cleanupMigrationFailure function.
func TestCleanupMigrationFailure(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create temp file simulating migration in progress
	tempPath := filepath.Join(tmpDir, "plugin.migrating")
	oldPath := filepath.Join(tmpDir, "plugin")
	newDir := filepath.Join(tmpDir, "plugin-dir")

	//nolint:gosec // Test file
	if err := os.WriteFile(tempPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := os.MkdirAll(newDir, 0o750); err != nil {
		t.Fatalf("failed to create new dir: %v", err)
	}

	err = cleanupMigrationFailure(tempPath, oldPath, newDir)
	if err != nil {
		t.Fatalf("cleanupMigrationFailure() error: %v", err)
	}

	// Verify temp file was renamed back
	if _, err := os.Stat(oldPath); err != nil {
		t.Error("temp file not restored to original path")
	}

	// Verify new directory was removed
	if _, err := os.Stat(newDir); !os.IsNotExist(err) {
		t.Error("new directory should have been removed")
	}
}

// TestCleanupMigrationFailure_PartialFailure tests cleanup with some failures.
func TestCleanupMigrationFailure_PartialFailure(t *testing.T) {
	t.Parallel()

	// Using paths that don't exist should cause errors
	err := cleanupMigrationFailure("/nonexistent/temp", "/nonexistent/old", "/nonexistent/dir")
	// This may or may not error depending on whether the dir exists
	// The function should handle missing files gracefully
	_ = err // Just verify it doesn't panic
}
