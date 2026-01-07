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

// RGBState holds the current state of an RGB component.
type RGBState struct {
	ID         int
	Name       string
	Output     bool
	Brightness int // 0-100
	Red        int // 0-255
	Green      int // 0-255
	Blue       int // 0-255
	White      int // 0-255 (for RGBW devices)
	Power      float64
	Source     string
}

// rgbFocus tracks which element is focused.
type rgbFocus int

const (
	rgbFocusToggle rgbFocus = iota
	rgbFocusBrightness
	rgbFocusRed
	rgbFocusGreen
	rgbFocusBlue
	rgbFocusPresets
)

// RGBModel is the control panel for RGB/RGBW components.
type RGBModel struct {
	ctx         context.Context
	svc         Service
	device      string
	state       RGBState
	toggle      form.Toggle
	brightness  form.Slider
	red         form.Slider
	green       form.Slider
	blue        form.Slider
	presetIdx   int
	styles      Styles
	focused     bool
	focus       rgbFocus
	width       int
	height      int
	loading     bool
	errorMsg    string
	showPresets bool
}

// NewRGB creates a new RGB control panel.
func NewRGB(ctx context.Context, svc Service, device string, state RGBState) RGBModel {
	return RGBModel{
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
			form.WithSliderValue(float64(clamp(state.Brightness, 0, 100))),
			form.WithSliderWidth(20),
			form.WithSliderFormat("%.0f%%"),
		),
		red: form.NewSlider(
			form.WithSliderLabel("Red"),
			form.WithSliderMin(0),
			form.WithSliderMax(255),
			form.WithSliderStep(5),
			form.WithSliderValue(float64(clamp(state.Red, 0, 255))),
			form.WithSliderWidth(20),
			form.WithSliderFormat("%.0f"),
		),
		green: form.NewSlider(
			form.WithSliderLabel("Green"),
			form.WithSliderMin(0),
			form.WithSliderMax(255),
			form.WithSliderStep(5),
			form.WithSliderValue(float64(clamp(state.Green, 0, 255))),
			form.WithSliderWidth(20),
			form.WithSliderFormat("%.0f"),
		),
		blue: form.NewSlider(
			form.WithSliderLabel("Blue"),
			form.WithSliderMin(0),
			form.WithSliderMax(255),
			form.WithSliderStep(5),
			form.WithSliderValue(float64(clamp(state.Blue, 0, 255))),
			form.WithSliderWidth(20),
			form.WithSliderFormat("%.0f"),
		),
		styles:  DefaultStyles(),
		focused: true,
		focus:   rgbFocusToggle,
	}
}

// clamp restricts val to the range [minVal, maxVal].
//
//nolint:unparam // minVal is always 0 but kept for clarity
func clamp(val, minVal, maxVal int) int {
	return max(minVal, min(val, maxVal))
}

// Init initializes the RGB control.
func (m RGBModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the RGB control.
func (m RGBModel) Update(msg tea.Msg) (RGBModel, tea.Cmd) {
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

func (m RGBModel) handleActionResult(msg ActionMsg) RGBModel {
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
	case actionColor:
		if rgb, ok := msg.Value.([3]int); ok {
			m.state.Red = rgb[0]
			m.state.Green = rgb[1]
			m.state.Blue = rgb[2]
			m.red = m.red.SetValue(float64(rgb[0]))
			m.green = m.green.SetValue(float64(rgb[1]))
			m.blue = m.blue.SetValue(float64(rgb[2]))
		}
	}
	return m
}

func (m RGBModel) handleKeyPress(msg tea.KeyPressMsg) (RGBModel, tea.Cmd) {
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
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		m.showPresets = !m.showPresets
		if m.showPresets {
			m.focus = rgbFocusPresets
		}
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
		return m.cycleFocus(), nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if m.focus == rgbFocusPresets && m.showPresets {
			return m.applyPreset()
		}
	}

	// Preset navigation
	if m.showPresets && m.focus == rgbFocusPresets {
		return m.handlePresetNavigation(msg)
	}

	// Focus-specific slider handling
	return m.handleSliderInput(msg)
}

func (m RGBModel) handlePresetNavigation(msg tea.KeyPressMsg) (RGBModel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
		if m.presetIdx > 0 {
			m.presetIdx--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
		if m.presetIdx < len(PresetColors)-1 {
			m.presetIdx++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("escape"))):
		m.showPresets = false
		m.focus = rgbFocusToggle
	}
	return m, nil
}

