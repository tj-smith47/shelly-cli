// Package plugins provides plugin management functionality.
package plugins

import (
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
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
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

// Install installs a plugin from a local file.
func (r *Registry) Install(sourcePath string) error {
	// Get the filename
	filename := filepath.Base(sourcePath)

	// Ensure it has the correct prefix
	if !strings.HasPrefix(filename, PluginPrefix) {
		return fmt.Errorf("plugin must be named with prefix %q (got %q)", PluginPrefix, filename)
	}

	// Destination path
	destPath := filepath.Join(r.pluginsDir, filename)

	// Copy the file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read plugin: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0o755); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	return nil
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

	// Remove the file
	if err := os.Remove(plugin.Path); err != nil {
		return fmt.Errorf("failed to remove plugin: %w", err)
	}

	return nil
}

// List returns all installed plugins (in user plugins directory).
func (r *Registry) List() ([]Plugin, error) {
	var plugins []Plugin

	entries, err := os.ReadDir(r.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return plugins, nil
		}
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, PluginPrefix) {
			continue
		}

		pluginName := strings.TrimPrefix(name, PluginPrefix)
		fullPath := filepath.Join(r.pluginsDir, name)

		if !isExecutable(fullPath) {
			continue
		}

		plugins = append(plugins, Plugin{
			Name:    pluginName,
			Path:    fullPath,
			Version: getPluginVersion(fullPath),
		})
	}

	return plugins, nil
}

// IsInstalled checks if a plugin is installed in the user plugins directory.
func (r *Registry) IsInstalled(name string) bool {
	// Check exact name
	path := filepath.Join(r.pluginsDir, PluginPrefix+name)
	if isExecutable(path) {
		return true
	}

	// Check with .exe extension
	if isExecutable(path + ".exe") {
		return true
	}

	return false
}
