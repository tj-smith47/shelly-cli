package views

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/fleet"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// FleetDeps holds dependencies for the fleet view.
type FleetDeps struct {
	Ctx        context.Context
	Svc        *shelly.Service
	IOS        *iostreams.IOStreams
	Cfg        *config.Config
	FocusState *focus.State
}

// Validate ensures all required dependencies are set.
func (d FleetDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
	}
	if d.FocusState == nil {
		return errNilFocusState
	}
	return nil
}

// FleetConnectMsg signals a fleet connection attempt result.
type FleetConnectMsg struct {
	Fleet *shelly.FleetConnection
	Err   error
}

// Fleet is the fleet view for Shelly Cloud Fleet management.
// This provides cloud-based fleet operations (NOT local device management).
type Fleet struct {
	ctx context.Context
	svc *shelly.Service
	ios *iostreams.IOStreams
	cfg *config.Config
	id  ViewID

	// Fleet manager
	fleetConn *shelly.FleetConnection

	// Component models
	devices    fleet.DevicesModel
	groups     fleet.GroupsModel
	health     fleet.HealthModel
	operations fleet.OperationsModel

	// State
	focusState *focus.State // Unified focus state (single source of truth)
	connecting bool
	connErr    error
	width      int
	height     int
	styles     FleetStyles

	// Layout calculator for flexible panel sizing
	layoutCalc *layout.TwoColumnLayout
}

// FleetStyles holds styles for the fleet view.
type FleetStyles struct {
	Panel       lipgloss.Style
	PanelActive lipgloss.Style
	Title       lipgloss.Style
	Muted       lipgloss.Style
	Connected   lipgloss.Style
	Error       lipgloss.Style
}

// DefaultFleetStyles returns default styles for the fleet view.
func DefaultFleetStyles() FleetStyles {
	colors := theme.GetSemanticColors()
	return FleetStyles{
		Panel:       styles.PanelBorder(),
		PanelActive: styles.PanelBorderActive(),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Connected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
	}
}

// NewFleet creates a new fleet view.
func NewFleet(deps FleetDeps) *Fleet {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("fleet view init", err)
		panic("fleet: " + err.Error())
	}

	devicesDeps := fleet.DevicesDeps{Ctx: deps.Ctx}
	groupsDeps := fleet.GroupsDeps{Ctx: deps.Ctx}
	healthDeps := fleet.HealthDeps{Ctx: deps.Ctx}
	operationsDeps := fleet.OperationsDeps{Ctx: deps.Ctx}

	// Create flexible layout with 60/40 column split (left/right)
	layoutCalc := layout.NewTwoColumnLayout(0.6, 1)

	// Configure left column (Devices takes full height)
	layoutCalc.LeftColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(focus.PanelFleetDevices), MinHeight: 10, ExpandOnFocus: true},
	}

	// Configure right column panels (Groups, Health, Operations) with expansion on focus
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(focus.PanelFleetGroups), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelFleetHealth), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelFleetOperations), MinHeight: 5, ExpandOnFocus: true},
	}

	f := &Fleet{
		ctx:        deps.Ctx,
		svc:        deps.Svc,
		ios:        deps.IOS,
		cfg:        deps.Cfg,
		id:         tabs.TabFleet,
		devices:    fleet.NewDevices(devicesDeps),
		groups:     fleet.NewGroups(groupsDeps),
		health:     fleet.NewHealth(healthDeps),
		operations: fleet.NewOperations(operationsDeps),
		focusState: deps.FocusState,
		styles:     DefaultFleetStyles(),
		layoutCalc: layoutCalc,
	}

	// Initialize focus states so the default focused panel (Devices) receives key events
	f.updateFocusStates()

	return f
}

// Init returns the initial command.
func (f *Fleet) Init() tea.Cmd {
	return tea.Batch(
		f.devices.Init(),
		f.groups.Init(),
		f.health.Init(),
		f.operations.Init(),
	)
}

// ID returns the view ID.
func (f *Fleet) ID() ViewID {
	return f.id
}

