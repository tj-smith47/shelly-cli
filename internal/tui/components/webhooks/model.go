// Package webhooks provides TUI components for managing device webhooks.
package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
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
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
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

// TestResultMsg signals the result of a webhook URL test.
type TestResultMsg struct {
	WebhookID   int
	URL         string
	StatusCode  int
	Err         error
	TestedCount int // How many URLs were tested
	TotalURLs   int // Total URLs in the webhook
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
	panel.Sizable
	ctx           context.Context
	svc           *shelly.Service
	fileCache     *cache.FileCache
	device        string
	webhooks      []Webhook
	loading       bool
	editing       bool
	testing       bool // True when testing a webhook URL
	err           error
	focused       bool
	panelIndex    int // 1-based panel index for Shift+N hotkey hint
	pendingDelete int // Webhook ID pending delete confirmation (-1 = none)
	styles        Styles
	editModal     EditModel
	cacheStatus   cachestatus.Model
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

	m := Model{
		Sizable:       panel.NewSizable(4, panel.NewScroller(0, 1)),
		ctx:           deps.Ctx,
		svc:           deps.Svc,
		fileCache:     deps.FileCache,
		loading:       false,
		pendingDelete: -1, // -1 means no pending delete
		styles:        DefaultStyles(),
		cacheStatus:   cachestatus.New(),
		editModal:     NewEditModel(deps.Ctx, deps.Svc),
	}
	m.Loader = m.Loader.SetMessage("Loading webhooks...")
	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to list webhooks for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.webhooks = nil
	m.Scroller.SetItemCount(0)
	m.err = nil
	m.pendingDelete = -1 // Clear any pending delete when changing device
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
	m.ApplySize(width, height)
	return m
}

// SetEditModalSize sets the edit modal dimensions.
// This should be called with screen-based dimensions when the modal is visible.
func (m Model) SetEditModalSize(width, height int) Model {
	m.ModalWidth = width
	m.ModalHeight = height
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
	case TestResultMsg:
		return m.handleTestResult(msg)
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.ViewRequestMsg, messages.EditRequestMsg, messages.ToggleEnableRequestMsg,
		messages.TestRequestMsg, messages.DeleteRequestMsg, messages.NewRequestMsg,
		messages.RefreshRequestMsg:
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
	case messages.ViewRequestMsg:
		return m, m.selectWebhook()
	case messages.EditRequestMsg:
		return m.handleEditKey()
	case messages.ToggleEnableRequestMsg:
		return m, m.toggleWebhook()
	case messages.TestRequestMsg:
		return m.handleTestKey()
	case messages.DeleteRequestMsg:
		return m.handleDeleteKey()
	case messages.NewRequestMsg:
		return m.handleCreateKey()
	case messages.RefreshRequestMsg:
		m.loading = true
		return m, tea.Batch(
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeWebhooks),
			m.fetchAndCacheWebhooks(),
		)
	}
	return m, nil
}

func (m Model) updateLoading(msg tea.Msg) (Model, tea.Cmd, bool) {
	result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
		switch msg.(type) {
		case LoadedMsg, ActionMsg:
			return true
		}
		return generics.IsPanelCacheMsg(msg)
	})
	m.Loader = result.Loader
	return m, result.Cmd, result.Consumed
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWebhooks {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedWebhooksData](msg.Data)
	if err == nil {
		m.webhooks = data.Webhooks
		m.Scroller.SetItemCount(len(m.webhooks))
		m.Scroller.CursorToStart()
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit LoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return LoadedMsg{Webhooks: m.webhooks} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWebhooks {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheWebhooks())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeWebhooks {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("webhooks background refresh", msg.Err)
		m.err = msg.Err
		// Emit LoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return LoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedWebhooksData); ok {
		m.webhooks = data.Webhooks
		m.Scroller.SetItemCount(len(m.webhooks))
	}
	// Emit LoadedMsg so sequential loading can advance
	return m, func() tea.Msg { return LoadedMsg{Webhooks: m.webhooks} }
}

func (m Model) handleLoaded(msg LoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.webhooks = msg.Webhooks
	m.Scroller.SetItemCount(len(m.webhooks))
	m.Scroller.CursorToStart()
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
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeWebhooks),
		m.fetchAndCacheWebhooks(),
	)
}

func (m Model) handleTestResult(msg TestResultMsg) (Model, tea.Cmd) {
	m.testing = false
	// The TestResultMsg is re-emitted so the parent view can show a toast
	return m, func() tea.Msg { return msg }
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
				m.Loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeWebhooks),
				m.fetchAndCacheWebhooks(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	m.Scroller.HandleNavigation(msg)
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Handle escape to cancel pending delete
	if msg.String() == "esc" || msg.String() == "ctrl+[" {
		if m.pendingDelete >= 0 {
			m.pendingDelete = -1
			return m, nil
		}
	}
	return m, nil
}

