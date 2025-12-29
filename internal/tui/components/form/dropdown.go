package form

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DropdownStyles holds the visual styles for a dropdown.
type DropdownStyles struct {
	Label        lipgloss.Style
	Option       lipgloss.Style
	SelectedMark lipgloss.Style
	Cursor       lipgloss.Style
	Help         lipgloss.Style
	Container    lipgloss.Style
	Focused      lipgloss.Style
}

// DefaultDropdownStyles returns the default styles for a dropdown.
func DefaultDropdownStyles() DropdownStyles {
	colors := theme.GetSemanticColors()
	return DropdownStyles{
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
	}
}

// DropdownOption configures a dropdown.
type DropdownOption func(*Dropdown)

// WithDropdownLabel sets the dropdown label.
func WithDropdownLabel(label string) DropdownOption {
	return func(d *Dropdown) {
		d.label = label
	}
}

// WithDropdownOptions sets the dropdown options.
func WithDropdownOptions(options []string) DropdownOption {
	return func(d *Dropdown) {
		d.options = options
	}
}

// WithDropdownHelp sets the help text.
func WithDropdownHelp(help string) DropdownOption {
	return func(d *Dropdown) {
		d.help = help
	}
}

// WithDropdownStyles sets custom styles.
func WithDropdownStyles(styles DropdownStyles) DropdownOption {
	return func(d *Dropdown) {
		d.styles = styles
	}
}

// WithDropdownSelected sets the initially selected index.
func WithDropdownSelected(index int) DropdownOption {
	return func(d *Dropdown) {
		d.selected = index
	}
}

// WithDropdownMaxVisible sets the maximum visible options.
func WithDropdownMaxVisible(maxItems int) DropdownOption {
	return func(d *Dropdown) {
		d.maxVisible = maxItems
	}
}

// Dropdown is a single-select dropdown component.
type Dropdown struct {
	label      string
	options    []string
	selected   int
	cursor     int
	help       string
	expanded   bool
	focused    bool
	maxVisible int
	offset     int
	styles     DropdownStyles
}

// NewDropdown creates a new dropdown with options.
func NewDropdown(opts ...DropdownOption) Dropdown {
	d := Dropdown{
		maxVisible: 5,
		styles:     DefaultDropdownStyles(),
	}

	for _, opt := range opts {
		opt(&d)
	}

	// Initialize cursor to selected
	d.cursor = d.selected

	return d
}

// Init returns the initial command.
func (d Dropdown) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (d Dropdown) Update(msg tea.Msg) (Dropdown, tea.Cmd) {
	if !d.focused {
		return d, nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return d, nil
	}

	return d.handleKey(keyMsg), nil
}

func (d Dropdown) handleKey(keyMsg tea.KeyPressMsg) Dropdown {
	switch keyMsg.String() {
	case "enter", " ":
		return d.handleEnter()
	case "esc", "ctrl+[":
		return d.handleEsc()
	case "j", "down":
		return d.handleDown()
	case "k", "up":
		return d.handleUp()
	case "g":
		return d.handleHome()
	case "G":
		return d.handleEnd()
	}
	return d
}

func (d Dropdown) handleEnter() Dropdown {
	if d.expanded {
		d.selected = d.cursor
		d.expanded = false
	} else {
		d.expanded = true
		d.cursor = d.selected
	}
	return d
}

func (d Dropdown) handleEsc() Dropdown {
	if d.expanded {
		d.expanded = false
		d.cursor = d.selected
	}
	return d
}

func (d Dropdown) handleDown() Dropdown {
	if d.expanded {
		return d.cursorDown()
	}
	d.expanded = true
	return d
}

func (d Dropdown) handleUp() Dropdown {
	if d.expanded {
		return d.cursorUp()
	}
	return d
}

func (d Dropdown) handleHome() Dropdown {
	if d.expanded {
		d.cursor = 0
		d.offset = 0
	}
	return d
}

func (d Dropdown) handleEnd() Dropdown {
	if d.expanded {
		d.cursor = len(d.options) - 1
		d = d.adjustOffset()
	}
	return d
}

func (d Dropdown) cursorDown() Dropdown {
	if d.cursor < len(d.options)-1 {
		d.cursor++
		d = d.adjustOffset()
	}
	return d
}

func (d Dropdown) cursorUp() Dropdown {
	if d.cursor > 0 {
		d.cursor--
		d = d.adjustOffset()
	}
	return d
}

func (d Dropdown) adjustOffset() Dropdown {
	// Ensure cursor is visible
	if d.cursor < d.offset {
		d.offset = d.cursor
	}
	if d.cursor >= d.offset+d.maxVisible {
		d.offset = d.cursor - d.maxVisible + 1
	}
	return d
}

// View renders the dropdown.
func (d Dropdown) View() string {
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
	if d.help != "" {
		result += "\n" + d.styles.Help.Render(d.help)
	}

	return result
}

func (d Dropdown) viewCollapsed() string {
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

func (d Dropdown) viewExpanded() string {
	var options string
	end := d.offset + d.maxVisible
	if end > len(d.options) {
		end = len(d.options)
	}

	for i := d.offset; i < end; i++ {
		options += d.renderOption(i, i < end-1)
	}

	// Scroll indicators
	if d.offset > 0 {
		options = d.styles.Help.Render("↑ more") + "\n" + options
	}
	if end < len(d.options) {
		options += "\n" + d.styles.Help.Render("↓ more")
	}

	return d.styles.Focused.Render(options)
}

func (d Dropdown) renderOption(index int, addNewline bool) string {
	opt := d.options[index]

	// Cursor indicator
	prefix := "  "
	if index == d.cursor {
		prefix = d.styles.Cursor.Render("▶ ")
	}

	// Selection mark
	suffix := ""
	if index == d.selected {
		suffix = d.styles.SelectedMark.Render(" ✓")
	}

	optStyle := d.styles.Option
	if index == d.cursor {
		optStyle = d.styles.Cursor
	}

	result := prefix + optStyle.Render(opt) + suffix
	if addNewline {
		result += "\n"
	}
	return result
}

// Focus focuses the dropdown.
func (d Dropdown) Focus() Dropdown {
	d.focused = true
	return d
}

// Blur removes focus from the dropdown.
func (d Dropdown) Blur() Dropdown {
	d.focused = false
	d.expanded = false
	return d
}

// Focused returns whether the dropdown is focused.
func (d Dropdown) Focused() bool {
	return d.focused
}

// Selected returns the selected index.
func (d Dropdown) Selected() int {
	return d.selected
}

// SelectedValue returns the selected option value.
func (d Dropdown) SelectedValue() string {
	if d.selected >= 0 && d.selected < len(d.options) {
		return d.options[d.selected]
	}
	return ""
}

// SetSelected sets the selected index.
func (d Dropdown) SetSelected(index int) Dropdown {
	if index >= 0 && index < len(d.options) {
		d.selected = index
		d.cursor = index
	}
	return d
}

// SetOptions updates the dropdown options.
func (d Dropdown) SetOptions(options []string) Dropdown {
	d.options = options
	if d.selected >= len(options) {
		d.selected = len(options) - 1
	}
	if d.selected < 0 {
		d.selected = 0
	}
	d.cursor = d.selected
	d.offset = 0
	return d
}

// IsExpanded returns whether the dropdown is expanded.
func (d Dropdown) IsExpanded() bool {
	return d.expanded
}

// Collapse collapses the dropdown.
func (d Dropdown) Collapse() Dropdown {
	d.expanded = false
	return d
}

// Expand expands the dropdown.
func (d Dropdown) Expand() Dropdown {
	d.expanded = true
	return d
}
