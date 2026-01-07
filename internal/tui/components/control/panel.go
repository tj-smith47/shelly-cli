package control

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// PanelCloseMsg is sent when the control panel should close.
type PanelCloseMsg struct{}

// Panel is the main control panel that can display any component type.
type Panel struct {
	ctx        context.Context
	svc        Service
	device     string
	compType   ComponentType
	compID     int
	switchCtrl *SwitchModel
	lightCtrl  *LightModel
	rgbCtrl    *RGBModel
	coverCtrl  *CoverModel
	thermoCtrl *ThermostatModel
	visible    bool
	width      int
	height     int
	styles     panelStyles
}

type panelStyles struct {
	Overlay lipgloss.Style
	Panel   lipgloss.Style
}

func defaultPanelStyles() panelStyles {
	colors := theme.GetSemanticColors()
	return panelStyles{
		Overlay: lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")).
			Foreground(colors.Text),
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(1, 2),
	}
}

// NewPanel creates a new control panel.
func NewPanel(ctx context.Context, svc Service) Panel {
	return Panel{
		ctx:    ctx,
		svc:    svc,
		styles: defaultPanelStyles(),
	}
}

// ShowSwitch displays the switch control panel.
func (p Panel) ShowSwitch(device string, state SwitchState) Panel {
	ctrl := NewSwitch(p.ctx, p.svc, device, state)
	p.switchCtrl = &ctrl
	p.device = device
	p.compType = TypeSwitch
	p.compID = state.ID
	p.visible = true
	return p.clearOthers(TypeSwitch)
}

// ShowLight displays the light control panel.
func (p Panel) ShowLight(device string, state LightState) Panel {
	ctrl := NewLight(p.ctx, p.svc, device, state)
	p.lightCtrl = &ctrl
	p.device = device
	p.compType = TypeLight
	p.compID = state.ID
	p.visible = true
	return p.clearOthers(TypeLight)
}

// ShowRGB displays the RGB control panel.
func (p Panel) ShowRGB(device string, state RGBState) Panel {
	ctrl := NewRGB(p.ctx, p.svc, device, state)
	p.rgbCtrl = &ctrl
	p.device = device
	p.compType = TypeRGB
	p.compID = state.ID
	p.visible = true
	return p.clearOthers(TypeRGB)
}

// ShowCover displays the cover control panel.
func (p Panel) ShowCover(device string, state CoverState) Panel {
	ctrl := NewCover(p.ctx, p.svc, device, state)
	p.coverCtrl = &ctrl
	p.device = device
	p.compType = TypeCover
	p.compID = state.ID
	p.visible = true
	return p.clearOthers(TypeCover)
}

// ShowThermostat displays the thermostat control panel.
func (p Panel) ShowThermostat(device string, state ThermostatState) Panel {
	ctrl := NewThermostat(p.ctx, p.svc, device, state)
	p.thermoCtrl = &ctrl
	p.device = device
	p.compType = TypeThermostat
	p.compID = state.ID
	p.visible = true
	return p.clearOthers(TypeThermostat)
}

func (p Panel) clearOthers(keep ComponentType) Panel {
	if keep != TypeSwitch {
		p.switchCtrl = nil
	}
	if keep != TypeLight {
		p.lightCtrl = nil
	}
	if keep != TypeRGB {
		p.rgbCtrl = nil
	}
	if keep != TypeCover {
		p.coverCtrl = nil
	}
	if keep != TypeThermostat {
		p.thermoCtrl = nil
	}
	return p
}

// Hide hides the control panel.
func (p Panel) Hide() Panel {
	p.visible = false
	return p
}

// Visible returns whether the panel is visible.
func (p Panel) Visible() bool {
	return p.visible
}

// ComponentType returns the current component type.
func (p Panel) ComponentType() ComponentType {
	return p.compType
}

// ComponentID returns the current component ID.
func (p Panel) ComponentID() int {
	return p.compID
}

// Device returns the current device.
func (p Panel) Device() string {
	return p.device
}

