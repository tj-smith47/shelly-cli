package shelly

import (
	"testing"
)

const (
	testModel     = "SNSW-001P16EU"
	testValueName = "test"
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

func TestConvertToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name: "simple struct",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  testValueName,
				Value: 42,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				t.Helper()
				if result["name"] != testValueName {
					t.Errorf("name = %v, want %s", result["name"], testValueName)
				}
				if result["value"] != float64(42) { // JSON numbers are float64
					t.Errorf("value = %v, want 42", result["value"])
				}
			},
		},
		{
			name: "nested struct",
			input: struct {
				Device struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"device"`
			}{
				Device: struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				}{
					ID:   "test-id",
					Type: "plug",
				},
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				t.Helper()
				device, ok := result["device"].(map[string]any)
				if !ok {
					t.Fatal("device is not a map")
				}
				if device["id"] != "test-id" {
					t.Errorf("device.id = %v, want test-id", device["id"])
				}
			},
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				t.Helper()
				if result != nil {
					t.Errorf("expected nil result for nil input, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := convertToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}
