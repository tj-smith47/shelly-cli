package editmodal

import "github.com/tj-smith47/shelly-cli/internal/tui/messages"

// NextField advances the cursor with wrapping, returning old and new indices
// so the modal can blur the old field and focus the new one.
func (b *Base) NextField() (oldCursor, newCursor int) {
	old := b.Cursor
	b.Cursor = (b.Cursor + 1) % b.FieldCount
	return old, b.Cursor
}

// PrevField moves the cursor back with wrapping, returning old and new indices.
func (b *Base) PrevField() (oldCursor, newCursor int) {
	old := b.Cursor
	if b.Cursor == 0 {
		b.Cursor = b.FieldCount - 1
	} else {
		b.Cursor--
	}
	return old, b.Cursor
}

// SetCursor directly sets the cursor position (for conditional skip logic).
func (b *Base) SetCursor(i int) {
	b.Cursor = i
}

// HandleNavigation translates a NavigationMsg into a KeyAction.
// Returns ActionNavUp or ActionNavDown for vertical navigation,
// ActionNone for other directions.
func (b *Base) HandleNavigation(msg messages.NavigationMsg) KeyAction {
	switch msg.Direction {
	case messages.NavUp:
		return ActionNavUp
	case messages.NavDown:
		return ActionNavDown
	default:
		return ActionNone
	}
}