// Init initializes the panel.
func (p Panel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the panel.
func (p Panel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	// Handle escape to close
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if keyMsg.String() == "escape" || keyMsg.String() == "q" {
			p.visible = false
			return p, func() tea.Msg { return PanelCloseMsg{} }
		}
	}

	// Delegate to active control
	var cmd tea.Cmd
	switch p.compType {
	case TypeSwitch:
		if p.switchCtrl != nil {
			ctrl, c := p.switchCtrl.Update(msg)
			p.switchCtrl = &ctrl
			cmd = c
		}
	case TypeLight:
		if p.lightCtrl != nil {
			ctrl, c := p.lightCtrl.Update(msg)
			p.lightCtrl = &ctrl
			cmd = c
		}
	case TypeRGB:
		if p.rgbCtrl != nil {
			ctrl, c := p.rgbCtrl.Update(msg)
			p.rgbCtrl = &ctrl
			cmd = c
		}
	case TypeCover:
		if p.coverCtrl != nil {
			ctrl, c := p.coverCtrl.Update(msg)
			p.coverCtrl = &ctrl
			cmd = c
		}
	case TypeThermostat:
		if p.thermoCtrl != nil {
			ctrl, c := p.thermoCtrl.Update(msg)
			p.thermoCtrl = &ctrl
			cmd = c
		}
	}

	return p, cmd
}

// View renders the panel.
func (p Panel) View() string {
	if !p.visible {
		return ""
	}

	var content string
	switch p.compType {
	case TypeSwitch:
		if p.switchCtrl != nil {
			content = p.switchCtrl.View()
		}
	case TypeLight:
		if p.lightCtrl != nil {
			content = p.lightCtrl.View()
		}
	case TypeRGB:
		if p.rgbCtrl != nil {
			content = p.rgbCtrl.View()
		}
	case TypeCover:
		if p.coverCtrl != nil {
			content = p.coverCtrl.View()
		}
	case TypeThermostat:
		if p.thermoCtrl != nil {
			content = p.thermoCtrl.View()
		}
	}

	if content == "" {
		return ""
	}

	// Center the panel
	panelWidth := 50
	if p.width > 0 && panelWidth > p.width-4 {
		panelWidth = p.width - 4
	}

	return p.styles.Panel.Width(panelWidth).Render(content)
}

// SetSize sets the panel dimensions.
func (p Panel) SetSize(width, height int) Panel {
	p.width = width
	p.height = height

	// Propagate to active control
	innerWidth := width - 10
	innerHeight := height - 6
	if innerWidth < 40 {
		innerWidth = 40
	}
	if innerHeight < 20 {
		innerHeight = 20
	}

	if p.switchCtrl != nil {
		ctrl := p.switchCtrl.SetSize(innerWidth, innerHeight)
		p.switchCtrl = &ctrl
	}
	if p.lightCtrl != nil {
		ctrl := p.lightCtrl.SetSize(innerWidth, innerHeight)
		p.lightCtrl = &ctrl
	}
	if p.rgbCtrl != nil {
		ctrl := p.rgbCtrl.SetSize(innerWidth, innerHeight)
		p.rgbCtrl = &ctrl
	}
	if p.coverCtrl != nil {
		ctrl := p.coverCtrl.SetSize(innerWidth, innerHeight)
		p.coverCtrl = &ctrl
	}
	if p.thermoCtrl != nil {
		ctrl := p.thermoCtrl.SetSize(innerWidth, innerHeight)
		p.thermoCtrl = &ctrl
	}

	return p
}

// UpdateSwitchState updates the switch state if active.
func (p Panel) UpdateSwitchState(state SwitchState) Panel {
	if p.compType == TypeSwitch && p.compID == state.ID && p.switchCtrl != nil {
		ctrl := p.switchCtrl.SetState(state)
		p.switchCtrl = &ctrl
	}
	return p
}

// UpdateLightState updates the light state if active.
func (p Panel) UpdateLightState(state LightState) Panel {
	if p.compType == TypeLight && p.compID == state.ID && p.lightCtrl != nil {
		ctrl := p.lightCtrl.SetState(state)
		p.lightCtrl = &ctrl
	}
	return p
}

// UpdateRGBState updates the RGB state if active.
func (p Panel) UpdateRGBState(state RGBState) Panel {
	if p.compType == TypeRGB && p.compID == state.ID && p.rgbCtrl != nil {
		ctrl := p.rgbCtrl.SetState(state)
		p.rgbCtrl = &ctrl
	}
	return p
}

// UpdateCoverState updates the cover state if active.
func (p Panel) UpdateCoverState(state CoverState) Panel {
	if p.compType == TypeCover && p.compID == state.ID && p.coverCtrl != nil {
		ctrl := p.coverCtrl.SetState(state)
		p.coverCtrl = &ctrl
	}
	return p
}

// UpdateThermostatState updates the thermostat state if active.
func (p Panel) UpdateThermostatState(state ThermostatState) Panel {
	if p.compType == TypeThermostat && p.compID == state.ID && p.thermoCtrl != nil {
		ctrl := p.thermoCtrl.SetState(state)
		p.thermoCtrl = &ctrl
	}
	return p
}
