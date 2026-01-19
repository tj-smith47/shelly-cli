package messages

import (
	"testing"
)

func TestNavDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dir  NavDirection
		want int
	}{
		{NavUp, 0},
		{NavDown, 1},
		{NavLeft, 2},
		{NavRight, 3},
		{NavPageUp, 4},
		{NavPageDown, 5},
		{NavHome, 6},
		{NavEnd, 7},
	}

	for _, tt := range tests {
		if int(tt.dir) != tt.want {
			t.Errorf("NavDirection %d: got %d, want %d", tt.dir, int(tt.dir), tt.want)
		}
	}
}

func TestNavigationMsg(t *testing.T) {
	t.Parallel()

	msg := NavigationMsg{Direction: NavDown}
	if msg.Direction != NavDown {
		t.Errorf("NavigationMsg.Direction = %v, want NavDown", msg.Direction)
	}
}

func TestSnoozeRequestMsg(t *testing.T) {
	t.Parallel()

	msg := SnoozeRequestMsg{Duration: "1h"}
	if msg.Duration != "1h" {
		t.Errorf("SnoozeRequestMsg.Duration = %q, want %q", msg.Duration, "1h")
	}
}

func TestModeSelectMsg(t *testing.T) {
	t.Parallel()

	msg := ModeSelectMsg{Mode: 2}
	if msg.Mode != 2 {
		t.Errorf("ModeSelectMsg.Mode = %d, want 2", msg.Mode)
	}
}
