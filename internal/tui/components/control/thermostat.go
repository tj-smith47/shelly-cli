package control

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
)

// ThermostatState holds the current state of a thermostat.
type ThermostatState struct {
	ID              int
	Name            string
	Enabled         bool
	Mode            string  // "heat", "cool", "auto", "off"
	TargetC         float64 // Target temperature in Celsius
	CurrentC        float64 // Current temperature in Celsius
	CurrentHumidity float64 // Current humidity percentage
	ValvePosition   int     // Valve position 0-100%
	BoostActive     bool
	BoostRemaining  int // Seconds remaining in boost
	OverrideActive  bool
	Source          string
}

// thermostatFocus tracks which element is focused.
type thermostatFocus int

const (
	thermostatFocusTarget thermostatFocus = iota
	thermostatFocusMode
	thermostatFocusActions
)

// ThermostatModel is the control panel for thermostat components.
type ThermostatModel struct {
	ctx       context.Context
	svc       Service
	device    string
	state     ThermostatState
	target    form.Slider
	modeIdx   int
	actionIdx int
	styles    Styles
	focused   bool
	focus     thermostatFocus
	width     int
	height    int
	loading   bool
	errorMsg  string
}

// NewThermostat creates a new thermostat control panel.
func NewThermostat(ctx context.Context, svc Service, device string, state ThermostatState) ThermostatModel {
	targetC := state.TargetC
	if targetC < 5 {
		targetC = 5
	}
	if targetC > 35 {
		targetC = 35
	}

	// Find mode index
	modeIdx := 0
	for i, m := range ThermostatModes {
		if m.ID == state.Mode {
			modeIdx = i
			break
		}
	}

	return ThermostatModel{
		ctx:    ctx,
		svc:    svc,
		device: device,
		state:  state,
		target: form.NewSlider(
			form.WithSliderLabel("Target"),
			form.WithSliderMin(5),
			form.WithSliderMax(35),
			form.WithSliderStep(0.5),
			form.WithSliderValue(targetC),
			form.WithSliderWidth(20),
			form.WithSliderFormat("%.1f°C"),
		).Focus(),
		modeIdx: modeIdx,
		styles:  DefaultStyles(),
		focused: true,
		focus:   thermostatFocusTarget,
	}
}

// Init initializes the thermostat control.
func (m ThermostatModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the thermostat control.
func (m ThermostatModel) Update(msg tea.Msg) (ThermostatModel, tea.Cmd) {
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

func (m ThermostatModel) handleActionResult(msg ActionMsg) ThermostatModel {
	switch msg.Action {
	case actionTarget:
		if val, ok := msg.Value.(float64); ok {
			m.state.TargetC = val
		}
	case actionMode:
		if val, ok := msg.Value.(string); ok {
			m.state.Mode = val
		}
	case actionBoost:
		m.state.BoostActive = true
	case actionCancelBst:
		m.state.BoostActive = false
	}
	return m
}

func (m ThermostatModel) handleKeyPress(msg tea.KeyPressMsg) (ThermostatModel, tea.Cmd) {
	// Global shortcuts
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("+", "="))):
		return m.adjustTarget(0.5)
	case key.Matches(msg, key.NewBinding(key.WithKeys("-", "_"))):
		return m.adjustTarget(-0.5)
	case key.Matches(msg, key.NewBinding(key.WithKeys("m"))):
		// Cycle to next mode
		return m.cycleMode()
	case key.Matches(msg, key.NewBinding(key.WithKeys("b"))):
		return m.executeBoost()
	case key.Matches(msg, key.NewBinding(key.WithKeys("B"))):
		return m.executeCancelBoost()
	case key.Matches(msg, key.NewBinding(key.WithKeys(keyconst.KeyTab))):
		return m.cycleFocus(), nil
	case key.Matches(msg, key.NewBinding(key.WithKeys(keyconst.KeyEnter))):
		return m.handleEnter()
	}

	// Focus-specific handling
	switch m.focus {
	case thermostatFocusTarget:
		var cmd tea.Cmd
		m.target, cmd = m.target.Update(msg)
		newVal := m.target.Value()
		if newVal != m.state.TargetC {
			return m.executeSetTarget(newVal)
		}
		return m, cmd
	case thermostatFocusMode:
		return m.handleModeNavigation(msg)
	case thermostatFocusActions:
		return m.handleActionNavigation(msg)
	}

	return m, nil
}