func (m Model) handleDeleteKey() (Model, tea.Cmd) {
	cursor := m.Scroller.Cursor()
	if len(m.webhooks) == 0 || cursor >= len(m.webhooks) {
		return m, nil
	}
	webhook := m.webhooks[cursor]

	// If this webhook is already pending delete, confirm and delete
	if m.pendingDelete == webhook.ID {
		m.pendingDelete = -1
		return m, m.deleteWebhook()
	}

	// Otherwise, mark as pending delete
	m.pendingDelete = webhook.ID
	return m, nil
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.webhooks) == 0 {
		return m, nil
	}
	cursor := m.Scroller.Cursor()
	if cursor >= len(m.webhooks) {
		return m, nil
	}
	webhook := m.webhooks[cursor]
	m.editing = true
	w, h := m.EditModalDims()
	m.editModal = m.editModal.SetSize(w, h)
	m.editModal = m.editModal.Show(m.device, &webhook)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleCreateKey() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.editing = true
	w, h := m.EditModalDims()
	m.editModal = m.editModal.SetSize(w, h)
	m.editModal = m.editModal.ShowCreate(m.device)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleTestKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.testing || len(m.webhooks) == 0 {
		return m, nil
	}
	cursor := m.Scroller.Cursor()
	if cursor >= len(m.webhooks) {
		return m, nil
	}
	webhook := m.webhooks[cursor]
	if len(webhook.URLs) == 0 {
		return m, nil
	}
	m.testing = true
	return m, m.testWebhook(webhook)
}

func (m Model) testWebhook(webhook Webhook) tea.Cmd {
	return func() tea.Msg {
		// Test the first URL with a GET request
		testURL := webhook.URLs[0]

		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, http.NoBody)
		if err != nil {
			return TestResultMsg{
				WebhookID:   webhook.ID,
				URL:         testURL,
				Err:         err,
				TestedCount: 1,
				TotalURLs:   len(webhook.URLs),
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return TestResultMsg{
				WebhookID:   webhook.ID,
				URL:         testURL,
				Err:         err,
				TestedCount: 1,
				TotalURLs:   len(webhook.URLs),
			}
		}
		if err := resp.Body.Close(); err != nil {
			iostreams.DebugErr("close response body", err)
		}

		return TestResultMsg{
			WebhookID:   webhook.ID,
			URL:         testURL,
			StatusCode:  resp.StatusCode,
			TestedCount: 1,
			TotalURLs:   len(webhook.URLs),
		}
	}
}

func (m Model) selectWebhook() tea.Cmd {
	cursor := m.Scroller.Cursor()
	if len(m.webhooks) == 0 || cursor >= len(m.webhooks) {
		return nil
	}
	webhook := m.webhooks[cursor]
	return func() tea.Msg {
		return SelectMsg{Webhook: webhook}
	}
}

func (m Model) toggleWebhook() tea.Cmd {
	cursor := m.Scroller.Cursor()
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
	cursor := m.Scroller.Cursor()
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

	r := rendering.New(m.Width, m.Height).
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

	// Show delete confirmation prompt
	if m.pendingDelete >= 0 {
		r.SetFooter(m.styles.Enabled.Render("Press 'd' again to confirm delete, Esc to cancel"))
		return
	}

	var hints []keys.Hint
	if len(m.webhooks) > 0 {
		hints = []keys.Hint{
			{Key: "e", Desc: "edit"},
			{Key: "t", Desc: "toggle"},
			{Key: "T", Desc: "test"},
			{Key: "d", Desc: "del"},
			{Key: "n", Desc: "new"},
			{Key: "r", Desc: "refresh"},
		}
	} else {
		hints = []keys.Hint{
			{Key: "n", Desc: "new"},
			{Key: "r", Desc: "refresh"},
		}
	}
	footer := theme.StyledKeybindings(keys.FormatHints(hints, keys.FooterHintWidth(m.Width)))
	if cs := m.cacheStatus.View(); cs != "" {
		footer += " | " + cs
	}
	r.SetFooter(footer)
}

// getStateContent returns content for non-list states and whether to use it.
func (m Model) getStateContent() (string, bool) {
	if m.device == "" {
		return styles.NoDeviceSelected(m.Width, m.Height), true
	}
	if m.loading {
		return m.Loader.View(), true
	}
	if m.err != nil {
		return m.getErrorContent(), true
	}
	if len(m.webhooks) == 0 {
		return styles.NoItemsConfigured("webhooks", m.Width, m.Height), true
	}
	return "", false
}

// getErrorContent returns the appropriate error message.
func (m Model) getErrorContent() string {
	if tuierrors.IsUnsupportedFeature(m.err) {
		return styles.EmptyStateWithBorder(tuierrors.UnsupportedMessage("Webhooks"), m.Width, m.Height)
	}
	return errorview.RenderInline(m.err)
}

// renderWebhookList renders the list of webhooks.
func (m Model) renderWebhookList() string {
	return generics.RenderScrollableList(generics.ListRenderConfig[Webhook]{
		Items:    m.webhooks,
		Scroller: m.Scroller,
		RenderItem: func(webhook Webhook, _ int, isCursor bool) string {
			return m.renderWebhookLine(webhook, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoWhenNeeded,
	})
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
	available := output.ContentWidth(m.Width, 4+6)
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
	cursor := m.Scroller.Cursor()
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
	return m, tea.Batch(m.Loader.Tick(), m.fetchWebhooks())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return keys.FormatHints([]keys.Hint{
		{Key: "j/k", Desc: "scroll"},
		{Key: "g/G", Desc: "top/btm"},
		{Key: "enter", Desc: "details"},
		{Key: "d", Desc: "delete"},
	}, keys.FooterHintWidth(m.Width))
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
