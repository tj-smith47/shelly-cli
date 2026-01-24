// Package ble provides TUI components for managing device Bluetooth settings.
// This includes BLE configuration and BTHome device management.
package ble

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
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
	BLE       *shelly.BLEConfig              `json:"ble"`
	Discovery *shelly.BTHomeDiscovery        `json:"discovery"`
	Devices   []model.BTHomeDeviceInfo       `json:"devices"`
	Sensors   []model.BTHomeSensorInfo       `json:"sensors"`
	ObjInfos  map[int]model.BTHomeObjectInfo `json:"obj_infos"`
}

// StatusLoadedMsg signals that BLE status was loaded.
type StatusLoadedMsg struct {
	BLE       *shelly.BLEConfig
	Discovery *shelly.BTHomeDiscovery
	Devices   []model.BTHomeDeviceInfo
	Sensors   []model.BTHomeSensorInfo
	ObjInfos  map[int]model.BTHomeObjectInfo
	Err       error
}

// DiscoveryStartedMsg signals that BTHome discovery was started.
type DiscoveryStartedMsg struct {
	Err error
}

// DeviceRemovedMsg signals that a BTHome device was removed.
type DeviceRemovedMsg struct {
	DeviceID int
	Err      error
}

// Model displays BLE and BTHome settings for a device.
type Model struct {
	helpers.Sizable
	ctx           context.Context
	svc           *shelly.Service
	fileCache     *cache.FileCache
	device        string
	ble           *shelly.BLEConfig
	discovery     *shelly.BTHomeDiscovery
	devices       []model.BTHomeDeviceInfo       // BTHome paired devices
	sensors       []model.BTHomeSensorInfo       // BTHome sensor readings
	objInfos      map[int]model.BTHomeObjectInfo // Object ID to name/unit mapping
	cursor        int                            // Selected device index
	pendingRemove int                            // Device ID pending removal (-1 = none)
	loading       bool
	starting      bool
	editing       bool
	pairing       bool
	err           error
	focused       bool
	panelIndex    int // 1-based panel index for Shift+N hotkey hint
	styles        Styles
	editModal     EditModel
	pairModal     PairModel
	cacheStatus   cachestatus.Model
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
	Selected  lipgloss.Style
	DevName   lipgloss.Style
	DevAddr   lipgloss.Style
	Signal    lipgloss.Style
	Battery   lipgloss.Style
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
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground),
		DevName: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		DevAddr: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Signal: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Battery: lipgloss.NewStyle().
			Foreground(colors.Success),
	}
}

