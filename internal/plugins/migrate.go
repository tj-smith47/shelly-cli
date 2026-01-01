// Package plugins provides plugin discovery, loading, and manifest management.
package plugins

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// ErrAlreadyMigrated is returned when migration has already been completed.
var ErrAlreadyMigrated = errors.New("plugins already migrated")

// MigrationResult holds the results of a migration operation.
type MigrationResult struct {
	Migrated []string // Names of successfully migrated plugins
	Skipped  []string // Names of plugins that were skipped
	Errors   []string // Error messages for failed migrations
}

// MigratePlugins migrates plugins from the old flat-binary format to the new
// directory-with-manifest format. This is a one-time migration that runs
// on first CLI invocation after upgrade.
//
// Old format: ~/.config/shelly/plugins/shelly-myext (bare binary).
// New format: ~/.config/shelly/plugins/shelly-myext/shelly-myext + manifest.json.
func MigratePlugins() (*MigrationResult, error) {
	pluginsDir, err := config.PluginsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get plugins directory: %w", err)
	}

	fs := config.Fs()

	// Ensure plugins directory exists
	if err := fs.MkdirAll(pluginsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create plugins directory: %w", err)
	}

	markerFile := filepath.Join(pluginsDir, MigrationMarkerFile)

	// Skip if already migrated
	if _, err := fs.Stat(markerFile); err == nil {
		return nil, ErrAlreadyMigrated
	}

	result := &MigrationResult{}

	entries, err := afero.ReadDir(fs, pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		migrateEntry(fs, pluginsDir, entry, result)
	}

	// Create marker file to prevent re-migration
	if err := afero.WriteFile(fs, markerFile, []byte(ManifestSchemaVersion), 0o644); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("warning: failed to create migration marker: %v", err))
	}

	return result, nil
}

// migrateEntry handles migration of a single plugin entry.
func migrateEntry(fs afero.Fs, pluginsDir string, entry os.FileInfo, result *MigrationResult) {
	name := entry.Name()

	// Skip directories, non-plugin files, and hidden files
	if entry.IsDir() || !strings.HasPrefix(name, PluginPrefix) || strings.HasPrefix(name, ".") {
		return
	}

	oldPath := filepath.Join(pluginsDir, name)
	info, err := fs.Stat(oldPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", name, err))
		return
	}

	if info.IsDir() {
		result.Skipped = append(result.Skipped, name)
		return
	}

	if err := migratePlugin(fs, pluginsDir, name, oldPath, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", name, err))
	}
}

// migratePlugin performs the actual migration of a single plugin.
func migratePlugin(fs afero.Fs, pluginsDir, name, oldPath string, result *MigrationResult) error {
	newDir := filepath.Join(pluginsDir, name)
	newPath := filepath.Join(newDir, name)

	// Create a temp name to avoid conflict
	tempPath := oldPath + ".migrating"
	if err := fs.Rename(oldPath, tempPath); err != nil {
		return fmt.Errorf("failed to prepare: %w", err)
	}

	// Create directory
	if err := fs.MkdirAll(newDir, 0o755); err != nil {
		restoreErr := fs.Rename(tempPath, oldPath)
		if restoreErr != nil {
			return errors.Join(
				fmt.Errorf("failed to create directory: %w", err),
				fmt.Errorf("restore also failed: %w", restoreErr),
			)
		}
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Move binary into directory
	if err := fs.Rename(tempPath, newPath); err != nil {
		if cleanupErr := cleanupMigrationFailure(fs, tempPath, oldPath, newDir); cleanupErr != nil {
			return errors.Join(
				fmt.Errorf("failed to move binary: %w", err),
				fmt.Errorf("cleanup failed: %w", cleanupErr),
			)
		}
		return fmt.Errorf("failed to move binary: %w", err)
	}

	// Create manifest
	pluginName := strings.TrimPrefix(name, PluginPrefix)
	manifest := NewManifest(pluginName, UnknownSource())
	manifest.Binary.Name = name
	manifest.Binary.Platform = runtime.GOOS + "-" + runtime.GOARCH

	// Compute checksum (non-fatal if fails)
	if err := manifest.SetBinaryInfo(newPath); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: warning: checksum failed: %v", name, err))
	}

	// Save manifest (non-fatal if fails)
	if err := manifest.Save(newDir); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: warning: manifest save failed: %v", name, err))
	}

	result.Migrated = append(result.Migrated, pluginName)
	return nil
}

// cleanupMigrationFailure attempts to restore state after a failed migration.
// Returns any errors encountered during cleanup.
func cleanupMigrationFailure(fs afero.Fs, tempPath, oldPath, newDir string) error {
	var errs []string

	// Try to remove the new directory first
	if err := fs.RemoveAll(newDir); err != nil {
		errs = append(errs, fmt.Sprintf("remove dir: %v", err))
	}

	// Try to restore the original file
	if err := fs.Rename(tempPath, oldPath); err != nil {
		errs = append(errs, fmt.Sprintf("restore file: %v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// NeedsMigration checks if plugin migration is needed.
func NeedsMigration() (bool, error) {
	pluginsDir, err := config.PluginsDir()
	if err != nil {
		return false, err
	}

	fs := config.Fs()
	markerFile := filepath.Join(pluginsDir, MigrationMarkerFile)
	if _, err := fs.Stat(markerFile); err == nil {
		return false, nil // Marker exists, no migration needed
	}

	return hasOldFormatPlugins(fs, pluginsDir)
}

// hasOldFormatPlugins checks if there are any old-format (bare binary) plugins.
func hasOldFormatPlugins(fs afero.Fs, pluginsDir string) (bool, error) {
	entries, err := afero.ReadDir(fs, pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // No plugins dir, nothing to migrate
		}
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), PluginPrefix) {
			return true, nil // Found an old-format plugin
		}
	}
	return false, nil
}
