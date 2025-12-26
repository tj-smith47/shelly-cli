// Package wifi provides TUI components for managing device WiFi settings.
package wifi

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Deps holds the dependencies for the WiFi component.
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

// StatusLoadedMsg signals that WiFi status was loaded.
type StatusLoadedMsg struct {
	Status *shelly.WifiStatus
	Config *shelly.WifiConfig
	Err    error
}

// ScanResultMsg signals that WiFi scan completed.
type ScanResultMsg struct {
	Networks []shelly.WifiNetwork
	Err      error
}

// Model displays WiFi settings for a device.
type Model struct {
	ctx        context.Context
	svc        *shelly.Service
	device     string
	status     *shelly.WifiStatus
	config     *shelly.WifiConfig
	networks   []shelly.WifiNetwork
	scroller   *panel.Scroller
	loading    bool
	scanning   bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     Styles
}

// Styles holds styles for the WiFi component.
type Styles struct {
	Connected    lipgloss.Style
	Disconnected lipgloss.Style
	SSID         lipgloss.Style
	Signal       lipgloss.Style
	SignalWeak   lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Selected     lipgloss.Style
	Error        lipgloss.Style
	Muted        lipgloss.Style
}

// DefaultStyles returns the default styles for the WiFi component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Connected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disconnected: lipgloss.NewStyle().
			Foreground(colors.Offline),
		SSID: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Signal: lipgloss.NewStyle().
			Foreground(colors.Online),
		SignalWeak: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// New creates a new WiFi model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("wifi: invalid deps: %v", err))
	}

	return Model{
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scroller: panel.NewScroller(0, 10),
		loading:  false,
		styles:   DefaultStyles(),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display WiFi settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.status = nil
	m.config = nil
	m.networks = nil
	m.scroller.SetItemCount(0)
	m.scroller.CursorToStart()
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, m.fetchStatus()
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetWifiStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		config, configErr := m.svc.GetWifiConfig(ctx, m.device)
		if configErr != nil {
			return StatusLoadedMsg{Status: status, Err: configErr}
		}

		return StatusLoadedMsg{Status: status, Config: config}
	}
}

func (m Model) scanNetworks() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()

		networks, err := m.svc.ScanWifiNetworks(ctx, m.device)
		return ScanResultMsg{Networks: networks, Err: err}
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	visibleRows := height - 12 // Reserve space for status section
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.status = msg.Status
		m.config = msg.Config
		return m, nil

	case ScanResultMsg:
		m.scanning = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.networks = msg.Networks
		m.scroller.SetItemCount(len(m.networks))
		m.scroller.CursorToStart()
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.scroller.CursorDown()
	case "k", "up":
		m.scroller.CursorUp()
	case "g":
		m.scroller.CursorToStart()
	case "G":
		m.scroller.CursorToEnd()
	case "ctrl+d", "pgdown":
		m.scroller.PageDown()
	case "ctrl+u", "pgup":
		m.scroller.PageUp()
	case "s":
		if !m.scanning && m.device != "" {
			m.scanning = true
			m.err = nil
			return m, m.scanNetworks()
		}
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			return m, m.fetchStatus()
		}
	}

	return m, nil
}

// View renders the WiFi component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("WiFi").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading WiFi status..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	var content strings.Builder

	// Current connection status
	content.WriteString(m.renderStatus())
	content.WriteString("\n\n")

	// Configuration summary
	content.WriteString(m.renderConfig())

	// Scanned networks
	if len(m.networks) > 0 || m.scanning {
		content.WriteString("\n\n")
		content.WriteString(m.renderNetworks())
	}

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("s: scan | r: refresh"))

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderStatus() string {
	var content strings.Builder

	if m.status == nil {
		return m.styles.Muted.Render("No status available")
	}

	// Connection status
	switch m.status.Status {
	case "got ip":
		content.WriteString(m.styles.Connected.Render("● Connected"))
	case "disconnected":
		content.WriteString(m.styles.Disconnected.Render("○ Disconnected"))
	default:
		content.WriteString(m.styles.Muted.Render("◐ " + m.status.Status))
	}
	content.WriteString("\n")

	// SSID
	if m.status.SSID != "" {
		content.WriteString(m.styles.Label.Render("SSID:    "))
		content.WriteString(m.styles.SSID.Render(m.status.SSID))
		content.WriteString("\n")
	}

	// IP Address
	if m.status.StaIP != "" {
		content.WriteString(m.styles.Label.Render("IP:      "))
		content.WriteString(m.styles.Value.Render(m.status.StaIP))
		content.WriteString("\n")
	}

	// Signal strength
	if m.status.RSSI != 0 {
		content.WriteString(m.styles.Label.Render("Signal:  "))
		content.WriteString(m.renderSignalStrength(m.status.RSSI))
		content.WriteString("\n")
	}

	// AP client count
	if m.status.APClientCount > 0 {
		content.WriteString(m.styles.Label.Render("AP Clients: "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.status.APClientCount)))
	}

	return content.String()
}

