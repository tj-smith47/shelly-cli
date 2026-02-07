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
	editmodal.Base

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

	// Sending state for test packets
	sending bool
}

// NewLoRaEditModel creates a new LoRa configuration edit modal.
func NewLoRaEditModel(ctx context.Context, svc *shelly.Service) LoRaEditModel {
	return LoRaEditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(14),
		},
	}
}

// Show displays the edit modal with the given device and LoRa state.
func (m LoRaEditModel) Show(device string, lora *shelly.TUILoRaStatus) (LoRaEditModel, tea.Cmd) {
	m.Base.Show(device, loraFieldCount)
	m.sending = false

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
	m.Base.Hide()
	return m
}

// Visible returns whether the modal is visible.
func (m LoRaEditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m LoRaEditModel) SetSize(width, height int) LoRaEditModel {
	m.Base.SetSize(width, height)
	return m
}

// Update handles messages.
func (m LoRaEditModel) Update(msg tea.Msg) (LoRaEditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m LoRaEditModel) handleMessage(msg tea.Msg) (LoRaEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LoRaTestSendResultMsg:
		m.sending = false
		if msg.Err != nil {
			m.Err = msg.Err
			return m, nil
		}
		m.Err = nil
		return m, nil

	case messages.SaveResultMsg:
		saved, cmd := m.HandleSaveResult(msg)
		if saved {
			m.freq = m.pendingFreq
			m.bw = m.pendingBW
			m.dr = m.pendingDR
			m.txp = m.pendingTxP
		}
		return m, cmd

	case messages.NavigationMsg:
		action := m.HandleNavigation(msg)
		if action != editmodal.ActionNone {
			return m.applyAction(action)
		}
		// Handle left/right for value adjustment
		switch msg.Direction {
		case messages.NavLeft:
			return m.adjustFieldValue(-1), nil
		case messages.NavRight:
			return m.adjustFieldValue(1), nil
		default:
		}
		return m, nil

	case messages.ToggleEnableRequestMsg:
		// LoRa does not have an enable toggle
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

func (m LoRaEditModel) applyAction(action editmodal.KeyAction) (LoRaEditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionNone:
		return m, nil
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		if loraEditField(m.Cursor) == loraFieldTestSend {
			return m.sendTestPacket()
		}
		return m.save()
	case editmodal.ActionNavDown:
		if m.Cursor < loraFieldCount-1 {
			m.Cursor++
		}
		return m, nil
	case editmodal.ActionNavUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil
	case editmodal.ActionNext:
		if m.Cursor < loraFieldCount-1 {
			m.Cursor++
		}
		return m, nil
	case editmodal.ActionPrev:
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil
	}
	return m, nil
}

func (m LoRaEditModel) handleCustomKey(msg tea.KeyPressMsg) (LoRaEditModel, tea.Cmd) {
	switch msg.String() {
	case "T":
		return m.sendTestPacket()

	case "h", "left":
		return m.adjustFieldValue(-1), nil

	case "l", "right":
		return m.adjustFieldValue(1), nil

	case "j", keyconst.KeyDown:
		if m.Cursor < loraFieldCount-1 {
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

func (m LoRaEditModel) adjustFieldValue(delta int) LoRaEditModel {
	if m.Saving || m.sending {
		return m
	}

	switch loraEditField(m.Cursor) {
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
	if m.Saving || m.sending {
		return m, nil
	}

	// Check if anything changed
	if !m.hasChanges() {
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.StartSave()

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

	cmd := m.SaveCmd(func(ctx context.Context) error {
		return m.Svc.SetLoRaConfig(ctx, m.Device, cfg)
	})
	return m, cmd
}

func (m LoRaEditModel) sendTestPacket() (LoRaEditModel, tea.Cmd) {
	if m.Saving || m.sending {
		return m, nil
	}

	m.sending = true
	m.Err = nil

	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.Ctx, 30*time.Second)
		defer cancel()

		err := m.Svc.SendLoRaTestPacket(ctx, m.Device)
		return LoRaTestSendResultMsg{Err: err}
	}
}

// View renders the edit modal.
func (m LoRaEditModel) View() string {
	if !m.Visible() {
		return ""
	}

	footer := m.buildFooter()

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
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("LoRa Configuration", content.String(), footer)
}

func (m LoRaEditModel) buildFooter() string {
	if m.Saving {
		return footerSaving
	}
	if m.sending {
		return "Sending test packet..."
	}
	return "←/→: Adjust | Enter: Save | T: Test | j/k: Nav | Esc: Cancel"
}

func (m LoRaEditModel) renderStatus() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Status:"))
	content.WriteString(" ")
	content.WriteString(m.Styles.StatusOn.Render("● Active"))

	// Last RSSI/SNR if available
	if m.rssi != 0 {
		content.WriteString("\n")
		content.WriteString(m.Styles.Label.Render("Last RSSI:"))
		content.WriteString(" ")
		content.WriteString(m.Styles.Value.Render(fmt.Sprintf("%d dBm", m.rssi)))
		content.WriteString("  ")
		content.WriteString(m.Styles.Label.Render("SNR:"))
		content.WriteString(" ")
		content.WriteString(m.Styles.Value.Render(fmt.Sprintf("%.1f dB", m.snr)))
	}

	return content.String()
}

func (m LoRaEditModel) renderFreqField() string {
	selected := loraEditField(m.Cursor) == loraFieldFreq
	freqMHz := float64(m.pendingFreq) / 1000000

	value := m.Styles.Value.Render(fmt.Sprintf("%.2f MHz", freqMHz))
	if selected {
		value += m.Styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingFreq != m.freq {
		value += m.Styles.Warning.Render(" *")
	}

	return m.Styles.RenderFieldRow(selected, "Frequency:", value)
}

func (m LoRaEditModel) renderBWField() string {
	selected := loraEditField(m.Cursor) == loraFieldBW

	value := m.Styles.Value.Render(fmt.Sprintf("%d kHz", m.pendingBW))
	if selected {
		value += m.Styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingBW != m.bw {
		value += m.Styles.Warning.Render(" *")
	}

	return m.Styles.RenderFieldRow(selected, "Bandwidth:", value)
}

func (m LoRaEditModel) renderDRField() string {
	selected := loraEditField(m.Cursor) == loraFieldDR

	value := m.Styles.Value.Render(fmt.Sprintf("SF%d", m.pendingDR))
	if selected {
		value += m.Styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingDR != m.dr {
		value += m.Styles.Warning.Render(" *")
	}

	return m.Styles.RenderFieldRow(selected, "Data Rate:", value)
}

func (m LoRaEditModel) renderTxPField() string {
	selected := loraEditField(m.Cursor) == loraFieldTxP

	value := m.Styles.Value.Render(fmt.Sprintf("%d dBm", m.pendingTxP))
	if selected {
		value += m.Styles.Muted.Render("  ◀ ▶")
	}
	if m.pendingTxP != m.txp {
		value += m.Styles.Warning.Render(" *")
	}

	return m.Styles.RenderFieldRow(selected, "TX Power:", value)
}

func (m LoRaEditModel) renderChangeIndicator() string {
	return m.Styles.Warning.Render("  ⚡ Configuration changed — Enter to save")
}

func (m LoRaEditModel) renderTestSendButton() string {
	selected := loraEditField(m.Cursor) == loraFieldTestSend

	selector := m.Styles.RenderSelector(selected)
	label := "Send Test Packet"
	if m.sending {
		label = "Sending..."
	}
	if selected {
		return selector + m.Styles.ButtonFocus.Render(label)
	}
	return selector + m.Styles.Button.Render(label)
}
