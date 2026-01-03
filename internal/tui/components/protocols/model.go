// Package protocols provides TUI components for managing device protocol settings.
// This includes MQTT, Modbus, and Ethernet configuration.
package protocols

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Deps holds the dependencies for the Protocols component.
type Deps struct {
	Ctx       context.Context
	Svc       *shelly.Service
	FileCache *cache.FileCache
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	// FileCache is optional - caching is disabled if nil
	return nil
}

// CachedProtocolsData holds protocol status for caching.
type CachedProtocolsData struct {
	MQTT     *MQTTData     `json:"mqtt"`
	Modbus   *ModbusData   `json:"modbus"`
	Ethernet *EthernetData `json:"ethernet"`
}

// Protocol identifies which protocol section is focused.
type Protocol int

// Protocol constants.
const (
	ProtocolMQTT     Protocol = iota // MQTT protocol section
	ProtocolModbus                   // Modbus TCP section
	ProtocolEthernet                 // Ethernet section
)

// StatusLoadedMsg signals that protocol statuses were loaded.
type StatusLoadedMsg struct {
	MQTT     *MQTTData
	Modbus   *ModbusData
	Ethernet *EthernetData
	Err      error
}

// MQTTData holds MQTT status and configuration.
type MQTTData struct {
	Connected   bool
	Enable      bool
	Server      string
	User        string
	ClientID    string
	TopicPrefix string
	SSLCA       string
}

// ModbusData holds Modbus status and configuration.
type ModbusData struct {
	Enabled bool
	Enable  bool
}

// EthernetData holds Ethernet status and configuration.
type EthernetData struct {
	IP         string
	Enable     bool
	IPv4Mode   string
	StaticIP   string
	Netmask    string
	Gateway    string
	Nameserver string
}

// Model displays protocol settings for a device.
type Model struct {
	ctx            context.Context
	svc            *shelly.Service
	fileCache      *cache.FileCache
	device         string
	mqtt           *MQTTData
	modbus         *ModbusData
	ethernet       *EthernetData
	activeProtocol Protocol
	loading        bool
	editing        bool
	err            error
	width          int
	height         int
	focused        bool
	panelIndex     int // 1-based panel index for Shift+N hotkey hint
	styles         Styles
	loader         loading.Model
	mqttEdit       MQTTEditModel
	cacheStatus    cachestatus.Model
}

// Styles holds styles for the Protocols component.
type Styles struct {
	Enabled   lipgloss.Style
	Disabled  lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Highlight lipgloss.Style
	Error     lipgloss.Style
	Muted     lipgloss.Style
	Section   lipgloss.Style
	Active    lipgloss.Style
}

// DefaultStyles returns the default styles for the Protocols component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Highlight: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Section: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Active: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new Protocols model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("protocols component init", err)
		panic(fmt.Sprintf("protocols: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading protocols..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		mqttEdit: NewMQTTEditModel(deps.Ctx, deps.Svc),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display protocol settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.mqtt = nil
	m.modbus = nil
	m.ethernet = nil
	m.activeProtocol = ProtocolMQTT
	m.err = nil
	m.cacheStatus = cachestatus.New()

	if device == "" {
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeMQTT)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var msg StatusLoadedMsg
		msg.MQTT = m.fetchMQTT(ctx)
		msg.Modbus = m.fetchModbus(ctx)
		msg.Ethernet = m.fetchEthernet(ctx)

		// If we got nothing at all, set an error
		if msg.MQTT == nil && msg.Modbus == nil && msg.Ethernet == nil {
			msg.Err = fmt.Errorf("no protocol data available")
		}

		return msg
	}
}