func (m Model) renderSignalStrength(rssi float64) string {
	signalStr := fmt.Sprintf("%.0f dBm", rssi)
	switch {
	case rssi > -50:
		return m.styles.Signal.Render(signalStr + " (excellent)")
	case rssi > -60:
		return m.styles.Signal.Render(signalStr + " (good)")
	case rssi > -70:
		return m.styles.SignalWeak.Render(signalStr + " (fair)")
	default:
		return m.styles.SignalWeak.Render(signalStr + " (weak)")
	}
}

func (m Model) getSignalIconAndStyle(rssi float64) (string, lipgloss.Style) {
	switch {
	case rssi > -50:
		return "████", m.styles.Signal
	case rssi > -60:
		return "███░", m.styles.Signal
	case rssi > -70:
		return "██░░", m.styles.SignalWeak
	default:
		return "█░░░", m.styles.SignalWeak
	}
}

func (m Model) renderConfig() string {
	if m.config == nil {
		return ""
	}

	var content strings.Builder
	content.WriteString(m.styles.Label.Render("Configuration:\n"))

	// Station config
	if m.config.STA != nil {
		sta := m.config.STA
		enabled := "disabled"
		if sta.Enabled {
			enabled = "enabled"
		}
		content.WriteString(fmt.Sprintf("  STA: %s (%s)\n", sta.SSID, enabled))
	}

	// AP config
	if m.config.AP != nil {
		ap := m.config.AP
		enabled := "disabled"
		if ap.Enabled {
			enabled = "enabled"
		}
		content.WriteString(fmt.Sprintf("  AP:  %s (%s)\n", ap.SSID, enabled))
	}

	return content.String()
}

func (m Model) renderNetworks() string {
	var content strings.Builder

	if m.scanning {
		content.WriteString(m.styles.Muted.Render("Scanning for networks..."))
		return content.String()
	}

	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Available Networks (%d):\n", len(m.networks))))

	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		network := m.networks[i]
		isSelected := m.scroller.IsCursorAt(i)

		line := m.renderNetworkLine(network, isSelected)
		content.WriteString(line)
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

	return content.String()
}

func (m Model) renderNetworkLine(network shelly.WifiNetwork, isSelected bool) string {
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Signal indicator
	signalIcon, signalStyle := m.getSignalIconAndStyle(network.RSSI)

	// Auth indicator
	authIcon := "🔓"
	if network.Auth != "open" && network.Auth != "" {
		authIcon = "🔒"
	}

	ssid := network.SSID
	if len(ssid) > 20 {
		ssid = ssid[:17] + "..."
	}

	line := fmt.Sprintf("%s%s %-20s %s",
		selector,
		signalStyle.Render(signalIcon),
		m.styles.SSID.Render(ssid),
		authIcon,
	)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// Status returns the current WiFi status.
func (m Model) Status() *shelly.WifiStatus {
	return m.status
}

// Config returns the current WiFi config.
func (m Model) Config() *shelly.WifiConfig {
	return m.config
}

// Networks returns the scanned networks.
func (m Model) Networks() []shelly.WifiNetwork {
	return m.networks
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Scanning returns whether a scan is in progress.
func (m Model) Scanning() bool {
	return m.scanning
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.scroller.Cursor()
}

// Refresh triggers a refresh of the WiFi status.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, m.fetchStatus()
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom s:scan"
}
