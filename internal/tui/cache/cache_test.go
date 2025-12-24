// Package cache provides a shared device data cache for the TUI.
package cache

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestSortStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "already sorted",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "reverse order",
			input: []string{"c", "b", "a"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "random order",
			input: []string{"kitchen", "bedroom", "office", "garage"},
			want:  []string{"bedroom", "garage", "kitchen", "office"},
		},
		{
			name:  "single element",
			input: []string{"only"},
			want:  []string{"only"},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := make([]string, len(tt.input))
			copy(input, tt.input)

			sortStrings(input)

			if len(input) != len(tt.want) {
				t.Fatalf("sortStrings() length = %d, want %d", len(input), len(tt.want))
			}

			for i := range input {
				if input[i] != tt.want[i] {
					t.Errorf("sortStrings()[%d] = %q, want %q", i, input[i], tt.want[i])
				}
			}
		})
	}
}

func TestDeviceData_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	data := DeviceData{
		Device: model.Device{
			Address: "192.168.1.100",
		},
		Online:      true,
		Fetched:     true,
		Power:       45.2,
		Voltage:     120.5,
		Current:     0.38,
		TotalEnergy: 1234.5,
		Temperature: 25.3,
		Switches: []SwitchState{
			{ID: 0, On: true, Source: "button"},
			{ID: 1, On: false, Source: "init"},
		},
		UpdatedAt: now,
	}

	if !data.Online {
		t.Error("DeviceData.Online should be true")
	}
	if !data.Fetched {
		t.Error("DeviceData.Fetched should be true")
	}
	if data.Power != 45.2 {
		t.Errorf("DeviceData.Power = %f, want 45.2", data.Power)
	}
	if data.Voltage != 120.5 {
		t.Errorf("DeviceData.Voltage = %f, want 120.5", data.Voltage)
	}
	if len(data.Switches) != 2 {
		t.Errorf("DeviceData.Switches length = %d, want 2", len(data.Switches))
	}
	if !data.Switches[0].On {
		t.Error("DeviceData.Switches[0].On should be true")
	}
}

func TestSwitchState_Fields(t *testing.T) {
	t.Parallel()

	state := SwitchState{
		ID:     0,
		On:     true,
		Source: "button",
	}

	if state.ID != 0 {
		t.Errorf("SwitchState.ID = %d, want 0", state.ID)
	}
	if !state.On {
		t.Error("SwitchState.On should be true")
	}
	if state.Source != "button" {
		t.Errorf("SwitchState.Source = %q, want 'button'", state.Source)
	}
}

func TestDeviceUpdateMsg_Fields(t *testing.T) {
	t.Parallel()

	data := &DeviceData{
		Device: model.Device{Address: "192.168.1.100"},
		Online: true,
	}

	msg := DeviceUpdateMsg{
		Name: "kitchen",
		Data: data,
	}

	if msg.Name != "kitchen" {
		t.Errorf("DeviceUpdateMsg.Name = %q, want 'kitchen'", msg.Name)
	}
	if msg.Data != data {
		t.Error("DeviceUpdateMsg.Data should match input")
	}
}

func TestAllDevicesLoadedMsg(t *testing.T) {
	t.Parallel()

	// Just verify the type exists and can be instantiated
	msg := AllDevicesLoadedMsg{}
	_ = msg
}

func TestRefreshTickMsg(t *testing.T) {
	t.Parallel()

	// Just verify the type exists and can be instantiated
	msg := RefreshTickMsg{}
	_ = msg
}

func TestCache_DeviceCountEmpty(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices: make(map[string]*DeviceData),
	}

	if count := c.DeviceCount(); count != 0 {
		t.Errorf("DeviceCount() = %d, want 0", count)
	}
}

func TestCache_OnlineCountEmpty(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices: make(map[string]*DeviceData),
	}

	if count := c.OnlineCount(); count != 0 {
		t.Errorf("OnlineCount() = %d, want 0", count)
	}
}

func TestCache_TotalPowerEmpty(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices: make(map[string]*DeviceData),
	}

	if power := c.TotalPower(); power != 0 {
		t.Errorf("TotalPower() = %f, want 0", power)
	}
}

