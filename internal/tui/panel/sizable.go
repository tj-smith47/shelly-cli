// Package panel provides shared utilities for TUI panel components.
package panel

import (
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
)

// LoaderBorderOffset is the standard offset for loader positioning within borders.
const LoaderBorderOffset = 4

// Sizable provides common size-related state and methods for TUI components.
// Embed this struct in component models to get standardized SetSize handling.
//
// Example usage:
//
//	type Model struct {
//	    panel.Sizable
//	    items []Item
//	}
//
//	func New() Model {
//	    return Model{
//	        Sizable: panel.NewSizable(4, panel.NewScroller(0, 10)),
//	    }
//	}
//
//	func (m Model) SetSize(width, height int) Model {
//	    m.ApplySize(width, height)
//	    return m
//	}
type Sizable struct {
	Width        int
	Height       int
	ModalWidth   int // Screen-based modal dimensions (set by SetEditModalSize)
	ModalHeight  int
	Loader       loading.Model
	Scroller     *Scroller
	scrollOffset int // Header offset for scroller visible rows calculation
}

// NewSizable creates a Sizable with a scroller and scroll offset.
// The scrollOffset accounts for borders, title, footer, and other header content.
// Common offsets: 4 (minimal), 5-6 (with stats), 8-10 (with controls), 12 (complex).
func NewSizable(scrollOffset int, scroller *Scroller) Sizable {
	return Sizable{
		scrollOffset: scrollOffset,
		Scroller:     scroller,
		Loader: loading.New(
			loading.WithMessage("Loading..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// NewSizableLoaderOnly creates a Sizable without a scroller.
// Use this for components that only need loader sizing.
func NewSizableLoaderOnly() Sizable {
	return Sizable{
		Loader: loading.New(
			loading.WithMessage("Loading..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// ApplySize updates dimensions, resizes the loader, and updates scroller visible rows.
// Call this from your component's SetSize method.
func (s *Sizable) ApplySize(width, height int) {
	s.Width = width
	s.Height = height
	s.Loader = s.Loader.SetSize(width-LoaderBorderOffset, height-LoaderBorderOffset)

	if s.Scroller != nil && s.scrollOffset > 0 {
		rows := height - s.scrollOffset
		if rows < 1 {
			rows = 1
		}
		s.Scroller.SetVisibleRows(rows)
	}
}

// EditModalDims returns the screen-based modal dimensions if available,
// falling back to panel dimensions. Use this when opening edit modals
// to avoid a first-frame size mismatch.
func (s *Sizable) EditModalDims() (width, height int) {
	if s.ModalWidth > 0 && s.ModalHeight > 0 {
		return s.ModalWidth, s.ModalHeight
	}
	return s.Width, s.Height
}

// SetScrollOffset updates the scroll offset for dynamic layouts.
// Most components set this once at construction via NewSizable.
func (s *Sizable) SetScrollOffset(offset int) {
	s.scrollOffset = offset
}

// ScrollOffset returns the current scroll offset.
func (s *Sizable) ScrollOffset() int {
	return s.scrollOffset
}

// VisibleRows returns the calculated visible rows based on current height and offset.
func (s *Sizable) VisibleRows() int {
	rows := s.Height - s.scrollOffset
	if rows < 1 {
		return 1
	}
	return rows
}

// ResizeLoader resizes just the loader without affecting other dimensions.
// Useful for components that need to resize loader independently.
func (s *Sizable) ResizeLoader() {
	s.Loader = s.Loader.SetSize(s.Width-LoaderBorderOffset, s.Height-LoaderBorderOffset)
}

// ApplySizeWithExtraLoaders handles components with multiple loaders.
// Updates the embedded Loader plus any additional loaders passed in.
// Returns the resized extra loaders in the same order they were passed.
//
// Example:
//
//	func (m Model) SetSize(width, height int) Model {
//	    extras := m.ApplySizeWithExtraLoaders(width, height, m.updateLoader, m.checkLoader)
//	    m.updateLoader, m.checkLoader = extras[0], extras[1]
//	    return m
//	}
func (s *Sizable) ApplySizeWithExtraLoaders(width, height int, extraLoaders ...loading.Model) []loading.Model {
	s.ApplySize(width, height)

	result := make([]loading.Model, len(extraLoaders))
	for i, loader := range extraLoaders {
		result[i] = loader.SetSize(width-LoaderBorderOffset, height-LoaderBorderOffset)
	}
	return result
}
