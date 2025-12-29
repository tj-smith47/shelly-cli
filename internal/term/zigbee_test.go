package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayZigbeeDevices_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayZigbeeDevices(ios, []model.ZigbeeDevice{})

	output := out.String()
	if !strings.Contains(output, "No Zigbee-capable devices") {
		t.Error("expected no devices message")
	}
	if !strings.Contains(output, "Gen4") {
		t.Error("expected Gen4 hint")
	}
}

func TestDisplayZigbeeDevices_WithDevices(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []model.ZigbeeDevice{
		{
			Name:         "Hub Device",
			Address:      "192.168.1.100",
			Model:        "SNSN-0043X",
			Enabled:      true,
			NetworkState: "joined",
			EUI64:        "00:11:22:FF:FE:33:44:55",
		},
		{
			Name:         "Second Hub",
			Address:      "192.168.1.101",
			Model:        "",
			Enabled:      false,
			NetworkState: "",
			EUI64:        "",
		},
	}
	DisplayZigbeeDevices(ios, devices)

	output := out.String()
	if !strings.Contains(output, "Zigbee-Capable Devices (2)") {
		t.Error("expected device count header")
	}
	if !strings.Contains(output, "Hub Device") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "192.168.1.100") {
		t.Error("expected address")
	}
	if !strings.Contains(output, "SNSN-0043X") {
		t.Error("expected model")
	}
	if !strings.Contains(output, "00:11:22:FF:FE:33:44:55") {
		t.Error("expected EUI64")
	}
}

func TestOutputZigbeeDevicesJSON(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []model.ZigbeeDevice{
		{Name: "test", Address: "192.168.1.1", Enabled: true},
	}
	err := OutputZigbeeDevicesJSON(ios, devices)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "test") {
		t.Error("expected device name in JSON")
	}
	if !strings.Contains(output, "192.168.1.1") {
		t.Error("expected address in JSON")
	}
}

func TestDisplayZigbeeStatus_Joined(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.ZigbeeStatus{
		Enabled:          true,
		NetworkState:     "joined",
		EUI64:            "AA:BB:CC:FF:FE:DD:EE:FF",
		PANID:            0x1234,
		Channel:          15,
		CoordinatorEUI64: "11:22:33:FF:FE:44:55:66",
	}
	DisplayZigbeeStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "Zigbee Status") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "Enabled") {
		t.Error("expected enabled status")
	}
	if !strings.Contains(output, "joined") {
		t.Error("expected network state")
	}
	if !strings.Contains(output, "Network Info") {
		t.Error("expected network info section")
	}
	if !strings.Contains(output, "PAN ID: 0x1234") {
		t.Error("expected PAN ID")
	}
	if !strings.Contains(output, "Channel: 15") {
		t.Error("expected channel")
	}
	if !strings.Contains(output, "11:22:33:FF:FE:44:55:66") {
		t.Error("expected coordinator EUI64")
	}
}

func TestDisplayZigbeeStatus_NotJoined(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.ZigbeeStatus{
		Enabled:      true,
		NetworkState: "steering",
		EUI64:        "AA:BB:CC:FF:FE:DD:EE:FF",
	}
	DisplayZigbeeStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "steering") {
		t.Error("expected network state")
	}
	// Should not show network info when not joined
	if strings.Contains(output, "Network Info") {
		t.Error("should not show network info when not joined")
	}
}

func TestDisplayZigbeeStatus_Disabled(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.ZigbeeStatus{
		Enabled: false,
	}
	DisplayZigbeeStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "Zigbee Status") {
		t.Error("expected header")
	}
}

func TestOutputZigbeeStatusJSON(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.ZigbeeStatus{
		Enabled:      true,
		NetworkState: "joined",
		PANID:        0xABCD,
		Channel:      20,
	}
	err := OutputZigbeeStatusJSON(ios, status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "enabled") {
		t.Error("expected enabled field in JSON")
	}
	if !strings.Contains(output, "joined") {
		t.Error("expected network_state in JSON")
	}
}
