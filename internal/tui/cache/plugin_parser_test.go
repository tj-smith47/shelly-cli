package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const (
	testInputTypeButton = "button"
	testPlatformName    = "tasmota"
	testPluginName      = "shelly-tasmota"
)

// setupTestPlugin creates a temp directory with a mock plugin that has a status hook
// returning the given JSON. Returns the shelly.Service with the plugin registry configured.
func setupTestPlugin(t *testing.T, statusJSON string) *shelly.Service {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "shelly-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create plugin directory (must have "shelly-" prefix)
	pluginDir := filepath.Join(tmpDir, testPluginName)
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create status hook script that outputs the given JSON
	hookScript := "#!/bin/sh\necho '" + statusJSON + "'\n"
	testutil.WriteTestScript(t, filepath.Join(pluginDir, "status"), hookScript)

	// Create manifest
	manifest := plugins.Manifest{
		SchemaVersion: "1",
		Name:          testPlatformName,
		Version:       "1.0.0",
		Capabilities: &plugins.Capabilities{
			Platform:   testPlatformName,
			Components: []string{"switch", "light"},
		},
		Hooks: &plugins.Hooks{
			Status: "./status",
		},
		Binary: plugins.Binary{
			Name: testPluginName,
		},
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestData, 0o600); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Create dummy binary so plugin is considered valid
	testutil.WriteTestScript(t, filepath.Join(pluginDir, testPluginName), "#!/bin/sh\necho test\n")

	// Create service with plugin registry
	registry := plugins.NewRegistryWithDir(tmpDir)
	svc := shelly.New(shelly.NewConfigResolver(), shelly.WithPluginRegistry(registry))

	return svc
}

func TestParsePluginStatus_Nil(t *testing.T) {
	t.Parallel()

	result := ParsePluginStatus("test", nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestParsePluginStatus_SwitchComponent(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"switch:0": map[string]any{
				"output": true,
				"name":   "Main Switch",
				"source": testInputTypeButton,
				"apower": 45.2,
			},
			"switch:1": map[string]any{
				"output": false,
				"name":   "Secondary",
			},
		},
	}

	parsed := ParsePluginStatus("test-device", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if len(parsed.Switches) != 2 {
		t.Fatalf("expected 2 switches, got %d", len(parsed.Switches))
	}

	// Check power tracking
	if power, ok := parsed.SwitchPowers[0]; !ok || power != 45.2 {
		t.Errorf("expected switch:0 power=45.2, got %v (ok=%v)", power, ok)
	}
}

func TestParsePluginStatus_LightComponent(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"light:0": map[string]any{
				"output": true,
				"name":   "Dimmer",
			},
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if len(parsed.Lights) != 1 {
		t.Fatalf("expected 1 light, got %d", len(parsed.Lights))
	}

	if !parsed.Lights[0].On {
		t.Error("expected light to be on")
	}
	if parsed.Lights[0].Name != "Dimmer" {
		t.Errorf("expected name 'Dimmer', got %q", parsed.Lights[0].Name)
	}
}

func TestParsePluginStatus_CoverComponent(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"cover:0": map[string]any{
				"state": "open",
				"name":  "Blinds",
			},
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if len(parsed.Covers) != 1 {
		t.Fatalf("expected 1 cover, got %d", len(parsed.Covers))
	}
	if parsed.Covers[0].State != CoverStateOpen {
		t.Errorf("expected state %q, got %q", CoverStateOpen, parsed.Covers[0].State)
	}
}

func TestParsePluginStatus_InputComponent(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"input:0": map[string]any{
				"state": true,
				"type":  testInputTypeButton,
				"name":  "Wall Button",
			},
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if len(parsed.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(parsed.Inputs))
	}
	if !parsed.Inputs[0].State {
		t.Error("expected input state to be true")
	}
	if parsed.Inputs[0].Type != testInputTypeButton {
		t.Errorf("expected type 'button', got %q", parsed.Inputs[0].Type)
	}
}

