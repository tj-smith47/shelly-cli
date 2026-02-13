package term

import (
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestDisplayOnboardDevices_Empty(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	DisplayOnboardDevices(ios, nil)

	combined := out.String() + errOut.String()
	if !strings.Contains(combined, "No") {
		t.Error("expected no results message for empty devices")
	}
}

func TestDisplayOnboardDevices_WithDevices(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []shelly.OnboardDevice{
		{
			Name:       "shelly-plus-1",
			Model:      "SNSW-001P16EU",
			Source:     shelly.OnboardSourceBLE,
			Generation: 2,
		},
		{
			Name:        "shelly-1",
			Model:       "SHSW-1",
			Source:      shelly.OnboardSourceWiFiAP,
			Generation:  1,
			Provisioned: false,
		},
	}
	DisplayOnboardDevices(ios, devices)

	output := out.String()
	if !strings.Contains(output, "shelly-plus-1") {
		t.Error("expected device name in output")
	}
	if !strings.Contains(output, "BLE") {
		t.Error("expected BLE source in output")
	}
	if !strings.Contains(output, "2 device") {
		t.Error("expected device count")
	}
}

func TestDisplayOnboardDevices_RegisteredStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []shelly.OnboardDevice{
		{Name: "registered-dev", Registered: true, Generation: 2},
		{Name: "network-dev", Provisioned: true, Generation: 2},
		{Name: "new-dev", Generation: 2},
	}
	DisplayOnboardDevices(ios, devices)

	output := out.String()
	if !strings.Contains(output, "registered") {
		t.Error("expected 'registered' status")
	}
	if !strings.Contains(output, "on network") {
		t.Error("expected 'on network' status")
	}
	if !strings.Contains(output, "new") {
		t.Error("expected 'new' status")
	}
}

func TestDisplayOnboardResults_Success(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	results := []*shelly.OnboardResult{
		{
			Device:     &shelly.OnboardDevice{Name: "dev-1"},
			NewAddress: "192.168.1.50",
			Registered: true,
			Method:     "BLE",
		},
	}
	DisplayOnboardResults(ios, results)

	output := out.String()
	if !strings.Contains(output, "dev-1") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "BLE") {
		t.Error("expected method")
	}
	if !strings.Contains(output, "192.168.1.50") {
		t.Error("expected new address")
	}
	if !strings.Contains(output, "registered") {
		t.Error("expected registered indicator")
	}
}

func TestDisplayOnboardResults_Error(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()
	results := []*shelly.OnboardResult{
		{
			Device: &shelly.OnboardDevice{Name: "dev-fail"},
			Error:  errors.New("BLE init failed"),
			Method: "BLE",
		},
	}
	DisplayOnboardResults(ios, results)

	output := errOut.String()
	if !strings.Contains(output, "dev-fail") {
		t.Error("expected device name in error")
	}
	if !strings.Contains(output, "BLE init failed") {
		t.Error("expected error message")
	}
}

func TestDisplayOnboardSummary_AllSuccess(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	results := []*shelly.OnboardResult{
		{Device: &shelly.OnboardDevice{Name: "dev-1"}, NewAddress: "192.168.1.50"},
		{Device: &shelly.OnboardDevice{Name: "dev-2"}, NewAddress: "192.168.1.51"},
	}
	DisplayOnboardSummary(ios, results)

	output := out.String()
	if !strings.Contains(output, "All 2 devices provisioned successfully") {
		t.Error("expected all success message")
	}
	if !strings.Contains(output, "192.168.1.50") {
		t.Error("expected address mapping")
	}
}

func TestDisplayOnboardSummary_SomeFailures(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()
	results := []*shelly.OnboardResult{
		{Device: &shelly.OnboardDevice{Name: "dev-1"}, NewAddress: "192.168.1.50"},
		{Device: &shelly.OnboardDevice{Name: "dev-2"}, Error: errors.New("timeout")},
	}
	DisplayOnboardSummary(ios, results)

	output := errOut.String()
	if !strings.Contains(output, "1 of 2") {
		t.Error("expected partial success message")
	}
	if !strings.Contains(output, "1 failed") {
		t.Error("expected failure count")
	}
}

func TestDisplayOnboardSummary_AllFailed(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()
	results := []*shelly.OnboardResult{
		{Device: &shelly.OnboardDevice{Name: "dev-1"}, Error: errors.New("err1")},
		{Device: &shelly.OnboardDevice{Name: "dev-2"}, Error: errors.New("err2")},
	}
	DisplayOnboardSummary(ios, results)

	output := errOut.String()
	if !strings.Contains(output, "0 of 2") {
		t.Error("expected zero success message")
	}
}

func TestFormatOnboardDeviceOptions(t *testing.T) {
	t.Parallel()

	devices := []shelly.OnboardDevice{
		{Name: "shelly-plus-1", Generation: 2, Source: shelly.OnboardSourceBLE},
		{Name: "shelly-1", Generation: 1, Source: shelly.OnboardSourceWiFiAP},
		{Name: "", Address: "192.168.1.50", Generation: 0, Source: shelly.OnboardSourceHTTP},
	}

	options := FormatOnboardDeviceOptions(devices)
	if len(options) != 3 {
		t.Fatalf("len(options) = %d, want 3", len(options))
	}
	if !strings.Contains(options[0], "shelly-plus-1") {
		t.Errorf("options[0] = %q, want to contain device name", options[0])
	}
	if !strings.Contains(options[0], "Gen2") {
		t.Errorf("options[0] = %q, want to contain Gen2", options[0])
	}
	if !strings.Contains(options[0], "BLE") {
		t.Errorf("options[0] = %q, want to contain BLE", options[0])
	}
	if !strings.Contains(options[1], "Gen1") {
		t.Errorf("options[1] = %q, want to contain Gen1", options[1])
	}
	if !strings.Contains(options[1], "WiFi AP") {
		t.Errorf("options[1] = %q, want to contain WiFi AP", options[1])
	}
	// Unknown generation
	if !strings.Contains(options[2], "Gen?") {
		t.Errorf("options[2] = %q, want to contain Gen?", options[2])
	}
	// Fallback to address
	if !strings.Contains(options[2], "192.168.1.50") {
		t.Errorf("options[2] = %q, want to contain address", options[2])
	}
}

func TestFormatOnboardDeviceOptions_Empty(t *testing.T) {
	t.Parallel()

	options := FormatOnboardDeviceOptions(nil)
	if len(options) != 0 {
		t.Errorf("len(options) = %d, want 0", len(options))
	}
}

func TestSelectOnboardDevices_AutoConfirm(t *testing.T) {
	t.Parallel()

	devices := []shelly.OnboardDevice{
		{Name: "dev-1"},
		{Name: "dev-2"},
	}

	selected, err := SelectOnboardDevices(nil, devices, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 2 {
		t.Errorf("len(selected) = %d, want 2", len(selected))
	}
}

func TestSelectOnboardDevices_AutoConfirm_Empty(t *testing.T) {
	t.Parallel()

	selected, err := SelectOnboardDevices(nil, nil, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 0 {
		t.Errorf("len(selected) = %d, want 0", len(selected))
	}
}
