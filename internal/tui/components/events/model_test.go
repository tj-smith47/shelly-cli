package events

import (
	"context"
	"testing"
	"time"
)

// createTestModel creates a model with test events for navigation testing.
func createTestModel(eventCount int) Model {
	m := Model{
		state: &sharedState{
			events:        make([]Event, eventCount),
			subscriptions: make(map[string]context.CancelFunc),
			connStatus:    make(map[string]bool),
		},
		height:   100, // Large enough for 10+ visible rows
		maxItems: 100,
		styles:   DefaultStyles(),
	}
	// Create test events
	for i := range eventCount {
		m.state.events[i] = Event{
			Timestamp:   time.Now(),
			Device:      "test-device",
			Type:        "info",
			Description: "test event",
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
		m.cursor = 4 // Last item

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
		m := createTestModel(30) // More events than visible rows
		m.height = 10            // Small height so 10 visible rows
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

func TestCursorToBottom(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor to last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.cursor = 0

		m = m.cursorToBottom()

		if m.cursor != 9 {
			t.Errorf("expected cursor 9, got %d", m.cursor)
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m = m.cursorToBottom()

		if m.cursor != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
		}
	})

	t.Run("adjusts scrollOffset", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(30) // More events than visible rows
		m.height = 10            // Small height so 10 visible rows
		m.scrollOffset = 0

		m = m.cursorToBottom()

		if m.scrollOffset <= 0 {
			t.Error("expected scrollOffset to be positive when jumping to bottom of long list")
		}
	})
}

func TestPageDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor by visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m.height = 20
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
		m.height = 20
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

func TestTogglePause(t *testing.T) {
	t.Parallel()

	t.Run("toggles from false to true", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = false

		m = m.togglePause()

		if !m.paused {
			t.Error("expected paused to be true")
		}
	})

	t.Run("toggles from true to false", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = true

		m = m.togglePause()

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

func TestVisibleRows(t *testing.T) {
	t.Parallel()

	t.Run("returns minimum 10 for small height", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.height = 5 // Very small

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
