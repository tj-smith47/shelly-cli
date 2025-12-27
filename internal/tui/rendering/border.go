package rendering

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// charWidth returns the visual width of a string (using ansi.StringWidth).
func charWidth(s string) int {
	return ansi.StringWidth(s)
}

// BuildTopBorder creates a top border with an embedded title.
// Example output: "╭──├─ Title ─┤────────────────╮".
func BuildTopBorder(width int, title string, border lipgloss.Border) string {
	if width < 5 {
		return ""
	}

	topLeft := border.TopLeft
	topRight := border.TopRight
	top := border.Top
	midLeft := border.MiddleLeft
	midRight := border.MiddleRight

	topLeftW := charWidth(topLeft)
	topRightW := charWidth(topRight)
	topW := charWidth(top)
	midLeftW := charWidth(midLeft)
	midRightW := charWidth(midRight)

	if title == "" {
		// No title - simple top border
		fillWidth := width - topLeftW - topRightW
		if fillWidth < 0 {
			fillWidth = 0
		}
		return topLeft + strings.Repeat(top, fillWidth/topW) + topRight
	}

	// Title with padding: ├─ Title ─┤
	titleText := " " + title + " "
	titleWidth := charWidth(titleText)

	// Calculate available space
	// Structure: TopLeft + fill + midLeft + ─ + titleText + ─ + midRight + fill + TopRight
	minTitleWidth := midLeftW + topW + titleWidth + topW + midRightW // ├─ Title ─┤
	availableForFill := width - topLeftW - topRightW - minTitleWidth

	if availableForFill < 0 {
		// Title too long, truncate it
		maxTitleLen := width - topLeftW - topRightW - midLeftW - midRightW - topW*2 - 2
		if maxTitleLen < 3 {
			// Can't fit any title
			fillWidth := width - topLeftW - topRightW
			if fillWidth < 0 {
				fillWidth = 0
			}
			return topLeft + strings.Repeat(top, fillWidth/topW) + topRight
		}
		title = ansi.Truncate(title, maxTitleLen, "..")
		titleText = " " + title + " "
		titleWidth = charWidth(titleText)
		minTitleWidth = midLeftW + topW + titleWidth + topW + midRightW
		availableForFill = width - topLeftW - topRightW - minTitleWidth
	}

	// Distribute fill space: minimal left padding to keep title near corner
	// Pattern: ╭─├─ Title ─┤────────────────────────╮
	totalFill := availableForFill / topW
	leftFillCount := 1 // Single dash before title for visual balance
	rightFillCount := totalFill - leftFillCount

	if rightFillCount < 0 {
		rightFillCount = 0
	}

	return topLeft +
		strings.Repeat(top, leftFillCount) +
		midLeft + top + titleText + top + midRight +
		strings.Repeat(top, rightFillCount) +
		topRight
}

// BuildTopBorderWithBadge creates a top border with a title and a separate badge section.
// Example output: "╭─├─ Devices ─┼─ 17/18 ─┤────╮" (superfile style).
func BuildTopBorderWithBadge(width int, title, badge string, border lipgloss.Border) string {
	if width < 5 {
		return ""
	}

	// If no badge, fall back to regular title border
	if badge == "" {
		return BuildTopBorder(width, title, border)
	}

	topLeft := border.TopLeft
	topRight := border.TopRight
	top := border.Top
	midLeft := border.MiddleLeft
	midRight := border.MiddleRight
	middle := border.Middle // ┼ for separator between title and badge

	topLeftW := charWidth(topLeft)
	topRightW := charWidth(topRight)
	topW := charWidth(top)
	midLeftW := charWidth(midLeft)
	midRightW := charWidth(midRight)
	middleW := charWidth(middle)

	// Title section: ├─ Title ─┼
	titleText := " " + title + " "
	titleWidth := charWidth(titleText)

	// Badge section: ─ badge ─┤
	badgeText := " " + badge + " "
	badgeWidth := charWidth(badgeText)

	// Total structure: TopLeft + fill + midLeft + ─ + title + ─ + middle + ─ + badge + ─ + midRight + fill + TopRight
	minWidth := midLeftW + topW + titleWidth + topW + middleW + topW + badgeWidth + topW + midRightW
	availableForFill := width - topLeftW - topRightW - minWidth

	if availableForFill < 0 {
		// Not enough space for badge, fall back to title only
		return BuildTopBorder(width, title+" "+badge, border)
	}

	// Minimal left padding, rest on right
	totalFill := availableForFill / topW
	leftFillCount := 1
	rightFillCount := totalFill - leftFillCount

	if rightFillCount < 0 {
		rightFillCount = 0
	}

	return topLeft +
		strings.Repeat(top, leftFillCount) +
		midLeft + top + titleText + top + middle + top + badgeText + top + midRight +
		strings.Repeat(top, rightFillCount) +
		topRight
}

