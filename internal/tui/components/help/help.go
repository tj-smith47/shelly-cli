// Package help provides a help overlay component for the TUI.
package help

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ViewContext defines context-specific help content.
type ViewContext int

const (
	// ContextDevices shows help for the devices view.
	ContextDevices ViewContext = iota
	// ContextMonitor shows help for the monitor view.
	ContextMonitor
	// ContextEvents shows help for the events view.
	ContextEvents
	// ContextEnergy shows help for the energy view.
	ContextEnergy
)

// Model holds the help overlay state.
type Model struct {
	visible bool
	context ViewContext
	width   int
	height  int
	styles  Styles
}

// Styles for the help component.
type Styles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Section   lipgloss.Style
	Key       lipgloss.Style
	Desc      lipgloss.Style
	Footer    lipgloss.Style
}

// DefaultStyles returns default styles for the help component.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Underline(true).
			MarginBottom(1),
		Section: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true).
			MarginTop(1),
		Key: lipgloss.NewStyle().
			Foreground(colors.Success).
			Bold(true).
			Width(12),
		Desc: lipgloss.NewStyle().
			Foreground(colors.Text),
		Footer: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true).
			MarginTop(1),
	}
}

// New creates a new help model.
func New() Model {
	return Model{
		styles: DefaultStyles(),
	}
}

// Init initializes the help component.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help component.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the help overlay.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	content := m.styles.Title.Render("Keyboard Shortcuts") + "\n\n"

	// Navigation section
	content += m.styles.Section.Render("Navigation") + "\n"
	content += m.renderBinding("j/k", "Move up/down")
	content += m.renderBinding("g/G", "Go to top/bottom")
	content += m.renderBinding("Tab", "Next view")
	content += m.renderBinding("1-4", "Jump to view")

	// Actions section
	content += m.styles.Section.Render("Actions") + "\n"
	content += m.renderBinding("r", "Refresh data")
	content += m.renderBinding("/", "Filter devices")
	content += m.renderBinding("Enter", "Select item")
	content += m.renderBinding("Esc", "Back/close")

	// Device control section (only in devices/monitor view)
	if m.context == ContextDevices || m.context == ContextMonitor {
		content += m.styles.Section.Render("Device Control") + "\n"
		content += m.renderBinding("t", "Toggle switch")
		content += m.renderBinding("o", "Turn on")
		content += m.renderBinding("O", "Turn off")
		content += m.renderBinding("R", "Reboot device")
	}

	// View-specific help
	content += m.contextHelp()

	// General section
	content += m.styles.Section.Render("General") + "\n"
	content += m.renderBinding("?", "Toggle help")
	content += m.renderBinding("q", "Quit")

	content += m.styles.Footer.Render("Press ? or Esc to close")

	return m.styles.Container.
		Width(m.width-4).
		Height(m.height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// renderBinding renders a single key binding line.
func (m Model) renderBinding(keyStr, desc string) string {
	return m.styles.Key.Render(keyStr) + m.styles.Desc.Render(desc) + "\n"
}

// contextHelp returns view-specific help content.
func (m Model) contextHelp() string {
	switch m.context {
	case ContextDevices:
		return m.styles.Section.Render("Devices View") + "\n" +
			m.renderBinding("/", "Filter by name/type")

	case ContextMonitor:
		return m.styles.Section.Render("Monitor View") + "\n" +
			m.renderBinding("Enter", "View device details")

	case ContextEvents:
		return m.styles.Section.Render("Events View") + "\n" +
			m.renderBinding("c", "Clear events") +
			m.renderBinding("p", "Pause/resume")

	case ContextEnergy:
		return m.styles.Section.Render("Energy View") + "\n" +
			m.renderBinding("d", "Toggle daily/weekly")

	default:
		return ""
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetContext sets the current view context for context-sensitive help.
func (m Model) SetContext(ctx ViewContext) Model {
	m.context = ctx
	return m
}

// Show shows the help overlay.
func (m Model) Show() Model {
	m.visible = true
	return m
}

// Hide hides the help overlay.
func (m Model) Hide() Model {
	m.visible = false
	return m
}

// Toggle toggles the help overlay visibility.
func (m Model) Toggle() Model {
	m.visible = !m.visible
	return m
}

// Visible returns whether the help overlay is visible.
func (m Model) Visible() bool {
	return m.visible
}

// ShortHelp returns keybindings for the short help view (implements key.Map).
func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

// FullHelp returns keybindings for the full help view (implements key.Map).
func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/down", "move down")),
			key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/up", "move up")),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next view")),
		},
		{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
			key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		},
	}
}
