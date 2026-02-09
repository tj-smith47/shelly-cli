package monitor

import (
	"strings"
	"testing"
)

func TestNewSummary(t *testing.T) {
	t.Parallel()

	m := NewSummary()
	if m.width != 0 {
		t.Errorf("expected initial width 0, got %d", m.width)
	}
	if m.height != 3 {
		t.Errorf("expected initial height 3, got %d", m.height)
	}
	if m.IsRefreshing() {
		t.Error("expected not refreshing initially")
	}
}

func TestSummaryModel_SetData(t *testing.T) {
	t.Parallel()

	m := NewSummary()
	data := SummaryData{
		TotalPower:  500.0,
		TotalEnergy: 12400.0,
		OnlineCount: 8,
		TotalCount:  10,
		CostRate:    0.12,
		Currency:    "$",
	}

	m = m.SetData(data)

	if m.Data().TotalPower != 500.0 {
		t.Errorf("expected total power 500, got %f", m.Data().TotalPower)
	}
	if m.Data().TotalEnergy != 12400.0 {
		t.Errorf("expected total energy 12400, got %f", m.Data().TotalEnergy)
	}
	if m.Data().OnlineCount != 8 {
		t.Errorf("expected online count 8, got %d", m.Data().OnlineCount)
	}
	if m.Data().TotalCount != 10 {
		t.Errorf("expected total count 10, got %d", m.Data().TotalCount)
	}
}

func TestSummaryModel_PeakPowerTracking(t *testing.T) {
	t.Parallel()

	m := NewSummary()

	// Set initial data with high power
	m = m.SetData(SummaryData{TotalPower: 500.0})
	if m.Data().PeakPower != 500.0 {
		t.Errorf("expected peak power 500, got %f", m.Data().PeakPower)
	}

	// Lower power should not lower peak
	m = m.SetData(SummaryData{TotalPower: 300.0})
	if m.Data().PeakPower != 500.0 {
		t.Errorf("expected peak power to stay at 500, got %f", m.Data().PeakPower)
	}

	// Higher power should update peak
	m = m.SetData(SummaryData{TotalPower: 700.0})
	if m.Data().PeakPower != 700.0 {
		t.Errorf("expected peak power to update to 700, got %f", m.Data().PeakPower)
	}
}

func TestSummaryModel_SetSize(t *testing.T) {
	t.Parallel()

	m := NewSummary()
	m = m.SetSize(120, 3)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
}

func TestSummaryModel_View(t *testing.T) {
	t.Parallel()

	t.Run("too narrow returns empty", func(t *testing.T) {
		t.Parallel()
		m := NewSummary()
		m = m.SetSize(15, 3)
		if m.View() != "" {
			t.Error("expected empty view for narrow width")
		}
	})

	t.Run("renders with data", func(t *testing.T) {
		t.Parallel()
		m := NewSummary()
		m = m.SetSize(120, 5)
		m = m.SetData(SummaryData{
			TotalPower:  500.0,
			TotalEnergy: 12400.0,
			OnlineCount: 8,
			TotalCount:  10,
			CostRate:    0.12,
			Currency:    "$",
		})

		view := m.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
		// Should contain key information
		if !strings.Contains(view, "Monitor") {
			t.Error("expected view to contain 'Monitor' title")
		}
	})
}

func TestSummaryModel_Refresh(t *testing.T) {
	t.Parallel()

	m := NewSummary()

	// Start refresh
	var cmd interface{}
	m, _ = m.StartRefresh()
	_ = cmd
	if !m.IsRefreshing() {
		t.Error("expected refreshing after StartRefresh")
	}

	// Stop refresh
	m = m.StopRefresh()
	if m.IsRefreshing() {
		t.Error("expected not refreshing after StopRefresh")
	}
	if m.UpdatedAt().IsZero() {
		t.Error("expected non-zero UpdatedAt after StopRefresh")
	}
}

func TestFormatPower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		watts float64
		want  string
	}{
		{0, "0.0 W"},
		{123.4, "123.4 W"},
		{999.9, "999.9 W"},
		{1000, "1.00 kW"},
		{1500.5, "1.50 kW"},
		{-1500, "-1.50 kW"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := formatPower(tt.watts)
			if got != tt.want {
				t.Errorf("formatPower(%f) = %q, want %q", tt.watts, got, tt.want)
			}
		})
	}
}

func TestFormatEnergy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wh   float64
		want string
	}{
		{0, "0.0 Wh"},
		{500, "500.0 Wh"},
		{999.9, "999.9 Wh"},
		{1000, "1.00 kWh"},
		{12400, "12.40 kWh"},
		{1000000, "1.00 MWh"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := formatEnergy(tt.wh)
			if got != tt.want {
				t.Errorf("formatEnergy(%f) = %q, want %q", tt.wh, got, tt.want)
			}
		})
	}
}