// New creates a new BLE model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("ble component init", err)
		panic(fmt.Sprintf("ble: invalid deps: %v", err))
	}

	m := Model{
		Sizable:       helpers.NewSizableLoaderOnly(),
		ctx:           deps.Ctx,
		svc:           deps.Svc,
		fileCache:     deps.FileCache,
		pendingRemove: -1, // -1 means no pending removal
		styles:        DefaultStyles(),
		cacheStatus:   cachestatus.New(),
		editModal:     NewEditModel(deps.Ctx, deps.Svc),
		pairModal:     NewPairModel(deps.Ctx, deps.Svc),
	}
	m.Loader = m.Loader.SetMessage("Loading Bluetooth settings...")
	return m
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
	m.devices = nil
	m.sensors = nil
	m.objInfos = nil
	m.cursor = 0
	m.pendingRemove = -1
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

		// Fetch BTHome devices and sensors if BLE is enabled
		if data.BLE != nil && data.BLE.Enable {
			devices, err := m.svc.Wireless().FetchBTHomeDevices(ctx, m.device, nil)
			if err == nil {
				data.Devices = devices
			}

			sensors, err := m.svc.Wireless().FetchBTHomeSensors(ctx, m.device, nil)
			if err == nil {
				data.Sensors = sensors
				// Fetch object info for sensor types
				data.ObjInfos = fetchObjectInfos(ctx, m.svc, m.device, sensors)
			}
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

		// Fetch BTHome devices and sensors if BLE is enabled
		if data.BLE != nil && data.BLE.Enable {
			devices, err := m.svc.Wireless().FetchBTHomeDevices(ctx, m.device, nil)
			if err == nil {
				data.Devices = devices
			}

			sensors, err := m.svc.Wireless().FetchBTHomeSensors(ctx, m.device, nil)
			if err == nil {
				data.Sensors = sensors
				// Fetch object info for sensor types
				data.ObjInfos = fetchObjectInfos(ctx, m.svc, m.device, sensors)
			}
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
	m.ApplySize(width, height)
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

// SetPairModalSize sets the pairing modal dimensions.
// This should be called with screen-based dimensions when the modal is visible.
func (m Model) SetPairModalSize(width, height int) Model {
	if m.pairing {
		m.pairModal = m.pairModal.SetSize(width, height)
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

	// Handle pairing modal if visible
	if m.pairing {
		return m.handlePairModalUpdate(msg)
	}

	// Forward tick messages to loader when loading
	if m.loading {
		if updated, cmd, done := m.updateLoading(msg); done {
			return updated, cmd
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
	case DeviceRemovedMsg:
		return m.handleDeviceRemoved(msg)
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.RefreshRequestMsg, messages.ScanRequestMsg, messages.EditRequestMsg,
		messages.ViewRequestMsg, messages.DeleteRequestMsg:
		return m.handleActionMsg(msg)
	case tea.KeyPressMsg:
		return m.handleKeyMsg(msg)
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
	case messages.RefreshRequestMsg:
		return m.handleRefresh()
	case messages.ScanRequestMsg:
		return m.handleScanDiscovery()
	case messages.EditRequestMsg, messages.ViewRequestMsg:
		return m.handleEditKey()
	case messages.DeleteRequestMsg:
		return m.handleRemoveKey()
	case messages.NewRequestMsg:
		return m.handlePairKey()
	}
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	return m.handleKey(msg)
}

func (m Model) updateLoading(msg tea.Msg) (Model, tea.Cmd, bool) {
	result := generics.UpdateLoader(m.Loader, msg, func(inMsg tea.Msg) bool {
		switch inMsg.(type) {
		case StatusLoadedMsg, DiscoveryStartedMsg:
			return true
		}
		return generics.IsPanelCacheMsg(inMsg)
	})
	m.Loader = result.Loader
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
		m.devices = sortBTHomeDevices(data.Devices)
		m.sensors = data.Sensors
		m.objInfos = data.ObjInfos
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit StatusLoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg {
		return StatusLoadedMsg{BLE: m.ble, Discovery: m.discovery, Devices: m.devices, Sensors: m.sensors, ObjInfos: m.objInfos}
	}

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
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeBLE {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("ble background refresh", msg.Err)
		m.err = msg.Err
		// Emit StatusLoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return StatusLoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedBLEData); ok {
		m.ble = data.BLE
		m.discovery = data.Discovery
		m.devices = sortBTHomeDevices(data.Devices)
		m.sensors = data.Sensors
		m.objInfos = data.ObjInfos
	}
	// Emit StatusLoadedMsg so sequential loading can advance
	return m, func() tea.Msg {
		return StatusLoadedMsg{BLE: m.ble, Discovery: m.discovery, Devices: m.devices, Sensors: m.sensors, ObjInfos: m.objInfos}
	}
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
	m.devices = sortBTHomeDevices(msg.Devices)
	m.sensors = msg.Sensors
	m.objInfos = msg.ObjInfos
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
		m.Loader.Tick(),
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
			m.Loader.Tick(),
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
				m.Loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handlePairModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.pairModal, cmd = m.pairModal.Update(msg)

	// Check if modal was closed
	if !m.pairModal.Visible() {
		m.pairing = false
		// Invalidate cache and refresh data after pairing
		m.loading = true
		return m, tea.Batch(
			cmd,
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
			m.fetchAndCacheStatus(),
		)
	}

	// Handle device added message
	if addMsg, ok := msg.(DeviceAddedMsg); ok {
		if addMsg.Err == nil {
			m.pairing = false
			m.pairModal = m.pairModal.Hide()
			// Invalidate cache and refresh data after successful add
			m.loading = true
			return m, tea.Batch(
				m.Loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return PairClosedMsg{Added: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handlePairKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.ble == nil || !m.ble.Enable {
		return m, nil
	}
	m.pairing = true
	m.pairModal = m.pairModal.SetSize(m.Width, m.Height)
	var cmd tea.Cmd
	m.pairModal, cmd = m.pairModal.Show(m.device)
	return m, cmd
}

func (m Model) handleRefresh() (Model, tea.Cmd) {
	if m.loading || m.device == "" {
		return m, nil
	}
	m.loading = true
	// Invalidate cache and fetch fresh data
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleScanDiscovery() (Model, tea.Cmd) {
	if m.starting || m.loading || m.device == "" || m.ble == nil || !m.ble.Enable {
		return m, nil
	}
	m.starting = true
	m.err = nil
	return m, m.startDiscovery()
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.ble == nil {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.Width, m.Height)
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Show(m.device, m.ble)
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+[":
		// Handle escape to cancel pending removal
		if m.pendingRemove >= 0 {
			m.pendingRemove = -1
			return m, nil
		}
	case "p":
		// Pair new BTHome device
		return m.handlePairKey()
	}
	return m, nil
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	if len(m.devices) == 0 {
		return m, nil
	}
	switch msg.Direction {
	case messages.NavUp:
		if m.cursor > 0 {
			m.cursor--
			m.pendingRemove = -1 // Cancel pending removal when navigating
		}
	case messages.NavDown:
		if m.cursor < len(m.devices)-1 {
			m.cursor++
			m.pendingRemove = -1 // Cancel pending removal when navigating
		}
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m Model) handleRemoveKey() (Model, tea.Cmd) {
	if len(m.devices) == 0 || m.cursor >= len(m.devices) {
		return m, nil
	}
	device := m.devices[m.cursor]

	// If this device is already pending removal, confirm and remove
	if m.pendingRemove == device.ID {
		m.pendingRemove = -1
		return m, m.removeDevice(device.ID)
	}

	// Otherwise, mark as pending removal
	m.pendingRemove = device.ID
	return m, nil
}

func (m Model) removeDevice(deviceID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		err := m.svc.Wireless().BTHomeRemoveDevice(ctx, m.device, deviceID)
		return DeviceRemovedMsg{DeviceID: deviceID, Err: err}
	}
}

func (m Model) handleDeviceRemoved(msg DeviceRemovedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}

	// Remove from list and adjust cursor
	for i, dev := range m.devices {
		if dev.ID == msg.DeviceID {
			m.devices = append(m.devices[:i], m.devices[i+1:]...)
			break
		}
	}
	if m.cursor >= len(m.devices) && m.cursor > 0 {
		m.cursor--
	}

	// Invalidate cache and refresh
	m.loading = true
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeBLE),
		m.fetchAndCacheStatus(),
	)
}

// View renders the BLE component.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	// Render pairing modal if pairing
	if m.pairing {
		return m.pairModal.View()
	}

	r := rendering.New(m.Width, m.Height).
		SetTitle("Bluetooth").
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

	// BLE Configuration Section
	content.WriteString(m.renderBLEConfig())
	content.WriteString("\n\n")

	// BTHome Section
	content.WriteString(m.renderBTHome())

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := m.buildFooter()
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) buildFooter() string {
	// Show removal confirmation prompt
	if m.pendingRemove >= 0 {
		return m.styles.Enabled.Render("Press 'x' again to confirm removal, Esc to cancel")
	}

	var footer string
	switch {
	case m.ble != nil && m.ble.Enable && len(m.devices) > 0:
		footer = "e:edit d:discover p:pair x:remove r:refresh"
	case m.ble != nil && m.ble.Enable:
		footer = "e:edit d:discover p:pair r:refresh"
	default:
		footer = "e:edit r:refresh"
	}

	if cs := m.cacheStatus.View(); cs != "" {
		footer += " | " + cs
	}
	return footer
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
	if len(m.devices) > 0 {
		content.WriteString(m.styles.Muted.Render(fmt.Sprintf(" (%d)", len(m.devices))))
	}
	content.WriteString("\n")

	if m.ble == nil || !m.ble.Enable {
		content.WriteString(m.styles.Muted.Render("  Enable Bluetooth to manage BTHome devices"))
		return content.String()
	}

	// Discovery status
	switch {
	case m.discovery != nil && m.discovery.Active:
		content.WriteString("  " + m.styles.Warning.Render("◐ Discovery in progress...") + "\n")
	case m.starting:
		content.WriteString("  " + m.styles.Muted.Render("◐ Starting discovery...") + "\n")
	}

	// Device list
	if len(m.devices) == 0 {
		content.WriteString("  " + m.styles.Muted.Render("No paired devices"))
		content.WriteString("\n")
		content.WriteString("  " + m.styles.Muted.Render("Press 'd' to scan for BTHome devices"))
	} else {
		content.WriteString("\n")
		for i, dev := range m.devices {
			line := m.renderBTHomeDevice(dev, i == m.cursor)
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	return content.String()
}

func (m Model) renderBTHomeDevice(dev model.BTHomeDeviceInfo, selected bool) string {
	var lines []string

	// First line: status, name, address, RSSI, battery
	var parts []string

	// Status indicator (pending removal or signal quality)
	parts = append(parts, m.renderDeviceStatus(dev))

	// Name and Address
	name := dev.Name
	if name == "" {
		name = "Unnamed"
	}
	nameStyle := m.styles.DevName
	if selected {
		nameStyle = m.styles.Selected.Inherit(nameStyle)
	}
	parts = append(parts, nameStyle.Render(truncate(name, 18)), m.styles.DevAddr.Render(dev.Addr))

	// RSSI
	if dev.RSSI != nil {
		parts = append(parts, m.styles.Signal.Render(fmt.Sprintf("%ddBm", *dev.RSSI)))
	}

	// Battery
	if dev.Battery != nil {
		parts = append(parts, m.styles.Battery.Render(fmt.Sprintf("🔋%d%%", *dev.Battery)))
	}

	lines = append(lines, "  "+strings.Join(parts, "  "))

	// Find sensors for this device by matching address
	deviceSensors := m.getSensorsForDevice(dev.Addr)
	if len(deviceSensors) > 0 {
		sensorLine := m.renderDeviceSensors(deviceSensors)
		if sensorLine != "" {
			lines = append(lines, "    "+sensorLine)
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) getSensorsForDevice(addr string) []model.BTHomeSensorInfo {
	var sensors []model.BTHomeSensorInfo
	for _, s := range m.sensors {
		if s.Addr == addr {
			sensors = append(sensors, s)
		}
	}
	return sensors
}

func (m Model) renderDeviceSensors(sensors []model.BTHomeSensorInfo) string {
	var parts []string

	for _, s := range sensors {
		valueStr := m.formatSensorValue(s)
		if valueStr != "" {
			parts = append(parts, valueStr)
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return m.styles.Muted.Render(strings.Join(parts, " │ "))
}

func (m Model) formatSensorValue(s model.BTHomeSensorInfo) string {
	if s.Value == nil {
		return ""
	}

	// Get sensor name and unit from object info
	name := s.Name
	unit := ""
	if info, ok := m.objInfos[s.ObjID]; ok {
		if name == "" {
			name = info.Name
		}
		unit = info.Unit
	}
	if name == "" {
		name = fmt.Sprintf("Obj%d", s.ObjID)
	}

	// Format value based on type
	var valueStr string
	switch v := s.Value.(type) {
	case float64:
		if v == float64(int(v)) {
			valueStr = fmt.Sprintf("%d", int(v))
		} else {
			valueStr = fmt.Sprintf("%.1f", v)
		}
	case bool:
		if v {
			valueStr = "Yes"
		} else {
			valueStr = "No"
		}
	case string:
		valueStr = v
	default:
		valueStr = fmt.Sprintf("%v", v)
	}

	if unit != "" {
		return fmt.Sprintf("%s: %s%s", name, valueStr, unit)
	}
	return fmt.Sprintf("%s: %s", name, valueStr)
}

func (m Model) renderDeviceStatus(dev model.BTHomeDeviceInfo) string {
	switch {
	case m.pendingRemove == dev.ID:
		return m.styles.Error.Render("✗")
	case dev.RSSI != nil && *dev.RSSI >= -60:
		return m.styles.Enabled.Render("●")
	case dev.RSSI != nil && *dev.RSSI >= -80:
		return m.styles.Warning.Render("●")
	case dev.RSSI != nil:
		return m.styles.Disabled.Render("●")
	default:
		return m.styles.Muted.Render("○")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func sortBTHomeDevices(devices []model.BTHomeDeviceInfo) []model.BTHomeDeviceInfo {
	if devices == nil {
		return nil
	}
	// Sort by name, then by ID
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Name != devices[j].Name {
			return devices[i].Name < devices[j].Name
		}
		return devices[i].ID < devices[j].ID
	})
	return devices
}

// fetchObjectInfos fetches BTHome object info for a list of sensors.
func fetchObjectInfos(ctx context.Context, svc *shelly.Service, device string, sensors []model.BTHomeSensorInfo) map[int]model.BTHomeObjectInfo {
	if len(sensors) == 0 {
		return nil
	}

	// Collect unique object IDs
	objIDSet := make(map[int]struct{})
	for _, s := range sensors {
		objIDSet[s.ObjID] = struct{}{}
	}

	objIDs := make([]int, 0, len(objIDSet))
	for id := range objIDSet {
		objIDs = append(objIDs, id)
	}

	infos, err := svc.Wireless().FetchBTHomeObjectInfos(ctx, device, objIDs)
	if err != nil {
		return nil
	}

	result := make(map[int]model.BTHomeObjectInfo, len(infos))
	for _, info := range infos {
		result[info.ObjID] = info
	}
	return result
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
	return m, tea.Batch(m.Loader.Tick(), m.fetchStatus())
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

// IsPairing returns whether the pairing modal is currently visible.
func (m Model) IsPairing() bool {
	return m.pairing
}

// RenderPairModal returns the pairing modal view for full-screen overlay rendering.
func (m Model) RenderPairModal() string {
	if !m.pairing {
		return ""
	}
	return m.pairModal.View()
}
