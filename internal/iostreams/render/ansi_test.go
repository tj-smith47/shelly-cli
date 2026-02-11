package render_test

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams/render"
)

func TestMoveUp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		n    int
		want string
	}{
		{0, ""},
		{-1, ""},
		{1, "\x1b[1A"},
		{5, "\x1b[5A"},
	}

	for _, tt := range tests {
		if got := render.MoveUp(tt.n); got != tt.want {
			t.Errorf("MoveUp(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestMoveDown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		n    int
		want string
	}{
		{0, ""},
		{-1, ""},
		{1, "\x1b[1B"},
		{3, "\x1b[3B"},
	}

	for _, tt := range tests {
		if got := render.MoveDown(tt.n); got != tt.want {
			t.Errorf("MoveDown(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestClearLine(t *testing.T) {
	t.Parallel()

	if got := render.ClearLine(); got != "\x1b[2K" {
		t.Errorf("ClearLine() = %q, want %q", got, "\x1b[2K")
	}
}

func TestClearDown(t *testing.T) {
	t.Parallel()

	if got := render.ClearDown(); got != "\x1b[0J" {
		t.Errorf("ClearDown() = %q, want %q", got, "\x1b[0J")
	}
}

func TestHideCursor(t *testing.T) {
	t.Parallel()

	if got := render.HideCursor(); got != "\x1b[?25l" {
		t.Errorf("HideCursor() = %q, want %q", got, "\x1b[?25l")
	}
}

func TestShowCursor(t *testing.T) {
	t.Parallel()

	if got := render.ShowCursor(); got != "\x1b[?25h" {
		t.Errorf("ShowCursor() = %q, want %q", got, "\x1b[?25h")
	}
}

func TestVisibleLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"plain ascii", "hello", 5},
		{"with ansi color", "\x1b[31mred\x1b[0m", 3},
		{"unicode", "héllo", 5},
		{"emoji", "✓ done", 6},
		{"multiple ansi", "\x1b[1m\x1b[31mbold red\x1b[0m", 8},
		{"ansi only", "\x1b[31m\x1b[0m", 0},
		{"mixed", "pre\x1b[32mgreen\x1b[0mpost", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := render.VisibleLen(tt.s); got != tt.want {
				t.Errorf("VisibleLen(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestTruncateLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		maxWidth int
		want     string
	}{
		{"empty", "", 10, ""},
		{"zero width", "hello", 0, ""},
		{"fits", "hello", 10, "hello"},
		{"exact fit", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hell\u2026"},
		{"ansi preserved", "\x1b[31mhello world\x1b[0m", 5, "\x1b[31mhell\u2026"},
		{"unicode truncated", "héllo wörld", 5, "héll\u2026"},
		{"width 1", "hello", 1, "\u2026"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := render.TruncateLine(tt.s, tt.maxWidth); got != tt.want {
				t.Errorf("TruncateLine(%q, %d) = %q, want %q", tt.s, tt.maxWidth, got, tt.want)
			}
		})
	}
}
