// Package provisioning provides a TUI wizard for provisioning new Shelly devices.
// This guides users through connecting to a device's AP and configuring WiFi.
package provisioning

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Step represents the current wizard step.
type Step int

// Wizard steps.
const (
	StepInstructions Step = iota // Show connection instructions
	StepWaiting                  // Waiting for device connection
	StepCredentials              // Enter WiFi credentials
	StepConfiguring              // Configuring device
	StepSuccess                  // Successfully configured
	StepError                    // Error occurred
)

// DefaultAPAddress is the default IP address for Shelly devices in AP mode.
const DefaultAPAddress = "192.168.33.1"

// Key constants for input handling (single-char keys only, others use keyconst).
const (
	keyBackspace = "backspace"
	keyQ         = "q"
	keyR         = "r"
)

// Deps holds the dependencies for the Provisioning component.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// DeviceFoundMsg signals that a device was found at the AP address.
type DeviceFoundMsg struct {
	DeviceInfo *shelly.ProvisioningDeviceInfo
	Err        error
}

// ConfiguredMsg signals that WiFi was configured on the device.
type ConfiguredMsg struct {
	Err error
}

// PollMsg signals a poll attempt should be made.
type PollMsg struct{}

// Model is the provisioning wizard model.
type Model struct {
	helpers.Sizable
	ctx          context.Context
	svc          *shelly.Service
	step         Step
	deviceInfo   *shelly.ProvisioningDeviceInfo
	ssid         string
	password     string
	inputField   int // 0 = SSID, 1 = password
	err          error
	focused      bool
	panelIndex   int
	polling      bool
	styles       Styles
	configLoader loading.Model // Extra loader for config step
}

// Styles holds styles for the Provisioning component.
type Styles struct {
	Title     lipgloss.Style
	Step      lipgloss.Style
	Text      lipgloss.Style
	Highlight lipgloss.Style
	Input     lipgloss.Style
	Label     lipgloss.Style
	Muted     lipgloss.Style
	Success   lipgloss.Style
	Error     lipgloss.Style
	Warning   lipgloss.Style
}

// DefaultStyles returns the default styles for the Provisioning component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Step: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Text: lipgloss.NewStyle().
			Foreground(colors.Text),
		Highlight: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Input: lipgloss.NewStyle().
			Foreground(colors.Text).
			Background(colors.AltBackground).
			Padding(0, 1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Success: lipgloss.NewStyle().
			Foreground(colors.Online),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
	}
}

// New creates a new Provisioning model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("provisioning component init", err)
		panic(fmt.Sprintf("provisioning: invalid deps: %v", err))
	}

	m := Model{
		Sizable: helpers.NewSizableLoaderOnly(),
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		step:    StepInstructions,
		styles:  DefaultStyles(),
		configLoader: loading.New(
			loading.WithMessage("Please wait..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(false, false),
		),
	}
	m.Loader = m.Loader.SetMessage("Scanning...")
	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Reset resets the wizard to the initial state.
func (m Model) Reset() Model {
	m.step = StepInstructions
	m.deviceInfo = nil
	m.ssid = ""
	m.password = ""
	m.inputField = 0
	m.err = nil
	m.polling = false
	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	resized := m.ApplySizeWithExtraLoaders(width, height, m.configLoader)
	m.configLoader = resized[0]
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward tick messages to loaders when active
	if m, cmd, handled := m.updateLoaders(msg); handled {
		return m, cmd
	}

	return m.handleMessage(msg)
}

func (m Model) updateLoaders(msg tea.Msg) (Model, tea.Cmd, bool) {
	if m.step == StepWaiting {
		var cmd tea.Cmd
		m.Loader, cmd = m.Loader.Update(msg)
		switch msg.(type) {
		case DeviceFoundMsg, PollMsg:
			return m, nil, false // Pass through
		default:
			if cmd != nil {
				return m, cmd, true
			}
		}
	}
	if m.step == StepConfiguring {
		var cmd tea.Cmd
		m.configLoader, cmd = m.configLoader.Update(msg)
		switch msg.(type) {
		case ConfiguredMsg:
			return m, nil, false // Pass through
		default:
			if cmd != nil {
				return m, cmd, true
			}
		}
	}
	return m, nil, false
}

func (m Model) handleMessage(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DeviceFoundMsg:
		return m.handleDeviceFound(msg)
	case ConfiguredMsg:
		return m.handleConfigured(msg)
	case PollMsg:
		if m.step == StepWaiting {
			return m, tea.Batch(m.Loader.Tick(), m.checkDevice())
		}
	case messages.NavigationMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleNavigation(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleDeviceFound(msg DeviceFoundMsg) (Model, tea.Cmd) {
	m.polling = false
	if msg.Err != nil {
		// Keep polling on error
		return m, tea.Batch(m.Loader.Tick(), m.pollAfterDelay())
	}
	m.deviceInfo = msg.DeviceInfo
	m.step = StepCredentials
	return m, nil
}

func (m Model) handleConfigured(msg ConfiguredMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		m.step = StepError
		return m, nil
	}
	m.step = StepSuccess
	return m, nil
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	// Navigation only applicable during credential entry
	if m.step != StepCredentials {
		return m, nil
	}
	switch msg.Direction {
	case messages.NavUp, messages.NavDown:
		// Cycle between SSID (0) and Password (1) fields
		m.inputField = (m.inputField + 1) % 2
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for 2-field form
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch m.step {
	case StepInstructions:
		return m.handleInstructionsKey(msg)
	case StepWaiting:
		// No key handling while waiting
		return m, nil
	case StepCredentials:
		return m.handleCredentialsKey(msg)
	case StepConfiguring:
		// No key handling while configuring
		return m, nil
	case StepSuccess, StepError:
		return m.handleFinalKey(msg)
	}
	return m, nil
}

func (m Model) handleInstructionsKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEnter:
		m.step = StepWaiting
		m.polling = true
		return m, tea.Batch(m.Loader.Tick(), m.checkDevice())
	case keyconst.KeyEsc, keyQ:
		m = m.Reset()
	}
	return m, nil
}

func (m Model) handleCredentialsKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyTab:
		// Tab cycles through fields (NavigationMsg handles j/k/arrows)
		m.inputField = (m.inputField + 1) % 2
	case keyconst.KeyEnter:
		if m.ssid != "" {
			m.step = StepConfiguring
			return m, tea.Batch(m.configLoader.Tick(), m.configureDevice())
		}
	case keyconst.KeyEsc:
		m = m.Reset()
	case keyBackspace:
		m = m.handleBackspace()
	default:
		m = m.handleCharInput(msg.String())
	}
	return m, nil
}

