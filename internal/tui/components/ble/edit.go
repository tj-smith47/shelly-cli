// Package ble provides TUI components for managing device Bluetooth settings.
package ble

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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
	editmodal.Base

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
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(14),
		},
		enableToggle:   enableToggle,
		rpcToggle:      rpcToggle,
		observerToggle: observerToggle,
	}
}

// Show displays the edit modal for BLE configuration.
func (m EditModel) Show(device string, config *shelly.BLEConfig) (EditModel, tea.Cmd) {
	m.Base.Show(device, int(EditFieldCount))
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
	m.Base.Hide()
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.Base.SetSize(width, height)
	return m
}

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m EditModel) handleMessage(msg tea.Msg) (EditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case EditSaveResultMsg:
		_, cmd := m.HandleSaveResult(msg)
		return m, cmd

	case messages.NavigationMsg:
		if m.Saving {
			return m, nil
		}
		action := m.HandleNavigation(msg)
		return m.applyAction(action)
	case messages.ToggleEnableRequestMsg:
		return m.handleSpace()
	case tea.KeyPressMsg:
		action := m.HandleKey(msg)
		if action != editmodal.ActionNone {
			return m.applyAction(action)
		}
		return m.handleCustomKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) applyAction(action editmodal.KeyAction) (EditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		return m.save()
	case editmodal.ActionNext, editmodal.ActionNavDown:
		m = m.moveFocus(m.NextField())
		return m, nil
	case editmodal.ActionPrev, editmodal.ActionNavUp:
		m = m.moveFocus(m.PrevField())
		return m, nil
	case editmodal.ActionNone:
		// No action to take
	}
	return m, nil
}

func (m EditModel) handleCustomKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	switch msg.String() {
	case "t", keyconst.KeySpace:
		if !m.Saving {
			return m.handleSpace()
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if !m.Saving {
			m = m.moveFocus(m.NextField())
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if !m.Saving {
			m = m.moveFocus(m.PrevField())
		}
		return m, nil
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleSpace() (EditModel, tea.Cmd) {
	switch EditField(m.Cursor) {
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

	switch EditField(m.Cursor) {
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

// moveFocus blurs the old field and focuses the new one.
func (m EditModel) moveFocus(oldCursor, newCursor int) EditModel {
	m = m.blurField(EditField(oldCursor))
	m = m.focusField(EditField(newCursor))
	return m
}

func (m EditModel) blurField(field EditField) EditModel {
	switch field {
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

func (m EditModel) focusField(field EditField) EditModel {
	switch field {
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
	if m.Saving {
		return m, nil
	}

	// Get values from toggles
	enable := m.enableToggle.Value()
	rpc := m.rpcToggle.Value()
	observer := m.observerToggle.Value()

	m.StartSave()

	device := m.Base.Device
	cmd := m.SaveCmd(func(ctx context.Context) error {
		return m.Svc.SetBLEConfig(ctx, device, &enable, &rpc, &observer)
	})
	return m, cmd
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.Visible() {
		return ""
	}

	footer := m.RenderSavingFooter("Tab/j/k: Navigate | Space: Toggle | Ctrl+S: Save | Esc: Cancel")

	// Build content
	var content strings.Builder

	// Info text
	content.WriteString(m.Styles.Info.Render("Configure Bluetooth options for this device"))
	content.WriteString("\n\n")

	// Form fields
	content.WriteString(m.renderFormFields())

	// Warning about observer mode
	if m.observerToggle.Value() && !m.enableToggle.Value() {
		content.WriteString("\n")
		content.WriteString(m.Styles.Warning.Render("⚠ Observer requires Bluetooth enabled"))
	}

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("Bluetooth Settings", content.String(), footer)
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Enable
	content.WriteString(m.RenderField(int(EditFieldEnable), "Bluetooth:", m.enableToggle.View()))
	content.WriteString("\n")
	content.WriteString(m.Styles.Info.Render("    Main Bluetooth on/off switch"))
	content.WriteString("\n\n")

	// RPC
	content.WriteString(m.RenderField(int(EditFieldRPC), "RPC Service:", m.rpcToggle.View()))
	content.WriteString("\n")
	content.WriteString(m.Styles.Info.Render("    Accept RPC commands via Bluetooth"))
	content.WriteString("\n\n")

	// Observer
	content.WriteString(m.RenderField(int(EditFieldObserver), "Observer:", m.observerToggle.View()))
	content.WriteString("\n")
	content.WriteString(m.Styles.Info.Render("    Receive BLU sensor broadcasts"))

	return content.String()
}

// Device returns the current device.
func (m EditModel) Device() string {
	return m.Base.Device
}
