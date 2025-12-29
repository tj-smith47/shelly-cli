package views

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// ConfigPanel identifies which panel is focused in the Config view.
type ConfigPanel int

// Config panel constants.
const (
	PanelWiFi ConfigPanel = iota
	PanelSystem
	PanelCloud
	PanelInputs
	PanelBLE
	PanelProtocols
	PanelSecurity
	PanelSmartHome
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
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d ConfigDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
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
	device       string
	focusedPanel ConfigPanel
	viewFocused  bool // Whether the view content has focus (vs device list)
	width        int
	height       int
	styles       ConfigStyles
	loadPhase    configLoadPhase // Tracks sequential loading progress

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
	}
}

// NewConfig creates a new config view.
func NewConfig(deps ConfigDeps) *Config {
	if err := deps.Validate(); err != nil {
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
		{ID: layout.PanelID(PanelWiFi), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelSystem), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelCloud), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelSecurity), MinHeight: 5, ExpandOnFocus: true},
	}

	// Configure right column panels (BLE, Inputs, Protocols, SmartHome)
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(PanelBLE), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelInputs), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelProtocols), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelSmartHome), MinHeight: 5, ExpandOnFocus: true},
	}

	c := &Config{
		ctx:          deps.Ctx,
		svc:          deps.Svc,
		id:           tabs.TabConfig,
		wifi:         wifi.New(wifiDeps),
		system:       system.New(systemDeps),
		cloud:        cloud.New(cloudDeps),
		inputs:       inputs.New(inputsDeps),
		ble:          ble.New(bleDeps),
		protocols:    protocols.New(protocolsDeps),
		security:     security.New(securityDeps),
		smarthome:    smarthome.New(smarthomeDeps),
		focusedPanel: PanelWiFi,
		styles:       DefaultConfigStyles(),
		layoutCalc:   layoutCalc,
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

	// Handle view focus changes from app.go
	if focusMsg, ok := msg.(ViewFocusChangedMsg); ok {
		// When regaining focus, reset to first panel so Tab cycling starts fresh
		if focusMsg.Focused && !c.viewFocused {
			c.focusedPanel = PanelWiFi
		}
		c.viewFocused = focusMsg.Focused
		c.updateFocusStates()
		return c, nil
	}

	// Handle edit completion messages with toast feedback
	if cmd := c.handleEditClosedMsg(msg); cmd != nil {
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
	} else {
		// For non-key messages (async results), update ALL components
		cmd := c.updateAllComponents(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

// handleEditClosedMsg processes edit completion messages and returns toast commands.
func (c *Config) handleEditClosedMsg(msg tea.Msg) tea.Cmd {
	switch editMsg := msg.(type) {
	case wifi.EditClosedMsg:
		if editMsg.Saved {
			return toast.Success("WiFi settings saved")
		}
	case system.EditClosedMsg:
		if editMsg.Saved {
			return toast.Success("System settings saved")
		}
	case security.EditClosedMsg:
		if editMsg.Saved {
			return toast.Success("Authentication settings saved")
		}
	case cloud.EditClosedMsg:
		if editMsg.Saved {
			return toast.Success("Cloud settings saved")
		}
	case protocols.MQTTEditClosedMsg:
		if editMsg.Saved {
			return toast.Success("MQTT settings saved")
		}
	case inputs.EditClosedMsg:
		if editMsg.Saved {
			return toast.Success("Input settings saved")
		}
	case ble.EditClosedMsg:
		if editMsg.Saved {
			return toast.Success("Bluetooth settings saved")
		}
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
	switch msg.String() {
	case keyTab:
		// If on last panel, return focus to device list
		if c.focusedPanel == PanelSmartHome {
			c.viewFocused = false
			c.updateFocusStates()
			return func() tea.Msg { return ReturnFocusMsg{} }
		}
		c.viewFocused = true // View has focus when cycling panels
		c.focusNext()
	case keyShiftTab:
		// If on first panel, return focus to device list
		if c.focusedPanel == PanelWiFi {
			c.viewFocused = false
			c.updateFocusStates()
			return func() tea.Msg { return ReturnFocusMsg{} }
		}
		c.viewFocused = true // View has focus when cycling panels
		c.focusPrev()
	// Shift+N hotkeys match column-by-column order: left column (2-5), right column (6-9)
	case keyconst.Shift2:
		c.viewFocused = true
		c.focusedPanel = PanelWiFi
		c.updateFocusStates()
	case keyconst.Shift3:
		c.viewFocused = true
		c.focusedPanel = PanelSystem
		c.updateFocusStates()
	case keyconst.Shift4:
		c.viewFocused = true
		c.focusedPanel = PanelCloud
		c.updateFocusStates()
	case keyconst.Shift5:
		c.viewFocused = true
		c.focusedPanel = PanelSecurity
		c.updateFocusStates()
	case keyconst.Shift6:
		c.viewFocused = true
		c.focusedPanel = PanelBLE
		c.updateFocusStates()
	case keyconst.Shift7:
		c.viewFocused = true
		c.focusedPanel = PanelInputs
		c.updateFocusStates()
	case keyconst.Shift8:
		c.viewFocused = true
		c.focusedPanel = PanelProtocols
		c.updateFocusStates()
	case keyconst.Shift9:
		c.viewFocused = true
		c.focusedPanel = PanelSmartHome
		c.updateFocusStates()
	}
	return nil
}

func (c *Config) focusNext() {
	// Column-by-column: left column top-to-bottom, then right column top-to-bottom
	panels := []ConfigPanel{
		PanelWiFi, PanelSystem, PanelCloud, PanelSecurity, // Left column
		PanelBLE, PanelInputs, PanelProtocols, PanelSmartHome, // Right column
	}
	for i, p := range panels {
		if p == c.focusedPanel {
			c.focusedPanel = panels[(i+1)%len(panels)]
			break
		}
	}
	c.updateFocusStates()
}

func (c *Config) focusPrev() {
	// Column-by-column: left column top-to-bottom, then right column top-to-bottom
	panels := []ConfigPanel{
		PanelWiFi, PanelSystem, PanelCloud, PanelSecurity, // Left column
		PanelBLE, PanelInputs, PanelProtocols, PanelSmartHome, // Right column
	}
	for i, p := range panels {
		if p == c.focusedPanel {
			prevIdx := (i - 1 + len(panels)) % len(panels)
			c.focusedPanel = panels[prevIdx]
			break
		}
	}
	c.updateFocusStates()
}

func (c *Config) updateFocusStates() {
	// Panels only show focused when the view has overall focus AND it's the active panel
	// Panel indices match column-by-column order: left column (2-5), right column (6-9)
	c.wifi = c.wifi.SetFocused(c.viewFocused && c.focusedPanel == PanelWiFi).SetPanelIndex(2)
	c.system = c.system.SetFocused(c.viewFocused && c.focusedPanel == PanelSystem).SetPanelIndex(3)
	c.cloud = c.cloud.SetFocused(c.viewFocused && c.focusedPanel == PanelCloud).SetPanelIndex(4)
	c.security = c.security.SetFocused(c.viewFocused && c.focusedPanel == PanelSecurity).SetPanelIndex(5)
	c.ble = c.ble.SetFocused(c.viewFocused && c.focusedPanel == PanelBLE).SetPanelIndex(6)
	c.inputs = c.inputs.SetFocused(c.viewFocused && c.focusedPanel == PanelInputs).SetPanelIndex(7)
	c.protocols = c.protocols.SetFocused(c.viewFocused && c.focusedPanel == PanelProtocols).SetPanelIndex(8)
	c.smarthome = c.smarthome.SetFocused(c.viewFocused && c.focusedPanel == PanelSmartHome).SetPanelIndex(9)

	// Recalculate layout with new focus (panels resize on focus change)
	if c.layoutCalc != nil && c.width > 0 && c.height > 0 {
		c.SetSize(c.width, c.height)
	}
}

func (c *Config) updateFocusedComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch c.focusedPanel {
	case PanelWiFi:
		c.wifi, cmd = c.wifi.Update(msg)
	case PanelSystem:
		c.system, cmd = c.system.Update(msg)
	case PanelCloud:
		c.cloud, cmd = c.cloud.Update(msg)
	case PanelInputs:
		c.inputs, cmd = c.inputs.Update(msg)
	case PanelBLE:
		c.ble, cmd = c.ble.Update(msg)
	case PanelProtocols:
		c.protocols, cmd = c.protocols.Update(msg)
	case PanelSecurity:
		c.security, cmd = c.security.Update(msg)
	case PanelSmartHome:
		c.smarthome, cmd = c.smarthome.Update(msg)
	}
	return cmd
}

func (c *Config) updateAllComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

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
	switch c.focusedPanel {
	case PanelWiFi:
		return c.wifi.View()
	case PanelSystem:
		return c.system.View()
	case PanelCloud:
		return c.cloud.View()
	case PanelInputs:
		return c.inputs.View()
	case PanelBLE:
		return c.ble.View()
	case PanelProtocols:
		return c.protocols.View()
	case PanelSecurity:
		return c.security.View()
	case PanelSmartHome:
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
	// Only expand panels when view has focus, otherwise distribute evenly
	if c.viewFocused {
		c.layoutCalc.SetFocus(layout.PanelID(c.focusedPanel))
	} else {
		c.layoutCalc.SetFocus(-1) // No expansion when device list is focused
	}

	// Calculate panel dimensions using flexible layout
	dims := c.layoutCalc.Calculate()

	// Apply sizes to left column components (WiFi, System, Cloud, Security)
	// Pass full panel dimensions - components handle their own borders via rendering.New()
	if d, ok := dims[layout.PanelID(PanelWiFi)]; ok {
		c.wifi = c.wifi.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelSystem)]; ok {
		c.system = c.system.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelCloud)]; ok {
		c.cloud = c.cloud.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelSecurity)]; ok {
		c.security = c.security.SetSize(d.Width, d.Height)
	}

	// Apply sizes to right column components (BLE, Inputs, Protocols, SmartHome)
	if d, ok := dims[layout.PanelID(PanelBLE)]; ok {
		c.ble = c.ble.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelInputs)]; ok {
		c.inputs = c.inputs.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelProtocols)]; ok {
		c.protocols = c.protocols.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelSmartHome)]; ok {
		c.smarthome = c.smarthome.SetSize(d.Width, d.Height)
	}

	return c
}

// Device returns the current device.
func (c *Config) Device() string {
	return c.device
}

// FocusedPanel returns the currently focused panel.
func (c *Config) FocusedPanel() ConfigPanel {
	return c.focusedPanel
}

// SetFocusedPanel sets the focused panel.
func (c *Config) SetFocusedPanel(panel ConfigPanel) *Config {
	c.focusedPanel = panel
	c.updateFocusStates()
	return c
}

// SetViewFocused sets whether the view has overall focus (vs device list).
// When false, all panels show as unfocused.
func (c *Config) SetViewFocused(focused bool) *Config {
	c.viewFocused = focused
	c.updateFocusStates()
	return c
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
		c.security.IsEditing()
	// smarthome doesn't have an edit modal
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
	return ""
}