func (m Model) fetchMQTT(ctx context.Context) *MQTTData {
	status, err := m.svc.GetMQTTStatus(ctx, m.device)
	if err != nil {
		return nil
	}

	config, err := m.svc.GetMQTTConfig(ctx, m.device)
	if err != nil {
		return nil
	}

	data := &MQTTData{Connected: status.Connected}
	if v, ok := config["enable"].(bool); ok {
		data.Enable = v
	}
	if v, ok := config["server"].(string); ok {
		data.Server = v
	}
	if v, ok := config["user"].(string); ok {
		data.User = v
	}
	if v, ok := config["client_id"].(string); ok {
		data.ClientID = v
	}
	if v, ok := config["topic_prefix"].(string); ok {
		data.TopicPrefix = v
	}
	if v, ok := config["ssl_ca"].(string); ok {
		data.SSLCA = v
	}
	return data
}

func (m Model) fetchModbus(ctx context.Context) *ModbusData {
	status, err := m.svc.GetModbusStatus(ctx, m.device)
	if err != nil {
		return nil
	}

	config, err := m.svc.GetModbusConfig(ctx, m.device)
	if err != nil {
		return nil
	}

	data := &ModbusData{Enabled: status.Enabled}
	if v, ok := config["enable"].(bool); ok {
		data.Enable = v
	}
	return data
}

func (m Model) fetchEthernet(ctx context.Context) *EthernetData {
	status, err := m.svc.GetEthernetStatus(ctx, m.device)
	if err != nil {
		return nil
	}

	config, err := m.svc.GetEthernetConfig(ctx, m.device)
	if err != nil {
		return nil
	}

	data := &EthernetData{IP: status.IP}
	if v, ok := config["enable"].(bool); ok {
		data.Enable = v
	}
	if v, ok := config["ipv4mode"].(string); ok {
		data.IPv4Mode = v
	}
	if v, ok := config["ip"].(string); ok {
		data.StaticIP = v
	}
	if v, ok := config["netmask"].(string); ok {
		data.Netmask = v
	}
	if v, ok := config["gw"].(string); ok {
		data.Gateway = v
	}
	if v, ok := config["nameserver"].(string); ok {
		data.Nameserver = v
	}
	return data
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeMQTT, cache.TTLProtocols, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var data CachedProtocolsData
		data.MQTT = m.fetchMQTT(ctx)
		data.Modbus = m.fetchModbus(ctx)
		data.Ethernet = m.fetchEthernet(ctx)

		// If we got nothing at all, return an error
		if data.MQTT == nil && data.Modbus == nil && data.Ethernet == nil {
			return nil, fmt.Errorf("no protocol data available")
		}

		return data, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeMQTT, cache.TTLProtocols, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var data CachedProtocolsData
		data.MQTT = m.fetchMQTT(ctx)
		data.Modbus = m.fetchModbus(ctx)
		data.Ethernet = m.fetchEthernet(ctx)

		// If we got nothing at all, return an error
		if data.MQTT == nil && data.Modbus == nil && data.Ethernet == nil {
			return nil, fmt.Errorf("no protocol data available")
		}

		return data, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	// Update edit modal size
	m.mqttEdit = m.mqttEdit.SetSize(width, height)
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
	// Handle edit modal if visible
	if m.editing {
		return m.handleEditModalUpdate(msg)
	}

	// Forward tick messages to loader when loading
	if m.loading {
		if model, cmd, done := m.updateLoading(msg); done {
			return model, cmd
		}
	}

	// Update cache status spinner
	if m.cacheStatus.IsRefreshing() {
		var cmd tea.Cmd
		m.cacheStatus, cmd = m.cacheStatus.Update(msg)
		if cmd != nil {
			return m, cmd
		}
	}

	return m.handleMessage(msg)
}

