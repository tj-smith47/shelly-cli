// Package smarthome provides TUI components for managing smart home protocol settings.
package smarthome

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// ModbusToggleResultMsg signals that a Modbus enable/disable toggle completed.
type ModbusToggleResultMsg struct {
	Enabled bool
	Err     error
}

// modbusEditField identifies focusable fields in the Modbus edit modal.
type modbusEditField int

const (
	modbusFieldEnable modbusEditField = iota // Enable/disable toggle
)

// modbusFieldCount is the total number of focusable fields (enable toggle only).
const modbusFieldCount = 1

// ModbusEditModel represents the Modbus configuration edit modal.
type ModbusEditModel struct {
	ctx    context.Context
	svc    *shelly.Service
	device string

	visible bool
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

	// Modbus state
	enabled bool

	// Pending changes
	pendingEnabled bool

	// Focus
	field modbusEditField
}

// NewModbusEditModel creates a new Modbus configuration edit modal.
func NewModbusEditModel(ctx context.Context, svc *shelly.Service) ModbusEditModel {
	return ModbusEditModel{
		ctx:    ctx,
		svc:    svc,
		styles: editmodal.DefaultStyles().WithLabelWidth(14),
	}
}

// Show displays the edit modal with the given device and Modbus state.
func (m ModbusEditModel) Show(device string, modbus *shelly.TUIModbusStatus) (ModbusEditModel, tea.Cmd) {
	m.device = device
	m.visible = true
	m.saving = false
	m.err = nil
	m.field = modbusFieldEnable

	if modbus != nil {
		m.enabled = modbus.Enabled
		m.pendingEnabled = modbus.Enabled
	}

	return m, nil
}

// Hide hides the edit modal.
func (m ModbusEditModel) Hide() ModbusEditModel {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m ModbusEditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m ModbusEditModel) SetSize(width, height int) ModbusEditModel {
	m.width = width
	m.height = height
	return m
}

// Update handles messages.
func (m ModbusEditModel) Update(msg tea.Msg) (ModbusEditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m ModbusEditModel) handleMessage(msg tea.Msg) (ModbusEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - update state and close modal
		m.enabled = m.pendingEnabled
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		if !m.saving {
			m.pendingEnabled = !m.pendingEnabled
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m ModbusEditModel) handleNavigation(msg messages.NavigationMsg) (ModbusEditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		if m.field > 0 {
			m.field--
		}
	case messages.NavDown:
		if int(m.field) < modbusFieldCount-1 {
			m.field++
		}
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m ModbusEditModel) handleKey(msg tea.KeyPressMsg) (ModbusEditModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEsc, keyconst.KeyCtrlOpenBracket:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter, keyconst.KeyCtrlS:
		return m.save()

	case "t", keyconst.KeySpace:
		if !m.saving && m.field == modbusFieldEnable {
			m.pendingEnabled = !m.pendingEnabled
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if int(m.field) < modbusFieldCount-1 {
			m.field++
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if m.field > 0 {
			m.field--
		}
		return m, nil
	}

	return m, nil
}

func (m ModbusEditModel) save() (ModbusEditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	// Check if anything changed
	if m.pendingEnabled == m.enabled {
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd()
}

func (m ModbusEditModel) createSaveCmd() tea.Cmd {
	newEnabled := m.pendingEnabled
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetModbusConfig(ctx, m.device, newEnabled)
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

// View renders the edit modal.
func (m ModbusEditModel) View() string {
	if !m.visible {
		return ""
	}

	footer := m.buildFooter()
	r := rendering.NewModal(m.width, m.height, "Modbus Configuration", footer)

	var content strings.Builder

	// Status summary
	content.WriteString(m.renderStatus())
	content.WriteString("\n\n")

	// Enable toggle
	content.WriteString(m.renderEnableToggle())

	// Change indicator
	if m.pendingEnabled != m.enabled {
		content.WriteString("\n")
		content.WriteString(m.renderChangeIndicator())
	}

	// Port info (when enabled or will be enabled)
	if m.enabled || m.pendingEnabled {
		content.WriteString("\n\n")
		content.WriteString(m.renderPortInfo())
	}

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return r.SetContent(content.String()).Render()
}

func (m ModbusEditModel) buildFooter() string {
	if m.saving {
		return footerSaving
	}
	return "Space/t: Toggle | Enter: Save | j/k: Navigate | Esc: Cancel"
}

func (m ModbusEditModel) renderStatus() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Status:"))
	content.WriteString(" ")
	if m.enabled {
		content.WriteString(m.styles.StatusOn.Render("● Enabled"))
	} else {
		content.WriteString(m.styles.StatusOff.Render("○ Disabled"))
	}

	if m.enabled {
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Port:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Value.Render("502 (TCP)"))
	}

	return content.String()
}

func (m ModbusEditModel) renderEnableToggle() string {
	selected := m.field == modbusFieldEnable

	var value string
	if m.pendingEnabled {
		value = m.styles.StatusOn.Render("[●] ON ")
	} else {
		value = m.styles.StatusOff.Render("[ ] OFF")
	}

	return m.styles.RenderFieldRow(selected, "Enabled:", value)
}

func (m ModbusEditModel) renderChangeIndicator() string {
	var msg string
	if m.pendingEnabled {
		msg = "Will enable Modbus-TCP on port 502"
	} else {
		msg = "Will disable Modbus-TCP server"
	}
	return m.styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

func (m ModbusEditModel) renderPortInfo() string {
	var content strings.Builder

	content.WriteString(m.styles.LabelFocus.Render("Modbus-TCP Details"))
	content.WriteString("\n")
	content.WriteString("  " + m.styles.Label.Render("Protocol:  "))
	content.WriteString(m.styles.Value.Render("Modbus-TCP"))
	content.WriteString("\n")
	content.WriteString("  " + m.styles.Label.Render("Port:      "))
	content.WriteString(m.styles.Value.Render("502"))
	content.WriteString("\n")
	content.WriteString("  " + m.styles.Label.Render("Registers: "))
	content.WriteString(m.styles.Muted.Render("Auto-exposed per component"))

	return content.String()
}
