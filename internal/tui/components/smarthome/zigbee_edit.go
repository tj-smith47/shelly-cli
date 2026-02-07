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

// ZigbeeToggleResultMsg signals that a Zigbee enable/disable toggle completed.
type ZigbeeToggleResultMsg struct {
	Enabled bool
	Err     error
}

// ZigbeeSteeringResultMsg signals that Zigbee network steering completed.
type ZigbeeSteeringResultMsg struct {
	Err error
}

// ZigbeeLeaveResultMsg signals that leaving a Zigbee network completed.
type ZigbeeLeaveResultMsg struct {
	Err error
}

// zigbeeEditField identifies focusable fields in the Zigbee edit modal.
type zigbeeEditField int

const (
	zigbeeFieldEnable zigbeeEditField = iota // Enable/disable toggle
	zigbeeFieldPair                          // Start network steering button
	zigbeeFieldLeave                         // Leave network button
)

// ZigbeeEditModel represents the Zigbee configuration edit modal.
type ZigbeeEditModel struct {
	editmodal.Base

	// Zigbee state
	enabled      bool
	networkState string
	channel      int
	panID        uint16
	eui64        string
	coordinator  string

	// Pending changes
	pendingEnabled bool

	// Network steering state
	steering bool

	// Leave confirmation
	pendingLeave bool
	leaving      bool
}

// NewZigbeeEditModel creates a new Zigbee configuration edit modal.
func NewZigbeeEditModel(ctx context.Context, svc *shelly.Service) ZigbeeEditModel {
	return ZigbeeEditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(14),
		},
	}
}

// Show displays the edit modal with the given device and Zigbee state.
func (m ZigbeeEditModel) Show(device string, zigbee *shelly.TUIZigbeeStatus) (ZigbeeEditModel, tea.Cmd) {
	m.steering = false
	m.leaving = false
	m.pendingLeave = false

	if zigbee != nil {
		m.enabled = zigbee.Enabled
		m.networkState = zigbee.NetworkState
		m.channel = zigbee.Channel
		m.panID = zigbee.PANID
		m.eui64 = zigbee.EUI64
		m.coordinator = zigbee.CoordinatorEUI64
		m.pendingEnabled = zigbee.Enabled
	}

	// Calculate field count based on state
	fieldCount := m.calcFieldCount()
	m.Base.Show(device, fieldCount)

	return m, nil
}

// Hide hides the edit modal.
func (m ZigbeeEditModel) Hide() ZigbeeEditModel {
	m.Base.Hide()
	m.pendingLeave = false
	return m
}

// Visible returns whether the modal is visible.
func (m ZigbeeEditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m ZigbeeEditModel) SetSize(width, height int) ZigbeeEditModel {
	m.Base.SetSize(width, height)
	return m
}

// calcFieldCount returns the number of focusable fields based on current state.
func (m ZigbeeEditModel) calcFieldCount() int {
	count := 1 // Enable toggle always present
	if !m.enabled {
		return count
	}
	count++ // Pair button
	if m.networkState == zigbeeStateJoined {
		count++ // Leave button
	}
	return count
}

// Update handles messages.
func (m ZigbeeEditModel) Update(msg tea.Msg) (ZigbeeEditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m ZigbeeEditModel) handleMessage(msg tea.Msg) (ZigbeeEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ZigbeeSteeringResultMsg:
		m.steering = false
		if msg.Err != nil {
			m.Err = msg.Err
			return m, nil
		}
		// Steering started successfully - close modal to show status update
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case ZigbeeLeaveResultMsg:
		m.leaving = false
		m.pendingLeave = false
		if msg.Err != nil {
			m.Err = msg.Err
			return m, nil
		}
		// Left network successfully - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

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
		if !m.Saving && !m.steering && !m.leaving {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingLeave = false
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

func (m ZigbeeEditModel) applyAction(action editmodal.KeyAction) (ZigbeeEditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionNone:
		return m, nil
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		return m.handleSaveOrAction()
	case editmodal.ActionNavDown, editmodal.ActionNext:
		if m.Cursor < m.FieldCount-1 {
			m.Cursor++
			m.pendingLeave = false
		}
		return m, nil
	case editmodal.ActionNavUp, editmodal.ActionPrev:
		if m.Cursor > 0 {
			m.Cursor--
			m.pendingLeave = false
		}
		return m, nil
	}
	return m, nil
}

func (m ZigbeeEditModel) handleCustomKey(msg tea.KeyPressMsg) (ZigbeeEditModel, tea.Cmd) {
	switch msg.String() {
	case "t", keyconst.KeySpace:
		if !m.Saving && !m.steering && !m.leaving && zigbeeEditField(m.Cursor) == zigbeeFieldEnable {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingLeave = false
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if m.Cursor < m.FieldCount-1 {
			m.Cursor++
			m.pendingLeave = false
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
			m.pendingLeave = false
		}
		return m, nil
	}

	return m, nil
}

func (m ZigbeeEditModel) handleSaveOrAction() (ZigbeeEditModel, tea.Cmd) {
	if m.Saving || m.steering || m.leaving {
		return m, nil
	}

	switch zigbeeEditField(m.Cursor) {
	case zigbeeFieldPair:
		return m.startSteering()
	case zigbeeFieldLeave:
		return m.handleLeaveAction()
	default:
		// Save enable toggle change
		return m.save()
	}
}

func (m ZigbeeEditModel) startSteering() (ZigbeeEditModel, tea.Cmd) {
	m.steering = true
	m.Err = nil
	return m, m.createSteeringCmd()
}

func (m ZigbeeEditModel) handleLeaveAction() (ZigbeeEditModel, tea.Cmd) {
	if m.pendingLeave {
		// Second press - confirm leave
		m.leaving = true
		m.pendingLeave = false
		m.Err = nil
		return m, m.createLeaveCmd()
	}
	// First press - request confirmation
	m.pendingLeave = true
	return m, nil
}

func (m ZigbeeEditModel) save() (ZigbeeEditModel, tea.Cmd) {
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
			return m.Svc.Wireless().ZigbeeEnable(ctx, m.Device)
		}
		return m.Svc.Wireless().ZigbeeDisable(ctx, m.Device)
	})
	return m, cmd
}

