// Package plugins provides plugin discovery and loading functionality.
package plugins

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

const (
	// PluginPrefix is the prefix for plugin executables.
	PluginPrefix = "shelly-"
)

// Plugin represents a discovered plugin.
type Plugin struct {
	Name     string    // Plugin name (without shelly- prefix)
	Path     string    // Full path to executable
	Version  string    // Plugin version (if available)
	Dir      string    // Plugin directory (new format only)
	Manifest *Manifest // Manifest (new format only)
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
		paths = append(paths, strings.Split(pathEnv, string(os.PathListSeparator))...)
	}

	return paths
}

// Discover finds all available plugins.
// Supports both new format (directory with manifest) and old format (bare binary).
func (l *Loader) Discover() ([]Plugin, error) {
	seen := make(map[string]bool)
	var plugins []Plugin

	for _, dir := range l.paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories that don't exist or can't be read
		}

		for _, entry := range entries {
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

			// Check if this is the new format (directory with manifest)
			if entry.IsDir() {
				pluginDir := filepath.Join(dir, name)
				plugin := l.loadFromDir(pluginDir, pluginName)
				if plugin != nil {
					seen[pluginName] = true
					plugins = append(plugins, *plugin)
				}
				continue
			}

			// Old format: bare binary
			fullPath := filepath.Join(dir, name)
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

	// Try to get versions for plugins without manifests concurrently
	var wg sync.WaitGroup
	for i := range plugins {
		if plugins[i].Version == "" && plugins[i].Manifest == nil {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				plugins[idx].Version = getPluginVersion(plugins[idx].Path)
			}(i)
		}
	}
	wg.Wait()

	return plugins, nil
}

// loadFromDir loads a plugin from a directory (new format).
func (l *Loader) loadFromDir(pluginDir, pluginName string) *Plugin {
	// Try to load manifest
	manifest, err := LoadManifest(pluginDir)
	if err != nil {
		// No manifest, try to find binary in directory
		binaryName := PluginPrefix + pluginName
		binaryPath := filepath.Join(pluginDir, binaryName)
		if !isExecutable(binaryPath) {
			// Try with .exe
			binaryPath += ".exe"
			if !isExecutable(binaryPath) {
				return nil
			}
		}
		return &Plugin{
			Name: pluginName,
			Path: binaryPath,
			Dir:  pluginDir,
		}
	}

	// Have manifest, use it
	binaryPath := manifest.BinaryPath(pluginDir)
	if !isExecutable(binaryPath) {
		// Try with .exe
		binaryPath += ".exe"
		if !isExecutable(binaryPath) {
			return nil
		}
	}

	return &Plugin{
		Name:     pluginName,
		Path:     binaryPath,
		Version:  manifest.Version,
		Dir:      pluginDir,
		Manifest: manifest,
	}
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

	return nil, nil //nolint:nilnil // Plugin not found is valid, not an error
}

func (l *Loader) findByName(name string) *Plugin {
	pluginName := strings.TrimPrefix(name, PluginPrefix)

	for _, dir := range l.paths {
		// First, check for new format (directory)
		pluginDir := filepath.Join(dir, name)
		if info, err := os.Stat(pluginDir); err == nil && info.IsDir() {
			if plugin := l.loadFromDir(pluginDir, pluginName); plugin != nil {
				return plugin
			}
		}

		// Then check old format (bare binary)
		path := filepath.Join(dir, name)
		if isExecutable(path) {
			return &Plugin{
				Name:    pluginName,
				Path:    path,
				Version: getPluginVersion(path),
			}
		}

		// Try with .exe extension (Windows)
		pathExe := path + ".exe"
		if isExecutable(pathExe) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
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
