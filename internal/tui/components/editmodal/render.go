package editmodal

import (
	"github.com/tj-smith47/shelly-cli/internal/tui/components/errorview"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// RenderField renders a single field row with cursor indicator.
// When the cursor matches fieldIndex, renders "▶ Label value",
// otherwise renders "  Label value".
func (b *Base) RenderField(fieldIndex int, label, value string) string {
	selected := b.Cursor == fieldIndex
	return b.Styles.RenderFieldRow(selected, label, value)
}

// RenderModal wraps content in the standard modal frame with title and footer.
func (b *Base) RenderModal(title, content, footer string) string {
	r := rendering.NewModal(b.Width, b.Height, title, footer)
	return r.SetContent(content).Render()
}

// RenderError returns a styled error string if Err is set, otherwise "".
func (b *Base) RenderError() string {
	if b.Err == nil {
		return ""
	}
	return errorview.RenderInline(b.Err)
}

// RenderSavingFooter returns "Saving..." if a save is in progress,
// otherwise returns the provided normalFooter.
func (b *Base) RenderSavingFooter(normalFooter string) string {
	if b.Saving {
		return "Saving..."
	}
	return normalFooter
}
