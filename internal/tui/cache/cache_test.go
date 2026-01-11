// Package cache provides a shared device data cache for the TUI.
package cache

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const (
	deviceKitchen = "kitchen"
	deviceOffice  = "office"
	deviceBedroom = "bedroom"
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
		Name: deviceKitchen,
		Data: data,
	}

	if msg.Name != deviceKitchen {
		t.Errorf("DeviceUpdateMsg.Name = %q, want %q", msg.Name, deviceKitchen)
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

// === Tests for Phase 0.5.2: Wave-Based Device Loading ===

func TestCreateWaves_Empty(t *testing.T) {
	t.Parallel()

	waves := createWaves(map[string]model.Device{})
	if waves != nil {
		t.Errorf("createWaves(empty) = %v, want nil", waves)
	}
}

func TestCreateWaves_SingleDevice(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"kitchen": {Address: "192.168.1.100", Generation: 2},
	}

	waves := createWaves(devices)
	if len(waves) != 1 {
		t.Fatalf("createWaves() returned %d waves, want 1", len(waves))
	}
	if len(waves[0]) != 1 {
		t.Errorf("waves[0] has %d devices, want 1", len(waves[0]))
	}
}

func TestCreateWaves_ThreeDevices(t *testing.T) {
	t.Parallel()

	// Three devices should fit in a single wave (first wave is 3)
	devices := map[string]model.Device{
		"a": {Address: "192.168.1.100", Generation: 2},
		"b": {Address: "192.168.1.101", Generation: 2},
		"c": {Address: "192.168.1.102", Generation: 2},
	}

	waves := createWaves(devices)
	if len(waves) != 1 {
		t.Fatalf("createWaves() returned %d waves, want 1", len(waves))
	}
	if len(waves[0]) != 3 {
		t.Errorf("waves[0] has %d devices, want 3", len(waves[0]))
	}
}

func TestCreateWaves_MultipleWaves(t *testing.T) {
	t.Parallel()

	// 7 devices: first wave = 3, then 2, then 2
	devices := map[string]model.Device{
		"a": {Address: "192.168.1.100", Generation: 2},
		"b": {Address: "192.168.1.101", Generation: 2},
		"c": {Address: "192.168.1.102", Generation: 2},
		"d": {Address: "192.168.1.103", Generation: 2},
		"e": {Address: "192.168.1.104", Generation: 2},
		"f": {Address: "192.168.1.105", Generation: 2},
		"g": {Address: "192.168.1.106", Generation: 2},
	}

	waves := createWaves(devices)
	if len(waves) != 3 {
		t.Fatalf("createWaves() returned %d waves, want 3", len(waves))
	}
	if len(waves[0]) != 3 {
		t.Errorf("waves[0] has %d devices, want 3", len(waves[0]))
	}
	if len(waves[1]) != 2 {
		t.Errorf("waves[1] has %d devices, want 2", len(waves[1]))
	}
	if len(waves[2]) != 2 {
		t.Errorf("waves[2] has %d devices, want 2", len(waves[2]))
	}
}

func TestCreateWaves_Gen2First(t *testing.T) {
	t.Parallel()

	// Gen2 devices should come before Gen1
	devices := map[string]model.Device{
		"gen1_a": {Address: "192.168.1.100", Generation: 1},
		"gen2_b": {Address: "192.168.1.101", Generation: 2},
		"gen1_c": {Address: "192.168.1.102", Generation: 1},
		"gen2_d": {Address: "192.168.1.103", Generation: 2},
	}

	waves := createWaves(devices)
	if len(waves) < 1 {
		t.Fatal("createWaves() returned no waves")
	}

	// First devices in first wave should be Gen2
	firstWave := waves[0]
	gen2Count := 0
	for _, df := range firstWave {
		if df.Device.Generation != 1 {
			gen2Count++
		}
	}

	// With 4 devices (2 Gen1, 2 Gen2), first wave of 3 should have both Gen2
	if gen2Count < 2 {
		t.Errorf("first wave has %d Gen2 devices, want at least 2", gen2Count)
	}
}

