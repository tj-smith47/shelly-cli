package keys

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

func TestHandleScrollNavigation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		key            string
		itemCount      int
		startCursor    int
		expectedCursor int
		handled        bool
	}{
		{
			name:           "j moves cursor down",
			key:            "j",
			itemCount:      10,
			startCursor:    0,
			expectedCursor: 1,
			handled:        true,
		},
		{
			name:           "down moves cursor down",
			key:            "down",
			itemCount:      10,
			startCursor:    0,
			expectedCursor: 1,
			handled:        true,
		},
		{
			name:           "k moves cursor up",
			key:            "k",
			itemCount:      10,
			startCursor:    5,
			expectedCursor: 4,
			handled:        true,
		},
		{
			name:           "up moves cursor up",
			key:            "up",
			itemCount:      10,
			startCursor:    5,
			expectedCursor: 4,
			handled:        true,
		},
		{
			name:           "g moves to start",
			key:            "g",
			itemCount:      10,
			startCursor:    5,
			expectedCursor: 0,
			handled:        true,
		},
		{
			name:           "G moves to end",
			key:            "G",
			itemCount:      10,
			startCursor:    0,
			expectedCursor: 9,
			handled:        true,
		},
		{
			name:           "ctrl+d pages down",
			key:            "ctrl+d",
			itemCount:      20,
			startCursor:    0,
			expectedCursor: 5, // visibleRows is 5
			handled:        true,
		},
		{
			name:           "pgdown pages down",
			key:            "pgdown",
			itemCount:      20,
			startCursor:    0,
			expectedCursor: 5,
			handled:        true,
		},
		{
			name:           "ctrl+u pages up",
			key:            "ctrl+u",
			itemCount:      20,
			startCursor:    10,
			expectedCursor: 5,
			handled:        true,
		},
		{
			name:           "pgup pages up",
			key:            "pgup",
			itemCount:      20,
			startCursor:    10,
			expectedCursor: 5,
			handled:        true,
		},
		{
			name:           "unrecognized key not handled",
			key:            "x",
			itemCount:      10,
			startCursor:    5,
			expectedCursor: 5,
			handled:        false,
		},
		{
			name:           "empty key not handled",
			key:            "",
			itemCount:      10,
			startCursor:    5,
			expectedCursor: 5,
			handled:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scroller := panel.NewScroller(tt.itemCount, 5)
			scroller.SetCursor(tt.startCursor)

			handled := HandleScrollNavigation(tt.key, scroller)

			if handled != tt.handled {
				t.Errorf("expected handled=%v, got %v", tt.handled, handled)
			}
			if scroller.Cursor() != tt.expectedCursor {
				t.Errorf("expected cursor=%d, got %d", tt.expectedCursor, scroller.Cursor())
			}
		})
	}
}

func TestHandleScrollNavigationScroller(t *testing.T) {
	t.Parallel()

	t.Run("nil scroller returns false", func(t *testing.T) {
		t.Parallel()
		handled := HandleScrollNavigationScroller("j", nil)
		if handled {
			t.Error("expected false for nil scroller")
		}
	})

	t.Run("valid scroller delegates to HandleScrollNavigation", func(t *testing.T) {
		t.Parallel()
		scroller := panel.NewScroller(10, 5)
		handled := HandleScrollNavigationScroller("j", scroller)
		if !handled {
			t.Error("expected true for valid navigation key")
		}
		if scroller.Cursor() != 1 {
			t.Errorf("expected cursor=1, got %d", scroller.Cursor())
		}
	})
}

func TestHandleScrollNavigationBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("k at start stays at 0", func(t *testing.T) {
		t.Parallel()
		scroller := panel.NewScroller(10, 5)
		scroller.SetCursor(0)
		HandleScrollNavigation("k", scroller)
		if scroller.Cursor() != 0 {
			t.Errorf("expected cursor=0, got %d", scroller.Cursor())
		}
	})

	t.Run("j at end stays at last", func(t *testing.T) {
		t.Parallel()
		scroller := panel.NewScroller(10, 5)
		scroller.SetCursor(9)
		HandleScrollNavigation("j", scroller)
		if scroller.Cursor() != 9 {
			t.Errorf("expected cursor=9, got %d", scroller.Cursor())
		}
	})

	t.Run("pageup from near start clamps to 0", func(t *testing.T) {
		t.Parallel()
		scroller := panel.NewScroller(10, 5)
		scroller.SetCursor(2)
		HandleScrollNavigation("pgup", scroller)
		if scroller.Cursor() != 0 {
			t.Errorf("expected cursor=0, got %d", scroller.Cursor())
		}
	})

	t.Run("pagedown from near end clamps to last", func(t *testing.T) {
		t.Parallel()
		scroller := panel.NewScroller(10, 5)
		scroller.SetCursor(7)
		HandleScrollNavigation("pgdown", scroller)
		if scroller.Cursor() != 9 {
			t.Errorf("expected cursor=9, got %d", scroller.Cursor())
		}
	})
}
