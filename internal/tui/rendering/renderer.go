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

	borderStyle := r.getBorderStyle()
	contentWidth := r.width - 2
	contentHeight := r.height - 2

	lines := make([]string, 0, r.height)

	// Top border - pass border style to handle styled badges properly
	lines = append(lines, r.buildTopBorder(borderStyle))

	// Content with borders and padding
	contentLines := r.buildContentLines(contentWidth, borderStyle)
	lines = append(lines, r.renderContentWithBorders(contentLines, contentWidth, contentHeight, borderStyle)...)

	// Bottom border - handle styled footer content properly
	lines = append(lines, r.buildBottomBorderStyled(borderStyle))

	return strings.Join(lines, "\n")
}

// getBorderStyle returns the appropriate border style based on focus.
func (r *Renderer) getBorderStyle() lipgloss.Style {
	borderColor := r.blurColor
	if r.focused {
		borderColor = r.focusColor
	}
	return lipgloss.NewStyle().Foreground(borderColor)
}

// buildTopBorder constructs the top border with optional title and badge.
// When a badge contains styled text, the border style is needed to re-color
// the border parts after the badge.
func (r *Renderer) buildTopBorder(borderStyle lipgloss.Style) string {
	if r.badge != "" {
		return BuildTopBorderWithBadge(r.width, r.title, r.badge, r.border, borderStyle)
	}
	return borderStyle.Render(BuildTopBorder(r.width, r.title, r.border))
}

// buildBottomBorderStyled constructs the bottom border with proper handling of styled footer content.
// If the footer contains ANSI codes (styled text), it renders border parts separately to prevent
// style leakage from resetting the border color.
func (r *Renderer) buildBottomBorderStyled(borderStyle lipgloss.Style) string {
	hint := r.buildPanelHint()

	// If footer might contain styled text (ANSI codes), handle it specially
	if r.footer != "" && strings.Contains(r.footer, "\x1b[") {
		return BuildBottomBorderWithStyledFooter(r.width, r.footer, hint, r.border, borderStyle)
	}

	// No styled content - use simple render
	if r.footer != "" || r.footerBadge != "" || hint != "" {
		return borderStyle.Render(BuildBottomBorderWithFooterBadgeAndHint(r.width, r.footer, r.footerBadge, hint, r.border))
	}
	return borderStyle.Render(BuildBottomBorder(r.width, r.border))
}

// buildContentLines assembles the main content and sections.
func (r *Renderer) buildContentLines(contentWidth int, borderStyle lipgloss.Style) []string {
	lines := make([]string, 0)

	if r.content != "" {
		lines = append(lines, r.wrapAndTruncate(r.content, contentWidth)...)
	}

	for _, sec := range r.sections {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		divider := BuildDivider(r.width, sec.name, r.border)
		lines = append(lines, borderStyle.Render(divider[1:len(divider)-1]))
		if sec.content != "" {
			lines = append(lines, r.wrapAndTruncate(sec.content, contentWidth)...)
		}
	}

	return lines
}

// renderContentWithBorders wraps content lines with borders and padding.
// Returns exactly contentHeight lines to ensure panels fit their allocated space.
// Adds 1-line top and bottom padding when space allows (contentHeight >= 3).
func (r *Renderer) renderContentWithBorders(contentLines []string, contentWidth, contentHeight int, borderStyle lipgloss.Style) []string {
	if contentHeight <= 0 {
		return []string{}
	}

	leftBorder := borderStyle.Render(r.border.Left) + " "
	rightBorder := " " + borderStyle.Render(r.border.Right)
	paddedWidth := contentWidth - 2

	emptyLine := leftBorder + strings.Repeat(" ", paddedWidth) + rightBorder

	lines := make([]string, 0, contentHeight)

	// Determine padding based on available height
	// >= 3 lines: top padding + content + bottom padding
	// == 2 lines: top padding + content
	// == 1 line:  content only
	topPadding := 0
	bottomPadding := 0
	if contentHeight >= 2 {
		topPadding = 1
	}
	if contentHeight >= 3 {
		bottomPadding = 1
	}

	// Add top padding
	for range topPadding {
		lines = append(lines, emptyLine)
	}

	// Add content lines (between padding)
	contentSpace := contentHeight - topPadding - bottomPadding
	for i := range contentSpace {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		line = r.padOrTruncate(line, paddedWidth)
		lines = append(lines, leftBorder+line+rightBorder)
	}

	// Add bottom padding
	for range bottomPadding {
		lines = append(lines, emptyLine)
	}

	return lines
}

// padOrTruncate adjusts line to exact width.
func (r *Renderer) padOrTruncate(line string, width int) string {
	lineWidth := ansi.StringWidth(line)
	if lineWidth < width {
		return line + strings.Repeat(" ", width-lineWidth)
	}
	if lineWidth > width {
		return ansi.Truncate(line, width-3, "...")
	}
	return line
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

// ContentHeight returns the usable content height inside the borders (including padding).
func (r *Renderer) ContentHeight() int {
	return r.height - 4 // -2 for borders, -2 for vertical padding
}

// buildPanelHint returns the Shift+N hint string if applicable.
// Only shows hint when panel is not focused and has a valid index.
func (r *Renderer) buildPanelHint() string {
	if r.focused || r.panelIndex < 1 || r.panelIndex > 9 {
		return ""
	}
	return fmt.Sprintf("⇧%d", r.panelIndex)
}

// NewModal creates a modal-style renderer optimized for full-screen overlays.
// Uses the provided dimensions directly with minimum sizes, focused border, and highlight color.
// Callers should pass properly computed modal dimensions (typically 90% of screen).
// This is the standard pattern for all edit modals.
func NewModal(width, height int, title, footer string) *Renderer {
	// Apply minimum sizes
	if width < 60 {
		width = 60
	}
	if height < 20 {
		height = 20
	}

	colors := theme.GetSemanticColors()
	return New(width, height).
		SetTitle(title).
		SetFocused(true).
		SetFocusColor(colors.Highlight).
		SetFooter(footer)
}

// ModalInputWidth returns the recommended input width for modals given modal dimensions.
// This ensures inputs are appropriately sized without arbitrary limits.
// The modalWidth parameter should be the actual modal width (not screen width).
func ModalInputWidth(modalWidth int) int {
	inputWidth := modalWidth - 20 // Account for borders, padding, labels
	if inputWidth < 40 {
		inputWidth = 40
	}
	return inputWidth
}
