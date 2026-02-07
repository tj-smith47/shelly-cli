package editmodal

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func TestBase_NextField_Wraps(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 3, Cursor: 2}
	old, cur := b.NextField()

	if old != 2 {
		t.Errorf("old = %d, want 2", old)
	}
	if cur != 0 {
		t.Errorf("new cursor = %d, want 0 (wrapped)", cur)
	}
}

func TestBase_NextField_Increment(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 3, Cursor: 0}
	old, cur := b.NextField()

	if old != 0 {
		t.Errorf("old = %d, want 0", old)
	}
	if cur != 1 {
		t.Errorf("new cursor = %d, want 1", cur)
	}
}

func TestBase_PrevField_Wraps(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 3, Cursor: 0}
	old, cur := b.PrevField()

	if old != 0 {
		t.Errorf("old = %d, want 0", old)
	}
	if cur != 2 {
		t.Errorf("new cursor = %d, want 2 (wrapped)", cur)
	}
}

func TestBase_PrevField_Decrement(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 3, Cursor: 2}
	old, cur := b.PrevField()

	if old != 2 {
		t.Errorf("old = %d, want 2", old)
	}
	if cur != 1 {
		t.Errorf("new cursor = %d, want 1", cur)
	}
}

func TestBase_SetCursor(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 5}
	b.SetCursor(3)

	if b.Cursor != 3 {
		t.Errorf("Cursor = %d, want 3", b.Cursor)
	}
}

func TestBase_HandleNavigation(t *testing.T) {
	t.Parallel()
	b := Base{FieldCount: 3}

	tests := []struct {
		direction messages.NavDirection
		want      KeyAction
	}{
		{messages.NavUp, ActionNavUp},
		{messages.NavDown, ActionNavDown},
		{messages.NavLeft, ActionNone},
		{messages.NavRight, ActionNone},
		{messages.NavPageUp, ActionNone},
		{messages.NavPageDown, ActionNone},
		{messages.NavHome, ActionNone},
		{messages.NavEnd, ActionNone},
	}

	for _, tt := range tests {
		msg := messages.NavigationMsg{Direction: tt.direction}
		got := b.HandleNavigation(msg)
		if got != tt.want {
			t.Errorf("HandleNavigation(%d) = %d, want %d", tt.direction, got, tt.want)
		}
	}
}
