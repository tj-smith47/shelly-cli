// Package form provides form components for the TUI.
package form

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	tuistyles "github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// SelectStyles holds the visual styles for a select component.
type SelectStyles struct {
	Label        lipgloss.Style
	Option       lipgloss.Style
	SelectedMark lipgloss.Style
	Cursor       lipgloss.Style
	Help         lipgloss.Style
	Container    lipgloss.Style
	Focused      lipgloss.Style
	NoMatch      lipgloss.Style
}

// DefaultSelectStyles returns the default styles for a select.
func DefaultSelectStyles() SelectStyles {
	colors := theme.GetSemanticColors()
	return SelectStyles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Option: lipgloss.NewStyle().
			Foreground(colors.Text),
		SelectedMark: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		Cursor: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Container: tuistyles.PanelBorder().Padding(0, 1),
		Focused:   tuistyles.PanelBorderActive().Padding(0, 1),
		NoMatch: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// SelectOption configures a select component.
type SelectOption func(*Select)

// WithSelectLabel sets the select label.
func WithSelectLabel(label string) SelectOption {
	return func(s *Select) {
		s.label = label
	}
}

// WithSelectOptions sets the select options.
func WithSelectOptions(options []string) SelectOption {
	return func(s *Select) {
		s.options = options
		s.filtered = options
	}
}

// WithSelectHelp sets the help text.
func WithSelectHelp(help string) SelectOption {
	return func(s *Select) {
		s.help = help
	}
}

// WithSelectStyles sets custom styles.
func WithSelectStyles(styles SelectStyles) SelectOption {
	return func(s *Select) {
		s.styles = styles
	}
}

// WithSelectSelected sets the initially selected index.
func WithSelectSelected(index int) SelectOption {
	return func(s *Select) {
		s.selected = index
		s.cursor = index
	}
}

// WithSelectMaxVisible sets the maximum visible options.
func WithSelectMaxVisible(maxItems int) SelectOption {
	return func(s *Select) {
		s.maxVisible = maxItems
	}
}

// WithSelectFiltering enables type-to-filter search.
func WithSelectFiltering(enabled bool) SelectOption {
	return func(s *Select) {
		s.filterEnabled = enabled
	}
}

// Select is a single-select component with optional filtering.
type Select struct {
	label         string
	options       []string // All options
	filtered      []string // Options matching filter (or all if no filter)
	selected      int      // Index in options
	cursor        int      // Index in filtered
	help          string
	expanded      bool
	focused       bool
	filtering     bool // Currently in filter mode
	filterEnabled bool // Filtering feature enabled
	maxVisible    int
	offset        int
	styles        SelectStyles
	filterInput   textinput.Model
}

// NewSelect creates a new select component with options.
func NewSelect(opts ...SelectOption) Select {
	colors := theme.GetSemanticColors()

	inputStyles := textinput.Styles{}
	inputStyles.Focused.Text = inputStyles.Focused.Text.Foreground(colors.Highlight)
	inputStyles.Focused.Placeholder = inputStyles.Focused.Placeholder.Foreground(colors.Muted)
	inputStyles.Blurred.Text = inputStyles.Blurred.Text.Foreground(colors.Text)
	inputStyles.Blurred.Placeholder = inputStyles.Blurred.Placeholder.Foreground(colors.Muted)

	input := textinput.New()
	input.Placeholder = "Type to filter..."
	input.CharLimit = 50
	input.SetWidth(30)
	input.SetStyles(inputStyles)

	s := Select{
		maxVisible:  8,
		styles:      DefaultSelectStyles(),
		filterInput: input,
	}

	for _, opt := range opts {
		opt(&s)
	}

	// Initialize filtered to all options
	if s.filtered == nil {
		s.filtered = s.options
	}

	return s
}

// Init returns the initial command.
func (s Select) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s Select) Update(msg tea.Msg) (Select, tea.Cmd) {
	if !s.focused {
		return s, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return s.handleKey(keyMsg)
	}

	// Forward to filter input when filtering
	if s.filtering {
		var cmd tea.Cmd
		oldValue := s.filterInput.Value()
		s.filterInput, cmd = s.filterInput.Update(msg)
		if s.filterInput.Value() != oldValue {
			s = s.applyFilter()
		}
		return s, cmd
	}

	return s, nil
}

func (s Select) handleKey(msg tea.KeyPressMsg) (Select, tea.Cmd) {
	key := msg.String()

	// Handle filter mode keys first
	if s.filtering {
		return s.handleFilterKey(key, msg)
	}

	// Normal mode keys
	switch key {
	case "enter", "space":
		return s.handleConfirm(), nil
	case "/":
		return s.handleSlash()
	case "esc", "ctrl+[":
		return s.handleEscape(), nil
	case "j", "down":
		return s.handleDown(), nil
	case "k", "up":
		return s.handleUp(), nil
	case "g", "home":
		return s.handleHome(), nil
	case "shift+G", "G", "end":
		return s.handleEnd(), nil
	case "tab":
		return s.handleTab(), nil
	}

	return s, nil
}