func (m ZigbeeEditModel) createSteeringCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.Ctx, editmodal.SaveTimeout)
		defer cancel()

		err := m.Svc.Wireless().ZigbeeStartNetworkSteering(ctx, m.Device)
		return ZigbeeSteeringResultMsg{Err: err}
	}
}

func (m ZigbeeEditModel) createLeaveCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.Ctx, editmodal.SaveTimeout)
		defer cancel()

		err := m.Svc.Wireless().ZigbeeLeaveNetwork(ctx, m.Device)
		return ZigbeeLeaveResultMsg{Err: err}
	}
}

// View renders the edit modal.
func (m ZigbeeEditModel) View() string {
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

	// Pair button (when enabled)
	if m.enabled {
		content.WriteString("\n\n")
		content.WriteString(m.renderPairButton())
	}

	// Leave button (when joined)
	if m.enabled && m.networkState == zigbeeStateJoined {
		content.WriteString("\n\n")
		content.WriteString(m.renderLeaveButton())
	}

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("Zigbee Configuration", content.String(), footer)
}

func (m ZigbeeEditModel) buildFooter() string {
	if m.Saving {
		return footerSaving
	}
	if m.steering {
		return "Starting network steering..."
	}
	if m.leaving {
		return "Leaving Zigbee network..."
	}
	if m.pendingLeave {
		return "Press Enter again to confirm leaving network, Esc to cancel"
	}
	return "Space/t: Toggle | Enter: Save/Confirm | j/k: Navigate | Esc: Cancel"
}

func (m ZigbeeEditModel) renderStatus() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Status:"))
	content.WriteString(" ")
	if m.enabled {
		content.WriteString(m.Styles.StatusOn.Render("● Enabled"))
	} else {
		content.WriteString(m.Styles.StatusOff.Render("○ Disabled"))
		return content.String()
	}

	content.WriteString("\n")
	content.WriteString(m.Styles.Label.Render("Network:"))
	content.WriteString(" ")
	content.WriteString(m.renderNetworkState())

	if m.networkState == zigbeeStateJoined {
		m.renderJoinedDetails(&content)
	}

	return content.String()
}

func (m ZigbeeEditModel) renderNetworkState() string {
	switch m.networkState {
	case zigbeeStateJoined:
		return m.Styles.StatusOn.Render("Joined")
	case zigbeeStateSteering:
		return m.Styles.Warning.Render("Searching...")
	case zigbeeStateReady:
		return m.Styles.Muted.Render("Ready (not joined)")
	default:
		return m.Styles.Muted.Render(m.networkState)
	}
}

func (m ZigbeeEditModel) renderJoinedDetails(content *strings.Builder) {
	content.WriteString("\n")
	content.WriteString(m.Styles.Label.Render("Channel:"))
	content.WriteString(" ")
	content.WriteString(m.Styles.Value.Render(fmt.Sprintf("%d", m.channel)))

	content.WriteString("\n")
	content.WriteString(m.Styles.Label.Render("PAN ID:"))
	content.WriteString(" ")
	content.WriteString(m.Styles.Value.Render(fmt.Sprintf("0x%04X", m.panID)))

	if m.eui64 != "" {
		content.WriteString("\n")
		content.WriteString(m.Styles.Label.Render("EUI-64:"))
		content.WriteString(" ")
		content.WriteString(m.Styles.Value.Render(m.eui64))
	}

	if m.coordinator != "" {
		content.WriteString("\n")
		content.WriteString(m.Styles.Label.Render("Coordinator:"))
		content.WriteString(" ")
		content.WriteString(m.Styles.Value.Render(m.coordinator))
	}
}

func (m ZigbeeEditModel) renderEnableToggle() string {
	selected := zigbeeEditField(m.Cursor) == zigbeeFieldEnable

	var value string
	if m.pendingEnabled {
		value = m.Styles.StatusOn.Render("[●] ON ")
	} else {
		value = m.Styles.StatusOff.Render("[ ] OFF")
	}

	return m.Styles.RenderFieldRow(selected, "Enabled:", value)
}

func (m ZigbeeEditModel) renderChangeIndicator() string {
	var msg string
	if m.pendingEnabled {
		msg = "Will enable Zigbee"
	} else {
		msg = "Will disable Zigbee"
	}
	return m.Styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

func (m ZigbeeEditModel) renderPairButton() string {
	selected := zigbeeEditField(m.Cursor) == zigbeeFieldPair

	selector := m.Styles.RenderSelector(selected)
	label := "Start Pair Mode"
	if m.networkState == zigbeeStateSteering {
		label = "Steering in progress..."
	}
	if selected {
		return selector + m.Styles.ButtonFocus.Render(label)
	}
	return selector + m.Styles.Button.Render(label)
}

func (m ZigbeeEditModel) renderLeaveButton() string {
	selected := zigbeeEditField(m.Cursor) == zigbeeFieldLeave

	if m.pendingLeave {
		selector := m.Styles.RenderSelector(selected)
		return selector + m.Styles.ButtonDanger.Render("⚠ CONFIRM LEAVE NETWORK")
	}

	selector := m.Styles.RenderSelector(selected)
	if selected {
		return selector + m.Styles.ButtonDanger.Render("Leave Network")
	}
	return selector + m.Styles.Button.Render("Leave Network")
}
