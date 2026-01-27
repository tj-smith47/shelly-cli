package views

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/ble"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cloud"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/inputs"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/protocols"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/security"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/smarthome"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/system"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/wifi"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// configLoadPhase tracks which component is being loaded.
type configLoadPhase int

const (
	configLoadIdle      configLoadPhase = iota
	configLoadWifi                      // Panel 2
	configLoadSystem                    // Panel 3
	configLoadCloud                     // Panel 4
	configLoadSecurity                  // Panel 5
	configLoadBLE                       // Panel 6
	configLoadInputs                    // Panel 7
	configLoadProtocols                 // Panel 8
	configLoadSmartHome                 // Panel 9
)

// configLoadNextMsg triggers loading the next component in sequence.
type configLoadNextMsg struct {
	phase  configLoadPhase
	device string // Track device to prevent stale messages from advancing phase
}

// ConfigDeps holds dependencies for the config view.
type ConfigDeps struct {
	Ctx        context.Context
	Svc        *shelly.Service
	FocusState *focus.State
}

// Validate ensures all required dependencies are set.
func (d ConfigDeps) Validate() error {
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

// Config is the config view that composes WiFi, System, Cloud, Inputs, BLE, Protocols, Security, and SmartHome.
type Config struct {
	ctx context.Context
	svc *shelly.Service
	id  ViewID

	// Component models
	wifi      wifi.Model
	system    system.Model
	cloud     cloud.Model
	inputs    inputs.Model
	ble       ble.Model
	protocols protocols.Model
	security  security.Model
	smarthome smarthome.Model

	// State
	device     string
	focusState *focus.State // Unified focus state (single source of truth)
	width      int
	height     int
	styles     ConfigStyles
	loadPhase  configLoadPhase // Tracks sequential loading progress

	// Layout calculator for flexible panel sizing
	layoutCalc *layout.TwoColumnLayout
}

// ConfigStyles holds styles for the config view.
type ConfigStyles struct {
	Panel       lipgloss.Style
	PanelActive lipgloss.Style
	Title       lipgloss.Style
	Muted       lipgloss.Style
}

// DefaultConfigStyles returns default styles for the config view.
func DefaultConfigStyles() ConfigStyles {
	colors := theme.GetSemanticColors()
	return ConfigStyles{
		Panel:       styles.PanelBorder(),
		PanelActive: styles.PanelBorderActive(),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewConfig creates a new config view.
func NewConfig(deps ConfigDeps) *Config {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("config view init", err)
		panic("config: " + err.Error())
	}

	wifiDeps := wifi.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	systemDeps := system.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	cloudDeps := cloud.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	inputsDeps := inputs.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	bleDeps := ble.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	protocolsDeps := protocols.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	securityDeps := security.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	smarthomeDeps := smarthome.Deps{Ctx: deps.Ctx, Svc: deps.Svc}

	// Create flexible layout with 50/50 column split
	layoutCalc := layout.NewTwoColumnLayout(0.5, 1)

	// Configure left column panels (WiFi, System, Cloud, Security) with expansion on focus
	// MinHeight = borders (2) + padding (2) + content (1) = 5 minimum
	layoutCalc.LeftColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(focus.PanelConfigWiFi), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelConfigSystem), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelConfigCloud), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelConfigSecurity), MinHeight: 5, ExpandOnFocus: true},
	}

	// Configure right column panels (BLE, Inputs, Protocols, SmartHome)
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(focus.PanelConfigBLE), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelConfigInputs), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelConfigProtocols), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelConfigSmartHome), MinHeight: 5, ExpandOnFocus: true},
	}

	c := &Config{
		ctx:        deps.Ctx,
		svc:        deps.Svc,
		id:         tabs.TabConfig,
		wifi:       wifi.New(wifiDeps),
		system:     system.New(systemDeps),
		cloud:      cloud.New(cloudDeps),
		inputs:     inputs.New(inputsDeps),
		ble:        ble.New(bleDeps),
		protocols:  protocols.New(protocolsDeps),
		security:   security.New(securityDeps),
		smarthome:  smarthome.New(smarthomeDeps),
		focusState: deps.FocusState,
		styles:     DefaultConfigStyles(),
		layoutCalc: layoutCalc,
	}

	// Initialize focus states so the default focused panel (WiFi) receives key events
	c.updateFocusStates()

	return c
}

