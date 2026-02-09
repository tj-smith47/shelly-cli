// Package ble provides TUI components for managing device Bluetooth settings.
package ble

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/errorview"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// PairField represents a field in the BLE pairing form.
type PairField int

// Pair field constants.
const (
	PairFieldAddr PairField = iota
	PairFieldName
	PairFieldCount
)

// DeviceAddedMsg signals that a BTHome device was added.
type DeviceAddedMsg struct {
	Key string
	Err error
}

// PairOpenedMsg signals the pairing modal was opened.
type PairOpenedMsg struct{}

// PairClosedMsg signals the pairing modal was closed.
type PairClosedMsg struct {
	Added bool
}

// PairModel represents the BTHome device pairing modal.
type PairModel struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	visible bool
	cursor  PairField
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

	// Form inputs
	addrInput form.TextInput
	nameInput form.TextInput
}

// NewPairModel creates a new BTHome pairing modal.
func NewPairModel(ctx context.Context, svc *shelly.Service) PairModel {
	addrInput := form.NewTextInput(
		form.WithPlaceholder("AA:BB:CC:DD:EE:FF"),
		form.WithCharLimit(17),
	)

	nameInput := form.NewTextInput(
		form.WithPlaceholder("Living Room Sensor"),
		form.WithCharLimit(64),
	)

	return PairModel{
		ctx:       ctx,
		svc:       svc,
		styles:    editmodal.DefaultStyles().WithLabelWidth(12),
		addrInput: addrInput,
		nameInput: nameInput,
	}
}

// Show displays the pairing modal.
func (m PairModel) Show(device string) (PairModel, tea.Cmd) {
	m.device = device
	m.visible = true
	m.cursor = PairFieldAddr
	m.saving = false
	m.err = nil

	// Reset inputs
	m.addrInput = m.addrInput.SetValue("")
	m.nameInput = m.nameInput.SetValue("")

	// Blur all and focus first
	m = m.blurAllInputs()
	var focusCmd tea.Cmd
	m.addrInput, focusCmd = m.addrInput.Focus()

	return m, tea.Batch(focusCmd, func() tea.Msg { return PairOpenedMsg{} })
}

// Hide hides the pairing modal.
func (m PairModel) Hide() PairModel {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m PairModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m PairModel) SetSize(width, height int) PairModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m PairModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PairModel) Update(msg tea.Msg) (PairModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m PairModel) handleMessage(msg tea.Msg) (PairModel, tea.Cmd) {
	switch msg := msg.(type) {
	case DeviceAddedMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.visible = false
		return m, func() tea.Msg { return PairClosedMsg{Added: true} }

	case messages.NavigationMsg:
		if m.saving {
			return m, nil
		}
		return m.handleNavigation(msg)

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m PairModel) handleNavigation(msg messages.NavigationMsg) (PairModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		return m.prevField(), nil
	case messages.NavDown:
		return m.nextField(), nil
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this form
	}
	return m, nil
}

func (m PairModel) handleKey(msg tea.KeyPressMsg) (PairModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEsc, keyconst.KeyCtrlOpenBracket:
		if m.saving {
			return m, nil
		}
		m.visible = false
		return m, func() tea.Msg { return PairClosedMsg{Added: false} }

	case keyconst.KeyTab:
		if m.saving {
			return m, nil
		}
		return m.nextField(), nil

	case keyconst.KeyShiftTab:
		if m.saving {
			return m, nil
		}
		return m.prevField(), nil

	case keyconst.KeyEnter, keyconst.KeyCtrlS:
		if m.saving {
			return m, nil
		}
		return m.save()
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m PairModel) updateFocusedInput(msg tea.Msg) (PairModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case PairFieldAddr:
		m.addrInput, cmd = m.addrInput.Update(msg)
	case PairFieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case PairFieldCount:
		// Sentinel
	}

	return m, cmd
}

