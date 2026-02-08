package devicelist

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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
	if m.Cursor() != 0 {
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

	// Move down via NavigationMsg (how app.go dispatches j/k)
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor=1 after NavDown, got %d", m.Cursor())
	}

	// Move down again
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if m.Cursor() != 2 {
		t.Errorf("expected cursor=2 after NavDown, got %d", m.Cursor())
	}

	// Move up
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor=1 after NavUp, got %d", m.Cursor())
	}

	// Go to start via NavHome
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavHome})
	if m.Cursor() != 0 {
		t.Errorf("expected cursor=0 after NavHome, got %d", m.Cursor())
	}

	// Go to end via NavEnd
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavEnd})
	if m.Cursor() != 2 {
		t.Errorf("expected cursor=2 after NavEnd, got %d", m.Cursor())
	}

	// Test gg pattern via raw key (component-specific behavior retained)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g'})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g'})
	if m.Cursor() != 0 {
		t.Errorf("expected cursor=0 after gg, got %d", m.Cursor())
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

	if m.Width != 120 {
		t.Errorf("expected width=120, got %d", m.Width)
	}
	if m.Height != 40 {
		t.Errorf("expected height=40, got %d", m.Height)
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

func TestRenderListRow_PlatformBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		platform    string
		wantBadge   bool
		wantBadgeCh string
	}{
		{"no platform (native shelly)", "", false, ""},
		{"explicit shelly platform", "shelly", false, ""},
		{"tasmota platform", "tasmota", true, "T"},
		{"esphome platform", "esphome", true, "E"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := &cache.DeviceData{
				Device:  model.Device{Name: "test-device", Address: "10.0.0.1", Platform: tt.platform},
				Online:  true,
				Fetched: true,
			}

			m := New(cache.NewForTesting())
			m = m.SetSize(80, 40)
			row := m.renderListRow(d, false, 40)

			if tt.wantBadge {
				if !strings.Contains(row, "["+tt.wantBadgeCh+"]") {
					t.Errorf("expected badge [%s] in row, got: %s", tt.wantBadgeCh, row)
				}
			} else {
				// Check that no platform badge like [T] or [E] appears
				// (can't just check for "[" because lipgloss ANSI codes contain brackets)
				for _, ch := range []string{"[T]", "[E]", "[S]"} {
					if strings.Contains(row, ch) {
						t.Errorf("expected no platform badge in row for platform %q, found %s, got: %s", tt.platform, ch, row)
					}
				}
			}
		})
	}
}

func TestRenderBasicInfo_PlatformRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
		want     bool
	}{
		{"native shelly shows no platform", "", false},
		{"explicit shelly shows no platform", "shelly", false},
		{"tasmota shows platform", "tasmota", true},
		{"esphome shows platform", "esphome", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := &cache.DeviceData{
				Device: model.Device{
					Name:     "test-device",
					Address:  "10.0.0.1",
					Platform: tt.platform,
				},
			}

			m := New(cache.NewForTesting())
			var content strings.Builder
			m.renderBasicInfo(&content, d)
			result := content.String()

			switch {
			case tt.want && !strings.Contains(result, "Platform"):
				t.Errorf("expected Platform row for platform %q, got: %s", tt.platform, result)
			case tt.want && !strings.Contains(result, "shelly-"+tt.platform):
				t.Errorf("expected plugin name shelly-%s in output, got: %s", tt.platform, result)
			case !tt.want && strings.Contains(result, "Platform"):
				t.Errorf("expected no Platform row for platform %q, got: %s", tt.platform, result)
			}
		})
	}
}

func TestPlatformFilter_Cycling(t *testing.T) {
	t.Parallel()

	devices := []model.Device{
		{Name: "shelly1", Address: "10.0.0.1"},                       // default platform (shelly)
		{Name: "shelly2", Address: "10.0.0.2", Platform: "shelly"},   // explicit shelly
		{Name: "tasmota1", Address: "10.0.0.3", Platform: "tasmota"}, // tasmota
		{Name: "esphome1", Address: "10.0.0.4", Platform: "esphome"}, // esphome
		{Name: "tasmota2", Address: "10.0.0.5", Platform: "tasmota"}, // second tasmota
	}

	c := mockCache(devices, map[string]bool{
		"shelly1": true, "shelly2": true,
		"tasmota1": true, "esphome1": true, "tasmota2": true,
	})
	m := New(c)
	m = m.SetSize(80, 40)

	// Initially no filter
	if m.PlatformFilter() != "" {
		t.Errorf("expected empty platform filter, got %q", m.PlatformFilter())
	}
	if m.DeviceCount() != 5 {
		t.Errorf("expected 5 devices, got %d", m.DeviceCount())
	}

	// Cycle: "" → "esphome" (alphabetically first)
	m, _ = m.Update(messages.PlatformFilterMsg{})
	if m.PlatformFilter() != "esphome" {
		t.Errorf("expected platform filter 'esphome', got %q", m.PlatformFilter())
	}
	if m.DeviceCount() != 1 {
		t.Errorf("expected 1 esphome device, got %d", m.DeviceCount())
	}

	// Cycle: "esphome" → "shelly"
	m, _ = m.Update(messages.PlatformFilterMsg{})
	if m.PlatformFilter() != "shelly" {
		t.Errorf("expected platform filter 'shelly', got %q", m.PlatformFilter())
	}
	if m.DeviceCount() != 2 {
		t.Errorf("expected 2 shelly devices, got %d", m.DeviceCount())
	}

	// Cycle: "shelly" → "tasmota"
	m, _ = m.Update(messages.PlatformFilterMsg{})
	if m.PlatformFilter() != "tasmota" {
		t.Errorf("expected platform filter 'tasmota', got %q", m.PlatformFilter())
	}
	if m.DeviceCount() != 2 {
		t.Errorf("expected 2 tasmota devices, got %d", m.DeviceCount())
	}

	// Cycle: "tasmota" → "" (back to all)
	m, _ = m.Update(messages.PlatformFilterMsg{})
	if m.PlatformFilter() != "" {
		t.Errorf("expected empty platform filter, got %q", m.PlatformFilter())
	}
	if m.DeviceCount() != 5 {
		t.Errorf("expected 5 devices after clearing filter, got %d", m.DeviceCount())
	}
}

func TestPlatformFilter_CombinedWithTextFilter(t *testing.T) {
	t.Parallel()

	devices := []model.Device{
		{Name: "kitchen-shelly", Address: "10.0.0.1"},
		{Name: "kitchen-tasmota", Address: "10.0.0.2", Platform: "tasmota"},
		{Name: "bedroom-tasmota", Address: "10.0.0.3", Platform: "tasmota"},
	}

	c := mockCache(devices, map[string]bool{
		"kitchen-shelly": true, "kitchen-tasmota": true, "bedroom-tasmota": true,
	})
	m := New(c)
	m = m.SetSize(80, 40)

	// Set text filter for "kitchen"
	m = m.SetFilter("kitchen")
	if m.DeviceCount() != 2 {
		t.Errorf("expected 2 kitchen devices, got %d", m.DeviceCount())
	}

	// Add platform filter for "tasmota"
	m, _ = m.Update(messages.PlatformFilterMsg{}) // → "shelly"
	m, _ = m.Update(messages.PlatformFilterMsg{}) // → "tasmota"
	if m.DeviceCount() != 1 {
		t.Errorf("expected 1 kitchen+tasmota device, got %d", m.DeviceCount())
	}
}
