package energy

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// createTestModel creates a model with test devices for navigation testing.
func createTestModel(deviceCount int) Model {
	m := Model{
		devices:  make([]DeviceEnergy, deviceCount),
		scroller: panel.NewScroller(0, 10),
		height:   100,
		styles:   DefaultStyles(),
	}
	for i := range deviceCount {
		m.devices[i] = DeviceEnergy{
			Name:   "test-device",
			Online: true,
		}
	}
	m.scroller.SetItemCount(deviceCount)
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

func TestVisibleDevices(t *testing.T) {
	t.Parallel()

	t.Run("returns minimum for small height", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.height = 5

		if m.visibleDevices() < 1 {
			t.Error("expected at least 1 visible device")
		}
	})

	t.Run("calculates based on height", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.height = 100

		devices := m.visibleDevices()
		if devices <= 0 {
			t.Error("expected positive visible devices for height 100")
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
