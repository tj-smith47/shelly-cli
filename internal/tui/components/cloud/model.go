// Package cloud provides TUI components for managing device cloud settings.
package cloud

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

// Deps holds the dependencies for the Cloud component.
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

// CachedCloudData holds cloud status for caching.
type CachedCloudData struct {
	Connected bool   `json:"connected"`
	Enabled   bool   `json:"enabled"`
	Server    string `json:"server"`
}

// StatusLoadedMsg signals that cloud status was loaded.
type StatusLoadedMsg struct {
	Status  *shelly.CloudStatus
	Config  map[string]any
	Enabled bool
	Server  string
	Err     error
}

// ToggleResultMsg signals the result of a toggle operation.
type ToggleResultMsg struct {
	Enabled bool
	Err     error
}

// Model displays cloud settings for a device.
type Model struct {
	ctx         context.Context
	svc         *shelly.Service
	fileCache   *cache.FileCache
	device      string
	connected   bool
	enabled     bool
	server      string
	loading     bool
	toggling    bool
	err         error
	width       int
	height      int
	focused     bool
	panelIndex  int // 1-based panel index for Shift+N hotkey hint
	styles      Styles
	loader      loading.Model
	cacheStatus cachestatus.Model

	// Edit modal
	editModel EditModel
}

// Styles holds styles for the Cloud component.
type Styles struct {
	Connected    lipgloss.Style
	Disconnected lipgloss.Style
	Enabled      lipgloss.Style
	Disabled     lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Error        lipgloss.Style
	Muted        lipgloss.Style
	Title        lipgloss.Style
}

// DefaultStyles returns the default styles for the Cloud component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Connected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disconnected: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new Cloud model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("cloud component init", err)
		panic(fmt.Sprintf("cloud: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading cloud status..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		editModel: NewEditModel(deps.Ctx, deps.Svc),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display cloud settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.connected = false
	m.enabled = false
	m.server = ""
	m.err = nil
	m.loading = true
	m.cacheStatus = cachestatus.New()

	if device == "" {
		m.loading = false
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeCloud)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetCloudStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		config, configErr := m.svc.GetCloudConfig(ctx, m.device)
		if configErr != nil {
			return StatusLoadedMsg{Status: status, Err: configErr}
		}

		// Extract enabled and server from config
		var enabled bool
		var server string
		if e, ok := config["enable"].(*bool); ok && e != nil {
			enabled = *e
		} else if e, ok := config["enable"].(bool); ok {
			enabled = e
		}
		if s, ok := config["server"].(*string); ok && s != nil {
			server = *s
		} else if s, ok := config["server"].(string); ok {
			server = s
		}

		return StatusLoadedMsg{
			Status:  status,
			Config:  config,
			Enabled: enabled,
			Server:  server,
		}
	}
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeCloud, cache.TTLCloud, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetCloudStatus(ctx, m.device)
		if err != nil {
			return nil, err
		}

		config, configErr := m.svc.GetCloudConfig(ctx, m.device)
		if configErr != nil {
			return nil, configErr
		}

		// Extract enabled and server from config
		var connected, enabled bool
		var server string
		if status != nil {
			connected = status.Connected
		}
		if e, ok := config["enable"].(*bool); ok && e != nil {
			enabled = *e
		} else if e, ok := config["enable"].(bool); ok {
			enabled = e
		}
		if s, ok := config["server"].(*string); ok && s != nil {
			server = *s
		} else if s, ok := config["server"].(string); ok {
			server = s
		}

		return CachedCloudData{Connected: connected, Enabled: enabled, Server: server}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeCloud, cache.TTLCloud, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetCloudStatus(ctx, m.device)
		if err != nil {
			return nil, err
		}

		config, configErr := m.svc.GetCloudConfig(ctx, m.device)
		if configErr != nil {
			return nil, configErr
		}

		// Extract enabled and server from config
		var connected, enabled bool
		var server string
		if status != nil {
			connected = status.Connected
		}
		if e, ok := config["enable"].(*bool); ok && e != nil {
			enabled = *e
		} else if e, ok := config["enable"].(bool); ok {
			enabled = e
		}
		if s, ok := config["server"].(*string); ok && s != nil {
			server = *s
		} else if s, ok := config["server"].(string); ok {
			server = s
		}

		return CachedCloudData{Connected: connected, Enabled: enabled, Server: server}, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.loader = m.loader.SetSize(width-4, height-4)
	m.editModel = m.editModel.SetSize(width, height)
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
	case ToggleResultMsg:
		return m.handleToggleResult(msg)
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
		case StatusLoadedMsg, ToggleResultMsg:
			return true
		}
		return generics.IsPanelCacheMsg(msg)
	})
	m.loader = result.Loader
	return m, result.Cmd, result.Consumed
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeCloud {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedCloudData](msg.Data)
	if err == nil {
		m.connected = data.Connected
		m.enabled = data.Enabled
		m.server = data.Server
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, nil
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeCloud {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeCloud {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("cloud background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedCloudData); ok {
		m.connected = data.Connected
		m.enabled = data.Enabled
		m.server = data.Server
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
	if msg.Status != nil {
		m.connected = msg.Status.Connected
	}
	m.enabled = msg.Enabled
	m.server = msg.Server
	return m, nil
}

func (m Model) handleToggleResult(msg ToggleResultMsg) (Model, tea.Cmd) {
	m.toggling = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.enabled = msg.Enabled
	// Invalidate cache and refresh to get updated connection status
	m.loading = true
	return m, tea.Batch(
		m.loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeCloud),
		m.fetchAndCacheStatus(),
	)
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
			m.loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeCloud),
			m.fetchAndCacheStatus(),
		)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		// Open cloud configuration modal
		if m.device != "" && !m.loading && !m.toggling {
			m.editModel = m.editModel.Show(m.device, m.connected, m.enabled, m.server)
			return m, func() tea.Msg { return EditOpenedMsg{} }
		}
	case "t", "enter":
		if !m.toggling && !m.loading && m.device != "" {
			m.toggling = true
			m.err = nil
			return m, m.toggleCloud()
		}
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			// Invalidate cache and fetch fresh data
			return m, tea.Batch(
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeCloud),
				m.fetchAndCacheStatus(),
			)
		}
	}

	return m, nil
}

