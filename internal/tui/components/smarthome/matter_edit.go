// Package smarthome provides TUI components for managing smart home protocol settings.
package smarthome

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// MatterEditSaveResultMsg is an alias for the shared save result message.
type MatterEditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// MatterCodesLoadedMsg signals that commissioning codes were fetched.
type MatterCodesLoadedMsg struct {
	Codes model.CommissioningInfo
	Err   error
}

// MatterResetResultMsg signals that a Matter factory reset completed.
type MatterResetResultMsg struct {
	Err error
}

// MatterToggleResultMsg signals that a Matter enable/disable toggle completed.
type MatterToggleResultMsg struct {
	Enabled bool
	Err     error
}

// matterEditField identifies focusable fields in the edit modal.
type matterEditField int

const (
	matterFieldEnable matterEditField = iota // Enable/disable toggle
	matterFieldReset                         // Factory reset button
)

// MatterEditModel represents the Matter configuration edit modal.
type MatterEditModel struct {
	editmodal.Base

	// Matter state
	enabled        bool
	commissionable bool
	fabricsCount   int

	// Pending changes
	pendingEnabled bool

	// Commissioning codes
	codes        *model.CommissioningInfo
	loadingCodes bool

	// Reset confirmation
	pendingReset bool
	resetting    bool
}

// NewMatterEditModel creates a new Matter configuration edit modal.
func NewMatterEditModel(ctx context.Context, svc *shelly.Service) MatterEditModel {
	return MatterEditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(14),
		},
	}
}

// Show displays the edit modal with the given device and Matter state.
func (m MatterEditModel) Show(device string, matter *shelly.TUIMatterStatus) (MatterEditModel, tea.Cmd) {
	// Calculate field count: enable toggle + reset button (when enabled)
	fieldCount := 1 // Enable toggle always
	if matter != nil && matter.Enabled {
		fieldCount = 2 // Add reset button
	}

	m.Base.Show(device, fieldCount)
	m.resetting = false
	m.pendingReset = false
	m.codes = nil
	m.loadingCodes = false

	if matter != nil {
		m.enabled = matter.Enabled
		m.commissionable = matter.Commissionable
		m.fabricsCount = matter.FabricsCount
		m.pendingEnabled = matter.Enabled
	}

	// Fetch commissioning codes if enabled and commissionable
	var cmd tea.Cmd
	if matter != nil && matter.Enabled && matter.Commissionable {
		m.loadingCodes = true
		cmd = m.fetchCodes()
	}

	return m, cmd
}

// Hide hides the edit modal.
func (m MatterEditModel) Hide() MatterEditModel {
	m.Base.Hide()
	m.pendingReset = false
	return m
}

// Visible returns whether the modal is visible.
func (m MatterEditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m MatterEditModel) SetSize(width, height int) MatterEditModel {
	m.Base.SetSize(width, height)
	return m
}

// Update handles messages.
func (m MatterEditModel) Update(msg tea.Msg) (MatterEditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m MatterEditModel) handleMessage(msg tea.Msg) (MatterEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case MatterCodesLoadedMsg:
		m.loadingCodes = false
		if msg.Err != nil {
			// Codes not available, not a fatal error
			return m, nil
		}
		m.codes = &msg.Codes
		return m, nil

	case messages.SaveResultMsg:
		saved, cmd := m.HandleSaveResult(msg)
		if saved {
			m.enabled = m.pendingEnabled
		}
		return m, cmd

	case MatterResetResultMsg:
		m.resetting = false
		m.pendingReset = false
		if msg.Err != nil {
			m.Err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case messages.NavigationMsg:
		action := m.HandleNavigation(msg)
		return m.applyAction(action)
	case messages.ToggleEnableRequestMsg:
		if !m.Saving && !m.resetting {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingReset = false
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

func (m MatterEditModel) applyAction(action editmodal.KeyAction) (MatterEditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionNone:
		return m, nil
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		return m.handleSaveOrAction()
	case editmodal.ActionNext, editmodal.ActionNavDown:
		if m.Cursor < m.FieldCount-1 {
			m.Cursor++
			m.pendingReset = false
		}
		return m, nil
	case editmodal.ActionPrev, editmodal.ActionNavUp:
		if m.Cursor > 0 {
			m.Cursor--
			m.pendingReset = false
		}
		return m, nil
	}
	return m, nil
}

func (m MatterEditModel) handleCustomKey(msg tea.KeyPressMsg) (MatterEditModel, tea.Cmd) {
	switch msg.String() {
	case "t", keyconst.KeySpace:
		if !m.Saving && !m.resetting && matterEditField(m.Cursor) == matterFieldEnable {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingReset = false
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if m.Cursor < m.FieldCount-1 {
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

func (m MatterEditModel) handleSaveOrAction() (MatterEditModel, tea.Cmd) {
	if m.Saving || m.resetting {
		return m, nil
	}

	// If focused on reset button, handle reset confirmation
	if matterEditField(m.Cursor) == matterFieldReset {
		if m.pendingReset {
			// Second press - confirm reset
			m.resetting = true
			m.pendingReset = false
			m.Err = nil
			return m, m.createResetCmd()
		}
		// First press - request confirmation
		m.pendingReset = true
		return m, nil
	}

	// Save enable toggle change
	return m.save()
}

func (m MatterEditModel) save() (MatterEditModel, tea.Cmd) {
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
		if newEnabled {
			return m.Svc.Wireless().MatterEnable(ctx, m.Device)
		}
		return m.Svc.Wireless().MatterDisable(ctx, m.Device)
	})
	return m, cmd
}

func (m MatterEditModel) createResetCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.Ctx, 30*time.Second)
		defer cancel()

		err := m.Svc.Wireless().MatterReset(ctx, m.Device)
		return MatterResetResultMsg{Err: err}
	}
}

func (m MatterEditModel) fetchCodes() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.Ctx, 10*time.Second)
		defer cancel()

		info, err := m.Svc.Wireless().MatterGetCommissioningCode(ctx, m.Device)
		return MatterCodesLoadedMsg{Codes: info, Err: err}
	}
}