func TestParsePluginStatus_Energy(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Energy: &plugins.EnergyStatus{
			Power:   100.5,
			Voltage: 230.0,
			Current: 0.437,
			Total:   1234.56,
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if parsed.Power != 100.5 {
		t.Errorf("expected power=100.5, got %f", parsed.Power)
	}
	if parsed.Voltage != 230.0 {
		t.Errorf("expected voltage=230.0, got %f", parsed.Voltage)
	}
	if parsed.Current != 0.437 {
		t.Errorf("expected current=0.437, got %f", parsed.Current)
	}
	if parsed.TotalEnergy != 1234.56 {
		t.Errorf("expected totalEnergy=1234.56, got %f", parsed.TotalEnergy)
	}
}

func TestParsePluginStatus_Sensors(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Sensors: map[string]any{
			"temperature": 23.5,
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if parsed.Temperature != 23.5 {
		t.Errorf("expected temperature=23.5, got %f", parsed.Temperature)
	}
}

func TestParsePluginStatus_StringState(t *testing.T) {
	t.Parallel()

	// Test plugins that use "state": "on" instead of "output": true
	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"switch:0": map[string]any{
				"state": "on",
			},
			"light:0": map[string]any{
				"state": "ON",
			},
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if len(parsed.Switches) != 1 || !parsed.Switches[0].On {
		t.Error("expected switch to be on via string state")
	}
	if len(parsed.Lights) != 1 || !parsed.Lights[0].On {
		t.Error("expected light to be on via string state")
	}
}

func TestParsePluginStatus_InvalidComponentKey(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"invalid":    map[string]any{"output": true},
			"switch:abc": map[string]any{"output": true},
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	// Invalid keys should be skipped
	if len(parsed.Switches) != 0 {
		t.Errorf("expected 0 switches for invalid keys, got %d", len(parsed.Switches))
	}
}

func TestParsePluginStatus_NonMapComponent(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"switch:0": "not-a-map",
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if len(parsed.Switches) != 0 {
		t.Errorf("expected 0 switches for non-map value, got %d", len(parsed.Switches))
	}
}

func TestParsePluginStatus_MixedComponents(t *testing.T) {
	t.Parallel()

	result := &plugins.DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"switch:0": map[string]any{"output": true, "name": "Relay 1"},
			"switch:1": map[string]any{"output": false, "name": "Relay 2"},
			"light:0":  map[string]any{"output": true},
			"cover:0":  map[string]any{"state": "closed"},
			"input:0":  map[string]any{"state": true, "type": "switch"},
		},
		Energy: &plugins.EnergyStatus{Power: 50.0},
		Sensors: map[string]any{
			"temperature": 25.0,
		},
	}

	parsed := ParsePluginStatus("test", result)
	if parsed == nil {
		t.Fatal("expected non-nil parsed result")
	}

	if len(parsed.Switches) != 2 {
		t.Errorf("expected 2 switches, got %d", len(parsed.Switches))
	}
	if len(parsed.Lights) != 1 {
		t.Errorf("expected 1 light, got %d", len(parsed.Lights))
	}
	if len(parsed.Covers) != 1 {
		t.Errorf("expected 1 cover, got %d", len(parsed.Covers))
	}
	if len(parsed.Inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(parsed.Inputs))
	}
	if parsed.Power != 50.0 {
		t.Errorf("expected power=50.0, got %f", parsed.Power)
	}
	if parsed.Temperature != 25.0 {
		t.Errorf("expected temperature=25.0, got %f", parsed.Temperature)
	}
}

func TestBuildPluginDeviceInfo(t *testing.T) {
	t.Parallel()

	device := model.Device{
		Name:       "tasmota1",
		Address:    "10.0.0.5",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Model:      "Sonoff Basic",
		Platform:   "tasmota",
		Generation: 0,
	}

	info := BuildPluginDeviceInfo("tasmota1", device)
	if info.ID != "tasmota1" {
		t.Errorf("expected ID 'tasmota1', got %q", info.ID)
	}
	if info.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC 'AA:BB:CC:DD:EE:FF', got %q", info.MAC)
	}
	if info.Model != "Sonoff Basic" {
		t.Errorf("expected Model 'Sonoff Basic', got %q", info.Model)
	}
	if info.App != "tasmota" {
		t.Errorf("expected App 'tasmota', got %q", info.App)
	}
	if info.Firmware != "tasmota (plugin)" {
		t.Errorf("expected Firmware 'tasmota (plugin)', got %q", info.Firmware)
	}
}

func TestPluginParseComponentKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key      string
		wantType string
		wantID   int
		wantOK   bool
	}{
		{"switch:0", "switch", 0, true},
		{"light:1", "light", 1, true},
		{"cover:2", "cover", 2, true},
		{"input:0", "input", 0, true},
		{"invalid", "", 0, false},
		{"switch:abc", "", 0, false},
		{"", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			compType, compID, ok := pluginParseComponentKey(tt.key)
			if ok != tt.wantOK {
				t.Errorf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if compType != tt.wantType {
				t.Errorf("expected type=%q, got %q", tt.wantType, compType)
			}
			if compID != tt.wantID {
				t.Errorf("expected id=%d, got %d", tt.wantID, compID)
			}
		})
	}
}

// --- Cache Branch Tests ---
// These test fetchPluginDevice through real plugin hook execution.

func TestFetchPluginDevice_OnlineWithComponents(t *testing.T) {
	t.Parallel()

	statusJSON := `{"online":true,"components":{"switch:0":{"output":true,"name":"Relay","apower":42.5}},"energy":{"power":42.5,"voltage":230.0},"sensors":{"temperature":24.0}}`

	svc := setupTestPlugin(t, statusJSON)
	ios := iostreams.Test(nil, os.Stdout, os.Stderr)

	c := &Cache{
		ctx:                context.Background(),
		svc:                svc,
		ios:                ios,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		esManaged:          make(map[string]bool),
		pendingRefreshes:   make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		waveConfig:         DefaultWaveConfig(),
	}

	device := model.Device{
		Name:     "test-tasmota",
		Address:  "192.168.1.100",
		Platform: testPlatformName,
	}
	data := &DeviceData{
		Device:    device,
		Fetched:   true,
		UpdatedAt: time.Now(),
	}

	msg := c.fetchPluginDevice("test-tasmota", device, data, 1)
	updateMsg, ok := msg.(DeviceUpdateMsg)
	if !ok {
		t.Fatalf("expected DeviceUpdateMsg, got %T", msg)
	}

	if updateMsg.Name != "test-tasmota" {
		t.Errorf("expected name 'test-tasmota', got %q", updateMsg.Name)
	}
	if updateMsg.RequestID != 1 {
		t.Errorf("expected requestID=1, got %d", updateMsg.RequestID)
	}
	if !updateMsg.Data.Online {
		t.Error("expected device to be online")
	}
	if updateMsg.Data.Error != nil {
		t.Errorf("expected no error, got %v", updateMsg.Data.Error)
	}

	// Verify DeviceInfo was constructed
	if updateMsg.Data.Info == nil {
		t.Fatal("expected DeviceInfo to be set")
	}
	if updateMsg.Data.Info.ID != "test-tasmota" {
		t.Errorf("expected info ID 'test-tasmota', got %q", updateMsg.Data.Info.ID)
	}
	if updateMsg.Data.Info.App != testPlatformName {
		t.Errorf("expected info App %q, got %q", testPlatformName, updateMsg.Data.Info.App)
	}

	// Verify parsed status was applied
	if len(updateMsg.Data.Switches) != 1 {
		t.Fatalf("expected 1 switch, got %d", len(updateMsg.Data.Switches))
	}
	if !updateMsg.Data.Switches[0].On {
		t.Error("expected switch to be on")
	}
	if updateMsg.Data.Switches[0].Name != "Relay" {
		t.Errorf("expected switch name 'Relay', got %q", updateMsg.Data.Switches[0].Name)
	}
	if updateMsg.Data.Power != 42.5 {
		t.Errorf("expected power=42.5, got %f", updateMsg.Data.Power)
	}
	if updateMsg.Data.Voltage != 230.0 {
		t.Errorf("expected voltage=230.0, got %f", updateMsg.Data.Voltage)
	}
	if updateMsg.Data.Temperature != 24.0 {
		t.Errorf("expected temperature=24.0, got %f", updateMsg.Data.Temperature)
	}
}

