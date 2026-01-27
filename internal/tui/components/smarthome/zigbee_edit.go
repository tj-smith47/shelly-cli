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
	ctx     context.Context
	svc     *shelly.Service
	device  string
	visible bool
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

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

	// Focus
	field      zigbeeEditField
	fieldCount int // Dynamic based on whether pair/leave are visible
}

// NewZigbeeEditModel creates a new Zigbee configuration edit modal.
func NewZigbeeEditModel(ctx context.Context, svc *shelly.Service) ZigbeeEditModel {
	return ZigbeeEditModel{
		ctx:    ctx,
		svc:    svc,
		styles: editmodal.DefaultStyles().WithLabelWidth(14),
	}
}

// Show displays the edit modal with the given device and Zigbee state.
func (m ZigbeeEditModel) Show(device string, zigbee *shelly.TUIZigbeeStatus) (ZigbeeEditModel, tea.Cmd) {
	m.device = device
	m.visible = true
	m.saving = false
	m.steering = false
	m.leaving = false
	m.pendingLeave = false
	m.err = nil
	m.field = zigbeeFieldEnable

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
	m.fieldCount = m.calcFieldCount()

	return m, nil
}

// Hide hides the edit modal.
func (m ZigbeeEditModel) Hide() ZigbeeEditModel {
	m.visible = false
	m.pendingLeave = false
	return m
}

// Visible returns whether the modal is visible.
func (m ZigbeeEditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m ZigbeeEditModel) SetSize(width, height int) ZigbeeEditModel {
	m.width = width
	m.height = height
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
	if !m.visible {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m ZigbeeEditModel) handleMessage(msg tea.Msg) (ZigbeeEditModel, tea.Cmd) {
	switch msg := msg.(type) {
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

	case ZigbeeSteeringResultMsg:
		m.steering = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Steering started successfully - close modal to show status update
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case ZigbeeLeaveResultMsg:
		m.leaving = false
		m.pendingLeave = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Left network successfully - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		if !m.saving && !m.steering && !m.leaving {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingLeave = false
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m ZigbeeEditModel) handleNavigation(msg messages.NavigationMsg) (ZigbeeEditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		if m.field > 0 {
			m.field--
			m.pendingLeave = false
		}
	case messages.NavDown:
		if int(m.field) < m.fieldCount-1 {
			m.field++
			m.pendingLeave = false
		}
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m ZigbeeEditModel) handleKey(msg tea.KeyPressMsg) (ZigbeeEditModel, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter, keyconst.KeyCtrlS:
		return m.handleSaveOrAction()

	case "t", keyconst.KeySpace:
		if !m.saving && !m.steering && !m.leaving && m.field == zigbeeFieldEnable {
			m.pendingEnabled = !m.pendingEnabled
			m.pendingLeave = false
		}
		return m, nil

	case "j", keyconst.KeyDown:
		if int(m.field) < m.fieldCount-1 {
			m.field++
			m.pendingLeave = false
		}
		return m, nil

	case "k", keyconst.KeyUp:
		if m.field > 0 {
			m.field--
			m.pendingLeave = false
		}
		return m, nil
	}

	return m, nil
}

func (m ZigbeeEditModel) handleSaveOrAction() (ZigbeeEditModel, tea.Cmd) {
	if m.saving || m.steering || m.leaving {
		return m, nil
	}

	switch m.field {
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
	m.err = nil
	return m, m.createSteeringCmd()
}

func (m ZigbeeEditModel) handleLeaveAction() (ZigbeeEditModel, tea.Cmd) {
	if m.pendingLeave {
		// Second press - confirm leave
		m.leaving = true
		m.pendingLeave = false
		m.err = nil
		return m, m.createLeaveCmd()
	}
	// First press - request confirmation
	m.pendingLeave = true
	return m, nil
}

func (m ZigbeeEditModel) save() (ZigbeeEditModel, tea.Cmd) {
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

func (m ZigbeeEditModel) createSaveCmd() tea.Cmd {
	newEnabled := m.pendingEnabled
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var err error
		if newEnabled {
			err = m.svc.Wireless().ZigbeeEnable(ctx, m.device)
		} else {
			err = m.svc.Wireless().ZigbeeDisable(ctx, m.device)
		}
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

func (m ZigbeeEditModel) createSteeringCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Wireless().ZigbeeStartNetworkSteering(ctx, m.device)
		return ZigbeeSteeringResultMsg{Err: err}
	}
}

func (m ZigbeeEditModel) createLeaveCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Wireless().ZigbeeLeaveNetwork(ctx, m.device)
		return ZigbeeLeaveResultMsg{Err: err}
	}
}

// View renders the edit modal.
func (m ZigbeeEditModel) View() string {
	if !m.visible {
		return ""
	}

	footer := m.buildFooter()
	r := rendering.NewModal(m.width, m.height, "Zigbee Configuration", footer)

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
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return r.SetContent(content.String()).Render()
}

func (m ZigbeeEditModel) buildFooter() string {
	if m.saving {
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

	content.WriteString(m.styles.Label.Render("Status:"))
	content.WriteString(" ")
	if m.enabled {
		content.WriteString(m.styles.StatusOn.Render("● Enabled"))
	} else {
		content.WriteString(m.styles.StatusOff.Render("○ Disabled"))
		return content.String()
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Label.Render("Network:"))
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
		return m.styles.StatusOn.Render("Joined")
	case zigbeeStateSteering:
		return m.styles.Warning.Render("Searching...")
	case zigbeeStateReady:
		return m.styles.Muted.Render("Ready (not joined)")
	default:
		return m.styles.Muted.Render(m.networkState)
	}
}

func (m ZigbeeEditModel) renderJoinedDetails(content *strings.Builder) {
	content.WriteString("\n")
	content.WriteString(m.styles.Label.Render("Channel:"))
	content.WriteString(" ")
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.channel)))

	content.WriteString("\n")
	content.WriteString(m.styles.Label.Render("PAN ID:"))
	content.WriteString(" ")
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("0x%04X", m.panID)))

	if m.eui64 != "" {
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("EUI-64:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Value.Render(m.eui64))
	}

	if m.coordinator != "" {
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Coordinator:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Value.Render(m.coordinator))
	}
}