// BuildBottomBorder creates a bottom border.
// Example output: "╰─────────────────────────────╯".
func BuildBottomBorder(width int, border lipgloss.Border) string {
	if width < 2 {
		return ""
	}

	bottomLeft := border.BottomLeft
	bottomRight := border.BottomRight
	bottom := border.Bottom

	bottomLeftW := charWidth(bottomLeft)
	bottomRightW := charWidth(bottomRight)
	bottomW := charWidth(bottom)

	fillWidth := width - bottomLeftW - bottomRightW
	if fillWidth < 0 {
		fillWidth = 0
	}

	return bottomLeft + strings.Repeat(bottom, fillWidth/bottomW) + bottomRight
}

// BuildBottomBorderWithFooter creates a bottom border with embedded footer text.
// Example output: "╰───├─ e:edit r:run ─┤───────╯".
func BuildBottomBorderWithFooter(width int, footer string, border lipgloss.Border) string {
	if width < 5 {
		return ""
	}

	bottomLeft := border.BottomLeft
	bottomRight := border.BottomRight
	bottom := border.Bottom
	midLeft := border.MiddleLeft
	midRight := border.MiddleRight

	bottomLeftW := charWidth(bottomLeft)
	bottomRightW := charWidth(bottomRight)
	bottomW := charWidth(bottom)
	midLeftW := charWidth(midLeft)
	midRightW := charWidth(midRight)

	if footer == "" {
		// No footer - simple bottom border
		return BuildBottomBorder(width, border)
	}

	// Footer with padding: ├─ footer ─┤
	footerText := " " + footer + " "
	footerWidth := charWidth(footerText)

	// Calculate available space
	// Structure: BottomLeft + fill + midLeft + ─ + footerText + ─ + midRight + fill + BottomRight
	minFooterWidth := midLeftW + bottomW + footerWidth + bottomW + midRightW // ├─ footer ─┤
	availableForFill := width - bottomLeftW - bottomRightW - minFooterWidth

	if availableForFill < 0 {
		// Footer too long, truncate it
		maxFooterLen := width - bottomLeftW - bottomRightW - midLeftW - midRightW - bottomW*2 - 2
		if maxFooterLen < 3 {
			// Can't fit any footer
			return BuildBottomBorder(width, border)
		}
		footer = ansi.Truncate(footer, maxFooterLen, "..")
		footerText = " " + footer + " "
		footerWidth = charWidth(footerText)
		minFooterWidth = midLeftW + bottomW + footerWidth + bottomW + midRightW
		availableForFill = width - bottomLeftW - bottomRightW - minFooterWidth
	}

	// Distribute fill space evenly
	leftFillCount := availableForFill / bottomW / 2
	rightFillCount := (availableForFill / bottomW) - leftFillCount

	if leftFillCount < 1 {
		leftFillCount = 1
		rightFillCount = (availableForFill / bottomW) - leftFillCount
	}
	if rightFillCount < 0 {
		rightFillCount = 0
	}

	return bottomLeft +
		strings.Repeat(bottom, leftFillCount) +
		midLeft + bottom + footerText + bottom + midRight +
		strings.Repeat(bottom, rightFillCount) +
		bottomRight
}

