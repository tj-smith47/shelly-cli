// Package form provides form components for the TUI.
package form

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// SearchableDropdownStyles holds the visual styles for a searchable dropdown.
type SearchableDropdownStyles struct {
	Label        lipgloss.Style
	Option       lipgloss.Style
	SelectedMark lipgloss.Style
	Cursor       lipgloss.Style
	Help         lipgloss.Style
	Container    lipgloss.Style
	Focused      lipgloss.Style
	NoMatch      lipgloss.Style
}

// DefaultSearchableDropdownStyles returns the default styles.
func DefaultSearchableDropdownStyles() SearchableDropdownStyles {
	colors := theme.GetSemanticColors()
	return SearchableDropdownStyles{
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
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(0, 1),
		Focused: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(0, 1),
		NoMatch: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// SearchableDropdown is a dropdown with type-to-search filtering.
type SearchableDropdown struct {
	label       string
	options     []string // All options
	filtered    []string // Options matching the filter
	selected    int      // Index in options (not filtered)
	cursor      int      // Index in filtered
	help        string
	expanded    bool
	focused     bool
	maxVisible  int
	offset      int
	styles      SearchableDropdownStyles
	searchInput textinput.Model
}

// NewSearchableDropdown creates a new searchable dropdown.
func NewSearchableDropdown() SearchableDropdown {
	colors := theme.GetSemanticColors()

	inputStyles := textinput.Styles{}
	inputStyles.Focused.Text = inputStyles.Focused.Text.Foreground(colors.Highlight)
	inputStyles.Focused.Placeholder = inputStyles.Focused.Placeholder.Foreground(colors.Muted)
	inputStyles.Blurred.Text = inputStyles.Blurred.Text.Foreground(colors.Text)
	inputStyles.Blurred.Placeholder = inputStyles.Blurred.Placeholder.Foreground(colors.Muted)

	input := textinput.New()
	input.Placeholder = "Type to search..."
	input.CharLimit = 50
	input.SetWidth(30)
	input.SetStyles(inputStyles)

	return SearchableDropdown{
		maxVisible:  8,
		styles:      DefaultSearchableDropdownStyles(),
		searchInput: input,
	}
}

// SetLabel sets the dropdown label.
func (d SearchableDropdown) SetLabel(label string) SearchableDropdown {
	d.label = label
	return d
}

// SetOptions sets the dropdown options.
func (d SearchableDropdown) SetOptions(options []string) SearchableDropdown {
	d.options = options
	d.filtered = options
	d.cursor = 0
	d.offset = 0
	return d
}

// SetHelp sets the help text.
func (d SearchableDropdown) SetHelp(help string) SearchableDropdown {
	d.help = help
	return d
}

// SetMaxVisible sets the maximum visible options.
func (d SearchableDropdown) SetMaxVisible(maxItems int) SearchableDropdown {
	d.maxVisible = maxItems
	return d
}

// SetSelected sets the selected value by matching the string.
func (d SearchableDropdown) SetSelected(value string) SearchableDropdown {
	for i, opt := range d.options {
		if opt == value {
			d.selected = i
			// Also position cursor if matching option is in filtered list
			for j, filteredOpt := range d.filtered {
				if filteredOpt == value {
					d.cursor = j
					d = d.adjustOffset()
					break
				}
			}
			break
		}
	}
	return d
}

// Init returns the initial command.
func (d SearchableDropdown) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (d SearchableDropdown) Update(msg tea.Msg) (SearchableDropdown, tea.Cmd) {
	if !d.focused {
		return d, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return d.handleKey(keyMsg)
	}

	// Forward to search input when expanded
	if d.expanded {
		var cmd tea.Cmd
		oldValue := d.searchInput.Value()
		d.searchInput, cmd = d.searchInput.Update(msg)
		// Re-filter if search changed
		if d.searchInput.Value() != oldValue {
			d = d.applyFilter()
		}
		return d, cmd
	}

	return d, nil
}

func (d SearchableDropdown) handleKey(msg tea.KeyPressMsg) (SearchableDropdown, tea.Cmd) {
	key := msg.String()

	switch key {
	case "enter":
		if d.expanded {
			return d.selectCurrentAndClose()
		}
		d.expanded = true
		d.searchInput.Focus()
		d.searchInput.SetValue("")
		d = d.applyFilter()
		return d, textinput.Blink
	case "esc", "ctrl+[":
		if d.expanded {
			d.expanded = false
			d.searchInput.Blur()
			d.searchInput.SetValue("")
			d = d.applyFilter()
			return d, nil
		}
		return d, nil
	case "down", "ctrl+n":
		if d.expanded {
			return d.cursorDown(), nil
		}
		d.expanded = true
		d.searchInput.Focus()
		return d, textinput.Blink
	case "up", "ctrl+p":
		if d.expanded {
			return d.cursorUp(), nil
		}
		return d, nil
	case "tab":
		if d.expanded {
			return d.selectCurrentAndClose()
		}
		return d, nil
	}

	// Forward other keys to search input when expanded
	if d.expanded {
		var cmd tea.Cmd
		oldValue := d.searchInput.Value()
		d.searchInput, cmd = d.searchInput.Update(msg)
		if d.searchInput.Value() != oldValue {
			d = d.applyFilter()
		}
		return d, cmd
	}

	return d, nil
}

func (d SearchableDropdown) selectCurrentAndClose() (SearchableDropdown, tea.Cmd) {
	if len(d.filtered) > 0 && d.cursor < len(d.filtered) {
		selectedValue := d.filtered[d.cursor]
		// Find index in original options
		for i, opt := range d.options {
			if opt == selectedValue {
				d.selected = i
				break
			}
		}
	}
	d.expanded = false
	d.searchInput.Blur()
	d.searchInput.SetValue("")
	d = d.applyFilter()
	return d, nil
}

func (d SearchableDropdown) applyFilter() SearchableDropdown {
	query := strings.ToLower(d.searchInput.Value())
	if query == "" {
		d.filtered = d.options
	} else {
		d.filtered = make([]string, 0)
		for _, opt := range d.options {
			if strings.Contains(strings.ToLower(opt), query) {
				d.filtered = append(d.filtered, opt)
			}
		}
	}
	d.cursor = 0
	d.offset = 0
	return d
}

func (d SearchableDropdown) cursorDown() SearchableDropdown {
	if d.cursor < len(d.filtered)-1 {
		d.cursor++
		d = d.adjustOffset()
	}
	return d
}

func (d SearchableDropdown) cursorUp() SearchableDropdown {
	if d.cursor > 0 {
		d.cursor--
		d = d.adjustOffset()
	}
	return d
}

func (d SearchableDropdown) adjustOffset() SearchableDropdown {
	if d.cursor < d.offset {
		d.offset = d.cursor
	}
	if d.cursor >= d.offset+d.maxVisible {
		d.offset = d.cursor - d.maxVisible + 1
	}
	return d
}

// View renders the dropdown.
func (d SearchableDropdown) View() string {
	var result string

	// Label
	if d.label != "" {
		result += d.styles.Label.Render(d.label) + "\n"
	}

	if d.expanded {
		result += d.viewExpanded()
	} else {
		result += d.viewCollapsed()
	}

	// Help text
	if d.help != "" && !d.expanded {
		result += "\n" + d.styles.Help.Render(d.help)
	}

	return result
}

func (d SearchableDropdown) viewCollapsed() string {
	display := "(none selected)"
	if d.selected >= 0 && d.selected < len(d.options) {
		display = d.options[d.selected]
	}
	content := display + " ▼"

	if d.focused {
		return d.styles.Focused.Render(content)
	}
	return d.styles.Container.Render(content)
}

func (d SearchableDropdown) viewExpanded() string {
	var content strings.Builder

	// Search input
	content.WriteString(d.searchInput.View())
	content.WriteString("\n")

	// Options
	if len(d.filtered) == 0 {
		content.WriteString(d.styles.NoMatch.Render("No matches"))
		return d.styles.Focused.Render(content.String())
	}

	d.renderOptionsTo(&content)
	return d.styles.Focused.Render(content.String())
}

func (d SearchableDropdown) renderOptionsTo(content *strings.Builder) {
	end := d.offset + d.maxVisible
	if end > len(d.filtered) {
		end = len(d.filtered)
	}

	// Scroll up indicator
	if d.offset > 0 {
		content.WriteString(d.styles.Help.Render("↑ more"))
		content.WriteString("\n")
	}

	for i := d.offset; i < end; i++ {
		content.WriteString(d.renderOption(i))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll down indicator
	if end < len(d.filtered) {
		content.WriteString("\n")
		content.WriteString(d.styles.Help.Render("↓ more"))
	}
}

func (d SearchableDropdown) renderOption(index int) string {
	opt := d.filtered[index]

	// Cursor indicator
	prefix := "  "
	if index == d.cursor {
		prefix = d.styles.Cursor.Render("▶ ")
	}

	// Check if this is the currently selected value
	isSelected := false
	if d.selected >= 0 && d.selected < len(d.options) {
		isSelected = d.options[d.selected] == opt
	}

	// Selection mark
	suffix := ""
	if isSelected {
		suffix = d.styles.SelectedMark.Render(" ✓")
	}

	optStyle := d.styles.Option
	if index == d.cursor {
		optStyle = d.styles.Cursor
	}

	return prefix + optStyle.Render(opt) + suffix
}

// Focus focuses the dropdown.
func (d SearchableDropdown) Focus() SearchableDropdown {
	d.focused = true
	return d
}

// Blur removes focus from the dropdown.
func (d SearchableDropdown) Blur() SearchableDropdown {
	d.focused = false
	d.expanded = false
	d.searchInput.Blur()
	return d
}

// Focused returns whether the dropdown is focused.
func (d SearchableDropdown) Focused() bool {
	return d.focused
}

// Selected returns the selected index.
func (d SearchableDropdown) Selected() int {
	return d.selected
}

// SelectedValue returns the selected option value.
func (d SearchableDropdown) SelectedValue() string {
	if d.selected >= 0 && d.selected < len(d.options) {
		return d.options[d.selected]
	}
	return ""
}

// IsExpanded returns whether the dropdown is expanded.
func (d SearchableDropdown) IsExpanded() bool {
	return d.expanded
}