// View renders the edit modal.
func (m MatterEditModel) View() string {
	if !m.Visible() {
		return ""
	}

	footer := m.buildFooter()

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

	// Commissioning codes section (when enabled)
	if m.enabled || m.pendingEnabled {
		content.WriteString("\n\n")
		content.WriteString(m.renderCodes())
	}

	// Reset button (when enabled)
	if m.enabled {
		content.WriteString("\n\n")
		content.WriteString(m.renderResetButton())
	}

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("Matter Configuration", content.String(), footer)
}

func (m MatterEditModel) buildFooter() string {
	if m.Saving {
		return footerSaving
	}
	if m.resetting {
		return "Resetting Matter..."
	}
	if m.pendingReset {
		return "Press Enter again to confirm factory reset, Esc to cancel"
	}
	return "Space/t: Toggle | Enter: Save/Confirm | j/k: Navigate | Esc: Cancel"
}

func (m MatterEditModel) renderStatus() string {
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
		content.WriteString(m.Styles.Label.Render("Commission:"))
		content.WriteString(" ")
		if m.commissionable {
			content.WriteString(m.Styles.Warning.Render("Ready to pair"))
		} else {
			content.WriteString(m.Styles.Muted.Render("Already paired"))
		}

		content.WriteString("\n")
		content.WriteString(m.Styles.Label.Render("Fabrics:"))
		content.WriteString(" ")
		content.WriteString(m.Styles.Value.Render(fmt.Sprintf("%d", m.fabricsCount)))
	}

	return content.String()
}

func (m MatterEditModel) renderEnableToggle() string {
	selected := matterEditField(m.Cursor) == matterFieldEnable

	var value string
	if m.pendingEnabled {
		value = m.Styles.StatusOn.Render("[●] ON ")
	} else {
		value = m.Styles.StatusOff.Render("[ ] OFF")
	}

	return m.Styles.RenderFieldRow(selected, "Enabled:", value)
}

func (m MatterEditModel) renderChangeIndicator() string {
	var msg string
	if m.pendingEnabled {
		msg = "Will enable Matter"
	} else {
		msg = "Will disable Matter"
	}
	return m.Styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

func (m MatterEditModel) renderCodes() string {
	var content strings.Builder

	content.WriteString(m.Styles.LabelFocus.Render("Commissioning Codes"))
	content.WriteString("\n")

	if m.loadingCodes {
		content.WriteString("  " + m.Styles.Muted.Render("Loading codes..."))
		return content.String()
	}

	if !m.commissionable {
		content.WriteString("  " + m.Styles.Muted.Render("Device is already paired"))
		return content.String()
	}

	if m.codes == nil || !m.codes.Available {
		content.WriteString("  " + m.Styles.Muted.Render("Codes not available"))
		return content.String()
	}

	// Manual code
	if m.codes.ManualCode != "" {
		content.WriteString("  " + m.Styles.Label.Render("Manual Code:  "))
		content.WriteString(m.Styles.Value.Render(m.codes.ManualCode))
		content.WriteString("\n")
	}

	// QR code string
	if m.codes.QRCode != "" {
		content.WriteString("  " + m.Styles.Label.Render("QR Code:      "))
		content.WriteString(m.Styles.Value.Render(m.codes.QRCode))
		content.WriteString("\n")
	}

	// Setup PIN
	if m.codes.SetupPINCode != 0 {
		content.WriteString("  " + m.Styles.Label.Render("Setup PIN:    "))
		content.WriteString(m.Styles.Value.Render(fmt.Sprintf("%d", m.codes.SetupPINCode)))
		content.WriteString("\n")
	}

	// Discriminator
	if m.codes.Discriminator != 0 {
		content.WriteString("  " + m.Styles.Label.Render("Discriminator:"))
		content.WriteString(" ")
		content.WriteString(m.Styles.Value.Render(fmt.Sprintf("%d", m.codes.Discriminator)))
	}

	return content.String()
}

func (m MatterEditModel) renderResetButton() string {
	selected := matterEditField(m.Cursor) == matterFieldReset

	if m.pendingReset {
		selector := m.Styles.RenderSelector(selected)
		return selector + m.Styles.ButtonDanger.Render("⚠ CONFIRM FACTORY RESET")
	}

	selector := m.Styles.RenderSelector(selected)
	if selected {
		return selector + m.Styles.ButtonDanger.Render("Factory Reset")
	}
	return selector + m.Styles.Button.Render("Factory Reset")
}
