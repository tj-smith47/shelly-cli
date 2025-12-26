package monitor

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// createTestModel creates a model with test statuses for navigation testing.
func createTestModel(statusCount int) Model {
	m := Model{
		statuses: make([]DeviceStatus, statusCount),
		scroller: panel.NewScroller(0, 10),
		height:   100,
		styles:   DefaultStyles(),
	}
	for i := range statusCount {
		m.statuses[i] = DeviceStatus{
			Name:   "test-device",
			Online: true,
		}
	}
	m.scroller.SetItemCount(statusCount)
	return m
}

func TestScrollerCursorDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor down", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m.scroller.CursorDown()

		if m.Cursor() != 1 {
			t.Errorf("expected cursor 1, got %d", m.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.scroller.SetCursor(4)

		m.scroller.CursorDown()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor to stay at 4, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.scroller.CursorDown()

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
		m.scroller.SetCursor(5)

		m.scroller.CursorUp()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)

		m.scroller.CursorUp()

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

		m.scroller.CursorToEnd()

		if m.Cursor() != 9 {
			t.Errorf("expected cursor 9, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.scroller.CursorToEnd()

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

		m.scroller.PageDown()

		if m.Cursor() <= 0 {
			t.Errorf("expected cursor to move forward, got %d", m.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.scroller.SetCursor(3)

		m.scroller.PageDown()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.scroller.PageDown()

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
		m.scroller.SetCursor(20)

		m.scroller.PageUp()

		if m.Cursor() >= 20 {
			t.Errorf("expected cursor to move backward from 20, got %d", m.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.scroller.SetCursor(2)

		m.scroller.PageUp()

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
		m.scroller.SetCursor(2)

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

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}
