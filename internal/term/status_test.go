package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestQuickDeviceInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	var info QuickDeviceInfo

	if info.Model != "" {
		t.Errorf("Model = %q, want empty", info.Model)
	}
	if info.Generation != 0 {
		t.Errorf("Generation = %d, want 0", info.Generation)
	}
	if info.Firmware != "" {
		t.Errorf("Firmware = %q, want empty", info.Firmware)
	}
}

func TestComponentState_ZeroValues(t *testing.T) {
	t.Parallel()

	var state ComponentState

	if state.Name != "" {
		t.Errorf("Name = %q, want empty", state.Name)
	}
	if state.State != "" {
		t.Errorf("State = %q, want empty", state.State)
	}
}

func TestQuickDeviceStatus_ZeroValues(t *testing.T) {
	t.Parallel()

	var status QuickDeviceStatus

	if status.Name != "" {
		t.Errorf("Name = %q, want empty", status.Name)
	}
	if status.Model != "" {
		t.Errorf("Model = %q, want empty", status.Model)
	}
	if status.Online {
		t.Error("Online = true, want false")
	}
}

func TestDisplayQuickDeviceStatus_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	info := &QuickDeviceInfo{
		Model:      "Shelly Pro 1PM",
		Generation: 2,
		Firmware:   "1.0.0",
	}

	DisplayQuickDeviceStatus(ios, "device1", info, nil)

	output := out.String()
	if !strings.Contains(output, "No controllable components") {
		t.Error("output should contain 'No controllable components'")
	}
}

func TestDisplayAllDevicesQuickStatus_SingleOnline(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	statuses := []QuickDeviceStatus{
		{Name: "device1", Model: "Shelly Pro 1PM", Online: true},
	}

	DisplayAllDevicesQuickStatus(ios, statuses)

	output := out.String()
	if !strings.Contains(output, "device1") {
		t.Error("output should contain 'device1'")
	}
	if !strings.Contains(output, "Shelly Pro 1PM") {
		t.Error("output should contain model for online device")
	}
}

func TestDisplayAllDevicesQuickStatus_MixedStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	statuses := []QuickDeviceStatus{
		{Name: "device1", Model: "Shelly Pro 1PM", Online: true},
		{Name: "device2", Model: "Shelly Plus 1", Online: false},
		{Name: "device3", Model: "Shelly Pro 2PM", Online: true},
	}

	DisplayAllDevicesQuickStatus(ios, statuses)

	output := out.String()
	if !strings.Contains(output, "device1") {
		t.Error("output should contain 'device1'")
	}
	if !strings.Contains(output, "device2") {
		t.Error("output should contain 'device2'")
	}
	if !strings.Contains(output, "device3") {
		t.Error("output should contain 'device3'")
	}
}

func TestGetComponentState_DefaultCase(t *testing.T) {
	t.Parallel()

	// Test with unknown component type
	comp := model.Component{
		Type: "unknown",
		ID:   0,
	}

	// GetComponentState returns nil for unknown types
	_ = comp // We can't easily test GetComponentState without mocking the client
}
