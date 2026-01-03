// Package smarthome provides TUI components for managing smart home protocol settings.
// This includes Matter, Zigbee, and LoRa configuration.
package smarthome

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

// Deps holds the dependencies for the SmartHome component.
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

// CachedSmartHomeData holds smart home status for caching.
type CachedSmartHomeData struct {
	Matter *shelly.TUIMatterStatus `json:"matter"`
	Zigbee *shelly.TUIZigbeeStatus `json:"zigbee"`
	LoRa   *shelly.TUILoRaStatus   `json:"lora"`
}

// Protocol identifies which protocol section is focused.
type Protocol int

// Protocol constants.
const (
	ProtocolMatter Protocol = iota // Matter protocol section
	ProtocolZigbee                 // Zigbee protocol section
	ProtocolLoRa                   // LoRa protocol section
)

// StatusLoadedMsg signals that smart home statuses were loaded.
type StatusLoadedMsg struct {
	Matter *shelly.TUIMatterStatus
	Zigbee *shelly.TUIZigbeeStatus
	LoRa   *shelly.TUILoRaStatus
	Err    error
}

// Model displays smart home protocol settings for a device.
type Model struct {
	ctx            context.Context
	svc            *shelly.Service
	fileCache      *cache.FileCache
	device         string
	matter         *shelly.TUIMatterStatus
	zigbee         *shelly.TUIZigbeeStatus
	lora           *shelly.TUILoRaStatus
	activeProtocol Protocol
	loading        bool
	err            error
	width          int
	height         int
	focused        bool
	panelIndex     int // 1-based panel index for Shift+N hotkey hint
	styles         Styles
	loader         loading.Model
	cacheStatus    cachestatus.Model
}

// Styles holds styles for the SmartHome component.
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
	Warning   lipgloss.Style
}

// DefaultStyles returns the default styles for the SmartHome component.
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
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
	}
}

// New creates a new SmartHome model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("smarthome component init", err)
		panic(fmt.Sprintf("smarthome: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading smart home protocols..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display smart home settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.matter = nil
	m.zigbee = nil
	m.lora = nil
	m.activeProtocol = ProtocolMatter
	m.err = nil
	m.cacheStatus = cachestatus.New()

	if device == "" {
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeMatter)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var msg StatusLoadedMsg
		msg.Matter = m.fetchMatter(ctx)
		msg.Zigbee = m.fetchZigbee(ctx)
		msg.LoRa = m.fetchLoRa(ctx)

		return msg
	}
}

func (m Model) fetchMatter(ctx context.Context) *shelly.TUIMatterStatus {
	status, err := m.svc.GetTUIMatterStatus(ctx, m.device)
	if err != nil {
		return nil
	}
	return status
}

func (m Model) fetchZigbee(ctx context.Context) *shelly.TUIZigbeeStatus {
	status, err := m.svc.GetTUIZigbeeStatus(ctx, m.device)
	if err != nil {
		return nil
	}
	return status
}