func (m Model) handleBackspace() Model {
	if m.inputField == 0 && m.ssid != "" {
		m.ssid = m.ssid[:len(m.ssid)-1]
	} else if m.inputField == 1 && m.password != "" {
		m.password = m.password[:len(m.password)-1]
	}
	return m
}

func (m Model) handleCharInput(char string) Model {
	if len(char) != 1 {
		return m
	}
	if m.inputField == 0 && len(m.ssid) < 32 {
		m.ssid += char
	} else if m.inputField == 1 && len(m.password) < 64 {
		m.password += char
	}
	return m
}

func (m Model) handleFinalKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case keyconst.KeyEnter, keyconst.KeyEsc, keyQ, keyR:
		m = m.Reset()
	}
	return m, nil
}

func (m Model) checkDevice() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		info, err := m.svc.GetDeviceInfoByAddress(ctx, DefaultAPAddress)
		return DeviceFoundMsg{
			DeviceInfo: info,
			Err:        err,
		}
	}
}

func (m Model) pollAfterDelay() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return PollMsg{}
	})
}

func (m Model) configureDevice() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.ConfigureWiFi(ctx, DefaultAPAddress, m.ssid, m.password)
		return ConfiguredMsg{Err: err}
	}
}

// View renders the Provisioning component.
func (m Model) View() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Setup New Device").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	var content strings.Builder

	// Step indicator
	stepText := m.getStepText()
	content.WriteString(m.styles.Step.Render(stepText))
	content.WriteString("\n\n")

	// Step content
	switch m.step {
	case StepInstructions:
		content.WriteString(m.renderInstructions())
	case StepWaiting:
		content.WriteString(m.renderWaiting())
	case StepCredentials:
		content.WriteString(m.renderCredentials())
	case StepConfiguring:
		content.WriteString(m.renderConfiguring())
	case StepSuccess:
		content.WriteString(m.renderSuccess())
	case StepError:
		content.WriteString(m.renderError())
	}

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) getStepText() string {
	switch m.step {
	case StepInstructions:
		return "Step 1 of 4: Connect to Device"
	case StepWaiting:
		return "Step 2 of 4: Detecting Device"
	case StepCredentials:
		return "Step 3 of 4: Enter WiFi Credentials"
	case StepConfiguring:
		return "Step 4 of 4: Configuring Device"
	case StepSuccess:
		return "Setup Complete"
	case StepError:
		return "Setup Failed"
	default:
		return ""
	}
}

