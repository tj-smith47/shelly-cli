// Package webhooks provides TUI components for managing device webhooks.
package webhooks

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
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Webhook represents a webhook configuration on a device.
type Webhook struct {
	ID     int
	Name   string
	Event  string
	Enable bool
	URLs   []string
	Cid    int
}

// Deps holds the dependencies for the webhooks component.
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

// CachedWebhooksData holds webhooks for caching.
type CachedWebhooksData struct {
	Webhooks []Webhook `json:"webhooks"`
}

// LoadedMsg signals that webhooks were loaded.
type LoadedMsg struct {
	Webhooks []Webhook
	Err      error
}

// ActionMsg signals a webhook action result.
type ActionMsg struct {
	Action    string // "enable", "disable", "delete"
	WebhookID int
	Err       error
}

// SelectMsg signals that a webhook was selected.
type SelectMsg struct {
	Webhook Webhook
}

// CreateMsg signals that a new webhook should be created.
type CreateMsg struct {
	Device string
}

// Model displays webhooks for a device.
type Model struct {
	ctx         context.Context
	svc         *shelly.Service
	fileCache   *cache.FileCache
	device      string
	webhooks    []Webhook
	scroller    *panel.Scroller
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

// Styles holds styles for the webhook list component.
type Styles struct {
	Enabled  lipgloss.Style
	Disabled lipgloss.Style
	Event    lipgloss.Style
	URL      lipgloss.Style
	Name     lipgloss.Style
	Selected lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
}

// DefaultStyles returns the default styles for the webhook list.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Event: lipgloss.NewStyle().
			Foreground(colors.Warning),
		URL: lipgloss.NewStyle().
			Foreground(colors.Info),
		Name: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
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

// New creates a new webhooks model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("webhooks component init", err)
		panic(fmt.Sprintf("webhooks: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		scroller:    panel.NewScroller(0, 1),
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading webhooks..."),
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

// SetDevice sets the device to list webhooks for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.webhooks = nil
	m.scroller.SetItemCount(0)
	m.err = nil
	m.cacheStatus = cachestatus.New()

	if device == "" {
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeWebhooks)
}

// fetchWebhooks creates a command to fetch webhooks from the device.
func (m Model) fetchWebhooks() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		hooks, err := m.svc.ListWebhooks(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Webhook, len(hooks))
		for i, h := range hooks {
			result[i] = Webhook{
				ID:     h.ID,
				Name:   h.Name,
				Event:  h.Event,
				Enable: h.Enable,
				URLs:   h.URLs,
				Cid:    h.Cid,
			}
		}

		return LoadedMsg{Webhooks: result}
	}
}

// fetchAndCacheWebhooks fetches fresh data and caches it.
func (m Model) fetchAndCacheWebhooks() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeWebhooks, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		hooks, err := m.svc.ListWebhooks(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Webhook, len(hooks))
		for i, h := range hooks {
			result[i] = Webhook{
				ID:     h.ID,
				Name:   h.Name,
				Event:  h.Event,
				Enable: h.Enable,
				URLs:   h.URLs,
				Cid:    h.Cid,
			}
		}

		return CachedWebhooksData{Webhooks: result}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeWebhooks, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		hooks, err := m.svc.ListWebhooks(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Webhook, len(hooks))
		for i, h := range hooks {
			result[i] = Webhook{
				ID:     h.ID,
				Name:   h.Name,
				Event:  h.Event,
				Enable: h.Enable,
				URLs:   h.URLs,
				Cid:    h.Cid,
			}
		}

		return CachedWebhooksData{Webhooks: result}, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Calculate visible rows: height - borders (2) - title (1) - footer (1)
	visibleRows := height - 4
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
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
	case LoadedMsg:
		return m.handleLoaded(msg)
	case ActionMsg:
		return m.handleAction(msg)
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
	case LoadedMsg, ActionMsg, panelcache.CacheHitMsg, panelcache.CacheMissMsg, panelcache.RefreshCompleteMsg:
		return m, nil, false
	default:
		if cmd != nil {
			return m, cmd, true
		}
	}
	return m, nil, false
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWebhooks {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedWebhooksData](msg.Data)
	if err == nil {
		m.webhooks = data.Webhooks
		m.scroller.SetItemCount(len(m.webhooks))
		m.scroller.CursorToStart()
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, nil
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWebhooks {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheWebhooks())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWebhooks {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("webhooks background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedWebhooksData); ok {
		m.webhooks = data.Webhooks
		m.scroller.SetItemCount(len(m.webhooks))
	}
	return m, nil
}

func (m Model) handleLoaded(msg LoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.webhooks = msg.Webhooks
	m.scroller.SetItemCount(len(m.webhooks))
	m.scroller.CursorToStart()
	return m, nil
}

func (m Model) handleAction(msg ActionMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh after action
	m.loading = true
	return m, tea.Batch(
		m.loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeWebhooks),
		m.fetchAndCacheWebhooks(),
	)
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
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeWebhooks),
			m.fetchAndCacheWebhooks(),
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
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeWebhooks),
				m.fetchAndCacheWebhooks(),
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
	case "enter":
		return m, m.selectWebhook()
	case "e":
		return m.handleEditKey()
	case "t":
		return m, m.toggleWebhook()
	case "d":
		return m, m.deleteWebhook()
	case "n":
		return m, m.createWebhook()
	case "r":
		// Invalidate cache and fetch fresh data
		m.loading = true
		return m, tea.Batch(
			m.loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeWebhooks),
			m.fetchAndCacheWebhooks(),
		)
	}

	return m, nil
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.webhooks) == 0 {
		return m, nil
	}
	cursor := m.scroller.Cursor()
	if cursor >= len(m.webhooks) {
		return m, nil
	}
	webhook := m.webhooks[cursor]
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.Show(m.device, &webhook)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) createWebhook() tea.Cmd {
	if m.device == "" {
		return nil
	}
	return func() tea.Msg {
		return CreateMsg{Device: m.device}
	}
}

