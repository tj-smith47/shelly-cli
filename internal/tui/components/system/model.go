// Package system provides TUI components for managing device system settings.
package system

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

// Deps holds the dependencies for the System component.
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

// CachedSystemData holds system status and config for caching.
type CachedSystemData struct {
	Status *shelly.SysStatus `json:"status"`
	Config *shelly.SysConfig `json:"config"`
}

// StatusLoadedMsg signals that system status was loaded.
type StatusLoadedMsg struct {
	Status *shelly.SysStatus
	Config *shelly.SysConfig
	Err    error
}

// SettingField represents a configurable field.
type SettingField int

// Setting field constants.
const (
	FieldName SettingField = iota
	FieldTimezone
	FieldEcoMode
	FieldDiscoverable
	FieldCount
)

// Model displays system settings for a device.
type Model struct {
	ctx         context.Context
	svc         *shelly.Service
	fileCache   *cache.FileCache
	device      string
	status      *shelly.SysStatus
	config      *shelly.SysConfig
	cursor      SettingField
	loading     bool
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

// Styles holds styles for the System component.
type Styles struct {
	Label        lipgloss.Style
	Value        lipgloss.Style
	ValueEnabled lipgloss.Style
	ValueMuted   lipgloss.Style
	Selected     lipgloss.Style
	Error        lipgloss.Style
	Muted        lipgloss.Style
	Title        lipgloss.Style
}

// DefaultStyles returns the default styles for the System component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Text),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		ValueEnabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		ValueMuted: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new System model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("system component init", err)
		panic(fmt.Sprintf("system: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading system settings..."),
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

// SetDevice sets the device to display system settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.status = nil
	m.config = nil
	m.cursor = FieldName
	m.err = nil
	m.loading = true
	m.cacheStatus = cachestatus.New()

	if device == "" {
		m.loading = false
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeSystem)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		config, configErr := m.svc.GetSysConfig(ctx, m.device)
		if configErr != nil {
			return StatusLoadedMsg{Status: status, Err: configErr}
		}

		return StatusLoadedMsg{Status: status, Config: config}
	}
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeSystem, cache.TTLSystem, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return nil, err
		}

		config, configErr := m.svc.GetSysConfig(ctx, m.device)
		if configErr != nil {
			return nil, configErr
		}

		return CachedSystemData{Status: status, Config: config}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeSystem, cache.TTLSystem, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return nil, err
		}

		config, configErr := m.svc.GetSysConfig(ctx, m.device)
		if configErr != nil {
			return nil, configErr
		}

		return CachedSystemData{Status: status, Config: config}, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	// Update edit modal size
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
	// Forward to edit modal if editing
	if m.editing {
		return m.updateEditing(msg)
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

	switch msg := msg.(type) {
	case panelcache.CacheHitMsg:
		return m.handleCacheHit(msg)
	case panelcache.CacheMissMsg:
		return m.handleCacheMiss(msg)
	case panelcache.RefreshCompleteMsg:
		return m.handleRefreshComplete(msg)
	case StatusLoadedMsg:
		return m.handleStatusLoaded(msg)
	case EditClosedMsg:
		return m.handleEditClosed(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) updateEditing(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Update(msg)

	// Check if modal was closed
	if !m.editModal.Visible() {
		m.editing = false
		// Check if it was a save - invalidate cache and refresh data
		if closedMsg, ok := msg.(EditClosedMsg); ok && closedMsg.Saved {
			m.loading = true
			return m, tea.Batch(
				cmd,
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeSystem),
				m.fetchAndCacheStatus(),
			)
		}
	}

	return m, cmd
}

func (m Model) updateLoading(msg tea.Msg) (Model, tea.Cmd, bool) {
	result := generics.UpdateLoader(m.loader, msg, func(msg tea.Msg) bool {
		if _, ok := msg.(StatusLoadedMsg); ok {
			return true
		}
		return generics.IsPanelCacheMsg(msg)
	})
	m.loader = result.Loader
	return m, result.Cmd, result.Consumed
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeSystem {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedSystemData](msg.Data)
	if err == nil {
		m.status = data.Status
		m.config = data.Config
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit StatusLoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return StatusLoadedMsg{Status: m.status, Config: m.config} }

	// Background refresh if stale
	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeSystem {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeSystem {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("system background refresh", msg.Err)
		m.err = msg.Err
		// Emit StatusLoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return StatusLoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedSystemData); ok {
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

func (m Model) handleEditClosed(msg EditClosedMsg) (Model, tea.Cmd) {
	if !msg.Saved {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(
		m.loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeSystem),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "r":
		return m.handleRefreshKey()
	case "t":
		return m.toggleCurrentField()
	case "e":
		return m.handleEditKey()
	case "z":
		return m.handleTimezoneKey()
	}

	return m, nil
}

func (m Model) handleRefreshKey() (Model, tea.Cmd) {
	if m.loading || m.device == "" {
		return m, nil
	}
	m.loading = true
	// Invalidate cache and fetch fresh data
	return m, tea.Batch(
		m.loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeSystem),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.config == nil {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.Show(m.device, m.config)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleTimezoneKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.config == nil {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowAtTimezone(m.device, m.config)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) cursorDown() Model {
	if m.cursor < FieldCount-1 {
		m.cursor++
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
	}
	return m
}

func (m Model) toggleCurrentField() (Model, tea.Cmd) {
	if m.config == nil || m.device == "" {
		return m, nil
	}

	switch m.cursor {
	case FieldEcoMode:
		newVal := !m.config.EcoMode
		return m, m.setEcoMode(newVal)
	case FieldDiscoverable:
		newVal := !m.config.Discoverable
		return m, m.setDiscoverable(newVal)
	case FieldName, FieldTimezone, FieldCount:
		// These fields are not toggleable
	}

	return m, nil
}

func (m Model) setEcoMode(enable bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetSysEcoMode(ctx, m.device, enable)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		// Refresh status after change
		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}
		config, err := m.svc.GetSysConfig(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Status: status, Err: err}
		}

		// Update cache with new data
		if m.fileCache != nil {
			data := CachedSystemData{Status: status, Config: config}
			iostreams.DebugErr("cache system after eco mode change", m.fileCache.Set(m.device, cache.TypeSystem, data, cache.TTLSystem))
		}

		return StatusLoadedMsg{Status: status, Config: config}
	}
}

func (m Model) setDiscoverable(discoverable bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetSysDiscoverable(ctx, m.device, discoverable)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		// Refresh status after change
		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}
		config, err := m.svc.GetSysConfig(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Status: status, Err: err}
		}

		// Update cache with new data
		if m.fileCache != nil {
			data := CachedSystemData{Status: status, Config: config}
			iostreams.DebugErr("cache system after discoverable change", m.fileCache.Set(m.device, cache.TypeSystem, data, cache.TTLSystem))
		}

		return StatusLoadedMsg{Status: status, Config: config}
	}
}

