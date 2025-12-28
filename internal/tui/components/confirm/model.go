// Package confirm provides a confirmation dialog for dangerous operations.
// It requires users to type a confirmation phrase before proceeding.
package confirm

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/modal"
)

// ConfirmedMsg signals that the user confirmed the operation.
type ConfirmedMsg struct {
	Operation string
}

// CancelledMsg signals that the user cancelled the operation.
type CancelledMsg struct {
	Operation string
}

// Model displays a confirmation dialog requiring typed confirmation.
type Model struct {
	operation     string // Name of the operation (e.g., "factory-reset")
	title         string
	message       string
	confirmPhrase string // Phrase user must type
	input         string // Current user input
	width         int
	height        int
	visible       bool
	useModal      bool        // Whether to use modal overlay
	modal         modal.Model // Modal component for overlay support
	styles        Styles
}

// Styles holds styles for the Confirm component.
type Styles struct {
	Border    lipgloss.Style
	Title     lipgloss.Style
	Message   lipgloss.Style
	Warning   lipgloss.Style
	Input     lipgloss.Style
	Prompt    lipgloss.Style
	Muted     lipgloss.Style
	Danger    lipgloss.Style
	Highlight lipgloss.Style
}

// DefaultStyles returns the default styles for the Confirm component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Error).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
		Message: lipgloss.NewStyle().
			Foreground(colors.Text),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Input: lipgloss.NewStyle().
			Foreground(colors.Text).
			Background(colors.AltBackground).
			Padding(0, 1),
		Prompt: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Danger: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
		Highlight: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// Option configures the confirm model.
type Option func(*Model)

// WithModalOverlay enables modal overlay mode.
func WithModalOverlay() Option {
	return func(m *Model) {
		m.useModal = true
	}
}

// WithStyles sets custom styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.styles = styles
	}
}

// New creates a new Confirm model.
func New(opts ...Option) Model {
	m := Model{
		styles: DefaultStyles(),
		modal: modal.New(
			modal.WithCloseOnEsc(false),     // We handle Esc ourselves
			modal.WithConfirmOnEnter(false), // We handle Enter ourselves
		),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Show displays the confirmation dialog for a dangerous operation.
func (m Model) Show(operation, title, message, confirmPhrase string) Model {
	m.operation = operation
	m.title = title
	m.message = message
	m.confirmPhrase = confirmPhrase
	m.input = ""
	m.visible = true
	if m.useModal {
		m.modal = m.modal.SetTitle("⚠ " + title).Show()
	}
	return m
}

// Hide hides the confirmation dialog.
func (m Model) Hide() Model {
	m.visible = false
	m.input = ""
	if m.useModal {
		m.modal = m.modal.Hide()
	}
	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	if m.useModal {
		m.modal = m.modal.SetSize(width, height)
	}
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKey(keyMsg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.visible = false
		m.input = ""
		return m, func() tea.Msg {
			return CancelledMsg{Operation: m.operation}
		}
	case "enter":
		if m.inputMatches() {
			m.visible = false
			m.input = ""
			return m, func() tea.Msg {
				return ConfirmedMsg{Operation: m.operation}
			}
		}
	case "backspace":
		if m.input != "" {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		// Single character input
		if len(msg.String()) == 1 && len(m.input) < len(m.confirmPhrase)+10 {
			m.input += msg.String()
		}
	}

	return m, nil
}

func (m Model) inputMatches() bool {
	return strings.EqualFold(m.input, m.confirmPhrase)
}

// View renders the Confirm component.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	content := m.buildContent()

	if m.useModal {
		// When using modal, update modal content and return its view
		m.modal = m.modal.SetContent(content)
		return m.modal.View()
	}

	// Traditional bordered view
	dialog := m.styles.Border.Render(content)

	// Center the dialog
	if m.width > 0 {
		dialogWidth := lipgloss.Width(dialog)
		if dialogWidth < m.width {
			padding := (m.width - dialogWidth) / 2
			dialog = strings.Repeat(" ", padding) + dialog
		}
	}

	return dialog
}

func (m Model) buildContent() string {
	var content strings.Builder

	// Message
	content.WriteString(m.styles.Message.Render(m.message))
	content.WriteString("\n\n")

	// Warning about confirmation
	content.WriteString(m.styles.Warning.Render(
		fmt.Sprintf("Type %q to confirm:", m.confirmPhrase),
	))
	content.WriteString("\n")

	// Input field
	inputText := m.input
	if inputText == "" {
		inputText = " "
	}
	content.WriteString(m.styles.Input.Render(inputText))
	content.WriteString("\n\n")

	// Match indicator
	if m.input != "" {
		if m.inputMatches() {
			content.WriteString(m.styles.Highlight.Render("✓ Matches - Press Enter to confirm"))
		} else {
			content.WriteString(m.styles.Muted.Render("✗ Does not match"))
		}
		content.WriteString("\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("Esc: cancel"))

	return content.String()
}

// Overlay renders the confirmation dialog as an overlay on base content.
func (m Model) Overlay(base string) string {
	if !m.visible {
		return base
	}

	if m.useModal {
		m.modal = m.modal.SetContent(m.buildContent())
		return m.modal.Overlay(base)
	}

	// Fallback: just return the view for non-modal mode
	return m.View()
}

// Visible returns whether the dialog is visible.
func (m Model) Visible() bool {
	return m.visible
}

// Operation returns the current operation name.
func (m Model) Operation() string {
	return m.operation
}

// Input returns the current input text.
func (m Model) Input() string {
	return m.input
}

// ConfirmPhrase returns the required confirmation phrase.
func (m Model) ConfirmPhrase() string {
	return m.confirmPhrase
}

// InputMatches returns whether the current input matches the confirmation phrase.
func (m Model) InputMatches() bool {
	return m.inputMatches()
}
