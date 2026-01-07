package control

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
)

// CoverState holds the current state of a cover/roller.
type CoverState struct {
	ID          int
	Name        string
	State       string // "open", "closed", "opening", "closing", "stopped", "calibrating"
	Position    int    // 0-100 (0=closed, 100=open)
	TargetPos   int    // Target position during movement
	Power       float64
	Calibrating bool
	Source      string
}

// coverFocus tracks which element is focused.
type coverFocus int

const (
	coverFocusActions coverFocus = iota
	coverFocusPosition
)

// coverAction represents the currently highlighted action button.
type coverAction int

const (
	coverActionOpen coverAction = iota
	coverActionStop
	coverActionClose
	coverActionCalibrate
)

// CoverModel is the control panel for cover/roller components.
type CoverModel struct {
	ctx       context.Context
	svc       Service
	device    string
	state     CoverState
	position  form.Slider
	styles    Styles
	focused   bool
	focus     coverFocus
	actionIdx coverAction
	width     int
	height    int
	loading   bool
	errorMsg  string
}

// NewCover creates a new cover control panel.
func NewCover(ctx context.Context, svc Service, device string, state CoverState) CoverModel {
	pos := state.Position
	if pos < 0 {
		pos = 0
	}
	if pos > 100 {
		pos = 100
	}

	return CoverModel{
		ctx:    ctx,
		svc:    svc,
		device: device,
		state:  state,
		position: form.NewSlider(
			form.WithSliderLabel("Position"),
			form.WithSliderMin(0),
			form.WithSliderMax(100),
			form.WithSliderStep(5),
			form.WithSliderValue(float64(pos)),
			form.WithSliderWidth(20),
			form.WithSliderFormat("%.0f%%"),
			form.WithSliderHelp("0%=closed, 100%=open"),
		),
		styles:    DefaultStyles(),
		focused:   true,
		focus:     coverFocusActions,
		actionIdx: coverActionStop,
	}
}

// Init initializes the cover control.
func (m CoverModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the cover control.
func (m CoverModel) Update(msg tea.Msg) (CoverModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ActionMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		} else {
			m.errorMsg = ""
			m = m.handleActionResult(msg)
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused || m.loading {
			return m, nil
		}
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m CoverModel) handleActionResult(msg ActionMsg) CoverModel {
	switch msg.Action {
	case actionOpen:
		m.state.State = "opening"
	case actionClose:
		m.state.State = "closing"
	case actionStop:
		m.state.State = "stopped"
	case actionPosition:
		if val, ok := msg.Value.(int); ok {
			m.state.Position = val
			m.state.TargetPos = val
		}
	case actionCalibrate:
		m.state.Calibrating = true
		m.state.State = "calibrating"
	}
	return m
}

func (m CoverModel) handleKeyPress(msg tea.KeyPressMsg) (CoverModel, tea.Cmd) {
	// Global shortcuts
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
		return m.executeOpen()
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		return m.executeClose()
	case key.Matches(msg, key.NewBinding(key.WithKeys("s", " "))):
		return m.executeStop()
	case key.Matches(msg, key.NewBinding(key.WithKeys("p"))):
		// Switch to position focus
		m.focus = coverFocusPosition
		m.position = m.position.Focus()
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("C"))):
		return m.executeCalibrate()
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
		return m.cycleFocus(), nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if m.focus == coverFocusActions {
			return m.executeSelectedAction()
		}
	}

	// Focus-specific handling
	if m.focus == coverFocusActions {
		return m.handleActionNavigation(msg)
	}

	if m.focus == coverFocusPosition {
		var cmd tea.Cmd
		m.position, cmd = m.position.Update(msg)
		newVal := int(m.position.Value())
		if newVal != m.state.Position {
			return m.executePosition(newVal)
		}
		return m, cmd
	}

	return m, nil
}

func (m CoverModel) handleActionNavigation(msg tea.KeyPressMsg) (CoverModel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
		if m.actionIdx > 0 {
			m.actionIdx--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
		if m.actionIdx < coverActionCalibrate {
			m.actionIdx++
		}
	}
	return m, nil
}

func (m CoverModel) cycleFocus() CoverModel {
	if m.focus == coverFocusActions {
		m.focus = coverFocusPosition
		m.position = m.position.Focus()
	} else {
		m.focus = coverFocusActions
		m.position = m.position.Blur()
	}
	return m
}

func (m CoverModel) executeSelectedAction() (CoverModel, tea.Cmd) {
	switch m.actionIdx {
	case coverActionOpen:
		return m.executeOpen()
	case coverActionStop:
		return m.executeStop()
	case coverActionClose:
		return m.executeClose()
	case coverActionCalibrate:
		return m.executeCalibrate()
	}
	return m, nil
}

func (m CoverModel) executeOpen() (CoverModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeCover, m.state.ID, actionOpen, func() error {
		return m.svc.CoverOpen(m.ctx, m.device, m.state.ID, nil)
	})
}

