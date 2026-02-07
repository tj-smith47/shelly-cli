package panel

import (
	"testing"
)

func TestNewSizable(t *testing.T) {
	t.Parallel()

	scroller := NewScroller(0, 10)
	s := NewSizable(4, scroller)

	if s.Scroller != scroller {
		t.Error("expected scroller to be set")
	}
	if s.scrollOffset != 4 {
		t.Errorf("expected scrollOffset 4, got %d", s.scrollOffset)
	}
	if s.Loader.Message() != "Loading..." {
		t.Error("expected default loader message")
	}
}

func TestNewSizableLoaderOnly(t *testing.T) {
	t.Parallel()

	s := NewSizableLoaderOnly()

	if s.Scroller != nil {
		t.Error("expected nil scroller")
	}
	if s.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0, got %d", s.scrollOffset)
	}
	if s.Loader.Message() != "Loading..." {
		t.Error("expected default loader message")
	}
}

func TestSizable_ApplySize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scrollOffset int
		width        int
		height       int
		expectRows   int
	}{
		{
			name:         "standard offset 4",
			scrollOffset: 4,
			width:        80,
			height:       20,
			expectRows:   16,
		},
		{
			name:         "offset 10",
			scrollOffset: 10,
			width:        100,
			height:       30,
			expectRows:   20,
		},
		{
			name:         "clamp to minimum 1",
			scrollOffset: 10,
			width:        50,
			height:       5,
			expectRows:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scroller := NewScroller(0, 100)
			s := NewSizable(tt.scrollOffset, scroller)

			s.ApplySize(tt.width, tt.height)

			if s.Width != tt.width {
				t.Errorf("expected Width %d, got %d", tt.width, s.Width)
			}
			if s.Height != tt.height {
				t.Errorf("expected Height %d, got %d", tt.height, s.Height)
			}
			if scroller.VisibleRows() != tt.expectRows {
				t.Errorf("expected %d visible rows, got %d", tt.expectRows, scroller.VisibleRows())
			}
		})
	}
}

func TestSizable_ApplySizeLoaderOnly(t *testing.T) {
	t.Parallel()

	s := NewSizableLoaderOnly()
	s.ApplySize(80, 40)

	if s.Width != 80 {
		t.Errorf("expected Width 80, got %d", s.Width)
	}
	if s.Height != 40 {
		t.Errorf("expected Height 40, got %d", s.Height)
	}
	// Should not panic with nil scroller
}

func TestSizable_VisibleRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scrollOffset int
		height       int
		expectRows   int
	}{
		{
			name:         "normal calculation",
			scrollOffset: 4,
			height:       20,
			expectRows:   16,
		},
		{
			name:         "clamp to 1",
			scrollOffset: 10,
			height:       5,
			expectRows:   1,
		},
		{
			name:         "zero offset",
			scrollOffset: 0,
			height:       20,
			expectRows:   20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := Sizable{
				Height:       tt.height,
				scrollOffset: tt.scrollOffset,
			}

			if got := s.VisibleRows(); got != tt.expectRows {
				t.Errorf("expected %d, got %d", tt.expectRows, got)
			}
		})
	}
}

func TestSizable_SetScrollOffset(t *testing.T) {
	t.Parallel()

	s := NewSizable(4, nil)
	if s.ScrollOffset() != 4 {
		t.Errorf("expected 4, got %d", s.ScrollOffset())
	}

	s.SetScrollOffset(10)
	if s.ScrollOffset() != 10 {
		t.Errorf("expected 10, got %d", s.ScrollOffset())
	}
}

func TestSizable_ApplySizeWithExtraLoaders(t *testing.T) {
	t.Parallel()

	scroller := NewScroller(0, 100)
	s := NewSizable(4, scroller)

	extra1 := s.Loader
	extra2 := s.Loader

	result := s.ApplySizeWithExtraLoaders(80, 40, extra1, extra2)

	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	// Verify dimensions were applied
	if s.Width != 80 {
		t.Errorf("expected Width 80, got %d", s.Width)
	}
	if s.Height != 40 {
		t.Errorf("expected Height 40, got %d", s.Height)
	}

	// Verify scroller was updated (height 40 - offset 4 = 36)
	if scroller.VisibleRows() != 36 {
		t.Errorf("expected 36 visible rows, got %d", scroller.VisibleRows())
	}
}

func TestLoaderBorderOffset(t *testing.T) {
	t.Parallel()

	if LoaderBorderOffset != 4 {
		t.Errorf("expected LoaderBorderOffset to be 4, got %d", LoaderBorderOffset)
	}
}