func TestCreateWaves_UnknownGenTreatedAsGen2(t *testing.T) {
	t.Parallel()

	// Generation 0 (unknown) should be treated as Gen2 (prioritized)
	devices := map[string]model.Device{
		"gen1":    {Address: "192.168.1.100", Generation: 1},
		"unknown": {Address: "192.168.1.101", Generation: 0},
	}

	waves := createWaves(devices)
	if len(waves) < 1 {
		t.Fatal("createWaves() returned no waves")
	}

	// Unknown generation should be first (treated as Gen2)
	if waves[0][0].Device.Generation == 1 {
		t.Error("Gen1 device should not be first; unknown generation should be prioritized")
	}
}

// === Tests for Phase 0.5.3: Adaptive Refresh Intervals ===

func TestDefaultRefreshConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultRefreshConfig()

	if cfg.Gen1Online != 15*time.Second {
		t.Errorf("Gen1Online = %v, want 15s", cfg.Gen1Online)
	}
	if cfg.Gen1Offline != 60*time.Second {
		t.Errorf("Gen1Offline = %v, want 60s", cfg.Gen1Offline)
	}
	if cfg.Gen2Online != 5*time.Second {
		t.Errorf("Gen2Online = %v, want 5s", cfg.Gen2Online)
	}
	if cfg.Gen2Offline != 30*time.Second {
		t.Errorf("Gen2Offline = %v, want 30s", cfg.Gen2Offline)
	}
	if cfg.FocusedBoost != 3*time.Second {
		t.Errorf("FocusedBoost = %v, want 3s", cfg.FocusedBoost)
	}
}

func TestGetRefreshInterval_NilData(t *testing.T) {
	t.Parallel()

	c := &Cache{refreshConfig: DefaultRefreshConfig()}

	interval := c.getRefreshInterval(nil)
	if interval != 5*time.Second {
		t.Errorf("getRefreshInterval(nil) = %v, want 5s (Gen2Online)", interval)
	}
}

func TestGetRefreshInterval_Gen1Online(t *testing.T) {
	t.Parallel()

	c := &Cache{refreshConfig: DefaultRefreshConfig()}
	data := &DeviceData{
		Online: true,
		Info:   &shelly.DeviceInfo{Generation: 1},
	}

	interval := c.getRefreshInterval(data)
	if interval != 15*time.Second {
		t.Errorf("getRefreshInterval(Gen1 online) = %v, want 15s", interval)
	}
}

func TestGetRefreshInterval_Gen1Offline(t *testing.T) {
	t.Parallel()

	c := &Cache{refreshConfig: DefaultRefreshConfig()}
	data := &DeviceData{
		Online: false,
		Info:   &shelly.DeviceInfo{Generation: 1},
	}

	interval := c.getRefreshInterval(data)
	if interval != 60*time.Second {
		t.Errorf("getRefreshInterval(Gen1 offline) = %v, want 60s", interval)
	}
}

func TestGetRefreshInterval_Gen2Online(t *testing.T) {
	t.Parallel()

	c := &Cache{refreshConfig: DefaultRefreshConfig()}
	data := &DeviceData{
		Online: true,
		Info:   &shelly.DeviceInfo{Generation: 2},
	}

	interval := c.getRefreshInterval(data)
	if interval != 5*time.Second {
		t.Errorf("getRefreshInterval(Gen2 online) = %v, want 5s", interval)
	}
}

func TestGetRefreshInterval_Gen2Offline(t *testing.T) {
	t.Parallel()

	c := &Cache{refreshConfig: DefaultRefreshConfig()}
	data := &DeviceData{
		Online: false,
		Info:   &shelly.DeviceInfo{Generation: 2},
	}

	interval := c.getRefreshInterval(data)
	if interval != 30*time.Second {
		t.Errorf("getRefreshInterval(Gen2 offline) = %v, want 30s", interval)
	}
}

