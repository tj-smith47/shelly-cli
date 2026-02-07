// Package scripts provides TUI components for managing device scripts.
package scripts

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	shellyevents "github.com/tj-smith47/shelly-go/events"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// maxConsoleLines is the maximum number of lines to keep in the console buffer.
const maxConsoleLines = 500

// ConsoleOutputMsg delivers script output to the console component.
type ConsoleOutputMsg struct {
	Device   string
	ScriptID int
	Output   string
	Time     time.Time
}

// consoleLine represents a single line of console output.
type consoleLine struct {
	Time     time.Time
	ScriptID int
	Output   string
}

// consoleState holds the shared mutable state for the console (accessed via pointer).
type consoleState struct {
	mu    sync.Mutex
	lines []consoleLine
}

// ConsoleModel displays script console output.
type ConsoleModel struct {
	eventStream *automation.EventStream
	device      string
	scriptID    int // 0 = show all scripts, >0 = filter by script
	state       *consoleState
	scrollPos   int
	width       int
	height      int
	focused     bool
	paused      bool
	styles      consoleStyles

	// Thread-safe message channel
	msgChan chan ConsoleOutputMsg
}

type consoleStyles struct {
	Timestamp lipgloss.Style
	ScriptID  lipgloss.Style
	Output    lipgloss.Style
	Paused    lipgloss.Style
}

func defaultConsoleStyles() consoleStyles {
	colors := theme.GetSemanticColors()
	return consoleStyles{
		Timestamp: lipgloss.NewStyle().Foreground(colors.Muted),
		ScriptID:  lipgloss.NewStyle().Foreground(colors.Highlight).Bold(true),
		Output:    lipgloss.NewStyle().Foreground(colors.Text),
		Paused:    lipgloss.NewStyle().Foreground(colors.Warning).Bold(true),
	}
}

// NewConsoleModel creates a new script console model.
func NewConsoleModel(es *automation.EventStream) ConsoleModel {
	state := &consoleState{
		lines: make([]consoleLine, 0),
	}
	msgChan := make(chan ConsoleOutputMsg, 100)

	m := ConsoleModel{
		eventStream: es,
		state:       state,
		msgChan:     msgChan,
		styles:      defaultConsoleStyles(),
	}

	// Subscribe to script events if event stream is provided
	if es != nil {
		es.SubscribeFiltered(
			shellyevents.WithEventType(shellyevents.EventTypeScript),
			func(evt shellyevents.Event) {
				scriptEvt, ok := evt.(*shellyevents.ScriptEvent)
				if !ok {
					return
				}

				// Send to channel (non-blocking)
				select {
				case msgChan <- ConsoleOutputMsg{
					Device:   scriptEvt.DeviceID(),
					ScriptID: scriptEvt.ScriptID,
					Output:   scriptEvt.Output,
					Time:     scriptEvt.Timestamp(),
				}:
				default:
					// Channel full, drop message
				}
			},
		)
	}

	return m
}

// SetDevice sets the device to filter console output by.
func (m ConsoleModel) SetDevice(device string) ConsoleModel {
	m.device = device
	return m
}

// SetScriptID sets the script ID to filter console output by.
// Set to 0 to show output from all scripts.
func (m ConsoleModel) SetScriptID(id int) ConsoleModel {
	m.scriptID = id
	return m
}

// SetSize sets the console dimensions.
func (m ConsoleModel) SetSize(width, height int) ConsoleModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focused state.
func (m ConsoleModel) SetFocused(focused bool) ConsoleModel {
	m.focused = focused
	return m
}

// Clear clears all console output.
func (m ConsoleModel) Clear() ConsoleModel {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	m.state.lines = make([]consoleLine, 0)
	m.scrollPos = 0
	return m
}

// TogglePause toggles the paused state.
func (m ConsoleModel) TogglePause() ConsoleModel {
	m.paused = !m.paused
	return m
}

// IsPaused returns whether the console is paused.
func (m ConsoleModel) IsPaused() bool {
	return m.paused
}

// Init returns the initial command.
func (m ConsoleModel) Init() tea.Cmd {
	return m.pollMessages()
}

// pollMessages polls for new console output messages.
func (m ConsoleModel) pollMessages() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-m.msgChan:
			return msg
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}
}

