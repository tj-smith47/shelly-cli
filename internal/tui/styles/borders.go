// Package styles provides shared TUI style helpers.
package styles

import (
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// PanelBorder returns a style with a rounded border using the standard table border color.
// Use this for inactive/unfocused panels.
func PanelBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.TableBorder)
}

// PanelBorderActive returns a style with a rounded border using the highlight color.
// Use this for active/focused panels.
func PanelBorderActive() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Highlight)
}

// PanelBorderFocused returns the appropriate panel border style based on focus state.
func PanelBorderFocused(focused bool) lipgloss.Style {
	if focused {
		return PanelBorderActive()
	}
	return PanelBorder()
}

// ModalBorder returns a style with a rounded border suitable for modals.
// Uses the highlight color for visibility.
func ModalBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Highlight)
}

// ErrorBorder returns a style with a rounded border using the error color.
// Use this for error states or destructive action confirmations.
func ErrorBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Error)
}

// InfoBorder returns a style with a rounded border using the info color.
// Use this for informational messages and hints.
func InfoBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Info)
}

// SuccessBorder returns a style with a rounded border using the success color.
// Use this for success messages and confirmations.
func SuccessBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Success)
}

// WarningBorder returns a style with a rounded border using the warning color.
// Use this for warning messages and cautions.
func WarningBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Warning)
}

// MutedBorder returns a style with a rounded border using the muted color.
// Use this for unfocused inputs or secondary elements.
func MutedBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Muted)
}

// PrimaryBorder returns a style with a rounded border using the primary color.
// Use this for primary interactive elements like command mode.
func PrimaryBorder() lipgloss.Style {
	colors := theme.GetSemanticColors()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.Primary)
}
