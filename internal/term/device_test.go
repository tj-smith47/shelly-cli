package term

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestDisplayDeviceStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &shelly.DeviceStatus{
		Info: &shelly.DeviceInfo{
			ID:         "shellypro1pm-123456",
			Model:      "SNSW-001P16EU",
			Generation: 2,
			Firmware:   "1.0.0",
		},
		Status: map[string]any{
			"switch:0": map[string]any{"output": true},
		},
	}

	DisplayDeviceStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "shellypro1pm-123456") {
		t.Error("output should contain device ID")
	}
	if !strings.Contains(output, "Gen2") {
		t.Error("output should contain generation")
	}
	if !strings.Contains(output, "switch:0") {
		t.Error("output should contain component")
	}
}

func TestDisplayAllSnapshots(t *testing.T) {
	t.Parallel()

	t.Run("with online devices", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		snapshots := map[string]*shelly.DeviceSnapshot{
			"device1": {
				Info: &shelly.DeviceInfo{
					ID:    "device1",
					Model: "Shelly Pro 1PM",
				},
				Snapshot: &shelly.MonitoringSnapshot{
					Timestamp: time.Now(),
					PM: []shelly.PMStatus{
						{ID: 0, APower: 100.0},
					},
				},
				Error: nil,
			},
		}

		DisplayAllSnapshots(ios, snapshots)

		output := out.String()
		if !strings.Contains(output, "Device Status Summary") {
			t.Error("output should contain 'Device Status Summary'")
		}
		if !strings.Contains(output, "device1") {
			t.Error("output should contain 'device1'")
		}
	})

	t.Run("with offline device", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		snapshots := map[string]*shelly.DeviceSnapshot{
			"device1": {
				Info:     nil,
				Snapshot: nil,
				Error:    errors.New("connection timeout"),
			},
		}

		DisplayAllSnapshots(ios, snapshots)

		output := out.String()
		if !strings.Contains(output, "connection timeout") {
			t.Error("output should contain error message")
		}
	})
}

func TestDisplayAuthStatus(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &shelly.AuthStatus{
			Enabled: true,
		}

		DisplayAuthStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "Authentication Status") {
			t.Error("output should contain 'Authentication Status'")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &shelly.AuthStatus{
			Enabled: false,
		}

		DisplayAuthStatus(ios, status)

		output := out.String()
		if output == "" {
			t.Error("DisplayAuthStatus should produce output")
		}
	})
}

func TestDisplayPluginDeviceStatus(t *testing.T) {
	t.Parallel()

	t.Run("full status", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		device := model.Device{
			Name:    "smartthings-device",
			Address: "192.168.1.100",
			Model:   "ST-SWITCH-01",
		}
		status := &plugins.DeviceStatusResult{
			Online: true,
			Components: map[string]any{
				"switch": true,
			},
			Sensors: map[string]any{
				"temperature": 22.5,
			},
			Energy: &plugins.EnergyStatus{
				Power:   100.0,
				Voltage: 220.0,
				Current: 0.45,
			},
		}

		DisplayPluginDeviceStatus(ios, device, status)

		output := out.String()
		if !strings.Contains(output, "smartthings-device") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "Components") {
			t.Error("output should contain 'Components'")
		}
		if !strings.Contains(output, "Sensors") {
			t.Error("output should contain 'Sensors'")
		}
		if !strings.Contains(output, "Energy") {
			t.Error("output should contain 'Energy'")
		}
	})

	t.Run("minimal status", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		device := model.Device{
			Name:    "device1",
			Address: "192.168.1.100",
		}
		status := &plugins.DeviceStatusResult{
			Online: false,
		}

		DisplayPluginDeviceStatus(ios, device, status)

		output := out.String()
		if output == "" {
			t.Error("DisplayPluginDeviceStatus should produce output")
		}
	})
}

func TestQuickDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := QuickDeviceInfo{
		Model:      "Shelly Pro 1PM",
		Generation: 2,
		Firmware:   "1.0.0",
	}

	if info.Model != "Shelly Pro 1PM" {
		t.Errorf("Model = %q, want 'Shelly Pro 1PM'", info.Model)
	}
	if info.Generation != 2 {
		t.Errorf("Generation = %d, want 2", info.Generation)
	}
	if info.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want '1.0.0'", info.Firmware)
	}
}

func TestComponentState_Fields(t *testing.T) {
	t.Parallel()

	state := ComponentState{
		Name:  "switch:0",
		State: "ON",
	}

	if state.Name != "switch:0" {
		t.Errorf("Name = %q, want 'switch:0'", state.Name)
	}
	if state.State != "ON" {
		t.Errorf("State = %q, want 'ON'", state.State)
	}
}

func TestQuickDeviceStatus_Fields(t *testing.T) {
	t.Parallel()

	status := QuickDeviceStatus{
		Name:   "device1",
		Model:  "Shelly Pro 1PM",
		Online: true,
	}

	if status.Name != "device1" {
		t.Errorf("Name = %q, want 'device1'", status.Name)
	}
	if status.Model != "Shelly Pro 1PM" {
		t.Errorf("Model = %q, want 'Shelly Pro 1PM'", status.Model)
	}
	if !status.Online {
		t.Error("Online = false, want true")
	}
}
