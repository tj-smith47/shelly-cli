package form

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// PasswordOption configures a password input.
type PasswordOption func(*Password)

// WithPasswordLabel sets the password label.
func WithPasswordLabel(label string) PasswordOption {
	return func(p *Password) {
		p.label = label
	}
}

// WithPasswordPlaceholder sets the placeholder text.
func WithPasswordPlaceholder(placeholder string) PasswordOption {
	return func(p *Password) {
		p.input.Placeholder = placeholder
	}
}

// WithPasswordCharLimit sets the maximum character limit.
func WithPasswordCharLimit(limit int) PasswordOption {
	return func(p *Password) {
		p.input.CharLimit = limit
	}
}

// WithPasswordWidth sets the input width.
func WithPasswordWidth(width int) PasswordOption {
	return func(p *Password) {
		p.input.SetWidth(width)
	}
}

// WithPasswordValidation sets a validation function.
func WithPasswordValidation(fn ValidateFunc) PasswordOption {
	return func(p *Password) {
		p.validate = fn
	}
}

// WithPasswordHelp sets the help text.
func WithPasswordHelp(help string) PasswordOption {
	return func(p *Password) {
		p.help = help
	}
}

// WithMaskChar sets the character used to mask the password.
func WithMaskChar(char rune) PasswordOption {
	return func(p *Password) {
		p.input.EchoCharacter = char
	}
}

// Password is a password input component with masking.
type Password struct {
	input    textinput.Model
	label    string
	help     string
	err      error
	validate ValidateFunc
	styles   TextInputStyles
	showPass bool
}

// NewPassword creates a new password input with options.
func NewPassword(opts ...PasswordOption) Password {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

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

	p := Password{
		input:  ti,
		styles: DefaultTextInputStyles(),
	}

	for _, opt := range opts {
		opt(&p)
	}

	return p
}

// Init returns the initial command.
func (p Password) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (p Password) Update(msg tea.Msg) (Password, tea.Cmd) {
	// Handle toggle visibility
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "ctrl+t" {
		return p.toggleVisibility(), nil
	}

	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)

	// Run validation on value change
	if p.validate != nil {
		p.err = p.validate(p.input.Value())
	}

	return p, cmd
}

func (p Password) toggleVisibility() Password {
	p.showPass = !p.showPass
	if p.showPass {
		p.input.EchoMode = textinput.EchoNormal
	} else {
		p.input.EchoMode = textinput.EchoPassword
	}
	return p
}

// View renders the password input.
func (p Password) View() string {
	colors := theme.GetSemanticColors()
	var result string

	// Label with toggle hint
	if p.label != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true)
		hintStyle := lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true)
		hint := " (Ctrl+T to toggle visibility)"
		result += labelStyle.Render(p.label) + hintStyle.Render(hint) + "\n"
	}

	// Input
	result += p.input.View()

	// Error message
	if p.err != nil {
		result += "\n" + p.styles.Error.Render(p.err.Error())
	} else if p.help != "" {
		result += "\n" + p.styles.Help.Render(p.help)
	}

	return result
}

// Focus focuses the input.
func (p Password) Focus() (Password, tea.Cmd) {
	p.input.Focus()
	return p, textinput.Blink
}

// Blur removes focus from the input.
func (p Password) Blur() Password {
	p.input.Blur()
	return p
}

// Focused returns whether the input is focused.
func (p Password) Focused() bool {
	return p.input.Focused()
}

// Value returns the current value (unmasked).
func (p Password) Value() string {
	return p.input.Value()
}

// SetValue sets the input value.
func (p Password) SetValue(value string) Password {
	p.input.SetValue(value)
	if p.validate != nil {
		p.err = p.validate(value)
	}
	return p
}

// SetWidth sets the input width.
func (p Password) SetWidth(width int) Password {
	p.input.SetWidth(width)
	return p
}

// Error returns any validation error.
func (p Password) Error() error {
	return p.err
}

// Valid returns whether the input is valid.
func (p Password) Valid() bool {
	return p.err == nil
}

// Reset clears the input value and error.
func (p Password) Reset() Password {
	p.input.SetValue("")
	p.err = nil
	return p
}

// IsVisible returns whether the password is currently visible.
func (p Password) IsVisible() bool {
	return p.showPass
}
