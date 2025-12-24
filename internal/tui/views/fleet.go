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
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// FleetPanel identifies which panel is focused in the Fleet view.
type FleetPanel int

// Fleet panel constants.
const (
	FleetPanelDevices FleetPanel = iota
	FleetPanelGroups
	FleetPanelHealth
	FleetPanelOperations
)

// FleetDeps holds dependencies for the fleet view.
type FleetDeps struct {
	Ctx context.Context
	Svc *shelly.Service
	IOS *iostreams.IOStreams
	Cfg *config.Config
}

// Validate ensures all required dependencies are set.
func (d FleetDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
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
	focusedPanel FleetPanel
	connecting   bool
	connErr      error
	width        int
	height       int
	styles       FleetStyles
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
		Panel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder),
		PanelActive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight),
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
		panic("fleet: " + err.Error())
	}

	devicesDeps := fleet.DevicesDeps{Ctx: deps.Ctx}
	groupsDeps := fleet.GroupsDeps{Ctx: deps.Ctx}
	healthDeps := fleet.HealthDeps{Ctx: deps.Ctx}
	operationsDeps := fleet.OperationsDeps{Ctx: deps.Ctx}

	f := &Fleet{
		ctx:          deps.Ctx,
		svc:          deps.Svc,
		ios:          deps.IOS,
		cfg:          deps.Cfg,
		id:           tabs.TabFleet,
		devices:      fleet.NewDevices(devicesDeps),
		groups:       fleet.NewGroups(groupsDeps),
		health:       fleet.NewHealth(healthDeps),
		operations:   fleet.NewOperations(operationsDeps),
		focusedPanel: FleetPanelDevices,
		styles:       DefaultFleetStyles(),
	}

	// Initialize components with focus
	f.devices = f.devices.SetFocused(true)

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
		f.devices = f.devices.SetSize(width, height-4)
		f.groups = f.groups.SetSize(width, height-4)
		f.health = f.health.SetSize(width, height-4)
		f.operations = f.operations.SetSize(width, height-4)
		return f
	}

	// Standard layout: 2-column layout
	// Left column: Devices (60% width, full height)
	// Right column: Groups, Health, Operations (40% width, stacked)
	leftWidth := (width - 1) * 3 / 5
	rightWidth := width - leftWidth - 1

	// Right column has 3 stacked panels
	rightPanelHeight := (height - 2) / 3

	f.devices = f.devices.SetSize(leftWidth, height)
	f.groups = f.groups.SetSize(rightWidth, rightPanelHeight)
	f.health = f.health.SetSize(rightWidth, rightPanelHeight)
	f.operations = f.operations.SetSize(rightWidth, rightPanelHeight)

	return f
}

// Update handles messages.
func (f *Fleet) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

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

	// Handle keyboard input
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		f.handleKeyPress(keyMsg)
	}

	// Update components
	cmd := f.updateComponents(msg)
	cmds = append(cmds, cmd)

	return f, tea.Batch(cmds...)
}

func (f *Fleet) handleKeyPress(msg tea.KeyPressMsg) {
	switch msg.String() {
	case keyTab:
		f.focusNext()
	case keyShiftTab:
		f.focusPrev()
	case "1":
		f.focusedPanel = FleetPanelDevices
		f.updateFocusStates()
	case "2":
		f.focusedPanel = FleetPanelGroups
		f.updateFocusStates()
	case "3":
		f.focusedPanel = FleetPanelHealth
		f.updateFocusStates()
	case "4":
		f.focusedPanel = FleetPanelOperations
		f.updateFocusStates()
	case "c":
		// Connect/disconnect
		if f.fleetConn == nil && !f.connecting {
			f.connecting = true
			f.connErr = nil
		}
	}
}

func (f *Fleet) focusNext() {
	panels := []FleetPanel{FleetPanelDevices, FleetPanelGroups, FleetPanelHealth, FleetPanelOperations}
	for i, p := range panels {
		if p == f.focusedPanel {
			f.focusedPanel = panels[(i+1)%len(panels)]
			break
		}
	}
	f.updateFocusStates()
}

