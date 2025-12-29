package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayBTHomeAddResult(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := BTHomeAddResult{
		Key:  "bthomedevice:200",
		Name: "Temperature Sensor",
		Addr: "AA:BB:CC:DD:EE:FF",
	}
	DisplayBTHomeAddResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "BTHome device added") {
		t.Error("expected success message")
	}
	if !strings.Contains(output, "bthomedevice:200") {
		t.Error("expected key")
	}
	if !strings.Contains(output, "Temperature Sensor") {
		t.Error("expected name")
	}
	if !strings.Contains(output, "AA:BB:CC:DD:EE:FF") {
		t.Error("expected address")
	}
}

func TestDisplayBTHomeAddResult_NoName(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := BTHomeAddResult{
		Key:  "bthomedevice:201",
		Name: "",
		Addr: "11:22:33:44:55:66",
	}
	DisplayBTHomeAddResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "bthomedevice:201") {
		t.Error("expected key")
	}
	// Name line should not appear when empty
	if !strings.Contains(output, "Address:") {
		t.Error("expected address line")
	}
}

func TestDisplayBTHomeDiscoveryStarted(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayBTHomeDiscoveryStarted(ios, "shellyplus1-abc123", 60)

	output := out.String()
	if !strings.Contains(output, "BTHome Device Discovery Started") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "60 seconds") {
		t.Error("expected duration")
	}
	if !strings.Contains(output, "shellyplus1-abc123") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "device_discovered") {
		t.Error("expected event info")
	}
	if !strings.Contains(output, "shelly monitor events") {
		t.Error("expected monitor command")
	}
	if !strings.Contains(output, "shelly bthome add") {
		t.Error("expected add command")
	}
}

func TestDisplayBTHomeDevices_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayBTHomeDevices(ios, []model.BTHomeDeviceInfo{}, "gateway-device")

	output := out.String()
	if !strings.Contains(output, "No BTHome devices found") {
		t.Error("expected no devices message")
	}
	if !strings.Contains(output, "shelly bthome add gateway-device") {
		t.Error("expected add command hint")
	}
}

func TestDisplayBTHomeDevices_WithDevices(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	rssi := -65
	battery := 85
	devices := []model.BTHomeDeviceInfo{
		{
			ID:      100,
			Name:    "Temperature Sensor",
			Addr:    "AA:BB:CC:DD:EE:FF",
			RSSI:    &rssi,
			Battery: &battery,
		},
		{
			ID:   101,
			Name: "",
			Addr: "11:22:33:44:55:66",
		},
	}
	DisplayBTHomeDevices(ios, devices, "gateway")

	output := out.String()
	if !strings.Contains(output, "BTHome Devices (2)") {
		t.Error("expected device count header")
	}
	if !strings.Contains(output, "Temperature Sensor") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "ID: 100") {
		t.Error("expected device ID")
	}
	if !strings.Contains(output, "-65 dBm") {
		t.Error("expected RSSI")
	}
	if !strings.Contains(output, "85%") {
		t.Error("expected battery percentage")
	}
	if !strings.Contains(output, "Device 101") {
		t.Error("expected generated name for device without name")
	}
}

func TestDisplayBTHomeComponentStatus_WithDiscovery(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.BTHomeComponentStatus{
		Discovery: &model.BTHomeDiscoveryStatus{
			StartedAt: 1700000000,
			Duration:  60,
		},
	}
	DisplayBTHomeComponentStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "BTHome Status") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "Discovery") {
		t.Error("expected discovery section")
	}
	if !strings.Contains(output, "60s") {
		t.Error("expected duration")
	}
}

func TestDisplayBTHomeComponentStatus_NoDiscovery(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.BTHomeComponentStatus{
		Discovery: nil,
	}
	DisplayBTHomeComponentStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "No active discovery scan") {
		t.Error("expected no discovery message")
	}
}

func TestDisplayBTHomeComponentStatus_WithErrors(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	status := model.BTHomeComponentStatus{
		Errors: []string{"Connection timeout", "Invalid data"},
	}
	DisplayBTHomeComponentStatus(ios, status)

	// Error header goes to stderr
	if !strings.Contains(errOut.String(), "Errors") {
		t.Error("expected errors section")
	}
	// Error details go to stdout via Printf
	if !strings.Contains(out.String(), "Connection timeout") {
		t.Error("expected first error")
	}
	if !strings.Contains(out.String(), "Invalid data") {
		t.Error("expected second error")
	}
}

func TestDisplayBTHomeDeviceStatus_Full(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	rssi := -70
	battery := 90
	packetID := 42
	component := "temperature:100"
	status := model.BTHomeDeviceStatus{
		ID:           100,
		Name:         "Temp Sensor",
		Addr:         "AA:BB:CC:DD:EE:FF",
		RSSI:         &rssi,
		Battery:      &battery,
		PacketID:     &packetID,
		LastUpdateTS: 1700000000,
		KnownObjects: []model.BTHomeKnownObj{
			{ObjID: 1, Idx: 0, Component: &component},
		},
	}
	DisplayBTHomeDeviceStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "BTHome Device: Temp Sensor") {
		t.Error("expected device header")
	}
	if !strings.Contains(output, "ID: 100") {
		t.Error("expected ID")
	}
	if !strings.Contains(output, "-70 dBm") {
		t.Error("expected RSSI")
	}
	if !strings.Contains(output, "90%") {
		t.Error("expected battery")
	}
	if !strings.Contains(output, "Packet ID: 42") {
		t.Error("expected packet ID")
	}
	if !strings.Contains(output, "Known Objects") {
		t.Error("expected known objects section")
	}
	if !strings.Contains(output, "temperature:100") {
		t.Error("expected component reference")
	}
}

func TestDisplayBTHomeDeviceStatus_NoName(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.BTHomeDeviceStatus{
		ID:   200,
		Name: "",
	}
	DisplayBTHomeDeviceStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "Device 200") {
		t.Error("expected generated name")
	}
}

func TestDisplayBTHomeDeviceStatus_WithErrors(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	status := model.BTHomeDeviceStatus{
		ID:     100,
		Name:   "Test",
		Errors: []string{"Parse error", "Checksum mismatch"},
	}
	DisplayBTHomeDeviceStatus(ios, status)

	// Error header goes to stderr
	if !strings.Contains(errOut.String(), "Errors") {
		t.Error("expected errors section")
	}
	// Error details go to stdout via Printf
	if !strings.Contains(out.String(), "Parse error") {
		t.Error("expected first error")
	}
}
