// Package editmodal provides shared styles and utilities for edit modal components.
package editmodal

import (
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Styles contains the common styles used by edit modal components.
// Components can embed this struct and add component-specific styles.
type Styles struct {
	// Modal is the container style for the modal dialog.
	Modal lipgloss.Style

	// Title is the style for the modal title.
	Title lipgloss.Style

	// Label is the style for field labels (unfocused).
	Label lipgloss.Style

	// LabelFocus is the style for field labels when focused.
	LabelFocus lipgloss.Style

	// Error is the style for error messages.
	Error lipgloss.Style

	// Help is the style for help text and keybinding hints.
	Help lipgloss.Style

	// Selector is the style for the cursor indicator (▶).
	Selector lipgloss.Style

	// Warning is the style for warning messages.
	Warning lipgloss.Style

	// Info is the style for informational text.
	Info lipgloss.Style

	// Muted is the style for de-emphasized text.
	Muted lipgloss.Style

	// Overlay is the style for the backdrop behind the modal.
	Overlay lipgloss.Style

	// Input is the style for text inputs (unfocused).
	Input lipgloss.Style

	// InputFocus is the style for text inputs when focused.
	InputFocus lipgloss.Style

	// Button is the style for buttons (unfocused).
	Button lipgloss.Style

	// ButtonFocus is the style for buttons when focused.
	ButtonFocus lipgloss.Style

	// ButtonDanger is the style for destructive action buttons.
	ButtonDanger lipgloss.Style

	// Value is the style for displaying field values.
	Value lipgloss.Style

	// StatusOn is the style for enabled/on status indicators.
	StatusOn lipgloss.Style

	// StatusOff is the style for disabled/off status indicators.
	StatusOff lipgloss.Style

	// Selected is the style for selected/highlighted items in lists.
	Selected lipgloss.Style

	// Tab is the style for inactive tab labels.
	Tab lipgloss.Style

	// TabActive is the style for the active tab label.
	TabActive lipgloss.Style
}

// DefaultStyles returns the default edit modal styles.
// Components can use this directly or override specific fields.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Background(colors.Background).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Width(14),
		LabelFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Width(14),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selector: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Info: lipgloss.NewStyle().
			Foreground(colors.Info),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Overlay: lipgloss.NewStyle().
			Background(colors.Background),
		Input: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Muted).
			Padding(0, 1),
		InputFocus: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(0, 1),
		Button: lipgloss.NewStyle().
			Foreground(colors.Text).
			Background(colors.AltBackground).
			Padding(0, 2),
		ButtonFocus: lipgloss.NewStyle().
			Foreground(colors.Background).
			Background(colors.Highlight).
			Bold(true).
			Padding(0, 2),
		ButtonDanger: lipgloss.NewStyle().
			Foreground(colors.Background).
			Background(colors.Error).
			Bold(true).
			Padding(0, 2),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		StatusOn: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		StatusOff: lipgloss.NewStyle().
			Foreground(colors.Offline).
			Bold(true),
		Selected: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Background(colors.AltBackground),
		Tab: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Padding(0, 2),
		TabActive: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colors.Highlight),
	}
}

// DefaultLabelWidth is the standard width for field labels.
const DefaultLabelWidth = 14

// WithLabelWidth returns a copy of the styles with a different label width.
func (s Styles) WithLabelWidth(width int) Styles {
	s.Label = s.Label.Width(width)
	s.LabelFocus = s.LabelFocus.Width(width)
	return s
}

// LabelStyle returns the appropriate label style based on focus state.
func (s Styles) LabelStyle(focused bool) lipgloss.Style {
	if focused {
		return s.LabelFocus
	}
	return s.Label
}

// InputStyle returns the appropriate input style based on focus state.
func (s Styles) InputStyle(focused bool) lipgloss.Style {
	if focused {
		return s.InputFocus
	}
	return s.Input
}

// ButtonStyle returns the appropriate button style based on focus state.
func (s Styles) ButtonStyle(focused bool) lipgloss.Style {
	if focused {
		return s.ButtonFocus
	}
	return s.Button
}

// StatusStyle returns the appropriate status style based on enabled state.
func (s Styles) StatusStyle(enabled bool) lipgloss.Style {
	if enabled {
		return s.StatusOn
	}
	return s.StatusOff
}

// RenderSelector returns the selector string (▶ or empty) based on selection state.
func (s Styles) RenderSelector(selected bool) string {
	if selected {
		return s.Selector.Render("▶ ")
	}
	return "  "
}

// RenderLabel renders a label with the appropriate style based on focus.
func (s Styles) RenderLabel(label string, focused bool) string {
	return s.LabelStyle(focused).Render(label)
}

// RenderFieldRow renders a complete field row with selector, label, and value.
func (s Styles) RenderFieldRow(selected bool, label, value string) string {
	selector := s.RenderSelector(selected)
	labelStr := s.LabelStyle(selected).Render(label)
	return selector + labelStr + value
}
