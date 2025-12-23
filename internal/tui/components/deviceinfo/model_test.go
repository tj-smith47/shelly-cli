package deviceinfo

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()
	if m.componentCursor != -1 {
		t.Errorf("componentCursor = %d, want -1", m.componentCursor)
	}
	if m.device != nil {
		t.Error("device should be nil initially")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := New().SetSize(80, 40)
	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := New().SetFocused(true)
	if !m.Focused() {
		t.Error("Focused() should be true")
	}

	m = m.SetFocused(false)
	if m.Focused() {
		t.Error("Focused() should be false")
	}
}

func TestModel_SetDevice(t *testing.T) {
	t.Parallel()
	m := New()
	data := &cache.DeviceData{
		Device: model.Device{Name: "test-device"},
		Online: true,
	}
	m = m.SetDevice(data)

	if m.Device() == nil {
		t.Error("Device() should not be nil")
	}
	if m.Device().Device.Name != "test-device" {
		t.Errorf("Device().Device.Name = %q, want %q", m.Device().Device.Name, "test-device")
	}
}

func TestModel_View_Empty(t *testing.T) {
	t.Parallel()
	m := New().SetSize(80, 20)
	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestModel_View_WithDevice(t *testing.T) {
	t.Parallel()
	m := New().SetSize(80, 30)
	data := &cache.DeviceData{
		Device:  model.Device{Name: "Kitchen Plug", Model: "SNSW-102"},
		Online:  true,
		Power:   150.5,
		Voltage: 120.0,
		Current: 1.25,
		Info: &shelly.DeviceInfo{
			Firmware: "1.0.8",
			App:      "plug",
		},
		Switches: []cache.SwitchState{
			{ID: 0, On: true},
		},
	}
	m = m.SetDevice(data)

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestModel_Update_Navigation(t *testing.T) {
	t.Parallel()
	m := New().SetSize(80, 30).SetFocused(true)
	data := &cache.DeviceData{
		Device: model.Device{Name: "test"},
		Online: true,
		Switches: []cache.SwitchState{
			{ID: 0, On: true},
			{ID: 1, On: false},
		},
	}
	m = m.SetDevice(data)

	// Test navigation key 'l' to select first component
	msg := tea.KeyPressMsg{Code: 108} // 'l'
	m, _ = m.Update(msg)
	if m.SelectedComponent() != 0 {
		t.Errorf("SelectedComponent() = %d, want 0", m.SelectedComponent())
	}

	// Test 'a' to toggle back to all
	msg = tea.KeyPressMsg{Code: 97} // 'a'
	m, _ = m.Update(msg)
	if m.SelectedComponent() != -1 {
		t.Errorf("SelectedComponent() = %d, want -1 (all)", m.SelectedComponent())
	}
}

func TestModel_SelectedEndpoint(t *testing.T) {
	t.Parallel()
	m := New().SetSize(80, 30).SetFocused(true)
	data := &cache.DeviceData{
		Device: model.Device{Name: "test"},
		Online: true,
		Switches: []cache.SwitchState{
			{ID: 0, On: true},
		},
	}
	m = m.SetDevice(data)

	// No component selected initially
	if m.SelectedEndpoint() != "" {
		t.Errorf("SelectedEndpoint() = %q, want empty string", m.SelectedEndpoint())
	}

	// Select first component
	m.componentCursor = 0
	endpoint := m.SelectedEndpoint()
	if endpoint != "Switch.GetStatus?id=0" {
		t.Errorf("SelectedEndpoint() = %q, want %q", endpoint, "Switch.GetStatus?id=0")
	}
}

func TestModel_getComponents(t *testing.T) {
	t.Parallel()
	m := New()

	// No device
	comps := m.getComponents()
	if len(comps) != 0 {
		t.Errorf("getComponents() length = %d, want 0", len(comps))
	}

	// Offline device
	m = m.SetDevice(&cache.DeviceData{
		Device: model.Device{Name: "offline"},
		Online: false,
	})
	comps = m.getComponents()
	if len(comps) != 0 {
		t.Errorf("getComponents() for offline device length = %d, want 0", len(comps))
	}

	// Online device with switches
	m = m.SetDevice(&cache.DeviceData{
		Device: model.Device{Name: "test"},
		Online: true,
		Switches: []cache.SwitchState{
			{ID: 0, On: true},
			{ID: 1, On: false},
		},
	})
	comps = m.getComponents()
	if len(comps) != 2 {
		t.Errorf("getComponents() length = %d, want 2", len(comps))
	}
}

func TestFormatPower(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value float64
		want  string
	}{
		{100, "100.0 W"},
		{1500, "1.50 kW"},
		{999, "999.0 W"},
		{-100, "-100.0 W"},
		{0, "0.0 W"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := formatPower(tt.value)
			if got != tt.want {
				t.Errorf("formatPower(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestComponentInfo(t *testing.T) {
	t.Parallel()
	power := 150.5
	comp := ComponentInfo{
		Name:     "Switch:0",
		Type:     "Switch",
		State:    "on",
		Power:    &power,
		Endpoint: "Switch.GetStatus?id=0",
	}

	if comp.Name != "Switch:0" {
		t.Errorf("Name = %q, want %q", comp.Name, "Switch:0")
	}
	if *comp.Power != 150.5 {
		t.Errorf("Power = %v, want 150.5", *comp.Power)
	}
}