// Init returns the initial command.
func (c *Config) Init() tea.Cmd {
	return tea.Batch(
		c.wifi.Init(),
		c.system.Init(),
		c.cloud.Init(),
		c.inputs.Init(),
		c.ble.Init(),
		c.protocols.Init(),
		c.security.Init(),
		c.smarthome.Init(),
	)
}

// ID returns the view ID.
func (c *Config) ID() ViewID {
	return c.id
}

// SetDevice sets the device for all components.
// Components are loaded sequentially to avoid overwhelming the device with concurrent RPC calls.
func (c *Config) SetDevice(device string) tea.Cmd {
	if device == c.device {
		return nil
	}
	c.device = device

	// Reset all components by clearing their device (ignore cmds - no loading yet)
	c.wifi, _ = c.wifi.SetDevice("")
	c.system, _ = c.system.SetDevice("")
	c.cloud, _ = c.cloud.SetDevice("")
	c.inputs, _ = c.inputs.SetDevice("")
	c.ble, _ = c.ble.SetDevice("")
	c.protocols, _ = c.protocols.SetDevice("")
	c.security, _ = c.security.SetDevice("")
	c.smarthome, _ = c.smarthome.SetDevice("")

	// Start sequential loading with first component
	c.loadPhase = configLoadWifi
	return func() tea.Msg {
		return configLoadNextMsg{phase: configLoadWifi, device: device}
	}
}

// loadNextComponent triggers loading for the current phase.
// Order matches Shift+N panel order: WiFi(2), System(3), Cloud(4), Security(5),
// BLE(6), Inputs(7), Protocols(8), SmartHome(9).
func (c *Config) loadNextComponent() tea.Cmd {
	switch c.loadPhase {
	case configLoadWifi: // Panel 2
		newWifi, cmd := c.wifi.SetDevice(c.device)
		c.wifi = newWifi
		return cmd
	case configLoadSystem: // Panel 3
		newSystem, cmd := c.system.SetDevice(c.device)
		c.system = newSystem
		return cmd
	case configLoadCloud: // Panel 4
		newCloud, cmd := c.cloud.SetDevice(c.device)
		c.cloud = newCloud
		return cmd
	case configLoadSecurity: // Panel 5
		newSecurity, cmd := c.security.SetDevice(c.device)
		c.security = newSecurity
		return cmd
	case configLoadBLE: // Panel 6
		newBLE, cmd := c.ble.SetDevice(c.device)
		c.ble = newBLE
		return cmd
	case configLoadInputs: // Panel 7
		newInputs, cmd := c.inputs.SetDevice(c.device)
		c.inputs = newInputs
		return cmd
	case configLoadProtocols: // Panel 8
		newProtocols, cmd := c.protocols.SetDevice(c.device)
		c.protocols = newProtocols
		return cmd
	case configLoadSmartHome: // Panel 9
		newSmartHome, cmd := c.smarthome.SetDevice(c.device)
		c.smarthome = newSmartHome
		return cmd
	default:
		return nil
	}
}

// advanceLoadPhase moves to the next loading phase and returns command to trigger it.
func (c *Config) advanceLoadPhase() tea.Cmd {
	device := c.device // Capture current device for the closure
	// Loading order follows Shift+N panel order: 2, 3, 4, 5, 6, 7, 8, 9
	switch c.loadPhase {
	case configLoadWifi: // Panel 2 -> Panel 3
		c.loadPhase = configLoadSystem
	case configLoadSystem: // Panel 3 -> Panel 4
		c.loadPhase = configLoadCloud
	case configLoadCloud: // Panel 4 -> Panel 5
		c.loadPhase = configLoadSecurity
	case configLoadSecurity: // Panel 5 -> Panel 6
		c.loadPhase = configLoadBLE
	case configLoadBLE: // Panel 6 -> Panel 7
		c.loadPhase = configLoadInputs
	case configLoadInputs: // Panel 7 -> Panel 8
		c.loadPhase = configLoadProtocols
	case configLoadProtocols: // Panel 8 -> Panel 9
		c.loadPhase = configLoadSmartHome
	case configLoadSmartHome:
		c.loadPhase = configLoadIdle
		return nil // All done
	default:
		return nil
	}
	nextPhase := c.loadPhase // Capture for closure
	return func() tea.Msg {
		return configLoadNextMsg{phase: nextPhase, device: device}
	}
}