func (m Model) handleMessage(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case panelcache.CacheHitMsg:
		return m.handleCacheHit(msg)
	case panelcache.CacheMissMsg:
		return m.handleCacheMiss(msg)
	case panelcache.RefreshCompleteMsg:
		return m.handleRefreshComplete(msg)
	case StatusLoadedMsg:
		return m.handleStatusLoaded(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) updateLoading(msg tea.Msg) (Model, tea.Cmd, bool) {
	var cmd tea.Cmd
	m.loader, cmd = m.loader.Update(msg)
	// Continue processing these messages even during loading
	switch msg.(type) {
	case StatusLoadedMsg, panelcache.CacheHitMsg, panelcache.CacheMissMsg, panelcache.RefreshCompleteMsg:
		return m, nil, false
	default:
		if cmd != nil {
			return m, cmd, true
		}
	}
	return m, nil, false
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeMQTT {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedProtocolsData](msg.Data)
	if err == nil {
		m.mqtt = data.MQTT
		m.modbus = data.Modbus
		m.ethernet = data.Ethernet
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, nil
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeMQTT {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeMQTT {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("protocols background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedProtocolsData); ok {
		m.mqtt = data.MQTT
		m.modbus = data.Modbus
		m.ethernet = data.Ethernet
	}
	return m, nil
}

func (m Model) handleStatusLoaded(msg StatusLoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.mqtt = msg.MQTT
	m.modbus = msg.Modbus
	m.ethernet = msg.Ethernet
	return m, nil
}

func (m Model) handleEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.mqttEdit, cmd = m.mqttEdit.Update(msg)

	// Check if modal was closed
	if !m.mqttEdit.IsVisible() {
		m.editing = false
		// Invalidate cache and refresh data after edit
		m.loading = true
		return m, tea.Batch(
			cmd,
			m.loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeMQTT),
			m.fetchAndCacheStatus(),
		)
	}

	// Handle save result message
	if saveMsg, ok := msg.(MQTTEditSaveResultMsg); ok {
		if saveMsg.Err == nil {
			m.editing = false
			m.mqttEdit = m.mqttEdit.Hide()
			// Invalidate cache and refresh data after successful save
			m.loading = true
			return m, tea.Batch(
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeMQTT),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return MQTTEditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m = m.nextProtocol()
	case "k", "up":
		m = m.prevProtocol()
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			// Invalidate cache and fetch fresh data
			return m, tea.Batch(
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeMQTT),
				m.fetchAndCacheStatus(),
			)
		}
	case "1":
		m.activeProtocol = ProtocolMQTT
	case "2":
		m.activeProtocol = ProtocolModbus
	case "3":
		m.activeProtocol = ProtocolEthernet
	case "m":
		// Open MQTT configuration modal when MQTT protocol is selected
		return m.handleMQTTEditKey()
	}

	return m, nil
}

func (m Model) handleMQTTEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.mqtt == nil {
		return m, nil
	}
	// Only allow editing when MQTT protocol is selected
	if m.activeProtocol != ProtocolMQTT {
		return m, nil
	}
	m.editing = true
	m.mqttEdit = m.mqttEdit.SetSize(m.width, m.height)
	m.mqttEdit = m.mqttEdit.Show(m.device, m.mqtt)
	return m, func() tea.Msg { return MQTTEditOpenedMsg{} }
}

func (m Model) nextProtocol() Model {
	switch m.activeProtocol {
	case ProtocolMQTT:
		m.activeProtocol = ProtocolModbus
	case ProtocolModbus:
		m.activeProtocol = ProtocolEthernet
	case ProtocolEthernet:
		m.activeProtocol = ProtocolMQTT
	}
	return m
}

func (m Model) prevProtocol() Model {
	switch m.activeProtocol {
	case ProtocolMQTT:
		m.activeProtocol = ProtocolEthernet
	case ProtocolModbus:
		m.activeProtocol = ProtocolMQTT
	case ProtocolEthernet:
		m.activeProtocol = ProtocolModbus
	}
	return m
}