// View renders the System component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("System").
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

	// System status section
	content.WriteString(m.renderStatus())
	content.WriteString("\n\n")

	// Configuration section
	content.WriteString(m.styles.Title.Render("Settings"))
	content.WriteString("\n")
	content.WriteString(m.renderSettings())

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := "e:edit z:timezone t:toggle r:refresh"
		if cacheView := m.cacheStatus.View(); cacheView != "" {
			footer += " | " + cacheView
		}
		r.SetFooter(footer)
	}

	// Render edit modal overlay if editing
	if m.editing {
		return m.editModal.View()
	}

	return r.Render()
}

func (m Model) renderStatus() string {
	if m.status == nil {
		return m.styles.Muted.Render("No status available")
	}

	var content strings.Builder

	// Uptime
	content.WriteString(m.styles.Label.Render("Uptime:    "))
	content.WriteString(m.styles.Value.Render(formatUptime(m.status.Uptime)))
	content.WriteString("\n")

	// Memory
	content.WriteString(m.styles.Label.Render("RAM:       "))
	ramUsed := m.status.RAMSize - m.status.RAMFree
	ramPct := float64(ramUsed) / float64(m.status.RAMSize) * 100
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d/%d KB (%.0f%%)",
		ramUsed/1024, m.status.RAMSize/1024, ramPct)))
	content.WriteString("\n")

	// Filesystem
	content.WriteString(m.styles.Label.Render("Storage:   "))
	fsUsed := m.status.FSSize - m.status.FSFree
	fsPct := float64(fsUsed) / float64(m.status.FSSize) * 100
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d/%d KB (%.0f%%)",
		fsUsed/1024, m.status.FSSize/1024, fsPct)))
	content.WriteString("\n")

	// Time
	if m.status.Time != "" {
		content.WriteString(m.styles.Label.Render("Time:      "))
		content.WriteString(m.styles.Value.Render(m.status.Time))
		content.WriteString("\n")
	}

	// Update available
	if m.status.UpdateAvailable != "" {
		content.WriteString(m.styles.Label.Render("Update:    "))
		content.WriteString(m.styles.ValueEnabled.Render(m.status.UpdateAvailable + " available"))
		content.WriteString("\n")
	}

	// Restart required
	if m.status.RestartRequired {
		content.WriteString(m.styles.Error.Render("⚠ Restart required"))
	}

	return content.String()
}