func (m ThermostatModel) handleModeNavigation(msg tea.KeyPressMsg) (ThermostatModel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
		if m.modeIdx > 0 {
			m.modeIdx--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
		if m.modeIdx < len(ThermostatModes)-1 {
			m.modeIdx++
		}
	}
	return m, nil
}

func (m ThermostatModel) handleActionNavigation(msg tea.KeyPressMsg) (ThermostatModel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
		if m.actionIdx > 0 {
			m.actionIdx--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
		if m.actionIdx < 1 { // Only 2 actions: Boost, Cancel Boost
			m.actionIdx++
		}
	}
	return m, nil
}

func (m ThermostatModel) handleEnter() (ThermostatModel, tea.Cmd) {
	switch m.focus {
	case thermostatFocusTarget:
		// Enter on target slider has no special action
		return m, nil
	case thermostatFocusMode:
		return m.executeSetMode(ThermostatModes[m.modeIdx].ID)
	case thermostatFocusActions:
		if m.actionIdx == 0 {
			return m.executeBoost()
		}
		return m.executeCancelBoost()
	}
	return m, nil
}

func (m ThermostatModel) cycleFocus() ThermostatModel {
	m.target = m.target.Blur()
	m.focus = (m.focus + 1) % 3
	if m.focus == thermostatFocusTarget {
		m.target = m.target.Focus()
	}
	return m
}

func (m ThermostatModel) cycleMode() (ThermostatModel, tea.Cmd) {
	m.modeIdx = (m.modeIdx + 1) % len(ThermostatModes)
	return m.executeSetMode(ThermostatModes[m.modeIdx].ID)
}

func (m ThermostatModel) adjustTarget(delta float64) (ThermostatModel, tea.Cmd) {
	newVal := m.state.TargetC + delta
	if newVal < 5 {
		newVal = 5
	}
	if newVal > 35 {
		newVal = 35
	}
	if newVal == m.state.TargetC {
		return m, nil
	}
	return m.executeSetTarget(newVal)
}

func (m ThermostatModel) executeSetTarget(targetC float64) (ThermostatModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.ThermostatSetTarget(m.ctx, m.device, m.state.ID, targetC)
		return ActionMsg{
			Device:    m.device,
			Component: TypeThermostat,
			ID:        m.state.ID,
			Action:    actionTarget,
			Value:     targetC,
			Err:       err,
		}
	}
}

func (m ThermostatModel) executeSetMode(mode string) (ThermostatModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.ThermostatSetMode(m.ctx, m.device, m.state.ID, mode)
		return ActionMsg{
			Device:    m.device,
			Component: TypeThermostat,
			ID:        m.state.ID,
			Action:    actionMode,
			Value:     mode,
			Err:       err,
		}
	}
}

func (m ThermostatModel) executeBoost() (ThermostatModel, tea.Cmd) {
	if m.state.BoostActive {
		return m, nil // Already boosting
	}
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.ThermostatBoost(m.ctx, m.device, m.state.ID, 0) // 0 = default duration
		return ActionMsg{
			Device:    m.device,
			Component: TypeThermostat,
			ID:        m.state.ID,
			Action:    actionBoost,
			Err:       err,
		}
	}
}

func (m ThermostatModel) executeCancelBoost() (ThermostatModel, tea.Cmd) {
	if !m.state.BoostActive {
		return m, nil // Not boosting
	}
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.ThermostatCancelBoost(m.ctx, m.device, m.state.ID)
		return ActionMsg{
			Device:    m.device,
			Component: TypeThermostat,
			ID:        m.state.ID,
			Action:    actionCancelBst,
			Err:       err,
		}
	}
}

