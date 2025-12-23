package overlay

import (
	"strings"
	"testing"
)

func TestPlaceOverlay_Centered(t *testing.T) {
	t.Parallel()
	base := "12345\n12345\n12345\n12345\n12345"
	ovr := "XX\nXX"

	result := PlaceOverlay(base, ovr, true, true)
	lines := strings.Split(result, "\n")

	if len(lines) != 5 {
		t.Errorf("lines count = %d, want 5", len(lines))
	}

	// Overlay should be in the middle (row 1-2, col 1-2)
	// Row 1: "1XX45"
	// Row 2: "1XX45"
	if !strings.Contains(lines[1], "XX") {
		t.Errorf("line 1 = %q, should contain XX", lines[1])
	}
	if !strings.Contains(lines[2], "XX") {
		t.Errorf("line 2 = %q, should contain XX", lines[2])
	}
}

func TestPlaceAt(t *testing.T) {
	t.Parallel()
	base := ".....\n.....\n....."
	ovr := "XX"

	result := PlaceAt(base, ovr, 2, 1)
	lines := strings.Split(result, "\n")

	if len(lines) != 3 {
		t.Errorf("lines count = %d, want 3", len(lines))
	}

	// Line 1 should have XX at position 2
	expected := "..XX."
	if lines[1] != expected {
		t.Errorf("line 1 = %q, want %q", lines[1], expected)
	}
}

func TestPlaceAt_MultiLine(t *testing.T) {
	t.Parallel()
	base := ".....\n.....\n.....\n....."
	ovr := "AB\nCD"

	result := PlaceAt(base, ovr, 1, 1)
	lines := strings.Split(result, "\n")

	if lines[1] != ".AB.." {
		t.Errorf("line 1 = %q, want '.AB..'", lines[1])
	}
	if lines[2] != ".CD.." {
		t.Errorf("line 2 = %q, want '.CD..'", lines[2])
	}
}

func TestPlaceAt_EdgeCase_StartOfLine(t *testing.T) {
	t.Parallel()
	base := ".....\n....."
	ovr := "XX"

	result := PlaceAt(base, ovr, 0, 0)
	lines := strings.Split(result, "\n")

	if lines[0] != "XX..." {
		t.Errorf("line 0 = %q, want 'XX...'", lines[0])
	}
}

func TestPlaceAt_EdgeCase_EndOfLine(t *testing.T) {
	t.Parallel()
	base := ".....\n....."
	ovr := "XX"

	result := PlaceAt(base, ovr, 3, 0)
	lines := strings.Split(result, "\n")

	if lines[0] != "...XX" {
		t.Errorf("line 0 = %q, want '...XX'", lines[0])
	}
}

func TestCenter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		baseW, baseH  int
		overlayW, ovH int
		wantX, wantY  int
	}{
		{"centered", 80, 24, 40, 10, 20, 7},
		{"small overlay", 100, 50, 10, 5, 45, 22},
		{"overlay larger than base", 20, 10, 40, 20, 0, 0}, // Clipped to 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			x, y := Center(tt.baseW, tt.baseH, tt.overlayW, tt.ovH)
			if x != tt.wantX {
				t.Errorf("x = %d, want %d", x, tt.wantX)
			}
			if y != tt.wantY {
				t.Errorf("y = %d, want %d", y, tt.wantY)
			}
		})
	}
}

func TestDimensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      string
		wantWidth  int
		wantHeight int
	}{
		{"single line", "hello", 5, 1},
		{"multi line", "hello\nworld", 5, 2},
		{"varied lengths", "hi\nhello\nbye", 5, 3},
		{"empty", "", 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w, h := Dimensions(tt.input)
			if w != tt.wantWidth {
				t.Errorf("width = %d, want %d", w, tt.wantWidth)
			}
			if h != tt.wantHeight {
				t.Errorf("height = %d, want %d", h, tt.wantHeight)
			}
		})
	}
}

func TestTruncateToWidth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		width   int
		wantLen int
	}{
		{"exact", "hello", 5, 5},
		{"shorter", "hi", 5, 5}, // Padded to 5
		{"longer", "hello world", 5, 5},
		{"zero", "hello", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := truncateToWidth(tt.input, tt.width)
			if len(result) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}
