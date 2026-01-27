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

// LoRaTestSendResultMsg signals that a LoRa test packet send completed.
type LoRaTestSendResultMsg struct {
	Err error
}

// loraEditField identifies focusable fields in the LoRa edit modal.
type loraEditField int

const (
	loraFieldFreq     loraEditField = iota // Frequency field
	loraFieldBW                            // Bandwidth field
	loraFieldDR                            // Data Rate (Spreading Factor) field
	loraFieldTxP                           // Transmit Power field
	loraFieldTestSend                      // Send Test Packet button
)

// loraFieldCount is the total number of focusable fields.
const loraFieldCount = 5

// LoRa frequency step: 100 kHz = 100000 Hz.
const loraFreqStep int64 = 100000

// LoRa frequency limits (Hz) — covers common LoRa bands.
const (
	loraFreqMin int64 = 433000000 // 433 MHz (ISM)
	loraFreqMax int64 = 928000000 // 928 MHz (US915 top)
)

// Bandwidth values supported by Shelly LoRa add-on.
var loraBandwidths = [2]int{125, 250}

// LoRaEditModel represents the LoRa configuration edit modal.
type LoRaEditModel struct {
	ctx    context.Context
	svc    *shelly.Service
	device string

	visible bool
	saving  bool
	sending bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

	// Current state (from server)
	freq int64
	bw   int
	dr   int
	txp  int
	rssi int
	snr  float64

	// Pending changes
	pendingFreq int64
	pendingBW   int
	pendingDR   int
	pendingTxP  int

	// Focus
	field loraEditField
}

// NewLoRaEditModel creates a new LoRa configuration edit modal.
func NewLoRaEditModel(ctx context.Context, svc *shelly.Service) LoRaEditModel {
	return LoRaEditModel{
		ctx:    ctx,
		svc:    svc,
		styles: editmodal.DefaultStyles().WithLabelWidth(14),
	}
}

// Show displays the edit modal with the given device and LoRa state.
func (m LoRaEditModel) Show(device string, lora *shelly.TUILoRaStatus) (LoRaEditModel, tea.Cmd) {
	m.device = device
	m.visible = true
	m.saving = false
	m.sending = false
	m.err = nil
	m.field = loraFieldFreq

	if lora != nil {
		m.freq = lora.Frequency
		m.bw = lora.Bandwidth
		m.dr = lora.DataRate
		m.txp = lora.TxPower
		m.rssi = lora.RSSI
		m.snr = lora.SNR

		m.pendingFreq = lora.Frequency
		m.pendingBW = lora.Bandwidth
		m.pendingDR = lora.DataRate
		m.pendingTxP = lora.TxPower
	}

	return m, nil
}

// Hide hides the edit modal.
func (m LoRaEditModel) Hide() LoRaEditModel {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m LoRaEditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m LoRaEditModel) SetSize(width, height int) LoRaEditModel {
	m.width = width
	m.height = height
	return m
}

