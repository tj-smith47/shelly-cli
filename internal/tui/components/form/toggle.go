package form

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ToggleStyles holds the visual styles for a toggle.
type ToggleStyles struct {
	Label     lipgloss.Style
	On        lipgloss.Style
	Off       lipgloss.Style
	Track     lipgloss.Style
	TrackOn   lipgloss.Style
	TrackOff  lipgloss.Style
	Help      lipgloss.Style
	Focused   lipgloss.Style
	Unfocused lipgloss.Style
}

// DefaultToggleStyles returns the default styles for a toggle.
func DefaultToggleStyles() ToggleStyles {
	colors := theme.GetSemanticColors()
	return ToggleStyles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true).
			MarginRight(1),
		On: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		Off: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Track: lipgloss.NewStyle().
			Foreground(colors.Muted),
		TrackOn: lipgloss.NewStyle().
			Foreground(colors.Online),
		TrackOff: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Focused: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Unfocused: lipgloss.NewStyle().
			Foreground(colors.Text),
	}
}

// ToggleOption configures a toggle.
type ToggleOption func(*Toggle)

// WithToggleLabel sets the toggle label.
func WithToggleLabel(label string) ToggleOption {
	return func(t *Toggle) {
		t.label = label
	}
}

// WithToggleOnLabel sets the label when toggle is on.
func WithToggleOnLabel(label string) ToggleOption {
	return func(t *Toggle) {
		t.onLabel = label
	}
}

// WithToggleOffLabel sets the label when toggle is off.
func WithToggleOffLabel(label string) ToggleOption {
	return func(t *Toggle) {
		t.offLabel = label
	}
}

// WithToggleHelp sets the help text.
func WithToggleHelp(help string) ToggleOption {
	return func(t *Toggle) {
		t.help = help
	}
}

// WithToggleStyles sets custom styles.
func WithToggleStyles(styles ToggleStyles) ToggleOption {
	return func(t *Toggle) {
		t.styles = styles
	}
}

// WithToggleValue sets the initial value.
func WithToggleValue(value bool) ToggleOption {
	return func(t *Toggle) {
		t.value = value
	}
}

// Toggle is a boolean toggle component.
type Toggle struct {
	label    string
	onLabel  string
	offLabel string
	help     string
	value    bool
	focused  bool
	styles   ToggleStyles
}

// NewToggle creates a new toggle with options.
func NewToggle(opts ...ToggleOption) Toggle {
	t := Toggle{
		onLabel:  "On",
		offLabel: "Off",
		styles:   DefaultToggleStyles(),
	}

	for _, opt := range opts {
		opt(&t)
	}

	return t
}

// Init returns the initial command.
func (t Toggle) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (t Toggle) Update(msg tea.Msg) (Toggle, tea.Cmd) {
	if !t.focused {
		return t, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "enter", "space", "tab":
			t.value = !t.value
		case "y", "Y":
			t.value = true
		case "n", "N":
			t.value = false
		}
	}

	return t, nil
}

// View renders the toggle.
func (t Toggle) View() string {
	var result string

	// Label
	if t.label != "" {
		result += t.styles.Label.Render(t.label) + " "
	}

	// Toggle track
	var track string
	if t.value {
		track = t.styles.TrackOn.Render("[") +
			t.styles.On.Render("●") +
			t.styles.TrackOn.Render("○]") +
			" " + t.styles.On.Render(t.onLabel)
	} else {
		track = t.styles.TrackOff.Render("[○") +
			t.styles.Off.Render("●") +
			t.styles.TrackOff.Render("]") +
			" " + t.styles.Off.Render(t.offLabel)
	}

	if t.focused {
		result += t.styles.Focused.Render(track)
	} else {
		result += t.styles.Unfocused.Render(track)
	}

	// Help text
	if t.help != "" {
		result += "\n" + t.styles.Help.Render(t.help)
	}

	return result
}

// Focus focuses the toggle.
func (t Toggle) Focus() Toggle {
	t.focused = true
	return t
}

// Blur removes focus from the toggle.
func (t Toggle) Blur() Toggle {
	t.focused = false
	return t
}

// Focused returns whether the toggle is focused.
func (t Toggle) Focused() bool {
	return t.focused
}

// Value returns the current value.
func (t Toggle) Value() bool {
	return t.value
}

// SetValue sets the toggle value.
func (t Toggle) SetValue(value bool) Toggle {
	t.value = value
	return t
}

// Toggle toggles the value.
func (t Toggle) Toggle() Toggle {
	t.value = !t.value
	return t
}
