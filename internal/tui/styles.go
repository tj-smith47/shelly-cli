package tui

import (
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// Styles contains all styles for the TUI components.
type Styles struct {
	// Layout
	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style

	// Status bar
	StatusBar     lipgloss.Style
	StatusNormal  lipgloss.Style
	StatusSuccess lipgloss.Style
	StatusError   lipgloss.Style
	StatusWarn    lipgloss.Style

	// Table
	TableHeader     lipgloss.Style
	TableRow        lipgloss.Style
	TableRowAlt     lipgloss.Style
	TableRowFocused lipgloss.Style

	// Tabs
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Device states
	DeviceOnline  lipgloss.Style
	DeviceOffline lipgloss.Style
	SwitchOn      lipgloss.Style
	SwitchOff     lipgloss.Style

	// Misc
	Title   lipgloss.Style
	Border  lipgloss.Style
	Spinner lipgloss.Style
}

// DefaultStyles returns theme-aware styles for the TUI.
// Uses semantic colors for consistent theming across the application.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		// Layout
		App:    lipgloss.NewStyle(),
		Header: lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Footer: lipgloss.NewStyle().Padding(0, 1),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.AltBackground),
		StatusNormal: lipgloss.NewStyle().
			Foreground(colors.Text),
		StatusSuccess: lipgloss.NewStyle().
			Foreground(colors.Success),
		StatusError: lipgloss.NewStyle().
			Foreground(colors.Error),
		StatusWarn: lipgloss.NewStyle().
			Foreground(colors.Warning),

		// Table
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.TableHeader).
			Padding(0, 1),
		TableRow: lipgloss.NewStyle().
			Foreground(colors.TableCell).
			Padding(0, 1),
		TableRowAlt: lipgloss.NewStyle().
			Foreground(colors.TableAltCell).
			Padding(0, 1),
		TableRowFocused: lipgloss.NewStyle().
			Foreground(colors.Text).
			Background(colors.AltBackground).
			Padding(0, 1),

		// Tabs
		TabActive: styles.TabUnderlineActive().
			Bold(true).
			Foreground(colors.Highlight).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Padding(0, 2),

		// Help
		HelpKey: lipgloss.NewStyle().
			Foreground(colors.Secondary),
		HelpDesc: lipgloss.NewStyle().
			Foreground(colors.Muted),

		// Device states
		DeviceOnline: lipgloss.NewStyle().
			Foreground(colors.Online),
		DeviceOffline: lipgloss.NewStyle().
			Foreground(colors.Offline),
		SwitchOn: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		SwitchOff: lipgloss.NewStyle().
			Foreground(colors.Muted),

		// Misc
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Primary),
		Border:  styles.PanelBorder(),
		Spinner: lipgloss.NewStyle().Foreground(colors.Info),
	}
}

// RefreshStyles updates styles to reflect the current theme.
// Call this after changing themes at runtime.
func (s *Styles) RefreshStyles() {
	*s = DefaultStyles()
}