// Update handles messages.
func (m LoRaEditModel) Update(msg tea.Msg) (LoRaEditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m LoRaEditModel) handleMessage(msg tea.Msg) (LoRaEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - update state and close modal
		m.freq = m.pendingFreq
		m.bw = m.pendingBW
		m.dr = m.pendingDR
		m.txp = m.pendingTxP
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case LoRaTestSendResultMsg:
		m.sending = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.err = nil
		return m, nil

	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m LoRaEditModel) handleNavigation(msg messages.NavigationMsg) (LoRaEditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		if m.field > 0 {
			m.field--
		}
	case messages.NavDown:
		if int(m.field) < loraFieldCount-1 {
			m.field++
		}
	case messages.NavLeft:
		return m.adjustFieldValue(-1), nil
	case messages.NavRight:
		return m.adjustFieldValue(1), nil
	case messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m LoRaEditModel) handleKey(msg tea.KeyPressMsg) (LoRaEditModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEsc, keyconst.KeyCtrlOpenBracket:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter, keyconst.KeyCtrlS:
		if m.field == loraFieldTestSend {
			return m.sendTestPacket()
		}
		return m.save()

	case "T":
		return m.sendTestPacket()

	case "h", "left":
		return m.adjustFieldValue(-1), nil

	case "l", "right":
		return m.adjustFieldValue(1), nil

	case "j", keyconst.KeyDown:
		if int(m.field) < loraFieldCount-1 {
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

func (m LoRaEditModel) adjustFieldValue(delta int) LoRaEditModel {
	if m.saving || m.sending {
		return m
	}

	switch m.field {
	case loraFieldFreq:
		newFreq := m.pendingFreq + int64(delta)*loraFreqStep
		if newFreq >= loraFreqMin && newFreq <= loraFreqMax {
			m.pendingFreq = newFreq
		}
	case loraFieldBW:
		// Toggle between 125 and 250
		if m.pendingBW == loraBandwidths[0] {
			m.pendingBW = loraBandwidths[1]
		} else {
			m.pendingBW = loraBandwidths[0]
		}
	case loraFieldDR:
		newDR := m.pendingDR + delta
		if newDR >= 7 && newDR <= 12 {
			m.pendingDR = newDR
		}
	case loraFieldTxP:
		newTxP := m.pendingTxP + delta
		if newTxP >= 0 && newTxP <= 14 {
			m.pendingTxP = newTxP
		}
	case loraFieldTestSend:
		// No value adjustment for the button
	}

	return m
}

func (m LoRaEditModel) hasChanges() bool {
	return m.pendingFreq != m.freq ||
		m.pendingBW != m.bw ||
		m.pendingDR != m.dr ||
		m.pendingTxP != m.txp
}

func (m LoRaEditModel) save() (LoRaEditModel, tea.Cmd) {
	if m.saving || m.sending {
		return m, nil
	}

	// Check if anything changed
	if !m.hasChanges() {
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd()
}

func (m LoRaEditModel) createSaveCmd() tea.Cmd {
	cfg := make(map[string]any)
	if m.pendingFreq != m.freq {
		cfg["freq"] = m.pendingFreq
	}
	if m.pendingBW != m.bw {
		cfg["bw"] = m.pendingBW
	}
	if m.pendingDR != m.dr {
		cfg["dr"] = m.pendingDR
	}
	if m.pendingTxP != m.txp {
		cfg["txp"] = m.pendingTxP
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetLoRaConfig(ctx, m.device, cfg)
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

func (m LoRaEditModel) sendTestPacket() (LoRaEditModel, tea.Cmd) {
	if m.saving || m.sending {
		return m, nil
	}

	m.sending = true
	m.err = nil

	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SendLoRaTestPacket(ctx, m.device)
		return LoRaTestSendResultMsg{Err: err}
	}
}

// View renders the edit modal.
func (m LoRaEditModel) View() string {
	if !m.visible {
		return ""
	}

	footer := m.buildFooter()
	r := rendering.NewModal(m.width, m.height, "LoRa Configuration", footer)

	var content strings.Builder

	// Status summary
	content.WriteString(m.renderStatus())
	content.WriteString("\n\n")

	// Editable fields
	content.WriteString(m.renderFreqField())
	content.WriteString("\n")
	content.WriteString(m.renderBWField())
	content.WriteString("\n")
	content.WriteString(m.renderDRField())
	content.WriteString("\n")
	content.WriteString(m.renderTxPField())

	// Change indicator
	if m.hasChanges() {
		content.WriteString("\n")
		content.WriteString(m.renderChangeIndicator())
	}

	// Test send button
	content.WriteString("\n\n")
	content.WriteString(m.renderTestSendButton())

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return r.SetContent(content.String()).Render()
}

func (m LoRaEditModel) buildFooter() string {
	if m.saving {
		return footerSaving
	}
	if m.sending {
		return "Sending test packet..."
	}
	return "←/→: Adjust | Enter: Save | T: Test | j/k: Nav | Esc: Cancel"
}

func (m LoRaEditModel) renderStatus() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Status:"))
	content.WriteString(" ")
	content.WriteString(m.styles.StatusOn.Render("● Active"))

	// Last RSSI/SNR if available
	if m.rssi != 0 {
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Last RSSI:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d dBm", m.rssi)))
		content.WriteString("  ")
		content.WriteString(m.styles.Label.Render("SNR:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%.1f dB", m.snr)))
	}

	return content.String()
}

func (m LoRaEditModel) renderFreqField() string {
	selected := m.field == loraFieldFreq
	freqMHz := float64(m.pendingFreq) / 1000000

	value := m.styles.Value.Render(fmt.Sprintf("%.2f MHz", freqMHz))
	if selected {
		value += m.styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingFreq != m.freq {
		value += m.styles.Warning.Render(" *")
	}

	return m.styles.RenderFieldRow(selected, "Frequency:", value)
}

func (m LoRaEditModel) renderBWField() string {
	selected := m.field == loraFieldBW

	value := m.styles.Value.Render(fmt.Sprintf("%d kHz", m.pendingBW))
	if selected {
		value += m.styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingBW != m.bw {
		value += m.styles.Warning.Render(" *")
	}

	return m.styles.RenderFieldRow(selected, "Bandwidth:", value)
}

func (m LoRaEditModel) renderDRField() string {
	selected := m.field == loraFieldDR

	value := m.styles.Value.Render(fmt.Sprintf("SF%d", m.pendingDR))
	if selected {
		value += m.styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingDR != m.dr {
		value += m.styles.Warning.Render(" *")
	}

	return m.styles.RenderFieldRow(selected, "Data Rate:", value)
}

func (m LoRaEditModel) renderTxPField() string {
	selected := m.field == loraFieldTxP

	value := m.styles.Value.Render(fmt.Sprintf("%d dBm", m.pendingTxP))
	if selected {
		value += m.styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingTxP != m.txp {
		value += m.styles.Warning.Render(" *")
	}

	return m.styles.RenderFieldRow(selected, "TX Power:", value)
}

func (m LoRaEditModel) renderChangeIndicator() string {
	return m.styles.Warning.Render("  ⚡ Configuration changed — Enter to save")
}

func (m LoRaEditModel) renderTestSendButton() string {
	selected := m.field == loraFieldTestSend

	selector := m.styles.RenderSelector(selected)
	label := "Send Test Packet"
	if m.sending {
		label = "Sending..."
	}
	if selected {
		return selector + m.styles.ButtonFocus.Render(label)
	}
	return selector + m.styles.Button.Render(label)
}
