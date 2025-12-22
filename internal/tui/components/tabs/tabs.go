// Package tabs provides the tab bar component for the TUI.
package tabs

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Model holds the tabs state.
type Model struct {
	tabs   []string
	active int
	styles Styles
}

// Styles for the tabs component.
type Styles struct {
	Container lipgloss.Style
	Active    lipgloss.Style
	Inactive  lipgloss.Style
	Separator lipgloss.Style
}

// DefaultStyles returns default styles for tabs.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.TableBorder).
			BorderBottom(true).
			Padding(0, 1),
		Active: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight).
			Background(colors.AltBackground).
			Padding(0, 2),
		Inactive: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Padding(0, 2),
		Separator: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// New creates a new tabs model.
func New(tabs []string, active int) Model {
	if active < 0 || active >= len(tabs) {
		active = 0
	}
	return Model{
		tabs:   tabs,
		active: active,
		styles: DefaultStyles(),
	}
}

// Init returns the initial command for tabs.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for tabs.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// SetActive sets the active tab.
func (m Model) SetActive(idx int) Model {
	if idx >= 0 && idx < len(m.tabs) {
		m.active = idx
	}
	return m
}

// Active returns the active tab index.
func (m Model) Active() int {
	return m.active
}

// ActiveName returns the name of the active tab.
func (m Model) ActiveName() string {
	if m.active >= 0 && m.active < len(m.tabs) {
		return m.tabs[m.active]
	}
	return ""
}

// Next moves to the next tab.
func (m Model) Next() Model {
	m.active = (m.active + 1) % len(m.tabs)
	return m
}

// Prev moves to the previous tab.
func (m Model) Prev() Model {
	m.active = (m.active + len(m.tabs) - 1) % len(m.tabs)
	return m
}

// View renders the tabs.
func (m Model) View() string {
	parts := make([]string, 0, len(m.tabs))

	for i, tab := range m.tabs {
		// Add tab number hint
		hint := string(rune('1' + i))

		var tabView string
		if i == m.active {
			tabView = m.styles.Active.Render(hint + " " + tab)
		} else {
			tabView = m.styles.Inactive.Render(hint + " " + tab)
		}
		parts = append(parts, tabView)
	}

	// Join tabs with separator
	sep := m.styles.Separator.Render(" | ")
	content := strings.Join(parts, sep)

	return m.styles.Container.Render(content)
}

// Count returns the number of tabs.
func (m Model) Count() int {
	return len(m.tabs)
}
