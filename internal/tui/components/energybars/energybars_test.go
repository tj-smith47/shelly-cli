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
	if m.Width != 80 {
		t.Errorf("width = %d, want 80", m.Width)
	}
	if m.Height != 40 {
		t.Errorf("height = %d, want 40", m.Height)
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
		{Label: "Device 1", Value: 100, Unit: "W", Color: theme.GetSemanticColors().Warning},
		{Label: "Device 2", Value: 200, Unit: "W", Color: theme.GetSemanticColors().Warning},
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
	bar := Bar{Label: "Test Device", Value: 50, Unit: "W", Color: theme.GetSemanticColors().Warning}

	result := m.renderBar(bar, 100, 20, 16)
	if result == "" {
		t.Error("renderBar() returned empty string")
	}
}

func TestFormatSwitchLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		deviceName  string
		switchName  string
		switchID    int
		switchCount int
		want        string
	}{
		{
			name:        "single switch returns device name only",
			deviceName:  "Office",
			switchName:  "Light",
			switchID:    0,
			switchCount: 1,
			want:        "Office",
		},
		{
			name:        "multiple switches with empty name shows Sw format",
			deviceName:  "Office",
			switchName:  "",
			switchID:    0,
			switchCount: 2,
			want:        "Office (Sw0)",
		},
		{
			name:        "multiple switches combines names",
			deviceName:  "Office",
			switchName:  "Light",
			switchID:    0,
			switchCount: 2,
			want:        "Office Light",
		},
		{
			name:        "dedupes when switch name starts with device name",
			deviceName:  "Office",
			switchName:  "Office Light",
			switchID:    0,
			switchCount: 2,
			want:        "Office Light",
		},
		{
			name:        "dedupes case insensitive",
			deviceName:  "Office",
			switchName:  "office Light",
			switchID:    0,
			switchCount: 2,
			want:        "office Light",
		},
		{
			name:        "no dedupe when device name is partial match",
			deviceName:  "Off",
			switchName:  "Office Light",
			switchID:    0,
			switchCount: 2,
			want:        "Office Light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatSwitchLabel(tt.deviceName, tt.switchName, tt.switchID, tt.switchCount)
			if got != tt.want {
				t.Errorf("formatSwitchLabel(%q, %q, %d, %d) = %q, want %q",
					tt.deviceName, tt.switchName, tt.switchID, tt.switchCount, got, tt.want)
			}
		})
	}
}
