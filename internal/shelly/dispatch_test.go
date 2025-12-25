// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

func TestSupportsPluginCommand_WithHints(t *testing.T) {
	t.Parallel()

	// Create temp directory for plugin
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create plugin directory with manifest containing hints
	// Directory must have "shelly-" prefix to be found by registry
	pluginDir := filepath.Join(tmpDir, "shelly-tasmota")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	manifest := plugins.Manifest{
		SchemaVersion: "1",
		Name:          "tasmota",
		Version:       "1.0.0",
		Capabilities: &plugins.Capabilities{
			Platform:   "tasmota",
			Components: []string{"switch", "light"},
			Hints: map[string]string{
				"scene":    "Tasmota uses Rules for automation",
				"script":   "Tasmota uses Berry scripting on ESP32",
				"schedule": "Tasmota uses Timers for scheduling",
			},
		},
		Hooks: &plugins.Hooks{
			Control: "./shelly-tasmota control",
			Status:  "./shelly-tasmota status",
		},
		Binary: plugins.Binary{
			Name:     "shelly-tasmota",
			Checksum: "sha256:abc123",
		},
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestData, 0o600); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Create a dummy binary so the plugin is considered valid
	binaryPath := filepath.Join(pluginDir, "shelly-tasmota")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho test"), 0o700); err != nil { //nolint:gosec // G306: test binary needs execute permission
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create registry pointing to our temp dir
	registry := plugins.NewRegistryWithDir(tmpDir)

	// Create service with registry
	service := NewService()
	service.SetPluginRegistry(registry)

	// Test device (plugin-managed)
	device := model.Device{
		Name:     "test-tasmota",
		Address:  "192.168.1.100",
		Platform: "tasmota",
	}

	t.Run("supported command - control", func(t *testing.T) {
		t.Parallel()
		err := service.SupportsPluginCommand(device, "on")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("supported command - status", func(t *testing.T) {
		t.Parallel()
		err := service.SupportsPluginCommand(device, "status")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unsupported command with hint - scene", func(t *testing.T) {
		t.Parallel()
		err := service.SupportsPluginCommand(device, "scene")
		assertPlatformErrorWithHint(t, err, "Tasmota uses Rules for automation")
	})

	t.Run("unsupported command with hint - script", func(t *testing.T) {
		t.Parallel()
		err := service.SupportsPluginCommand(device, "script")
		assertPlatformErrorWithHint(t, err, "Tasmota uses Berry scripting on ESP32")
	})

	t.Run("unsupported command with hint - schedule", func(t *testing.T) {
		t.Parallel()
		err := service.SupportsPluginCommand(device, "schedule")
		assertPlatformErrorWithHint(t, err, "Tasmota uses Timers for scheduling")
	})

	t.Run("unsupported command without hint - unknown", func(t *testing.T) {
		t.Parallel()
		err := service.SupportsPluginCommand(device, "unknown")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var platErr *PlatformError
		if !errors.As(err, &platErr) {
			t.Fatalf("expected *PlatformError, got %T", err)
		}
		if platErr.Hint != "" {
			t.Errorf("expected no hint, got %q", platErr.Hint)
		}
	})
}

// assertPlatformErrorWithHint checks that err is a PlatformError with the expected hint.
func assertPlatformErrorWithHint(t *testing.T, err error, wantHint string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var platErr *PlatformError
	if !errors.As(err, &platErr) {
		t.Fatalf("expected *PlatformError, got %T", err)
	}
	if platErr.Hint != wantHint {
		t.Errorf("hint = %q, want %q", platErr.Hint, wantHint)
	}
	if !strings.Contains(platErr.Error(), wantHint) {
		t.Errorf("Error() output doesn't contain hint: %s", platErr.Error())
	}
}

func TestSupportsPluginCommand_NoHintsInManifest(t *testing.T) {
	t.Parallel()

	// Create temp directory for plugin
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create plugin directory with manifest WITHOUT hints
	// Directory must have "shelly-" prefix to be found by registry
	pluginDir := filepath.Join(tmpDir, "shelly-esphome")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	manifest := plugins.Manifest{
		SchemaVersion: "1",
		Name:          "esphome",
		Version:       "1.0.0",
		Capabilities: &plugins.Capabilities{
			Platform:   "esphome",
			Components: []string{"switch"},
			// No hints defined
		},
		Hooks: &plugins.Hooks{
			Control: "./shelly-esphome control",
		},
		Binary: plugins.Binary{
			Name:     "shelly-esphome",
			Checksum: "sha256:abc123",
		},
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestData, 0o600); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Create a dummy binary
	binaryPath := filepath.Join(pluginDir, "shelly-esphome")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho test"), 0o700); err != nil { //nolint:gosec // G306: test binary needs execute permission
		t.Fatalf("failed to create binary: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	service := NewService()
	service.SetPluginRegistry(registry)

	device := model.Device{
		Name:     "test-esphome",
		Address:  "192.168.1.101",
		Platform: "esphome",
	}

	// Request an unsupported command - should get error without hint
	err = service.SupportsPluginCommand(device, "scene")
	if err == nil {
		t.Fatal("expected error for unsupported command")
	}

	var platErr *PlatformError
	if !errors.As(err, &platErr) {
		t.Fatalf("expected *PlatformError, got %T", err)
	}

	if platErr.Hint != "" {
		t.Errorf("expected empty hint for plugin without hints, got %q", platErr.Hint)
	}
}

func TestSupportsPluginCommand_ShellyDevice(t *testing.T) {
	t.Parallel()

	service := NewService()

	// Shelly device (not plugin-managed)
	device := model.Device{
		Name:     "shelly-plug",
		Address:  "192.168.1.50",
		Platform: "", // Empty platform = Shelly device
	}

	// All commands should be supported for Shelly devices
	commands := []string{"on", "off", "toggle", "status", "scene", "script", "schedule", "kvs"}
	for _, cmd := range commands {
		err := service.SupportsPluginCommand(device, cmd)
		if err != nil {
			t.Errorf("SupportsPluginCommand(%q) = %v, want nil for Shelly device", cmd, err)
		}
	}
}

func TestPlatformErrorWithHint(t *testing.T) {
	t.Parallel()

	t.Run("with hint", func(t *testing.T) {
		t.Parallel()
		err := NewPlatformErrorWithHint("tasmota", "scene", "Use Tasmota Rules instead")
		wantMsg := `"scene" command is not supported for tasmota devices` + "\nHint: Use Tasmota Rules instead"
		if err.Error() != wantMsg {
			t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("without hint", func(t *testing.T) {
		t.Parallel()
		err := NewPlatformError("tasmota", "unknown")
		wantMsg := `"unknown" command is not supported for tasmota devices`
		if err.Error() != wantMsg {
			t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
		}
	})
}