func (m Model) selectWebhook() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.webhooks) == 0 || cursor >= len(m.webhooks) {
		return nil
	}
	webhook := m.webhooks[cursor]
	return func() tea.Msg {
		return SelectMsg{Webhook: webhook}
	}
}

func (m Model) toggleWebhook() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.webhooks) == 0 || cursor >= len(m.webhooks) {
		return nil
	}
	webhook := m.webhooks[cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		newEnable := !webhook.Enable
		err := m.svc.UpdateWebhook(ctx, m.device, webhook.ID, shelly.UpdateWebhookParams{
			Event:  webhook.Event,
			URLs:   webhook.URLs,
			Name:   webhook.Name,
			Enable: &newEnable,
		})

		action := "enable"
		if !newEnable {
			action = "disable"
		}
		return ActionMsg{Action: action, WebhookID: webhook.ID, Err: err}
	}
}

func (m Model) deleteWebhook() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.webhooks) == 0 || cursor >= len(m.webhooks) {
		return nil
	}
	webhook := m.webhooks[cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteWebhook(ctx, m.device, webhook.ID)
		return ActionMsg{Action: "delete", WebhookID: webhook.ID, Err: err}
	}
}

// View renders the webhooks list.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Webhooks").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	m.setFooter(r)

	// Handle early return states
	if content, done := m.getStateContent(); done {
		r.SetContent(content)
		return r.Render()
	}

	// Render webhook list
	r.SetContent(m.renderWebhookList())
	return r.Render()
}

// setFooter adds the appropriate keybinding footer with cache status.
func (m Model) setFooter(r *rendering.Renderer) {
	if !m.focused || m.device == "" {
		return
	}
	var footer string
	if len(m.webhooks) > 0 {
		footer = "e:edit t:toggle d:del n:new r:refresh"
	} else {
		footer = "n:new r:refresh"
	}
	if cs := m.cacheStatus.View(); cs != "" {
		footer = cs + " " + footer
	}
	r.SetFooter(footer)
}

// getStateContent returns content for non-list states and whether to use it.
func (m Model) getStateContent() (string, bool) {
	if m.device == "" {
		return m.styles.Muted.Render("No device selected"), true
	}
	if m.loading {
		return m.loader.View(), true
	}
	if m.err != nil {
		return m.getErrorContent(), true
	}
	if len(m.webhooks) == 0 {
		return m.styles.Muted.Render("No webhooks configured"), true
	}
	return "", false
}

// getErrorContent returns the appropriate error message.
func (m Model) getErrorContent() string {
	errMsg := m.err.Error()
	if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "unknown method") ||
		strings.Contains(errMsg, "not found") {
		return m.styles.Muted.Render("Webhooks not supported on this device")
	}
	return m.styles.Error.Render("Error: " + errMsg)
}

// renderWebhookList renders the list of webhooks.
func (m Model) renderWebhookList() string {
	var content strings.Builder
	start, end := m.scroller.VisibleRange()

	for i := start; i < end; i++ {
		webhook := m.webhooks[i]
		isSelected := m.scroller.IsCursorAt(i)

		line := m.renderWebhookLine(webhook, isSelected)
		content.WriteString(line)
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	if m.scroller.HasMore() || m.scroller.HasPrevious() {
		content.WriteString(m.styles.Muted.Render("\n" + m.scroller.ScrollInfo()))
	}

	return content.String()
}

func (m Model) renderWebhookLine(webhook Webhook, isSelected bool) string {
	// Status icon
	var icon string
	if webhook.Enable {
		icon = m.styles.Enabled.Render("●")
	} else {
		icon = m.styles.Disabled.Render("○")
	}

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Calculate available width for event and URL
	// Fixed: selector(2) + icon(2) + spaces(2) = 6
	available := output.ContentWidth(m.width, 4+6)
	eventWidth, urlWidth := output.SplitWidth(available, 40, 15, 20)

	// Event type (truncate if too long)
	event := output.Truncate(webhook.Event, eventWidth)
	eventStr := m.styles.Event.Render(event)

	// URL count or first URL
	urlInfo := ""
	if len(webhook.URLs) > 0 {
		url := output.Truncate(webhook.URLs[0], urlWidth)
		if len(webhook.URLs) > 1 {
			urlInfo = fmt.Sprintf("%s +%d", url, len(webhook.URLs)-1)
		} else {
			urlInfo = url
		}
		urlInfo = m.styles.URL.Render(urlInfo)
	}

	line := fmt.Sprintf("%s%s %s %s", selector, icon, eventStr, urlInfo)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// SelectedWebhook returns the currently selected webhook, if any.
func (m Model) SelectedWebhook() *Webhook {
	cursor := m.scroller.Cursor()
	if len(m.webhooks) == 0 || cursor >= len(m.webhooks) {
		return nil
	}
	return &m.webhooks[cursor]
}

// WebhookCount returns the number of webhooks.
func (m Model) WebhookCount() int {
	return len(m.webhooks)
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

// Refresh triggers a refresh of the webhook list.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchWebhooks())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom enter:details d:delete"
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
