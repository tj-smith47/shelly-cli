// Package form provides reusable form input components for the TUI.
package form

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// TextInputStyles holds the visual styles for a text input.
type TextInputStyles struct {
	Label       lipgloss.Style
	Input       lipgloss.Style
	Placeholder lipgloss.Style
	Error       lipgloss.Style
	Help        lipgloss.Style
}

// DefaultTextInputStyles returns the default styles for a text input.
func DefaultTextInputStyles() TextInputStyles {
	colors := theme.GetSemanticColors()
	return TextInputStyles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true).
			MarginRight(1),
		Input: lipgloss.NewStyle().
			Foreground(colors.Text),
		Placeholder: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// ValidateFunc is a function that validates text input.
type ValidateFunc func(string) error

// TextInputOption configures a text input.
type TextInputOption func(*TextInput)

// WithLabel sets the input label.
func WithLabel(label string) TextInputOption {
	return func(ti *TextInput) {
		ti.label = label
	}
}

// WithPlaceholder sets the placeholder text.
func WithPlaceholder(placeholder string) TextInputOption {
	return func(ti *TextInput) {
		ti.input.Placeholder = placeholder
	}
}

// WithCharLimit sets the maximum character limit.
func WithCharLimit(limit int) TextInputOption {
	return func(ti *TextInput) {
		ti.input.CharLimit = limit
	}
}

// WithWidth sets the input width.
func WithWidth(width int) TextInputOption {
	return func(ti *TextInput) {
		ti.input.SetWidth(width)
	}
}

// WithValidation sets a validation function.
func WithValidation(fn ValidateFunc) TextInputOption {
	return func(ti *TextInput) {
		ti.validate = fn
	}
}

// WithHelp sets the help text.
func WithHelp(help string) TextInputOption {
	return func(ti *TextInput) {
		ti.help = help
	}
}

// WithTextInputStyles sets custom styles.
func WithTextInputStyles(styles TextInputStyles) TextInputOption {
	return func(ti *TextInput) {
		ti.styles = styles
	}
}

// TextInput is a styled text input component.
type TextInput struct {
	input    textinput.Model
	label    string
	help     string
	err      error
	validate ValidateFunc
	styles   TextInputStyles
}

// NewTextInput creates a new text input with options.
func NewTextInput(opts ...TextInputOption) TextInput {
	ti := textinput.New()

	// Configure default styles using semantic colors
	colors := theme.GetSemanticColors()
	inputStyles := textinput.DefaultStyles(true)
	inputStyles.Focused.Prompt = inputStyles.Focused.Prompt.Foreground(colors.Highlight)
	inputStyles.Focused.Text = inputStyles.Focused.Text.Foreground(colors.Text)
	inputStyles.Focused.Placeholder = inputStyles.Focused.Placeholder.Foreground(colors.Muted)
	inputStyles.Blurred.Prompt = inputStyles.Blurred.Prompt.Foreground(colors.Highlight)
	inputStyles.Blurred.Text = inputStyles.Blurred.Text.Foreground(colors.Text)
	inputStyles.Blurred.Placeholder = inputStyles.Blurred.Placeholder.Foreground(colors.Muted)
	ti.SetStyles(inputStyles)

	input := TextInput{
		input:  ti,
		styles: DefaultTextInputStyles(),
	}

	for _, opt := range opts {
		opt(&input)
	}

	return input
}

// Init returns the initial command.
func (ti TextInput) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (ti TextInput) Update(msg tea.Msg) (TextInput, tea.Cmd) {
	var cmd tea.Cmd
	ti.input, cmd = ti.input.Update(msg)

	// Run validation on value change
	if ti.validate != nil {
		ti.err = ti.validate(ti.input.Value())
	}

	return ti, cmd
}

// View renders the text input.
func (ti TextInput) View() string {
	var result string

	// Label
	if ti.label != "" {
		result += ti.styles.Label.Render(ti.label) + "\n"
	}

	// Input
	result += ti.input.View()

	// Error message
	if ti.err != nil {
		result += "\n" + ti.styles.Error.Render(ti.err.Error())
	} else if ti.help != "" {
		result += "\n" + ti.styles.Help.Render(ti.help)
	}

	return result
}

// Focus focuses the input.
func (ti TextInput) Focus() (TextInput, tea.Cmd) {
	ti.input.Focus()
	return ti, textinput.Blink
}

// Blur removes focus from the input.
func (ti TextInput) Blur() TextInput {
	ti.input.Blur()
	return ti
}

// Focused returns whether the input is focused.
func (ti TextInput) Focused() bool {
	return ti.input.Focused()
}

// Value returns the current value.
func (ti TextInput) Value() string {
	return ti.input.Value()
}

// SetValue sets the input value.
func (ti TextInput) SetValue(value string) TextInput {
	ti.input.SetValue(value)
	if ti.validate != nil {
		ti.err = ti.validate(value)
	}
	return ti
}

// SetWidth sets the input width.
func (ti TextInput) SetWidth(width int) TextInput {
	ti.input.SetWidth(width)
	return ti
}

// Error returns any validation error.
func (ti TextInput) Error() error {
	return ti.err
}

// Valid returns whether the input is valid.
func (ti TextInput) Valid() bool {
	return ti.err == nil
}

// Reset clears the input value and error.
func (ti TextInput) Reset() TextInput {
	ti.input.SetValue("")
	ti.err = nil
	return ti
}
