// Package ble provides TUI components for managing device Bluetooth settings.
package ble

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
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

// EditSaveResultMsg signals a save operation completed.
type EditSaveResultMsg struct {
	Err error
}

// EditOpenedMsg signals the edit modal was opened.
type EditOpenedMsg struct{}

// EditClosedMsg signals the edit modal was closed.
type EditClosedMsg struct {
	Saved bool
}

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
	styles  EditStyles

	// Original config for comparison
	original *shelly.BLEConfig

	// Form inputs
	enableToggle   form.Toggle
	rpcToggle      form.Toggle
	observerToggle form.Toggle
}

// EditStyles holds styles for the BLE edit modal.
type EditStyles struct {
	Modal      lipgloss.Style
	Title      lipgloss.Style
	Label      lipgloss.Style
	LabelFocus lipgloss.Style
	Error      lipgloss.Style
	Help       lipgloss.Style
	Selector   lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style
}

// DefaultEditStyles returns the default edit modal styles.
func DefaultEditStyles() EditStyles {
	colors := theme.GetSemanticColors()
	return EditStyles{
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Background(colors.Background).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(14),
		LabelFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Width(14),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selector: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Info: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
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
		styles:         DefaultEditStyles(),
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
	case "esc":
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
		return EditSaveResultMsg{Err: err}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(m.styles.Title.Render("Bluetooth Settings"))
	content.WriteString("\n\n")

	// Info text
	content.WriteString(m.styles.Info.Render("Configure Bluetooth options for this device"))
	content.WriteString("\n\n")

	// Form fields
	content.WriteString(m.renderFormFields())
	content.WriteString("\n\n")

	// Warning about observer mode
	if m.observerToggle.Value() && !m.enableToggle.Value() {
		content.WriteString(m.styles.Warning.Render("⚠ Observer requires Bluetooth enabled"))
		content.WriteString("\n\n")
	}

	// Error message
	if m.err != nil {
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
		content.WriteString("\n\n")
	}

	// Saving indicator
	if m.saving {
		content.WriteString(m.styles.Info.Render("Saving..."))
		content.WriteString("\n\n")
	}

	// Help
	content.WriteString(m.styles.Help.Render("Tab/j/k: navigate • Space: toggle • Enter/Ctrl+S: save • Esc: cancel"))

	// Calculate modal dimensions
	modalWidth := 55
	if m.width > 0 && m.width < modalWidth+10 {
		modalWidth = m.width - 10
	}

	modal := m.styles.Modal.Width(modalWidth).Render(content.String())

	// Center the modal
	return m.centerModal(modal)
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
	labelStyle := m.styles.Label
	if m.cursor == field {
		labelStyle = m.styles.LabelFocus
	}
	return labelStyle.Render(label) + " " + value
}

func (m EditModel) centerModal(modal string) string {
	if m.width == 0 || m.height == 0 {
		return modal
	}

	lines := strings.Split(modal, "\n")
	modalHeight := len(lines)
	modalWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > modalWidth {
			modalWidth = lipgloss.Width(line)
		}
	}

	// Calculate padding
	topPadding := (m.height - modalHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}
	leftPadding := (m.width - modalWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Build centered output
	var result strings.Builder
	for range topPadding {
		result.WriteString("\n")
	}
	for _, line := range lines {
		result.WriteString(strings.Repeat(" ", leftPadding))
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// Device returns the current device.
func (m EditModel) Device() string {
	return m.device
}
