package styles

import (
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// EmptyState renders a centered message for empty or no-data states.
// The message is centered both horizontally and vertically within the given dimensions.
// Width and height should be the content area dimensions (after accounting for borders).
func EmptyState(message string, width, height int) string {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Foreground(colors.Muted).
		Align(lipgloss.Center, lipgloss.Center).
		Render(message)
}

// EmptyStateWithBorder renders a centered message with standard border padding applied.
// Use this when passing panel dimensions (m.Width, m.Height) and the panel has borders.
// Subtracts 4 from width (2 border + 2 padding) and 4 from height (2 border + 2 padding)
// to match rendering.Renderer's ContentWidth/ContentHeight calculations.
func EmptyStateWithBorder(message string, width, height int) string {
	return EmptyState(message, width-4, height-4)
}

// NoDeviceSelected is a convenience function for the common "No device selected" message.
// Pass the panel's full dimensions (m.Width, m.Height).
func NoDeviceSelected(width, height int) string {
	return EmptyStateWithBorder("No device selected", width, height)
}

// NoDevicesOnline is a convenience function for the common "No devices online" message.
// Pass the panel's full dimensions (m.Width, m.Height).
func NoDevicesOnline(width, height int) string {
	return EmptyStateWithBorder("No devices online", width, height)
}

// NoDataAvailable is a convenience function for the common "No data available" message.
// Pass the panel's full dimensions (m.Width, m.Height).
func NoDataAvailable(width, height int) string {
	return EmptyStateWithBorder("No data available", width, height)
}

// NoItemsFound is a convenience function for empty list states.
// Pass the panel's full dimensions (m.Width, m.Height).
func NoItemsFound(itemType string, width, height int) string {
	return EmptyStateWithBorder("No "+itemType+" found", width, height)
}

// NoItemsConfigured is a convenience function for empty configuration states.
// Pass the panel's full dimensions (m.Width, m.Height).
func NoItemsConfigured(itemType string, width, height int) string {
	return EmptyStateWithBorder("No "+itemType+" configured", width, height)
}
