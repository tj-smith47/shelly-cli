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

	// Distribute fill space: more on right side for visual balance
	leftFillCount := availableForFill / topW / 3
	rightFillCount := (availableForFill / topW) - leftFillCount

	if leftFillCount < 1 {
		leftFillCount = 1
		rightFillCount = (availableForFill / topW) - leftFillCount
	}
	if rightFillCount < 0 {
		rightFillCount = 0
	}

	return topLeft +
		strings.Repeat(top, leftFillCount) +
		midLeft + top + titleText + top + midRight +
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
