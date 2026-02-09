// Package security provides TUI components for displaying device security settings.
// This includes authentication status, debug mode, and device visibility.
package security

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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/errorview"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// Deps holds the dependencies for the Security component.
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

// CachedSecurityData holds security status for caching.
type CachedSecurityData struct {
	Status *shelly.TUISecurityStatus `json:"status"`
}

// StatusLoadedMsg signals that security status was loaded.
type StatusLoadedMsg struct {
	Status *shelly.TUISecurityStatus
	Err    error
}

// Model displays security settings for a device.
type Model struct {
	panel.Sizable
	ctx         context.Context
	svc         *shelly.Service
	fileCache   *cache.FileCache
	device      string
	status      *shelly.TUISecurityStatus
	loading     bool
	err         error
	focused     bool
	panelIndex  int // 1-based panel index for Shift+N hotkey hint
	styles      Styles
	cacheStatus cachestatus.Model

	// Edit modal
	editModel EditModel
}

// Styles holds styles for the Security component.
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
	Danger    lipgloss.Style
}

// DefaultStyles returns the default styles for the Security component.
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
		Danger: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
	}
}

// New creates a new Security model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("security component init", err)
		panic(fmt.Sprintf("security: invalid deps: %v", err))
	}

	m := Model{
		Sizable:     panel.NewSizableLoaderOnly(),
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		editModel:   NewEditModel(deps.Ctx, deps.Svc),
	}
	m.Loader = m.Loader.SetMessage("Loading security settings...")
	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display security settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.status = nil
	m.err = nil
	m.loading = true
	m.cacheStatus = cachestatus.New()

	if device == "" {
		m.loading = false
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeSecurity)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		status, err := m.svc.GetTUISecurityStatus(ctx, m.device)
		return StatusLoadedMsg{
			Status: status,
			Err:    err,
		}
	}
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeSecurity, cache.TTLSecurity, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		status, err := m.svc.GetTUISecurityStatus(ctx, m.device)
		if err != nil {
			return nil, err
		}
		return CachedSecurityData{Status: status}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeSecurity, cache.TTLSecurity, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		status, err := m.svc.GetTUISecurityStatus(ctx, m.device)
		if err != nil {
			return nil, err
		}
		return CachedSecurityData{Status: status}, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.ApplySize(width, height)
	m.editModel = m.editModel.SetSize(width, height)
	return m
}