func (m PairModel) nextField() PairModel {
	m = m.blurCurrentField()
	m.cursor++
	if m.cursor >= PairFieldCount {
		m.cursor = 0
	}
	m = m.focusCurrentField()
	return m
}

func (m PairModel) prevField() PairModel {
	m = m.blurCurrentField()
	m.cursor--
	if m.cursor < 0 {
		m.cursor = PairFieldCount - 1
	}
	m = m.focusCurrentField()
	return m
}

func (m PairModel) blurCurrentField() PairModel {
	switch m.cursor {
	case PairFieldAddr:
		m.addrInput = m.addrInput.Blur()
	case PairFieldName:
		m.nameInput = m.nameInput.Blur()
	case PairFieldCount:
		// Sentinel
	}
	return m
}

func (m PairModel) focusCurrentField() PairModel {
	switch m.cursor {
	case PairFieldAddr:
		m.addrInput, _ = m.addrInput.Focus()
	case PairFieldName:
		m.nameInput, _ = m.nameInput.Focus()
	case PairFieldCount:
		// Sentinel
	}
	return m
}

func (m PairModel) blurAllInputs() PairModel {
	m.addrInput = m.addrInput.Blur()
	m.nameInput = m.nameInput.Blur()
	return m
}

func (m PairModel) save() (PairModel, tea.Cmd) {
	m.err = nil

	// Validate MAC address
	addr := strings.TrimSpace(m.addrInput.Value())
	if addr == "" {
		m.err = errMACRequired
		return m, nil
	}

	if !isValidMAC(addr) {
		m.err = errMACInvalid
		return m, nil
	}

	name := strings.TrimSpace(m.nameInput.Value())

	m.saving = true
	return m, m.createAddCmd(addr, name)
}

func (m PairModel) createAddCmd(addr, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		result, err := m.svc.Wireless().BTHomeAddDevice(ctx, m.device, addr, name)
		if err != nil {
			return DeviceAddedMsg{Err: err}
		}
		return DeviceAddedMsg{Key: result.Key}
	}
}

// View renders the pairing modal.
func (m PairModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Tab: Navigate | Ctrl+S: Add | Esc: Cancel"
	if m.saving {
		footer = "Adding device..."
	}

	r := rendering.NewModal(m.width, m.height, "Add BTHome Device", footer)

	var content strings.Builder

	// Info text
	content.WriteString(m.styles.Info.Render("Enter the MAC address of the BTHome device to pair."))
	content.WriteString("\n")
	content.WriteString(m.styles.Info.Render("Run discovery scan first to find nearby devices."))
	content.WriteString("\n\n")

	// Form fields
	content.WriteString(m.renderFormFields())

	// Error message
	if m.err != nil {
		content.WriteString("\n\n")
		content.WriteString(errorview.RenderInline(m.err))
	}

	return r.SetContent(content.String()).Render()
}

func (m PairModel) renderFormFields() string {
	var content strings.Builder

	// MAC Address
	content.WriteString(m.renderField(PairFieldAddr, "MAC Address:", m.addrInput.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Info.Render("    Format: AA:BB:CC:DD:EE:FF"))
	content.WriteString("\n\n")

	// Name (optional)
	content.WriteString(m.renderField(PairFieldName, "Name:", m.nameInput.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Info.Render("    Optional friendly name"))

	return content.String()
}

func (m PairModel) renderField(field PairField, label, value string) string {
	return m.styles.RenderLabel(label, m.cursor == field) + " " + value
}

// Device returns the current device.
func (m PairModel) Device() string {
	return m.device
}

// Validation errors.
var (
	errMACRequired = errors.New("MAC address is required")
	errMACInvalid  = errors.New("invalid MAC address format (use AA:BB:CC:DD:EE:FF)")
)

// macPattern matches MAC addresses in common formats.
var macPattern = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

// isValidMAC validates a MAC address format.
func isValidMAC(addr string) bool {
	return macPattern.MatchString(addr)
}
