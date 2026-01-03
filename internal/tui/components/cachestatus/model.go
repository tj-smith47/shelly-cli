// Package cachestatus provides a cache status indicator component for TUI panels.
package cachestatus

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Styles for the cache status component.
type Styles struct {
	Muted   lipgloss.Style
	Spinner lipgloss.Style
}

// DefaultStyles returns default styles for the cache status.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Spinner: lipgloss.NewStyle().
			Foreground(colors.Highlight),
	}
}

// Model holds the cache status component state.
type Model struct {
	updatedAt  time.Time
	refreshing bool
	spinner    spinner.Model
	styles     Styles
}

// Option configures the cache status model.
type Option func(*Model)

// WithStyles sets custom styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.styles = styles
		m.spinner.Style = styles.Spinner
	}
}

// New creates a new cache status model.
func New(opts ...Option) Model {
	styles := DefaultStyles()
	s := spinner.New(
		spinner.WithSpinner(spinner.Spinner{
			Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
			FPS:    time.Second / 10,
		}),
		spinner.WithStyle(styles.Spinner),
	)

	m := Model{
		spinner: s,
		styles:  styles,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Init returns the initial command for the spinner.
func (m Model) Init() tea.Cmd {
	if m.refreshing {
		return m.spinner.Tick
	}
	return nil
}

// Update handles messages for the cache status component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.refreshing {
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// SetUpdatedAt sets the cache update timestamp.
func (m Model) SetUpdatedAt(t time.Time) Model {
	m.updatedAt = t
	return m
}

// SetRefreshing sets whether a refresh is in progress.
func (m Model) SetRefreshing(refreshing bool) Model {
	m.refreshing = refreshing
	return m
}

// IsRefreshing returns whether a refresh is in progress.
func (m Model) IsRefreshing() bool {
	return m.refreshing
}

// UpdatedAt returns the last update timestamp.
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// StartRefresh sets refreshing to true and returns the tick command.
func (m Model) StartRefresh() (Model, tea.Cmd) {
	m.refreshing = true
	return m, m.spinner.Tick
}

// StopRefresh sets refreshing to false and updates the timestamp.
func (m Model) StopRefresh() Model {
	m.refreshing = false
	m.updatedAt = time.Now()
	return m
}

// View renders the cache status.
// Returns "Updated X ago" or a refreshing spinner.
func (m Model) View() string {
	if m.refreshing {
		return m.spinner.View() + m.styles.Muted.Render(" Refreshing...")
	}

	if m.updatedAt.IsZero() {
		return ""
	}

	return m.styles.Muted.Render("Updated " + formatAge(m.updatedAt))
}

// ViewCompact renders a compact version for tight spaces.
// Returns "Xm ago" or a refreshing spinner without text.
func (m Model) ViewCompact() string {
	if m.refreshing {
		return m.spinner.View()
	}

	if m.updatedAt.IsZero() {
		return ""
	}

	return m.styles.Muted.Render(formatAgeCompact(m.updatedAt))
}

// Tick returns the spinner tick command.
func (m Model) Tick() tea.Cmd {
	return m.spinner.Tick
}

// formatAge returns a human-readable age string.
func formatAge(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	age := time.Since(t)
	switch {
	case age < time.Minute:
		return "just now"
	case age < time.Hour:
		mins := int(age.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case age < 24*time.Hour:
		hours := int(age.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(age.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// formatAgeCompact returns a compact age string.
func formatAgeCompact(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	age := time.Since(t)
	switch {
	case age < time.Minute:
		return "now"
	case age < time.Hour:
		return fmt.Sprintf("%dm", int(age.Minutes()))
	case age < 24*time.Hour:
		return fmt.Sprintf("%dh", int(age.Hours()))
	default:
		return fmt.Sprintf("%dd", int(age.Hours()/24))
	}
}
