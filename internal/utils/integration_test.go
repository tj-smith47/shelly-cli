// Package utils provides common functionality shared across CLI commands.
//
//go:build integration
// +build integration

package utils

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// Integration tests that require config state.
// Run with: go test -tags=integration ./internal/utils/...

func setupTestConfig(t *testing.T) {
	t.Helper()

	// Reset the config singleton BEFORE changing HOME
	// This ensures the next call to getDefaultManager will use the new HOME
	config.ResetDefaultManagerForTesting()

	// Create temp directory for test config
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "shelly")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write minimal config
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `
devices: {}
groups: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Reset singleton on cleanup
	t.Cleanup(func() {
		config.ResetDefaultManagerForTesting()
	})
}

func TestRegisterDiscoveredDevices_Integration(t *testing.T) {
	setupTestConfig(t)

	// Create test devices
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "test-device-1",
			Address:    net.ParseIP("192.168.1.100"),
			Generation: 2,
			Model:      "SHSW-1",
		},
		{
			ID:         "test-device-2",
			Address:    net.ParseIP("192.168.1.101"),
			Generation: 2,
			Model:      "SHSW-PM",
		},
	}

	added, err := RegisterDiscoveredDevices(devices, false)
	if err != nil {
		t.Fatalf("RegisterDiscoveredDevices() error = %v", err)
	}

	if added != 2 {
		t.Errorf("added = %d, want 2", added)
	}
}

func TestRegisterDiscoveredDevices_SkipExisting(t *testing.T) {
	setupTestConfig(t)

	// Register first device
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "skip-test-device",
			Address:    net.ParseIP("192.168.1.200"),
			Generation: 2,
			Model:      "SHSW-1",
		},
	}

	added1, err := RegisterDiscoveredDevices(devices, false)
	if err != nil {
		t.Fatalf("RegisterDiscoveredDevices() error = %v", err)
	}
	if added1 != 1 {
		t.Errorf("first registration: added = %d, want 1", added1)
	}

	// Try to register again with skipExisting=true
	added2, err := RegisterDiscoveredDevices(devices, true)
	if err != nil {
		t.Fatalf("RegisterDiscoveredDevices() error = %v", err)
	}
	if added2 != 0 {
		t.Errorf("second registration with skip: added = %d, want 0", added2)
	}
}

func TestRegisterDevicesFromFlags_Integration(t *testing.T) {
	setupTestConfig(t)

	// Test with device specs
	deviceSpecs := []string{
		"flag-test-1=192.168.1.50",
		"flag-test-2=192.168.1.51:admin:password",
	}

	registered, err := RegisterDevicesFromFlags(deviceSpecs, nil)
	if err != nil {
		t.Fatalf("RegisterDevicesFromFlags() error = %v", err)
	}

	if registered != 2 {
		t.Errorf("registered = %d, want 2", registered)
	}
}

func TestRegisterDevicesFromFlags_JSON(t *testing.T) {
	setupTestConfig(t)

	// Test with JSON input
	devicesJSON := []string{
		`[{"name":"json-dev-1","address":"192.168.1.60"},{"name":"json-dev-2","address":"192.168.1.61"}]`,
	}

	registered, err := RegisterDevicesFromFlags(nil, devicesJSON)
	if err != nil {
		t.Fatalf("RegisterDevicesFromFlags() error = %v", err)
	}

	if registered != 2 {
		t.Errorf("registered = %d, want 2", registered)
	}
}

func TestRegisterDevicesFromFlags_InvalidSpec(t *testing.T) {
	setupTestConfig(t)

	// Test with invalid device spec
	deviceSpecs := []string{"invalid-no-equals"}

	_, err := RegisterDevicesFromFlags(deviceSpecs, nil)
	if err == nil {
		t.Error("RegisterDevicesFromFlags() should error for invalid spec")
	}
}

func TestRegisterDevicesFromFlags_InvalidJSON(t *testing.T) {
	setupTestConfig(t)

	// Test with invalid JSON
	devicesJSON := []string{`{invalid-json`}

	_, err := RegisterDevicesFromFlags(nil, devicesJSON)
	if err == nil {
		t.Error("RegisterDevicesFromFlags() should error for invalid JSON")
	}
}

func TestRegisterDevicesFromFlags_MissingFields(t *testing.T) {
	setupTestConfig(t)

	// Test with missing required fields
	devicesJSON := []string{
		`{"address":"192.168.1.70"}`, // missing name
	}

	_, err := RegisterDevicesFromFlags(nil, devicesJSON)
	if err == nil {
		t.Error("RegisterDevicesFromFlags() should error for missing name")
	}

	devicesJSON = []string{
		`{"name":"no-address"}`, // missing address
	}

	_, err = RegisterDevicesFromFlags(nil, devicesJSON)
	if err == nil {
		t.Error("RegisterDevicesFromFlags() should error for missing address")
	}
}

func TestRegisterPluginDiscoveredDevice_Integration(t *testing.T) {
	setupTestConfig(t)

	result := &plugins.DeviceDetectionResult{
		Detected:   true,
		Platform:   "test-platform",
		DeviceID:   "plugin-test-123",
		DeviceName: "Plugin Test Device",
		Model:      "TEST-MODEL",
	}

	added, err := RegisterPluginDiscoveredDevice(result, "192.168.1.80", false)
	if err != nil {
		t.Fatalf("RegisterPluginDiscoveredDevice() error = %v", err)
	}
	if !added {
		t.Error("RegisterPluginDiscoveredDevice() should return true for new device")
	}
}

func TestRegisterPluginDiscoveredDevice_SkipExisting(t *testing.T) {
	setupTestConfig(t)

	result := &plugins.DeviceDetectionResult{
		Detected:   true,
		Platform:   "test-platform",
		DeviceID:   "plugin-skip-test",
		DeviceName: "Skip Test",
		Model:      "TEST",
	}

	// First registration
	added1, err := RegisterPluginDiscoveredDevice(result, "192.168.1.81", false)
	if err != nil {
		t.Fatalf("first RegisterPluginDiscoveredDevice() error = %v", err)
	}
	if !added1 {
		t.Error("first registration should succeed")
	}

	// Second registration with skipExisting
	added2, err := RegisterPluginDiscoveredDevice(result, "192.168.1.81", true)
	if err != nil {
		t.Fatalf("second RegisterPluginDiscoveredDevice() error = %v", err)
	}
	if added2 {
		t.Error("second registration with skip should return false")
	}
}

func TestRegisterPluginDiscoveredDevices_Batch(t *testing.T) {
	setupTestConfig(t)

	devices := []PluginDevice{
		{
			Address:  "192.168.1.90",
			Platform: "batch-test",
			ID:       "batch-1",
			Name:     "Batch Device 1",
			Model:    "BATCH",
		},
		{
			Address:  "192.168.1.91",
			Platform: "batch-test",
			ID:       "batch-2",
			Name:     "Batch Device 2",
			Model:    "BATCH",
		},
		{
			Address:  "", // Empty address should be skipped
			Platform: "batch-test",
			ID:       "batch-skip",
		},
	}

	added := RegisterPluginDiscoveredDevices(devices, false)
	if added != 2 {
		t.Errorf("RegisterPluginDiscoveredDevices() = %d, want 2", added)
	}
}

func TestIsPluginDeviceRegistered_Integration(t *testing.T) {
	setupTestConfig(t)

	// Register a device
	_ = config.RegisterDevice("registered-device", "192.168.1.95", 2, "", "", nil)

	// Check if registered
	if !IsPluginDeviceRegistered("192.168.1.95") {
		t.Error("IsPluginDeviceRegistered() should return true for registered address")
	}

	// Check unregistered address
	if IsPluginDeviceRegistered("192.168.1.99") {
		t.Error("IsPluginDeviceRegistered() should return false for unregistered address")
	}
}

func TestResolveBatchTargets_All(t *testing.T) {
	setupTestConfig(t)

	// Register some devices
	_ = config.RegisterDevice("all-test-1", "192.168.1.110", 2, "", "", nil)
	_ = config.RegisterDevice("all-test-2", "192.168.1.111", 2, "", "", nil)

	targets, err := ResolveBatchTargets("", true, nil)
	if err != nil {
		t.Fatalf("ResolveBatchTargets(all=true) error = %v", err)
	}

	if len(targets) < 2 {
		t.Errorf("len(targets) = %d, want >= 2", len(targets))
	}
}

func TestResolveBatchTargets_Group(t *testing.T) {
	setupTestConfig(t)

	// Register devices
	_ = config.RegisterDevice("group-dev-1", "192.168.1.120", 2, "", "", nil)
	_ = config.RegisterDevice("group-dev-2", "192.168.1.121", 2, "", "", nil)

	// Create a group and add devices
	_ = config.CreateGroup("test-group")
	_ = config.AddDeviceToGroup("test-group", "group-dev-1")
	_ = config.AddDeviceToGroup("test-group", "group-dev-2")

	targets, err := ResolveBatchTargets("test-group", false, nil)
	if err != nil {
		t.Fatalf("ResolveBatchTargets(group) error = %v", err)
	}

	if len(targets) != 2 {
		t.Errorf("len(targets) = %d, want 2", len(targets))
	}
}

func TestResolveBatchTargets_EmptyGroup(t *testing.T) {
	setupTestConfig(t)

	// Create an empty group
	_ = config.CreateGroup("empty-group")

	_, err := ResolveBatchTargets("empty-group", false, nil)
	if err == nil {
		t.Error("ResolveBatchTargets() should error for empty group")
	}
}

func TestResolveBatchTargets_NonexistentGroup(t *testing.T) {
	setupTestConfig(t)

	_, err := ResolveBatchTargets("nonexistent-group", false, nil)
	if err == nil {
		t.Error("ResolveBatchTargets() should error for nonexistent group")
	}
}
