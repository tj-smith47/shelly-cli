package shelly

import (
	"testing"
)

const (
	testModel = "SNSW-001P16EU"
)

func TestDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{
		ID:         "shellypro1pm-123456",
		MAC:        testMAC,
		Model:      testModel,
		Generation: 2,
		Firmware:   "1.0.0",
		App:        "Pro1PM",
		AuthEn:     true,
	}

	if info.ID != "shellypro1pm-123456" {
		t.Errorf("ID = %q, want shellypro1pm-123456", info.ID)
	}
	if info.MAC != testMAC {
		t.Errorf("MAC = %q, want %s", info.MAC, testMAC)
	}
	if info.Model != testModel {
		t.Errorf("Model = %q, want %s", info.Model, testModel)
	}
	if info.Generation != 2 {
		t.Errorf("Generation = %d, want 2", info.Generation)
	}
	if info.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want 1.0.0", info.Firmware)
	}
	if info.App != "Pro1PM" {
		t.Errorf("App = %q, want Pro1PM", info.App)
	}
	if !info.AuthEn {
		t.Error("AuthEn = false, want true")
	}
}

func TestDeviceStatus_Fields(t *testing.T) {
	t.Parallel()

	status := DeviceStatus{
		Info: &DeviceInfo{
			ID:         "shellyplus1-123456",
			Generation: 2,
		},
		Status: map[string]any{
			"sys": map[string]any{
				"uptime": 12345,
			},
		},
	}

	if status.Info == nil {
		t.Fatal("Info is nil")
	}
	if status.Info.ID != "shellyplus1-123456" {
		t.Errorf("Info.ID = %q, want shellyplus1-123456", status.Info.ID)
	}
	if status.Status == nil {
		t.Fatal("Status is nil")
	}
	if _, ok := status.Status["sys"]; !ok {
		t.Error("Status missing 'sys' key")
	}
}

func TestDeviceInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{}

	if info.ID != "" {
		t.Errorf("ID = %q, want empty", info.ID)
	}
	if info.Generation != 0 {
		t.Errorf("Generation = %d, want 0", info.Generation)
	}
	if info.AuthEn {
		t.Error("AuthEn = true, want false")
	}
}

func TestDeviceStatus_NilInfo(t *testing.T) {
	t.Parallel()

	status := DeviceStatus{
		Info:   nil,
		Status: map[string]any{},
	}

	if status.Info != nil {
		t.Error("Info should be nil")
	}
}