// Update handles messages.
func (c *Config) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle focus changes from the unified focus state
	if _, ok := msg.(focus.ChangedMsg); ok {
		c.updateFocusStates()
		return c, nil
	}

	// Handle edit modal opened messages - notify app.go of modal state change
	if _, ok := msg.(messages.EditOpenedMsg); ok {
		c.setEditModalDimensions()
		cmds = append(cmds, func() tea.Msg {
			return messages.ModalOpenedMsg{ID: focus.OverlayEditModal, Mode: focus.ModeModal}
		})
	}

	// Handle edit completion messages with toast feedback
	if cmd := c.handleEditClosedMsg(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Handle input trigger result messages with toast feedback
	if cmd := handleInputTriggerMsg(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Handle sequential loading messages
	if loadMsg, ok := msg.(configLoadNextMsg); ok {
		// Only process if device matches current device (prevents stale messages)
		if loadMsg.phase == c.loadPhase && loadMsg.device == c.device {
			cmd := c.loadNextComponent()
			cmds = append(cmds, cmd)
		}
		return c, tea.Batch(cmds...)
	}

	// Check for component completion to advance sequential loading
	if cmd := c.handleComponentLoaded(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Handle keyboard input - only update focused component for key messages
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd := c.handleKeyPress(keyMsg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmd := c.updateFocusedComponent(msg)
		cmds = append(cmds, cmd)
	} else if messages.IsActionRequest(msg) {
		// Action request messages go only to the focused component
		cmd := c.updateFocusedComponent(msg)
		cmds = append(cmds, cmd)
	} else {
		// For non-key messages (async results), update ALL components
		cmd := c.updateAllComponents(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

// handleEditClosedMsg processes edit completion messages and returns toast and coordination commands.
func (c *Config) handleEditClosedMsg(msg tea.Msg) tea.Cmd {
	if editMsg, ok := msg.(messages.EditClosedMsg); ok {
		cmds := []tea.Cmd{
			// Notify app.go that modal is closed
			func() tea.Msg { return messages.ModalClosedMsg{ID: focus.OverlayEditModal} },
		}
		if editMsg.Saved {
			cmds = append(cmds, toast.Success("Settings saved"))
		}
		return tea.Batch(cmds...)
	}
	return nil
}

// handleInputTriggerMsg processes input trigger result messages and returns toast commands.
func handleInputTriggerMsg(msg tea.Msg) tea.Cmd {
	if triggerMsg, ok := msg.(inputs.TriggerResultMsg); ok {
		if triggerMsg.Err != nil {
			return toast.Error("Trigger failed: " + triggerMsg.Err.Error())
		}
		return toast.Success(fmt.Sprintf("Input %d: %s triggered", triggerMsg.InputID, triggerMsg.EventType))
	}
	return nil
}

// handleComponentLoaded checks for component completion messages and advances loading.
// Only advances if the message is for the current device to prevent stale responses.
func (c *Config) handleComponentLoaded(msg tea.Msg) tea.Cmd {
	expectedPhase := c.phaseForMessage(msg)
	if expectedPhase == configLoadIdle || c.loadPhase != expectedPhase {
		return nil
	}

	// Validate the component's device matches current device
	// This prevents stale responses from advancing the phase
	if !c.isComponentForCurrentDevice(expectedPhase) {
		return nil
	}

	return c.advanceLoadPhase()
}

// isComponentForCurrentDevice checks if the component for the given phase
// is configured for the current device.
func (c *Config) isComponentForCurrentDevice(phase configLoadPhase) bool {
	switch phase {
	case configLoadWifi:
		return c.wifi.Device() == c.device
	case configLoadSystem:
		return c.system.Device() == c.device
	case configLoadCloud:
		return c.cloud.Device() == c.device
	case configLoadInputs:
		return c.inputs.Device() == c.device
	case configLoadBLE:
		return c.ble.Device() == c.device
	case configLoadProtocols:
		return c.protocols.Device() == c.device
	case configLoadSecurity:
		return c.security.Device() == c.device
	case configLoadSmartHome:
		return c.smarthome.Device() == c.device
	default:
		return false
	}
}

// phaseForMessage returns the load phase that corresponds to a message type.
func (c *Config) phaseForMessage(msg tea.Msg) configLoadPhase {
	switch msg.(type) {
	case wifi.StatusLoadedMsg:
		return configLoadWifi
	case system.StatusLoadedMsg:
		return configLoadSystem
	case cloud.StatusLoadedMsg:
		return configLoadCloud
	case inputs.LoadedMsg:
		return configLoadInputs
	case ble.StatusLoadedMsg:
		return configLoadBLE
	case protocols.StatusLoadedMsg:
		return configLoadProtocols
	case security.StatusLoadedMsg:
		return configLoadSecurity
	case smarthome.StatusLoadedMsg:
		return configLoadSmartHome
	default:
		return configLoadIdle
	}
}

func (c *Config) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	prevPanel := c.focusState.ActivePanel()

	switch msg.String() {
	case keyconst.KeyTab:
		c.focusState.NextPanel()
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.KeyShiftTab:
		c.focusState.PrevPanel()
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	// Shift+N hotkeys match column-by-column order: left column (2-5), right column (6-9)
	case keyconst.Shift2:
		c.focusState.JumpToPanel(2) // WiFi
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.Shift3:
		c.focusState.JumpToPanel(3) // System
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.Shift4:
		c.focusState.JumpToPanel(4) // Cloud
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.Shift5:
		c.focusState.JumpToPanel(5) // Security
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.Shift6:
		c.focusState.JumpToPanel(6) // BLE
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.Shift7:
		c.focusState.JumpToPanel(7) // Inputs
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.Shift8:
		c.focusState.JumpToPanel(8) // Protocols
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	case keyconst.Shift9:
		c.focusState.JumpToPanel(9) // SmartHome
		c.updateFocusStates()
		return c.emitFocusChanged(prevPanel)
	}
	return nil
}

// emitFocusChanged returns a command that emits a FocusChangedMsg if panel actually changed.
func (c *Config) emitFocusChanged(prevPanel focus.GlobalPanelID) tea.Cmd {
	newPanel := c.focusState.ActivePanel()
	if newPanel == prevPanel {
		return nil
	}
	return func() tea.Msg {
		return c.focusState.NewChangedMsg(
			c.focusState.ActiveTab(),
			prevPanel,
			false, // tab didn't change
			true,  // panel changed
			false, // overlay didn't change
		)
	}
}

func (c *Config) updateFocusStates() {
	// Query focusState for panel focus (single source of truth)
	// Panel indices match column-by-column order: left column (2-5), right column (6-9)
	c.wifi = c.wifi.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigWiFi)).SetPanelIndex(2)
	c.system = c.system.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigSystem)).SetPanelIndex(3)
	c.cloud = c.cloud.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigCloud)).SetPanelIndex(4)
	c.security = c.security.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigSecurity)).SetPanelIndex(5)
	c.ble = c.ble.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigBLE)).SetPanelIndex(6)
	c.inputs = c.inputs.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigInputs)).SetPanelIndex(7)
	c.protocols = c.protocols.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigProtocols)).SetPanelIndex(8)
	c.smarthome = c.smarthome.SetFocused(c.focusState.IsPanelFocused(focus.PanelConfigSmartHome)).SetPanelIndex(9)

	// Recalculate layout with new focus (panels resize on focus change)
	if c.layoutCalc != nil && c.width > 0 && c.height > 0 {
		c.SetSize(c.width, c.height)
	}
}

