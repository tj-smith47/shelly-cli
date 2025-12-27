package energyhistory

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New(nil)
	if m.maxItems != 60 {
		t.Errorf("maxItems = %d, want 60", m.maxItems)
	}
	if m.history == nil {
		t.Error("history map should be initialized")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := New(nil)
	m = m.SetSize(80, 40)
	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestModel_addDataPoint(t *testing.T) {
	t.Parallel()
	m := New(nil)
	m.addDataPoint("device1", 100.5)
	m.addDataPoint("device1", 150.0)
	m.addDataPoint("device2", 75.0)

	if m.DeviceCount() != 2 {
		t.Errorf("DeviceCount() = %d, want 2", m.DeviceCount())
	}
	if m.HistoryCount() != 3 {
		t.Errorf("HistoryCount() = %d, want 3", m.HistoryCount())
	}
}

func TestModel_addDataPoint_MaxItems(t *testing.T) {
	t.Parallel()
	m := New(nil)
	m.maxItems = 5 // Set small limit for testing

	// Add more than max items
	for i := range 10 {
		m.addDataPoint("device1", float64(i*10))
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history["device1"]) != 5 {
		t.Errorf("history length = %d, want 5", len(m.history["device1"]))
	}

	// Should keep the last 5 values (50, 60, 70, 80, 90)
	if m.history["device1"][0].Value != 50 {
		t.Errorf("first value = %f, want 50", m.history["device1"][0].Value)
	}
	if m.history["device1"][4].Value != 90 {
		t.Errorf("last value = %f, want 90", m.history["device1"][4].Value)
	}
}

func TestModel_View_Empty(t *testing.T) {
	t.Parallel()
	m := New(nil)
	m = m.SetSize(80, 20)
	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestModel_Clear(t *testing.T) {
	t.Parallel()
	m := New(nil)
	m.addDataPoint("device1", 100.0)
	m.addDataPoint("device2", 200.0)

	if m.DeviceCount() != 2 {
		t.Errorf("DeviceCount() before clear = %d, want 2", m.DeviceCount())
	}

	m.Clear()

	if m.DeviceCount() != 0 {
		t.Errorf("DeviceCount() after clear = %d, want 0", m.DeviceCount())
	}
}

func TestGenerateSparkline(t *testing.T) {
	t.Parallel()
	m := New(nil)

	tests := []struct {
		name    string
		history []DataPoint
		width   int
		wantLen int
	}{
		{
			name:    "empty history",
			history: []DataPoint{},
			width:   10,
			wantLen: 10, // Should be all spaces
		},
		{
			name: "single point",
			history: []DataPoint{
				{Value: 100, Timestamp: time.Now()},
			},
			width:   10,
			wantLen: 10, // Padded with spaces
		},
		{
			name: "multiple points",
			history: []DataPoint{
				{Value: 0, Timestamp: time.Now()},
				{Value: 50, Timestamp: time.Now()},
				{Value: 100, Timestamp: time.Now()},
			},
			width:   10,
			wantLen: 10,
		},
		{
			name: "more points than width",
			history: []DataPoint{
				{Value: 10, Timestamp: time.Now()},
				{Value: 20, Timestamp: time.Now()},
				{Value: 30, Timestamp: time.Now()},
				{Value: 40, Timestamp: time.Now()},
				{Value: 50, Timestamp: time.Now()},
			},
			width:   3,
			wantLen: 3, // Takes last 3 points
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := m.generateSparkline(tt.history, tt.width)
			// Note: We count runes, not bytes, since sparkline uses Unicode
			runeCount := 0
			for range result {
				runeCount++
			}
			if runeCount != tt.wantLen {
				t.Errorf("sparkline rune count = %d, want %d", runeCount, tt.wantLen)
			}
		})
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
		{999, "W", "999.0 W"},
		{-100, "W", "-100.0 W"},
		{0, "W", "0.0 W"},
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

func TestDataPoint(t *testing.T) {
	t.Parallel()
	now := time.Now()
	dp := DataPoint{
		Value:     123.45,
		Timestamp: now,
	}

	if dp.Value != 123.45 {
		t.Errorf("Value = %f, want 123.45", dp.Value)
	}
	if !dp.Timestamp.Equal(now) {
		t.Errorf("Timestamp mismatch")
	}
}
