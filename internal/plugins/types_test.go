package plugins

import (
	"encoding/json"
	"testing"
)

// TestComponentInfo tests ComponentInfo struct.
func TestComponentInfo(t *testing.T) {
	t.Parallel()

	info := ComponentInfo{
		Type: "switch",
		ID:   0,
		Name: "Main Switch",
	}

	if info.Type != "switch" {
		t.Errorf("Type = %q, want %q", info.Type, "switch")
	}
	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "Main Switch" {
		t.Errorf("Name = %q, want %q", info.Name, "Main Switch")
	}
}

// TestComponentInfo_JSONMarshaling tests ComponentInfo JSON marshaling.
func TestComponentInfo_JSONMarshaling(t *testing.T) {
	t.Parallel()

	info := ComponentInfo{
		Type: "light",
		ID:   1,
		Name: "Bedroom Light",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed ComponentInfo
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Type != info.Type {
		t.Errorf("Type = %q, want %q", parsed.Type, info.Type)
	}
	if parsed.ID != info.ID {
		t.Errorf("ID = %d, want %d", parsed.ID, info.ID)
	}
	if parsed.Name != info.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, info.Name)
	}
}

// TestDeviceDetectionResult tests DeviceDetectionResult struct.
func TestDeviceDetectionResult(t *testing.T) {
	t.Parallel()

	result := DeviceDetectionResult{
		Detected:   true,
		Platform:   "tasmota",
		DeviceID:   "tasmota-123456",
		DeviceName: "Kitchen Plug",
		Model:      "Sonoff Basic",
		Firmware:   "14.3.0",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Components: []ComponentInfo{
			{Type: "switch", ID: 0},
			{Type: "energy", ID: 0},
		},
	}

	if !result.Detected {
		t.Error("Detected = false, want true")
	}
	if result.Platform != testPlatformTasmota {
		t.Errorf("Platform = %q, want %q", result.Platform, testPlatformTasmota)
	}
	if result.DeviceID != "tasmota-123456" {
		t.Errorf("DeviceID = %q, want %q", result.DeviceID, "tasmota-123456")
	}
	if result.DeviceName != "Kitchen Plug" {
		t.Errorf("DeviceName = %q, want %q", result.DeviceName, "Kitchen Plug")
	}
	if len(result.Components) != 2 {
		t.Errorf("Components len = %d, want 2", len(result.Components))
	}
}

// TestDeviceDetectionResult_NotDetected tests DeviceDetectionResult when not detected.
func TestDeviceDetectionResult_NotDetected(t *testing.T) {
	t.Parallel()

	result := DeviceDetectionResult{
		Detected: false,
	}

	if result.Detected {
		t.Error("Detected = true, want false")
	}
	if result.Platform != "" {
		t.Errorf("Platform = %q, want empty", result.Platform)
	}
}

// TestDeviceStatusResult tests DeviceStatusResult struct.
func TestDeviceStatusResult(t *testing.T) {
	t.Parallel()

	result := DeviceStatusResult{
		Online: true,
		Components: map[string]any{
			"switch:0": map[string]any{"output": true},
			"light:0":  map[string]any{"brightness": 75},
		},
		Sensors: map[string]any{
			"temperature": 22.5,
			"humidity":    55.0,
		},
		Energy: &EnergyStatus{
			Power:       50.5,
			Voltage:     230.0,
			Current:     0.22,
			Total:       1250.5,
			PowerFactor: 0.95,
		},
	}

	if !result.Online {
		t.Error("Online = false, want true")
	}
	if len(result.Components) != 2 {
		t.Errorf("Components len = %d, want 2", len(result.Components))
	}
	if len(result.Sensors) != 2 {
		t.Errorf("Sensors len = %d, want 2", len(result.Sensors))
	}
	if result.Energy == nil {
		t.Fatal("Energy should not be nil")
	}
	if result.Energy.Power != 50.5 {
		t.Errorf("Energy.Power = %f, want 50.5", result.Energy.Power)
	}
}

// TestDeviceStatusResult_Offline tests DeviceStatusResult when offline.
func TestDeviceStatusResult_Offline(t *testing.T) {
	t.Parallel()

	result := DeviceStatusResult{
		Online: false,
	}

	if result.Online {
		t.Error("Online = true, want false")
	}
	if result.Components != nil {
		t.Error("Components should be nil for offline device")
	}
}

// TestEnergyStatus tests EnergyStatus struct.
func TestEnergyStatus(t *testing.T) {
	t.Parallel()

	status := EnergyStatus{
		Power:         100.5,
		Voltage:       230.0,
		Current:       0.44,
		Total:         5000.25,
		ApparentPower: 105.0,
		ReactivePower: 25.0,
		PowerFactor:   0.96,
	}

	if status.Power != 100.5 {
		t.Errorf("Power = %f, want 100.5", status.Power)
	}
	if status.Voltage != 230.0 {
		t.Errorf("Voltage = %f, want 230.0", status.Voltage)
	}
	if status.Current != 0.44 {
		t.Errorf("Current = %f, want 0.44", status.Current)
	}
	if status.Total != 5000.25 {
		t.Errorf("Total = %f, want 5000.25", status.Total)
	}
	if status.ApparentPower != 105.0 {
		t.Errorf("ApparentPower = %f, want 105.0", status.ApparentPower)
	}
	if status.ReactivePower != 25.0 {
		t.Errorf("ReactivePower = %f, want 25.0", status.ReactivePower)
	}
	if status.PowerFactor != 0.96 {
		t.Errorf("PowerFactor = %f, want 0.96", status.PowerFactor)
	}
}

