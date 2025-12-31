// Package base provides shared structs and methods for TUI components.
// It reduces duplication by extracting common patterns like size management,
// focus handling, and loading states into reusable, embeddable types.
package base

// SizableModel provides width and height management for TUI components.
// Components that need size tracking should embed this struct.
//
// Example usage:
//
//	type Model struct {
//	    base.SizableModel
//	    // other fields
//	}
//
//	func (m Model) SetSize(width, height int) Model {
//	    m.SizableModel = m.SizableModel.SetSize(width, height)
//	    // component-specific size handling
//	    return m
//	}
type SizableModel struct {
	width  int
	height int
}

// Width returns the current width.
func (s SizableModel) Width() int {
	return s.width
}

// Height returns the current height.
func (s SizableModel) Height() int {
	return s.height
}

// SetSize sets the width and height, returning a new SizableModel.
// This method uses value semantics to work with BubbleTea's model pattern.
func (s SizableModel) SetSize(width, height int) SizableModel {
	s.width = width
	s.height = height
	return s
}

// FocusableModel provides focus state and panel index management for TUI components.
// Components that need focus tracking should embed this struct.
//
// Example usage:
//
//	type Model struct {
//	    base.FocusableModel
//	    // other fields
//	}
//
//	func (m Model) SetFocused(focused bool) Model {
//	    m.FocusableModel = m.FocusableModel.SetFocused(focused)
//	    return m
//	}
type FocusableModel struct {
	focused    bool
	panelIndex int
}

// IsFocused returns whether the component is focused.
func (f FocusableModel) IsFocused() bool {
	return f.focused
}

// PanelIndex returns the panel index for Shift+N hints.
func (f FocusableModel) PanelIndex() int {
	return f.panelIndex
}

// SetFocused sets the focus state, returning a new FocusableModel.
func (f FocusableModel) SetFocused(focused bool) FocusableModel {
	f.focused = focused
	return f
}

// SetPanelIndex sets the panel index, returning a new FocusableModel.
func (f FocusableModel) SetPanelIndex(index int) FocusableModel {
	f.panelIndex = index
	return f
}

// PanelModel combines SizableModel and FocusableModel for components
// that need both size and focus management.
//
// Example usage:
//
//	type Model struct {
//	    base.PanelModel
//	    // other fields
//	}
type PanelModel struct {
	SizableModel
	FocusableModel
}

// SetSize sets the width and height, returning a new PanelModel.
func (p PanelModel) SetSize(width, height int) PanelModel {
	p.SizableModel = p.SizableModel.SetSize(width, height)
	return p
}

// SetFocused sets the focus state, returning a new PanelModel.
func (p PanelModel) SetFocused(focused bool) PanelModel {
	p.FocusableModel = p.FocusableModel.SetFocused(focused)
	return p
}

// SetPanelIndex sets the panel index, returning a new PanelModel.
func (p PanelModel) SetPanelIndex(index int) PanelModel {
	p.FocusableModel = p.FocusableModel.SetPanelIndex(index)
	return p
}
