package monitor

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewPowerRanking(t *testing.T) {
	t.Parallel()

	m := NewPowerRanking()
	if len(m.devices) != 0 {
		t.Errorf("expected 0 devices initially, got %d", len(m.devices))
	}
	if m.focused {
		t.Error("expected not focused initially")
	}
	if m.SelectedDevice() != nil {
		t.Error("expected nil selected device initially")
	}
}

func TestPowerRankingModel_SetDevices(t *testing.T) {
	t.Parallel()

	m := NewPowerRanking()
	m = m.SetSize(80, 30)

	statuses := []DeviceStatus{
		{Name: "kitchen", Online: true, Power: 342},
		{Name: "office", Online: true, Power: 180},
		{Name: "garage", Online: true, Power: 45},
		{Name: "bedroom", Online: true, Power: 12},
		{Name: "living", Online: true, Power: 0},
		{Name: "outdoor", Online: false, Error: fmt.Errorf("timeout")},
	}

	m = m.SetDevices(statuses)

	if len(m.Devices()) != 6 {
		t.Fatalf("expected 6 devices, got %d", len(m.Devices()))
	}

	// Verify sort order: highest power first, zero power after, offline last
	expected := []string{"kitchen", "office", "garage", "bedroom", "living", "outdoor"}
	for i, d := range m.Devices() {
		if d.Name != expected[i] {
			t.Errorf("device[%d] = %q, want %q", i, d.Name, expected[i])
		}
	}
}

func TestPowerRankingModel_SortOrder(t *testing.T) {
	t.Parallel()

	m := NewPowerRanking()
	m = m.SetSize(80, 30)

	statuses := []DeviceStatus{
		{Name: "low-power", Online: true, Power: 10},
		{Name: "offline-a", Online: false},
		{Name: "zero-power", Online: true, Power: 0},
		{Name: "high-power", Online: true, Power: 500},
		{Name: "offline-b", Online: false},
		{Name: "mid-power", Online: true, Power: 100},
	}

	m = m.SetDevices(statuses)

	devices := m.Devices()
	// First: online devices with power (desc)
	if devices[0].Name != "high-power" {
		t.Errorf("expected high-power first, got %q", devices[0].Name)
	}
	if devices[1].Name != "mid-power" {
		t.Errorf("expected mid-power second, got %q", devices[1].Name)
	}
	if devices[2].Name != "low-power" {
		t.Errorf("expected low-power third, got %q", devices[2].Name)
	}
	// Then: zero power
	if devices[3].Name != "zero-power" {
		t.Errorf("expected zero-power fourth, got %q", devices[3].Name)
	}
	// Then: offline (sorted by name)
	if devices[4].Name != "offline-a" {
		t.Errorf("expected offline-a fifth, got %q", devices[4].Name)
	}
	if devices[5].Name != "offline-b" {
		t.Errorf("expected offline-b sixth, got %q", devices[5].Name)
	}
}

func TestPowerRankingModel_TrendDetection(t *testing.T) {
	t.Parallel()

	m := NewPowerRanking()
	m = m.SetSize(80, 30)

	// First reading - no trend (no previous data)
	statuses := []DeviceStatus{
		{Name: "rising", Online: true, Power: 100},
		{Name: "falling", Online: true, Power: 200},
		{Name: "stable", Online: true, Power: 50},
	}
	m = m.SetDevices(statuses)

	// All should be stable (no previous data)
	for _, d := range m.Devices() {
		if d.Trend != TrendStable {
			t.Errorf("device %q should have stable trend on first reading, got %d", d.Name, d.Trend)
		}
	}

	// Second reading - with changes
	statuses = []DeviceStatus{
		{Name: "rising", Online: true, Power: 150},  // +50
		{Name: "falling", Online: true, Power: 100}, // -100
		{Name: "stable", Online: true, Power: 50.5}, // +0.5 (within threshold)
	}
	m = m.SetDevices(statuses)

	// Find devices by name in sorted order
	deviceMap := make(map[string]RankedDevice)
	for _, d := range m.Devices() {
		deviceMap[d.Name] = d
	}

	if deviceMap["rising"].Trend != TrendRising {
		t.Errorf("expected rising trend for 'rising', got %d", deviceMap["rising"].Trend)
	}
	if deviceMap["falling"].Trend != TrendFalling {
		t.Errorf("expected falling trend for 'falling', got %d", deviceMap["falling"].Trend)
	}
	if deviceMap["stable"].Trend != TrendStable {
		t.Errorf("expected stable trend for 'stable', got %d", deviceMap["stable"].Trend)
	}
}