// SetSize sets the view dimensions.
func (f *Fleet) SetSize(width, height int) View {
	f.width = width
	f.height = height

	if f.isNarrow() {
		// Narrow mode: all components get full width, full height
		contentWidth := width - 4
		contentHeight := height - 6

		f.devices = f.devices.SetSize(contentWidth, contentHeight)
		f.groups = f.groups.SetSize(contentWidth, contentHeight)
		f.health = f.health.SetSize(contentWidth, contentHeight)
		f.operations = f.operations.SetSize(contentWidth, contentHeight)
		return f
	}

	// Update layout with new dimensions and focus
	f.layoutCalc.SetSize(width, height)
	activePanel := f.focusState.ActivePanel()
	if activePanel.TabFor() == tabs.TabFleet {
		f.layoutCalc.SetFocus(layout.PanelID(activePanel))
	} else {
		f.layoutCalc.SetFocus(-1)
	}

	// Calculate panel dimensions using flexible layout
	dims := f.layoutCalc.Calculate()

	// Apply size to left column (Devices)
	// Pass full panel dimensions - components handle their own borders via rendering.New()
	if d, ok := dims[layout.PanelID(focus.PanelFleetDevices)]; ok {
		f.devices = f.devices.SetSize(d.Width, d.Height)
	}

	// Apply sizes to right column components
	if d, ok := dims[layout.PanelID(focus.PanelFleetGroups)]; ok {
		f.groups = f.groups.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelFleetHealth)]; ok {
		f.health = f.health.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelFleetOperations)]; ok {
		f.operations = f.operations.SetSize(d.Width, d.Height)
	}

	return f
}

// Update handles messages.
func (f *Fleet) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle focus changes from the unified focus state
	if _, ok := msg.(focus.ChangedMsg); ok {
		f.updateFocusStates()
		return f, nil
	}

	// Handle fleet connection result
	if connMsg, ok := msg.(FleetConnectMsg); ok {
		f.connecting = false
		if connMsg.Err != nil {
			f.connErr = connMsg.Err
			return f, nil
		}
		f.fleetConn = connMsg.Fleet
		f.connErr = nil

		// Set fleet manager on all components
		var cmd tea.Cmd
		f.devices, cmd = f.devices.SetFleetManager(connMsg.Fleet.Manager)
		cmds = append(cmds, cmd)
		f.groups, cmd = f.groups.SetFleetManager(connMsg.Fleet.Manager)
		cmds = append(cmds, cmd)
		f.health, cmd = f.health.SetFleetManager(connMsg.Fleet.Manager)
		cmds = append(cmds, cmd)
		f.operations = f.operations.SetFleetManager(connMsg.Fleet.Manager)

		return f, tea.Batch(cmds...)
	}

	// Handle edit modal coordination messages - notify app.go of modal state changes
	if _, ok := msg.(fleet.GroupEditOpenedMsg); ok {
		cmds = append(cmds, func() tea.Msg {
			return messages.ModalOpenedMsg{ID: focus.OverlayEditModal, Mode: focus.ModeModal}
		})
	}
	if _, ok := msg.(messages.EditClosedMsg); ok {
		cmds = append(cmds, func() tea.Msg {
			return messages.ModalClosedMsg{ID: focus.OverlayEditModal}
		})
	}

	// Handle group edit completion with toast
	if cmd := f.handleGroupEditMsg(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Handle keyboard input
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		f.handleKeyPress(keyMsg)
	}

	// Update components
	cmd := f.updateComponents(msg)
	cmds = append(cmds, cmd)

	return f, tea.Batch(cmds...)
}

// handleGroupEditMsg processes group edit completion messages and returns toast commands.
func (f *Fleet) handleGroupEditMsg(msg tea.Msg) tea.Cmd {
	switch editMsg := msg.(type) {
	case messages.EditClosedMsg:
		if editMsg.Saved {
			return toast.Success("Group saved")
		}
	case fleet.GroupCommandResultMsg:
		if editMsg.Err != nil {
			return toast.Error("Group command failed: " + editMsg.Err.Error())
		}
		switch editMsg.Action {
		case fleet.GroupCommandOn:
			return toast.Success("Group relays turned on")
		case fleet.GroupCommandOff:
			return toast.Success("Group relays turned off")
		case fleet.GroupCommandToggle:
			return toast.Success("Group relays toggled")
		}
	}
	return nil
}

