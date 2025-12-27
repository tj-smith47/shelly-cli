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

// StatusItem represents a piece of information in the status bar.
type StatusItem struct {
	Full    string // Full text for wide terminals (>120)
	Compact string // Compact text for medium terminals (>80)
	Minimal string // Minimal text for narrow terminals (<80)
}

// Model holds the status bar state.
type Model struct {
	width       int
	message     string
	messageType MessageType
	lastUpdate  time.Time
	styles      Styles
	items       []StatusItem
	debugActive bool
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
	Debug   lipgloss.Style
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
		Debug: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
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

// SetItems sets the context-specific status items.
func (m Model) SetItems(items []StatusItem) Model {
	m.items = items
	return m
}

// AddItem adds a status item.
func (m Model) AddItem(full, compact, minimal string) Model {
	m.items = append(m.items, StatusItem{
		Full:    full,
		Compact: compact,
		Minimal: minimal,
	})
	return m
}

// ClearItems clears all status items.
func (m Model) ClearItems() Model {
	m.items = nil
	return m
}

// SetDebugActive sets whether a debug session is active.
func (m Model) SetDebugActive(active bool) Model {
	m.debugActive = active
	return m
}

// IsDebugActive returns whether a debug session is active.
func (m Model) IsDebugActive() bool {
	return m.debugActive
}

// Tier represents the status bar display tier.
type Tier int

const (
	// TierMinimal is for narrow terminals (<80 columns).
	TierMinimal Tier = iota
	// TierCompact is for medium terminals (80-120 columns).
	TierCompact
	// TierFull is for wide terminals (>120 columns).
	TierFull
)

// GetTier returns the appropriate tier based on width.
func (m Model) GetTier() Tier {
	if m.width >= 120 {
		return TierFull
	}
	if m.width >= 80 {
		return TierCompact
	}
	return TierMinimal
}

// View renders the status bar.
func (m Model) View() string {
	tier := m.GetTier()

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

	// Debug indicator (recording dot + text)
	if m.debugActive {
		debugText := "REC"
		if tier == TierFull {
			debugText = "Debug active"
		}
		left += "  " + m.styles.Debug.Render("● "+debugText)
	}

	// Middle: context-specific items (if any)
	middle := m.renderItems(tier)
	if middle != "" {
		left += "  " + middle
	}

	// Right side: version and time (tier-dependent)
	timeStr := m.lastUpdate.Format("15:04:05")
	var right string
	switch tier {
	case TierFull:
		right = fmt.Sprintf("%s  %s",
			m.styles.Version.Render("v"+version.Version),
			m.styles.Time.Render(timeStr),
		)
	case TierCompact:
		right = fmt.Sprintf("%s  %s",
			m.styles.Version.Render(version.Version),
			m.styles.Time.Render(timeStr[:5]), // HH:MM only
		)
	default: // TierMinimal
		right = m.styles.Time.Render(timeStr[:5])
	}

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

// renderItems renders the status items based on tier.
func (m Model) renderItems(tier Tier) string {
	if len(m.items) == 0 {
		return ""
	}

	separator := " │ "
	if tier == TierMinimal {
		separator = " "
	}

	var parts []string
	for _, item := range m.items {
		var text string
		switch tier {
		case TierFull:
			text = item.Full
		case TierCompact:
			text = item.Compact
		default:
			text = item.Minimal
		}
		if text != "" {
			parts = append(parts, m.styles.Normal.Render(text))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return m.styles.Normal.Render(fmt.Sprintf("│ %s", joinStrings(parts, separator)))
}

// joinStrings joins strings with a separator.
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
