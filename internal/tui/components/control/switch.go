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

// SwitchState holds the current state of a switch.
type SwitchState struct {
	ID      int
	Name    string
	Output  bool
	Power   float64
	Voltage float64
	Current float64
	Source  string
}

// SwitchModel is the control panel for switch components.
type SwitchModel struct {
	ctx      context.Context
	svc      Service
	device   string
	state    SwitchState
	toggle   form.Toggle
	styles   Styles
	focused  bool
	width    int
	height   int
	loading  bool
	errorMsg string
}

// NewSwitch creates a new switch control panel.
func NewSwitch(ctx context.Context, svc Service, device string, state SwitchState) SwitchModel {
	return SwitchModel{
		ctx:    ctx,
		svc:    svc,
		device: device,
		state:  state,
		toggle: form.NewToggle(
			form.WithToggleLabel("Power"),
			form.WithToggleOnLabel("ON"),
			form.WithToggleOffLabel("OFF"),
			form.WithToggleValue(state.Output),
		).Focus(),
		styles:  DefaultStyles(),
		focused: true,
	}
}

// Init initializes the switch control.
func (m SwitchModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the switch control.
func (m SwitchModel) Update(msg tea.Msg) (SwitchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ActionMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		} else {
			m.errorMsg = ""
			// Update state based on action
			switch msg.Action {
			case actionToggle:
				m.state.Output = !m.state.Output
			case actionOn:
				m.state.Output = true
			case actionOff:
				m.state.Output = false
			}
			m.toggle = m.toggle.SetValue(m.state.Output)
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

func (m SwitchModel) handleKeyPress(msg tea.KeyPressMsg) (SwitchModel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("t", " "))):
		return m.executeToggle()
	case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
		return m.executeOn()
	case key.Matches(msg, key.NewBinding(key.WithKeys("O"))):
		return m.executeOff()
	}
	return m, nil
}

func (m SwitchModel) executeToggle() (SwitchModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeSwitch, m.state.ID, actionToggle, func() error {
		return m.svc.SwitchToggle(m.ctx, m.device, m.state.ID)
	})
}

func (m SwitchModel) executeOn() (SwitchModel, tea.Cmd) {
	if m.state.Output {
		return m, nil // Already on
	}
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeSwitch, m.state.ID, actionOn, func() error {
		return m.svc.SwitchOn(m.ctx, m.device, m.state.ID)
	})
}

func (m SwitchModel) executeOff() (SwitchModel, tea.Cmd) {
	if !m.state.Output {
		return m, nil // Already off
	}
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeSwitch, m.state.ID, actionOff, func() error {
		return m.svc.SwitchOff(m.ctx, m.device, m.state.ID)
	})
}

// View renders the switch control panel.
func (m SwitchModel) View() string {
	var b strings.Builder

	// Title
	name := m.state.Name
	if name == "" {
		name = fmt.Sprintf("Switch %d", m.state.ID)
	}
	b.WriteString(m.styles.Title.Render(name))
	b.WriteString("\n\n")

	// State indicator
	stateStr := m.styles.OffState.Render("○ OFF")
	if m.state.Output {
		stateStr = m.styles.OnState.Render("● ON")
	}
	if m.loading {
		stateStr = m.styles.Muted.Render("⋯ Loading...")
	}
	b.WriteString(m.styles.Label.Render("State:"))
	b.WriteString(stateStr)
	b.WriteString("\n")

	// Power reading (if available)
	if m.state.Power != 0 || m.state.Voltage != 0 {
		b.WriteString(m.styles.Label.Render("Power:"))
		var powerParts []string
		if m.state.Power != 0 {
			powerParts = append(powerParts, m.styles.Power.Render(formatPower(m.state.Power)))
		}
		if m.state.Voltage != 0 {
			powerParts = append(powerParts, fmt.Sprintf("%.0fV", m.state.Voltage))
		}
		if m.state.Current != 0 {
			powerParts = append(powerParts, fmt.Sprintf("%.2fA", m.state.Current))
		}
		b.WriteString(strings.Join(powerParts, " │ "))
		b.WriteString("\n")
	}

	// Source
	if m.state.Source != "" {
		b.WriteString(m.styles.Label.Render("Source:"))
		b.WriteString(m.styles.Muted.Render(m.state.Source))
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
	b.WriteString(m.styles.Help.Render("t/space:toggle  o:on  O:off  esc:close"))

	return m.styles.Container.Render(b.String())
}

func (m SwitchModel) renderActions() string {
	// Toggle is always the primary action
	toggleBtn := m.styles.ActionFocus.Render("Toggle")
	onBtn := m.styles.Action.Render("On")
	offBtn := m.styles.Action.Render("Off")

	return lipgloss.JoinHorizontal(lipgloss.Center, toggleBtn, "  ", onBtn, "  ", offBtn)
}

// SetState updates the switch state.
func (m SwitchModel) SetState(state SwitchState) SwitchModel {
	m.state = state
	m.toggle = m.toggle.SetValue(state.Output)
	return m
}

// SetSize sets the panel dimensions.
func (m SwitchModel) SetSize(width, height int) SwitchModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m SwitchModel) SetFocused(focused bool) SwitchModel {
	m.focused = focused
	if focused {
		m.toggle = m.toggle.Focus()
	} else {
		m.toggle = m.toggle.Blur()
	}
	return m
}

// Focused returns whether the panel is focused.
func (m SwitchModel) Focused() bool {
	return m.focused
}

// State returns the current switch state.
func (m SwitchModel) State() SwitchState {
	return m.state
}

// formatPower formats a power value with appropriate units.
func formatPower(value float64) string {
	absVal := value
	if absVal < 0 {
		absVal = -absVal
	}
	if absVal >= 1000 {
		return fmt.Sprintf("%.2f kW", value/1000)
	}
	return fmt.Sprintf("%.1f W", value)
}