func (m RGBModel) handleSliderInput(msg tea.KeyPressMsg) (RGBModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {
	case rgbFocusToggle, rgbFocusPresets:
		// No slider input for toggle or presets
		return m, nil
	case rgbFocusBrightness:
		m.brightness, cmd = m.brightness.Update(msg)
		newVal := int(m.brightness.Value())
		if newVal != m.state.Brightness {
			return m.executeBrightness(newVal)
		}
	case rgbFocusRed:
		m.red, cmd = m.red.Update(msg)
		newVal := int(m.red.Value())
		if newVal != m.state.Red {
			return m.executeColor(newVal, m.state.Green, m.state.Blue)
		}
	case rgbFocusGreen:
		m.green, cmd = m.green.Update(msg)
		newVal := int(m.green.Value())
		if newVal != m.state.Green {
			return m.executeColor(m.state.Red, newVal, m.state.Blue)
		}
	case rgbFocusBlue:
		m.blue, cmd = m.blue.Update(msg)
		newVal := int(m.blue.Value())
		if newVal != m.state.Blue {
			return m.executeColor(m.state.Red, m.state.Green, newVal)
		}
	}

	return m, cmd
}

func (m RGBModel) cycleFocus() RGBModel {
	m = m.blurAll()
	if m.showPresets {
		// Skip presets focus when cycling via tab
		m.focus = (m.focus + 1) % 5
		if m.focus == rgbFocusPresets {
			m.focus = rgbFocusToggle
		}
	} else {
		m.focus = (m.focus + 1) % 5
	}
	return m.focusCurrent()
}

func (m RGBModel) blurAll() RGBModel {
	m.toggle = m.toggle.Blur()
	m.brightness = m.brightness.Blur()
	m.red = m.red.Blur()
	m.green = m.green.Blur()
	m.blue = m.blue.Blur()
	return m
}

func (m RGBModel) focusCurrent() RGBModel {
	switch m.focus {
	case rgbFocusToggle:
		m.toggle = m.toggle.Focus()
	case rgbFocusBrightness:
		m.brightness = m.brightness.Focus()
	case rgbFocusRed:
		m.red = m.red.Focus()
	case rgbFocusGreen:
		m.green = m.green.Focus()
	case rgbFocusBlue:
		m.blue = m.blue.Focus()
	case rgbFocusPresets:
		// Presets don't have a focusable element
	}
	return m
}

func (m RGBModel) executeToggle() (RGBModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeRGB, m.state.ID, actionToggle, func() error {
		return m.svc.RGBToggle(m.ctx, m.device, m.state.ID)
	})
}

func (m RGBModel) executeOn() (RGBModel, tea.Cmd) {
	if m.state.Output {
		return m, nil
	}
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeRGB, m.state.ID, actionOn, func() error {
		return m.svc.RGBOn(m.ctx, m.device, m.state.ID)
	})
}

func (m RGBModel) executeOff() (RGBModel, tea.Cmd) {
	if !m.state.Output {
		return m, nil
	}
	m.loading = true
	m.errorMsg = ""
	return m, executeAction(m.device, TypeRGB, m.state.ID, actionOff, func() error {
		return m.svc.RGBOff(m.ctx, m.device, m.state.ID)
	})
}

func (m RGBModel) adjustBrightness(delta int) (RGBModel, tea.Cmd) {
	newVal := clamp(m.state.Brightness+delta, 0, 100)
	if newVal == m.state.Brightness {
		return m, nil
	}
	return m.executeBrightness(newVal)
}

func (m RGBModel) executeBrightness(brightness int) (RGBModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.RGBBrightness(m.ctx, m.device, m.state.ID, brightness)
		return ActionMsg{
			Device:    m.device,
			Component: TypeRGB,
			ID:        m.state.ID,
			Action:    actionBrightness,
			Value:     brightness,
			Err:       err,
		}
	}
}

func (m RGBModel) executeColor(r, g, b int) (RGBModel, tea.Cmd) {
	m.loading = true
	m.errorMsg = ""
	return m, func() tea.Msg {
		err := m.svc.RGBColor(m.ctx, m.device, m.state.ID, r, g, b)
		return ActionMsg{
			Device:    m.device,
			Component: TypeRGB,
			ID:        m.state.ID,
			Action:    actionColor,
			Value:     [3]int{r, g, b},
			Err:       err,
		}
	}
}

