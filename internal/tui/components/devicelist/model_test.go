package devicelist

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
)

// mockCache creates a mock cache for testing.
// It directly sets device data without network calls.
func mockCache(devices []model.Device, online map[string]bool) *cache.Cache {
	c := cache.NewForTesting()
	for _, d := range devices {
		isOnline := online[d.Name]
		c.SetDeviceForTesting(d, isOnline)
	}
	return c
}

func TestNewDeviceList(t *testing.T) {
	t.Parallel()

	c := mockCache(nil, nil)
	m := New(c)

	if m.cache != c {
		t.Error("expected cache to be set")
	}
	if m.cursor != 0 {
		t.Error("expected cursor to start at 0")
	}
	if m.filter != "" {
		t.Error("expected filter to start empty")
	}
}

func TestDeviceCount(t *testing.T) {
	t.Parallel()

	devices := []model.Device{
		{Name: "test1", Address: "10.0.0.1"},
		{Name: "test2", Address: "10.0.0.2"},
	}

	c := mockCache(devices, map[string]bool{"test1": true, "test2": true})
	m := New(c)

	// No filter - all devices
	if m.DeviceCount() != 2 {
		t.Errorf("expected 2 devices, got %d", m.DeviceCount())
	}

	if m.TotalDeviceCount() != 2 {
		t.Errorf("expected TotalDeviceCount()=2, got %d", m.TotalDeviceCount())
	}
}

func TestSetFilter(t *testing.T) {
	t.Parallel()

	devices := []model.Device{
		{Name: "kitchen-light", Address: "10.0.0.1"},
		{Name: "bedroom-fan", Address: "10.0.0.2"},
		{Name: "kitchen-outlet", Address: "10.0.0.3"},
	}

	c := mockCache(devices, map[string]bool{
		"kitchen-light":  true,
		"bedroom-fan":    true,
		"kitchen-outlet": true,
	})
	m := New(c)

	// No filter - all devices
	if m.DeviceCount() != 3 {
		t.Errorf("expected 3 devices, got %d", m.DeviceCount())
	}

	// Filter by kitchen
	m = m.SetFilter("kitchen")
	if m.DeviceCount() != 2 {
		t.Errorf("expected 2 devices with 'kitchen' filter, got %d", m.DeviceCount())
	}

	// Clear filter
	m = m.SetFilter("")
	if m.DeviceCount() != 3 {
		t.Errorf("expected 3 devices after clearing filter, got %d", m.DeviceCount())
	}
}

func TestCursorNavigation(t *testing.T) {
	t.Parallel()

	devices := []model.Device{
		{Name: "device1", Address: "10.0.0.1"},
		{Name: "device2", Address: "10.0.0.2"},
		{Name: "device3", Address: "10.0.0.3"},
	}

	c := mockCache(devices, map[string]bool{
		"device1": true,
		"device2": true,
		"device3": true,
	})
	m := New(c)
	m = m.SetSize(100, 50)

	// Start at 0
	if m.Cursor() != 0 {
		t.Errorf("expected cursor=0, got %d", m.Cursor())
	}

	// Move down
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor=1 after j, got %d", m.Cursor())
	}

	// Move down again
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 2 {
		t.Errorf("expected cursor=2 after j, got %d", m.Cursor())
	}

	// Move up
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor=1 after k, got %d", m.Cursor())
	}

	// Go to start (gg pattern)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g'})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g'})
	if m.Cursor() != 0 {
		t.Errorf("expected cursor=0 after gg, got %d", m.Cursor())
	}

	// Go to end
	m, _ = m.Update(tea.KeyPressMsg{Code: 'G'})
	if m.Cursor() != 2 {
		t.Errorf("expected cursor=2 after G, got %d", m.Cursor())
	}
}

func TestSelectedDevice(t *testing.T) {
	t.Parallel()

	devices := []model.Device{
		{Name: "device1", Address: "10.0.0.1"},
		{Name: "device2", Address: "10.0.0.2"},
	}

	c := mockCache(devices, map[string]bool{
		"device1": true,
		"device2": true,
	})
	m := New(c)

	// First device selected
	selected := m.SelectedDevice()
	if selected == nil {
		t.Fatal("expected selected device, got nil")
	}
	if selected.Device.Name != "device1" {
		t.Errorf("expected device1, got %s", selected.Device.Name)
	}

	// Move cursor and check again
	m = m.SetCursor(1)
	selected = m.SelectedDevice()
	if selected == nil {
		t.Fatal("expected selected device, got nil")
	}
	if selected.Device.Name != "device2" {
		t.Errorf("expected device2, got %s", selected.Device.Name)
	}
}

func TestSetFocused(t *testing.T) {
	t.Parallel()

	c := mockCache(nil, nil)
	m := New(c)

	// Not focused initially
	if m.Focused() {
		t.Error("expected not focused initially")
	}

	// Set focused
	m = m.SetFocused(true)
	if !m.Focused() {
		t.Error("expected focused after SetFocused(true)")
	}

	// Set unfocused
	m = m.SetFocused(false)
	if m.Focused() {
		t.Error("expected not focused after SetFocused(false)")
	}
}

func TestSetSize(t *testing.T) {
	t.Parallel()

	c := mockCache(nil, nil)
	m := New(c)

	m = m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("expected width=120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height=40, got %d", m.height)
	}
}

func TestInitReturnsNil(t *testing.T) {
	t.Parallel()

	c := mockCache(nil, nil)
	m := New(c)

	// Init should return nil since cache handles loading
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init() to return nil")
	}
}
