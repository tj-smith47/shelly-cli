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
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
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
	helpers.Sizable
	ctx           context.Context
	svc           *shelly.Service
	fileCache     *cache.FileCache
	device        string
	status        *network.WiFiStatusFull
	config        *network.WiFiConfigFull
	networks      []network.WiFiNetworkFull
	loading       bool
	scanning      bool
	editing       bool
	err           error
	focused       bool
	panelIndex    int // 1-based panel index for Shift+N hotkey hint
	styles        Styles
	scannerLoader loading.Model // Extra loader for scanning
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

	m := Model{
		Sizable:     helpers.NewSizable(12, panel.NewScroller(0, 10)),
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		scannerLoader: loading.New(
			loading.WithMessage("Scanning for networks..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		editModal: NewEditModel(deps.Ctx, deps.Svc),
	}
	m.Loader = m.Loader.SetMessage("Loading WiFi status...")
	return m
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
	m.Scroller.SetItemCount(0)
	m.Scroller.CursorToStart()
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
	resized := m.ApplySizeWithExtraLoaders(width, height, m.scannerLoader)
	m.scannerLoader = resized[0]
	m.editModal = m.editModal.SetSize(width, height)
	return m
}

// SetEditModalSize sets the edit modal dimensions.
// This should be called with screen-based dimensions when the modal is visible.
func (m Model) SetEditModalSize(width, height int) Model {
	if m.editing {
		m.editModal = m.editModal.SetSize(width, height)
	}
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
	// Action messages from context system - consolidated for complexity
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.ScanRequestMsg, messages.RefreshRequestMsg, messages.EditRequestMsg:
		return m.handleActionMsg(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleNavigationMsg(msg messages.NavigationMsg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	return m.handleNavigation(msg)
}

func (m Model) handleActionMsg(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	switch msg.(type) {
	case messages.ScanRequestMsg:
		return m.handleScanKey()
	case messages.RefreshRequestMsg:
		return m.handleRefreshKey()
	case messages.EditRequestMsg:
		return m.handleEditKey()
	}
	return m, nil
}

func (m Model) updateLoading(msg tea.Msg) (Model, tea.Cmd, bool) {
	result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
		if _, ok := msg.(StatusLoadedMsg); ok {
			return true
		}
		return generics.IsPanelCacheMsg(msg)
	})
	m.Loader = result.Loader
	return m, result.Cmd, result.Consumed
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

	// Emit StatusLoadedMsg with cached data so sequential loading can advance
	// and handleStatusLoaded won't overwrite with nil
	loadedCmd := func() tea.Msg { return StatusLoadedMsg{Status: m.status, Config: m.config} }

	// Background refresh if stale
	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWiFi {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWiFi {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("wifi background refresh", msg.Err)
		m.err = msg.Err
		// Emit StatusLoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return StatusLoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedWiFiData); ok {
		m.status = data.Status
		m.config = data.Config
	}
	// Emit StatusLoadedMsg so sequential loading can advance
	return m, func() tea.Msg { return StatusLoadedMsg{Status: m.status, Config: m.config} }
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
	m.Scroller.SetItemCount(len(m.networks))
	m.Scroller.CursorToStart()
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
			m.Loader.Tick(),
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
				m.Loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeWiFi),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		m.Scroller.CursorUp()
	case messages.NavDown:
		m.Scroller.CursorDown()
	case messages.NavPageUp:
		m.Scroller.PageUp()
	case messages.NavPageDown:
		m.Scroller.PageDown()
	case messages.NavHome:
		m.Scroller.CursorToStart()
	case messages.NavEnd:
		m.Scroller.CursorToEnd()
	case messages.NavLeft, messages.NavRight:
		// Not applicable for this component
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not covered by action messages
	// (none currently - all keys migrated to action messages)
	_ = msg
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
		m.Loader.Tick(),
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

	r := rendering.New(m.Width, m.Height).
		SetTitle("WiFi").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.device == "" {
		r.SetContent(styles.NoDeviceSelected(m.Width, m.Height))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.Loader.View())
		return r.Render()
	}

	if m.err != nil {
		msg, _ := tuierrors.FormatError(m.err)
		r.SetContent(m.styles.Error.Render(msg))
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

	// Network list with scroll indicator
	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[network.WiFiNetworkFull]{
		Items:    m.networks,
		Scroller: m.Scroller,
		RenderItem: func(netw network.WiFiNetworkFull, _ int, isCursor bool) string {
			return m.renderNetworkLine(netw, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	}))

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

	// Calculate available width for SSID
	// Fixed: selector(2) + signalIcon(2) + space(1) + authIcon(2) = 7
	ssidWidth := output.ContentWidth(m.Width, 4+7)
	if ssidWidth < 10 {
		ssidWidth = 10
	}
	ssid := output.Truncate(netw.SSID, ssidWidth)

	line := fmt.Sprintf("%s%s %-*s %s",
		selector,
		signalStyle.Render(signalIcon),
		ssidWidth,
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
	return m.Scroller.Cursor()
}

// Refresh triggers a refresh of the WiFi status.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchStatus())
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
