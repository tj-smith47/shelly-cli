// Package toast provides toast notification component for the TUI.
package toast

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Level indicates the toast notification level.
type Level int

const (
	// LevelInfo is an informational toast.
	LevelInfo Level = iota
	// LevelSuccess is a success toast.
	LevelSuccess
	// LevelWarning is a warning toast.
	LevelWarning
	// LevelError is an error toast.
	LevelError
)

// DefaultDuration is the default toast display duration.
const DefaultDuration = 3 * time.Second

// Toast represents a single toast notification.
type Toast struct {
	ID       int
	Message  string
	Level    Level
	Duration time.Duration
	Created  time.Time
}

// dismissMsg is sent when a toast should be dismissed.
type dismissMsg struct {
	ID int
}

// Model holds the toast notification state.
type Model struct {
	toasts  []Toast
	nextID  int
	width   int
	height  int
	visible bool
	styles  Styles
}

// Styles for toast notifications.
type Styles struct {
	Container lipgloss.Style
	Info      lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
}

// DefaultStyles returns default styles for toasts.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	baseStyle := lipgloss.NewStyle().
		Padding(0, 2).
		MarginBottom(1).
		Bold(true)

	return Styles{
		Container: lipgloss.NewStyle().
			Align(lipgloss.Right).
			Padding(1),
		Info: baseStyle.
			Foreground(colors.Text).
			Background(colors.Info),
		Success: baseStyle.
			Foreground(colors.Background).
			Background(colors.Success),
		Warning: baseStyle.
			Foreground(colors.Background).
			Background(colors.Warning),
		Error: baseStyle.
			Foreground(colors.Text).
			Background(colors.Error),
	}
}

// New creates a new toast notification model.
func New() Model {
	return Model{
		toasts:  make([]Toast, 0),
		visible: true,
		styles:  DefaultStyles(),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the toast component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ShowMsg:
		toast := Toast{
			ID:       m.nextID,
			Message:  msg.Message,
			Level:    msg.Level,
			Duration: msg.Duration,
			Created:  time.Now(),
		}
		if toast.Duration == 0 {
			toast.Duration = DefaultDuration
		}
		m.nextID++
		m.toasts = append(m.toasts, toast)

		// Schedule dismissal
		return m, scheduleDismiss(toast.ID, toast.Duration)

	case dismissMsg:
		// Remove toast with matching ID
		for i, t := range m.toasts {
			if t.ID == msg.ID {
				m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
				break
			}
		}
		return m, nil

	case ClearAllMsg:
		m.toasts = make([]Toast, 0)
		return m, nil
	}

	return m, nil
}

// scheduleDismiss returns a command that dismisses a toast after duration.
func scheduleDismiss(id int, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(_ time.Time) tea.Msg {
		return dismissMsg{ID: id}
	})
}

// ShowMsg shows a new toast notification.
type ShowMsg struct {
	Message  string
	Level    Level
	Duration time.Duration
}

// ClearAllMsg clears all toasts.
type ClearAllMsg struct{}

// Show returns a command to show a toast.
func Show(message string, level Level) tea.Cmd {
	return ShowWithDuration(message, level, DefaultDuration)
}

// ShowWithDuration returns a command to show a toast with custom duration.
func ShowWithDuration(message string, level Level, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return ShowMsg{
			Message:  message,
			Level:    level,
			Duration: duration,
		}
	}
}

// Info shows an info toast.
func Info(message string) tea.Cmd {
	return Show(message, LevelInfo)
}

// Success shows a success toast.
func Success(message string) tea.Cmd {
	return Show(message, LevelSuccess)
}

// Warning shows a warning toast.
func Warning(message string) tea.Cmd {
	return Show(message, LevelWarning)
}

// Error shows an error toast.
func Error(message string) tea.Cmd {
	return Show(message, LevelError)
}

// ClearAll returns a command to clear all toasts.
func ClearAll() tea.Cmd {
	return func() tea.Msg {
		return ClearAllMsg{}
	}
}

// SetSize sets the toast container size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetVisible sets whether toasts are visible.
func (m Model) SetVisible(visible bool) Model {
	m.visible = visible
	return m
}

// HasToasts returns true if there are any toasts to display.
func (m Model) HasToasts() bool {
	return len(m.toasts) > 0
}

// View renders the toast notifications.
func (m Model) View() string {
	if !m.visible || len(m.toasts) == 0 {
		return ""
	}

	// Render toasts (max 5 visible)
	maxVisible := 5
	startIdx := 0
	if len(m.toasts) > maxVisible {
		startIdx = len(m.toasts) - maxVisible
	}

	var rendered string
	for _, toast := range m.toasts[startIdx:] {
		var style lipgloss.Style
		var icon string
		switch toast.Level {
		case LevelSuccess:
			style = m.styles.Success
			icon = "✓ "
		case LevelWarning:
			style = m.styles.Warning
			icon = "! "
		case LevelError:
			style = m.styles.Error
			icon = "✗ "
		default:
			style = m.styles.Info
			icon = "i "
		}
		rendered += style.Render(icon+toast.Message) + "\n"
	}

	return m.styles.Container.Width(m.width).Render(rendered)
}

// Overlay renders the toasts as an overlay positioned at top-right.
// Overlays toast content on top of base without affecting layout.
func (m Model) Overlay(base string) string {
	if !m.visible || len(m.toasts) == 0 {
		return base
	}

	// Simply return the base - toasts are rendered in the status bar instead
	// This prevents layout disruption from overlay attempts
	return base
}