func (m Model) toggleCloud() tea.Cmd {
	newEnabled := !m.enabled
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetCloudEnabled(ctx, m.device, newEnabled)
		if err != nil {
			return ToggleResultMsg{Err: err}
		}

		return ToggleResultMsg{Enabled: newEnabled}
	}
}

// View renders the Cloud component.
func (m Model) View() string {
	// If edit modal is visible, render it as overlay
	if m.editModel.Visible() {
		return m.editModel.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Cloud").
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

	// Connection status
	content.WriteString(m.styles.Label.Render("Status:  "))
	if m.connected {
		content.WriteString(m.styles.Connected.Render("● Connected"))
	} else {
		content.WriteString(m.styles.Disconnected.Render("○ Disconnected"))
	}
	content.WriteString("\n\n")

	// Enabled status
	content.WriteString(m.styles.Label.Render("Enabled: "))
	if m.enabled {
		content.WriteString(m.styles.Enabled.Render("Yes"))
	} else {
		content.WriteString(m.styles.Disabled.Render("No"))
	}
	content.WriteString("\n")

	// Server
	if m.server != "" {
		content.WriteString(m.styles.Label.Render("Server:  "))
		content.WriteString(m.styles.Value.Render(m.server))
		content.WriteString("\n")
	}

	// Toggling indicator
	if m.toggling {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("Updating..."))
	}

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := "c:config t:toggle r:refresh"
		if cacheView := m.cacheStatus.View(); cacheView != "" {
			footer += " | " + cacheView
		}
		r.SetFooter(footer)
	}
	return r.Render()
}

// Connected returns whether the device is connected to cloud.
func (m Model) Connected() bool {
	return m.connected
}

// Enabled returns whether cloud is enabled.
func (m Model) Enabled() bool {
	return m.enabled
}

// Server returns the cloud server address.
func (m Model) Server() string {
	return m.server
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Toggling returns whether a toggle is in progress.
func (m Model) Toggling() bool {
	return m.toggling
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Refresh triggers a refresh of the cloud status.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
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
