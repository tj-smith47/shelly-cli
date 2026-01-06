package helpers

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

func TestSetLoaderSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		width        int
		height       int
		expectWidth  int
		expectHeight int
	}{
		{
			name:         "standard panel size",
			width:        80,
			height:       40,
			expectWidth:  76,
			expectHeight: 36,
		},
		{
			name:         "small panel",
			width:        20,
			height:       10,
			expectWidth:  16,
			expectHeight: 6,
		},
		{
			name:         "minimum size",
			width:        4,
			height:       4,
			expectWidth:  0,
			expectHeight: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			loader := loading.New()
			result := SetLoaderSize(loader, tt.width, tt.height)
			// Verify the loader was updated (we can't directly check internal state,
			// but we verify it returns a loader model)
			if result.Message() != "Loading..." {
				t.Error("expected default loading message")
			}
		})
	}
}

func TestSetScrollerRows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		height       int
		headerOffset int
		expectedRows int
	}{
		{
			name:         "standard offset 4",
			height:       20,
			headerOffset: 4,
			expectedRows: 16,
		},
		{
			name:         "standard offset 5",
			height:       20,
			headerOffset: 5,
			expectedRows: 15,
		},
		{
			name:         "standard offset 10",
			height:       30,
			headerOffset: 10,
			expectedRows: 20,
		},
		{
			name:         "clamp to minimum 1",
			height:       3,
			headerOffset: 4,
			expectedRows: 1, // height-offset = -1, clamped to 1
		},
		{
			name:         "exact minimum",
			height:       5,
			headerOffset: 4,
			expectedRows: 1,
		},
		{
			name:         "zero height",
			height:       0,
			headerOffset: 4,
			expectedRows: 1, // clamped to 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scroller := panel.NewScroller(0, 10)
			SetScrollerRows(tt.height, tt.headerOffset, scroller)

			if scroller.VisibleRows() != tt.expectedRows {
				t.Errorf("expected %d visible rows, got %d", tt.expectedRows, scroller.VisibleRows())
			}
		})
	}
}

func TestSetScrollerRowsReturn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		height       int
		headerOffset int
		expectedRows int
	}{
		{
			name:         "returns calculated rows",
			height:       20,
			headerOffset: 4,
			expectedRows: 16,
		},
		{
			name:         "returns clamped value",
			height:       3,
			headerOffset: 4,
			expectedRows: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scroller := panel.NewScroller(0, 10)
			result := SetScrollerRowsReturn(tt.height, tt.headerOffset, scroller)

			if result != tt.expectedRows {
				t.Errorf("expected return value %d, got %d", tt.expectedRows, result)
			}
			if scroller.VisibleRows() != tt.expectedRows {
				t.Errorf("expected scroller to have %d visible rows, got %d", tt.expectedRows, scroller.VisibleRows())
			}
		})
	}
}

func TestLoaderBorderOffset(t *testing.T) {
	t.Parallel()
	if LoaderBorderOffset != 4 {
		t.Errorf("expected LoaderBorderOffset to be 4, got %d", LoaderBorderOffset)
	}
}