// BuildDivider creates a section divider with embedded name.
// Example output: "├─ Section ─┤─────────────────┤".
func BuildDivider(width int, name string, border lipgloss.Border) string {
	if width < 5 {
		return ""
	}

	left := border.MiddleLeft
	right := border.MiddleRight
	fill := border.Top // Use top border char for horizontal fill

	leftW := charWidth(left)
	rightW := charWidth(right)
	fillW := charWidth(fill)

	if name == "" {
		// No name - simple horizontal divider
		fillWidth := width - leftW - rightW
		if fillWidth < 0 {
			fillWidth = 0
		}
		return left + strings.Repeat(fill, fillWidth/fillW) + right
	}

	// Name with padding: ├─ Name ─┤
	nameText := " " + name + " "
	nameWidth := charWidth(nameText)

	// Calculate available space
	// Structure: left + ─ + nameText + ─ + right + fill
	minNameWidth := fillW + nameWidth + fillW + rightW // ─ Name ─┤
	availableForFill := width - leftW - minNameWidth

	if availableForFill < 0 {
		// Name too long, truncate it
		maxNameLen := width - leftW - rightW - fillW*2 - 2
		if maxNameLen < 3 {
			fillWidth := width - leftW - rightW
			if fillWidth < 0 {
				fillWidth = 0
			}
			return left + strings.Repeat(fill, fillWidth/fillW) + right
		}
		name = ansi.Truncate(name, maxNameLen, "..")
		nameText = " " + name + " "
		nameWidth = charWidth(nameText)
		minNameWidth = fillW + nameWidth + fillW + rightW
		availableForFill = width - leftW - minNameWidth
	}

	if availableForFill < 0 {
		availableForFill = 0
	}

	return left + fill + nameText + fill + right + strings.Repeat(fill, availableForFill/fillW)
}

// BuildSideBorders creates left and right border strings for a content line.
func BuildSideBorders(border lipgloss.Border) (left, right string) {
	return border.Left, border.Right
}

// BuildBottomBorderWithHint creates a bottom border with a hint on the right side.
// Example output: "╰─────────────────────├─ ⇧1 ─┤╯".
func BuildBottomBorderWithHint(width int, hint string, border lipgloss.Border) string {
	if width < 5 || hint == "" {
		return BuildBottomBorder(width, border)
	}

	bottomLeft := border.BottomLeft
	bottomRight := border.BottomRight
	bottom := border.Bottom
	midLeft := border.MiddleLeft
	midRight := border.MiddleRight

	bottomLeftW := charWidth(bottomLeft)
	bottomRightW := charWidth(bottomRight)
	bottomW := charWidth(bottom)
	midLeftW := charWidth(midLeft)
	midRightW := charWidth(midRight)

	// Hint with padding: ├─ hint ─┤
	hintText := " " + hint + " "
	hintWidth := charWidth(hintText)

	// Structure: BottomLeft + fill + midLeft + ─ + hintText + ─ + midRight + BottomRight
	// Hint goes on the right side
	minHintWidth := midLeftW + bottomW + hintWidth + bottomW + midRightW
	availableForFill := width - bottomLeftW - bottomRightW - minHintWidth

	if availableForFill < 0 {
		// Not enough space for hint
		return BuildBottomBorder(width, border)
	}

	// All fill goes on the left side (hint is right-aligned)
	leftFillCount := availableForFill / bottomW

	return bottomLeft +
		strings.Repeat(bottom, leftFillCount) +
		midLeft + bottom + hintText + bottom + midRight +
		bottomRight
}

