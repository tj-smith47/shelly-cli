package control

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// PluginComponent represents a controllable component on a plugin device.
type PluginComponent struct {
	Type  string // "switch", "light", "cover"
	ID    int
	Name  string
	On    bool   // For switch/light
	State string // For cover (open/closed/opening/closing/stopped)
}

// PluginModel is the control panel for plugin-managed devices.
type PluginModel struct {
	ctx        context.Context
	svc        PluginService
	device     string
	platform   string
	components []PluginComponent
	cursor     int
	styles     Styles
	focused    bool
	width      int
	height     int
	loading    bool
	errorMsg   string
}

// NewPlugin creates a new plugin control panel.
func NewPlugin(ctx context.Context, svc PluginService, device, platform string, components []PluginComponent) PluginModel {
	return PluginModel{
		ctx:        ctx,
		svc:        svc,
		device:     device,
		platform:   platform,
		components: components,
		styles:     DefaultStyles(),
		focused:    true,
	}
}

// Init initializes the plugin control.
func (m PluginModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the plugin control.
func (m PluginModel) Update(msg tea.Msg) (PluginModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ActionMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		} else {
			m.errorMsg = ""
			m.updateComponentState(msg)
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

func (m PluginModel) updateComponentState(msg ActionMsg) {
	if m.cursor >= len(m.components) {
		return
	}
	comp := &m.components[m.cursor]
	switch msg.Action {
	case actionToggle:
		comp.On = !comp.On
	case actionOn:
		comp.On = true
	case actionOff:
		comp.On = false
	}
}

func (m PluginModel) handleKeyPress(msg tea.KeyPressMsg) (PluginModel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
		if m.cursor < len(m.components)-1 {
			m.cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("t", " "))):
		return m.executeAction(actionToggle)
	case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
		return m.executeAction(actionOn)
	case key.Matches(msg, key.NewBinding(key.WithKeys("O"))):
		return m.executeAction(actionOff)
	}
	return m, nil
}

func (m PluginModel) executeAction(action string) (PluginModel, tea.Cmd) {
	if len(m.components) == 0 || m.cursor >= len(m.components) {
		return m, nil
	}
	comp := m.components[m.cursor]
	m.loading = true
	m.errorMsg = ""

	return m, executeAction(m.device, TypePlugin, comp.ID, action, func() error {
		return m.svc.PluginControl(m.ctx, m.device, action, comp.Type, comp.ID)
	})
}

// View renders the plugin control panel.
func (m PluginModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(m.styles.Title.Render(fmt.Sprintf("Plugin Device (%s)", m.platform)))
	b.WriteString("\n\n")

	if len(m.components) == 0 {
		b.WriteString(m.styles.Muted.Render("No controllable components"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.Help.Render("esc:close"))
		return m.styles.Container.Render(b.String())
	}

	// Component list
	for i, comp := range m.components {
		selected := i == m.cursor
		b.WriteString(m.renderComponent(comp, selected))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Loading indicator
	if m.loading {
		b.WriteString(m.styles.Muted.Render("⋯ Executing..."))
		b.WriteString("\n")
	}

	// Error message
	if m.errorMsg != "" {
		b.WriteString(m.styles.Error.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	// Help
	b.WriteString(m.styles.Help.Render("t/space:toggle  o:on  O:off  j/k:nav  esc:close"))

	return m.styles.Container.Render(b.String())
}

func (m PluginModel) renderComponent(comp PluginComponent, selected bool) string {
	selector := "  "
	if selected {
		selector = "▶ "
	}

	// State indicator
	var stateStr string
	switch comp.Type {
	case "cover":
		stateStr = m.renderCoverState(comp.State)
	default:
		if comp.On {
			stateStr = m.styles.OnState.Render("● ON")
		} else {
			stateStr = m.styles.OffState.Render("○ OFF")
		}
	}

	// Name
	name := comp.Name
	if name == "" {
		name = fmt.Sprintf("%s:%d", comp.Type, comp.ID)
	}

	// Type badge
	badge := m.styles.Muted.Render(fmt.Sprintf("[%s]", comp.Type))

	line := fmt.Sprintf("%s%s  %s  %s", selector, stateStr, name, badge)
	if selected {
		return lipgloss.NewStyle().Bold(true).Render(line)
	}
	return line
}

func (m PluginModel) renderCoverState(state string) string {
	switch state {
	case "open":
		return m.styles.OnState.Render("▲ OPEN")
	case "closed":
		return m.styles.OffState.Render("▼ CLOSED")
	case coverStateOpening:
		return m.styles.Value.Render("▲ OPENING")
	case coverStateClosing:
		return m.styles.Value.Render("▼ CLOSING")
	case coverStateStopped:
		return m.styles.Muted.Render("■ STOPPED")
	default:
		return m.styles.Muted.Render("? " + state)
	}
}

// SetSize sets the panel dimensions.
func (m PluginModel) SetSize(width, height int) PluginModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m PluginModel) SetFocused(focused bool) PluginModel {
	m.focused = focused
	return m
}

// Focused returns whether the panel is focused.
func (m PluginModel) Focused() bool {
	return m.focused
}

// Components returns the plugin components.
func (m PluginModel) Components() []PluginComponent {
	return m.components
}

// Cursor returns the current cursor position.
func (m PluginModel) Cursor() int {
	return m.cursor
}