func (c *Config) updateFocusedComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch c.focusState.ActivePanel() {
	case focus.PanelConfigWiFi:
		c.wifi, cmd = c.wifi.Update(msg)
	case focus.PanelConfigSystem:
		c.system, cmd = c.system.Update(msg)
	case focus.PanelConfigCloud:
		c.cloud, cmd = c.cloud.Update(msg)
	case focus.PanelConfigInputs:
		c.inputs, cmd = c.inputs.Update(msg)
	case focus.PanelConfigBLE:
		c.ble, cmd = c.ble.Update(msg)
	case focus.PanelConfigProtocols:
		c.protocols, cmd = c.protocols.Update(msg)
	case focus.PanelConfigSecurity:
		c.security, cmd = c.security.Update(msg)
	case focus.PanelConfigSmartHome:
		c.smarthome, cmd = c.smarthome.Update(msg)
	default:
		// Panels from other tabs - no action needed
	}
	return cmd
}

func (c *Config) updateAllComponents(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0, 8)

	var wifiCmd, systemCmd, cloudCmd, inputsCmd, bleCmd, protocolsCmd, securityCmd, smarthomeCmd tea.Cmd
	c.wifi, wifiCmd = c.wifi.Update(msg)
	c.system, systemCmd = c.system.Update(msg)
	c.cloud, cloudCmd = c.cloud.Update(msg)
	c.inputs, inputsCmd = c.inputs.Update(msg)
	c.ble, bleCmd = c.ble.Update(msg)
	c.protocols, protocolsCmd = c.protocols.Update(msg)
	c.security, securityCmd = c.security.Update(msg)
	c.smarthome, smarthomeCmd = c.smarthome.Update(msg)

	cmds = append(cmds, wifiCmd, systemCmd, cloudCmd, inputsCmd, bleCmd, protocolsCmd, securityCmd, smarthomeCmd)
	return tea.Batch(cmds...)
}

