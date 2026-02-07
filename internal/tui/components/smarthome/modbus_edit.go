// Package smarthome provides TUI components for managing smart home protocol settings.
package smarthome

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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
	editmodal.Base

	// Modbus state
	enabled bool

	// Pending changes
	pendingEnabled bool
}

// NewModbusEditModel creates a new Modbus configuration edit modal.
func NewModbusEditModel(ctx context.Context, svc *shelly.Service) ModbusEditModel {
	return ModbusEditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(14),
		},
	}
}

// Show displays the edit modal with the given device and Modbus state.
func (m ModbusEditModel) Show(device string, modbus *shelly.TUIModbusStatus) (ModbusEditModel, tea.Cmd) {
	m.Base.Show(device, modbusFieldCount)

	if modbus != nil {
		m.enabled = modbus.Enabled
		m.pendingEnabled = modbus.Enabled
	}

	return m, nil
}

// Hide hides the edit modal.
func (m ModbusEditModel) Hide() ModbusEditModel {
	m.Base.Hide()
	return m
}

// Visible returns whether the modal is visible.
func (m ModbusEditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m ModbusEditModel) SetSize(width, height int) ModbusEditModel {
	m.Base.SetSize(width, height)
	return m
}

// Update handles messages.
func (m ModbusEditModel) Update(msg tea.Msg) (ModbusEditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m ModbusEditModel) handleMessage(msg tea.Msg) (ModbusEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		saved, cmd := m.HandleSaveResult(msg)
		if saved {
			m.enabled = m.pendingEnabled
		}
		return m, cmd

	case messages.NavigationMsg:
		action := m.HandleNavigation(msg)
		return m.applyAction(action)
	case messages.ToggleEnableRequestMsg:
		if !m.Saving {
			m.pendingEnabled = !m.pendingEnabled
		}
		return m, nil
	case tea.KeyPressMsg:
		action := m.HandleKey(msg)
		if action != editmodal.ActionNone {
			return m.applyAction(action)
		}
		return m.handleCustomKey(msg)
	}

	return m, nil
}

func (m ModbusEditModel) applyAction(action editmodal.KeyAction) (ModbusEditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionNone:
		return m, nil
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		return m.save()
	case editmodal.ActionNext, editmodal.ActionNavDown:
		// Only 1 field, no navigation needed
		return m, nil
	case editmodal.ActionPrev, editmodal.ActionNavUp:
		return m, nil
	}
	return m, nil
}

func (m ModbusEditModel) handleCustomKey(msg tea.KeyPressMsg) (ModbusEditModel, tea.Cmd) {
	switch msg.String() {
	case "t", keyconst.KeySpace:
		if !m.Saving && modbusEditField(m.Cursor) == modbusFieldEnable {
			m.pendingEnabled = !m.pendingEnabled
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if m.Cursor < modbusFieldCount-1 {
			m.Cursor++
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil
	}

	return m, nil
}

func (m ModbusEditModel) save() (ModbusEditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	// Check if anything changed
	if m.pendingEnabled == m.enabled {
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.StartSave()
	newEnabled := m.pendingEnabled

	cmd := m.SaveCmd(func(ctx context.Context) error {
		return m.Svc.SetModbusConfig(ctx, m.Device, newEnabled)
	})
	return m, cmd
}

// View renders the edit modal.
func (m ModbusEditModel) View() string {
	if !m.Visible() {
		return ""
	}

	footer := m.RenderSavingFooter("Space/t: Toggle | Enter: Save | j/k: Navigate | Esc: Cancel")

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
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("Modbus Configuration", content.String(), footer)
}

func (m ModbusEditModel) renderStatus() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Status:"))
	content.WriteString(" ")
	if m.enabled {
		content.WriteString(m.Styles.StatusOn.Render("● Enabled"))
	} else {
		content.WriteString(m.Styles.StatusOff.Render("○ Disabled"))
	}

	if m.enabled {
		content.WriteString("\n")
		content.WriteString(m.Styles.Label.Render("Port:"))
		content.WriteString(" ")
		content.WriteString(m.Styles.Value.Render("502 (TCP)"))
	}

	return content.String()
}

func (m ModbusEditModel) renderEnableToggle() string {
	selected := modbusEditField(m.Cursor) == modbusFieldEnable

	var value string
	if m.pendingEnabled {
		value = m.Styles.StatusOn.Render("[●] ON ")
	} else {
		value = m.Styles.StatusOff.Render("[ ] OFF")
	}

	return m.Styles.RenderFieldRow(selected, "Enabled:", value)
}

func (m ModbusEditModel) renderChangeIndicator() string {
	var msg string
	if m.pendingEnabled {
		msg = "Will enable Modbus-TCP on port 502"
	} else {
		msg = "Will disable Modbus-TCP server"
	}
	return m.Styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

func (m ModbusEditModel) renderPortInfo() string {
	var content strings.Builder

	content.WriteString(m.Styles.LabelFocus.Render("Modbus-TCP Details"))
	content.WriteString("\n")
	content.WriteString("  " + m.Styles.Label.Render("Protocol:  "))
	content.WriteString(m.Styles.Value.Render("Modbus-TCP"))
	content.WriteString("\n")
	content.WriteString("  " + m.Styles.Label.Render("Port:      "))
	content.WriteString(m.Styles.Value.Render("502"))
	content.WriteString("\n")
	content.WriteString("  " + m.Styles.Label.Render("Registers: "))
	content.WriteString(m.Styles.Muted.Render("Auto-exposed per component"))

	return content.String()
}