func TestCache_GetDevice(t *testing.T) {
	t.Parallel()

	data := &DeviceData{
		Device: model.Device{Address: "192.168.1.100"},
		Online: true,
		Power:  45.2,
	}

	c := &Cache{
		devices: map[string]*DeviceData{
			"kitchen": data,
		},
	}

	// Get existing device
	got := c.GetDevice("kitchen")
	if got != data {
		t.Error("GetDevice('kitchen') should return cached data")
	}

	// Get non-existing device
	got = c.GetDevice("nonexistent")
	if got != nil {
		t.Error("GetDevice('nonexistent') should return nil")
	}
}

func TestCache_GetAllDevices(t *testing.T) {
	t.Parallel()

	kitchen := &DeviceData{Device: model.Device{Address: "192.168.1.100"}}
	office := &DeviceData{Device: model.Device{Address: "192.168.1.101"}}

	c := &Cache{
		devices: map[string]*DeviceData{
			"kitchen": kitchen,
			"office":  office,
		},
		order: []string{"kitchen", "office"},
	}

	all := c.GetAllDevices()
	if len(all) != 2 {
		t.Errorf("GetAllDevices() length = %d, want 2", len(all))
	}

	// Verify order is preserved
	if all[0] != kitchen {
		t.Error("GetAllDevices()[0] should be kitchen")
	}
	if all[1] != office {
		t.Error("GetAllDevices()[1] should be office")
	}
}

func TestCache_GetOnlineDevices(t *testing.T) {
	t.Parallel()

	kitchen := &DeviceData{
		Device: model.Device{Address: "192.168.1.100"},
		Online: true,
	}
	office := &DeviceData{
		Device: model.Device{Address: "192.168.1.101"},
		Online: false,
	}
	bedroom := &DeviceData{
		Device: model.Device{Address: "192.168.1.102"},
		Online: true,
	}

	c := &Cache{
		devices: map[string]*DeviceData{
			"kitchen": kitchen,
			"office":  office,
			"bedroom": bedroom,
		},
		order: []string{"bedroom", "kitchen", "office"},
	}

	online := c.GetOnlineDevices()
	if len(online) != 2 {
		t.Errorf("GetOnlineDevices() length = %d, want 2", len(online))
	}

	// All returned devices should be online
	for _, d := range online {
		if !d.Online {
			t.Errorf("GetOnlineDevices() returned offline device: %s", d.Device.Address)
		}
	}
}

func TestCache_DeviceCount(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices: map[string]*DeviceData{
			"a": {},
			"b": {},
			"c": {},
		},
	}

	if count := c.DeviceCount(); count != 3 {
		t.Errorf("DeviceCount() = %d, want 3", count)
	}
}

func TestCache_OnlineCount(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices: map[string]*DeviceData{
			"a": {Online: true},
			"b": {Online: false},
			"c": {Online: true},
			"d": {Online: true},
		},
	}

	if count := c.OnlineCount(); count != 3 {
		t.Errorf("OnlineCount() = %d, want 3", count)
	}
}

func TestCache_TotalPower(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices: map[string]*DeviceData{
			"a": {Online: true, Power: 10.5},
			"b": {Online: false, Power: 20.0}, // Offline - should not count
			"c": {Online: true, Power: 30.2},
		},
	}

	expected := 10.5 + 30.2 // Only online devices
	if power := c.TotalPower(); power != expected {
		t.Errorf("TotalPower() = %f, want %f", power, expected)
	}
}

func TestCache_IsLoading(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialLoad  bool
		pendingCount int
		want         bool
	}{
		{"initial load with pending", true, 5, true},
		{"initial load complete", true, 0, false},
		{"not initial load", false, 5, false},
		{"not initial load no pending", false, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Cache{
				devices:      make(map[string]*DeviceData),
				initialLoad:  tt.initialLoad,
				pendingCount: tt.pendingCount,
			}

			if got := c.IsLoading(); got != tt.want {
				t.Errorf("IsLoading() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_FetchedCount(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices: map[string]*DeviceData{
			"a": {Fetched: true},
			"b": {Fetched: false},
			"c": {Fetched: true},
			"d": {Fetched: true},
		},
	}

	if count := c.FetchedCount(); count != 3 {
		t.Errorf("FetchedCount() = %d, want 3", count)
	}
}
