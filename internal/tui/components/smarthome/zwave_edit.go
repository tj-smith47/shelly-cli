// Package smarthome provides TUI components for managing smart home protocol settings.
package smarthome

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/wireless"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// ZWaveResetResultMsg signals that a Z-Wave factory reset info view was confirmed.
type ZWaveResetResultMsg struct{}

// zwaveEditField identifies focusable fields in the Z-Wave edit modal.
type zwaveEditField int

const (
	zwaveFieldInclusion zwaveEditField = iota // Inclusion mode selector
	zwaveFieldExclusion                       // Exclusion instructions
	zwaveFieldConfig                          // Configuration parameters
	zwaveFieldReset                           // Factory reset instructions
)

// ZWaveEditModel represents the Z-Wave information edit modal.
type ZWaveEditModel struct {
	device  string
	visible bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

	// Device info
	deviceModel string
	deviceName  string
	isPro       bool
	supportsLR  bool

	// Inclusion mode selection
	inclusionIdx int
	modes        []wireless.ZWaveInclusionMode

	// Reset confirmation
	pendingReset bool

	// Focus
	field zwaveEditField
}

// NewZWaveEditModel creates a new Z-Wave information edit modal.
func NewZWaveEditModel() ZWaveEditModel {
	return ZWaveEditModel{
		styles: editmodal.DefaultStyles().WithLabelWidth(14),
		modes:  wireless.ZWaveInclusionModes(),
	}
}

// Show displays the edit modal with the given device and Z-Wave status.
func (m ZWaveEditModel) Show(device string, zw *shelly.TUIZWaveStatus) (ZWaveEditModel, tea.Cmd) {
	m.device = device
	m.visible = true
	m.pendingReset = false
	m.err = nil
	m.field = zwaveFieldInclusion
	m.inclusionIdx = 0

	if zw != nil {
		m.deviceModel = zw.DeviceModel
		m.deviceName = zw.DeviceName
		m.isPro = zw.IsPro
		m.supportsLR = zw.SupportsLR
	}

	return m, nil
}

// Hide hides the edit modal.
func (m ZWaveEditModel) Hide() ZWaveEditModel {
	m.visible = false
	m.pendingReset = false
	return m
}

// Visible returns whether the modal is visible.
func (m ZWaveEditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m ZWaveEditModel) SetSize(width, height int) ZWaveEditModel {
	m.width = width
	m.height = height
	return m
}

// Update handles messages.
func (m ZWaveEditModel) Update(msg tea.Msg) (ZWaveEditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m ZWaveEditModel) handleMessage(msg tea.Msg) (ZWaveEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m ZWaveEditModel) handleNavigation(msg messages.NavigationMsg) (ZWaveEditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		if m.field > 0 {
			m.field--
			m.pendingReset = false
		}
	case messages.NavDown:
		if m.field < zwaveFieldReset {
			m.field++
			m.pendingReset = false
		}
	case messages.NavLeft:
		if m.field == zwaveFieldInclusion && m.inclusionIdx > 0 {
			m.inclusionIdx--
		}
	case messages.NavRight:
		if m.field == zwaveFieldInclusion && m.inclusionIdx < len(m.modes)-1 {
			m.inclusionIdx++
		}
	case messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m ZWaveEditModel) handleKey(msg tea.KeyPressMsg) (ZWaveEditModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter, keyconst.KeyCtrlS:
		return m.handleAction()

	case keyconst.KeySpace, "t":
		if m.field == zwaveFieldInclusion {
			m.inclusionIdx = (m.inclusionIdx + 1) % len(m.modes)
		}
		return m, nil

	case "h", "left":
		if m.field == zwaveFieldInclusion && m.inclusionIdx > 0 {
			m.inclusionIdx--
		}
		return m, nil

	case "l", "right":
		if m.field == zwaveFieldInclusion && m.inclusionIdx < len(m.modes)-1 {
			m.inclusionIdx++
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if m.field < zwaveFieldReset {
			m.field++
			m.pendingReset = false
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if m.field > 0 {
			m.field--
			m.pendingReset = false
		}
		return m, nil
	}

	return m, nil
}

func (m ZWaveEditModel) handleAction() (ZWaveEditModel, tea.Cmd) {
	switch m.field {
	case zwaveFieldReset:
		return m.handleResetAction()
	default:
		// Other fields are informational - Enter just acknowledges
		return m, nil
	}
}

func (m ZWaveEditModel) handleResetAction() (ZWaveEditModel, tea.Cmd) {
	if m.pendingReset {
		// Second press - close modal (no RPC action, just informational)
		m.pendingReset = false
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}
	// First press - show reset warning and instructions
	m.pendingReset = true
	return m, nil
}

// View renders the edit modal.
func (m ZWaveEditModel) View() string {
	if !m.visible {
		return ""
	}

	footer := m.buildFooter()
	r := rendering.NewModal(m.width, m.height, "Z-Wave Configuration", footer)

	var content strings.Builder

	// Device info summary
	content.WriteString(m.renderDeviceInfo())
	content.WriteString("\n\n")

	// Inclusion mode selector
	content.WriteString(m.renderInclusionSection())

	// Exclusion section
	content.WriteString("\n\n")
	content.WriteString(m.renderExclusionSection())

	// Config parameters
	content.WriteString("\n\n")
	content.WriteString(m.renderConfigSection())

	// Factory reset
	content.WriteString("\n\n")
	content.WriteString(m.renderResetSection())

	return r.SetContent(content.String()).Render()
}

func (m ZWaveEditModel) buildFooter() string {
	if m.pendingReset {
		return "Press Enter again to view reset instructions, Esc to cancel"
	}
	switch m.field {
	case zwaveFieldInclusion:
		return "Space/h/l: Change mode | j/k: Navigate | Esc: Close"
	default:
		return "Enter: View details | j/k: Navigate | Esc: Close"
	}
}

func (m ZWaveEditModel) renderDeviceInfo() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Device:"))
	content.WriteString(" ")
	content.WriteString(m.styles.Value.Render(m.deviceName))

	content.WriteString("\n")
	content.WriteString(m.styles.Label.Render("Model:"))
	content.WriteString(" ")
	content.WriteString(m.styles.Value.Render(m.deviceModel))

	content.WriteString("\n")
	content.WriteString(m.styles.Label.Render("Series:"))
	content.WriteString(" ")
	if m.isPro {
		content.WriteString(m.styles.StatusOn.Render("Wave Pro"))
	} else {
		content.WriteString(m.styles.Value.Render("Wave"))
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Label.Render("Topology:"))
	content.WriteString(" ")
	if m.supportsLR {
		content.WriteString(m.styles.Value.Render("Mesh + Long Range"))
	} else {
		content.WriteString(m.styles.Value.Render("Mesh"))
	}

	return content.String()
}