func (s Select) handleConfirm() Select {
	if s.expanded {
		s = s.selectCurrent()
		s.expanded = false
	} else {
		s.expanded = true
		s.cursor = s.selectedCursorPosition()
		s = s.adjustOffset()
	}
	return s
}

func (s Select) handleSlash() (Select, tea.Cmd) {
	if s.filterEnabled && s.expanded {
		s.filtering = true
		s.filterInput.Focus()
		s.filterInput.SetValue("")
		return s, textinput.Blink
	}
	return s, nil
}

func (s Select) handleEscape() Select {
	if s.expanded {
		s.expanded = false
		s.cursor = s.selectedCursorPosition()
	}
	return s
}

func (s Select) handleDown() Select {
	if !s.expanded {
		s.expanded = true
		s.cursor = s.selectedCursorPosition()
		s = s.adjustOffset()
	} else {
		s = s.cursorDown()
	}
	return s
}

func (s Select) handleUp() Select {
	if s.expanded {
		s = s.cursorUp()
	}
	return s
}

func (s Select) handleHome() Select {
	if s.expanded {
		s.cursor = 0
		s.offset = 0
	}
	return s
}

func (s Select) handleEnd() Select {
	if s.expanded && len(s.filtered) > 0 {
		s.cursor = len(s.filtered) - 1
		s = s.adjustOffset()
	}
	return s
}

func (s Select) handleTab() Select {
	if s.expanded {
		s = s.selectCurrent()
		s.expanded = false
	}
	return s
}

func (s Select) handleFilterKey(key string, msg tea.KeyPressMsg) (Select, tea.Cmd) {
	switch key {
	case "enter":
		// Select current and exit filter mode
		s = s.selectCurrent()
		s.expanded = false
		s.filtering = false
		s.filterInput.Blur()
		s.filterInput.SetValue("")
		s = s.applyFilter()
		return s, nil

	case "esc", "ctrl+[":
		// Exit filter mode but stay expanded
		s.filtering = false
		s.filterInput.Blur()
		s.filterInput.SetValue("")
		s = s.applyFilter()
		return s, nil

	case "down", "ctrl+n":
		s = s.cursorDown()
		return s, nil

	case "up", "ctrl+p":
		s = s.cursorUp()
		return s, nil
	}

	// Forward to filter input
	var cmd tea.Cmd
	oldValue := s.filterInput.Value()
	s.filterInput, cmd = s.filterInput.Update(msg)
	if s.filterInput.Value() != oldValue {
		s = s.applyFilter()
	}
	return s, cmd
}

func (s Select) selectCurrent() Select {
	if len(s.filtered) > 0 && s.cursor >= 0 && s.cursor < len(s.filtered) {
		selectedValue := s.filtered[s.cursor]
		// Find index in original options
		for i, opt := range s.options {
			if opt == selectedValue {
				s.selected = i
				break
			}
		}
	}
	return s
}

func (s Select) selectedCursorPosition() int {
	if s.selected < 0 || s.selected >= len(s.options) {
		return 0
	}
	selectedValue := s.options[s.selected]
	for i, opt := range s.filtered {
		if opt == selectedValue {
			return i
		}
	}
	return 0
}

func (s Select) applyFilter() Select {
	query := strings.ToLower(s.filterInput.Value())
	if query == "" {
		s.filtered = s.options
	} else {
		s.filtered = make([]string, 0)
		for _, opt := range s.options {
			if strings.Contains(strings.ToLower(opt), query) {
				s.filtered = append(s.filtered, opt)
			}
		}
	}
	s.cursor = 0
	s.offset = 0
	return s
}

func (s Select) cursorDown() Select {
	if s.cursor < len(s.filtered)-1 {
		s.cursor++
		s = s.adjustOffset()
	}
	return s
}

func (s Select) cursorUp() Select {
	if s.cursor > 0 {
		s.cursor--
		s = s.adjustOffset()
	}
	return s
}

func (s Select) adjustOffset() Select {
	if s.cursor < s.offset {
		s.offset = s.cursor
	}
	if s.cursor >= s.offset+s.maxVisible {
		s.offset = s.cursor - s.maxVisible + 1
	}
	return s
}