// View renders the Protocols component.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.mqttEdit.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Protocols").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.loader.View())
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	var content strings.Builder

	// MQTT Section
	content.WriteString(m.renderMQTT())
	content.WriteString("\n\n")

	// Modbus Section
	content.WriteString(m.renderModbus())
	content.WriteString("\n\n")

	// Ethernet Section
	content.WriteString(m.renderEthernet())

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		var footer string
		if m.activeProtocol == ProtocolMQTT && m.mqtt != nil {
			footer = "1-3:sel j/k:nav m:mqtt r:refresh"
		} else {
			footer = "1-3:sel j/k:nav r:refresh"
		}
		if cs := m.cacheStatus.View(); cs != "" {
			footer = cs + " " + footer
		}
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) renderMQTT() string {
	var content strings.Builder

	// Section header
	header := "MQTT"
	if m.activeProtocol == ProtocolMQTT {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.mqtt == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported"))
		return content.String()
	}

	// Status
	status := m.styles.Disabled.Render("○ Disabled")
	if m.mqtt.Enable {
		if m.mqtt.Connected {
			status = m.styles.Enabled.Render("● Connected")
		} else {
			status = m.styles.Muted.Render("◐ Enabled (disconnected)")
		}
	}
	content.WriteString("    " + status + "\n")

	if m.mqtt.Enable {
		if m.mqtt.Server != "" {
			content.WriteString("    " + m.styles.Label.Render("Server: "))
			content.WriteString(m.styles.Value.Render(m.mqtt.Server) + "\n")
		}
		if m.mqtt.User != "" {
			content.WriteString("    " + m.styles.Label.Render("User:   "))
			content.WriteString(m.styles.Value.Render(m.mqtt.User) + "\n")
		}
		if m.mqtt.TopicPrefix != "" {
			content.WriteString("    " + m.styles.Label.Render("Prefix: "))
			content.WriteString(m.styles.Value.Render(m.mqtt.TopicPrefix))
		}
	}

	return content.String()
}

func (m Model) renderModbus() string {
	var content strings.Builder

	// Section header
	header := "Modbus TCP"
	if m.activeProtocol == ProtocolModbus {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.modbus == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported"))
		return content.String()
	}

	// Status
	if m.modbus.Enable || m.modbus.Enabled {
		content.WriteString("    " + m.styles.Enabled.Render("● Enabled (port 502)"))
	} else {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled"))
	}

	return content.String()
}

func (m Model) renderEthernet() string {
	var content strings.Builder

	// Section header
	header := "Ethernet"
	if m.activeProtocol == ProtocolEthernet {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.ethernet == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported"))
		return content.String()
	}

	// Status
	if !m.ethernet.Enable {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled"))
		return content.String()
	}

	if m.ethernet.IP != "" {
		content.WriteString("    " + m.styles.Enabled.Render("● Connected") + "\n")
		content.WriteString("    " + m.styles.Label.Render("IP:   "))
		content.WriteString(m.styles.Value.Render(m.ethernet.IP) + "\n")
	} else {
		content.WriteString("    " + m.styles.Muted.Render("◐ Enabled (no link)") + "\n")
	}

	content.WriteString("    " + m.styles.Label.Render("Mode: "))
	mode := m.ethernet.IPv4Mode
	if mode == "" {
		mode = "dhcp"
	}
	content.WriteString(m.styles.Value.Render(mode))

	if mode == "static" {
		if m.ethernet.StaticIP != "" {
			content.WriteString("\n    " + m.styles.Label.Render("Static IP: "))
			content.WriteString(m.styles.Value.Render(m.ethernet.StaticIP))
		}
		if m.ethernet.Gateway != "" {
			content.WriteString("\n    " + m.styles.Label.Render("Gateway:   "))
			content.WriteString(m.styles.Value.Render(m.ethernet.Gateway))
		}
	}

	return content.String()
}

// MQTT returns the current MQTT data.
func (m Model) MQTT() *MQTTData {
	return m.mqtt
}

// Modbus returns the current Modbus data.
func (m Model) Modbus() *ModbusData {
	return m.modbus
}

// Ethernet returns the current Ethernet data.
func (m Model) Ethernet() *EthernetData {
	return m.ethernet
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// ActiveProtocol returns the currently selected protocol.
func (m Model) ActiveProtocol() Protocol {
	return m.activeProtocol
}

// Refresh triggers a refresh of the protocol data.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

// IsEditing returns whether the edit modal is open.
func (m Model) IsEditing() bool {
	return m.editing
}

// RenderEditModal returns the edit modal view for full-screen overlay rendering.
func (m Model) RenderEditModal() string {
	if !m.editing {
		return ""
	}
	return m.mqttEdit.View()
}