func (m RGBModel) applyPreset() (RGBModel, tea.Cmd) {
	if m.presetIdx >= 0 && m.presetIdx < len(PresetColors) {
		preset := PresetColors[m.presetIdx]
		return m.executeColor(preset.RGB[0], preset.RGB[1], preset.RGB[2])
	}
	return m, nil
}

// View renders the RGB control panel.
func (m RGBModel) View() string {
	var b strings.Builder

	// Title
	name := m.state.Name
	if name == "" {
		name = fmt.Sprintf("RGB %d", m.state.ID)
	}
	b.WriteString(m.styles.Title.Render(name))
	b.WriteString("\n\n")

	// State indicator with color preview
	b.WriteString(m.renderStateWithColor())
	b.WriteString("\n\n")

	// Brightness slider
	b.WriteString(m.brightness.View())
	b.WriteString("\n\n")

	// RGB sliders
	b.WriteString(m.red.View())
	b.WriteString("\n")
	b.WriteString(m.green.View())
	b.WriteString("\n")
	b.WriteString(m.blue.View())
	b.WriteString("\n")

	// Preset colors
	if m.showPresets {
		b.WriteString("\n")
		b.WriteString(m.renderPresets())
		b.WriteString("\n")
	}

	// Power reading
	if m.state.Power != 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Label.Render("Power:"))
		b.WriteString(m.styles.Power.Render(formatPower(m.state.Power)))
	}

	b.WriteString("\n\n")

	// Action buttons
	b.WriteString(m.renderActions())
	b.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		b.WriteString(m.styles.Error.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	// Help
	help := "t:toggle  +/-:brightness  c:colors  tab:focus  esc:close"
	b.WriteString(m.styles.Help.Render(help))

	return m.styles.Container.Render(b.String())
}

func (m RGBModel) renderStateWithColor() string {
	stateStr := m.styles.OffState.Render("○ OFF")
	if m.state.Output {
		stateStr = m.styles.OnState.Render("● ON")
	}
	if m.loading {
		stateStr = m.styles.Muted.Render("⋯ Loading...")
	}

	// Color preview block
	colorPreview := lipgloss.NewStyle().
		Background(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", m.state.Red, m.state.Green, m.state.Blue))).
		Width(4).
		Render("    ")

	return m.styles.Label.Render("State:") + stateStr + "  " + colorPreview
}

func (m RGBModel) renderPresets() string {
	parts := make([]string, 0, len(PresetColors))
	for i, preset := range PresetColors {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", preset.RGB[0], preset.RGB[1], preset.RGB[2]))).
			Padding(0, 1)
		if i == m.presetIdx {
			style = style.Bold(true).Underline(true)
		}
		parts = append(parts, style.Render(preset.Name[:1]))
	}
	return m.styles.Label.Render("Presets: ") + strings.Join(parts, " ")
}

func (m RGBModel) renderActions() string {
	toggleBtn := m.styles.ActionFocus.Render("Toggle")
	onBtn := m.styles.Action.Render("On")
	offBtn := m.styles.Action.Render("Off")

	return lipgloss.JoinHorizontal(lipgloss.Center, toggleBtn, "  ", onBtn, "  ", offBtn)
}

// SetState updates the RGB state.
func (m RGBModel) SetState(state RGBState) RGBModel {
	m.state = state
	m.toggle = m.toggle.SetValue(state.Output)
	m.brightness = m.brightness.SetValue(float64(clamp(state.Brightness, 0, 100)))
	m.red = m.red.SetValue(float64(clamp(state.Red, 0, 255)))
	m.green = m.green.SetValue(float64(clamp(state.Green, 0, 255)))
	m.blue = m.blue.SetValue(float64(clamp(state.Blue, 0, 255)))
	return m
}

// SetSize sets the panel dimensions.
func (m RGBModel) SetSize(width, height int) RGBModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m RGBModel) SetFocused(focused bool) RGBModel {
	m.focused = focused
	if !focused {
		m = m.blurAll()
	} else {
		m = m.focusCurrent()
	}
	return m
}

// Focused returns whether the panel is focused.
func (m RGBModel) Focused() bool {
	return m.focused
}

// State returns the current RGB state.
func (m RGBModel) State() RGBState {
	return m.state
}