// BuildBottomBorderWithFooterBadgeAndHint creates a bottom border with footer, badge, and hint sections.
// Example output: "╰─├─ footer ─┼─ badge ─┤───├─ ⇧1 ─┤╯".
func BuildBottomBorderWithFooterBadgeAndHint(width int, footer, badge, hint string, border lipgloss.Border) string {
	if width < 5 {
		return ""
	}

	// Handle degenerate cases
	if badge == "" {
		return BuildBottomBorderWithFooterAndHint(width, footer, hint, border)
	}
	if footer == "" && hint == "" {
		return BuildBottomBorderWithFooter(width, badge, border)
	}

	bottomLeft := border.BottomLeft
	bottomRight := border.BottomRight
	bottom := border.Bottom
	midLeft := border.MiddleLeft
	midRight := border.MiddleRight
	middle := border.Middle // ┼ for separator between footer and badge

	bottomLeftW := charWidth(bottomLeft)
	bottomRightW := charWidth(bottomRight)
	bottomW := charWidth(bottom)
	midLeftW := charWidth(midLeft)
	midRightW := charWidth(midRight)
	middleW := charWidth(middle)

	// Footer section: ├─ footer ─┼
	footerText := ""
	footerWidth := 0
	if footer != "" {
		footerText = " " + footer + " "
		footerWidth = charWidth(footerText)
	}

	// Badge section: ─ badge ─┤
	badgeText := " " + badge + " "
	badgeWidth := charWidth(badgeText)

	// Hint section: ├─ hint ─┤
	hintText := ""
	hintWidth := 0
	if hint != "" {
		hintText = " " + hint + " "
		hintWidth = charWidth(hintText)
	}

	// Calculate minimum widths
	var minWidth int
	if footer != "" {
		// Structure with footer: BottomLeft + fill + midLeft + ─ + footer + ─ + middle + ─ + badge + ─ + midRight + fill + [hint section] + BottomRight
		minWidth = midLeftW + bottomW + footerWidth + bottomW + middleW + bottomW + badgeWidth + bottomW + midRightW
	} else {
		// Structure without footer: BottomLeft + fill + midLeft + ─ + badge + ─ + midRight + fill + [hint section] + BottomRight
		minWidth = midLeftW + bottomW + badgeWidth + bottomW + midRightW
	}

	if hint != "" {
		minWidth += midLeftW + bottomW + hintWidth + bottomW + midRightW
	}

	availableForFill := width - bottomLeftW - bottomRightW - minWidth

	if availableForFill < 2 {
		// Not enough space, fall back to simpler version
		return BuildBottomBorderWithFooterAndHint(width, footer+" "+badge, hint, border)
	}

	// Minimal left fill (1), rest between sections
	leftFillCount := 1
	midFillCount := (availableForFill / bottomW) - leftFillCount

	if midFillCount < 0 {
		midFillCount = 0
	}

	var result string
	result = bottomLeft + strings.Repeat(bottom, leftFillCount)

	if footer != "" {
		// Footer + badge with separator
		result += midLeft + bottom + footerText + bottom + middle + bottom + badgeText + bottom + midRight
	} else {
		// Badge only
		result += midLeft + bottom + badgeText + bottom + midRight
	}

	result += strings.Repeat(bottom, midFillCount)

	if hint != "" {
		result += midLeft + bottom + hintText + bottom + midRight
	}

	result += bottomRight

	return result
}

// BuildBottomBorderWithFooterAndHint creates a bottom border with footer on left and hint on right.
// Example output: "╰─├─ footer ─┤─────────├─ ⇧1 ─┤╯".
func BuildBottomBorderWithFooterAndHint(width int, footer, hint string, border lipgloss.Border) string {
	if width < 5 {
		return ""
	}

	// Handle degenerate cases
	if footer == "" && hint == "" {
		return BuildBottomBorder(width, border)
	}
	if footer == "" {
		return BuildBottomBorderWithHint(width, hint, border)
	}
	if hint == "" {
		return BuildBottomBorderWithFooter(width, footer, border)
	}

	bottomLeft := border.BottomLeft
	bottomRight := border.BottomRight
	bottom := border.Bottom
	midLeft := border.MiddleLeft
	midRight := border.MiddleRight

	bottomLeftW := charWidth(bottomLeft)
	bottomRightW := charWidth(bottomRight)
	bottomW := charWidth(bottom)
	midLeftW := charWidth(midLeft)
	midRightW := charWidth(midRight)

	// Footer section: ├─ footer ─┤
	footerText := " " + footer + " "
	footerWidth := charWidth(footerText)

	// Hint section: ├─ hint ─┤
	hintText := " " + hint + " "
	hintWidth := charWidth(hintText)

	// Structure: BottomLeft + fill + midLeft + ─ + footer + ─ + midRight + fill + midLeft + ─ + hint + ─ + midRight + BottomRight
	minFooterWidth := midLeftW + bottomW + footerWidth + bottomW + midRightW
	minHintWidth := midLeftW + bottomW + hintWidth + bottomW + midRightW
	availableForFill := width - bottomLeftW - bottomRightW - minFooterWidth - minHintWidth

	if availableForFill < 2 {
		// Not enough space for both, prioritize hint
		return BuildBottomBorderWithHint(width, hint, border)
	}

	// Minimal left fill (1), rest between footer and hint
	leftFillCount := 1
	midFillCount := (availableForFill / bottomW) - leftFillCount

	if midFillCount < 0 {
		midFillCount = 0
	}

	return bottomLeft +
		strings.Repeat(bottom, leftFillCount) +
		midLeft + bottom + footerText + bottom + midRight +
		strings.Repeat(bottom, midFillCount) +
		midLeft + bottom + hintText + bottom + midRight +
		bottomRight
}