func (m CoverModel) executeClose() (CoverModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeCover, m.state.ID, actionClose, func() error {
		return m.svc.CoverClose(m.ctx, m.device, m.state.ID, nil)
	})
}

func (m CoverModel) executeStop() (CoverModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeCover, m.state.ID, actionStop, func() error {
		return m.svc.CoverStop(m.ctx, m.device, m.state.ID)
	})
}

func (m CoverModel) executePosition(position int) (CoverModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.CoverPosition(m.ctx, m.device, m.state.ID, position)
		return ActionMsg{
			Device:    m.device,
			Component: TypeCover,
			ID:        m.state.ID,
			Action:    actionPosition,
			Value:     position,
			Err:       err,
		}
	}
}

func (m CoverModel) executeCalibrate() (CoverModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeCover, m.state.ID, actionCalibrate, func() error {
		return m.svc.CoverCalibrate(m.ctx, m.device, m.state.ID)
	})
}

// View renders the cover control panel.
func (m CoverModel) View() string {
	var b strings.Builder

	// Title
	name := m.state.Name
	if name == "" {
		name = fmt.Sprintf("Cover %d", m.state.ID)
	}
	b.WriteString(m.styles.Title.Render(name))
	b.WriteString("\n\n")

	// State indicator
	b.WriteString(m.renderState())
	b.WriteString("\n\n")

	// Position bar visualization
	b.WriteString(m.renderPositionBar())
	b.WriteString("\n\n")

	// Position slider
	b.WriteString(m.position.View())
	b.WriteString("\n")

	// Power reading
	if m.state.Power != 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Label.Render("Power:"))
		b.WriteString(m.styles.Power.Render(formatPower(m.state.Power)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Action buttons
	b.WriteString(m.renderActions())
	b.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		b.WriteString(m.styles.Error.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	// Help
	help := "o:open  c:close  s/space:stop  p:position  C:calibrate  tab:focus"
	b.WriteString(m.styles.Help.Render(help))

	return m.styles.Container.Render(b.String())
}

func (m CoverModel) renderState() string {
	var stateStr string
	var stateStyle lipgloss.Style

	switch m.state.State {
	case "open":
		stateStr = "▲ OPEN"
		stateStyle = m.styles.OnState
	case "closed":
		stateStr = "▼ CLOSED"
		stateStyle = m.styles.OffState
	case "opening":
		stateStr = "⋯ OPENING"
		stateStyle = m.styles.Value
	case "closing":
		stateStr = "⋯ CLOSING"
		stateStyle = m.styles.Value
	case "stopped":
		stateStr = "◼ STOPPED"
		stateStyle = m.styles.Muted
	case "calibrating":
		stateStr = "⟳ CALIBRATING"
		stateStyle = m.styles.Value
	default:
		stateStr = m.state.State
		stateStyle = m.styles.Muted
	}

	if m.loading {
		stateStr = "⋯ Loading..."
		stateStyle = m.styles.Muted
	}

	return m.styles.Label.Render("State:") + stateStyle.Render(stateStr)
}

func (m CoverModel) renderPositionBar() string {
	pos := m.state.Position
	if pos < 0 {
		pos = 0
	}
	if pos > 100 {
		pos = 100
	}

	barWidth := 20
	filledWidth := (pos * barWidth) / 100
	emptyWidth := barWidth - filledWidth

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", emptyWidth)

	bar := m.styles.OnState.Render(filled) + m.styles.Muted.Render(empty)

	return m.styles.Label.Render("Position:") + "[" + bar + "] " + fmt.Sprintf("%d%%", pos)
}

func (m CoverModel) renderActions() string {
	actions := []struct {
		label string
		idx   coverAction
	}{
		{"Open", coverActionOpen},
		{"Stop", coverActionStop},
		{"Close", coverActionClose},
		{"Calibrate", coverActionCalibrate},
	}

	buttons := make([]string, 0, len(actions))
	for _, a := range actions {
		style := m.styles.Action
		if m.focus == coverFocusActions && m.actionIdx == a.idx {
			style = m.styles.ActionFocus
		}
		buttons = append(buttons, style.Render(a.label))
	}

	return strings.Join(buttons, "  ")
}

// SetState updates the cover state.
func (m CoverModel) SetState(state CoverState) CoverModel {
	m.state = state
	m.position = m.position.SetValue(float64(clamp(state.Position, 0, 100)))
	return m
}

// SetSize sets the panel dimensions.
func (m CoverModel) SetSize(width, height int) CoverModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m CoverModel) SetFocused(focused bool) CoverModel {
	m.focused = focused
	if !focused {
		m.position = m.position.Blur()
	} else if m.focus == coverFocusPosition {
		m.position = m.position.Focus()
	}
	return m
}

// Focused returns whether the panel is focused.
func (m CoverModel) Focused() bool {
	return m.focused
}

// State returns the current cover state.
func (m CoverModel) State() CoverState {
	return m.state
}
