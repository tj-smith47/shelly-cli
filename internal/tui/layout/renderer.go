// Package layout provides flexible panel sizing and rendering utilities for the TUI.
package layout

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Renderer provides consistent panel rendering with borders, titles, and footers.
// Inspired by superfile's rendering pattern.
type Renderer struct {
	width         int
	height        int
	title         string
	lines         []string
	footerItems   []string
	focused       bool
	borderStyle   lipgloss.Border
	focusedColor  lipgloss.Color
	unfocusedColor lipgloss.Color
	scrollOffset  int
	totalLines    int
}

// NewRenderer creates a new panel renderer.
func NewRenderer(width, height int) *Renderer {
	colors := theme.GetSemanticColors()
	return &Renderer{
		width:         width,
		height:        height,
		borderStyle:   lipgloss.RoundedBorder(),
		focusedColor:  lipgloss.Color(colors.Highlight.RGBA()),
		unfocusedColor: lipgloss.Color(colors.TableBorder.RGBA()),
	}
}

// SetTitle sets the panel title.
func (r *Renderer) SetTitle(title string) *Renderer {
	r.title = title
	return r
}

// SetFocused sets whether the panel is focused.
func (r *Renderer) SetFocused(focused bool) *Renderer {
	r.focused = focused
	return r
}

// AddLine adds a single line of content.
func (r *Renderer) AddLine(line string) *Renderer {
	r.lines = append(r.lines, line)
	return r
}

// AddLines adds multiple lines of content.
func (r *Renderer) AddLines(lines ...string) *Renderer {
	r.lines = append(r.lines, lines...)
	return r
}

// SetContent replaces all content with the given lines.
func (r *Renderer) SetContent(lines []string) *Renderer {
	r.lines = lines
	return r
}

// SetFooterItems sets the footer items (shown at bottom of panel).
func (r *Renderer) SetFooterItems(items ...string) *Renderer {
	r.footerItems = items
	return r
}

// SetScrollOffset sets the scroll position for content.
func (r *Renderer) SetScrollOffset(offset, total int) *Renderer {
	r.scrollOffset = offset
	r.totalLines = total
	return r
}

// ContentHeight returns the available height for content (excluding borders).
func (r *Renderer) ContentHeight() int {
	h := r.height - 2 // Top and bottom border
	if len(r.footerItems) > 0 {
		h-- // Footer line
	}
	if h < 0 {
		h = 0
	}
	return h
}

// ContentWidth returns the available width for content (excluding borders).
func (r *Renderer) ContentWidth() int {
	w := r.width - 4 // Left and right border + padding
	if w < 0 {
		w = 0
	}
	return w
}

// Render renders the panel with borders, title, content, and footer.
func (r *Renderer) Render() string {
	if r.width <= 0 || r.height <= 0 {
		return ""
	}

	// Determine border color
	borderColor := r.unfocusedColor
	if r.focused {
		borderColor = r.focusedColor
	}

	// Build content
	contentHeight := r.ContentHeight()
	contentWidth := r.ContentWidth()

	// Apply scroll offset and truncate to visible area
	visibleLines := r.getVisibleLines(contentHeight)

	// Pad/truncate each line to content width
	for i := range visibleLines {
		visibleLines[i] = truncateOrPad(visibleLines[i], contentWidth)
	}

	// Pad to fill content height
	for len(visibleLines) < contentHeight {
		visibleLines = append(visibleLines, strings.Repeat(" ", contentWidth))
	}

	content := strings.Join(visibleLines, "\n")

	// Build the panel style
	panelStyle := lipgloss.NewStyle().
		Border(r.borderStyle).
		BorderForeground(borderColor).
		Width(r.width - 2). // Account for border
		Height(r.height - 2)

	// Add title to border
	if r.title != "" {
		panelStyle = panelStyle.BorderTop(true)
	}

	rendered := panelStyle.Render(content)

	// Inject title into top border
	if r.title != "" {
		rendered = r.injectTitle(rendered)
	}

	// Inject footer into bottom border
	if len(r.footerItems) > 0 {
		rendered = r.injectFooter(rendered)
	}

	return rendered
}

// getVisibleLines returns the lines that should be visible given the scroll offset.
func (r *Renderer) getVisibleLines(maxLines int) []string {
	if len(r.lines) == 0 {
		return nil
	}

	start := r.scrollOffset
	if start < 0 {
		start = 0
	}
	if start >= len(r.lines) {
		start = len(r.lines) - 1
	}

	end := start + maxLines
	if end > len(r.lines) {
		end = len(r.lines)
	}

	return r.lines[start:end]
}

// injectTitle injects the title into the top border.
func (r *Renderer) injectTitle(rendered string) string {
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		return rendered
	}

	// Format title with delimiters
	titleStr := "├─ " + r.title + " ─┤"

	// Find position to inject (centered or left-aligned)
	topBorder := lines[0]
	topBorderRunes := []rune(topBorder)

	// Find a good position (after the corner)
	insertPos := 3 // After "╭──"
	titleRunes := []rune(titleStr)

	if insertPos+len(titleRunes) < len(topBorderRunes)-3 {
		// Replace portion of border with title
		for i, tr := range titleRunes {
			if insertPos+i < len(topBorderRunes) {
				topBorderRunes[insertPos+i] = tr
			}
		}
		lines[0] = string(topBorderRunes)
	}

	return strings.Join(lines, "\n")
}

// injectFooter injects footer items into the bottom border.
func (r *Renderer) injectFooter(rendered string) string {
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		return rendered
	}

	// Build footer string
	colors := theme.GetSemanticColors()
	footerStyle := lipgloss.NewStyle().Foreground(colors.Muted)
	footerStr := footerStyle.Render(strings.Join(r.footerItems, " │ "))

	// Get the last line (bottom border)
	lastIdx := len(lines) - 1
	bottomBorder := lines[lastIdx]
	bottomRunes := []rune(bottomBorder)
	footerRunes := []rune(footerStr)

	// Insert footer near the start of bottom border
	insertPos := 2
	if insertPos+len(footerRunes) < len(bottomRunes)-2 {
		for i, fr := range footerRunes {
			if insertPos+i < len(bottomRunes) {
				bottomRunes[insertPos+i] = fr
			}
		}
		lines[lastIdx] = string(bottomRunes)
	}

	return strings.Join(lines, "\n")
}

// truncateOrPad truncates or pads a string to exactly the given width.
func truncateOrPad(s string, width int) string {
	if width <= 0 {
		return ""
	}

	// Count visible width (accounting for ANSI codes)
	visWidth := lipgloss.Width(s)

	if visWidth > width {
		// Truncate - this is tricky with ANSI codes
		// For now, use lipgloss truncation
		return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(s)
	}

	// Pad with spaces
	padding := width - visWidth
	return s + strings.Repeat(" ", padding)
}

// QuickPanel is a convenience function to render a simple panel.
func QuickPanel(width, height int, title string, content string, focused bool) string {
	lines := strings.Split(content, "\n")
	return NewRenderer(width, height).
		SetTitle(title).
		SetContent(lines).
		SetFocused(focused).
		Render()
}

// QuickPanelWithFooter renders a panel with footer shortcuts.
func QuickPanelWithFooter(width, height int, title string, content string, focused bool, footer ...string) string {
	lines := strings.Split(content, "\n")
	return NewRenderer(width, height).
		SetTitle(title).
		SetContent(lines).
		SetFocused(focused).
		SetFooterItems(footer...).
		Render()
}
