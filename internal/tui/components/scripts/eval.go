// Package scripts provides TUI components for managing device scripts.
package scripts

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// EvalResultMsg signals the result of evaluating a script expression.
type EvalResultMsg struct {
	Device   string
	ScriptID int
	Code     string
	Result   any
	Err      error
}

// EvalModel represents the script eval input modal.
type EvalModel struct {
	ctx      context.Context
	svc      *automation.Service
	device   string
	scriptID int
	visible  bool
	input    form.TextInput
	result   string
	err      error
	running  bool
	width    int
	height   int
	styles   evalStyles
}

type evalStyles struct {
	Label   lipgloss.Style
	Result  lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style
	Muted   lipgloss.Style
}

func defaultEvalStyles() evalStyles {
	colors := theme.GetSemanticColors()
	return evalStyles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Result: lipgloss.NewStyle().
			Foreground(colors.Text),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Success: lipgloss.NewStyle().
			Foreground(colors.Success),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewEvalModel creates a new script eval modal.
func NewEvalModel(ctx context.Context, svc *automation.Service) EvalModel {
	input := form.NewTextInput(
		form.WithPlaceholder("Enter JavaScript expression..."),
		form.WithCharLimit(256),
		form.WithWidth(60),
		form.WithHelp("Expression to evaluate in running script"),
	)

	return EvalModel{
		ctx:    ctx,
		svc:    svc,
		input:  input,
		styles: defaultEvalStyles(),
	}
}

// Show displays the eval modal.
func (m EvalModel) Show(device string, scriptID int) EvalModel {
	m.device = device
	m.scriptID = scriptID
	m.visible = true
	m.result = ""
	m.err = nil
	m.running = false

	// Reset and focus input
	m.input = m.input.SetValue("")
	m.input, _ = m.input.Focus()

	return m
}

// Hide hides the eval modal.
func (m EvalModel) Hide() EvalModel {
	m.visible = false
	m.input = m.input.Blur()
	return m
}

// IsVisible returns whether the modal is visible.
func (m EvalModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m EvalModel) SetSize(width, height int) EvalModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m EvalModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m EvalModel) Update(msg tea.Msg) (EvalModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case EvalResultMsg:
		m.running = false
		if msg.Err != nil {
			m.err = msg.Err
			m.result = ""
			return m, nil
		}
		m.result = formatResult(msg.Result)
		m.err = nil
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m EvalModel) handleKey(msg tea.KeyPressMsg) (EvalModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEsc, "q":
		m = m.Hide()
		return m, func() tea.Msg { return messages.EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter:
		code := strings.TrimSpace(m.input.Value())
		if code == "" || m.running {
			return m, nil
		}
		m.running = true
		m.result = ""
		m.err = nil
		return m, m.evalCmd(code)

	case "ctrl+c":
		// Clear input
		m.input = m.input.SetValue("")
		m.result = ""
		m.err = nil
	}

	// Forward to input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m EvalModel) evalCmd(code string) tea.Cmd {
	device := m.device
	scriptID := m.scriptID

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		result, err := m.svc.EvalScript(ctx, device, scriptID, code)
		return EvalResultMsg{
			Device:   device,
			ScriptID: scriptID,
			Code:     code,
			Result:   result,
			Err:      err,
		}
	}
}

// View renders the eval modal.
func (m EvalModel) View() string {
	if !m.visible {
		return ""
	}

	footer := "Enter: Evaluate | Ctrl+C: Clear | Esc: Close"
	if m.running {
		footer = "Evaluating..."
	}

	title := fmt.Sprintf("Eval Script %d", m.scriptID)
	r := rendering.NewModal(m.width, m.height, title, footer)
	return r.SetContent(m.renderContent()).Render()
}

func (m EvalModel) renderContent() string {
	var content strings.Builder

	// Input label and field
	content.WriteString(m.styles.Label.Render("Expression:"))
	content.WriteString("\n")
	content.WriteString(m.input.View())
	content.WriteString("\n\n")

	// Result or error
	switch {
	case m.running:
		content.WriteString(m.styles.Muted.Render("Evaluating..."))
	case m.err != nil:
		content.WriteString(m.styles.Label.Render("Error:"))
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render(m.err.Error()))
	case m.result != "":
		content.WriteString(m.styles.Label.Render("Result:"))
		content.WriteString("\n")
		content.WriteString(m.styles.Success.Render(m.result))
	default:
		content.WriteString(m.styles.Muted.Render("Enter a JavaScript expression to evaluate"))
	}

	return content.String()
}

func formatResult(result any) string {
	if result == nil {
		return "null"
	}
	return fmt.Sprintf("%v", result)
}
