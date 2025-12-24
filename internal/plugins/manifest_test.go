package plugins

import (
	"encoding/json"
	"testing"
)

func TestManifest_BackwardCompatibility(t *testing.T) {
	t.Parallel()

	// Old manifest without capabilities/hooks fields.
	oldManifest := `{
		"schema_version": "1",
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "Test plugin",
		"installed_at": "2024-01-01T00:00:00Z",
		"source": {"type": "local"},
		"binary": {"name": "shelly-test", "checksum": "sha256:abc"}
	}`

	var m Manifest
	if err := json.Unmarshal([]byte(oldManifest), &m); err != nil {
		t.Fatalf("failed to unmarshal old manifest: %v", err)
	}

	if m.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got %q", m.Name)
	}
	if m.Capabilities != nil {
		t.Error("expected Capabilities to be nil for old manifest")
	}
	if m.Hooks != nil {
		t.Error("expected Hooks to be nil for old manifest")
	}
}

func TestManifest_WithCapabilitiesAndHooks(t *testing.T) {
	t.Parallel()

	// New manifest with capabilities and hooks.
	newManifest := `{
		"schema_version": "1",
		"name": "tasmota",
		"version": "1.0.0",
		"installed_at": "2024-01-01T00:00:00Z",
		"source": {"type": "github", "url": "https://github.com/user/repo"},
		"binary": {"name": "shelly-tasmota", "checksum": "sha256:def"},
		"capabilities": {
			"device_detection": true,
			"platform": "tasmota",
			"components": ["switch", "light", "energy"],
			"firmware_updates": true
		},
		"hooks": {
			"detect": "./shelly-tasmota detect",
			"status": "./shelly-tasmota status",
			"control": "./shelly-tasmota control",
			"check_updates": "./shelly-tasmota check-updates",
			"apply_update": "./shelly-tasmota apply-update"
		}
	}`

	var m Manifest
	if err := json.Unmarshal([]byte(newManifest), &m); err != nil {
		t.Fatalf("failed to unmarshal new manifest: %v", err)
	}

	if m.Name != testPlatformTasmota {
		t.Errorf("expected name %q, got %q", testPlatformTasmota, m.Name)
	}

	// Verify capabilities.
	if m.Capabilities == nil {
		t.Fatal("expected Capabilities to be set")
	}
	if !m.Capabilities.DeviceDetection {
		t.Error("expected DeviceDetection to be true")
	}
	if m.Capabilities.Platform != testPlatformTasmota {
		t.Errorf("expected Platform %q, got %q", testPlatformTasmota, m.Capabilities.Platform)
	}
	if len(m.Capabilities.Components) != 3 {
		t.Errorf("expected 3 components, got %d", len(m.Capabilities.Components))
	}
	if !m.Capabilities.FirmwareUpdates {
		t.Error("expected FirmwareUpdates to be true")
	}

	// Verify hooks.
	if m.Hooks == nil {
		t.Fatal("expected Hooks to be set")
	}
	if m.Hooks.Detect != "./shelly-tasmota detect" {
		t.Errorf("expected Detect hook, got %q", m.Hooks.Detect)
	}
	if m.Hooks.Status != "./shelly-tasmota status" {
		t.Errorf("expected Status hook, got %q", m.Hooks.Status)
	}
	if m.Hooks.Control != "./shelly-tasmota control" {
		t.Errorf("expected Control hook, got %q", m.Hooks.Control)
	}
	if m.Hooks.CheckUpdates != "./shelly-tasmota check-updates" {
		t.Errorf("expected CheckUpdates hook, got %q", m.Hooks.CheckUpdates)
	}
	if m.Hooks.ApplyUpdate != "./shelly-tasmota apply-update" {
		t.Errorf("expected ApplyUpdate hook, got %q", m.Hooks.ApplyUpdate)
	}
}

func TestManifest_PartialCapabilities(t *testing.T) {
	t.Parallel()

	// Manifest with partial capabilities (only detection, no firmware updates).
	partialManifest := `{
		"schema_version": "1",
		"name": "esphome",
		"version": "0.5.0",
		"installed_at": "2024-01-01T00:00:00Z",
		"source": {"type": "local"},
		"binary": {"name": "shelly-esphome", "checksum": "sha256:ghi"},
		"capabilities": {
			"device_detection": true,
			"platform": "esphome"
		}
	}`

	var m Manifest
	if err := json.Unmarshal([]byte(partialManifest), &m); err != nil {
		t.Fatalf("failed to unmarshal partial manifest: %v", err)
	}

	if m.Capabilities == nil {
		t.Fatal("expected Capabilities to be set")
	}
	if !m.Capabilities.DeviceDetection {
		t.Error("expected DeviceDetection to be true")
	}
	if m.Capabilities.Platform != "esphome" {
		t.Errorf("expected Platform 'esphome', got %q", m.Capabilities.Platform)
	}
	if m.Capabilities.FirmwareUpdates {
		t.Error("expected FirmwareUpdates to be false (not set)")
	}
	if len(m.Capabilities.Components) != 0 {
		t.Errorf("expected 0 components, got %d", len(m.Capabilities.Components))
	}
	if m.Hooks != nil {
		t.Error("expected Hooks to be nil for manifest without hooks")
	}
}
