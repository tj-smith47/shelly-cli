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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/system"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/wifi"
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
)

// configLoadPhase tracks which component is being loaded.
type configLoadPhase int

const (
	configLoadIdle configLoadPhase = iota
	configLoadWifi
	configLoadSystem
	configLoadCloud
	configLoadInputs
	configLoadBLE
	configLoadProtocols
	configLoadSecurity
)

// configLoadNextMsg triggers loading the next component in sequence.
type configLoadNextMsg struct {
	phase configLoadPhase
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

// Config is the config view that composes WiFi, System, Cloud, Inputs, BLE, Protocols, and Security.
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

	// State
	device       string
	focusedPanel ConfigPanel
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

	// Create flexible layout with 50/50 column split
	layoutCalc := layout.NewTwoColumnLayout(0.5, 1)

	// Configure left column panels (WiFi, System, BLE) with expansion on focus
	layoutCalc.LeftColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(PanelWiFi), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelSystem), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelBLE), MinHeight: 5, ExpandOnFocus: true},
	}

	// Configure right column panels (Cloud, Inputs, Protocols, Security)
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(PanelCloud), MinHeight: 4, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelInputs), MinHeight: 4, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelProtocols), MinHeight: 4, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelSecurity), MinHeight: 4, ExpandOnFocus: true},
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

	// Start sequential loading with first component
	c.loadPhase = configLoadWifi
	return func() tea.Msg {
		return configLoadNextMsg{phase: configLoadWifi}
	}
}

// loadNextComponent triggers loading for the current phase.
func (c *Config) loadNextComponent() tea.Cmd {
	switch c.loadPhase {
	case configLoadWifi:
		newWifi, cmd := c.wifi.SetDevice(c.device)
		c.wifi = newWifi
		return cmd
	case configLoadSystem:
		newSystem, cmd := c.system.SetDevice(c.device)
		c.system = newSystem
		return cmd
	case configLoadCloud:
		newCloud, cmd := c.cloud.SetDevice(c.device)
		c.cloud = newCloud
		return cmd
	case configLoadInputs:
		newInputs, cmd := c.inputs.SetDevice(c.device)
		c.inputs = newInputs
		return cmd
	case configLoadBLE:
		newBLE, cmd := c.ble.SetDevice(c.device)
		c.ble = newBLE
		return cmd
	case configLoadProtocols:
		newProtocols, cmd := c.protocols.SetDevice(c.device)
		c.protocols = newProtocols
		return cmd
	case configLoadSecurity:
		newSecurity, cmd := c.security.SetDevice(c.device)
		c.security = newSecurity
		return cmd
	default:
		return nil
	}
}

// advanceLoadPhase moves to the next loading phase and returns command to trigger it.
func (c *Config) advanceLoadPhase() tea.Cmd {
	switch c.loadPhase {
	case configLoadWifi:
		c.loadPhase = configLoadSystem
	case configLoadSystem:
		c.loadPhase = configLoadCloud
	case configLoadCloud:
		c.loadPhase = configLoadInputs
	case configLoadInputs:
		c.loadPhase = configLoadBLE
	case configLoadBLE:
		c.loadPhase = configLoadProtocols
	case configLoadProtocols:
		c.loadPhase = configLoadSecurity
	case configLoadSecurity:
		c.loadPhase = configLoadIdle
		return nil // All done
	default:
		return nil
	}
	return func() tea.Msg {
		return configLoadNextMsg{phase: c.loadPhase}
	}
}