func TestGetRefreshInterval_FocusedDevice(t *testing.T) {
	t.Parallel()

	c := &Cache{
		refreshConfig: DefaultRefreshConfig(),
		focusedDevice: "kitchen",
	}
	data := &DeviceData{
		Device: model.Device{Name: "kitchen"},
		Online: true,
		Info:   &shelly.DeviceInfo{Generation: 2},
	}

	interval := c.getRefreshInterval(data)
	if interval != 3*time.Second {
		t.Errorf("getRefreshInterval(focused) = %v, want 3s", interval)
	}
}

func TestGetRefreshInterval_UnknownGenOnline(t *testing.T) {
	t.Parallel()

	c := &Cache{refreshConfig: DefaultRefreshConfig()}
	data := &DeviceData{
		Online: true,
		Info:   nil, // Unknown generation
	}

	interval := c.getRefreshInterval(data)
	if interval != 5*time.Second {
		t.Errorf("getRefreshInterval(unknown gen online) = %v, want 5s (Gen2Online)", interval)
	}
}

func TestGetRefreshInterval_UnknownGenOffline(t *testing.T) {
	t.Parallel()

	c := &Cache{refreshConfig: DefaultRefreshConfig()}
	data := &DeviceData{
		Online: false,
		Info:   nil, // Unknown generation
	}

	interval := c.getRefreshInterval(data)
	if interval != 30*time.Second {
		t.Errorf("getRefreshInterval(unknown gen offline) = %v, want 30s (Gen2Offline)", interval)
	}
}

func TestSetFocusedDevice(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
	}

	// Set focused device
	_ = c.SetFocusedDevice(deviceKitchen)
	if c.focusedDevice != deviceKitchen {
		t.Errorf("focusedDevice = %q, want %q", c.focusedDevice, deviceKitchen)
	}

	// Setting same device should not trigger refresh
	_ = c.SetFocusedDevice(deviceKitchen)
	if c.focusedDevice != deviceKitchen {
		t.Errorf("focusedDevice = %q, want %q", c.focusedDevice, deviceKitchen)
	}

	// Setting different device should update
	_ = c.SetFocusedDevice(deviceOffice)
	if c.focusedDevice != deviceOffice {
		t.Errorf("focusedDevice = %q, want %q", c.focusedDevice, deviceOffice)
	}

	// Clearing focused device
	_ = c.SetFocusedDevice("")
	if c.focusedDevice != "" {
		t.Errorf("focusedDevice = %q, want ''", c.focusedDevice)
	}
}

func TestGetFocusedDevice(t *testing.T) {
	t.Parallel()

	c := &Cache{focusedDevice: deviceKitchen}

	if got := c.GetFocusedDevice(); got != deviceKitchen {
		t.Errorf("GetFocusedDevice() = %q, want %q", got, deviceKitchen)
	}
}

// === Tests for Phase 0.5.4: Stale Response Handling ===

func TestDeviceUpdateMsg_RequestID(t *testing.T) {
	t.Parallel()

	data := &DeviceData{
		Device: model.Device{Address: "192.168.1.100"},
		Online: true,
	}

	msg := DeviceUpdateMsg{
		Name:      "kitchen",
		Data:      data,
		RequestID: 42,
	}

	if msg.RequestID != 42 {
		t.Errorf("DeviceUpdateMsg.RequestID = %d, want 42", msg.RequestID)
	}
}

func TestDeviceData_LastRequestID(t *testing.T) {
	t.Parallel()

	data := DeviceData{
		Device:        model.Device{Address: "192.168.1.100"},
		lastRequestID: 123,
	}

	if data.lastRequestID != 123 {
		t.Errorf("DeviceData.lastRequestID = %d, want 123", data.lastRequestID)
	}
}

