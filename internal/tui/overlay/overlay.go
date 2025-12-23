// Package overlay provides utilities for positioning overlays in the TUI.
package overlay

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// PlaceOverlay positions an overlay on top of base content.
// The overlay is centered if centerX and centerY are true.
func PlaceOverlay(base, overlay string, centerX, centerY bool) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	baseHeight := len(baseLines)
	overlayHeight := len(overlayLines)

	// Calculate base width (max line width)
	baseWidth := 0
	for _, line := range baseLines {
		w := ansi.StringWidth(line)
		if w > baseWidth {
			baseWidth = w
		}
	}

	// Calculate overlay width
	overlayWidth := 0
	for _, line := range overlayLines {
		w := ansi.StringWidth(line)
		if w > overlayWidth {
			overlayWidth = w
		}
	}

	// Calculate position
	x, y := 0, 0
	if centerX {
		x = (baseWidth - overlayWidth) / 2
		if x < 0 {
			x = 0
		}
	}
	if centerY {
		y = (baseHeight - overlayHeight) / 2
		if y < 0 {
			y = 0
		}
	}

	return PlaceAt(base, overlay, x, y)
}

// PlaceAt positions an overlay at specific x, y coordinates.
func PlaceAt(base, overlay string, x, y int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Ensure base has enough lines
	for len(baseLines) < y+len(overlayLines) {
		baseLines = append(baseLines, "")
	}

	// Place each overlay line
	for i, overlayLine := range overlayLines {
		lineIdx := y + i
		if lineIdx < 0 || lineIdx >= len(baseLines) {
			continue
		}

		baseLine := baseLines[lineIdx]
		baseLines[lineIdx] = placeLineAt(baseLine, overlayLine, x)
	}

	return strings.Join(baseLines, "\n")
}

// placeLineAt places overlay text onto a base line at position x.
func placeLineAt(baseLine, overlayLine string, x int) string {
	if x < 0 {
		x = 0
	}

	baseWidth := ansi.StringWidth(baseLine)
	overlayWidth := ansi.StringWidth(overlayLine)

	// Extend base line if needed
	if baseWidth < x {
		baseLine += strings.Repeat(" ", x-baseWidth)
		baseWidth = x
	}

	// Build the result
	var result strings.Builder

	// Content before overlay
	if x > 0 {
		prefix := truncateToWidth(baseLine, x)
		result.WriteString(prefix)
	}

	// Add overlay content
	result.WriteString(overlayLine)

	// Add remaining base content after overlay
	afterPos := x + overlayWidth
	if afterPos < baseWidth {
		suffix := skipToPosition(baseLine, afterPos)
		result.WriteString(suffix)
	}

	return result.String()
}

// truncateToWidth returns the string truncated to the given visual width.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	currentWidth := 0
	var result strings.Builder

	for _, r := range s {
		rWidth := ansi.StringWidth(string(r))
		if currentWidth+rWidth > width {
			break
		}
		result.WriteRune(r)
		currentWidth += rWidth
	}

	// Pad if needed
	if currentWidth < width {
		result.WriteString(strings.Repeat(" ", width-currentWidth))
	}

	return result.String()
}

// skipToPosition skips to the given position and returns the rest.
func skipToPosition(s string, pos int) string {
	currentWidth := 0

	for i, r := range s {
		rWidth := ansi.StringWidth(string(r))
		if currentWidth >= pos {
			return s[i:]
		}
		currentWidth += rWidth
	}

	return ""
}

// Center calculates center position for an overlay.
func Center(baseWidth, baseHeight, overlayWidth, overlayHeight int) (x, y int) {
	x = (baseWidth - overlayWidth) / 2
	if x < 0 {
		x = 0
	}
	y = (baseHeight - overlayHeight) / 2
	if y < 0 {
		y = 0
	}
	return x, y
}

// Dimensions returns the width and height of a string.
func Dimensions(s string) (width, height int) {
	lines := strings.Split(s, "\n")
	height = len(lines)
	for _, line := range lines {
		w := ansi.StringWidth(line)
		if w > width {
			width = w
		}
	}
	return width, height
}