func TestPowerRankingModel_SelectedDevice(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for empty list", func(t *testing.T) {
		t.Parallel()
		m := NewPowerRanking()
		m = m.SetSize(80, 30)
		if m.SelectedDevice() != nil {
			t.Error("expected nil selected device for empty list")
		}
	})

	t.Run("returns device at cursor", func(t *testing.T) {
		t.Parallel()
		m := NewPowerRanking()
		m = m.SetSize(80, 30)
		m = m.SetDevices([]DeviceStatus{
			{Name: "first", Online: true, Power: 200},
			{Name: "second", Online: true, Power: 100},
		})

		sel := m.SelectedDevice()
		if sel == nil {
			t.Fatal("expected non-nil selected device")
		}
		// First device should be "first" (highest power)
		if sel.Name != "first" {
			t.Errorf("expected 'first', got %q", sel.Name)
		}

		// Move cursor down
		m.Scroller.CursorDown()
		sel = m.SelectedDevice()
		if sel == nil {
			t.Fatal("expected non-nil selected device after cursor down")
		}
		if sel.Name != "second" {
			t.Errorf("expected 'second', got %q", sel.Name)
		}
	})
}

func TestPowerRankingModel_SetFocused(t *testing.T) {
	t.Parallel()

	m := NewPowerRanking()
	if m.IsFocused() {
		t.Error("expected not focused initially")
	}

	m = m.SetFocused(true)
	if !m.IsFocused() {
		t.Error("expected focused after SetFocused(true)")
	}

	m = m.SetFocused(false)
	if m.IsFocused() {
		t.Error("expected not focused after SetFocused(false)")
	}
}

func TestPowerRankingModel_SetSize(t *testing.T) {
	t.Parallel()

	m := NewPowerRanking()
	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("expected width 100, got %d", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("expected height 50, got %d", m.Height)
	}
}

func TestPowerRankingModel_View(t *testing.T) {
	t.Parallel()

	t.Run("too small returns empty", func(t *testing.T) {
		t.Parallel()
		m := NewPowerRanking()
		m = m.SetSize(5, 2)
		if m.View() != "" {
			t.Error("expected empty view for tiny dimensions")
		}
	})

	t.Run("renders with devices", func(t *testing.T) {
		t.Parallel()
		m := NewPowerRanking()
		m = m.SetSize(80, 20)
		m = m.SetDevices([]DeviceStatus{
			{Name: "kitchen", Online: true, Power: 342},
			{Name: "outdoor", Online: false, Error: fmt.Errorf("timeout")},
		})

		view := m.View()
		if !strings.Contains(view, "Power Ranking") {
			t.Error("expected view to contain title")
		}
	})

	t.Run("empty list shows message", func(t *testing.T) {
		t.Parallel()
		m := NewPowerRanking()
		m = m.SetSize(80, 20)
		m = m.SetDevices([]DeviceStatus{})

		view := m.View()
		if !strings.Contains(view, "No devices") {
			t.Error("expected empty message")
		}
	})
}

func TestTrendIndicator(t *testing.T) {
	t.Parallel()

	m := NewPowerRanking()

	// Trend indicators should contain the expected arrows
	up := m.trendIndicator(TrendRising)
	if !strings.Contains(up, "▲") {
		t.Errorf("expected rising trend to contain ▲, got %q", up)
	}

	down := m.trendIndicator(TrendFalling)
	if !strings.Contains(down, "▼") {
		t.Errorf("expected falling trend to contain ▼, got %q", down)
	}

	flat := m.trendIndicator(TrendStable)
	if !strings.Contains(flat, "─") {
		t.Errorf("expected stable trend to contain ─, got %q", flat)
	}
}
