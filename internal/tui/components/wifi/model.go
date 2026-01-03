// Package wifi provides TUI components for managing device WiFi settings.
package wifi

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Deps holds the dependencies for the WiFi component.
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

// CachedWiFiData holds WiFi status and config for caching.
type CachedWiFiData struct {
	Status *network.WiFiStatusFull `json:"status"`
	Config *network.WiFiConfigFull `json:"config"`
}

// StatusLoadedMsg signals that WiFi status was loaded.
type StatusLoadedMsg struct {
	Status *network.WiFiStatusFull
	Config *network.WiFiConfigFull
	Err    error
}

// ScanResultMsg signals that WiFi scan completed.
type ScanResultMsg struct {
	Networks []network.WiFiNetworkFull
	Err      error
}

// EditOpenedMsg signals that the edit modal was opened.
type EditOpenedMsg struct{}

// EditClosedMsg signals that the edit modal was closed.
type EditClosedMsg struct {
	Saved bool
}

// Model displays WiFi settings for a device.
type Model struct {
	ctx           context.Context
	svc           *shelly.Service
	fileCache     *cache.FileCache
	device        string
	status        *network.WiFiStatusFull
	config        *network.WiFiConfigFull
	networks      []network.WiFiNetworkFull
	scroller      *panel.Scroller
	loading       bool
	scanning      bool
	editing       bool
	err           error
	width         int
	height        int
	focused       bool
	panelIndex    int // 1-based panel index for Shift+N hotkey hint
	styles        Styles
	loader        loading.Model
	scannerLoader loading.Model
	editModal     EditModel
	cacheStatus   cachestatus.Model
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
			Foreground(colors.Text),
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
		iostreams.DebugErr("wifi component init", err)
		panic(fmt.Sprintf("wifi: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		scroller:    panel.NewScroller(0, 10),
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading WiFi status..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		scannerLoader: loading.New(
			loading.WithMessage("Scanning for networks..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		editModal: NewEditModel(deps.Ctx, deps.Svc),
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
	m.loading = true
	m.cacheStatus = cachestatus.New()

	if device == "" {
		m.loading = false
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeWiFi)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetWiFiStatusFull(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		config, configErr := m.svc.GetWiFiConfigFull(ctx, m.device)
		if configErr != nil {
			return StatusLoadedMsg{Status: status, Err: configErr}
		}

		return StatusLoadedMsg{Status: status, Config: config}
	}
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeWiFi, cache.TTLWiFi, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetWiFiStatusFull(ctx, m.device)
		if err != nil {
			return nil, err
		}

		config, configErr := m.svc.GetWiFiConfigFull(ctx, m.device)
		if configErr != nil {
			return nil, configErr
		}

		return CachedWiFiData{Status: status, Config: config}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeWiFi, cache.TTLWiFi, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetWiFiStatusFull(ctx, m.device)
		if err != nil {
			return nil, err
		}

		config, configErr := m.svc.GetWiFiConfigFull(ctx, m.device)
		if configErr != nil {
			return nil, configErr
		}

		return CachedWiFiData{Status: status, Config: config}, nil
	})
}

func (m Model) scanNetworks() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()

		networks, err := m.svc.ScanWiFiNetworksFull(ctx, m.device)
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
	// Update loader sizes for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	m.scannerLoader = m.scannerLoader.SetSize(width-4, height-4)
	m.editModal = m.editModal.SetSize(width, height)
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

	// Forward tick messages to loaders when loading or scanning
	if m.loading {
		if model, cmd, done := m.updateLoading(msg); done {
			return model, cmd
		}
	}
	if m.scanning {
		if model, cmd, done := m.updateScanning(msg); done {
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
	case ScanResultMsg:
		return m.handleScanResult(msg)
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

func (m Model) updateScanning(msg tea.Msg) (Model, tea.Cmd, bool) {
	var cmd tea.Cmd
	m.scannerLoader, cmd = m.scannerLoader.Update(msg)
	// Continue processing ScanResultMsg even during scanning
	if _, ok := msg.(ScanResultMsg); !ok {
		if cmd != nil {
			return m, cmd, true
		}
	}
	return m, nil, false
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWiFi {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedWiFiData](msg.Data)
	if err == nil {
		m.status = data.Status
		m.config = data.Config
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Background refresh if stale
	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, nil
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWiFi {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWiFi {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("wifi background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedWiFiData); ok {
		m.status = data.Status
		m.config = data.Config
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
	m.status = msg.Status
	m.config = msg.Config
	return m, nil
}

func (m Model) handleScanResult(msg ScanResultMsg) (Model, tea.Cmd) {
	m.scanning = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.networks = msg.Networks
	m.scroller.SetItemCount(len(m.networks))
	m.scroller.CursorToStart()
	return m, nil
}

func (m Model) handleEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Update(msg)

	// Check if modal was closed
	if !m.editModal.IsVisible() {
		m.editing = false
		// Invalidate cache and refresh data after edit
		m.loading = true
		return m, tea.Batch(
			cmd,
			m.loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeWiFi),
			m.fetchAndCacheStatus(),
		)
	}

	// Handle save result message
	if saveMsg, ok := msg.(SaveResultMsg); ok {
		if saveMsg.Success {
			m.editing = false
			m.editModal = m.editModal.Hide()
			// Invalidate cache and refresh data after successful save
			m.loading = true
			return m, tea.Batch(
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeWiFi),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Handle navigation keys first
	if keys.HandleScrollNavigation(msg.String(), m.scroller) {
		return m, nil
	}

	// Handle action keys
	switch msg.String() {
	case "s":
		return m.handleScanKey()
	case "r":
		return m.handleRefreshKey()
	case "e":
		return m.handleEditKey()
	}

	return m, nil
}

func (m Model) handleScanKey() (Model, tea.Cmd) {
	if m.scanning || m.device == "" {
		return m, nil
	}
	m.scanning = true
	m.err = nil
	return m, tea.Batch(m.scannerLoader.Tick(), m.scanNetworks())
}

func (m Model) handleRefreshKey() (Model, tea.Cmd) {
	if m.loading || m.device == "" {
		return m, nil
	}
	m.loading = true
	// Invalidate cache and fetch fresh data
	return m, tea.Batch(
		m.loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeWiFi),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.scanning {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.Show(m.device, m.config, m.networks)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

// View renders the WiFi component.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("WiFi").
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

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := "e:edit s:scan r:refresh"
		if cacheView := m.cacheStatus.View(); cacheView != "" {
			footer += " | " + cacheView
		}
		r.SetFooter(footer)
	}
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
	content.WriteString(m.styles.Label.Render("Configuration:"))
	content.WriteString("\n")

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
		content.WriteString(m.scannerLoader.View())
		return content.String()
	}

	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Available Networks (%d):", len(m.networks))))
	content.WriteString("\n")

	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		netw := m.networks[i]
		isSelected := m.scroller.IsCursorAt(i)

		line := m.renderNetworkLine(netw, isSelected)
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

func (m Model) renderNetworkLine(netw network.WiFiNetworkFull, isSelected bool) string {
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Signal indicator
	signalIcon, signalStyle := m.getSignalIconAndStyle(netw.RSSI)

	// Auth indicator
	authIcon := "🔓"
	if netw.Auth != "open" && netw.Auth != "" {
		authIcon = "🔒"
	}

	ssid := output.Truncate(netw.SSID, 20)

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
func (m Model) Status() *network.WiFiStatusFull {
	return m.status
}

// Config returns the current WiFi config.
func (m Model) Config() *network.WiFiConfigFull {
	return m.config
}

// Networks returns the scanned networks.
func (m Model) Networks() []network.WiFiNetworkFull {
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
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom e:edit s:scan"
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
	return m.editModal.View()
}
