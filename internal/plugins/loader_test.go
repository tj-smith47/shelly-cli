package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestLoader_Discover_Empty(t *testing.T) {
	t.Parallel()

	// Create temp dir with no plugins.
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	loader := &Loader{paths: []string{tmpDir}}
	plugins, err := loader.Discover()
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestLoader_Discover_FindsPlugins(t *testing.T) {
	t.Parallel()

	// Create temp dir with a fake plugin.
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create fake plugin executable.
	pluginPath := filepath.Join(tmpDir, "shelly-test-plugin")
	//nolint:gosec // Test file needs to be executable
	err = os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0o755)
	if err != nil {
		t.Fatalf("failed to create fake plugin: %v", err)
	}

	loader := &Loader{paths: []string{tmpDir}}
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

func TestLoader_Find(t *testing.T) {
	t.Parallel()

	// Create temp dir with a fake plugin.
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create fake plugin executable.
	pluginPath := filepath.Join(tmpDir, "shelly-myplugin")
	//nolint:gosec // Test file needs to be executable
	err = os.WriteFile(pluginPath, []byte("#!/bin/bash\necho test"), 0o755)
	if err != nil {
		t.Fatalf("failed to create fake plugin: %v", err)
	}

	loader := &Loader{paths: []string{tmpDir}}

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

func TestIsExecutable(t *testing.T) {
	t.Parallel()

	// Create temp dir.
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create executable file.
	execPath := filepath.Join(tmpDir, "executable")
	//nolint:gosec // Test file needs to be executable
	err = os.WriteFile(execPath, []byte("test"), 0o755)
	if err != nil {
		t.Fatalf("failed to create executable: %v", err)
	}

	// Create non-executable file.
	noExecPath := filepath.Join(tmpDir, "not-executable")
	//nolint:gosec // G306: Test file intentionally needs specific permissions (0o644)
	err = os.WriteFile(noExecPath, []byte("test"), 0o644)
	if err != nil {
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

	if isExecutable(tmpDir) {
		t.Error("isExecutable() returned true for directory")
	}
}
