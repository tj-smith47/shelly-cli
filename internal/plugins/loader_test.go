package plugins

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

const testLoaderDir = "/test/loader"

func setupLoaderTestFs(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })
	return fs
}

func TestNewLoader(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}
	if len(loader.paths) == 0 {
		t.Error("NewLoader() created loader with no paths")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoader_Discover_Empty(t *testing.T) {
	fs := setupLoaderTestFs(t)
	pluginsDir := testLoaderDir + "/empty"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	loader := &Loader{paths: []string{pluginsDir}}
	plugins, err := loader.Discover()
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoader_Discover_FindsPlugins(t *testing.T) {
	fs := setupLoaderTestFs(t)
	pluginsDir := testLoaderDir + "/finds"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create fake plugin executable.
	pluginPath := filepath.Join(pluginsDir, "shelly-test-plugin")
	if err := afero.WriteFile(fs, pluginPath, []byte("#!/bin/bash\necho test"), 0o755); err != nil {
		t.Fatalf("failed to create fake plugin: %v", err)
	}

	loader := &Loader{paths: []string{pluginsDir}}
	plugins, err := loader.Discover()
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}

	if plugins[0].Name != "test-plugin" {
		t.Errorf("expected plugin name 'test-plugin', got %q", plugins[0].Name)
	}
	if plugins[0].Path != pluginPath {
		t.Errorf("expected path %q, got %q", pluginPath, plugins[0].Path)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoader_Find(t *testing.T) {
	fs := setupLoaderTestFs(t)
	pluginsDir := testLoaderDir + "/find"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create fake plugin executable.
	pluginPath := filepath.Join(pluginsDir, "shelly-myplugin")
	if err := afero.WriteFile(fs, pluginPath, []byte("#!/bin/bash\necho test"), 0o755); err != nil {
		t.Fatalf("failed to create fake plugin: %v", err)
	}

	loader := &Loader{paths: []string{pluginsDir}}

	// Find by name without prefix.
	plugin, err := loader.Find("myplugin")
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if plugin == nil {
		t.Fatal("Find() returned nil for existing plugin")
	}
	if plugin.Name != "myplugin" {
		t.Errorf("expected name 'myplugin', got %q", plugin.Name)
	}

	// Find by name with prefix.
	plugin, err = loader.Find("shelly-myplugin")
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if plugin == nil {
		t.Fatal("Find() returned nil for existing plugin (with prefix)")
	}

	// Find nonexistent.
	plugin, err = loader.Find("nonexistent")
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if plugin != nil {
		t.Error("Find() should return nil for nonexistent plugin")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestIsExecutable(t *testing.T) {
	fs := setupLoaderTestFs(t)
	pluginsDir := testLoaderDir + "/executable"

	if err := fs.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	// Create executable file.
	execPath := filepath.Join(pluginsDir, "executable")
	if err := afero.WriteFile(fs, execPath, []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create executable: %v", err)
	}

	// Create non-executable file.
	noExecPath := filepath.Join(pluginsDir, "not-executable")
	if err := afero.WriteFile(fs, noExecPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create non-executable: %v", err)
	}

	if !isExecutable(execPath) {
		t.Error("isExecutable() returned false for executable file")
	}

	if isExecutable(noExecPath) {
		t.Error("isExecutable() returned true for non-executable file")
	}

	if isExecutable("/nonexistent/path") {
		t.Error("isExecutable() returned true for nonexistent path")
	}

	if isExecutable(pluginsDir) {
		t.Error("isExecutable() returned true for directory")
	}
}
