package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry() error: %v", err)
	}

	dir := registry.PluginsDir()
	if dir == "" {
		t.Error("PluginsDir() returned empty string")
	}
}

func TestRegistry_Install(t *testing.T) {
	t.Parallel()

	// Create temp dirs for source and registry.
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	pluginsDir := filepath.Join(tmpDir, "plugins")
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(pluginsDir, 0o750); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0o750); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	registry := &Registry{pluginsDir: pluginsDir}

	// Create source plugin.
	sourcePath := filepath.Join(sourceDir, "shelly-testplugin")
	//nolint:gosec // Test file needs to be executable
	err = os.WriteFile(sourcePath, []byte("#!/bin/bash\necho test"), 0o755)
	if err != nil {
		t.Fatalf("failed to create source plugin: %v", err)
	}

	// Install.
	err = registry.Install(sourcePath)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Verify installed.
	installedPath := filepath.Join(pluginsDir, "shelly-testplugin")
	if _, err := os.Stat(installedPath); os.IsNotExist(err) {
		t.Error("plugin not installed to expected location")
	}
}

func TestRegistry_Install_InvalidName(t *testing.T) {
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

	registry := &Registry{pluginsDir: tmpDir}

	// Create source with wrong prefix.
	sourcePath := filepath.Join(tmpDir, "wrong-prefix")
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(sourcePath, []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	err = registry.Install(sourcePath)
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

func TestRegistry_Remove_NotFound(t *testing.T) {
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

	registry := &Registry{pluginsDir: tmpDir}

	err = registry.Remove("nonexistent")
	if err == nil {
		t.Error("Remove() should fail for nonexistent plugin")
	}
}

func TestRegistry_List(t *testing.T) {
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

	registry := &Registry{pluginsDir: tmpDir}

	// Empty list.
	plugins, err := registry.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}

	// Add some plugins.
	//nolint:gosec // Test files need to be executable
	if err := os.WriteFile(filepath.Join(tmpDir, "shelly-plugin1"), []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin1: %v", err)
	}
	//nolint:gosec // Test files need to be executable
	if err := os.WriteFile(filepath.Join(tmpDir, "shelly-plugin2"), []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create plugin2: %v", err)
	}
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(filepath.Join(tmpDir, "not-a-plugin"), []byte("test"), 0o755); err != nil {
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

func TestRegistry_IsInstalled(t *testing.T) {
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

	registry := &Registry{pluginsDir: tmpDir}

	// Create installed plugin.
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(filepath.Join(tmpDir, "shelly-installed"), []byte("test"), 0o755); err != nil {
		t.Fatalf("failed to create installed plugin: %v", err)
	}

	if !registry.IsInstalled("installed") {
		t.Error("IsInstalled() returned false for installed plugin")
	}

	if registry.IsInstalled("not-installed") {
		t.Error("IsInstalled() returned true for non-installed plugin")
	}
}
