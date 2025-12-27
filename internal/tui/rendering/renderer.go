// Package rendering provides superfile-style bordered panels with embedded titles.
package rendering

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Renderer creates bordered panels with embedded titles and sections.
type Renderer struct {
	width       int
	height      int
	title       string
	badge       string // Separate badge section in title bar (superfile style)
	footer      string
	footerBadge string // Separate badge section in footer (for cursor/scroll info)
	panelIndex  int    // 1-based panel index for Shift+N hotkey hint
	focused     bool
	sections    []section
	content     string
	border      lipgloss.Border
	focusColor  color.Color
	blurColor   color.Color
}

type section struct {
	name    string
	content string
}

// New creates a new Renderer with the given dimensions.
func New(width, height int) *Renderer {
	colors := theme.GetSemanticColors()
	return &Renderer{
		width:      width,
		height:     height,
		border:     lipgloss.RoundedBorder(),
		focusColor: colors.Highlight,
		blurColor:  colors.TableBorder,
	}
}

// SetTitle sets the title to embed in the top border.
// Title appears as: ├─ Title ─┤.
func (r *Renderer) SetTitle(title string) *Renderer {
	r.title = title
	return r
}

// SetBadge sets a badge that appears in a separate section after the title.
// Badge appears as: ├─ Title ─┼─ Badge ─┤ (superfile style).
func (r *Renderer) SetBadge(badge string) *Renderer {
	r.badge = badge
	return r
}

// SetFocused sets whether this panel has focus (affects border color).
func (r *Renderer) SetFocused(focused bool) *Renderer {
	r.focused = focused
	return r
}

// SetFocusColor sets the border color when focused.
func (r *Renderer) SetFocusColor(c color.Color) *Renderer {
	r.focusColor = c
	return r
}

// SetBlurColor sets the border color when not focused.
func (r *Renderer) SetBlurColor(c color.Color) *Renderer {
	r.blurColor = c
	return r
}

// SetBorder sets the border style.
func (r *Renderer) SetBorder(b lipgloss.Border) *Renderer {
	r.border = b
	return r
}

// AddSection adds a named section with divider.
// Creates: ├─ SectionName ─┤ followed by content.
func (r *Renderer) AddSection(name, content string) *Renderer {
	r.sections = append(r.sections, section{name: name, content: content})
	return r
}

// SetContent sets the main content (no section header).
func (r *Renderer) SetContent(content string) *Renderer {
	r.content = content
	return r
}

// SetFooter sets the footer text to embed in the bottom border.
// Footer appears as: ├─ footer text ─┤.
func (r *Renderer) SetFooter(footer string) *Renderer {
	r.footer = footer
	return r
}

// SetFooterBadge sets a badge that appears in a separate section in the footer.
// FooterBadge appears as: ├─ footer ─┼─ badge ─┤ (between footer and hint).
func (r *Renderer) SetFooterBadge(badge string) *Renderer {
	r.footerBadge = badge
	return r
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
// When the panel is not focused, shows "Shift+N" in the bottom-right corner.
func (r *Renderer) SetPanelIndex(index int) *Renderer {
	r.panelIndex = index
	return r
}

// Render produces the final bordered output.
func (r *Renderer) Render() string {
	if r.width < 5 || r.height < 3 {
		return ""
	}

	borderColor := r.blurColor
	if r.focused {
		borderColor = r.focusColor
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Content area dimensions
	contentWidth := r.width - 2   // Account for left and right borders
	contentHeight := r.height - 2 // Account for top and bottom borders

	// Build lines with estimated capacity (top + content + bottom)
	lines := make([]string, 0, r.height)

	// Top border with embedded title (and optional badge in superfile style)
	var topBorder string
	if r.badge != "" {
		topBorder = BuildTopBorderWithBadge(r.width, r.title, r.badge, r.border)
	} else {
		topBorder = BuildTopBorder(r.width, r.title, r.border)
	}
	lines = append(lines, borderStyle.Render(topBorder))

	// Build content lines with estimated capacity
	contentLines := make([]string, 0, contentHeight)

	// Add main content if set
	if r.content != "" {
		contentLines = append(contentLines, r.wrapAndTruncate(r.content, contentWidth)...)
	}

	// Add sections
	for _, sec := range r.sections {
		if len(contentLines) > 0 {
			contentLines = append(contentLines, "") // Empty line before section
		}
		// Section divider
		divider := BuildDivider(r.width, sec.name, r.border)
		contentLines = append(contentLines, borderStyle.Render(divider[1:len(divider)-1])) // Remove border chars for inline
		// Section content
		if sec.content != "" {
			contentLines = append(contentLines, r.wrapAndTruncate(sec.content, contentWidth)...)
		}
	}

	// Render content within borders with 1 char padding on each side
	leftBorder := borderStyle.Render(r.border.Left) + " "
	rightBorder := " " + borderStyle.Render(r.border.Right)
	paddedWidth := contentWidth - 2 // Account for the 1-char padding on each side

	for i := range contentHeight {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Pad line to padded width
		lineWidth := ansi.StringWidth(line)
		if lineWidth < paddedWidth {
			line += strings.Repeat(" ", paddedWidth-lineWidth)
		} else if lineWidth > paddedWidth {
			line = ansi.Truncate(line, paddedWidth-3, "...")
		}
		lines = append(lines, leftBorder+line+rightBorder)
	}

	// Bottom border with optional footer, footerBadge, and panel hint
	var bottomBorder string
	hint := r.buildPanelHint()
	if r.footer != "" || r.footerBadge != "" || hint != "" {
		bottomBorder = BuildBottomBorderWithFooterBadgeAndHint(r.width, r.footer, r.footerBadge, hint, r.border)
	} else {
		bottomBorder = BuildBottomBorder(r.width, r.border)
	}
	lines = append(lines, borderStyle.Render(bottomBorder))

	return strings.Join(lines, "\n")
}

// wrapAndTruncate splits content into lines and truncates each to fit width.
func (r *Renderer) wrapAndTruncate(content string, width int) []string {
	inputLines := strings.Split(content, "\n")
	var result []string

	for _, line := range inputLines {
		lineWidth := ansi.StringWidth(line)
		if lineWidth <= width {
			result = append(result, line)
		} else {
			// Truncate with ellipsis
			result = append(result, ansi.Truncate(line, width-3, "...")+"...")
		}
	}

	return result
}

// ContentWidth returns the usable content width inside the borders (including padding).
func (r *Renderer) ContentWidth() int {
	return r.width - 4 // -2 for borders, -2 for padding
}

// ContentHeight returns the usable content height inside the borders.
func (r *Renderer) ContentHeight() int {
	return r.height - 2
}

// buildPanelHint returns the Shift+N hint string if applicable.
// Only shows hint when panel is not focused and has a valid index.
func (r *Renderer) buildPanelHint() string {
	if r.focused || r.panelIndex < 1 || r.panelIndex > 9 {
		return ""
	}
	return fmt.Sprintf("⇧%d", r.panelIndex)
}
