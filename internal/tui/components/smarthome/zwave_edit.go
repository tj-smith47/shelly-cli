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

// zwaveFieldCount is the total number of focusable fields.
const zwaveFieldCount = 4

// ZWaveEditModel represents the Z-Wave information edit modal.
type ZWaveEditModel struct {
	editmodal.Base

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
}

// NewZWaveEditModel creates a new Z-Wave information edit modal.
func NewZWaveEditModel() ZWaveEditModel {
	return ZWaveEditModel{
		Base: editmodal.Base{
			Styles: editmodal.DefaultStyles().WithLabelWidth(14),
		},
		modes: wireless.ZWaveInclusionModes(),
	}
}

// Show displays the edit modal with the given device and Z-Wave status.
func (m ZWaveEditModel) Show(device string, zw *shelly.TUIZWaveStatus) (ZWaveEditModel, tea.Cmd) {
	m.Base.Show(device, zwaveFieldCount)
	m.pendingReset = false
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
	m.Base.Hide()
	m.pendingReset = false
	return m
}

// Visible returns whether the modal is visible.
func (m ZWaveEditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m ZWaveEditModel) SetSize(width, height int) ZWaveEditModel {
	m.Base.SetSize(width, height)
	return m
}

// Update handles messages.
func (m ZWaveEditModel) Update(msg tea.Msg) (ZWaveEditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m ZWaveEditModel) handleMessage(msg tea.Msg) (ZWaveEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case tea.KeyPressMsg:
		action := m.HandleKey(msg)
		if action != editmodal.ActionNone {
			return m.applyAction(action)
		}
		return m.handleCustomKey(msg)
	}

	return m, nil
}

func (m ZWaveEditModel) handleNavigation(msg messages.NavigationMsg) (ZWaveEditModel, tea.Cmd) {
	// Handle left/right for inclusion mode (custom, not in Base)
	switch msg.Direction {
	case messages.NavLeft:
		if zwaveEditField(m.Cursor) == zwaveFieldInclusion && m.inclusionIdx > 0 {
			m.inclusionIdx--
		}
		return m, nil
	case messages.NavRight:
		if zwaveEditField(m.Cursor) == zwaveFieldInclusion && m.inclusionIdx < len(m.modes)-1 {
			m.inclusionIdx++
		}
		return m, nil
	default:
		// Use Base for up/down
		action := m.HandleNavigation(msg)
		if action != editmodal.ActionNone {
			return m.applyAction(action)
		}
		return m, nil
	}
}

func (m ZWaveEditModel) applyAction(action editmodal.KeyAction) (ZWaveEditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionNone:
		return m, nil
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		// Z-Wave is info-only; Enter triggers handleAction instead of save
		return m.handleAction()
	case editmodal.ActionNavDown:
		if m.Cursor < zwaveFieldCount-1 {
			m.Cursor++
			m.pendingReset = false
		}
		return m, nil
	case editmodal.ActionNavUp:
		if m.Cursor > 0 {
			m.Cursor--
			m.pendingReset = false
		}
		return m, nil
	case editmodal.ActionNext:
		if m.Cursor < zwaveFieldCount-1 {
			m.Cursor++
			m.pendingReset = false
		}
		return m, nil
	case editmodal.ActionPrev:
		if m.Cursor > 0 {
			m.Cursor--
			m.pendingReset = false
		}
		return m, nil
	}
	return m, nil
}

func (m ZWaveEditModel) handleCustomKey(msg tea.KeyPressMsg) (ZWaveEditModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeySpace, "t":
		if zwaveEditField(m.Cursor) == zwaveFieldInclusion {
			m.inclusionIdx = (m.inclusionIdx + 1) % len(m.modes)
		}
		return m, nil

	case "h", "left":
		if zwaveEditField(m.Cursor) == zwaveFieldInclusion && m.inclusionIdx > 0 {
			m.inclusionIdx--
		}
		return m, nil

	case "l", "right":
		if zwaveEditField(m.Cursor) == zwaveFieldInclusion && m.inclusionIdx < len(m.modes)-1 {
			m.inclusionIdx++
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if m.Cursor < zwaveFieldCount-1 {
			m.Cursor++
			m.pendingReset = false
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
			m.pendingReset = false
		}
		return m, nil
	}

	return m, nil
}