// isNarrow returns true if the view should use narrow/vertical layout.
func (c *Config) isNarrow() bool {
	return c.width < 80
}

// View renders the config view.
func (c *Config) View() string {
	if c.device == "" {
		return c.styles.Muted.Render("No device selected. Select a device from the Devices tab.")
	}

	if c.isNarrow() {
		return c.renderNarrowLayout()
	}

	return c.renderStandardLayout()
}

func (c *Config) renderNarrowLayout() string {
	// In narrow mode, show only the focused panel at full width
	// Components already have embedded titles from rendering.New()
	switch c.focusState.ActivePanel() {
	case focus.PanelConfigWiFi:
		return c.wifi.View()
	case focus.PanelConfigSystem:
		return c.system.View()
	case focus.PanelConfigCloud:
		return c.cloud.View()
	case focus.PanelConfigInputs:
		return c.inputs.View()
	case focus.PanelConfigBLE:
		return c.ble.View()
	case focus.PanelConfigProtocols:
		return c.protocols.View()
	case focus.PanelConfigSecurity:
		return c.security.View()
	case focus.PanelConfigSmartHome:
		return c.smarthome.View()
	default:
		return c.wifi.View()
	}
}

func (c *Config) renderStandardLayout() string {
	// Render panels (components already have embedded titles)
	// Left column: WiFi, System, Cloud, Security
	leftPanels := []string{
		c.wifi.View(),
		c.system.View(),
		c.cloud.View(),
		c.security.View(),
	}

	// Right column: BLE, Inputs, Protocols, SmartHome
	rightPanels := []string{
		c.ble.View(),
		c.inputs.View(),
		c.protocols.View(),
		c.smarthome.View(),
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", rightCol)
}

// SetSize sets the view dimensions.
func (c *Config) SetSize(width, height int) View {
	c.width = width
	c.height = height

	if c.isNarrow() {
		// Narrow mode: all components get full width, full height
		contentWidth := width - 4
		contentHeight := height - 6

		c.wifi = c.wifi.SetSize(contentWidth, contentHeight)
		c.system = c.system.SetSize(contentWidth, contentHeight)
		c.cloud = c.cloud.SetSize(contentWidth, contentHeight)
		c.inputs = c.inputs.SetSize(contentWidth, contentHeight)
		c.ble = c.ble.SetSize(contentWidth, contentHeight)
		c.protocols = c.protocols.SetSize(contentWidth, contentHeight)
		c.security = c.security.SetSize(contentWidth, contentHeight)
		c.smarthome = c.smarthome.SetSize(contentWidth, contentHeight)
		return c
	}

	// Update layout with new dimensions and focus
	c.layoutCalc.SetSize(width, height)
	// Only expand panels when a config panel has focus, otherwise distribute evenly
	activePanel := c.focusState.ActivePanel()
	if activePanel.TabFor() == tabs.TabConfig && activePanel != focus.PanelDeviceList {
		c.layoutCalc.SetFocus(layout.PanelID(activePanel))
	} else {
		c.layoutCalc.SetFocus(-1) // No expansion when device list is focused
	}

	// Calculate panel dimensions using flexible layout
	dims := c.layoutCalc.Calculate()

	// Apply sizes to left column components (WiFi, System, Cloud, Security)
	// Pass full panel dimensions - components handle their own borders via rendering.New()
	if d, ok := dims[layout.PanelID(focus.PanelConfigWiFi)]; ok {
		c.wifi = c.wifi.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelConfigSystem)]; ok {
		c.system = c.system.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelConfigCloud)]; ok {
		c.cloud = c.cloud.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelConfigSecurity)]; ok {
		c.security = c.security.SetSize(d.Width, d.Height)
	}

	// Apply sizes to right column components (BLE, Inputs, Protocols, SmartHome)
	if d, ok := dims[layout.PanelID(focus.PanelConfigBLE)]; ok {
		c.ble = c.ble.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelConfigInputs)]; ok {
		c.inputs = c.inputs.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelConfigProtocols)]; ok {
		c.protocols = c.protocols.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelConfigSmartHome)]; ok {
		c.smarthome = c.smarthome.SetSize(d.Width, d.Height)
	}

	return c
}