// TestControlResult tests ControlResult struct.
func TestControlResult(t *testing.T) {
	t.Parallel()

	// Success case
	result := ControlResult{
		Success: true,
		State:   "on",
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}
	if result.State != "on" {
		t.Errorf("State = %q, want %q", result.State, "on")
	}
	if result.Error != "" {
		t.Errorf("Error = %q, want empty", result.Error)
	}
}

// TestControlResult_Failure tests ControlResult with failure.
func TestControlResult_Failure(t *testing.T) {
	t.Parallel()

	result := ControlResult{
		Success: false,
		Error:   "device unreachable",
	}

	if result.Success {
		t.Error("Success = true, want false")
	}
	if result.Error != "device unreachable" {
		t.Errorf("Error = %q, want %q", result.Error, "device unreachable")
	}
}

// TestFirmwareUpdateInfo tests FirmwareUpdateInfo struct.
func TestFirmwareUpdateInfo(t *testing.T) {
	t.Parallel()

	info := FirmwareUpdateInfo{
		CurrentVersion:  "14.2.0",
		LatestStable:    "14.3.0",
		LatestBeta:      "15.0.0-beta1",
		HasUpdate:       true,
		HasBetaUpdate:   true,
		OTAURLStable:    "http://ota.example.com/firmware/14.3.0.bin",
		OTAURLBeta:      "http://ota.example.com/firmware/15.0.0-beta1.bin",
		ChipType:        "ESP32",
		Variant:         "tasmota",
		ReleaseNotesURL: "http://example.com/release-notes",
	}

	if info.CurrentVersion != "14.2.0" {
		t.Errorf("CurrentVersion = %q, want %q", info.CurrentVersion, "14.2.0")
	}
	if info.LatestStable != "14.3.0" {
		t.Errorf("LatestStable = %q, want %q", info.LatestStable, "14.3.0")
	}
	if info.LatestBeta != "15.0.0-beta1" {
		t.Errorf("LatestBeta = %q, want %q", info.LatestBeta, "15.0.0-beta1")
	}
	if !info.HasUpdate {
		t.Error("HasUpdate = false, want true")
	}
	if !info.HasBetaUpdate {
		t.Error("HasBetaUpdate = false, want true")
	}
	if info.ChipType != "ESP32" {
		t.Errorf("ChipType = %q, want %q", info.ChipType, "ESP32")
	}
	if info.Variant != "tasmota" {
		t.Errorf("Variant = %q, want %q", info.Variant, "tasmota")
	}
}

// TestFirmwareUpdateInfo_NoUpdate tests FirmwareUpdateInfo when no update available.
func TestFirmwareUpdateInfo_NoUpdate(t *testing.T) {
	t.Parallel()

	info := FirmwareUpdateInfo{
		CurrentVersion: "14.3.0",
		LatestStable:   "14.3.0",
		HasUpdate:      false,
		HasBetaUpdate:  false,
	}

	if info.HasUpdate {
		t.Error("HasUpdate = true, want false")
	}
	if info.HasBetaUpdate {
		t.Error("HasBetaUpdate = true, want false")
	}
}

// TestUpdateResult tests UpdateResult struct.
func TestUpdateResult(t *testing.T) {
	t.Parallel()

	// Success case with reboot
	result := UpdateResult{
		Success:   true,
		Message:   "Update initiated",
		Rebooting: true,
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}
	if result.Message != "Update initiated" {
		t.Errorf("Message = %q, want %q", result.Message, "Update initiated")
	}
	if !result.Rebooting {
		t.Error("Rebooting = false, want true")
	}
	if result.Error != "" {
		t.Errorf("Error = %q, want empty", result.Error)
	}
}

// TestUpdateResult_Failure tests UpdateResult with failure.
func TestUpdateResult_Failure(t *testing.T) {
	t.Parallel()

	result := UpdateResult{
		Success: false,
		Error:   "insufficient space",
	}

	if result.Success {
		t.Error("Success = true, want false")
	}
	if result.Error != "insufficient space" {
		t.Errorf("Error = %q, want %q", result.Error, "insufficient space")
	}
	if result.Rebooting {
		t.Error("Rebooting = true, want false")
	}
}

// TestPlugin tests Plugin struct.
func TestPlugin(t *testing.T) {
	t.Parallel()

	plugin := Plugin{
		Name:    "tasmota",
		Path:    "/home/user/.config/shelly/plugins/shelly-tasmota/shelly-tasmota",
		Version: "1.0.0",
		Dir:     "/home/user/.config/shelly/plugins/shelly-tasmota",
		Manifest: &Manifest{
			Name:    "tasmota",
			Version: "1.0.0",
		},
	}

	if plugin.Name != "tasmota" {
		t.Errorf("Name = %q, want %q", plugin.Name, "tasmota")
	}
	if plugin.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", plugin.Version, "1.0.0")
	}
	if plugin.Manifest == nil {
		t.Fatal("Manifest should not be nil")
	}
}
