// Package help provides a help overlay component for the TUI.
package help

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
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
		Container: styles.ModalBorder().Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Underline(true).
			MarginBottom(1),
		Section: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Key: lipgloss.NewStyle().
			Foreground(colors.Success).
			Bold(true),
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

	switch msg := msg.(type) {
	case messages.NavigationMsg:
		return m.handleNavigation(msg), nil
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
}

func (m Model) handleNavigation(msg messages.NavigationMsg) Model {
	switch msg.Direction {
	case messages.NavDown:
		m.scrollOffset++
	case messages.NavUp:
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case messages.NavHome:
		m.scrollOffset = 0
	case messages.NavPageDown:
		m.scrollOffset += 10
	case messages.NavPageUp:
		if m.scrollOffset >= 10 {
			m.scrollOffset -= 10
		} else {
			m.scrollOffset = 0
		}
	case messages.NavLeft, messages.NavRight, messages.NavEnd:
		// Not applicable for help overlay
	}
	return m
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("?", "escape", "q"))):
		m.visible = false
		return m, func() tea.Msg { return CloseMsg{} }
	case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
		m.scrollOffset++
	case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
		m.scrollOffset = 0
	}
	return m, nil
}

// View renders the full help overlay.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	// Get bindings for current context
	sections := m.getContextBindings()

	// Build content first to measure it
	var content string
	gap := 4
	if m.width >= 120 && len(sections) == 2 {
		content = m.renderTwoColumnLayout(sections, gap)
	} else {
		content = m.renderSingleColumnLayout(sections)
	}

	// Footer
	footer := m.styles.Footer.Render("Press ? or Esc to close")

	// Measure content dimensions
	contentLines := strings.Split(content, "\n")
	maxLineWidth := lipgloss.Width(footer)
	for _, line := range contentLines {
		if w := lipgloss.Width(line); w > maxLineWidth {
			maxLineWidth = w
		}
	}

	// Calculate overlay dimensions based on content
	// Add padding: 2 for border, 4 for internal padding (2 each side)
	overlayWidth := maxLineWidth + 6
	overlayHeight := len(contentLines) + 5 // +5 for border, title, footer, padding

	// Ensure minimum size and don't exceed screen
	overlayWidth = max(40, min(overlayWidth, m.width-4))
	overlayHeight = max(10, min(overlayHeight, m.height-4))

	colors := theme.GetSemanticColors()
	r := rendering.New(overlayWidth, overlayHeight).
		SetTitle("Help - " + keys.ContextName(m.context)).
		SetFocused(true).
		SetFocusColor(colors.Highlight)

	// Center footer within content width
	contentWidth := overlayWidth - 6
	footerCentered := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(footer)
	content += "\n" + footerCentered

	return r.SetContent(content).Render()
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

	return styles.SeparatorTop().
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

// keyDescPair holds a key string and its description for alignment.
type keyDescPair struct {
	key  string
	desc string
}

func (m Model) formatBindings(bindings []keys.KeyBinding) []string {
	// Group bindings by description (action) to collapse duplicates
	descToKeys := make(map[string][]string)
	descOrder := make([]string, 0)

	for _, b := range bindings {
		if b.Desc == "" {
			continue
		}
		if _, exists := descToKeys[b.Desc]; !exists {
			descOrder = append(descOrder, b.Desc)
		}
		descToKeys[b.Desc] = append(descToKeys[b.Desc], b.Key)
	}

	// Build pairs and find max key width
	pairs := make([]keyDescPair, 0, len(descOrder))
	maxKeyLen := 0
	for _, desc := range descOrder {
		keysList := descToKeys[desc]
		sort.Strings(keysList)
		keyStr := strings.Join(keysList, "/")
		if len(keysList) > 3 {
			keyStr = strings.Join(keysList[:3], "/") + "..."
		}
		if len(keyStr) > maxKeyLen {
			maxKeyLen = len(keyStr)
		}
		pairs = append(pairs, keyDescPair{key: keyStr, desc: desc})
	}

	// Format lines with aligned keys and descriptions
	lines := make([]string, 0, len(pairs))
	for _, p := range pairs {
		// Left-align key, pad to max width, then description
		paddedKey := p.key + strings.Repeat(" ", maxKeyLen-len(p.key))
		line := m.styles.Key.Render(paddedKey) + " " + m.styles.Desc.Render(p.desc)
		lines = append(lines, line)
	}
	return lines
}

func (m Model) renderSingleColumnLayout(sections []BindingSection) string {
	var content strings.Builder

	for i, section := range sections {
		if i > 0 {
			content.WriteString("\n")
		}
		// Section title (will be centered by View)
		content.WriteString(m.styles.Section.Render(section.Name) + "\n")

		// Build the keybindings block (left-aligned internally)
		lines := m.formatBindings(section.Bindings)

		// Add each line
		for _, line := range lines {
			content.WriteString(line + "\n")
		}
	}
	return content.String()
}

func (m Model) renderTwoColumnLayout(sections []BindingSection, gap int) string {
	// Format each section's bindings (internally left-aligned)
	leftLines := m.formatBindings(sections[0].Bindings)
	rightLines := m.formatBindings(sections[1].Bindings)

	// Find max width in left column for padding
	leftMaxW := 0
	for _, line := range leftLines {
		if w := lipgloss.Width(line); w > leftMaxW {
			leftMaxW = w
		}
	}

	var content strings.Builder

	// Section headers side by side
	leftHeaderText := m.styles.Section.Render(sections[0].Name)
	rightHeaderText := m.styles.Section.Render(sections[1].Name)
	leftHeaderW := lipgloss.Width(leftHeaderText)

	// Pad left header to match left column width
	leftHeaderPadded := leftHeaderText + strings.Repeat(" ", leftMaxW-leftHeaderW)
	content.WriteString(leftHeaderPadded + strings.Repeat(" ", gap) + rightHeaderText + "\n")

	// Render rows side-by-side
	maxRows := max(len(leftLines), len(rightLines))

	for i := range maxRows {
		// Left column - pad to max width
		var leftCell string
		if i < len(leftLines) {
			line := leftLines[i]
			lineWidth := lipgloss.Width(line)
			leftCell = line + strings.Repeat(" ", leftMaxW-lineWidth)
		} else {
			leftCell = strings.Repeat(" ", leftMaxW)
		}

		// Right column
		var rightCell string
		if i < len(rightLines) {
			rightCell = rightLines[i]
		}

		content.WriteString(leftCell + strings.Repeat(" ", gap) + rightCell + "\n")
	}

	return content.String()
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
			key.NewBinding(key.WithKeys(keyconst.KeyTab), key.WithHelp("tab", "next view")),
		},
		{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
			key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		},
	}
}
