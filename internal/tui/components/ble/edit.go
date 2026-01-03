// Package ble provides TUI components for managing device Bluetooth settings.
package ble

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// EditField represents a field in the BLE edit form.
type EditField int

// Edit field constants.
const (
	EditFieldEnable EditField = iota
	EditFieldRPC
	EditFieldObserver
	EditFieldCount
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the BLE edit modal.
type EditModel struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	visible bool
	cursor  EditField
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

	// Original config for comparison
	original *shelly.BLEConfig

	// Form inputs
	enableToggle   form.Toggle
	rpcToggle      form.Toggle
	observerToggle form.Toggle
}

// NewEditModel creates a new BLE edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	enableToggle := form.NewToggle(
		form.WithToggleLabel("Enable"),
		form.WithToggleHelp("Enable or disable Bluetooth on this device"),
	)

	rpcToggle := form.NewToggle(
		form.WithToggleLabel("RPC"),
		form.WithToggleHelp("Allow RPC commands via Bluetooth"),
	)

	observerToggle := form.NewToggle(
		form.WithToggleLabel("Observer"),
		form.WithToggleHelp("Receive broadcasts from BLU sensors"),
	)

	return EditModel{
		ctx:            ctx,
		svc:            svc,
		styles:         editmodal.DefaultStyles(),
		enableToggle:   enableToggle,
		rpcToggle:      rpcToggle,
		observerToggle: observerToggle,
	}
}

// Show displays the edit modal for BLE configuration.
func (m EditModel) Show(device string, config *shelly.BLEConfig) (EditModel, tea.Cmd) {
	m.device = device
	m.visible = true
	m.cursor = EditFieldEnable
	m.saving = false
	m.err = nil
	m.original = config

	// Blur all inputs first
	m = m.blurAllInputs()

	// Focus enable toggle
	m.enableToggle = m.enableToggle.Focus()

	// Populate from config
	if config != nil {
		m.enableToggle = m.enableToggle.SetValue(config.Enable)
		m.rpcToggle = m.rpcToggle.SetValue(config.RPCEnabled)
		m.observerToggle = m.observerToggle.SetValue(config.ObserverMode)
	}

	return m, func() tea.Msg { return EditOpenedMsg{} }
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case EditSaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Save successful, close the modal
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "ctrl+[":
		if m.saving {
			return m, nil
		}
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case "tab", "down", "j":
		if m.saving {
			return m, nil
		}
		m = m.nextField()
		return m, nil

	case "shift+tab", "up", "k":
		if m.saving {
			return m, nil
		}
		m = m.prevField()
		return m, nil

	case "enter":
		if m.saving {
			return m, nil
		}
		return m.save()

	case "ctrl+s":
		if m.saving {
			return m, nil
		}
		return m.save()

	case " ":
		// Space toggles the current toggle
		return m.handleSpace()
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleSpace() (EditModel, tea.Cmd) {
	switch m.cursor {
	case EditFieldEnable:
		m.enableToggle = m.enableToggle.Toggle()
	case EditFieldRPC:
		m.rpcToggle = m.rpcToggle.Toggle()
	case EditFieldObserver:
		m.observerToggle = m.observerToggle.Toggle()
	case EditFieldCount:
		// Sentinel, no action
	}
	return m, nil
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case EditFieldEnable:
		m.enableToggle, cmd = m.enableToggle.Update(msg)
	case EditFieldRPC:
		m.rpcToggle, cmd = m.rpcToggle.Update(msg)
	case EditFieldObserver:
		m.observerToggle, cmd = m.observerToggle.Update(msg)
	case EditFieldCount:
		// Sentinel, no input to update
	}

	return m, cmd
}

func (m EditModel) nextField() EditModel {
	m = m.blurCurrentField()
	m.cursor++
	if m.cursor >= EditFieldCount {
		m.cursor = 0
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()
	m.cursor--
	if m.cursor < 0 {
		m.cursor = EditFieldCount - 1
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
	case EditFieldEnable:
		m.enableToggle = m.enableToggle.Blur()
	case EditFieldRPC:
		m.rpcToggle = m.rpcToggle.Blur()
	case EditFieldObserver:
		m.observerToggle = m.observerToggle.Blur()
	case EditFieldCount:
		// Sentinel, no field to blur
	}
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch m.cursor {
	case EditFieldEnable:
		m.enableToggle = m.enableToggle.Focus()
	case EditFieldRPC:
		m.rpcToggle = m.rpcToggle.Focus()
	case EditFieldObserver:
		m.observerToggle = m.observerToggle.Focus()
	case EditFieldCount:
		// Sentinel, no field to focus
	}
	return m
}

func (m EditModel) blurAllInputs() EditModel {
	m.enableToggle = m.enableToggle.Blur()
	m.rpcToggle = m.rpcToggle.Blur()
	m.observerToggle = m.observerToggle.Blur()
	return m
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	m.err = nil

	// Get values from toggles
	enable := m.enableToggle.Value()
	rpc := m.rpcToggle.Value()
	observer := m.observerToggle.Value()

	m.saving = true
	return m, m.createSaveCmd(enable, rpc, observer)
}

func (m EditModel) createSaveCmd(enable, rpc, observer bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetBLEConfig(ctx, m.device, &enable, &rpc, &observer)
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Tab/j/k: Navigate | Space: Toggle | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Saving..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "Bluetooth Settings", footer)

	// Build content
	var content strings.Builder

	// Info text
	content.WriteString(m.styles.Info.Render("Configure Bluetooth options for this device"))
	content.WriteString("\n\n")

	// Form fields
	content.WriteString(m.renderFormFields())

	// Warning about observer mode
	if m.observerToggle.Value() && !m.enableToggle.Value() {
		content.WriteString("\n")
		content.WriteString(m.styles.Warning.Render("⚠ Observer requires Bluetooth enabled"))
	}

	// Error message
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
	}

	return r.SetContent(content.String()).Render()
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Enable
	content.WriteString(m.renderField(EditFieldEnable, "Bluetooth:", m.enableToggle.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Info.Render("    Main Bluetooth on/off switch"))
	content.WriteString("\n\n")

	// RPC
	content.WriteString(m.renderField(EditFieldRPC, "RPC Service:", m.rpcToggle.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Info.Render("    Accept RPC commands via Bluetooth"))
	content.WriteString("\n\n")

	// Observer
	content.WriteString(m.renderField(EditFieldObserver, "Observer:", m.observerToggle.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Info.Render("    Receive BLU sensor broadcasts"))

	return content.String()
}

func (m EditModel) renderField(field EditField, label, value string) string {
	return m.styles.RenderLabel(label, m.cursor == field) + " " + value
}

// Device returns the current device.
func (m EditModel) Device() string {
	return m.device
}
