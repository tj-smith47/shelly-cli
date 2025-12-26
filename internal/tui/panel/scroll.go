// Package panel provides shared utilities for TUI panel components.
package panel

// Scroller provides cursor and scroll management for list-based panels.
// It handles navigation, visible range calculation, and scroll indicators.
type Scroller struct {
	cursor      int // Current cursor position (0-indexed)
	scroll      int // Scroll offset for viewport
	itemCount   int // Total number of items
	visibleRows int // Number of visible rows in viewport
}

// NewScroller creates a new scroller with the given item count and visible rows.
func NewScroller(itemCount, visibleRows int) *Scroller {
	return &Scroller{
		itemCount:   itemCount,
		visibleRows: visibleRows,
	}
}

// SetItemCount updates the item count and clamps cursor if needed.
func (s *Scroller) SetItemCount(count int) {
	s.itemCount = count
	if s.cursor >= count && count > 0 {
		s.cursor = count - 1
	}
	if count == 0 {
		s.cursor = 0
		s.scroll = 0
	}
	s.ensureVisible()
}

// SetVisibleRows updates the visible rows count.
func (s *Scroller) SetVisibleRows(rows int) {
	if rows < 1 {
		rows = 1
	}
	s.visibleRows = rows
	s.ensureVisible()
}

// Cursor returns the current cursor position.
func (s *Scroller) Cursor() int {
	return s.cursor
}

// SetCursor sets the cursor position, clamping to valid range.
func (s *Scroller) SetCursor(pos int) {
	s.cursor = s.clamp(pos)
	s.ensureVisible()
}

// CursorUp moves the cursor up by one position.
func (s *Scroller) CursorUp() bool {
	if s.cursor > 0 {
		s.cursor--
		s.ensureVisible()
		return true
	}
	return false
}

// CursorDown moves the cursor down by one position.
func (s *Scroller) CursorDown() bool {
	if s.cursor < s.itemCount-1 {
		s.cursor++
		s.ensureVisible()
		return true
	}
	return false
}

// CursorToStart moves the cursor to the first item.
func (s *Scroller) CursorToStart() {
	s.cursor = 0
	s.scroll = 0
}

// CursorToEnd moves the cursor to the last item.
func (s *Scroller) CursorToEnd() {
	if s.itemCount > 0 {
		s.cursor = s.itemCount - 1
		s.ensureVisible()
	}
}

// PageUp moves the cursor up by one page.
func (s *Scroller) PageUp() {
	s.cursor -= s.visibleRows
	if s.cursor < 0 {
		s.cursor = 0
	}
	s.ensureVisible()
}

// PageDown moves the cursor down by one page.
func (s *Scroller) PageDown() {
	s.cursor += s.visibleRows
	if s.cursor >= s.itemCount {
		s.cursor = s.itemCount - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	s.ensureVisible()
}

// VisibleRange returns the start (inclusive) and end (exclusive) indices
// of items that should be visible in the viewport.
func (s *Scroller) VisibleRange() (start, end int) {
	start = s.scroll
	end = start + s.visibleRows
	if end > s.itemCount {
		end = s.itemCount
	}
	if start < 0 {
		start = 0
	}
	return start, end
}

// IsVisible returns true if the item at the given index is visible.
func (s *Scroller) IsVisible(idx int) bool {
	start, end := s.VisibleRange()
	return idx >= start && idx < end
}

// ScrollInfo returns a formatted scroll info string like "[5/20]".
func (s *Scroller) ScrollInfo() string {
	if s.itemCount == 0 {
		return "[0/0]"
	}
	return "[" + itoa(s.cursor+1) + "/" + itoa(s.itemCount) + "]"
}

// ScrollInfoRange returns a range-based scroll info like "[1-10/20]".
func (s *Scroller) ScrollInfoRange() string {
	if s.itemCount == 0 {
		return "[0/0]"
	}
	start, end := s.VisibleRange()
	if start == end || end-start >= s.itemCount {
		return "[" + itoa(s.itemCount) + "]"
	}
	return "[" + itoa(start+1) + "-" + itoa(end) + "/" + itoa(s.itemCount) + "]"
}

// HasMore returns true if there are items below the visible range.
func (s *Scroller) HasMore() bool {
	_, end := s.VisibleRange()
	return end < s.itemCount
}

// HasPrevious returns true if there are items above the visible range.
func (s *Scroller) HasPrevious() bool {
	return s.scroll > 0
}

// IsCursorAt returns true if the cursor is at the given index.
func (s *Scroller) IsCursorAt(idx int) bool {
	return s.cursor == idx
}

// ItemCount returns the total number of items.
func (s *Scroller) ItemCount() int {
	return s.itemCount
}

// VisibleRows returns the number of visible rows.
func (s *Scroller) VisibleRows() int {
	return s.visibleRows
}

// ensureVisible adjusts scroll to keep cursor in view.
func (s *Scroller) ensureVisible() {
	if s.visibleRows <= 0 || s.itemCount == 0 {
		return
	}

	// Scroll down if cursor below viewport
	if s.cursor >= s.scroll+s.visibleRows {
		s.scroll = s.cursor - s.visibleRows + 1
	}

	// Scroll up if cursor above viewport
	if s.cursor < s.scroll {
		s.scroll = s.cursor
	}

	// Clamp scroll to valid range
	maxScroll := s.itemCount - s.visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}
	if s.scroll < 0 {
		s.scroll = 0
	}
}

// clamp returns the value clamped to valid cursor range [0, itemCount-1].
func (s *Scroller) clamp(val int) int {
	if val < 0 {
		return 0
	}
	if s.itemCount == 0 {
		return 0
	}
	if val >= s.itemCount {
		return s.itemCount - 1
	}
	return val
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