// Update handles messages.
func (c *Config) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle sequential loading messages
	if loadMsg, ok := msg.(configLoadNextMsg); ok {
		if loadMsg.phase == c.loadPhase {
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
		c.handleKeyPress(keyMsg)
		cmd := c.updateFocusedComponent(msg)
		cmds = append(cmds, cmd)
	} else {
		// For non-key messages (async results), update ALL components
		cmd := c.updateAllComponents(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

// handleComponentLoaded checks for component completion messages and advances loading.
func (c *Config) handleComponentLoaded(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case wifi.StatusLoadedMsg:
		if c.loadPhase == configLoadWifi {
			return c.advanceLoadPhase()
		}
	case system.StatusLoadedMsg:
		if c.loadPhase == configLoadSystem {
			return c.advanceLoadPhase()
		}
	case cloud.StatusLoadedMsg:
		if c.loadPhase == configLoadCloud {
			return c.advanceLoadPhase()
		}
	case inputs.LoadedMsg:
		if c.loadPhase == configLoadInputs {
			return c.advanceLoadPhase()
		}
	case ble.StatusLoadedMsg:
		if c.loadPhase == configLoadBLE {
			return c.advanceLoadPhase()
		}
	case protocols.StatusLoadedMsg:
		if c.loadPhase == configLoadProtocols {
			return c.advanceLoadPhase()
		}
	case security.StatusLoadedMsg:
		if c.loadPhase == configLoadSecurity {
			return c.advanceLoadPhase()
		}
	}
	return nil
}

func (c *Config) handleKeyPress(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "tab":
		c.focusNext()
	case "shift+tab":
		c.focusPrev()
	}
}

func (c *Config) focusNext() {
	panels := []ConfigPanel{
		PanelWiFi, PanelSystem, PanelBLE,
		PanelCloud, PanelInputs, PanelProtocols, PanelSecurity,
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
	panels := []ConfigPanel{
		PanelWiFi, PanelSystem, PanelBLE,
		PanelCloud, PanelInputs, PanelProtocols, PanelSecurity,
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
	c.wifi = c.wifi.SetFocused(c.focusedPanel == PanelWiFi)
	c.system = c.system.SetFocused(c.focusedPanel == PanelSystem)
	c.cloud = c.cloud.SetFocused(c.focusedPanel == PanelCloud)
	c.inputs = c.inputs.SetFocused(c.focusedPanel == PanelInputs)
	c.ble = c.ble.SetFocused(c.focusedPanel == PanelBLE)
	c.protocols = c.protocols.SetFocused(c.focusedPanel == PanelProtocols)
	c.security = c.security.SetFocused(c.focusedPanel == PanelSecurity)

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
	}
	return cmd
}

func (c *Config) updateAllComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	var wifiCmd, systemCmd, cloudCmd, inputsCmd, bleCmd, protocolsCmd, securityCmd tea.Cmd
	c.wifi, wifiCmd = c.wifi.Update(msg)
	c.system, systemCmd = c.system.Update(msg)
	c.cloud, cloudCmd = c.cloud.Update(msg)
	c.inputs, inputsCmd = c.inputs.Update(msg)
	c.ble, bleCmd = c.ble.Update(msg)
	c.protocols, protocolsCmd = c.protocols.Update(msg)
	c.security, securityCmd = c.security.Update(msg)

	cmds = append(cmds, wifiCmd, systemCmd, cloudCmd, inputsCmd, bleCmd, protocolsCmd, securityCmd)
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
	default:
		return c.wifi.View()
	}
}

func (c *Config) renderStandardLayout() string {
	// Render panels (components already have embedded titles)
	// Left column: WiFi, System, BLE
	leftPanels := []string{
		c.wifi.View(),
		c.system.View(),
		c.ble.View(),
	}

	// Right column: Cloud, Inputs, Protocols, Security
	rightPanels := []string{
		c.cloud.View(),
		c.inputs.View(),
		c.protocols.View(),
		c.security.View(),
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
		return c
	}

	// Update layout with new dimensions and focus
	c.layoutCalc.SetSize(width, height)
	c.layoutCalc.SetFocus(layout.PanelID(c.focusedPanel))

	// Calculate panel dimensions using flexible layout
	dims := c.layoutCalc.Calculate()

	// Apply sizes to left column components (WiFi, System, BLE)
	if d, ok := dims[layout.PanelID(PanelWiFi)]; ok {
		cw, ch := d.ContentDimensions(2)
		c.wifi = c.wifi.SetSize(cw, ch)
	}
	if d, ok := dims[layout.PanelID(PanelSystem)]; ok {
		cw, ch := d.ContentDimensions(2)
		c.system = c.system.SetSize(cw, ch)
	}
	if d, ok := dims[layout.PanelID(PanelBLE)]; ok {
		cw, ch := d.ContentDimensions(2)
		c.ble = c.ble.SetSize(cw, ch)
	}

	// Apply sizes to right column components (Cloud, Inputs, Protocols, Security)
	if d, ok := dims[layout.PanelID(PanelCloud)]; ok {
		cw, ch := d.ContentDimensions(2)
		c.cloud = c.cloud.SetSize(cw, ch)
	}
	if d, ok := dims[layout.PanelID(PanelInputs)]; ok {
		cw, ch := d.ContentDimensions(2)
		c.inputs = c.inputs.SetSize(cw, ch)
	}
	if d, ok := dims[layout.PanelID(PanelProtocols)]; ok {
		cw, ch := d.ContentDimensions(2)
		c.protocols = c.protocols.SetSize(cw, ch)
	}
	if d, ok := dims[layout.PanelID(PanelSecurity)]; ok {
		cw, ch := d.ContentDimensions(2)
		c.security = c.security.SetSize(cw, ch)
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
