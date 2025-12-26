package panel

import "testing"

func TestScroller_Navigation(t *testing.T) {
	t.Parallel()

	s := NewScroller(10, 5)

	// Initial state
	if s.Cursor() != 0 {
		t.Errorf("initial cursor = %d, want 0", s.Cursor())
	}

	// Move down
	if !s.CursorDown() {
		t.Error("CursorDown should return true when moving from 0")
	}
	if s.Cursor() != 1 {
		t.Errorf("cursor after down = %d, want 1", s.Cursor())
	}

	// Move up
	if !s.CursorUp() {
		t.Error("CursorUp should return true when moving from 1")
	}
	if s.Cursor() != 0 {
		t.Errorf("cursor after up = %d, want 0", s.Cursor())
	}

	// Can't move up from 0
	if s.CursorUp() {
		t.Error("CursorUp should return false at position 0")
	}

	// Move to end
	s.CursorToEnd()
	if s.Cursor() != 9 {
		t.Errorf("cursor at end = %d, want 9", s.Cursor())
	}

	// Can't move down from end
	if s.CursorDown() {
		t.Error("CursorDown should return false at end")
	}

	// Move to start
	s.CursorToStart()
	if s.Cursor() != 0 {
		t.Errorf("cursor at start = %d, want 0", s.Cursor())
	}
}

func TestScroller_VisibleRange(t *testing.T) {
	t.Parallel()

	s := NewScroller(20, 5)

	// Initial range
	start, end := s.VisibleRange()
	if start != 0 || end != 5 {
		t.Errorf("initial range = [%d,%d), want [0,5)", start, end)
	}

	// Move cursor down to trigger scroll
	for range 10 {
		s.CursorDown()
	}

	start, end = s.VisibleRange()
	if start != 6 || end != 11 {
		t.Errorf("scrolled range = [%d,%d), want [6,11)", start, end)
	}

	// Check visibility
	if s.IsVisible(5) {
		t.Error("index 5 should not be visible")
	}
	if !s.IsVisible(10) {
		t.Error("index 10 (cursor) should be visible")
	}
}

func TestScroller_ScrollInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		itemCount int
		visible   int
		cursor    int
		wantInfo  string
		wantRange string
	}{
		{"empty", 0, 5, 0, "[0/0]", "[0/0]"},
		{"single", 1, 5, 0, "[1/1]", "[1]"},
		{"partial", 3, 5, 1, "[2/3]", "[3]"},
		{"scrolled", 20, 5, 10, "[11/20]", "[7-11/20]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewScroller(tt.itemCount, tt.visible)
			s.SetCursor(tt.cursor)

			if got := s.ScrollInfo(); got != tt.wantInfo {
				t.Errorf("ScrollInfo() = %q, want %q", got, tt.wantInfo)
			}
			if got := s.ScrollInfoRange(); got != tt.wantRange {
				t.Errorf("ScrollInfoRange() = %q, want %q", got, tt.wantRange)
			}
		})
	}
}

func TestScroller_PageNavigation(t *testing.T) {
	t.Parallel()

	s := NewScroller(20, 5)

	// Page down
	s.PageDown()
	if s.Cursor() != 5 {
		t.Errorf("cursor after PageDown = %d, want 5", s.Cursor())
	}

	// Page down again
	s.PageDown()
	if s.Cursor() != 10 {
		t.Errorf("cursor after 2nd PageDown = %d, want 10", s.Cursor())
	}

	// Page up
	s.PageUp()
	if s.Cursor() != 5 {
		t.Errorf("cursor after PageUp = %d, want 5", s.Cursor())
	}

	// Page up to start
	s.PageUp()
	if s.Cursor() != 0 {
		t.Errorf("cursor at start = %d, want 0", s.Cursor())
	}

	// Page down to near end
	s.SetCursor(17)
	s.PageDown()
	if s.Cursor() != 19 {
		t.Errorf("cursor after PageDown at end = %d, want 19", s.Cursor())
	}
}

func TestScroller_SetItemCount(t *testing.T) {
	t.Parallel()

	s := NewScroller(10, 5)
	s.SetCursor(8)

	// Reduce item count - cursor should clamp
	s.SetItemCount(5)
	if s.Cursor() != 4 {
		t.Errorf("cursor after reduce = %d, want 4", s.Cursor())
	}

	// Set to empty
	s.SetItemCount(0)
	if s.Cursor() != 0 {
		t.Errorf("cursor after empty = %d, want 0", s.Cursor())
	}

	// Restore
	s.SetItemCount(20)
	if s.Cursor() != 0 {
		t.Errorf("cursor after restore = %d, want 0", s.Cursor())
	}
}

func TestScroller_HasMorePrevious(t *testing.T) {
	t.Parallel()

	s := NewScroller(10, 5)

	// At start
	if s.HasPrevious() {
		t.Error("HasPrevious should be false at start")
	}
	if !s.HasMore() {
		t.Error("HasMore should be true at start")
	}

	// In middle
	s.SetCursor(7)
	if !s.HasPrevious() {
		t.Error("HasPrevious should be true in middle")
	}
	if !s.HasMore() {
		t.Error("HasMore should be true in middle")
	}

	// At end
	s.CursorToEnd()
	if !s.HasPrevious() {
		t.Error("HasPrevious should be true at end")
	}
	if s.HasMore() {
		t.Error("HasMore should be false at end")
	}
}

func TestScroller_SmallList(t *testing.T) {
	t.Parallel()

	// List smaller than viewport
	s := NewScroller(3, 10)

	start, end := s.VisibleRange()
	if start != 0 || end != 3 {
		t.Errorf("range for small list = [%d,%d), want [0,3)", start, end)
	}

	if s.HasMore() {
		t.Error("HasMore should be false for small list")
	}
	if s.HasPrevious() {
		t.Error("HasPrevious should be false for small list")
	}
}

func Test_itoa(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{-1, "-1"},
		{-456, "-456"},
	}

	for _, tt := range tests {
		if got := itoa(tt.input); got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
