package energybars

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New(nil)
	if m.barHeight != 1 {
		t.Errorf("barHeight = %d, want 1", m.barHeight)
	}
	if !m.showTotal {
		t.Error("showTotal should be true by default")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := New(nil).SetSize(80, 40)
	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestModel_ShowTotal(t *testing.T) {
	t.Parallel()
	m := New(nil).ShowTotal(false)
	if m.showTotal {
		t.Error("showTotal should be false")
	}
}

func TestModel_SetBars(t *testing.T) {
	t.Parallel()
	bars := []Bar{
		{Label: "Device 1", Value: 100, Unit: "W"},
		{Label: "Device 2", Value: 200, Unit: "W"},
	}
	m := New(nil).SetBars(bars)
	if m.BarCount() != 2 {
		t.Errorf("BarCount() = %d, want 2", m.BarCount())
	}
}

func TestModel_View_Empty(t *testing.T) {
	t.Parallel()
	m := New(nil).SetSize(80, 20)
	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestModel_View_WithBars(t *testing.T) {
	t.Parallel()
	bars := []Bar{
		{Label: "Device 1", Value: 100, Unit: "W", Color: theme.Orange()},
		{Label: "Device 2", Value: 200, Unit: "W", Color: theme.Orange()},
	}
	m := New(nil).SetSize(80, 20).SetBars(bars)
	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
	if len(view) < 50 {
		t.Errorf("View() too short: %d chars", len(view))
	}
}

func TestFormatValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value float64
		unit  string
		want  string
	}{
		{100, "W", "100.0 W"},
		{1500, "W", "1.50 kW"},
		{1500000, "W", "1.50 MW"},
		{-100, "W", "-100.0 W"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := formatValue(tt.value, tt.unit)
			if got != tt.want {
				t.Errorf("formatValue(%v, %q) = %q, want %q", tt.value, tt.unit, got, tt.want)
			}
		})
	}
}

func TestModel_renderBar(t *testing.T) {
	t.Parallel()
	m := New(nil).SetSize(80, 20)
	bar := Bar{Label: "Test Device", Value: 50, Unit: "W", Color: theme.Orange()}

	result := m.renderBar(bar, 100, 20, 16)
	if result == "" {
		t.Error("renderBar() returned empty string")
	}
}