func (m Model) renderSettings() string {
	if m.config == nil {
		return m.styles.Muted.Render("No configuration available")
	}

	var content strings.Builder

	// Name field
	content.WriteString(m.renderSettingLine(FieldName, "Name", m.config.Name))
	content.WriteString("\n")

	// Timezone field
	tz := m.config.Timezone
	if tz == "" {
		tz = "(not set)"
	}
	content.WriteString(m.renderSettingLine(FieldTimezone, "Timezone", tz))
	content.WriteString("\n")

	// Eco Mode field
	ecoModeVal := "Disabled"
	if m.config.EcoMode {
		ecoModeVal = "Enabled"
	}
	content.WriteString(m.renderToggleLine(FieldEcoMode, "Eco Mode", m.config.EcoMode, ecoModeVal))
	content.WriteString("\n")

	// Discoverable field
	discVal := "Hidden"
	if m.config.Discoverable {
		discVal = "Visible"
	}
	content.WriteString(m.renderToggleLine(FieldDiscoverable, "Discoverable", m.config.Discoverable, discVal))

	return content.String()
}

func (m Model) renderSettingLine(field SettingField, label, value string) string {
	selector := "  "
	if m.cursor == field {
		selector = "▶ "
	}

	labelWidth := 14
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label+":")

	line := selector + m.styles.Label.Render(paddedLabel) + m.styles.Value.Render(value)

	if m.cursor == field {
		return m.styles.Selected.Render(line)
	}
	return line
}

func (m Model) renderToggleLine(field SettingField, label string, enabled bool, value string) string {
	selector := "  "
	if m.cursor == field {
		selector = "▶ "
	}

	labelWidth := 14
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label+":")

	var valueStyle lipgloss.Style
	if enabled {
		valueStyle = m.styles.ValueEnabled
	} else {
		valueStyle = m.styles.ValueMuted
	}

	line := selector + m.styles.Label.Render(paddedLabel) + valueStyle.Render(value)

	if m.cursor == field {
		return m.styles.Selected.Render(line)
	}
	return line
}

func formatUptime(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// Status returns the current system status.
func (m Model) Status() *shelly.SysStatus {
	return m.status
}

// Config returns the current system config.
func (m Model) Config() *shelly.SysConfig {
	return m.config
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

// Refresh triggers a refresh of the system status.
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
