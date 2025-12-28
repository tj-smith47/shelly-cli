package form

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// TextAreaStyles holds the visual styles for a textarea.
type TextAreaStyles struct {
	Label lipgloss.Style
	Error lipgloss.Style
	Help  lipgloss.Style
}

// DefaultTextAreaStyles returns the default styles for a textarea.
func DefaultTextAreaStyles() TextAreaStyles {
	colors := theme.GetSemanticColors()
	return TextAreaStyles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true).
			MarginBottom(1),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// TextAreaOption configures a textarea.
type TextAreaOption func(*TextArea)

// WithTextAreaLabel sets the textarea label.
func WithTextAreaLabel(label string) TextAreaOption {
	return func(ta *TextArea) {
		ta.label = label
	}
}

// WithTextAreaPlaceholder sets the placeholder text.
func WithTextAreaPlaceholder(placeholder string) TextAreaOption {
	return func(ta *TextArea) {
		ta.input.Placeholder = placeholder
	}
}

// WithTextAreaCharLimit sets the maximum character limit.
func WithTextAreaCharLimit(limit int) TextAreaOption {
	return func(ta *TextArea) {
		ta.input.CharLimit = limit
	}
}

// WithTextAreaDimensions sets the width and height.
func WithTextAreaDimensions(width, height int) TextAreaOption {
	return func(ta *TextArea) {
		ta.input.SetWidth(width)
		ta.input.SetHeight(height)
	}
}

// WithTextAreaValidation sets a validation function.
func WithTextAreaValidation(fn ValidateFunc) TextAreaOption {
	return func(ta *TextArea) {
		ta.validate = fn
	}
}

// WithTextAreaHelp sets the help text.
func WithTextAreaHelp(help string) TextAreaOption {
	return func(ta *TextArea) {
		ta.help = help
	}
}

// WithTextAreaStyles sets custom styles.
func WithTextAreaStyles(styles TextAreaStyles) TextAreaOption {
	return func(ta *TextArea) {
		ta.styles = styles
	}
}

// WithTextAreaShowLineNumbers sets whether to show line numbers.
func WithTextAreaShowLineNumbers(show bool) TextAreaOption {
	return func(ta *TextArea) {
		ta.input.ShowLineNumbers = show
	}
}

// TextArea is a multi-line text input component.
type TextArea struct {
	input    textarea.Model
	label    string
	help     string
	err      error
	validate ValidateFunc
	styles   TextAreaStyles
}

// NewTextArea creates a new textarea with options.
func NewTextArea(opts ...TextAreaOption) TextArea {
	ti := textarea.New()

	// Configure default styles using semantic colors
	colors := theme.GetSemanticColors()
	inputStyles := textarea.DefaultStyles(true)
	inputStyles.Focused.Base = inputStyles.Focused.Base.
		BorderForeground(colors.Highlight)
	inputStyles.Focused.Text = inputStyles.Focused.Text.
		Foreground(colors.Text)
	inputStyles.Focused.Placeholder = inputStyles.Focused.Placeholder.
		Foreground(colors.Muted)
	inputStyles.Focused.LineNumber = inputStyles.Focused.LineNumber.
		Foreground(colors.Muted)
	inputStyles.Blurred.Base = inputStyles.Blurred.Base.
		BorderForeground(colors.TableBorder)
	inputStyles.Blurred.Text = inputStyles.Blurred.Text.
		Foreground(colors.Text)
	inputStyles.Blurred.Placeholder = inputStyles.Blurred.Placeholder.
		Foreground(colors.Muted)
	inputStyles.Blurred.LineNumber = inputStyles.Blurred.LineNumber.
		Foreground(colors.Muted)
	ti.SetStyles(inputStyles)

	ta := TextArea{
		input:  ti,
		styles: DefaultTextAreaStyles(),
	}

	for _, opt := range opts {
		opt(&ta)
	}

	return ta
}

// Init returns the initial command.
func (ta TextArea) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (ta TextArea) Update(msg tea.Msg) (TextArea, tea.Cmd) {
	var cmd tea.Cmd
	ta.input, cmd = ta.input.Update(msg)

	// Run validation on value change
	if ta.validate != nil {
		ta.err = ta.validate(ta.input.Value())
	}

	return ta, cmd
}

// View renders the textarea.
func (ta TextArea) View() string {
	var result string

	// Label
	if ta.label != "" {
		result += ta.styles.Label.Render(ta.label) + "\n"
	}

	// Input
	result += ta.input.View()

	// Error message
	if ta.err != nil {
		result += "\n" + ta.styles.Error.Render(ta.err.Error())
	} else if ta.help != "" {
		result += "\n" + ta.styles.Help.Render(ta.help)
	}

	return result
}

// Focus focuses the textarea.
func (ta TextArea) Focus() (TextArea, tea.Cmd) {
	ta.input.Focus()
	return ta, textarea.Blink
}

// Blur removes focus from the textarea.
func (ta TextArea) Blur() TextArea {
	ta.input.Blur()
	return ta
}

// Focused returns whether the textarea is focused.
func (ta TextArea) Focused() bool {
	return ta.input.Focused()
}

// Value returns the current value.
func (ta TextArea) Value() string {
	return ta.input.Value()
}

// SetValue sets the textarea value.
func (ta TextArea) SetValue(value string) TextArea {
	ta.input.SetValue(value)
	if ta.validate != nil {
		ta.err = ta.validate(value)
	}
	return ta
}

// SetDimensions sets the width and height.
func (ta TextArea) SetDimensions(width, height int) TextArea {
	ta.input.SetWidth(width)
	ta.input.SetHeight(height)
	return ta
}

// Error returns any validation error.
func (ta TextArea) Error() error {
	return ta.err
}

// Valid returns whether the textarea is valid.
func (ta TextArea) Valid() bool {
	return ta.err == nil
}

// Reset clears the textarea value and error.
func (ta TextArea) Reset() TextArea {
	ta.input.SetValue("")
	ta.err = nil
	return ta
}

// LineCount returns the number of lines.
func (ta TextArea) LineCount() int {
	return ta.input.LineCount()
}

// CursorPosition returns the current cursor position (line, column).
func (ta TextArea) CursorPosition() (line, column int) {
	return ta.input.Line(), ta.input.LineInfo().ColumnOffset
}