func (m ZWaveEditModel) handleAction() (ZWaveEditModel, tea.Cmd) {
	switch zwaveEditField(m.Cursor) {
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
	if !m.Visible() {
		return ""
	}

	footer := m.buildFooter()

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

	return m.RenderModal("Z-Wave Configuration", content.String(), footer)
}

func (m ZWaveEditModel) buildFooter() string {
	if m.pendingReset {
		return "Press Enter again to view reset instructions, Esc to cancel"
	}
	switch zwaveEditField(m.Cursor) {
	case zwaveFieldInclusion:
		return "Space/h/l: Change mode | j/k: Navigate | Esc: Close"
	default:
		return "Enter: View details | j/k: Navigate | Esc: Close"
	}
}

func (m ZWaveEditModel) renderDeviceInfo() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Device:"))
	content.WriteString(" ")
	content.WriteString(m.Styles.Value.Render(m.deviceName))

	content.WriteString("\n")
	content.WriteString(m.Styles.Label.Render("Model:"))
	content.WriteString(" ")
	content.WriteString(m.Styles.Value.Render(m.deviceModel))

	content.WriteString("\n")
	content.WriteString(m.Styles.Label.Render("Series:"))
	content.WriteString(" ")
	if m.isPro {
		content.WriteString(m.Styles.StatusOn.Render("Wave Pro"))
	} else {
		content.WriteString(m.Styles.Value.Render("Wave"))
	}

	content.WriteString("\n")
	content.WriteString(m.Styles.Label.Render("Topology:"))
	content.WriteString(" ")
	if m.supportsLR {
		content.WriteString(m.Styles.Value.Render("Mesh + Long Range"))
	} else {
		content.WriteString(m.Styles.Value.Render("Mesh"))
	}

	return content.String()
}

func (m ZWaveEditModel) renderInclusionSection() string {
	selected := zwaveEditField(m.Cursor) == zwaveFieldInclusion
	var content strings.Builder

	// Mode selector row
	selector := m.Styles.RenderSelector(selected)
	label := m.Styles.LabelStyle(selected).Render("Inclusion:")
	content.WriteString(selector)
	content.WriteString(label)

	// Mode tabs
	for i, mode := range m.modes {
		name := wireless.ZWaveInclusionModeName(mode)
		if i == m.inclusionIdx {
			content.WriteString(m.Styles.TabActive.Render(name))
		} else {
			content.WriteString(m.Styles.Tab.Render(name))
		}
	}

	// Show instructions for selected mode when this section is focused
	if selected {
		steps := wireless.ZWaveInclusionSteps(m.modes[m.inclusionIdx])
		for _, step := range steps {
			content.WriteString("\n    ")
			content.WriteString(m.Styles.Muted.Render(step))
		}
	}

	return content.String()
}

func (m ZWaveEditModel) renderExclusionSection() string {
	selected := zwaveEditField(m.Cursor) == zwaveFieldExclusion

	selector := m.Styles.RenderSelector(selected)
	if selected {
		label := m.Styles.ButtonFocus.Render("Exclusion Mode")
		var content strings.Builder
		content.WriteString(selector)
		content.WriteString(label)

		// Show exclusion instructions (button mode by default)
		steps := wireless.ZWaveExclusionSteps(wireless.ZWaveInclusionButton)
		for _, step := range steps {
			content.WriteString("\n    ")
			content.WriteString(m.Styles.Muted.Render(step))
		}
		return content.String()
	}

	return selector + m.Styles.Button.Render("Exclusion Mode")
}

func (m ZWaveEditModel) renderConfigSection() string {
	selected := zwaveEditField(m.Cursor) == zwaveFieldConfig

	selector := m.Styles.RenderSelector(selected)
	if selected {
		label := m.Styles.ButtonFocus.Render("Configuration Parameters")
		var content strings.Builder
		content.WriteString(selector)
		content.WriteString(label)

		params := wireless.ZWaveCommonConfigParams()
		for _, p := range params {
			content.WriteString("\n    ")
			content.WriteString(m.Styles.Info.Render(fmt.Sprintf("P%d: %s", p.Number, p.Name)))
			content.WriteString("\n    ")
			content.WriteString(m.Styles.Muted.Render(
				fmt.Sprintf("  Default: %d  Range: %d-%d", p.DefaultValue, p.MinValue, p.MaxValue),
			))
		}
		return content.String()
	}

	return selector + m.Styles.Button.Render("Configuration Parameters")
}

func (m ZWaveEditModel) renderResetSection() string {
	selected := zwaveEditField(m.Cursor) == zwaveFieldReset

	if m.pendingReset {
		var content strings.Builder
		selector := m.Styles.RenderSelector(selected)
		content.WriteString(selector)
		content.WriteString(m.Styles.ButtonDanger.Render("FACTORY RESET INSTRUCTIONS"))

		// Show warning
		content.WriteString("\n\n    ")
		warning := wireless.ZWaveFactoryResetWarning()
		content.WriteString(m.Styles.Warning.Render(warning))

		// Show steps
		steps := wireless.ZWaveFactoryResetSteps()
		content.WriteString("\n")
		for _, step := range steps {
			content.WriteString("\n    ")
			content.WriteString(m.Styles.Muted.Render(step))
		}
		return content.String()
	}

	selector := m.Styles.RenderSelector(selected)
	if selected {
		return selector + m.Styles.ButtonDanger.Render("Factory Reset")
	}
	return selector + m.Styles.Button.Render("Factory Reset")
}