func TestStaleResponseDiscard(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		pendingCount:       1,
	}

	// Add existing device with request ID 100
	c.devices["kitchen"] = &DeviceData{
		Device:        model.Device{Address: "192.168.1.100"},
		Online:        true,
		lastRequestID: 100,
	}

	// Try to update with older request ID (stale response)
	staleMsg := DeviceUpdateMsg{
		Name: "kitchen",
		Data: &DeviceData{
			Device:        model.Device{Address: "192.168.1.100"},
			Online:        false, // Different state to detect if applied
			lastRequestID: 50,
		},
		RequestID: 50,
	}

	_ = c.Update(staleMsg)

	// Device should still be online (stale response discarded)
	if !c.devices["kitchen"].Online {
		t.Error("stale response should have been discarded, device should still be online")
	}
	if c.devices["kitchen"].lastRequestID != 100 {
		t.Errorf("lastRequestID = %d, want 100 (unchanged)", c.devices["kitchen"].lastRequestID)
	}
}

func TestNewerResponseAccepted(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		pendingCount:       1,
	}

	// Add existing device with request ID 100
	c.devices["kitchen"] = &DeviceData{
		Device:        model.Device{Address: "192.168.1.100"},
		Online:        true,
		lastRequestID: 100,
	}

	// Update with newer request ID
	newerMsg := DeviceUpdateMsg{
		Name: "kitchen",
		Data: &DeviceData{
			Device:        model.Device{Address: "192.168.1.100"},
			Online:        false, // Changed state
			lastRequestID: 150,
		},
		RequestID: 150,
	}

	_ = c.Update(newerMsg)

	// Device should now be offline (newer response accepted)
	if c.devices["kitchen"].Online {
		t.Error("newer response should have been accepted, device should be offline")
	}
	if c.devices["kitchen"].lastRequestID != 150 {
		t.Errorf("lastRequestID = %d, want 150", c.devices["kitchen"].lastRequestID)
	}
}

func TestZeroRequestIDAlwaysAccepted(t *testing.T) {
	t.Parallel()

	c := &Cache{
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		pendingCount:       1,
	}

	// Add existing device with request ID 100
	c.devices["kitchen"] = &DeviceData{
		Device:        model.Device{Address: "192.168.1.100"},
		Online:        true,
		lastRequestID: 100,
	}

	// Update with zero request ID (should be accepted for backwards compatibility)
	zeroMsg := DeviceUpdateMsg{
		Name: "kitchen",
		Data: &DeviceData{
			Device:        model.Device{Address: "192.168.1.100"},
			Online:        false,
			lastRequestID: 0,
		},
		RequestID: 0,
	}

	_ = c.Update(zeroMsg)

	// Zero request ID should be accepted
	if c.devices["kitchen"].Online {
		t.Error("zero request ID should be accepted, device should be offline")
	}
}

// === Tests for new message types ===

func TestDeviceRefreshMsg(t *testing.T) {
	t.Parallel()

	msg := DeviceRefreshMsg{Name: "kitchen"}
	if msg.Name != "kitchen" {
		t.Errorf("DeviceRefreshMsg.Name = %q, want 'kitchen'", msg.Name)
	}
}

func TestWaveMsg(t *testing.T) {
	t.Parallel()

	devices := []deviceFetch{
		{Name: "a", Device: model.Device{Address: "192.168.1.100"}},
		{Name: "b", Device: model.Device{Address: "192.168.1.101"}},
	}
	remaining := [][]deviceFetch{
		{{Name: "c", Device: model.Device{Address: "192.168.1.102"}}},
	}

	msg := WaveMsg{
		Wave:      1,
		Devices:   devices,
		Remaining: remaining,
	}

	if msg.Wave != 1 {
		t.Errorf("WaveMsg.Wave = %d, want 1", msg.Wave)
	}
	if len(msg.Devices) != 2 {
		t.Errorf("WaveMsg.Devices length = %d, want 2", len(msg.Devices))
	}
	if len(msg.Remaining) != 1 {
		t.Errorf("WaveMsg.Remaining length = %d, want 1", len(msg.Remaining))
	}
}