// View renders the thermostat control panel.
func (m ThermostatModel) View() string {
	var b strings.Builder

	// Title
	name := m.state.Name
	if name == "" {
		name = fmt.Sprintf("Thermostat %d", m.state.ID)
	}
	b.WriteString(m.styles.Title.Render(name))
	b.WriteString("\n\n")

	// Current temperature display
	b.WriteString(m.renderCurrentTemperature())
	b.WriteString("\n\n")

	// Target temperature slider
	b.WriteString(m.target.View())
	b.WriteString("\n\n")

	// Mode selector
	b.WriteString(m.renderModeSelector())
	b.WriteString("\n")

	// Humidity (if available)
	if m.state.CurrentHumidity > 0 {
		b.WriteString(m.styles.Label.Render("Humidity:"))
		b.WriteString(fmt.Sprintf("%.0f%%", m.state.CurrentHumidity))
		b.WriteString("\n")
	}

	// Valve position (if available)
	if m.state.ValvePosition > 0 {
		b.WriteString(m.renderValvePosition())
		b.WriteString("\n")
	}

	// Boost status
	if m.state.BoostActive {
		b.WriteString(m.styles.OnState.Render("🔥 BOOST ACTIVE"))
		if m.state.BoostRemaining > 0 {
			b.WriteString(fmt.Sprintf(" (%ds remaining)", m.state.BoostRemaining))
		}
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
	help := "+/-:temp  m:mode  b:boost  B:cancel boost  tab:focus"
	b.WriteString(m.styles.Help.Render(help))

	return m.styles.Container.Render(b.String())
}

func (m ThermostatModel) renderCurrentTemperature() string {
	currentStr := fmt.Sprintf("%.1f°C", m.state.CurrentC)
	targetStr := fmt.Sprintf("→ %.1f°C", m.state.TargetC)

	// Color code based on whether heating is needed
	var currentStyle lipgloss.Style
	if m.state.CurrentC < m.state.TargetC {
		currentStyle = m.styles.Value // Below target
	} else {
		currentStyle = m.styles.OnState // At or above target
	}

	return m.styles.Label.Render("Current:") +
		currentStyle.Render(currentStr) + "  " +
		m.styles.Muted.Render(targetStr)
}

func (m ThermostatModel) renderModeSelector() string {
	modes := make([]string, 0, len(ThermostatModes))
	for i, mode := range ThermostatModes {
		style := m.styles.Action
		if i == m.modeIdx {
			if m.focus == thermostatFocusMode {
				style = m.styles.ActionFocus
			} else {
				style = m.styles.Selected
			}
		}
		modes = append(modes, style.Render(mode.Label))
	}
	return m.styles.Label.Render("Mode:") + strings.Join(modes, " ")
}

func (m ThermostatModel) renderValvePosition() string {
	pos := m.state.ValvePosition
	barWidth := 15
	filledWidth := (pos * barWidth) / 100
	emptyWidth := barWidth - filledWidth

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", emptyWidth)

	bar := m.styles.OnState.Render(filled) + m.styles.Muted.Render(empty)
	return m.styles.Label.Render("Valve:") + "[" + bar + "] " + fmt.Sprintf("%d%%", pos)
}

func (m ThermostatModel) renderActions() string {
	boostStyle := m.styles.Action
	cancelStyle := m.styles.Action

	if m.focus == thermostatFocusActions {
		if m.actionIdx == 0 {
			boostStyle = m.styles.ActionFocus
		} else {
			cancelStyle = m.styles.ActionFocus
		}
	}

	boostLabel := "Boost"
	if m.state.BoostActive {
		boostLabel = "🔥 Boost"
	}

	return lipgloss.JoinHorizontal(lipgloss.Center,
		boostStyle.Render(boostLabel), "  ",
		cancelStyle.Render("Cancel Boost"),
	)
}

// SetState updates the thermostat state.
func (m ThermostatModel) SetState(state ThermostatState) ThermostatModel {
	m.state = state
	m.target = m.target.SetValue(state.TargetC)
	for i, mode := range ThermostatModes {
		if mode.ID == state.Mode {
			m.modeIdx = i
			break
		}
	}
	return m
}

// SetSize sets the panel dimensions.
func (m ThermostatModel) SetSize(width, height int) ThermostatModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m ThermostatModel) SetFocused(focused bool) ThermostatModel {
	m.focused = focused
	if !focused {
		m.target = m.target.Blur()
	} else if m.focus == thermostatFocusTarget {
		m.target = m.target.Focus()
	}
	return m
}

// Focused returns whether the panel is focused.
func (m ThermostatModel) Focused() bool {
	return m.focused
}

// State returns the current thermostat state.
func (m ThermostatModel) State() ThermostatState {
	return m.state
}
