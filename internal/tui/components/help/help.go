// Package help provides a help overlay component for the TUI.
package help

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// CloseMsg is sent when help is closed.
type CloseMsg struct{}

// BindingSection represents a group of keybindings.
type BindingSection struct {
	Name     string
	Bindings []keys.KeyBinding
}

// Model holds the help overlay state.
type Model struct {
	visible      bool
	context      keys.Context
	keyMap       *keys.ContextMap
	width        int
	height       int
	scrollOffset int
	styles       Styles
}

// Styles for the help component.
type Styles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Section   lipgloss.Style
	Key       lipgloss.Style
	Desc      lipgloss.Style
	Footer    lipgloss.Style
	Muted     lipgloss.Style
}

// DefaultStyles returns default styles for the help component.
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
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// New creates a new help model.
func New() Model {
	return Model{
		keyMap: keys.NewContextMap(),
		styles: DefaultStyles(),
	}
}

// Init initializes the help component.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, key.NewBinding(key.WithKeys("?", "escape", "q"))):
		m.visible = false
		return m, func() tea.Msg { return CloseMsg{} }
	case key.Matches(keyMsg, key.NewBinding(key.WithKeys("j", "down"))):
		m.scrollOffset++
	case key.Matches(keyMsg, key.NewBinding(key.WithKeys("k", "up"))):
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case key.Matches(keyMsg, key.NewBinding(key.WithKeys("g"))):
		m.scrollOffset = 0
	}

	return m, nil
}

// View renders the full help overlay.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	// Calculate overlay dimensions (2/3 of screen)
	overlayWidth := m.width * 2 / 3
	overlayHeight := m.height * 2 / 3
	if overlayWidth < 60 {
		overlayWidth = 60
	}
	if overlayHeight < 20 {
		overlayHeight = 20
	}

	colors := theme.GetSemanticColors()
	r := rendering.New(overlayWidth, overlayHeight).
		SetTitle("Help - " + keys.ContextName(m.context)).
		SetFocused(true).
		SetFocusColor(colors.Highlight)

	var content strings.Builder

	// Get bindings for current context
	sections := m.getContextBindings()

	for i, section := range sections {
		if i > 0 {
			content.WriteString("\n")
		}
		content.WriteString(m.styles.Section.Render(section.Name) + "\n")
		content.WriteString(m.formatBindings(section.Bindings))
	}

	// Footer
	content.WriteString("\n")
	content.WriteString(m.styles.Footer.Render("Press ? or Esc to close"))

	return r.SetContent(content.String()).Render()
}

// ViewCompact renders a compact help bar at the bottom of the screen.
func (m Model) ViewCompact() string {
	colors := theme.GetSemanticColors()
	keyStyle := lipgloss.NewStyle().Foreground(colors.Success).Bold(true)
	sepStyle := lipgloss.NewStyle().Foreground(colors.Muted)

	bindings := []string{
		keyStyle.Render("j/k") + " nav",
		keyStyle.Render("h/l") + " cmp",
		keyStyle.Render("Tab") + " panel",
		keyStyle.Render("t") + " toggle",
		keyStyle.Render("o/O") + " on/off",
		keyStyle.Render("/") + " filter",
		keyStyle.Render(":") + " cmd",
		keyStyle.Render("?") + " help",
		keyStyle.Render("q") + " quit",
	}

	sep := sepStyle.Render(" │ ")
	content := ""
	for i, b := range bindings {
		if i > 0 {
			content += sep
		}
		content += b
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.TableBorder).
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		Padding(0, 1).
		Width(m.width - 2).
		Render(content)
}

func (m Model) getContextBindings() []BindingSection {
	var sections []BindingSection

	// Context-specific bindings
	contextBindings := m.keyMap.GetBindings(m.context)
	if len(contextBindings) > 0 {
		sections = append(sections, BindingSection{
			Name:     keys.ContextName(m.context),
			Bindings: m.sortBindings(contextBindings),
		})
	}

	// Global bindings
	globalBindings := m.keyMap.GetGlobalBindings()
	if len(globalBindings) > 0 {
		sections = append(sections, BindingSection{
			Name:     "Global",
			Bindings: m.sortBindings(globalBindings),
		})
	}

	return sections
}

func (m Model) sortBindings(bindings []keys.KeyBinding) []keys.KeyBinding {
	sorted := make([]keys.KeyBinding, len(bindings))
	copy(sorted, bindings)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})
	return sorted
}

func (m Model) formatBindings(bindings []keys.KeyBinding) string {
	lines := make([]string, 0, len(bindings))
	for _, b := range bindings {
		if b.Desc == "" {
			continue
		}
		line := m.styles.Key.Render(b.Key) + " " + m.styles.Desc.Render(b.Desc)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// ViewHeight returns the height the compact help bar will take.
func (m Model) ViewHeight() int {
	return 2
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetContext sets the current view context for context-sensitive help.
func (m Model) SetContext(ctx keys.Context) Model {
	m.context = ctx
	return m
}

// SetKeyMap sets the key map for context-sensitive bindings.
func (m Model) SetKeyMap(km *keys.ContextMap) Model {
	m.keyMap = km
	return m
}

// Show shows the help overlay.
func (m Model) Show() Model {
	m.visible = true
	m.scrollOffset = 0
	return m
}

// Hide hides the help overlay.
func (m Model) Hide() Model {
	m.visible = false
	return m
}

// Toggle toggles the help overlay visibility.
func (m Model) Toggle() Model {
	if m.visible {
		m.visible = false
	} else {
		m.visible = true
		m.scrollOffset = 0
	}
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
