// Package ble provides TUI components for managing device Bluetooth settings.
// This includes BLE configuration and BTHome device management.
package ble

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
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Deps holds the dependencies for the BLE component.
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

// CachedBLEData holds BLE status for caching.
type CachedBLEData struct {
	BLE       *shelly.BLEConfig       `json:"ble"`
	Discovery *shelly.BTHomeDiscovery `json:"discovery"`
}

// StatusLoadedMsg signals that BLE status was loaded.
type StatusLoadedMsg struct {
	BLE       *shelly.BLEConfig
	Discovery *shelly.BTHomeDiscovery
	Err       error
}

// DiscoveryStartedMsg signals that BTHome discovery was started.
type DiscoveryStartedMsg struct {
	Err error
}

// Model displays BLE and BTHome settings for a device.
type Model struct {
	ctx         context.Context
	svc         *shelly.Service
	fileCache   *cache.FileCache
	device      string
	ble         *shelly.BLEConfig
	discovery   *shelly.BTHomeDiscovery
	loading     bool
	starting    bool
	editing     bool
	err         error
	width       int
	height      int
	focused     bool
	panelIndex  int // 1-based panel index for Shift+N hotkey hint
	styles      Styles
	loader      loading.Model
	editModal   EditModel
	cacheStatus cachestatus.Model
}

// Styles holds styles for the BLE component.
type Styles struct {
	Enabled   lipgloss.Style
	Disabled  lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Highlight lipgloss.Style
	Error     lipgloss.Style
	Muted     lipgloss.Style
	Section   lipgloss.Style
	Warning   lipgloss.Style
}

// DefaultStyles returns the default styles for the BLE component.
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
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
	}
}

// New creates a new BLE model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("ble component init", err)
		panic(fmt.Sprintf("ble: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading Bluetooth settings..."),
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

// SetDevice sets the device to display BLE settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.ble = nil
	m.discovery = nil
	m.err = nil
	m.loading = true
	m.cacheStatus = cachestatus.New()

	if device == "" {
		m.loading = false
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeBLE)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var msg StatusLoadedMsg

		// Fetch BLE config
		bleConfig, err := m.svc.GetBLEConfig(ctx, m.device)
		if err == nil {
			msg.BLE = bleConfig
		}

		// Fetch BTHome status
		discovery, err := m.svc.GetBTHomeStatus(ctx, m.device)
		if err == nil {
			msg.Discovery = discovery
		}

		// If we got nothing, set an error
		if msg.BLE == nil && msg.Discovery == nil {
			msg.Err = fmt.Errorf("BLE not supported on this device")
		}

		return msg
	}
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeBLE, cache.TTLBLE, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var data CachedBLEData

		// Fetch BLE config
		bleConfig, err := m.svc.GetBLEConfig(ctx, m.device)
		if err == nil {
			data.BLE = bleConfig
		}

		// Fetch BTHome status
		discovery, err := m.svc.GetBTHomeStatus(ctx, m.device)
		if err == nil {
			data.Discovery = discovery
		}

		// If we got nothing, return an error
		if data.BLE == nil && data.Discovery == nil {
			return nil, fmt.Errorf("BLE not supported on this device")
		}

		return data, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeBLE, cache.TTLBLE, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var data CachedBLEData

		// Fetch BLE config
		bleConfig, err := m.svc.GetBLEConfig(ctx, m.device)
		if err == nil {
			data.BLE = bleConfig
		}

		// Fetch BTHome status
		discovery, err := m.svc.GetBTHomeStatus(ctx, m.device)
		if err == nil {
			data.Discovery = discovery
		}

		// If we got nothing, return an error
		if data.BLE == nil && data.Discovery == nil {
			return nil, fmt.Errorf("BLE not supported on this device")
		}

		return data, nil
	})
}

