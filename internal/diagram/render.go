package diagram

import (
	"fmt"
	"strings"
)

// Renderer renders a wiring diagram for a device model.
type Renderer interface {
	Render(model DeviceModel) string
}

// NewRenderer returns a Renderer for the given style.
func NewRenderer(style Style) Renderer {
	switch style {
	case StyleCompact:
		return &compactRenderer{}
	case StyleDetailed:
		return &detailedRenderer{}
	default:
		return &schematicRenderer{}
	}
}

// header returns a formatted header block for a device model.
func header(m DeviceModel) string {
	gen := fmt.Sprintf("Gen%d", m.Generation)
	return fmt.Sprintf("  %s (%s)\n  %s", m.Name, gen, strings.Repeat("─", len(m.Name)+len(gen)+3))
}

// inputCount returns the number of digital inputs for an InputOnly topology device.
func inputCount(m DeviceModel) int {
	if strings.Contains(strings.ToLower(m.Slug), "i4") {
		return 4
	}
	return 3
}

// emChannelCount returns the number of CT clamp channels for an EnergyMonitor device.
func emChannelCount(m DeviceModel) int {
	if strings.Contains(strings.ToLower(m.Slug), "3em") {
		return 3
	}
	return 2
}

// formatAmps formats an amperage value, omitting unnecessary decimal places.
func formatAmps(a float64) string {
	return fmt.Sprintf("%gA", a)
}

// specsFooter returns a formatted specs block for a device model.
func specsFooter(m DeviceModel) string {
	lines := []string{
		"  Specs:",
		fmt.Sprintf("    Voltage:    %s", m.Specs.Voltage),
	}
	if m.Specs.MaxAmps > 0 {
		ampsStr := formatAmps(m.Specs.MaxAmps)
		if m.Specs.MaxAmpsTotal > 0 {
			totalStr := formatAmps(m.Specs.MaxAmpsTotal)
			lines = append(lines, fmt.Sprintf("    Max Amps:   %s per channel (%s total)", ampsStr, totalStr))
		} else {
			lines = append(lines, fmt.Sprintf("    Max Amps:   %s", ampsStr))
		}
	}
	if m.Specs.PowerMonitoring {
		lines = append(lines, "    Metering:   Yes")
	}
	if m.Specs.NeutralRequired {
		lines = append(lines, "    Neutral:    Required")
	} else {
		lines = append(lines, "    Neutral:    Optional")
	}
	if m.Specs.Notes != "" {
		lines = append(lines, fmt.Sprintf("    Note:       %s", m.Specs.Notes))
	}
	return strings.Join(lines, "\n")
}

// detailedWidth returns the interior box width for a device name in the detailed renderer.
// Minimum 29 to fit content.
func detailedWidth(name string) int {
	w := len(name) + 10 // 9 spaces prefix + 1 space suffix minimum
	if w < 29 {
		w = 29
	}
	return w
}

// powerSupplyCol is the column where the power supply vertical │ sits.
// All detailed topologies align their ┴ junction to this column.
const powerSupplyCol = 29

// detailedTopBorder returns a top border with a ┴ junction aligned to
// the power supply vertical at column 29.
func detailedTopBorder(indent string, width int) string {
	left := powerSupplyCol - len(indent) - 1 // dashes before ┴
	right := width - 1 - left                // dashes after ┴
	return fmt.Sprintf("%s┌%s┴%s┐\n", indent,
		strings.Repeat("─", left),
		strings.Repeat("─", right))
}

// detailedLine returns a padded interior line for a detailed box.
// Content is left-aligned and padded with spaces to fill the interior width.
func detailedLine(indent string, width int, content string) string {
	return fmt.Sprintf("%s│%-*s│\n", indent, width, content)
}
