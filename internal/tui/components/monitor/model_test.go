package monitor

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// createTestModel creates a model with test statuses for navigation testing.
func createTestModel(statusCount int) Model {
	m := Model{
		Sizable:  panel.NewSizable(11, panel.NewScroller(0, 10)),
		statuses: make([]DeviceStatus, statusCount),
		styles:   DefaultStyles(),
	}
	m = m.SetSize(100, 100)
	for i := range statusCount {
		m.statuses[i] = DeviceStatus{
			Name:   "test-device",
			Online: true,
		}
	}
	m.Scroller.SetItemCount(statusCount)
	return m
}

func TestScrollerCursorDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor down", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m.Scroller.CursorDown()

		if m.Cursor() != 1 {
			t.Errorf("expected cursor 1, got %d", m.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.Scroller.SetCursor(4)

		m.Scroller.CursorDown()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor to stay at 4, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.CursorDown()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerCursorUp(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor up", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.Scroller.SetCursor(5)

		m.Scroller.CursorUp()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)

		m.Scroller.CursorUp()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerCursorToEnd(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor to last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m.Scroller.CursorToEnd()

		if m.Cursor() != 9 {
			t.Errorf("expected cursor 9, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.CursorToEnd()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerPageDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor by visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m = m.SetSize(100, 30)

		m.Scroller.PageDown()

		if m.Cursor() <= 0 {
			t.Errorf("expected cursor to move forward, got %d", m.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.Scroller.SetCursor(3)

		m.Scroller.PageDown()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.PageDown()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerPageUp(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor backward", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m = m.SetSize(100, 30)
		m.Scroller.SetCursor(20)

		m.Scroller.PageUp()

		if m.Cursor() >= 20 {
			t.Errorf("expected cursor to move backward from 20, got %d", m.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.Scroller.SetCursor(2)

		m.Scroller.PageUp()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor 0, got %d", m.Cursor())
		}
	})
}

func TestSelectedDevice(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		if m.SelectedDevice() != nil {
			t.Error("expected nil for empty list")
		}
	})

	t.Run("returns selected device", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.statuses[2].Name = "selected-device"
		m.Scroller.SetCursor(2)

		selected := m.SelectedDevice()
		if selected == nil {
			t.Fatal("expected non-nil selected device")
		}
		if selected.Name != "selected-device" {
			t.Errorf("expected selected-device, got %s", selected.Name)
		}
	})
}

func TestSetSize(t *testing.T) {
	t.Parallel()

	m := createTestModel(0)
	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("expected width 100, got %d", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("expected height 50, got %d", m.Height)
	}
}

func TestAggregateMetrics_PM(t *testing.T) {
	t.Parallel()

	freq := 50.0
	energy := &model.PMEnergyCounters{Total: 1234.5}

	pms := []model.PMStatus{
		{APower: 100, Voltage: 230, Current: 0.43, Freq: &freq, AEnergy: energy},
		{APower: 50, Voltage: 231, Current: 0.22},
	}

	status := &DeviceStatus{}
	aggregateMetrics(status, pms, false)

	if status.Power != 150 {
		t.Errorf("expected power 150, got %f", status.Power)
	}
	if status.Voltage != 230 {
		t.Errorf("expected voltage 230 (first non-zero), got %f", status.Voltage)
	}
	if status.Current != 0.43 {
		t.Errorf("expected current 0.43 (first non-zero), got %f", status.Current)
	}
	if status.Frequency != 50 {
		t.Errorf("expected frequency 50, got %f", status.Frequency)
	}
	if status.TotalEnergy != 1234.5 {
		t.Errorf("expected total energy 1234.5, got %f", status.TotalEnergy)
	}
}

func TestAggregateMetrics_EM(t *testing.T) {
	t.Parallel()

	freq := 60.0
	ems := []model.EMStatus{
		{TotalActivePower: 500, AVoltage: 120, TotalCurrent: 4.2, AFreq: &freq},
		{TotalActivePower: 300, AVoltage: 121, TotalCurrent: 2.5},
	}

	status := &DeviceStatus{}
	aggregateMetrics(status, ems, true)

	if status.Power != 800 {
		t.Errorf("expected power 800, got %f", status.Power)
	}
	if status.Voltage != 120 {
		t.Errorf("expected voltage 120 (first non-zero), got %f", status.Voltage)
	}
	// EM accumulates current
	if status.Current != 6.7 {
		t.Errorf("expected current 6.7 (accumulated), got %f", status.Current)
	}
	if status.Frequency != 60 {
		t.Errorf("expected frequency 60, got %f", status.Frequency)
	}
	if status.TotalEnergy != 0 {
		t.Errorf("expected total energy 0 (EM has no energy), got %f", status.TotalEnergy)
	}
}

func TestAggregateMetrics_EM1(t *testing.T) {
	t.Parallel()

	freq := 50.0
	em1s := []model.EM1Status{
		{ActPower: 200, Voltage: 240, Current: 0.83, Freq: &freq},
		{ActPower: 100, Voltage: 241, Current: 0.42},
	}

	status := &DeviceStatus{}
	aggregateMetrics(status, em1s, false)

	if status.Power != 300 {
		t.Errorf("expected power 300, got %f", status.Power)
	}
	if status.Voltage != 240 {
		t.Errorf("expected voltage 240 (first non-zero), got %f", status.Voltage)
	}
	if status.Current != 0.83 {
		t.Errorf("expected current 0.83 (first non-zero), got %f", status.Current)
	}
	if status.Frequency != 50 {
		t.Errorf("expected frequency 50, got %f", status.Frequency)
	}
}

func TestAggregateMetrics_EmptySlice(t *testing.T) {
	t.Parallel()

	status := &DeviceStatus{}
	aggregateMetrics(status, []model.PMStatus{}, false)

	if status.Power != 0 {
		t.Errorf("expected power 0, got %f", status.Power)
	}
	if status.Voltage != 0 {
		t.Errorf("expected voltage 0, got %f", status.Voltage)
	}
}

func TestAggregateMetrics_MultipleTypes(t *testing.T) {
	t.Parallel()

	// Test aggregating across all three types into same status (like checkDeviceStatus does)
	freq := 50.0
	energy := &model.PMEnergyCounters{Total: 500}

	status := &DeviceStatus{}
	aggregateMetrics(status, []model.PMStatus{
		{APower: 100, Voltage: 230, Current: 0.43, Freq: &freq, AEnergy: energy},
	}, false)
	aggregateMetrics(status, []model.EMStatus{
		{TotalActivePower: 200, AVoltage: 231, TotalCurrent: 1.5},
	}, true)
	aggregateMetrics(status, []model.EM1Status{
		{ActPower: 50, Voltage: 232, Current: 0.21},
	}, false)

	// Power accumulated from all
	if status.Power != 350 {
		t.Errorf("expected power 350, got %f", status.Power)
	}
	// Voltage set from first PM (first non-zero)
	if status.Voltage != 230 {
		t.Errorf("expected voltage 230, got %f", status.Voltage)
	}
	// Current: PM sets 0.43, EM accumulates +1.5, EM1 skipped (already non-zero, not accumulate)
	expectedCurrent := 0.43 + 1.5
	if status.Current != expectedCurrent {
		t.Errorf("expected current %f, got %f", expectedCurrent, status.Current)
	}
	// Frequency from PM
	if status.Frequency != 50 {
		t.Errorf("expected frequency 50, got %f", status.Frequency)
	}
	// Energy from PM only
	if status.TotalEnergy != 500 {
		t.Errorf("expected total energy 500, got %f", status.TotalEnergy)
	}
}
