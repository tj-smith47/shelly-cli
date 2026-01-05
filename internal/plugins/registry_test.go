package plugins

import (
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

const testPluginsDir = "/test/plugins"

func TestNewRegistry(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry() error: %v", err)
	}

	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if registry.pluginsDir == "" {
		t.Error("NewRegistry() created registry with empty pluginsDir")
	}
}

func TestRegistry_PluginsDir(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry() error: %v", err)
	}

	dir := registry.PluginsDir()
	if dir == "" {
		t.Error("PluginsDir() returned empty string")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRegistry_Install(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Create directories
	sourceDir := "/test/source"
	if err := fs.MkdirAll(testPluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}
	if err := fs.MkdirAll(sourceDir, 0o750); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	registry := &Registry{pluginsDir: testPluginsDir}

	// Create source plugin.
	sourcePath := sourceDir + "/shelly-testplugin"
	if err := afero.WriteFile(fs, sourcePath, []byte("#!/bin/bash\necho test"), 0o755); err != nil {
		t.Fatalf("failed to create source plugin: %v", err)
	}

	// Install.
	err := registry.Install(sourcePath)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Verify installed (new format creates directory with binary inside)
	installedDir := testPluginsDir + "/shelly-testplugin"
	installedPath := installedDir + "/shelly-testplugin"
	if _, err := fs.Stat(installedPath); err != nil {
		t.Error("plugin not installed to expected location")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRegistry_Install_InvalidName(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testPluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	registry := &Registry{pluginsDir: testPluginsDir}

	// Create source with wrong prefix.
	sourcePath := testPluginsDir + "/wrong-prefix"
	if err := afero.WriteFile(fs, sourcePath, []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	err := registry.Install(sourcePath)
	if err == nil {
		t.Error("Install() should fail for plugin without shelly- prefix")
	}
}

//nolint:paralleltest // Test skipped - uses NewLoader with system paths
func TestRegistry_Remove(t *testing.T) {
	// Skip: Remove() uses NewLoader() which searches in default system paths,
	// making unit testing difficult without environment manipulation.
	// The integration works - Remove() is tested indirectly via CLI tests.
	t.Skip("Remove() requires integration test setup - uses NewLoader() which searches system paths")
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRegistry_Remove_NotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testPluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	registry := &Registry{pluginsDir: testPluginsDir}

	err := registry.Remove("nonexistent")
	if err == nil {
		t.Error("Remove() should fail for nonexistent plugin")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRegistry_List(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testPluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	registry := &Registry{pluginsDir: testPluginsDir}

	// Empty list.
	plugins, err := registry.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}

	// Add some plugins.
	if err := afero.WriteFile(fs, testPluginsDir+"/shelly-plugin1", []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin1: %v", err)
	}
	if err := afero.WriteFile(fs, testPluginsDir+"/shelly-plugin2", []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin2: %v", err)
	}
	if err := afero.WriteFile(fs, testPluginsDir+"/not-a-plugin", []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create not-a-plugin: %v", err)
	}

	plugins, err = registry.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRegistry_IsInstalled(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testPluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	registry := &Registry{pluginsDir: testPluginsDir}

	// Create installed plugin.
	if err := afero.WriteFile(fs, testPluginsDir+"/shelly-installed", []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create installed plugin: %v", err)
	}

	if !registry.IsInstalled("installed") {
		t.Error("IsInstalled() returned false for installed plugin")
	}

	if registry.IsInstalled("not-installed") {
		t.Error("IsInstalled() returned true for non-installed plugin")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRegistry_FindByPlatform(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testPluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	registry := &Registry{pluginsDir: testPluginsDir}

	// Create plugin directory with manifest
	pluginDir := testPluginsDir + "/shelly-tasmota"
	if err := fs.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create executable
	if err := afero.WriteFile(fs, pluginDir+"/shelly-tasmota", []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin binary: %v", err)
	}

	// Create manifest with platform capability
	manifest := `{
		"schema_version": "1",
		"name": "tasmota",
		"binary": {
			"name": "shelly-tasmota"
		},
		"capabilities": {
			"platform": "tasmota",
			"device_detection": true
		}
	}`
	if err := afero.WriteFile(fs, pluginDir+"/manifest.json", []byte(manifest), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Test finding by platform
	plugin, err := registry.FindByPlatform("tasmota")
	if err != nil {
		t.Fatalf("FindByPlatform() error: %v", err)
	}
	if plugin == nil {
		t.Fatal("FindByPlatform() returned nil for existing platform")
	}
	if plugin.Name != testPlatformTasmota {
		t.Errorf("expected plugin name %q, got %q", testPlatformTasmota, plugin.Name)
	}

	// Test finding non-existent platform
	plugin, err = registry.FindByPlatform("esphome")
	if err != nil {
		t.Fatalf("FindByPlatform() error for non-existent platform: %v", err)
	}
	if plugin != nil {
		t.Error("FindByPlatform() should return nil for non-existent platform")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRegistry_ListDetectionCapable(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testPluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	registry := &Registry{pluginsDir: testPluginsDir}

	// Create plugin with detection capability
	pluginDir1 := testPluginsDir + "/shelly-detector"
	if err := fs.MkdirAll(pluginDir1, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}
	if err := afero.WriteFile(fs, pluginDir1+"/shelly-detector", []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin binary: %v", err)
	}
	manifest1 := `{
		"schema_version": "1",
		"name": "detector",
		"binary": {
			"name": "shelly-detector"
		},
		"capabilities": {
			"device_detection": true
		}
	}`
	if err := afero.WriteFile(fs, pluginDir1+"/manifest.json", []byte(manifest1), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Create plugin without detection capability
	pluginDir2 := testPluginsDir + "/shelly-other"
	if err := fs.MkdirAll(pluginDir2, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}
	if err := afero.WriteFile(fs, pluginDir2+"/shelly-other", []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin binary: %v", err)
	}
	manifest2 := `{
		"schema_version": "1",
		"name": "other",
		"binary": {
			"name": "shelly-other"
		},
		"capabilities": {
			"device_detection": false
		}
	}`
	if err := afero.WriteFile(fs, pluginDir2+"/manifest.json", []byte(manifest2), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Test listing detection-capable plugins
	plugins, err := registry.ListDetectionCapable()
	if err != nil {
		t.Fatalf("ListDetectionCapable() error: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("expected 1 detection-capable plugin, got %d", len(plugins))
	}
	if len(plugins) > 0 && plugins[0].Name != "detector" {
		t.Errorf("expected plugin 'detector', got %q", plugins[0].Name)
	}
}
