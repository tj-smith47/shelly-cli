package monitor

import (
	"testing"
)

// createTestModel creates a model with test statuses for navigation testing.
func createTestModel(statusCount int) Model {
	m := Model{
		statuses: make([]DeviceStatus, statusCount),
		height:   100,
		styles:   DefaultStyles(),
	}
	for i := range statusCount {
		m.statuses[i] = DeviceStatus{
			Name:   "test-device",
			Online: true,
		}
	}
	return m
}

func TestCursorDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor down", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.cursor = 0

		m = m.cursorDown()

		if m.cursor != 1 {
			t.Errorf("expected cursor 1, got %d", m.cursor)
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.cursor = 4

		m = m.cursorDown()

		if m.cursor != 4 {
			t.Errorf("expected cursor to stay at 4, got %d", m.cursor)
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.cursor = 0

		m = m.cursorDown()

		if m.cursor != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
		}
	})

	t.Run("adjusts scroll when cursor below visible", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(20)
		m.height = 20
		m.cursor = m.visibleRows() - 1
		m.scrollOffset = 0

		m = m.cursorDown()

		if m.scrollOffset == 0 {
			t.Error("expected scrollOffset to increase when cursor moves below visible area")
		}
	})
}

func TestCursorUp(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor up", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.cursor = 5

		m = m.cursorUp()

		if m.cursor != 4 {
			t.Errorf("expected cursor 4, got %d", m.cursor)
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.cursor = 0

		m = m.cursorUp()

		if m.cursor != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
		}
	})

	t.Run("adjusts scroll when cursor above visible", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(20)
		m.cursor = 5
		m.scrollOffset = 5

		m = m.cursorUp()

		if m.scrollOffset != 4 {
			t.Errorf("expected scrollOffset 4, got %d", m.scrollOffset)
		}
	})
}

func TestCursorToEnd(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor to last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.cursor = 0

		m = m.cursorToEnd()

		if m.cursor != 9 {
			t.Errorf("expected cursor 9, got %d", m.cursor)
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m = m.cursorToEnd()

		if m.cursor != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
		}
	})

	t.Run("adjusts scrollOffset", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(20)
		m.height = 20
		m.scrollOffset = 0

		m = m.cursorToEnd()

		if m.scrollOffset <= 0 {
			t.Error("expected scrollOffset to be positive when jumping to end of long list")
		}
	})
}

func TestPageDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor by visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m.height = 30
		m.cursor = 0
		visibleRows := m.visibleRows()

		m = m.pageDown()

		if m.cursor != visibleRows {
			t.Errorf("expected cursor %d, got %d", visibleRows, m.cursor)
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.cursor = 3

		m = m.pageDown()

		if m.cursor != 4 {
			t.Errorf("expected cursor 4, got %d", m.cursor)
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m = m.pageDown()

		if m.cursor != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
		}
	})
}

func TestPageUp(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor by visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m.height = 30
		visibleRows := m.visibleRows()
		m.cursor = visibleRows * 2

		m = m.pageUp()

		if m.cursor != visibleRows {
			t.Errorf("expected cursor %d, got %d", visibleRows, m.cursor)
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.cursor = 2

		m = m.pageUp()

		if m.cursor != 0 {
			t.Errorf("expected cursor 0, got %d", m.cursor)
		}
	})
}

func TestVisibleRows(t *testing.T) {
	t.Parallel()

	t.Run("returns minimum for small height", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.height = 5

		if m.visibleRows() < 1 {
			t.Error("expected at least 1 visible row")
		}
	})

	t.Run("calculates based on height", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.height = 100

		rows := m.visibleRows()
		if rows <= 0 {
			t.Error("expected positive visible rows for height 100")
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
		m.cursor = 2

		selected := m.SelectedDevice()
		if selected == nil {
			t.Fatal("expected non-nil selected device")
		}
		if selected.Name != "selected-device" {
			t.Errorf("expected selected-device, got %s", selected.Name)
		}
	})

	t.Run("returns nil for out of bounds cursor", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.cursor = 10

		if m.SelectedDevice() != nil {
			t.Error("expected nil for out of bounds cursor")
		}
	})
}

func TestSetSize(t *testing.T) {
	t.Parallel()

	m := createTestModel(0)
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}
