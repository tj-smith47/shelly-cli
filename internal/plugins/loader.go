// Package plugins provides plugin discovery and loading functionality.
package plugins

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

const (
	// PluginPrefix is the prefix for plugin executables.
	PluginPrefix = "shelly-"
)

// Plugin represents a discovered plugin.
type Plugin struct {
	Name    string // Plugin name (without shelly- prefix)
	Path    string // Full path to executable
	Version string // Plugin version (if available)
}

// Loader discovers and loads plugins.
type Loader struct {
	paths []string
}

// NewLoader creates a new plugin loader.
func NewLoader() *Loader {
	return &Loader{
		paths: getSearchPaths(),
	}
}

// getSearchPaths returns all paths to search for plugins.
func getSearchPaths() []string {
	var paths []string

	// Add user plugin directory first (highest priority)
	if pluginsDir, err := config.PluginsDir(); err == nil {
		paths = append(paths, pluginsDir)
	}

	// Add custom paths from config
	cfg := config.Get()
	paths = append(paths, cfg.Plugins.Path...)

	// Add PATH directories
	if pathEnv := os.Getenv("PATH"); pathEnv != "" {
		for _, p := range strings.Split(pathEnv, string(os.PathListSeparator)) {
			paths = append(paths, p)
		}
	}

	return paths
}

// Discover finds all available plugins.
func (l *Loader) Discover() ([]Plugin, error) {
	seen := make(map[string]bool)
	var plugins []Plugin

	for _, dir := range l.paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories that don't exist or can't be read
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasPrefix(name, PluginPrefix) {
				continue
			}

			// Extract plugin name
			pluginName := strings.TrimPrefix(name, PluginPrefix)
			// Remove extension if present (e.g., .exe on Windows)
			if ext := filepath.Ext(pluginName); ext != "" {
				pluginName = strings.TrimSuffix(pluginName, ext)
			}

			// Skip if we've already seen this plugin (earlier paths take priority)
			if seen[pluginName] {
				continue
			}

			fullPath := filepath.Join(dir, name)

			// Check if executable
			if !isExecutable(fullPath) {
				continue
			}

			seen[pluginName] = true
			plugins = append(plugins, Plugin{
				Name: pluginName,
				Path: fullPath,
			})
		}
	}

	// Try to get versions for each plugin
	for i := range plugins {
		plugins[i].Version = getPluginVersion(plugins[i].Path)
	}

	return plugins, nil
}

// Find finds a specific plugin by name.
func (l *Loader) Find(name string) (*Plugin, error) {
	// Check for direct plugin name
	if p := l.findByName(name); p != nil {
		return p, nil
	}

	// Check with prefix
	if !strings.HasPrefix(name, PluginPrefix) {
		if p := l.findByName(PluginPrefix + name); p != nil {
			return p, nil
		}
	}

	return nil, nil
}

func (l *Loader) findByName(name string) *Plugin {
	for _, dir := range l.paths {
		path := filepath.Join(dir, name)

		// Try exact name
		if isExecutable(path) {
			pluginName := strings.TrimPrefix(name, PluginPrefix)
			return &Plugin{
				Name:    pluginName,
				Path:    path,
				Version: getPluginVersion(path),
			}
		}

		// Try with .exe extension (Windows)
		pathExe := path + ".exe"
		if isExecutable(pathExe) {
			pluginName := strings.TrimPrefix(name, PluginPrefix)
			return &Plugin{
				Name:    pluginName,
				Path:    pathExe,
				Version: getPluginVersion(pathExe),
			}
		}
	}
	return nil
}

// isExecutable checks if a file exists and is executable.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}

	// On Unix, check executable bit
	// On Windows, all files are considered executable
	mode := info.Mode()
	return mode&0o111 != 0 || strings.HasSuffix(strings.ToLower(path), ".exe")
}

// getPluginVersion attempts to get the version of a plugin by running it with --version.
func getPluginVersion(path string) string {
	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(output))
	// Take first line only
	if idx := strings.IndexByte(version, '\n'); idx != -1 {
		version = version[:idx]
	}
	return version
}
