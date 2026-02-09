package fleet

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/errorview"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// OperationsDeps holds the dependencies for the Operations component.
type OperationsDeps struct {
	Ctx context.Context
}

// Validate ensures all required dependencies are set.
func (d OperationsDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	return nil
}

// Op represents a fleet operation type.
type Op int

// Fleet operation constants.
const (
	OpAllOn Op = iota
	OpAllOff
	OpGroupOn
	OpGroupOff
)

// String returns the string representation of a fleet operation.
func (o Op) String() string {
	switch o {
	case OpAllOn:
		return "All Relays On"
	case OpAllOff:
		return "All Relays Off"
	case OpGroupOn:
		return "Group On"
	case OpGroupOff:
		return "Group Off"
	default:
		return "Unknown"
	}
}

// OperationResultMsg signals the result of a batch operation.
type OperationResultMsg struct {
	Results   []integrator.BatchResult
	Operation Op
	Err       error
}

// OperationsModel provides batch operations for the fleet.
type OperationsModel struct {
	ctx         context.Context
	fleet       *integrator.FleetManager
	operation   Op
	executing   bool
	lastResults []integrator.BatchResult
	lastErr     error
	width       int
	height      int
	focused     bool
	panelIndex  int
	styles      OperationsStyles
}

// OperationsStyles holds styles for the Operations component.
type OperationsStyles struct {
	Button       lipgloss.Style
	ButtonActive lipgloss.Style
	Success      lipgloss.Style
	Failure      lipgloss.Style
	Muted        lipgloss.Style
	Error        lipgloss.Style
	Title        lipgloss.Style
}

// DefaultOperationsStyles returns the default styles for the Operations component.
func DefaultOperationsStyles() OperationsStyles {
	colors := theme.GetSemanticColors()
	return OperationsStyles{
		Button: lipgloss.NewStyle().
			Padding(0, 2).
			Background(colors.TableBorder).
			Foreground(colors.Text),
		ButtonActive: lipgloss.NewStyle().
			Padding(0, 2).
			Background(colors.Highlight).
			Foreground(colors.Primary).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(colors.Online),
		Failure: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// NewOperations creates a new Operations model.
func NewOperations(deps OperationsDeps) OperationsModel {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("fleet/operations component init", err)
		panic(fmt.Sprintf("fleet/operations: invalid deps: %v", err))
	}

	return OperationsModel{
		ctx:       deps.Ctx,
		operation: OpAllOn,
		styles:    DefaultOperationsStyles(),
	}
}

// Init returns the initial command.
func (m OperationsModel) Init() tea.Cmd {
	return nil
}

// SetFleetManager sets the fleet manager.
func (m OperationsModel) SetFleetManager(fm *integrator.FleetManager) OperationsModel {
	m.fleet = fm
	return m
}

// SetSize sets the component dimensions.
func (m OperationsModel) SetSize(width, height int) OperationsModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m OperationsModel) SetFocused(focused bool) OperationsModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m OperationsModel) SetPanelIndex(index int) OperationsModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m OperationsModel) Update(msg tea.Msg) (OperationsModel, tea.Cmd) {
	return m.handleMessage(msg)
}

func (m OperationsModel) handleMessage(msg tea.Msg) (OperationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case OperationResultMsg:
		m.executing = false
		if msg.Err != nil {
			m.lastErr = msg.Err
			return m, nil
		}
		m.lastResults = msg.Results
		m.lastErr = nil
		return m, nil

	// Action messages from context system
	case messages.NavigationMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleNavigation(msg)
	case messages.ModeSelectMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleModeSelect(msg)
	case messages.ActivateRequestMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleExecute()
	case messages.RefreshRequestMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleRetry()
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m OperationsModel) handleNavigation(msg messages.NavigationMsg) (OperationsModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavLeft:
		if m.operation > OpAllOn {
			m.operation--
		}
	case messages.NavRight:
		if m.operation < OpAllOff {
			m.operation++
		}
	case messages.NavUp, messages.NavDown, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component - only horizontal selection
	}
	return m, nil
}

