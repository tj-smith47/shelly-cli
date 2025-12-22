// Package statusbar provides the status bar component for the TUI.
package statusbar

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// tickMsg is sent on each tick for time updates.
type tickMsg time.Time

// Model holds the status bar state.
type Model struct {
	width       int
	message     string
	messageType MessageType
	lastUpdate  time.Time
	styles      Styles
}

// MessageType indicates the type of status message.
type MessageType int

const (
	// MessageNormal is a normal informational message.
	MessageNormal MessageType = iota
	// MessageSuccess indicates a successful operation.
	MessageSuccess
	// MessageError indicates an error.
	MessageError
	// MessageWarning indicates a warning.
	MessageWarning
)

// Styles for the status bar.
type Styles struct {
	Bar     lipgloss.Style
	Left    lipgloss.Style
	Right   lipgloss.Style
	Normal  lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Version lipgloss.Style
	Time    lipgloss.Style
}

// DefaultStyles returns default styles for the status bar.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Bar: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Padding(0, 1),
		Left: lipgloss.NewStyle().
			Align(lipgloss.Left),
		Right: lipgloss.NewStyle().
			Align(lipgloss.Right),
		Normal: lipgloss.NewStyle().
			Foreground(colors.Text),
		Success: lipgloss.NewStyle().
			Foreground(colors.Success),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Version: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Time: lipgloss.NewStyle().
			Foreground(colors.Info),
	}
}

// New creates a new status bar model.
func New() Model {
	return Model{
		message:    "Ready",
		lastUpdate: time.Now(),
		styles:     DefaultStyles(),
	}
}

// Init returns the initial command for the status bar.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd returns a command that ticks every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages for the status bar.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.lastUpdate = time.Time(msg)
		return m, tickCmd()

	case StatusMsg:
		m.message = msg.Message
		m.messageType = msg.Type
		return m, nil
	}

	return m, nil
}

// StatusMsg is a message to update the status bar.
type StatusMsg struct {
	Message string
	Type    MessageType
}

// SetMessage sets the status bar message.
func SetMessage(msg string, t MessageType) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Message: msg, Type: t}
	}
}

// SetWidth sets the status bar width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	return m
}

// View renders the status bar.
func (m Model) View() string {
	// Left side: status message
	var msgStyle lipgloss.Style
	switch m.messageType {
	case MessageSuccess:
		msgStyle = m.styles.Success
	case MessageError:
		msgStyle = m.styles.Error
	case MessageWarning:
		msgStyle = m.styles.Warning
	default:
		msgStyle = m.styles.Normal
	}
	left := msgStyle.Render(m.message)

	// Right side: version and time
	timeStr := m.lastUpdate.Format("15:04:05")
	right := fmt.Sprintf("%s  %s",
		m.styles.Version.Render("v"+version.Version),
		m.styles.Time.Render(timeStr),
	)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := m.width - leftWidth - rightWidth - 4 // Account for padding
	if spacing < 1 {
		spacing = 1
	}

	// Build the bar
	content := left + lipgloss.NewStyle().Width(spacing).Render("") + right

	return m.styles.Bar.Width(m.width).Render(content)
}