func (m Model) fetchLoRa(ctx context.Context) *shelly.TUILoRaStatus {
	status, err := m.svc.GetTUILoRaStatus(ctx, m.device)
	if err != nil {
		return nil
	}
	return status
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeMatter, cache.TTLSmartHome, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		data := CachedSmartHomeData{
			Matter: m.fetchMatter(ctx),
			Zigbee: m.fetchZigbee(ctx),
			LoRa:   m.fetchLoRa(ctx),
		}

		return data, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeMatter, cache.TTLSmartHome, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		data := CachedSmartHomeData{
			Matter: m.fetchMatter(ctx),
			Zigbee: m.fetchZigbee(ctx),
			LoRa:   m.fetchLoRa(ctx),
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
	if msg.Device != m.device || msg.DataType != cache.TypeMatter {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedSmartHomeData](msg.Data)
	if err == nil {
		m.matter = data.Matter
		m.zigbee = data.Zigbee
		m.lora = data.LoRa
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, nil
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeMatter {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeMatter {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("smarthome background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedSmartHomeData); ok {
		m.matter = data.Matter
		m.zigbee = data.Zigbee
		m.lora = data.LoRa
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
	m.matter = msg.Matter
	m.zigbee = msg.Zigbee
	m.lora = msg.LoRa
	return m, nil
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
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
				m.fetchAndCacheStatus(),
			)
		}
	case "1":
		m.activeProtocol = ProtocolMatter
	case "2":
		m.activeProtocol = ProtocolZigbee
	case "3":
		m.activeProtocol = ProtocolLoRa
	}

	return m, nil
}

func (m Model) nextProtocol() Model {
	switch m.activeProtocol {
	case ProtocolMatter:
		m.activeProtocol = ProtocolZigbee
	case ProtocolZigbee:
		m.activeProtocol = ProtocolLoRa
	case ProtocolLoRa:
		m.activeProtocol = ProtocolMatter
	}
	return m
}

func (m Model) prevProtocol() Model {
	switch m.activeProtocol {
	case ProtocolMatter:
		m.activeProtocol = ProtocolLoRa
	case ProtocolZigbee:
		m.activeProtocol = ProtocolMatter
	case ProtocolLoRa:
		m.activeProtocol = ProtocolZigbee
	}
	return m
}

// View renders the SmartHome component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Smart Home").
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

	// Matter Section
	content.WriteString(m.renderMatter())
	content.WriteString("\n\n")

	// Zigbee Section
	content.WriteString(m.renderZigbee())
	content.WriteString("\n\n")

	// LoRa Section
	content.WriteString(m.renderLoRa())

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := "1-3:sel j/k:nav r:refresh"
		if cs := m.cacheStatus.View(); cs != "" {
			footer = cs + " " + footer
		}
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) renderMatter() string {
	var content strings.Builder

	// Section header
	header := "Matter"
	if m.activeProtocol == ProtocolMatter {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.matter == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported (Gen4+ only)"))
		return content.String()
	}

	// Enabled status
	if m.matter.Enabled {
		content.WriteString("    " + m.styles.Enabled.Render("● Enabled") + "\n")
	} else {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled") + "\n")
		return content.String()
	}

	// Commissionable status
	content.WriteString("    " + m.styles.Label.Render("Commission: "))
	if m.matter.Commissionable {
		content.WriteString(m.styles.Warning.Render("Ready to pair"))
	} else {
		content.WriteString(m.styles.Muted.Render("Already paired"))
	}
	content.WriteString("\n")

	// Fabrics count
	content.WriteString("    " + m.styles.Label.Render("Fabrics:    "))
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.matter.FabricsCount)))

	return content.String()
}

func (m Model) renderZigbee() string {
	var content strings.Builder

	// Section header
	header := "Zigbee"
	if m.activeProtocol == ProtocolZigbee {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.zigbee == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported"))
		return content.String()
	}

	// Enabled status
	if m.zigbee.Enabled {
		content.WriteString("    " + m.styles.Enabled.Render("● Enabled") + "\n")
	} else {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled") + "\n")
		return content.String()
	}

	// Network state
	content.WriteString("    " + m.styles.Label.Render("State:   "))
	switch m.zigbee.NetworkState {
	case "joined":
		content.WriteString(m.styles.Enabled.Render("Joined"))
	case "steering":
		content.WriteString(m.styles.Warning.Render("Searching..."))
	case "ready":
		content.WriteString(m.styles.Muted.Render("Ready"))
	default:
		content.WriteString(m.styles.Muted.Render(m.zigbee.NetworkState))
	}
	content.WriteString("\n")

	// Channel and PAN ID
	if m.zigbee.NetworkState == "joined" {
		content.WriteString("    " + m.styles.Label.Render("Channel: "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.zigbee.Channel)))
		content.WriteString("\n")

		content.WriteString("    " + m.styles.Label.Render("PAN ID:  "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("0x%04X", m.zigbee.PANID)))
	}

	return content.String()
}

func (m Model) renderLoRa() string {
	var content strings.Builder

	// Section header
	header := "LoRa"
	if m.activeProtocol == ProtocolLoRa {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.lora == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported (add-on required)"))
		return content.String()
	}

	// Enabled status
	if m.lora.Enabled {
		content.WriteString("    " + m.styles.Enabled.Render("● Enabled") + "\n")
	} else {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled") + "\n")
		return content.String()
	}

	// Frequency
	content.WriteString("    " + m.styles.Label.Render("Freq:   "))
	freqMHz := float64(m.lora.Frequency) / 1000000
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%.2f MHz", freqMHz)))
	content.WriteString("\n")

	// TX Power
	content.WriteString("    " + m.styles.Label.Render("Power:  "))
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d dBm", m.lora.TxPower)))

	// Last RSSI/SNR if available
	if m.lora.RSSI != 0 {
		content.WriteString("\n    " + m.styles.Label.Render("Last RSSI: "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d dBm", m.lora.RSSI)))
		content.WriteString(" | SNR: ")
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%.1f dB", m.lora.SNR)))
	}

	return content.String()
}

// Matter returns the current Matter status.
func (m Model) Matter() *shelly.TUIMatterStatus {
	return m.matter
}

// Zigbee returns the current Zigbee status.
func (m Model) Zigbee() *shelly.TUIZigbeeStatus {
	return m.zigbee
}

// LoRa returns the current LoRa status.
func (m Model) LoRa() *shelly.TUILoRaStatus {
	return m.lora
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

// Refresh triggers a refresh of the smart home data.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}