func (m OperationsModel) handleModeSelect(msg messages.ModeSelectMsg) (OperationsModel, tea.Cmd) {
	switch msg.Mode {
	case 1:
		m.operation = OpAllOn
	case 2:
		m.operation = OpAllOff
	}
	return m, nil
}

func (m OperationsModel) handleExecute() (OperationsModel, tea.Cmd) {
	if !m.executing && m.fleet != nil {
		m.executing = true
		m.lastErr = nil
		m.lastResults = nil
		return m, m.executeOperation()
	}
	return m, nil
}

func (m OperationsModel) handleRetry() (OperationsModel, tea.Cmd) {
	// Retry: clear error and re-execute
	if m.lastErr != nil && !m.executing && m.fleet != nil {
		m.executing = true
		m.lastErr = nil
		m.lastResults = nil
		return m, m.executeOperation()
	}
	return m, nil
}

func (m OperationsModel) handleKey(msg tea.KeyPressMsg) (OperationsModel, tea.Cmd) {
	// Component-specific keys not covered by action messages
	// (none currently - all keys migrated to action messages)
	_ = msg
	return m, nil
}

func (m OperationsModel) executeOperation() tea.Cmd {
	return func() tea.Msg {
		if m.fleet == nil {
			return OperationResultMsg{Err: fmt.Errorf("not connected to fleet")}
		}

		ctx := m.ctx
		var results []integrator.BatchResult

		switch m.operation {
		case OpAllOn:
			results = m.fleet.AllRelaysOn(ctx)
		case OpAllOff:
			results = m.fleet.AllRelaysOff(ctx)
		default:
			return OperationResultMsg{Err: fmt.Errorf("unsupported operation")}
		}

		return OperationResultMsg{Results: results, Operation: m.operation}
	}
}

// View renders the Operations component.
func (m OperationsModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Batch Operations").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter(theme.StyledKeybindings(keys.FormatHints([]keys.Hint{{Key: "1-2", Desc: "op"}, {Key: "h/l", Desc: "sel"}, {Key: "enter", Desc: "exec"}}, keys.FooterHintWidth(m.width))))
	}

	if m.fleet == nil {
		r.SetContent(m.styles.Muted.Render("Not connected to Shelly Cloud"))
		return r.Render()
	}

	var content strings.Builder

	// Operation buttons
	content.WriteString("Select operation:\n\n")

	ops := []struct {
		op    Op
		key   string
		label string
	}{
		{OpAllOn, "1", "All On"},
		{OpAllOff, "2", "All Off"},
	}

	for i, op := range ops {
		if i > 0 {
			content.WriteString("  ")
		}

		style := m.styles.Button
		if op.op == m.operation && m.focused {
			style = m.styles.ButtonActive
		}

		content.WriteString(m.styles.Muted.Render(op.key + ":"))
		content.WriteString(style.Render(op.label))
	}

	content.WriteString("\n\n")

	// Status
	switch {
	case m.executing:
		content.WriteString(m.styles.Muted.Render("Executing operation..."))
	case m.lastErr != nil:
		content.WriteString(errorview.RenderInline(m.lastErr))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  Press 'r' to retry"))
	case len(m.lastResults) > 0:
		success := 0
		failed := 0
		for _, r := range m.lastResults {
			if r.Success {
				success++
			} else {
				failed++
			}
		}

		resultLine := fmt.Sprintf("Result: %d succeeded", success)
		if failed > 0 {
			resultLine += fmt.Sprintf(", %d failed", failed)
			content.WriteString(m.styles.Failure.Render(resultLine))
		} else {
			content.WriteString(m.styles.Success.Render(resultLine))
		}
	}

	r.SetContent(content.String())
	return r.Render()
}

// Operation returns the selected operation.
func (m OperationsModel) Operation() Op {
	return m.operation
}

// Executing returns whether an operation is in progress.
func (m OperationsModel) Executing() bool {
	return m.executing
}

// LastResults returns the results of the last operation.
func (m OperationsModel) LastResults() []integrator.BatchResult {
	return m.lastResults
}

// LastError returns the last error that occurred.
func (m OperationsModel) LastError() error {
	return m.lastErr
}
