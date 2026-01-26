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

// specsFooter returns a formatted specs block for a device model.
func specsFooter(m DeviceModel) string {
	lines := []string{
		"  Specs:",
		fmt.Sprintf("    Voltage:    %s", m.Specs.Voltage),
	}
	if m.Specs.MaxAmps > 0 {
		lines = append(lines, fmt.Sprintf("    Max Amps:   %.0fA", m.Specs.MaxAmps))
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