// View renders the select component.
func (s Select) View() string {
	var result string

	// Label
	if s.label != "" {
		result += s.styles.Label.Render(s.label) + "\n"
	}

	if s.expanded {
		result += s.viewExpanded()
	} else {
		result += s.viewCollapsed()
	}

	// Help text (not when expanded and filtering)
	if s.help != "" && !s.filtering {
		result += "\n" + s.styles.Help.Render(s.help)
	}

	return result
}

func (s Select) viewCollapsed() string {
	display := "(none selected)"
	if s.selected >= 0 && s.selected < len(s.options) {
		display = s.options[s.selected]
	}
	content := display + " ▼"

	if s.focused {
		return s.styles.Focused.Render(content)
	}
	return s.styles.Container.Render(content)
}

func (s Select) viewExpanded() string {
	var content strings.Builder

	// Filter input (if filtering)
	if s.filtering {
		content.WriteString(s.filterInput.View())
		content.WriteString("\n")
	} else if s.filterEnabled {
		// Show hint that filtering is available
		content.WriteString(s.styles.Help.Render("Press / to filter"))
		content.WriteString("\n")
	}

	// Options
	if len(s.filtered) == 0 {
		content.WriteString(s.styles.NoMatch.Render("No matches"))
		return s.styles.Focused.Render(content.String())
	}

	s.renderOptionsTo(&content)
	return s.styles.Focused.Render(content.String())
}

func (s Select) renderOptionsTo(content *strings.Builder) {
	end := s.offset + s.maxVisible
	if end > len(s.filtered) {
		end = len(s.filtered)
	}

	// Scroll up indicator
	if s.offset > 0 {
		content.WriteString(s.styles.Help.Render("↑ more"))
		content.WriteString("\n")
	}

	for i := s.offset; i < end; i++ {
		content.WriteString(s.renderOption(i))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll down indicator
	if end < len(s.filtered) {
		content.WriteString("\n")
		content.WriteString(s.styles.Help.Render("↓ more"))
	}
}

func (s Select) renderOption(index int) string {
	opt := s.filtered[index]

	// Cursor indicator
	prefix := "  "
	if index == s.cursor {
		prefix = s.styles.Cursor.Render("▶ ")
	}

	// Check if this is the currently selected value
	isSelected := false
	if s.selected >= 0 && s.selected < len(s.options) {
		isSelected = s.options[s.selected] == opt
	}

	// Selection mark
	suffix := ""
	if isSelected {
		suffix = s.styles.SelectedMark.Render(" ✓")
	}

	optStyle := s.styles.Option
	if index == s.cursor {
		optStyle = s.styles.Cursor
	}

	return prefix + optStyle.Render(opt) + suffix
}

// Focus focuses the select component.
func (s Select) Focus() Select {
	s.focused = true
	return s
}

// Blur removes focus from the select component.
func (s Select) Blur() Select {
	s.focused = false
	s.expanded = false
	s.filtering = false
	s.filterInput.Blur()
	s.filterInput.SetValue("")
	s = s.applyFilter()
	return s
}

// Focused returns whether the select is focused.
func (s Select) Focused() bool {
	return s.focused
}

// Selected returns the selected index.
func (s Select) Selected() int {
	return s.selected
}

// SelectedValue returns the selected option value.
func (s Select) SelectedValue() string {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.options[s.selected]
	}
	return ""
}

// SetSelected sets the selected index.
func (s Select) SetSelected(index int) Select {
	if index >= 0 && index < len(s.options) {
		s.selected = index
		s.cursor = s.selectedCursorPosition()
	}
	return s
}

// SetSelectedValue sets the selected value by matching the string.
func (s Select) SetSelectedValue(value string) Select {
	for i, opt := range s.options {
		if opt == value {
			s.selected = i
			s.cursor = s.selectedCursorPosition()
			s = s.adjustOffset()
			break
		}
	}
	return s
}

// SetOptions updates the select options.
func (s Select) SetOptions(options []string) Select {
	s.options = options
	s.filtered = options
	if s.selected >= len(options) {
		s.selected = len(options) - 1
	}
	if s.selected < 0 {
		s.selected = 0
	}
	s.cursor = s.selectedCursorPosition()
	s.offset = 0
	return s
}

// IsExpanded returns whether the select is expanded.
func (s Select) IsExpanded() bool {
	return s.expanded
}

// Collapse collapses the select.
func (s Select) Collapse() Select {
	s.expanded = false
	s.filtering = false
	s.filterInput.Blur()
	s.filterInput.SetValue("")
	s = s.applyFilter()
	return s
}

// Expand expands the select.
func (s Select) Expand() Select {
	s.expanded = true
	s.cursor = s.selectedCursorPosition()
	s = s.adjustOffset()
	return s
}

// IsFiltering returns whether the select is in filter mode.
func (s Select) IsFiltering() bool {
	return s.filtering
}
