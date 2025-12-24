package views

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cloud"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/inputs"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/system"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/wifi"
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
)

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

// Config is the config view that composes WiFi, System, Cloud, and Inputs components.
type Config struct {
	ctx context.Context
	svc *shelly.Service
	id  ViewID

	// Component models
	wifi   wifi.Model
	system system.Model
	cloud  cloud.Model
	inputs inputs.Model

	// State
	device       string
	focusedPanel ConfigPanel
	width        int
	height       int
	styles       ConfigStyles
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

	return &Config{
		ctx:          deps.Ctx,
		svc:          deps.Svc,
		id:           tabs.TabConfig,
		wifi:         wifi.New(wifiDeps),
		system:       system.New(systemDeps),
		cloud:        cloud.New(cloudDeps),
		inputs:       inputs.New(inputsDeps),
		focusedPanel: PanelWiFi,
		styles:       DefaultConfigStyles(),
	}
}

// Init returns the initial command.
func (c *Config) Init() tea.Cmd {
	return tea.Batch(
		c.wifi.Init(),
		c.system.Init(),
		c.cloud.Init(),
		c.inputs.Init(),
	)
}

// ID returns the view ID.
func (c *Config) ID() ViewID {
	return c.id
}

// SetDevice sets the device for all components.
func (c *Config) SetDevice(device string) tea.Cmd {
	if device == c.device {
		return nil
	}
	c.device = device

	var cmds []tea.Cmd

	newWifi, wifiCmd := c.wifi.SetDevice(device)
	c.wifi = newWifi
	cmds = append(cmds, wifiCmd)

	newSystem, systemCmd := c.system.SetDevice(device)
	c.system = newSystem
	cmds = append(cmds, systemCmd)

	newCloud, cloudCmd := c.cloud.SetDevice(device)
	c.cloud = newCloud
	cmds = append(cmds, cloudCmd)

	newInputs, inputsCmd := c.inputs.SetDevice(device)
	c.inputs = newInputs
	cmds = append(cmds, inputsCmd)

	return tea.Batch(cmds...)
}

// Update handles messages.
func (c *Config) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

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

func (c *Config) handleKeyPress(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "tab":
		c.focusNext()
	case "shift+tab":
		c.focusPrev()
	}
}

func (c *Config) focusNext() {
	panels := []ConfigPanel{PanelWiFi, PanelSystem, PanelCloud, PanelInputs}
	for i, p := range panels {
		if p == c.focusedPanel {
			c.focusedPanel = panels[(i+1)%len(panels)]
			break
		}
	}
	c.updateFocusStates()
}

func (c *Config) focusPrev() {
	panels := []ConfigPanel{PanelWiFi, PanelSystem, PanelCloud, PanelInputs}
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
	}
	return cmd
}

func (c *Config) updateAllComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	var wifiCmd, systemCmd, cloudCmd, inputsCmd tea.Cmd
	c.wifi, wifiCmd = c.wifi.Update(msg)
	c.system, systemCmd = c.system.Update(msg)
	c.cloud, cloudCmd = c.cloud.Update(msg)
	c.inputs, inputsCmd = c.inputs.Update(msg)

	cmds = append(cmds, wifiCmd, systemCmd, cloudCmd, inputsCmd)
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
	default:
		return c.wifi.View()
	}
}

func (c *Config) renderStandardLayout() string {
	// Render panels (components already have embedded titles)
	leftPanels := []string{
		c.wifi.View(),
		c.system.View(),
	}

	rightPanels := []string{
		c.cloud.View(),
		c.inputs.View(),
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
		return c
	}

	// Standard layout: 2-column split
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	panelHeight := height / 2
	contentHeight := panelHeight - 4 // Account for border and title

	// Set sizes for left column components
	c.wifi = c.wifi.SetSize(leftWidth-4, contentHeight)
	c.system = c.system.SetSize(leftWidth-4, height-panelHeight-4)

	// Set sizes for right column components
	c.cloud = c.cloud.SetSize(rightWidth-4, contentHeight)
	c.inputs = c.inputs.SetSize(rightWidth-4, height-panelHeight-4)

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
