package editmodal

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Base provides common infrastructure for edit modals.
// Embed this struct in modal models to eliminate lifecycle, navigation,
// key handling, save, and render boilerplate. Each modal still handles
// its own focus/blur (typed form fields) and domain-specific logic.
type Base struct {
	Ctx    context.Context
	Svc    *shelly.Service
	Device string
	Styles Styles

	Cursor     int
	FieldCount int
	Saving     bool
	Err        error
	Width      int
	Height     int

	visible   bool
	scrollOff int
}

// Show resets the modal state and makes it visible.
// Modals call this in their Show method after setting domain-specific fields.
func (b *Base) Show(device string, fieldCount int) {
	b.Device = device
	b.FieldCount = fieldCount
	b.Cursor = 0
	b.Saving = false
	b.Err = nil
	b.visible = true
	b.scrollOff = 0
}

// Hide makes the modal invisible.
func (b *Base) Hide() {
	b.visible = false
}

// Visible returns whether the modal is currently shown.
func (b *Base) Visible() bool {
	return b.visible
}

// SetSize updates the modal dimensions.
func (b *Base) SetSize(w, h int) {
	b.Width = w
	b.Height = h
}

// SetErr sets a validation or save error on the modal.
func (b *Base) SetErr(err error) {
	b.Err = err
}

// ClearErr clears any current error.
func (b *Base) ClearErr() {
	b.Err = nil
}

// InputWidth returns the recommended input width for text fields.
func (b *Base) InputWidth() int {
	return rendering.ModalInputWidth(b.Width)
}

// ContentHeight returns the available lines for field content
// inside the modal chrome (borders + padding).
func (b *Base) ContentHeight() int {
	// Modal has 2 border lines + 2 padding lines = 4 lines of chrome
	h := b.Height - 4
	if h < 0 {
		return 0
	}
	return h
}