// SetEditModalSize sets the edit modal dimensions.
// This should be called with screen-based dimensions when the modal is visible.
func (m Model) SetEditModalSize(width, height int) Model {
	m.ModalWidth = width
	m.ModalHeight = height
	if m.editModel.Visible() {
		m.editModel = m.editModel.SetSize(width, height)
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
	// Forward messages to edit modal when visible
	if m.editModel.Visible() {
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
	// Action messages from context system
	case messages.RefreshRequestMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleRefresh()
	case messages.AuthRequestMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleAuthConfig()
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
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

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeSecurity {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedSecurityData](msg.Data)
	if err == nil {
		m.status = data.Status
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit StatusLoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return StatusLoadedMsg{Status: m.status} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeSecurity {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeSecurity {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("security background refresh", msg.Err)
		m.err = msg.Err
		// Emit StatusLoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return StatusLoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedSecurityData); ok {
		m.status = data.Status
	}
	// Emit StatusLoadedMsg so sequential loading can advance
	return m, func() tea.Msg { return StatusLoadedMsg{Status: m.status} }
}

func (m Model) handleStatusLoaded(msg StatusLoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.status = msg.Status
	return m, nil
}

func (m Model) handleEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.editModel, cmd = m.editModel.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	// Check for EditClosedMsg to invalidate cache and refresh data
	if closedMsg, ok := msg.(EditClosedMsg); ok && closedMsg.Saved && m.device != "" {
		m.loading = true
		cmds = append(cmds,
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeSecurity),
			m.fetchAndCacheStatus(),
		)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleRefresh() (Model, tea.Cmd) {
	if m.loading || m.device == "" {
		return m, nil
	}
	m.loading = true
	// Invalidate cache and fetch fresh data
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeSecurity),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleAuthConfig() (Model, tea.Cmd) {
	if m.device == "" || m.status == nil || m.loading {
		return m, nil
	}
	var cmd tea.Cmd
	m.editModel, cmd = m.editModel.Show(m.device, m.status.AuthEnabled)
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not covered by action messages
	// (none currently - all keys migrated to action messages)
	_ = msg
	return m, nil
}

// View renders the Security component.
func (m Model) View() string {
	// If edit modal is visible, render it as overlay
	if m.editModel.Visible() {
		return m.editModel.View()
	}

	r := rendering.New(m.Width, m.Height).
		SetTitle("Security").
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
		r.SetContent(errorview.RenderInline(m.err))
		return r.Render()
	}

	if m.status == nil {
		r.SetContent(styles.NoDataAvailable(m.Width, m.Height))
		return r.Render()
	}

	var content strings.Builder

	// Authentication Section
	content.WriteString(m.renderAuth())
	content.WriteString("\n\n")

	// Device Visibility Section
	content.WriteString(m.renderVisibility())
	content.WriteString("\n\n")

	// Debug Mode Section
	content.WriteString(m.renderDebug())

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := theme.StyledKeybindings(keys.FormatHints([]keys.Hint{{Key: "a", Desc: "auth"}, {Key: "r", Desc: "refresh"}}, keys.FooterHintWidth(m.Width)))
		if cacheView := m.cacheStatus.View(); cacheView != "" {
			footer += " | " + cacheView
		}
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) renderAuth() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("Authentication"))
	content.WriteString("\n")

	if m.status.AuthEnabled {
		content.WriteString("  " + m.styles.Enabled.Render("● Protected"))
		content.WriteString("\n")
		content.WriteString("  " + m.styles.Muted.Render("Device requires password for access"))
	} else {
		content.WriteString("  " + m.styles.Danger.Render("○ UNPROTECTED"))
		content.WriteString("\n")
		content.WriteString("  " + m.styles.Warning.Render("⚠ No password set - anyone can control this device"))
	}

	return content.String()
}

func (m Model) renderVisibility() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("Device Visibility"))
	content.WriteString("\n")

	// Discoverable
	content.WriteString("  " + m.styles.Label.Render("Discoverable: "))
	if m.status.Discoverable {
		content.WriteString(m.styles.Enabled.Render("Yes"))
		content.WriteString(m.styles.Muted.Render(" (visible on network)"))
	} else {
		content.WriteString(m.styles.Disabled.Render("No"))
		content.WriteString(m.styles.Muted.Render(" (hidden)"))
	}
	content.WriteString("\n")

	// Eco Mode
	content.WriteString("  " + m.styles.Label.Render("Eco Mode:     "))
	if m.status.EcoMode {
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString(m.styles.Muted.Render(" (reduced power)"))
	} else {
		content.WriteString(m.styles.Disabled.Render("Disabled"))
	}

	return content.String()
}

func (m Model) renderDebug() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("Debug Logging"))
	content.WriteString("\n")

	hasDebug := m.status.DebugMQTT || m.status.DebugWS || m.status.DebugUDP

	if !hasDebug {
		content.WriteString("  " + m.styles.Muted.Render("○ No debug logging enabled"))
		return content.String()
	}

	content.WriteString("  " + m.styles.Warning.Render("● Debug logging active"))
	content.WriteString("\n")

	if m.status.DebugMQTT {
		content.WriteString("  " + m.styles.Label.Render("  MQTT:      "))
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString("\n")
	}

	if m.status.DebugWS {
		content.WriteString("  " + m.styles.Label.Render("  WebSocket: "))
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString("\n")
	}

	if m.status.DebugUDP {
		content.WriteString("  " + m.styles.Label.Render("  UDP:       "))
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		if m.status.DebugUDPAddr != "" {
			content.WriteString(m.styles.Muted.Render(" → " + m.status.DebugUDPAddr))
		}
	}

	return strings.TrimSuffix(content.String(), "\n")
}

// Status returns the current security status.
func (m Model) Status() *shelly.TUISecurityStatus {
	return m.status
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

// Refresh triggers a refresh of the security data.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchStatus())
}

// IsEditing returns whether the edit modal is currently visible.
func (m Model) IsEditing() bool {
	return m.editModel.Visible()
}

// RenderEditModal returns the edit modal view for full-screen overlay rendering.
func (m Model) RenderEditModal() string {
	if !m.editModel.Visible() {
		return ""
	}
	return m.editModel.View()
}
