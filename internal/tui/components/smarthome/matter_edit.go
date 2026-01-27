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
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
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
	ctx     context.Context
	svc     *shelly.Service
	device  string
	visible bool
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

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

	// Focus
	field      matterEditField
	fieldCount int // Dynamic based on whether reset is visible
}

// NewMatterEditModel creates a new Matter configuration edit modal.
func NewMatterEditModel(ctx context.Context, svc *shelly.Service) MatterEditModel {
	return MatterEditModel{
		ctx:    ctx,
		svc:    svc,
		styles: editmodal.DefaultStyles().WithLabelWidth(14),
	}
}

// Show displays the edit modal with the given device and Matter state.
func (m MatterEditModel) Show(device string, matter *shelly.TUIMatterStatus) (MatterEditModel, tea.Cmd) {
	m.device = device
	m.visible = true
	m.saving = false
	m.resetting = false
	m.pendingReset = false
	m.err = nil
	m.field = matterFieldEnable
	m.codes = nil
	m.loadingCodes = false

	if matter != nil {
		m.enabled = matter.Enabled
		m.commissionable = matter.Commissionable
		m.fabricsCount = matter.FabricsCount
		m.pendingEnabled = matter.Enabled
	}

	// Calculate field count: enable toggle + reset button (when enabled)
	m.fieldCount = 1 // Enable toggle always
	if m.enabled {
		m.fieldCount = 2 // Add reset button
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
	m.visible = false
	m.pendingReset = false
	return m
}

// Visible returns whether the modal is visible.
func (m MatterEditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m MatterEditModel) SetSize(width, height int) MatterEditModel {
	m.width = width
	m.height = height
	return m
}

// Update handles messages.
func (m MatterEditModel) Update(msg tea.Msg) (MatterEditModel, tea.Cmd) {
	if !m.visible {
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
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - update state and close modal
		m.enabled = m.pendingEnabled
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case MatterResetResultMsg:
		m.resetting = false
		m.pendingReset = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		if !m.saving && !m.resetting {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingReset = false
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m MatterEditModel) handleNavigation(msg messages.NavigationMsg) (MatterEditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		if m.field > 0 {
			m.field--
			m.pendingReset = false
		}
	case messages.NavDown:
		if int(m.field) < m.fieldCount-1 {
			m.field++
			m.pendingReset = false
		}
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m MatterEditModel) handleKey(msg tea.KeyPressMsg) (MatterEditModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEsc, keyconst.KeyCtrlOpenBracket:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter, keyconst.KeyCtrlS:
		return m.handleSaveOrAction()

	case "t", keyconst.KeySpace:
		if !m.saving && !m.resetting && m.field == matterFieldEnable {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingReset = false
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if int(m.field) < m.fieldCount-1 {
			m.field++
			m.pendingReset = false
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if m.field > 0 {
			m.field--
			m.pendingReset = false
		}
		return m, nil
	}

	return m, nil
}

func (m MatterEditModel) handleSaveOrAction() (MatterEditModel, tea.Cmd) {
	if m.saving || m.resetting {
		return m, nil
	}

	// If focused on reset button, handle reset confirmation
	if m.field == matterFieldReset {
		if m.pendingReset {
			// Second press - confirm reset
			m.resetting = true
			m.pendingReset = false
			m.err = nil
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

func (m MatterEditModel) createSaveCmd() tea.Cmd {
	newEnabled := m.pendingEnabled
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var err error
		if newEnabled {
			err = m.svc.Wireless().MatterEnable(ctx, m.device)
		} else {
			err = m.svc.Wireless().MatterDisable(ctx, m.device)
		}
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

func (m MatterEditModel) createResetCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Wireless().MatterReset(ctx, m.device)
		return MatterResetResultMsg{Err: err}
	}
}

func (m MatterEditModel) fetchCodes() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		info, err := m.svc.Wireless().MatterGetCommissioningCode(ctx, m.device)
		return MatterCodesLoadedMsg{Codes: info, Err: err}
	}
}

// View renders the edit modal.
func (m MatterEditModel) View() string {
	if !m.visible {
		return ""
	}

	footer := m.buildFooter()
	r := rendering.NewModal(m.width, m.height, "Matter Configuration", footer)

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
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return r.SetContent(content.String()).Render()
}

func (m MatterEditModel) buildFooter() string {
	if m.saving {
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

	content.WriteString(m.styles.Label.Render("Status:"))
	content.WriteString(" ")
	if m.enabled {
		content.WriteString(m.styles.StatusOn.Render("● Enabled"))
	} else {
		content.WriteString(m.styles.StatusOff.Render("○ Disabled"))
	}

	if m.enabled {
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Commission:"))
		content.WriteString(" ")
		if m.commissionable {
			content.WriteString(m.styles.Warning.Render("Ready to pair"))
		} else {
			content.WriteString(m.styles.Muted.Render("Already paired"))
		}

		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Fabrics:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.fabricsCount)))
	}

	return content.String()
}

func (m MatterEditModel) renderEnableToggle() string {
	selected := m.field == matterFieldEnable

	var value string
	if m.pendingEnabled {
		value = m.styles.StatusOn.Render("[●] ON ")
	} else {
		value = m.styles.StatusOff.Render("[ ] OFF")
	}

	return m.styles.RenderFieldRow(selected, "Enabled:", value)
}

func (m MatterEditModel) renderChangeIndicator() string {
	var msg string
	if m.pendingEnabled {
		msg = "Will enable Matter"
	} else {
		msg = "Will disable Matter"
	}
	return m.styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

func (m MatterEditModel) renderCodes() string {
	var content strings.Builder

	content.WriteString(m.styles.LabelFocus.Render("Commissioning Codes"))
	content.WriteString("\n")

	if m.loadingCodes {
		content.WriteString("  " + m.styles.Muted.Render("Loading codes..."))
		return content.String()
	}

	if !m.commissionable {
		content.WriteString("  " + m.styles.Muted.Render("Device is already paired"))
		return content.String()
	}

	if m.codes == nil || !m.codes.Available {
		content.WriteString("  " + m.styles.Muted.Render("Codes not available"))
		return content.String()
	}

	// Manual code
	if m.codes.ManualCode != "" {
		content.WriteString("  " + m.styles.Label.Render("Manual Code:  "))
		content.WriteString(m.styles.Value.Render(m.codes.ManualCode))
		content.WriteString("\n")
	}

	// QR code string
	if m.codes.QRCode != "" {
		content.WriteString("  " + m.styles.Label.Render("QR Code:      "))
		content.WriteString(m.styles.Value.Render(m.codes.QRCode))
		content.WriteString("\n")
	}

	// Setup PIN
	if m.codes.SetupPINCode != 0 {
		content.WriteString("  " + m.styles.Label.Render("Setup PIN:    "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.codes.SetupPINCode)))
		content.WriteString("\n")
	}

	// Discriminator
	if m.codes.Discriminator != 0 {
		content.WriteString("  " + m.styles.Label.Render("Discriminator:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.codes.Discriminator)))
	}

	return content.String()
}

func (m MatterEditModel) renderResetButton() string {
	selected := m.field == matterFieldReset

	if m.pendingReset {
		selector := m.styles.RenderSelector(selected)
		return selector + m.styles.ButtonDanger.Render("⚠ CONFIRM FACTORY RESET")
	}

	selector := m.styles.RenderSelector(selected)
	if selected {
		return selector + m.styles.ButtonDanger.Render("Factory Reset")
	}
	return selector + m.styles.Button.Render("Factory Reset")
}