func (f *Fleet) handleKeyPress(msg tea.KeyPressMsg) {
	switch msg.String() {
	case keyconst.KeyTab:
		f.focusState.NextPanel()
		f.updateFocusStates()
	case keyconst.KeyShiftTab:
		f.focusState.PrevPanel()
		f.updateFocusStates()
	case keyconst.Shift1:
		f.focusState.JumpToPanel(1)
		f.updateFocusStates()
	case keyconst.Shift2:
		f.focusState.JumpToPanel(2)
		f.updateFocusStates()
	case keyconst.Shift3:
		f.focusState.JumpToPanel(3)
		f.updateFocusStates()
	case keyconst.Shift4:
		f.focusState.JumpToPanel(4)
		f.updateFocusStates()
	case "c":
		// Connect/disconnect
		if f.fleetConn == nil && !f.connecting {
			f.connecting = true
			f.connErr = nil
		}
	}
}

func (f *Fleet) updateFocusStates() {
	f.devices = f.devices.SetFocused(f.focusState.IsPanelFocused(focus.PanelFleetDevices)).SetPanelIndex(1)
	f.groups = f.groups.SetFocused(f.focusState.IsPanelFocused(focus.PanelFleetGroups)).SetPanelIndex(2)
	f.health = f.health.SetFocused(f.focusState.IsPanelFocused(focus.PanelFleetHealth)).SetPanelIndex(3)
	f.operations = f.operations.SetFocused(f.focusState.IsPanelFocused(focus.PanelFleetOperations)).SetPanelIndex(4)

	// Recalculate layout with new focus (panels resize on focus change)
	if f.layoutCalc != nil && f.width > 0 && f.height > 0 {
		f.SetSize(f.width, f.height)
	}
}

func (f *Fleet) updateComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Only update the focused component for key messages and action requests
	_, isKeyPress := msg.(tea.KeyPressMsg)
	if isKeyPress || messages.IsActionRequest(msg) {
		switch f.focusState.ActivePanel() {
		case focus.PanelFleetDevices:
			f.devices, cmd = f.devices.Update(msg)
		case focus.PanelFleetGroups:
			f.groups, cmd = f.groups.Update(msg)
		case focus.PanelFleetHealth:
			f.health, cmd = f.health.Update(msg)
		case focus.PanelFleetOperations:
			f.operations, cmd = f.operations.Update(msg)
		default:
			// Panels from other tabs - no action needed
		}
		cmds = append(cmds, cmd)
	} else {
		// For non-key messages, update all components
		f.devices, cmd = f.devices.Update(msg)
		cmds = append(cmds, cmd)
		f.groups, cmd = f.groups.Update(msg)
		cmds = append(cmds, cmd)
		f.health, cmd = f.health.Update(msg)
		cmds = append(cmds, cmd)
		f.operations, cmd = f.operations.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// isNarrow returns true if the view should use narrow/vertical layout.
func (f *Fleet) isNarrow() bool {
	return f.width < 80
}

// View renders the fleet view.
func (f *Fleet) View() string {
	if f.width == 0 || f.height == 0 {
		return ""
	}

	// If not connected, show connection prompt
	if f.fleetConn == nil {
		return f.renderConnectionPrompt()
	}

	if f.isNarrow() {
		return f.renderNarrowLayout()
	}

	return f.renderStandardLayout()
}

func (f *Fleet) renderNarrowLayout() string {
	// In narrow mode, show only the focused panel at full width
	// Components already have embedded titles from rendering.New()
	switch f.focusState.ActivePanel() {
	case focus.PanelFleetDevices:
		return f.devices.View()
	case focus.PanelFleetGroups:
		return f.groups.View()
	case focus.PanelFleetHealth:
		return f.health.View()
	case focus.PanelFleetOperations:
		return f.operations.View()
	default:
		return f.devices.View()
	}
}

func (f *Fleet) renderStandardLayout() string {
	// Render panels (components already have embedded titles)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left,
		f.groups.View(),
		f.health.View(),
		f.operations.View(),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, f.devices.View(), " ", rightColumn)
}