func (m ZigbeeEditModel) renderEnableToggle() string {
	selected := m.field == zigbeeFieldEnable

	var value string
	if m.pendingEnabled {
		value = m.styles.StatusOn.Render("[●] ON ")
	} else {
		value = m.styles.StatusOff.Render("[ ] OFF")
	}

	return m.styles.RenderFieldRow(selected, "Enabled:", value)
}

func (m ZigbeeEditModel) renderChangeIndicator() string {
	var msg string
	if m.pendingEnabled {
		msg = "Will enable Zigbee"
	} else {
		msg = "Will disable Zigbee"
	}
	return m.styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

func (m ZigbeeEditModel) renderPairButton() string {
	selected := m.field == zigbeeFieldPair

	selector := m.styles.RenderSelector(selected)
	label := "Start Pair Mode"
	if m.networkState == zigbeeStateSteering {
		label = "Steering in progress..."
	}
	if selected {
		return selector + m.styles.ButtonFocus.Render(label)
	}
	return selector + m.styles.Button.Render(label)
}

func (m ZigbeeEditModel) renderLeaveButton() string {
	selected := m.field == zigbeeFieldLeave

	if m.pendingLeave {
		selector := m.styles.RenderSelector(selected)
		return selector + m.styles.ButtonDanger.Render("⚠ CONFIRM LEAVE NETWORK")
	}

	selector := m.styles.RenderSelector(selected)
	if selected {
		return selector + m.styles.ButtonDanger.Render("Leave Network")
	}
	return selector + m.styles.Button.Render("Leave Network")
}