func (m Model) startDiscovery() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		err := m.svc.StartBTHomeDiscovery(ctx, m.device, 30)
		return DiscoveryStartedMsg{Err: err}
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
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
	case DiscoveryStartedMsg:
		return m.handleDiscoveryStarted(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) updateLoading(msg tea.Msg) (Model, tea.Cmd, bool) {
	result := generics.UpdateLoader(m.loader, msg, func(msg tea.Msg) bool {
		switch msg.(type) {
		case StatusLoadedMsg, DiscoveryStartedMsg:
			return true
		}
		return generics.IsPanelCacheMsg(msg)
	})
	m.loader = result.Loader
	return m, result.Cmd, result.Consumed
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeBLE {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedBLEData](msg.Data)
	if err == nil {
		m.ble = data.BLE
		m.discovery = data.Discovery
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit StatusLoadedMsg so sequential loading in Config view can advance
	loadedCmd := func() tea.Msg { return StatusLoadedMsg{} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeBLE {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeBLE {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("ble background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedBLEData); ok {
		m.ble = data.BLE
		m.discovery = data.Discovery
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
	m.ble = msg.BLE
	m.discovery = msg.Discovery
	return m, nil
}

func (m Model) handleDiscoveryStarted(msg DiscoveryStartedMsg) (Model, tea.Cmd) {
	m.starting = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh to see discovery status
	m.loading = true
	return m, tea.Batch(
		m.loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Update(msg)

	// Check if modal was closed
	if !m.editModal.Visible() {
		m.editing = false
		// Invalidate cache and refresh data after edit
		m.loading = true
		return m, tea.Batch(
			cmd,
			m.loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
			m.fetchAndCacheStatus(),
		)
	}

	// Handle save result message
	if saveMsg, ok := msg.(EditSaveResultMsg); ok {
		if saveMsg.Err == nil {
			m.editing = false
			m.editModal = m.editModal.Hide()
			// Invalidate cache and refresh data after successful save
			m.loading = true
			return m, tea.Batch(
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			// Invalidate cache and fetch fresh data
			return m, tea.Batch(
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
				m.fetchAndCacheStatus(),
			)
		}
	case "d":
		if !m.starting && !m.loading && m.device != "" && m.ble != nil && m.ble.Enable {
			m.starting = true
			m.err = nil
			return m, m.startDiscovery()
		}
	case "e", "enter":
		// Open edit modal
		if m.device != "" && !m.loading && m.ble != nil {
			m.editing = true
			m.editModal = m.editModal.SetSize(m.width, m.height)
			var cmd tea.Cmd
			m.editModal, cmd = m.editModal.Show(m.device, m.ble)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the BLE component.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Bluetooth").
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

	// BLE Configuration Section
	content.WriteString(m.renderBLEConfig())
	content.WriteString("\n\n")

	// BTHome Section
	content.WriteString(m.renderBTHome())

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		var footer string
		if m.ble != nil && m.ble.Enable {
			footer = "e:edit r:refresh d:discover"
		} else {
			footer = "e:edit r:refresh"
		}
		if cs := m.cacheStatus.View(); cs != "" {
			footer = cs + " " + footer
		}
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) renderBLEConfig() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("BLE Configuration"))
	content.WriteString("\n")

	if m.ble == nil {
		content.WriteString(m.styles.Muted.Render("  Not supported"))
		return content.String()
	}

	// Bluetooth enabled status
	if m.ble.Enable {
		content.WriteString("  " + m.styles.Enabled.Render("● Bluetooth Enabled") + "\n")
	} else {
		content.WriteString("  " + m.styles.Disabled.Render("○ Bluetooth Disabled") + "\n")
	}

	if !m.ble.Enable {
		return content.String()
	}

	// RPC status
	content.WriteString("  " + m.styles.Label.Render("RPC:      "))
	if m.ble.RPCEnabled {
		content.WriteString(m.styles.Enabled.Render("Enabled"))
	} else {
		content.WriteString(m.styles.Disabled.Render("Disabled"))
	}
	content.WriteString("\n")

	// Observer mode
	content.WriteString("  " + m.styles.Label.Render("Observer: "))
	if m.ble.ObserverMode {
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString(m.styles.Muted.Render(" (receives BLU broadcasts)"))
	} else {
		content.WriteString(m.styles.Disabled.Render("Disabled"))
	}

	return content.String()
}

func (m Model) renderBTHome() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("BTHome Devices"))
	content.WriteString("\n")

	if m.ble == nil || !m.ble.Enable {
		content.WriteString(m.styles.Muted.Render("  Enable Bluetooth to manage BTHome devices"))
		return content.String()
	}

	// Discovery status
	switch {
	case m.discovery != nil && m.discovery.Active:
		content.WriteString("  " + m.styles.Warning.Render("◐ Discovery in progress...") + "\n")
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("    Duration: %ds", m.discovery.Duration),
		))
	case m.starting:
		content.WriteString("  " + m.styles.Muted.Render("◐ Starting discovery..."))
	default:
		content.WriteString("  " + m.styles.Muted.Render("No active discovery"))
		content.WriteString("\n")
		content.WriteString("  " + m.styles.Muted.Render("Press 'd' to scan for BTHome devices"))
	}

	return content.String()
}

// BLE returns the current BLE configuration.
func (m Model) BLE() *shelly.BLEConfig {
	return m.ble
}

// Discovery returns the current BTHome discovery status.
func (m Model) Discovery() *shelly.BTHomeDiscovery {
	return m.discovery
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Starting returns whether discovery is starting.
func (m Model) Starting() bool {
	return m.starting
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Refresh triggers a refresh of the BLE data.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

// IsEditing returns whether the edit modal is currently visible.
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