func (f *Fleet) renderConnectionPrompt() string {
	var content strings.Builder

	content.WriteString(f.styles.Title.Render("Shelly Cloud Fleet"))
	content.WriteString("\n\n")

	switch {
	case f.connecting:
		content.WriteString(f.styles.Muted.Render("Connecting to Shelly Cloud..."))
	case f.connErr != nil:
		content.WriteString(f.styles.Error.Render("Connection failed: " + f.connErr.Error()))
		content.WriteString("\n\n")
		content.WriteString(f.styles.Muted.Render("Press 'c' to retry connection"))
		content.WriteString("\n")
		content.WriteString(f.styles.Muted.Render("Ensure SHELLY_INTEGRATOR_TAG and SHELLY_INTEGRATOR_TOKEN are set"))
	default:
		content.WriteString(f.styles.Muted.Render("Not connected to Shelly Cloud"))
		content.WriteString("\n\n")
		content.WriteString(f.styles.Muted.Render("To connect, you need:"))
		content.WriteString("\n")
		content.WriteString(f.styles.Muted.Render("1. A Shelly Cloud Integrator account"))
		content.WriteString("\n")
		content.WriteString(f.styles.Muted.Render("2. SHELLY_INTEGRATOR_TAG environment variable"))
		content.WriteString("\n")
		content.WriteString(f.styles.Muted.Render("3. SHELLY_INTEGRATOR_TOKEN environment variable"))
		content.WriteString("\n\n")
		content.WriteString(f.styles.Muted.Render("Press 'c' to connect"))
	}

	// Center the content
	style := lipgloss.NewStyle().
		Width(f.width).
		Height(f.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content.String())
}

// Refresh reloads all components.
func (f *Fleet) Refresh() tea.Cmd {
	if f.fleetConn == nil {
		return nil
	}

	cmds := make([]tea.Cmd, 0, 3)
	var cmd tea.Cmd

	f.devices, cmd = f.devices.Refresh()
	cmds = append(cmds, cmd)

	f.groups, cmd = f.groups.Refresh()
	cmds = append(cmds, cmd)

	f.health, cmd = f.health.Refresh()
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// Connect initiates a connection to the Shelly Cloud.
func (f *Fleet) Connect() tea.Cmd {
	return func() tea.Msg {
		creds, err := shelly.GetIntegratorCredentials(f.ios, f.cfg)
		if err != nil {
			return FleetConnectMsg{Err: err}
		}

		conn, err := shelly.ConnectFleet(f.ctx, f.ios, creds)
		if err != nil {
			return FleetConnectMsg{Err: err}
		}

		return FleetConnectMsg{Fleet: conn}
	}
}

// Devices returns the devices component.
func (f *Fleet) Devices() fleet.DevicesModel {
	return f.devices
}

// Groups returns the groups component.
func (f *Fleet) Groups() fleet.GroupsModel {
	return f.groups
}

// Health returns the health component.
func (f *Fleet) Health() fleet.HealthModel {
	return f.health
}

// Operations returns the operations component.
func (f *Fleet) Operations() fleet.OperationsModel {
	return f.operations
}

// Connected returns whether the fleet is connected.
func (f *Fleet) Connected() bool {
	return f.fleetConn != nil
}

// Connecting returns whether a connection is in progress.
func (f *Fleet) Connecting() bool {
	return f.connecting
}

// ConnectionError returns any connection error.
func (f *Fleet) ConnectionError() error {
	return f.connErr
}

// StatusSummary returns a status summary string.
func (f *Fleet) StatusSummary() string {
	if f.connecting {
		return f.styles.Muted.Render("Connecting to Shelly Cloud...")
	}

	if f.connErr != nil {
		return f.styles.Error.Render("Connection failed")
	}

	if f.fleetConn == nil {
		return f.styles.Muted.Render("Not connected to Shelly Cloud")
	}

	var parts []string

	// Connection status
	parts = append(parts, f.styles.Connected.Render("Connected"))

	// Device count
	deviceCount := f.devices.DeviceCount()
	onlineCount := f.devices.OnlineCount()
	if deviceCount > 0 {
		parts = append(parts, f.styles.Muted.Render(fmt.Sprintf("%d devices (%d online)", deviceCount, onlineCount)))
	}

	// Operation status
	if f.operations.Executing() {
		parts = append(parts, "Executing batch operation...")
	}

	if len(parts) == 0 {
		return f.styles.Muted.Render("Fleet ready")
	}

	return strings.Join(parts, " | ")
}

// Close closes the fleet connection.
func (f *Fleet) Close() {
	if f.fleetConn != nil {
		f.fleetConn.Close()
		f.fleetConn = nil
	}
}
