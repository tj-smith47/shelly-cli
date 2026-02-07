package events

import (
	"context"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// createTestModel creates a model with test events for navigation testing.
func createTestModel(eventCount int) Model {
	m := Model{
		Sizable: panel.NewSizable(1, panel.NewScroller(eventCount, 10)),
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

	t.Run("toggles from false to true via action", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = false

		updated, _ := m.Update(messages.PauseRequestMsg{})

		if !updated.paused {
			t.Error("expected paused to be true")
		}
	})

	t.Run("toggles from true to false via action", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)
		m.paused = true

		updated, _ := m.Update(messages.PauseRequestMsg{})

		if updated.paused {
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

func TestMatchEventPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		pattern   string
		eventName string
		want      bool
	}{
		// Exact match
		{"exact match", "ble.scan_result", "ble.scan_result", true},
		{"exact no match", "ble.scan_result", "ble.connect", false},

		// Glob patterns with trailing *
		{"glob matches prefix", "ble.*", "ble.scan_result", true},
		{"glob matches another prefix", "ble.*", "ble.connect", true},
		{"glob no match different prefix", "ble.*", "wifi.scan", false},
		{"glob matches exact prefix", "button_*", "button_push", true},
		{"glob no match wrong prefix", "button_*", "ble.scan_result", false},

		// Edge cases
		{"empty pattern no match", "", "ble.scan_result", false},
		{"empty event no match", "ble.*", "", false},
		{"star alone matches everything", "*", "ble.scan_result", true},
		{"star alone matches empty", "*", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchEventPattern(tt.pattern, tt.eventName)
			if got != tt.want {
				t.Errorf("matchEventPattern(%q, %q) = %v, want %v",
					tt.pattern, tt.eventName, got, tt.want)
			}
		})
	}
}

func TestModel_IsFilteredEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filteredEvents []string
		eventName      string
		want           bool
	}{
		{
			name:           "exact match filters",
			filteredEvents: []string{"ble.scan_result"},
			eventName:      "ble.scan_result",
			want:           true,
		},
		{
			name:           "glob pattern filters",
			filteredEvents: []string{"ble.*"},
			eventName:      "ble.connect",
			want:           true,
		},
		{
			name:           "no match passes through",
			filteredEvents: []string{"ble.scan_result"},
			eventName:      "button_push",
			want:           false,
		},
		{
			name:           "multiple patterns first match",
			filteredEvents: []string{"ble.*", "wifi.*"},
			eventName:      "ble.scan",
			want:           true,
		},
		{
			name:           "multiple patterns second match",
			filteredEvents: []string{"ble.*", "wifi.*"},
			eventName:      "wifi.disconnect",
			want:           true,
		},
		{
			name:           "empty filter list passes all",
			filteredEvents: []string{},
			eventName:      "ble.scan_result",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := Model{filteredEvents: tt.filteredEvents}
			got := m.isFilteredEvent(tt.eventName)
			if got != tt.want {
				t.Errorf("isFilteredEvent(%q) = %v, want %v", tt.eventName, got, tt.want)
			}
		})
	}
}

func TestModel_IsFilteredComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		filteredComponents []string
		compPrefix         string
		want               bool
	}{
		{
			name:               "exact match filters",
			filteredComponents: []string{"sys", "wifi"},
			compPrefix:         "sys",
			want:               true,
		},
		{
			name:               "second item match filters",
			filteredComponents: []string{"sys", "wifi"},
			compPrefix:         "wifi",
			want:               true,
		},
		{
			name:               "no match passes through",
			filteredComponents: []string{"sys", "wifi"},
			compPrefix:         "switch",
			want:               false,
		},
		{
			name:               "empty filter list passes all",
			filteredComponents: []string{},
			compPrefix:         "sys",
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := Model{filteredComponents: tt.filteredComponents}
			got := m.isFilteredComponent(tt.compPrefix)
			if got != tt.want {
				t.Errorf("isFilteredComponent(%q) = %v, want %v", tt.compPrefix, got, tt.want)
			}
		})
	}
}