func TestWaveCompleteMsg(t *testing.T) {
	t.Parallel()

	msg := WaveCompleteMsg{Wave: 2}
	if msg.Wave != 2 {
		t.Errorf("WaveCompleteMsg.Wave = %d, want 2", msg.Wave)
	}
}

// === Tests for Phase 3: Connection-Aware Refresh ===

func TestCache_SetWebSocketConnected(t *testing.T) {
	t.Parallel()

	c := NewForTesting()

	// Initially not connected
	if c.IsWebSocketConnected("kitchen") {
		t.Error("expected kitchen to not be WebSocket connected initially")
	}

	// Set connected
	c.SetWebSocketConnected("kitchen", true)
	if !c.IsWebSocketConnected("kitchen") {
		t.Error("expected kitchen to be WebSocket connected after SetWebSocketConnected(true)")
	}

	// Set disconnected
	c.SetWebSocketConnected("kitchen", false)
	if c.IsWebSocketConnected("kitchen") {
		t.Error("expected kitchen to not be WebSocket connected after SetWebSocketConnected(false)")
	}
}

func TestCache_SetWebSocketConnected_MultipleDevices(t *testing.T) {
	t.Parallel()

	c := NewForTesting()

	// Connect two devices
	c.SetWebSocketConnected("kitchen", true)
	c.SetWebSocketConnected("office", true)

	if !c.IsWebSocketConnected("kitchen") {
		t.Error("expected kitchen to be WebSocket connected")
	}
	if !c.IsWebSocketConnected("office") {
		t.Error("expected office to be WebSocket connected")
	}
	if c.IsWebSocketConnected("bedroom") {
		t.Error("expected bedroom to not be WebSocket connected")
	}

	// Disconnect one
	c.SetWebSocketConnected("kitchen", false)

	if c.IsWebSocketConnected("kitchen") {
		t.Error("expected kitchen to not be WebSocket connected after disconnect")
	}
	if !c.IsWebSocketConnected("office") {
		t.Error("expected office to still be WebSocket connected")
	}
}

func TestCache_ScheduleDeviceRefresh_SkipsWebSocketDevices(t *testing.T) {
	t.Parallel()

	c := NewForTesting()
	c.SetDeviceForTesting(model.Device{Name: "kitchen", Address: "192.168.1.100", Generation: 2}, true)

	// Without WebSocket - should return a command
	cmd := c.scheduleDeviceRefresh("kitchen", c.GetDevice("kitchen"))
	if cmd == nil {
		t.Error("expected scheduleDeviceRefresh to return a command for non-WebSocket device")
	}

	// With WebSocket - should return nil
	c.SetWebSocketConnected("kitchen", true)
	cmd = c.scheduleDeviceRefresh("kitchen", c.GetDevice("kitchen"))
	if cmd != nil {
		t.Error("expected scheduleDeviceRefresh to return nil for WebSocket-connected device")
	}
}

func TestCache_IsWebSocketConnected_ThreadSafe(t *testing.T) {
	t.Parallel()

	c := NewForTesting()

	// Run concurrent reads and writes
	done := make(chan bool)
	for i := range 10 {
		go func(i int) {
			device := "device" + string(rune('0'+i))
			c.SetWebSocketConnected(device, true)
			_ = c.IsWebSocketConnected(device)
			c.SetWebSocketConnected(device, false)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// All devices should be disconnected now
	for i := range 10 {
		device := "device" + string(rune('0'+i))
		if c.IsWebSocketConnected(device) {
			t.Errorf("expected %s to be disconnected", device)
		}
	}
}
