package tui

import (
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
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
func DefaultStyles() Styles {
	return Styles{
		// Layout
		App:    lipgloss.NewStyle(),
		Header: lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Footer: lipgloss.NewStyle().Padding(0, 1),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			Padding(0, 1).
			Background(theme.BrightBlack()),
		StatusNormal: lipgloss.NewStyle().
			Foreground(theme.Fg()),
		StatusSuccess: lipgloss.NewStyle().
			Foreground(theme.Green()),
		StatusError: lipgloss.NewStyle().
			Foreground(theme.Red()),
		StatusWarn: lipgloss.NewStyle().
			Foreground(theme.Yellow()),

		// Table
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Cyan()).
			Padding(0, 1),
		TableRow: lipgloss.NewStyle().
			Foreground(theme.Fg()).
			Padding(0, 1),
		TableRowAlt: lipgloss.NewStyle().
			Foreground(theme.BrightBlack()).
			Padding(0, 1),
		TableRowFocused: lipgloss.NewStyle().
			Foreground(theme.Fg()).
			Background(theme.BrightBlack()).
			Padding(0, 1),

		// Tabs
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Cyan()).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(theme.Cyan()).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(theme.BrightBlack()).
			Padding(0, 2),

		// Help
		HelpKey: lipgloss.NewStyle().
			Foreground(theme.Cyan()),
		HelpDesc: lipgloss.NewStyle().
			Foreground(theme.BrightBlack()),

		// Device states
		DeviceOnline: lipgloss.NewStyle().
			Foreground(theme.Green()),
		DeviceOffline: lipgloss.NewStyle().
			Foreground(theme.Red()),
		SwitchOn: lipgloss.NewStyle().
			Foreground(theme.Green()).
			Bold(true),
		SwitchOff: lipgloss.NewStyle().
			Foreground(theme.BrightBlack()),

		// Misc
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Magenta()),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BrightBlack()),
		Spinner: lipgloss.NewStyle().
			Foreground(theme.Cyan()),
	}
}

// RefreshStyles updates styles to reflect the current theme.
// Call this after changing themes at runtime.
func (s *Styles) RefreshStyles() {
	*s = DefaultStyles()
}