// Device returns the current device.
func (c *Config) Device() string {
	return c.device
}

// HasActiveModal returns true if any component has an edit modal visible.
// Implements ModalProvider interface.
func (c *Config) HasActiveModal() bool {
	return c.wifi.IsEditing() ||
		c.system.IsEditing() ||
		c.cloud.IsEditing() ||
		c.inputs.IsEditing() ||
		c.ble.IsEditing() ||
		c.protocols.IsEditing() ||
		c.security.IsEditing() ||
		c.smarthome.IsEditing()
}

// setEditModalDimensions sets proper modal dimensions using screen dimensions.
// This is called when an edit modal opens to ensure it gets screen-based sizing
// rather than panel-based sizing.
func (c *Config) setEditModalDimensions() {
	modalWidth := c.width * 90 / 100
	modalHeight := c.height * 90 / 100
	if modalWidth < 60 {
		modalWidth = 60
	}
	if modalHeight < 20 {
		modalHeight = 20
	}
	// Set dimensions for whichever component has an edit modal open
	if c.wifi.IsEditing() {
		c.wifi = c.wifi.SetEditModalSize(modalWidth, modalHeight)
	}
	if c.system.IsEditing() {
		c.system = c.system.SetEditModalSize(modalWidth, modalHeight)
	}
	if c.cloud.IsEditing() {
		c.cloud = c.cloud.SetEditModalSize(modalWidth, modalHeight)
	}
	if c.inputs.IsEditing() {
		c.inputs = c.inputs.SetEditModalSize(modalWidth, modalHeight)
	}
	if c.ble.IsEditing() {
		c.ble = c.ble.SetEditModalSize(modalWidth, modalHeight)
	}
	if c.protocols.IsEditing() {
		c.protocols = c.protocols.SetEditModalSize(modalWidth, modalHeight)
	}
	if c.security.IsEditing() {
		c.security = c.security.SetEditModalSize(modalWidth, modalHeight)
	}
	if c.smarthome.IsEditing() {
		c.smarthome = c.smarthome.SetEditModalSize(modalWidth, modalHeight)
	}
}

// RenderModal returns the active modal's view for full-screen overlay rendering.
// Implements ModalProvider interface.
func (c *Config) RenderModal() string {
	// Return the first active modal's view
	if c.wifi.IsEditing() {
		return c.wifi.RenderEditModal()
	}
	if c.system.IsEditing() {
		return c.system.RenderEditModal()
	}
	if c.cloud.IsEditing() {
		return c.cloud.RenderEditModal()
	}
	if c.inputs.IsEditing() {
		return c.inputs.RenderEditModal()
	}
	if c.ble.IsEditing() {
		return c.ble.RenderEditModal()
	}
	if c.protocols.IsEditing() {
		return c.protocols.RenderEditModal()
	}
	if c.security.IsEditing() {
		return c.security.RenderEditModal()
	}
	if c.smarthome.IsEditing() {
		return c.smarthome.RenderEditModal()
	}
	return ""
}