func (m Model) renderInstructions() string {
	var content strings.Builder

	content.WriteString(m.styles.Title.Render("Connect to Device Access Point"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Text.Render("1. Put your Shelly device in AP mode"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("   (Hold button for 10+ seconds until LED blinks)"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Text.Render("2. Connect your computer to the device's WiFi network"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("   Network name: ShellyXXXX-XXXXXXXXXXXX"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Text.Render("3. The device will be available at:"))
	content.WriteString("\n")
	content.WriteString(m.styles.Highlight.Render("   http://192.168.33.1"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Muted.Render("Press Enter when connected, Esc to cancel"))

	return content.String()
}

func (m Model) renderWaiting() string {
	var content strings.Builder

	content.WriteString(m.styles.Title.Render("Looking for Device..."))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Text.Render("Trying to connect to "))
	content.WriteString(m.styles.Highlight.Render(DefaultAPAddress))
	content.WriteString("\n\n")

	content.WriteString(m.Loader.View())
	content.WriteString("\n\n")

	content.WriteString(m.styles.Muted.Render("Make sure you're connected to the device's WiFi network."))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("Press Esc to cancel"))

	return content.String()
}

func (m Model) renderCredentials() string {
	var content strings.Builder

	content.WriteString(m.styles.Title.Render("Device Found!"))
	content.WriteString("\n\n")

	if m.deviceInfo != nil {
		content.WriteString(m.styles.Label.Render("Model: "))
		content.WriteString(m.styles.Text.Render(m.deviceInfo.Model))
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("MAC:   "))
		content.WriteString(m.styles.Text.Render(model.NormalizeMAC(m.deviceInfo.MAC)))
		content.WriteString("\n\n")
	}

	content.WriteString(m.styles.Text.Render("Enter your WiFi credentials:"))
	content.WriteString("\n\n")

	// SSID field
	ssidLabel := m.styles.Label.Render("SSID:     ")
	ssidValue := m.ssid
	if ssidValue == "" {
		ssidValue = " "
	}
	if m.inputField == 0 {
		ssidLabel = m.styles.Highlight.Render("SSID:     ")
	}
	content.WriteString(ssidLabel)
	content.WriteString(m.styles.Input.Render(ssidValue))
	content.WriteString("\n")

	// Password field
	passLabel := m.styles.Label.Render("Password: ")
	passValue := strings.Repeat("*", len(m.password))
	if passValue == "" {
		passValue = " "
	}
	if m.inputField == 1 {
		passLabel = m.styles.Highlight.Render("Password: ")
	}
	content.WriteString(passLabel)
	content.WriteString(m.styles.Input.Render(passValue))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Muted.Render("Tab: switch field | Enter: configure | Esc: cancel"))

	return content.String()
}

func (m Model) renderConfiguring() string {
	var content strings.Builder

	content.WriteString(m.styles.Title.Render("Configuring Device..."))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Text.Render("Setting WiFi network: "))
	content.WriteString(m.styles.Highlight.Render(m.ssid))
	content.WriteString("\n\n")

	content.WriteString(m.configLoader.View())
	content.WriteString("\n\n")

	content.WriteString(m.styles.Muted.Render("The device will restart and connect to your network."))

	return content.String()
}

func (m Model) renderSuccess() string {
	var content strings.Builder

	content.WriteString(m.styles.Success.Render("✓ Device Configured Successfully!"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Text.Render("Your device is now connecting to: "))
	content.WriteString(m.styles.Highlight.Render(m.ssid))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Text.Render("Next steps:"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("1. Reconnect your computer to your main WiFi network"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("2. Use device discovery to find the device"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("3. Add the device to your device list"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Muted.Render("Press Enter or Esc to close"))

	return content.String()
}

func (m Model) renderError() string {
	var content strings.Builder

	content.WriteString(m.styles.Error.Render("✗ Setup Failed"))
	content.WriteString("\n\n")

	if m.err != nil {
		content.WriteString(m.styles.Text.Render("Error: "))
		content.WriteString(m.styles.Error.Render(m.err.Error()))
		content.WriteString("\n\n")
	}

	content.WriteString(m.styles.Text.Render("Troubleshooting:"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("• Check that you're connected to the device's WiFi"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("• Verify the WiFi credentials are correct"))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("• Try power-cycling the device"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Muted.Render("Press Enter or R to retry, Esc to cancel"))

	return content.String()
}

// Step returns the current wizard step.
func (m Model) Step() Step {
	return m.step
}

// DeviceInfo returns the discovered device info.
func (m Model) DeviceInfo() *shelly.ProvisioningDeviceInfo {
	return m.deviceInfo
}

// SSID returns the entered SSID.
func (m Model) SSID() string {
	return m.ssid
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Polling returns whether the component is polling for a device.
func (m Model) Polling() bool {
	return m.polling
}
