// Package search provides a search/filter component for the TUI.
package search

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// FilterChangedMsg signals that the filter value has changed.
type FilterChangedMsg struct {
	Filter string
}

// ClosedMsg signals that the search input was closed.
type ClosedMsg struct{}

// Model holds the search component state.
type Model struct {
	textInput textinput.Model
	active    bool
	width     int
	styles    Styles
}

// Styles for the search component.
type Styles struct {
	Container  lipgloss.Style
	Label      lipgloss.Style
	Input      lipgloss.Style
	InputFocus lipgloss.Style
}

// DefaultStyles returns default styles for the search component.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder),
		Label: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Input: lipgloss.NewStyle().
			Foreground(colors.Text),
		InputFocus: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
	}
}

// New creates a new search model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "type to filter devices..."
	ti.CharLimit = 50
	ti.SetWidth(30)

	// Configure styles using semantic colors
	colors := theme.GetSemanticColors()
	styles := textinput.DefaultStyles(true) // dark mode
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(colors.Highlight)
	styles.Focused.Text = styles.Focused.Text.Foreground(colors.Text)
	styles.Focused.Placeholder = styles.Focused.Placeholder.Foreground(colors.Muted)
	styles.Blurred.Prompt = styles.Blurred.Prompt.Foreground(colors.Highlight)
	styles.Blurred.Text = styles.Blurred.Text.Foreground(colors.Text)
	styles.Blurred.Placeholder = styles.Blurred.Placeholder.Foreground(colors.Muted)
	ti.SetStyles(styles)

	return Model{
		textInput: ti,
		styles:    DefaultStyles(),
	}
}

// NewWithFilter creates a new search model with an initial filter value.
func NewWithFilter(filter string) Model {
	m := New()
	if filter != "" {
		m.textInput.SetValue(filter)
	}
	return m
}

// Init initializes the search component.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the search component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("esc", "ctrl+["))):
			m.active = false
			m.textInput.Blur()
			return m, func() tea.Msg { return ClosedMsg{} }

		case key.Matches(keyMsg, key.NewBinding(key.WithKeys("enter"))):
			m.active = false
			m.textInput.Blur()
			return m, func() tea.Msg { return ClosedMsg{} }
		}
	}

	var cmd tea.Cmd
	prevValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)

	// Notify of filter change
	if m.textInput.Value() != prevValue {
		filter := m.textInput.Value()
		return m, tea.Batch(cmd, func() tea.Msg {
			return FilterChangedMsg{Filter: filter}
		})
	}

	return m, cmd
}

// View renders the search component.
func (m Model) View() string {
	if !m.active {
		return ""
	}

	label := m.styles.Label.Render("Filter: ")
	input := m.textInput.View()

	return m.styles.Container.
		Width(m.width).
		Render(label + input)
}

// SetWidth sets the component width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	m.textInput.SetWidth(width - 20) // Account for label and padding
	return m
}

// Activate shows the search input and focuses it.
func (m Model) Activate() (Model, tea.Cmd) {
	m.active = true
	m.textInput.Focus()
	return m, textinput.Blink
}

// Deactivate hides the search input.
func (m Model) Deactivate() Model {
	m.active = false
	m.textInput.Blur()
	return m
}

// Clear clears the search input value.
func (m Model) Clear() Model {
	m.textInput.SetValue("")
	return m
}

// IsActive returns whether the search input is active.
func (m Model) IsActive() bool {
	return m.active
}

// Value returns the current filter value.
func (m Model) Value() string {
	return m.textInput.Value()
}

// MatchesFilter checks if a string matches the current filter.
// Returns true if filter is empty or string contains filter (case-insensitive).
func (m Model) MatchesFilter(s string) bool {
	filter := m.textInput.Value()
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(filter))
}
