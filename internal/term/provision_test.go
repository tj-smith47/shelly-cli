package term

import (
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayBulkProvisionDryRun_WithWiFi(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	cfg := &model.BulkProvisionConfig{
		WiFi: &model.ProvisionWiFiConfig{SSID: "GlobalSSID"},
		Devices: []model.DeviceProvisionConfig{
			{Name: "device1"},
			{Name: "device2", WiFi: &model.ProvisionWiFiConfig{SSID: "CustomSSID"}},
		},
	}
	DisplayBulkProvisionDryRun(ios, cfg)

	output := out.String()
	if !strings.Contains(output, "Dry run") {
		t.Error("expected dry run notice")
	}
	if !strings.Contains(output, "device1: SSID=GlobalSSID") {
		t.Error("expected device1 with global SSID")
	}
	if !strings.Contains(output, "device2: SSID=CustomSSID") {
		t.Error("expected device2 with custom SSID")
	}
}

func TestDisplayBulkProvisionDryRun_NoWiFi(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()
	cfg := &model.BulkProvisionConfig{
		WiFi: nil,
		Devices: []model.DeviceProvisionConfig{
			{Name: "device1"},
		},
	}
	DisplayBulkProvisionDryRun(ios, cfg)

	// Warning goes to stderr
	output := errOut.String()
	if !strings.Contains(output, "no WiFi config") {
		t.Error("expected no WiFi warning")
	}
}

func TestDisplayBulkProvisionResults_AllSuccess(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	results := []model.ProvisionResult{
		{Device: "device1", Err: nil},
		{Device: "device2", Err: nil},
		{Device: "device3", Err: nil},
	}
	failed := DisplayBulkProvisionResults(ios, results, 3)

	output := out.String()
	if failed != 0 {
		t.Errorf("expected 0 failures, got %d", failed)
	}
	if !strings.Contains(output, "All 3 devices provisioned successfully") {
		t.Error("expected all success message")
	}
}

func TestDisplayBulkProvisionResults_SomeFailures(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	results := []model.ProvisionResult{
		{Device: "device1", Err: nil},
		{Device: "device2", Err: errors.New("connection refused")},
		{Device: "device3", Err: nil},
	}
	failed := DisplayBulkProvisionResults(ios, results, 3)

	if failed != 1 {
		t.Errorf("expected 1 failure, got %d", failed)
	}
	if !strings.Contains(out.String(), "Provisioned device1") {
		t.Error("expected success message for device1")
	}
	if !strings.Contains(errOut.String(), "Failed to provision device2") {
		t.Error("expected failure message for device2")
	}
	if strings.Contains(out.String(), "All 3 devices") {
		t.Error("should not show all success message when failures exist")
	}
}

func TestDisplayBulkProvisionResults_AllFailed(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()
	results := []model.ProvisionResult{
		{Device: "device1", Err: errors.New("timeout")},
		{Device: "device2", Err: errors.New("auth failed")},
	}
	failed := DisplayBulkProvisionResults(ios, results, 2)

	if failed != 2 {
		t.Errorf("expected 2 failures, got %d", failed)
	}
	if !strings.Contains(errOut.String(), "Failed to provision device1") {
		t.Error("expected failure message for device1")
	}
	if !strings.Contains(errOut.String(), "Failed to provision device2") {
		t.Error("expected failure message for device2")
	}
}
