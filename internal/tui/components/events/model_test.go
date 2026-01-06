package events

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// createTestModel creates a model with test events for navigation testing.
func createTestModel(eventCount int) Model {
	m := Model{
		Sizable: helpers.NewSizable(1, panel.NewScroller(eventCount, 10)),
		state: &sharedState{
			userEvents:    make([]Event, eventCount),
			systemEvents:  make([]Event, 0),
			subscriptions: make(map[string]context.CancelFunc),
			connStatus:    make(map[string]bool),
		},
		maxItems:   100,
		autoScroll: true,
		styles:     DefaultStyles(),
	}
	m = m.SetSize(100, 100) // Large enough for 10+ visible rows
	// Create test events
	for i := range eventCount {
		m.state.userEvents[i] = Event{
			Timestamp:   time.Now(),
			Device:      "test-device",
			Type:        "info",
			Description: "test event",
		}
	}
	return m
}

func TestScrollerCursorDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor down", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m.Scroller.CursorDown()

		if m.Scroller.Cursor() != 1 {
			t.Errorf("expected cursor 1, got %d", m.Scroller.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.Scroller.SetCursor(4) // Last item

		m.Scroller.CursorDown()

		if m.Scroller.Cursor() != 4 {
			t.Errorf("expected cursor to stay at 4, got %d", m.Scroller.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.CursorDown()

		if m.Scroller.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Scroller.Cursor())
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

		if m.Scroller.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Scroller.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)

		m.Scroller.CursorUp()

		if m.Scroller.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Scroller.Cursor())
		}
	})
}

func TestScrollerCursorToEnd(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor to last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m.Scroller.CursorToEnd()

		if m.Scroller.Cursor() != 9 {
			t.Errorf("expected cursor 9, got %d", m.Scroller.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.CursorToEnd()

		if m.Scroller.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Scroller.Cursor())
		}
	})
}

func TestScrollerPageDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor by visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m.Scroller.SetVisibleRows(20)

		m.Scroller.PageDown()

		if m.Scroller.Cursor() != 20 {
			t.Errorf("expected cursor 20, got %d", m.Scroller.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.Scroller.SetCursor(3)

		m.Scroller.PageDown()

		if m.Scroller.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Scroller.Cursor())
		}
	})
}

func TestScrollerPageUp(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor by visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m.Scroller.SetVisibleRows(20)
		m.Scroller.SetCursor(40)

		m.Scroller.PageUp()

		if m.Scroller.Cursor() != 20 {
			t.Errorf("expected cursor 20, got %d", m.Scroller.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.Scroller.SetCursor(2)

		m.Scroller.PageUp()

		if m.Scroller.Cursor() != 0 {
			t.Errorf("expected cursor 0, got %d", m.Scroller.Cursor())
		}
	})
}

func TestTogglePause(t *testing.T) {
	t.Parallel()

	t.Run("toggles from false to true via key", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = false

		m = m.handleKeyPress(tea.KeyPressMsg{Code: 'p'})

		if !m.paused {
			t.Error("expected paused to be true")
		}
	})

	t.Run("toggles from true to false via key", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = true

		m = m.handleKeyPress(tea.KeyPressMsg{Code: 'p'})

		if m.paused {
			t.Error("expected paused to be false")
		}
	})
}

func TestAddEventWhenPaused(t *testing.T) {
	t.Parallel()

	t.Run("does not add events when paused", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = true

		m.addEvent(Event{
			Timestamp:   time.Now(),
			Device:      "test",
			Description: "should not be added",
		})

		if m.eventCount() != 0 {
			t.Errorf("expected 0 events when paused, got %d", m.eventCount())
		}
	})

	t.Run("adds events when not paused", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = false

		m.addEvent(Event{
			Timestamp:   time.Now(),
			Device:      "test",
			Description: "should be added",
		})

		if m.eventCount() != 1 {
			t.Errorf("expected 1 event, got %d", m.eventCount())
		}
	})
}

func TestEventCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{"empty", 0, 0},
		{"single", 1, 1},
		{"multiple", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := createTestModel(tt.count)

			if m.eventCount() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, m.eventCount())
			}
		})
	}
}

func TestScrollerVisibleRows(t *testing.T) {
	t.Parallel()

	t.Run("returns configured visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.Scroller.SetVisibleRows(15)

		if m.Scroller.VisibleRows() != 15 {
			t.Errorf("expected 15 visible rows, got %d", m.Scroller.VisibleRows())
		}
	})
}

func TestAutoScrollBehavior(t *testing.T) {
	t.Parallel()

	t.Run("autoScroll is true at cursor 0", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.autoScroll = true

		// Moving down should disable autoScroll
		m.Scroller.CursorDown()
		m.autoScroll = m.Scroller.Cursor() == 0

		if m.autoScroll {
			t.Error("expected autoScroll to be false after moving cursor from 0")
		}
	})

	t.Run("autoScroll re-enabled when returning to cursor 0", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.Scroller.SetCursor(1)
		m.autoScroll = false

		// Moving up to 0 should enable autoScroll
		m.Scroller.CursorUp()
		m.autoScroll = m.Scroller.Cursor() == 0

		if !m.autoScroll {
			t.Error("expected autoScroll to be true when cursor is at 0")
		}
	})
}

func TestSetSize(t *testing.T) {
	t.Parallel()

	t.Run("updates scroller visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m = m.SetSize(80, 20)

		// SetSize reserves 1 row for header
		if m.Scroller.VisibleRows() != 19 {
			t.Errorf("expected 19 visible rows, got %d", m.Scroller.VisibleRows())
		}
	})

	t.Run("minimum visible rows is 1", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m = m.SetSize(80, 1) // Very small

		if m.Scroller.VisibleRows() < 1 {
			t.Errorf("expected at least 1 visible row, got %d", m.Scroller.VisibleRows())
		}
	})
}

func TestScrollInfo(t *testing.T) {
	t.Parallel()

	t.Run("returns position info", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.Scroller.SetVisibleRows(10)

		info := m.ScrollInfo()

		// Scroller always returns position info like "[1/5]"
		if info != "[1/5]" {
			t.Errorf("expected [1/5], got %q", info)
		}
	})

	t.Run("returns scroll info with position", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(20)
		m.Scroller.SetVisibleRows(10)
		m.userCursor = 5

		info := m.ScrollInfo()

		if info != "[6/20]" {
			t.Errorf("expected [6/20], got %q", info)
		}
	})
}