// Update handles messages.
func (m ConsoleModel) Update(msg tea.Msg) (ConsoleModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ConsoleOutputMsg:
		m = m.addLine(msg)
		return m, m.pollMessages()

	// Action messages from context system
	case messages.NavigationMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleNavigation(msg), nil
	case messages.ClearRequestMsg:
		if !m.focused {
			return m, nil
		}
		m = m.Clear()
		return m, nil
	case messages.PauseRequestMsg:
		if !m.focused {
			return m, nil
		}
		m = m.TogglePause()
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	// Continue polling
	return m, m.pollMessages()
}

func (m ConsoleModel) handleNavigation(msg messages.NavigationMsg) ConsoleModel {
	m.state.mu.Lock()
	lineCount := len(m.state.lines)
	m.state.mu.Unlock()

	switch msg.Direction {
	case messages.NavUp:
		if m.scrollPos > 0 {
			m.scrollPos--
		}
	case messages.NavDown:
		maxScroll := max(0, lineCount-m.visibleLines())
		if m.scrollPos < maxScroll {
			m.scrollPos++
		}
	case messages.NavPageUp:
		m.scrollPos = max(0, m.scrollPos-m.visibleLines())
	case messages.NavPageDown:
		maxScroll := max(0, lineCount-m.visibleLines())
		m.scrollPos = min(maxScroll, m.scrollPos+m.visibleLines())
	case messages.NavHome:
		m.scrollPos = 0
	case messages.NavEnd:
		m.scrollPos = max(0, lineCount-m.visibleLines())
	case messages.NavLeft, messages.NavRight:
		// Not applicable for console output
	}
	return m
}

func (m ConsoleModel) handleKey(msg tea.KeyPressMsg) (ConsoleModel, tea.Cmd) {
	// Component-specific keys not covered by action messages
	// (none currently - all keys migrated to action messages)
	_ = msg
	return m, nil
}

func (m ConsoleModel) addLine(msg ConsoleOutputMsg) ConsoleModel {
	// Filter by device
	if m.device != "" && msg.Device != m.device {
		return m
	}

	// Filter by script ID if set
	if m.scriptID > 0 && msg.ScriptID != m.scriptID {
		return m
	}

	// Don't add lines when paused
	if m.paused {
		return m
	}

	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	m.state.lines = append(m.state.lines, consoleLine{
		Time:     msg.Time,
		ScriptID: msg.ScriptID,
		Output:   msg.Output,
	})

	// Trim to max lines
	if len(m.state.lines) > maxConsoleLines {
		m.state.lines = m.state.lines[len(m.state.lines)-maxConsoleLines:]
	}

	// Auto-scroll to bottom if not manually scrolled
	maxScroll := max(0, len(m.state.lines)-m.visibleLines())
	if m.scrollPos >= maxScroll-1 {
		m.scrollPos = maxScroll
	}

	return m
}

func (m ConsoleModel) visibleLines() int {
	// Account for header
	return max(1, m.height-4)
}

// View renders the console output.
func (m ConsoleModel) View() string {
	title := "Console"
	if m.scriptID > 0 {
		title = fmt.Sprintf("Console (Script %d)", m.scriptID)
	}

	footer := theme.StyledKeybindings(keys.FormatHints([]keys.Hint{
		{Key: "↑/↓", Desc: "scroll"},
		{Key: "c", Desc: "clear"},
		{Key: "p", Desc: "pause"},
	}, keys.FooterHintWidth(m.width)))
	if m.paused {
		footer = m.styles.Paused.Render("PAUSED") + " | " + footer
	}

	r := rendering.New(m.width, m.height).
		SetTitle(title).
		SetFocused(m.focused).
		SetFooter(footer).
		SetContent(m.renderLines())

	return r.Render()
}

func (m ConsoleModel) renderLines() string {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	if len(m.state.lines) == 0 {
		return m.styles.Output.Foreground(theme.GetSemanticColors().Muted).Render("No output yet...")
	}

	var content strings.Builder
	visible := m.visibleLines()
	start := m.scrollPos
	end := min(start+visible, len(m.state.lines))

	for i := start; i < end; i++ {
		line := m.state.lines[i]
		timestamp := m.styles.Timestamp.Render(line.Time.Format("15:04:05"))
		scriptID := m.styles.ScriptID.Render(fmt.Sprintf("[%d]", line.ScriptID))
		output := m.styles.Output.Render(line.Output)

		content.WriteString(fmt.Sprintf("%s %s %s", timestamp, scriptID, output))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}
