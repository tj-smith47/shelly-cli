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

// LightState holds the current state of a light.
type LightState struct {
	ID         int
	Name       string
	Output     bool
	Brightness int // 0-100
	Power      float64
	Source     string
}

// lightFocus tracks which element is focused.
type lightFocus int

const (
	lightFocusToggle lightFocus = iota
	lightFocusBrightness
)

// LightModel is the control panel for light/dimmer components.
type LightModel struct {
	ctx        context.Context
	svc        Service
	device     string
	state      LightState
	toggle     form.Toggle
	brightness form.Slider
	styles     Styles
	focused    bool
	focus      lightFocus
	width      int
	height     int
	loading    bool
	errorMsg   string
}

// NewLight creates a new light control panel.
func NewLight(ctx context.Context, svc Service, device string, state LightState) LightModel {
	brightness := state.Brightness
	if brightness < 0 {
		brightness = 0
	}
	if brightness > 100 {
		brightness = 100
	}

	return LightModel{
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
		brightness: form.NewSlider(
			form.WithSliderLabel("Brightness"),
			form.WithSliderMin(0),
			form.WithSliderMax(100),
			form.WithSliderStep(5),
			form.WithSliderValue(float64(brightness)),
			form.WithSliderWidth(20),
			form.WithSliderFormat("%.0f%%"),
		),
		styles:  DefaultStyles(),
		focused: true,
		focus:   lightFocusToggle,
	}
}

// Init initializes the light control.
func (m LightModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the light control.
func (m LightModel) Update(msg tea.Msg) (LightModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ActionMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		} else {
			m.errorMsg = ""
			switch msg.Action {
			case actionToggle:
				m.state.Output = !m.state.Output
				m.toggle = m.toggle.SetValue(m.state.Output)
			case actionOn:
				m.state.Output = true
				m.toggle = m.toggle.SetValue(true)
			case actionOff:
				m.state.Output = false
				m.toggle = m.toggle.SetValue(false)
			case actionBrightness:
				if val, ok := msg.Value.(int); ok {
					m.state.Brightness = val
				}
			}
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

func (m LightModel) handleKeyPress(msg tea.KeyPressMsg) (LightModel, tea.Cmd) {
	// Global shortcuts
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("t", " "))):
		return m.executeToggle()
	case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
		return m.executeOn()
	case key.Matches(msg, key.NewBinding(key.WithKeys("O"))):
		return m.executeOff()
	case key.Matches(msg, key.NewBinding(key.WithKeys("+", "="))):
		return m.adjustBrightness(5)
	case key.Matches(msg, key.NewBinding(key.WithKeys("-", "_"))):
		return m.adjustBrightness(-5)
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
		return m.cycleFocus(), nil
	}

	// Focus-specific handling
	if m.focus == lightFocusBrightness {
		var cmd tea.Cmd
		m.brightness, cmd = m.brightness.Update(msg)
		// Check if brightness changed
		newVal := int(m.brightness.Value())
		if newVal != m.state.Brightness {
			return m.executeBrightness(newVal)
		}
		return m, cmd
	}

	return m, nil
}

func (m LightModel) cycleFocus() LightModel {
	m.focus = (m.focus + 1) % 2
	if m.focus == lightFocusToggle {
		m.toggle = m.toggle.Focus()
		m.brightness = m.brightness.Blur()
	} else {
		m.toggle = m.toggle.Blur()
		m.brightness = m.brightness.Focus()
	}
	return m
}

func (m LightModel) executeToggle() (LightModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeLight, m.state.ID, actionToggle, func() error {
		return m.svc.LightToggle(m.ctx, m.device, m.state.ID)
	})
}

func (m LightModel) executeOn() (LightModel, tea.Cmd) {
	if m.state.Output {
		return m, nil
	}
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeLight, m.state.ID, actionOn, func() error {
		return m.svc.LightOn(m.ctx, m.device, m.state.ID)
	})
}

func (m LightModel) executeOff() (LightModel, tea.Cmd) {
	if !m.state.Output {
		return m, nil
	}
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeLight, m.state.ID, actionOff, func() error {
		return m.svc.LightOff(m.ctx, m.device, m.state.ID)
	})
}

func (m LightModel) adjustBrightness(delta int) (LightModel, tea.Cmd) {
	newVal := m.state.Brightness + delta
	if newVal < 0 {
		newVal = 0
	}
	if newVal > 100 {
		newVal = 100
	}
	if newVal == m.state.Brightness {
		return m, nil
	}
	return m.executeBrightness(newVal)
}

func (m LightModel) executeBrightness(brightness int) (LightModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.LightBrightness(m.ctx, m.device, m.state.ID, brightness)
		return ActionMsg{
			Device:    m.device,
			Component: TypeLight,
			ID:        m.state.ID,
			Action:    actionBrightness,
			Value:     brightness,
			Err:       err,
		}
	}
}

// View renders the light control panel.
func (m LightModel) View() string {
	var b strings.Builder

	// Title
	name := m.state.Name
	if name == "" {
		name = fmt.Sprintf("Light %d", m.state.ID)
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
	b.WriteString("\n\n")

	// Brightness slider
	b.WriteString(m.brightness.View())
	b.WriteString("\n")

	// Power reading
	if m.state.Power != 0 {
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
	b.WriteString(m.styles.Help.Render("t/space:toggle  +/-:brightness  tab:focus  esc:close"))

	return m.styles.Container.Render(b.String())
}

func (m LightModel) renderActions() string {
	toggleStyle := m.styles.ActionFocus
	onStyle := m.styles.Action
	offStyle := m.styles.Action

	toggleBtn := toggleStyle.Render("Toggle")
	onBtn := onStyle.Render("On")
	offBtn := offStyle.Render("Off")

	return lipgloss.JoinHorizontal(lipgloss.Center, toggleBtn, "  ", onBtn, "  ", offBtn)
}

// SetState updates the light state.
func (m LightModel) SetState(state LightState) LightModel {
	m.state = state
	m.toggle = m.toggle.SetValue(state.Output)
	if state.Brightness >= 0 && state.Brightness <= 100 {
		m.brightness = m.brightness.SetValue(float64(state.Brightness))
	}
	return m
}

// SetSize sets the panel dimensions.
func (m LightModel) SetSize(width, height int) LightModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m LightModel) SetFocused(focused bool) LightModel {
	m.focused = focused
	if focused && m.focus == lightFocusToggle {
		m.toggle = m.toggle.Focus()
	} else {
		m.toggle = m.toggle.Blur()
	}
	if focused && m.focus == lightFocusBrightness {
		m.brightness = m.brightness.Focus()
	} else {
		m.brightness = m.brightness.Blur()
	}
	return m
}

// Focused returns whether the panel is focused.
func (m LightModel) Focused() bool {
	return m.focused
}

// State returns the current light state.
func (m LightModel) State() LightState {
	return m.state
}
