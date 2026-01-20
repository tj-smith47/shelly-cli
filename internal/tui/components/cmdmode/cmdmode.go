// Package cmdmode provides a vim-style command mode component for the TUI.
package cmdmode

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	tuistyles "github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// Command types that can be executed.
const (
	CmdQuit   = "quit"
	CmdDevice = "device"
	CmdFilter = "filter"
	CmdTheme  = "theme"
	CmdView   = "view"
	CmdHelp   = "help"
	CmdToggle = "toggle"
)

// CommandMsg is sent when a command is executed.
type CommandMsg struct {
	Command string // The command name (quit, device, filter, theme, view, help)
	Args    string // Arguments passed to the command
}

// ErrorMsg is sent when a command fails to parse.
type ErrorMsg struct {
	Message string
}

// ClosedMsg signals that command mode was closed without executing.
type ClosedMsg struct{}

// Model holds the command mode state.
type Model struct {
	textInput textinput.Model
	active    bool
	width     int
	styles    Styles
	history   []string
	histIdx   int
}

// Styles for the command mode component.
type Styles struct {
	Container lipgloss.Style
	Prompt    lipgloss.Style
	Input     lipgloss.Style
	Error     lipgloss.Style
}

// DefaultStyles returns default styles for the command mode.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: tuistyles.PrimaryBorder().Padding(0, 1),
		Prompt: lipgloss.NewStyle().
			Foreground(colors.Primary).
			Bold(true),
		Input: lipgloss.NewStyle().
			Foreground(colors.Text),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
	}
}

// New creates a new command mode model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "command..."
	ti.CharLimit = 100
	ti.SetWidth(40)

	// Configure styles using semantic colors
	colors := theme.GetSemanticColors()
	styles := textinput.DefaultStyles(true) // dark mode
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(colors.Primary)
	styles.Focused.Text = styles.Focused.Text.Foreground(colors.Text)
	styles.Focused.Placeholder = styles.Focused.Placeholder.Foreground(colors.Muted)
	styles.Blurred.Prompt = styles.Blurred.Prompt.Foreground(colors.Primary)
	styles.Blurred.Text = styles.Blurred.Text.Foreground(colors.Text)
	styles.Blurred.Placeholder = styles.Blurred.Placeholder.Foreground(colors.Muted)
	ti.SetStyles(styles)

	return Model{
		textInput: ti,
		styles:    DefaultStyles(),
		history:   make([]string, 0, 50),
		histIdx:   -1,
	}
}

// Init initializes the command mode component.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the command mode component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, key.NewBinding(key.WithKeys(keyconst.KeyEsc, "ctrl+["))):
			m.active = false
			m.textInput.Blur()
			m.textInput.SetValue("")
			m.histIdx = -1
			return m, func() tea.Msg { return ClosedMsg{} }

		case key.Matches(keyMsg, key.NewBinding(key.WithKeys(keyconst.KeyEnter))):
			return m.executeCommand()

		case key.Matches(keyMsg, key.NewBinding(key.WithKeys(keyconst.KeyUp))):
			return m.historyUp(), nil

		case key.Matches(keyMsg, key.NewBinding(key.WithKeys(keyconst.KeyDown))):
			return m.historyDown(), nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// executeCommand parses and executes the current command.
func (m Model) executeCommand() (Model, tea.Cmd) {
	input := strings.TrimSpace(m.textInput.Value())
	if input == "" {
		m.active = false
		m.textInput.Blur()
		m.textInput.SetValue("")
		return m, func() tea.Msg { return ClosedMsg{} }
	}

	// Add to history
	m.history = append(m.history, input)
	if len(m.history) > 50 {
		m.history = m.history[1:]
	}

	// Parse command
	parts := strings.SplitN(input, " ", 2)
	cmdName := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	// Deactivate
	m.active = false
	m.textInput.Blur()
	m.textInput.SetValue("")
	m.histIdx = -1

	// Map command aliases
	switch cmdName {
	case "q", "quit", "exit":
		return m, func() tea.Msg { return CommandMsg{Command: CmdQuit} }

	case "d", "dev", "device":
		if args == "" {
			return m, func() tea.Msg { return ErrorMsg{Message: "device requires a name argument"} }
		}
		return m, func() tea.Msg { return CommandMsg{Command: CmdDevice, Args: args} }

	case "f", "filter":
		return m, func() tea.Msg { return CommandMsg{Command: CmdFilter, Args: args} }

	case "t", "theme":
		if args == "" {
			return m, func() tea.Msg { return ErrorMsg{Message: "theme requires a name argument"} }
		}
		return m, func() tea.Msg { return CommandMsg{Command: CmdTheme, Args: args} }

	case "v", "view":
		if args == "" {
			return m, func() tea.Msg {
				return ErrorMsg{Message: "view requires a name argument (devices, monitor, events, energy)"}
			}
		}
		return m, func() tea.Msg { return CommandMsg{Command: CmdView, Args: args} }

	case "h", "help":
		return m, func() tea.Msg { return CommandMsg{Command: CmdHelp, Args: args} }

	case "toggle", "x":
		return m, func() tea.Msg { return CommandMsg{Command: CmdToggle, Args: args} }

	default:
		return m, func() tea.Msg { return ErrorMsg{Message: "unknown command: " + cmdName} }
	}
}

// historyUp navigates to the previous command in history.
func (m Model) historyUp() Model {
	if len(m.history) == 0 {
		return m
	}

	if m.histIdx < 0 {
		m.histIdx = len(m.history) - 1
	} else if m.histIdx > 0 {
		m.histIdx--
	}

	m.textInput.SetValue(m.history[m.histIdx])
	m.textInput.CursorEnd()
	return m
}

// historyDown navigates to the next command in history.
func (m Model) historyDown() Model {
	if len(m.history) == 0 || m.histIdx < 0 {
		return m
	}

	if m.histIdx < len(m.history)-1 {
		m.histIdx++
		m.textInput.SetValue(m.history[m.histIdx])
	} else {
		m.histIdx = -1
		m.textInput.SetValue("")
	}

	m.textInput.CursorEnd()
	return m
}

// View renders the command mode component.
func (m Model) View() string {
	if !m.active {
		return ""
	}

	prompt := m.styles.Prompt.Render(":")
	input := m.textInput.View()

	return m.styles.Container.
		Width(m.width).
		Render(prompt + input)
}

// SetWidth sets the component width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	m.textInput.SetWidth(width - 10) // Account for prompt and padding
	return m
}

// Activate shows the command input and focuses it.
func (m Model) Activate() (Model, tea.Cmd) {
	m.active = true
	m.textInput.Focus()
	m.histIdx = -1
	return m, textinput.Blink
}

// Deactivate hides the command input.
func (m Model) Deactivate() Model {
	m.active = false
	m.textInput.Blur()
	m.textInput.SetValue("")
	m.histIdx = -1
	return m
}

// IsActive returns whether command mode is active.
func (m Model) IsActive() bool {
	return m.active
}

// AvailableCommands returns help text for available commands.
func AvailableCommands() string {
	return `Available commands:
  :q, :quit, :exit     Quit the dashboard
  :d, :device <name>   Jump to device by name
  :f, :filter [term]   Filter devices (clear if no term)
  :t, :theme <name>    Set color theme
  :v, :view <name>     Switch view (devices, monitor, events, energy)
  :x, :toggle          Toggle selected device
  :h, :help            Show help`
}
