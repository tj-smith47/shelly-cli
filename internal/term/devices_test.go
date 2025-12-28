package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayDeviceList(t *testing.T) {
	t.Parallel()

	t.Run("basic list", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []model.DeviceListItem{
			{Name: "device1", Address: "192.168.1.10", Type: "switch", Model: "SNSW-001P16EU", Generation: 2, Auth: false},
			{Name: "device2", Address: "192.168.1.11", Type: "dimmer", Model: "SNDM-0013US", Generation: 2, Auth: true},
		}

		DisplayDeviceList(ios, devices, false, false)

		output := out.String()
		if !strings.Contains(output, "device1") {
			t.Error("output should contain 'device1'")
		}
		if !strings.Contains(output, "device2") {
			t.Error("output should contain 'device2'")
		}
		if !strings.Contains(output, "192.168.1.10") {
			t.Error("output should contain address")
		}
	})

	t.Run("with platform", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []model.DeviceListItem{
			{Name: "device1", Address: "192.168.1.10", Platform: "shelly", Type: "switch", Model: "SNSW-001P16EU", Generation: 2},
		}

		DisplayDeviceList(ios, devices, true, false)

		output := out.String()
		// Platform value should appear in output when showPlatform is true
		if !strings.Contains(output, "shelly") {
			t.Error("output should contain platform value")
		}
	})

	t.Run("with version", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []model.DeviceListItem{
			{Name: "device1", Address: "192.168.1.10", Type: "switch", Model: "SNSW-001P16EU", Generation: 2, CurrentVersion: "1.0.0", AvailableVersion: "1.1.0", HasUpdate: true},
		}

		DisplayDeviceList(ios, devices, false, true)

		output := out.String()
		// Current version should appear in output when showVersion is true
		if !strings.Contains(output, "1.0.0") {
			t.Error("output should contain current version")
		}
	})

	t.Run("with update available", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []model.DeviceListItem{
			{Name: "device1", Address: "192.168.1.10", Type: "switch", Generation: 2, HasUpdate: true, AvailableVersion: "1.1.0"},
		}

		DisplayDeviceList(ios, devices, false, true)

		output := out.String()
		if !strings.Contains(output, "1.1.0") {
			t.Error("output should contain available version")
		}
	})

	t.Run("empty version", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []model.DeviceListItem{
			{Name: "device1", Address: "192.168.1.10", Type: "switch", Generation: 2, CurrentVersion: ""},
		}

		DisplayDeviceList(ios, devices, false, true)

		output := out.String()
		if !strings.Contains(output, "-") {
			t.Error("output should contain '-' for missing version")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()

		DisplayDeviceList(ios, []model.DeviceListItem{}, false, false)

		output := out.String()
		// Even empty list should print table headers
		if output == "" {
			t.Error("DisplayDeviceList should produce output")
		}
	})

	t.Run("all options enabled", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []model.DeviceListItem{
			{
				Name:             "device1",
				Address:          "192.168.1.10",
				Platform:         "shelly",
				Type:             "switch",
				Model:            "SNSW-001P16EU",
				Generation:       2,
				Auth:             true,
				CurrentVersion:   "1.0.0",
				AvailableVersion: "1.1.0",
				HasUpdate:        true,
			},
		}

		DisplayDeviceList(ios, devices, true, true)

		output := out.String()
		// Check that the output contains the device data
		if !strings.Contains(output, "device1") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "shelly") {
			t.Error("output should contain platform value")
		}
		if !strings.Contains(output, "1.0.0") {
			t.Error("output should contain version")
		}
	})
}

func TestDisplayQuickDeviceStatus(t *testing.T) {
	t.Parallel()

	t.Run("with components", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		info := &QuickDeviceInfo{
			Model:      "Shelly Pro 1PM",
			Generation: 2,
			Firmware:   "1.0.0",
		}
		states := []ComponentState{
			{Name: "switch:0", State: "ON"},
			{Name: "input:0", State: "inactive"},
		}

		DisplayQuickDeviceStatus(ios, "device1", info, states)

		output := out.String()
		if !strings.Contains(output, "device1") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "Shelly Pro 1PM") {
			t.Error("output should contain model")
		}
		if !strings.Contains(output, "switch:0") {
			t.Error("output should contain component name")
		}
		if !strings.Contains(output, "ON") {
			t.Error("output should contain component state")
		}
	})

	t.Run("no components", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		info := &QuickDeviceInfo{
			Model:      "Shelly Button",
			Generation: 2,
			Firmware:   "1.0.0",
		}

		DisplayQuickDeviceStatus(ios, "button1", info, []ComponentState{})

		output := out.String()
		if !strings.Contains(output, "No controllable components") {
			t.Error("output should contain 'No controllable components'")
		}
	})
}

func TestDisplayAllDevicesQuickStatus(t *testing.T) {
	t.Parallel()

	t.Run("with devices", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		statuses := []QuickDeviceStatus{
			{Name: "device1", Model: "Shelly Pro 1PM", Online: true},
			{Name: "device2", Model: "Shelly Plus 1", Online: false},
		}

		DisplayAllDevicesQuickStatus(ios, statuses)

		output := out.String()
		if !strings.Contains(output, "device1") {
			t.Error("output should contain 'device1'")
		}
		if !strings.Contains(output, "device2") {
			t.Error("output should contain 'device2'")
		}
	})

	t.Run("offline device shows dash model", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		statuses := []QuickDeviceStatus{
			{Name: "device1", Model: "Shelly Pro 1PM", Online: false},
		}

		DisplayAllDevicesQuickStatus(ios, statuses)

		output := out.String()
		// Offline devices should show "-" for model
		if !strings.Contains(output, "-") {
			t.Error("output should contain '-' for offline device model")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		ios, _, errOut := testIOStreams()

		DisplayAllDevicesQuickStatus(ios, []QuickDeviceStatus{})

		// Warning goes to errOut, not out
		errOutput := errOut.String()
		if !strings.Contains(errOutput, "No devices registered") {
			t.Errorf("errOut should contain 'No devices registered', got %q", errOutput)
		}
	})
}