func (m ZWaveEditModel) renderInclusionSection() string {
	selected := m.field == zwaveFieldInclusion
	var content strings.Builder

	// Mode selector row
	selector := m.styles.RenderSelector(selected)
	label := m.styles.LabelStyle(selected).Render("Inclusion:")
	content.WriteString(selector)
	content.WriteString(label)

	// Mode tabs
	for i, mode := range m.modes {
		name := wireless.ZWaveInclusionModeName(mode)
		if i == m.inclusionIdx {
			content.WriteString(m.styles.TabActive.Render(name))
		} else {
			content.WriteString(m.styles.Tab.Render(name))
		}
	}

	// Show instructions for selected mode when this section is focused
	if selected {
		steps := wireless.ZWaveInclusionSteps(m.modes[m.inclusionIdx])
		for _, step := range steps {
			content.WriteString("\n    ")
			content.WriteString(m.styles.Muted.Render(step))
		}
	}

	return content.String()
}

func (m ZWaveEditModel) renderExclusionSection() string {
	selected := m.field == zwaveFieldExclusion

	selector := m.styles.RenderSelector(selected)
	if selected {
		label := m.styles.ButtonFocus.Render("Exclusion Mode")
		var content strings.Builder
		content.WriteString(selector)
		content.WriteString(label)

		// Show exclusion instructions (button mode by default)
		steps := wireless.ZWaveExclusionSteps(wireless.ZWaveInclusionButton)
		for _, step := range steps {
			content.WriteString("\n    ")
			content.WriteString(m.styles.Muted.Render(step))
		}
		return content.String()
	}

	return selector + m.styles.Button.Render("Exclusion Mode")
}

func (m ZWaveEditModel) renderConfigSection() string {
	selected := m.field == zwaveFieldConfig

	selector := m.styles.RenderSelector(selected)
	if selected {
		label := m.styles.ButtonFocus.Render("Configuration Parameters")
		var content strings.Builder
		content.WriteString(selector)
		content.WriteString(label)

		params := wireless.ZWaveCommonConfigParams()
		for _, p := range params {
			content.WriteString("\n    ")
			content.WriteString(m.styles.Info.Render(fmt.Sprintf("P%d: %s", p.Number, p.Name)))
			content.WriteString("\n    ")
			content.WriteString(m.styles.Muted.Render(
				fmt.Sprintf("  Default: %d  Range: %d-%d", p.DefaultValue, p.MinValue, p.MaxValue),
			))
		}
		return content.String()
	}

	return selector + m.styles.Button.Render("Configuration Parameters")
}

func (m ZWaveEditModel) renderResetSection() string {
	selected := m.field == zwaveFieldReset

	if m.pendingReset {
		var content strings.Builder
		selector := m.styles.RenderSelector(selected)
		content.WriteString(selector)
		content.WriteString(m.styles.ButtonDanger.Render("FACTORY RESET INSTRUCTIONS"))

		// Show warning
		content.WriteString("\n\n    ")
		warning := wireless.ZWaveFactoryResetWarning()
		content.WriteString(m.styles.Warning.Render(warning))

		// Show steps
		steps := wireless.ZWaveFactoryResetSteps()
		content.WriteString("\n")
		for _, step := range steps {
			content.WriteString("\n    ")
			content.WriteString(m.styles.Muted.Render(step))
		}
		return content.String()
	}

	selector := m.styles.RenderSelector(selected)
	if selected {
		return selector + m.styles.ButtonDanger.Render("Factory Reset")
	}
	return selector + m.styles.Button.Render("Factory Reset")
}