func (f *Fleet) focusPrev() {
	panels := []FleetPanel{FleetPanelDevices, FleetPanelGroups, FleetPanelHealth, FleetPanelOperations}
	for i, p := range panels {
		if p == f.focusedPanel {
			prevIdx := (i - 1 + len(panels)) % len(panels)
			f.focusedPanel = panels[prevIdx]
			break
		}
	}
	f.updateFocusStates()
}

func (f *Fleet) updateFocusStates() {
	f.devices = f.devices.SetFocused(f.focusedPanel == FleetPanelDevices)
	f.groups = f.groups.SetFocused(f.focusedPanel == FleetPanelGroups)
	f.health = f.health.SetFocused(f.focusedPanel == FleetPanelHealth)
	f.operations = f.operations.SetFocused(f.focusedPanel == FleetPanelOperations)
}

func (f *Fleet) updateComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Only update the focused component for key messages
	if _, ok := msg.(tea.KeyPressMsg); ok {
		switch f.focusedPanel {
		case FleetPanelDevices:
			f.devices, cmd = f.devices.Update(msg)
		case FleetPanelGroups:
			f.groups, cmd = f.groups.Update(msg)
		case FleetPanelHealth:
			f.health, cmd = f.health.Update(msg)
		case FleetPanelOperations:
			f.operations, cmd = f.operations.Update(msg)
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
	panelHeight := f.height - 2

	switch f.focusedPanel {
	case FleetPanelDevices:
		return f.renderPanel(f.devices.View(), f.width, panelHeight, true)
	case FleetPanelGroups:
		return f.renderPanel(f.groups.View(), f.width, panelHeight, true)
	case FleetPanelHealth:
		return f.renderPanel(f.health.View(), f.width, panelHeight, true)
	case FleetPanelOperations:
		return f.renderPanel(f.operations.View(), f.width, panelHeight, true)
	default:
		return f.renderPanel(f.devices.View(), f.width, panelHeight, true)
	}
}

func (f *Fleet) renderStandardLayout() string {
	// Calculate panel dimensions
	leftWidth := (f.width - 1) * 3 / 5
	rightWidth := f.width - leftWidth - 1
	rightPanelHeight := (f.height - 2) / 3

	// Render left column (devices)
	devicesView := f.renderPanel(f.devices.View(), leftWidth, f.height, f.focusedPanel == FleetPanelDevices)

	// Render right column panels (stacked vertically)
	groupsView := f.renderPanel(f.groups.View(), rightWidth, rightPanelHeight, f.focusedPanel == FleetPanelGroups)
	healthView := f.renderPanel(f.health.View(), rightWidth, rightPanelHeight, f.focusedPanel == FleetPanelHealth)
	operationsView := f.renderPanel(f.operations.View(), rightWidth, rightPanelHeight, f.focusedPanel == FleetPanelOperations)

	// Build right column (stacked)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, groupsView, healthView, operationsView)

	// Join columns
	return lipgloss.JoinHorizontal(lipgloss.Top, devicesView, " ", rightColumn)
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
		content.WriteString("To connect, you need:\n")
		content.WriteString(f.styles.Muted.Render("  1. A Shelly Cloud Integrator account\n"))
		content.WriteString(f.styles.Muted.Render("  2. SHELLY_INTEGRATOR_TAG environment variable\n"))
		content.WriteString(f.styles.Muted.Render("  3. SHELLY_INTEGRATOR_TOKEN environment variable\n"))
		content.WriteString("\n")
		content.WriteString(f.styles.Muted.Render("Press 'c' to connect"))
	}

	// Center the content
	style := lipgloss.NewStyle().
		Width(f.width).
		Height(f.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content.String())
}

func (f *Fleet) renderPanel(content string, width, height int, focused bool) string {
	style := f.styles.Panel
	if focused {
		style = f.styles.PanelActive
	}

	return style.
		Width(width).
		Height(height).
		Render(content)
}

// Refresh reloads all components.
func (f *Fleet) Refresh() tea.Cmd {
	if f.fleetConn == nil {
		return nil
	}

	var cmds []tea.Cmd
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

// FocusedPanel returns the currently focused panel.
func (f *Fleet) FocusedPanel() FleetPanel {
	return f.focusedPanel
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