func TestFetchPluginDevice_Offline(t *testing.T) {
	t.Parallel()

	statusJSON := `{"online":false}`

	svc := setupTestPlugin(t, statusJSON)
	ios := iostreams.Test(nil, os.Stdout, os.Stderr)

	c := &Cache{
		ctx:                context.Background(),
		svc:                svc,
		ios:                ios,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		esManaged:          make(map[string]bool),
		pendingRefreshes:   make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		waveConfig:         DefaultWaveConfig(),
	}

	device := model.Device{
		Name:     "offline-dev",
		Address:  "192.168.1.200",
		Platform: testPlatformName,
	}
	data := &DeviceData{
		Device:    device,
		Fetched:   true,
		UpdatedAt: time.Now(),
	}

	msg := c.fetchPluginDevice("offline-dev", device, data, 2)
	updateMsg, ok := msg.(DeviceUpdateMsg)
	if !ok {
		t.Fatalf("expected DeviceUpdateMsg, got %T", msg)
	}

	if updateMsg.Data.Online {
		t.Error("expected device to be offline")
	}
	if updateMsg.Data.Error != nil {
		t.Errorf("expected no error for offline device, got %v", updateMsg.Data.Error)
	}
	// DeviceInfo should still be set even for offline devices
	if updateMsg.Data.Info == nil {
		t.Fatal("expected DeviceInfo to be set even for offline device")
	}
}

func TestFetchPluginDevice_HookError(t *testing.T) {
	t.Parallel()

	// Create a plugin whose status hook exits with error
	tmpDir, err := os.MkdirTemp("", "shelly-cache-err-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	pluginDir := filepath.Join(tmpDir, testPluginName)
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Status hook that exits with error
	testutil.WriteTestScript(t, filepath.Join(pluginDir, "status"), "#!/bin/sh\necho 'device unreachable' >&2\nexit 1\n")

	manifest := plugins.Manifest{
		SchemaVersion: "1",
		Name:          testPlatformName,
		Version:       "1.0.0",
		Capabilities:  &plugins.Capabilities{Platform: testPlatformName},
		Hooks:         &plugins.Hooks{Status: "./status"},
		Binary:        plugins.Binary{Name: testPluginName},
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestData, 0o600); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	testutil.WriteTestScript(t, filepath.Join(pluginDir, testPluginName), "#!/bin/sh\necho test\n")

	registry := plugins.NewRegistryWithDir(tmpDir)
	svc := shelly.New(shelly.NewConfigResolver(), shelly.WithPluginRegistry(registry))
	ios := iostreams.Test(nil, os.Stdout, os.Stderr)

	c := &Cache{
		ctx:                context.Background(),
		svc:                svc,
		ios:                ios,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		esManaged:          make(map[string]bool),
		pendingRefreshes:   make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		waveConfig:         DefaultWaveConfig(),
	}

	device := model.Device{
		Name:     "error-dev",
		Address:  "192.168.1.250",
		Platform: testPlatformName,
	}
	data := &DeviceData{
		Device:    device,
		Fetched:   true,
		UpdatedAt: time.Now(),
	}

	msg := c.fetchPluginDevice("error-dev", device, data, 3)
	updateMsg, ok := msg.(DeviceUpdateMsg)
	if !ok {
		t.Fatalf("expected DeviceUpdateMsg, got %T", msg)
	}

	if updateMsg.Data.Online {
		t.Error("expected device to be offline after hook error")
	}
	if updateMsg.Data.Error == nil {
		t.Error("expected error to be set after hook failure")
	}
	// DeviceInfo should still be set even on error
	if updateMsg.Data.Info == nil {
		t.Fatal("expected DeviceInfo to be set even on error")
	}
}

func TestFetchPluginDevice_NoPluginRegistry(t *testing.T) {
	t.Parallel()

	// Service without plugin registry
	svc := shelly.New(shelly.NewConfigResolver())
	ios := iostreams.Test(nil, os.Stdout, os.Stderr)

	c := &Cache{
		ctx:                context.Background(),
		svc:                svc,
		ios:                ios,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		esManaged:          make(map[string]bool),
		pendingRefreshes:   make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		waveConfig:         DefaultWaveConfig(),
	}

	device := model.Device{
		Name:     "no-registry",
		Address:  "192.168.1.50",
		Platform: testPlatformName,
	}
	data := &DeviceData{
		Device:    device,
		Fetched:   true,
		UpdatedAt: time.Now(),
	}

	msg := c.fetchPluginDevice("no-registry", device, data, 4)
	updateMsg, ok := msg.(DeviceUpdateMsg)
	if !ok {
		t.Fatalf("expected DeviceUpdateMsg, got %T", msg)
	}

	if updateMsg.Data.Online {
		t.Error("expected device to be offline when no registry")
	}
	if updateMsg.Data.Error == nil {
		t.Error("expected error when no plugin registry")
	}
}
