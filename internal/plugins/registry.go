// Package plugins provides plugin management functionality.
package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Registry manages installed plugins.
type Registry struct {
	pluginsDir string
}

// NewRegistry creates a new plugin registry.
func NewRegistry() (*Registry, error) {
	pluginsDir, err := config.PluginsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get plugins directory: %w", err)
	}

	// Ensure plugins directory exists
	if err := os.MkdirAll(pluginsDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create plugins directory: %w", err)
	}

	return &Registry{
		pluginsDir: pluginsDir,
	}, nil
}

// PluginsDir returns the plugins directory path.
func (r *Registry) PluginsDir() string {
	return r.pluginsDir
}

// InstallWithManifest installs a plugin with manifest metadata.
func (r *Registry) InstallWithManifest(sourcePath string, manifest *Manifest) error {
	filename := filepath.Base(sourcePath)

	// Ensure it has the correct prefix
	if !strings.HasPrefix(filename, PluginPrefix) {
		return fmt.Errorf("plugin must be named with prefix %q (got %q)", PluginPrefix, filename)
	}

	pluginName := strings.TrimPrefix(filename, PluginPrefix)
	pluginDir := filepath.Join(r.pluginsDir, PluginPrefix+pluginName)
	destPath := filepath.Join(pluginDir, filename)

	// Create plugin directory
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Read source file
	//nolint:gosec // G304: sourcePath is user-provided intentionally (install command)
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		// Clean up directory on failure
		if cleanupErr := os.RemoveAll(pluginDir); cleanupErr != nil {
			return fmt.Errorf("failed to read plugin (cleanup also failed: %w): %w", cleanupErr, err)
		}
		return fmt.Errorf("failed to read plugin: %w", err)
	}

	// Write binary
	//nolint:gosec // G306: Plugins need executable permission
	if err := os.WriteFile(destPath, data, 0o700); err != nil {
		if cleanupErr := os.RemoveAll(pluginDir); cleanupErr != nil {
			return fmt.Errorf("failed to install plugin (cleanup also failed: %w): %w", cleanupErr, err)
		}
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	// Update manifest with binary info
	manifest.Binary.Name = filename
	if err := manifest.SetBinaryInfo(destPath); err != nil {
		// Non-fatal, continue without checksum
		manifest.Binary.Checksum = ""
	}

	// Save manifest
	if err := manifest.Save(pluginDir); err != nil {
		// Non-fatal but worth noting - plugin is installed but without manifest
		return fmt.Errorf("plugin installed but manifest save failed: %w", err)
	}

	return nil
}

// Install installs a plugin from a local file (legacy API, uses unknown source).
func (r *Registry) Install(sourcePath string) error {
	filename := filepath.Base(sourcePath)
	pluginName := strings.TrimPrefix(filename, PluginPrefix)
	manifest := NewManifest(pluginName, ParseLocalSource(sourcePath))
	return r.InstallWithManifest(sourcePath, manifest)
}

// Remove removes an installed plugin.
func (r *Registry) Remove(name string) error {
	// Find the plugin
	loader := NewLoader()
	plugin, err := loader.Find(name)
	if err != nil {
		return fmt.Errorf("error finding plugin: %w", err)
	}
	if plugin == nil {
		return fmt.Errorf("plugin %q not found", name)
	}

	// Only remove if it's in our plugins directory
	if !strings.HasPrefix(plugin.Path, r.pluginsDir) {
		return fmt.Errorf("cannot remove plugin %q: not installed in user plugins directory (found at %s)", name, plugin.Path)
	}

	// If plugin has a directory (new format), remove the directory
	if plugin.Dir != "" {
		if err := os.RemoveAll(plugin.Dir); err != nil {
			return fmt.Errorf("failed to remove plugin directory: %w", err)
		}
		return nil
	}

	// Old format: just remove the file
	if err := os.Remove(plugin.Path); err != nil {
		return fmt.Errorf("failed to remove plugin: %w", err)
	}

	return nil
}

// List returns all installed plugins (in user plugins directory).
func (r *Registry) List() ([]Plugin, error) {
	entries, err := os.ReadDir(r.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Empty list is valid when directory doesn't exist
		}
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	result := make([]Plugin, 0, len(entries))
	loader := &Loader{paths: []string{r.pluginsDir}}
	// Use background context for list; cancellation happens via timeout in getPluginVersion
	ctx := context.Background()

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, PluginPrefix) {
			continue
		}

		pluginName := strings.TrimPrefix(name, PluginPrefix)

		// New format: directory with manifest
		if entry.IsDir() {
			pluginDir := filepath.Join(r.pluginsDir, name)
			plugin := loader.loadFromDir(pluginDir, pluginName)
			if plugin != nil {
				result = append(result, *plugin)
			}
			continue
		}

		// Old format: bare binary
		fullPath := filepath.Join(r.pluginsDir, name)
		if !isExecutable(fullPath) {
			continue
		}

		result = append(result, Plugin{
			Name:    pluginName,
			Path:    fullPath,
			Version: getPluginVersion(ctx, fullPath),
		})
	}

	return result, nil
}

// IsInstalled checks if a plugin is installed in the user plugins directory.
func (r *Registry) IsInstalled(name string) bool {
	dirPath := filepath.Join(r.pluginsDir, PluginPrefix+name)

	// Check new format (directory)
	if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
		return true
	}

	// Check old format (bare binary)
	if isExecutable(dirPath) {
		return true
	}

	// Check with .exe extension
	if isExecutable(dirPath + ".exe") {
		return true
	}

	return false
}

// GetManifest returns the manifest for an installed plugin.
func (r *Registry) GetManifest(name string) (*Manifest, error) {
	pluginDir := filepath.Join(r.pluginsDir, PluginPrefix+name)

	// Check if directory exists
	info, err := os.Stat(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plugin %q not found", name)
		}
		return nil, err
	}

	if !info.IsDir() {
		// Old format - no manifest
		return nil, fmt.Errorf("plugin %q uses old format (no manifest)", name)
	}

	return LoadManifest(pluginDir)
}

// FindByPlatform returns the plugin that manages a given platform.
// Returns nil if no plugin is found for the platform (not an error).
func (r *Registry) FindByPlatform(platform string) (*Plugin, error) {
	plugins, err := r.List()
	if err != nil {
		return nil, err
	}

	for i := range plugins {
		p := &plugins[i]
		if p.Manifest != nil && p.Manifest.Capabilities != nil {
			if p.Manifest.Capabilities.Platform == platform {
				return p, nil
			}
		}
	}

	return nil, nil //nolint:nilnil // Not found is valid, not an error
}

// ListDetectionCapable returns all plugins that can detect devices.
// These are plugins with DeviceDetection capability set to true.
func (r *Registry) ListDetectionCapable() ([]Plugin, error) {
	plugins, err := r.List()
	if err != nil {
		return nil, err
	}

	var result []Plugin
	for _, p := range plugins {
		if p.Manifest != nil && p.Manifest.Capabilities != nil {
			if p.Manifest.Capabilities.DeviceDetection {
				result = append(result, p)
			}
		}
	}

	return result, nil
}
