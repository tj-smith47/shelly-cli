// Package kvs provides TUI components for browsing device key-value store.
package kvs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	shellykvs "github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Item represents a key-value pair in the device KVS.
type Item struct {
	Key   string
	Value any
	Etag  string
}

// Deps holds the dependencies for the KVS browser component.
type Deps struct {
	Ctx       context.Context
	Svc       *shellykvs.Service
	FileCache *cache.FileCache
}

// CachedKVSData holds KVS data for caching.
type CachedKVSData struct {
	Items []Item `json:"items"`
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// LoadedMsg signals that KVS items were loaded.
type LoadedMsg struct {
	Items []Item
	Err   error
}

// ActionMsg signals a KVS action result.
type ActionMsg struct {
	Action string // "set", "delete"
	Key    string
	Err    error
}

// SelectMsg signals that an item was selected for viewing.
type SelectMsg struct {
	Item Item
}

// Model displays KVS items for a device.
type Model struct {
	ctx              context.Context
	svc              *shellykvs.Service
	fileCache        *cache.FileCache
	device           string
	items            []Item
	scroller         *panel.Scroller
	loading          bool
	editing          bool
	confirmingDelete bool
	deleteKey        string
	err              error
	width            int
	height           int
	focused          bool
	panelIndex       int // 1-based panel index for Shift+N hotkey hint
	styles           Styles
	loader           loading.Model
	editModal        EditModel
	cacheStatus      cachestatus.Model
}

// Styles holds styles for the KVS browser component.
type Styles struct {
	Key      lipgloss.Style
	Value    lipgloss.Style
	String   lipgloss.Style
	Number   lipgloss.Style
	Bool     lipgloss.Style
	Null     lipgloss.Style
	Object   lipgloss.Style
	Selected lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
	Warning  lipgloss.Style
	Confirm  lipgloss.Style
}

// DefaultStyles returns the default styles for the KVS browser.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Key: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		String: lipgloss.NewStyle().
			Foreground(colors.Online),
		Number: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Bool: lipgloss.NewStyle().
			Foreground(colors.Info),
		Null: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Object: lipgloss.NewStyle().
			Foreground(colors.Error),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Confirm: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
	}
}

// New creates a new KVS browser model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("kvs component init", err)
		panic(fmt.Sprintf("kvs: invalid deps: %v", err))
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
			loading.WithMessage("Loading KVS..."),
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

// SetDevice sets the device to browse KVS for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.items = nil
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
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeKVS)
}

// fetchItems creates a command to fetch KVS items from the device.
func (m Model) fetchItems() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		kvsItems, err := m.svc.GetAll(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Item, len(kvsItems))
		for i, item := range kvsItems {
			result[i] = Item{
				Key:   item.Key,
				Value: item.Value,
				Etag:  item.Etag,
			}
		}

		return LoadedMsg{Items: result}
	}
}

// fetchAndCacheItems fetches fresh data and caches it.
func (m Model) fetchAndCacheItems() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeKVS, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		kvsItems, err := m.svc.GetAll(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Item, len(kvsItems))
		for i, item := range kvsItems {
			result[i] = Item{
				Key:   item.Key,
				Value: item.Value,
				Etag:  item.Etag,
			}
		}

		return CachedKVSData{Items: result}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeKVS, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		kvsItems, err := m.svc.GetAll(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Item, len(kvsItems))
		for i, item := range kvsItems {
			result[i] = Item{
				Key:   item.Key,
				Value: item.Value,
				Etag:  item.Etag,
			}
		}

		return CachedKVSData{Items: result}, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.loader = m.loader.SetSize(width-4, height-4)
	visibleRows := height - 4
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
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

	// Handle delete confirmation
	if m.confirmingDelete {
		return m.handleDeleteConfirmation(msg)
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

func (m Model) updateLoading(msg tea.Msg) (Model, tea.Cmd, bool) {
	result := generics.UpdateLoader(m.loader, msg, func(msg tea.Msg) bool {
		switch msg.(type) {
		case LoadedMsg, ActionMsg:
			return true
		}
		return generics.IsPanelCacheMsg(msg)
	})
	m.loader = result.Loader
	return m, result.Cmd, result.Consumed
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

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeKVS {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedKVSData](msg.Data)
	if err == nil {
		m.items = data.Items
		m.scroller.SetItemCount(len(m.items))
		m.scroller.CursorToStart()
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit LoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return LoadedMsg{Items: m.items} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeKVS {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheItems())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeKVS {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("kvs background refresh", msg.Err)
		m.err = msg.Err
		// Emit LoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return LoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedKVSData); ok {
		m.items = data.Items
		m.scroller.SetItemCount(len(m.items))
	}
	// Emit LoadedMsg so sequential loading can advance
	return m, func() tea.Msg { return LoadedMsg{Items: m.items} }
}

func (m Model) handleLoaded(msg LoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.items = msg.Items
	m.scroller.SetItemCount(len(m.items))
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
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeKVS),
		m.fetchAndCacheItems(),
		func() tea.Msg { return EditClosedMsg{Saved: true} },
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
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeKVS),
			m.fetchAndCacheItems(),
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
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeKVS),
				m.fetchAndCacheItems(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleDeleteConfirmation(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "y", "Y":
			m.confirmingDelete = false
			return m, m.executeDelete(m.deleteKey)
		case "n", "N", "esc":
			m.confirmingDelete = false
			m.deleteKey = ""
			return m, nil
		}
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Handle navigation keys first
	if keys.HandleScrollNavigation(msg.String(), m.scroller) {
		return m, nil
	}

	// Handle action keys
	switch msg.String() {
	case "enter":
		return m, m.selectItem()
	case "e":
		return m.handleEditKey()
	case "n":
		return m.handleNewKey()
	case "d":
		return m.handleDeleteKey()
	case "R":
		// Refresh list - invalidate cache and fetch fresh data
		m.loading = true
		return m, tea.Batch(
			m.loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeKVS),
			m.fetchAndCacheItems(),
		)
	}

	return m, nil
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.items) == 0 {
		return m, nil
	}
	cursor := m.scroller.Cursor()
	if cursor >= len(m.items) {
		return m, nil
	}
	item := m.items[cursor]
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowEdit(m.device, &item)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleNewKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowNew(m.device)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleDeleteKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.items) == 0 {
		return m, nil
	}
	cursor := m.scroller.Cursor()
	if cursor >= len(m.items) {
		return m, nil
	}
	// Start delete confirmation
	m.confirmingDelete = true
	m.deleteKey = m.items[cursor].Key
	return m, nil
}

func (m Model) executeDelete(key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Delete(ctx, m.device, key)
		return ActionMsg{Action: "delete", Key: key, Err: err}
	}
}

func (m Model) selectItem() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.items) == 0 || cursor >= len(m.items) {
		return nil
	}
	item := m.items[cursor]
	return func() tea.Msg {
		return SelectMsg{Item: item}
	}
}

// View renders the KVS browser.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Key-Value Store").
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
		errMsg := m.err.Error()
		// Detect Gen1 or unsupported device errors and show a friendly message
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "unknown method") ||
			strings.Contains(errMsg, "not found") {
			r.SetContent(m.styles.Muted.Render("KVS not supported on this device"))
		} else {
			r.SetContent(m.styles.Error.Render("Error: " + errMsg))
		}
		return r.Render()
	}

	// Show delete confirmation
	if m.confirmingDelete {
		var content strings.Builder
		content.WriteString(m.styles.Confirm.Render("Delete key: " + m.deleteKey + "?"))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Warning.Render("Press Y to confirm, N or Esc to cancel"))
		r.SetContent(content.String())
		return r.Render()
	}

	if len(m.items) == 0 {
		r.SetContent(m.styles.Muted.Render("No KVS entries"))
		return r.Render()
	}

	var content strings.Builder

	// KVS items with scroll indicator
	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[Item]{
		Items:    m.items,
		Scroller: m.scroller,
		RenderItem: func(item Item, _ int, isCursor bool) string {
			return m.renderItemLine(item, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	}))

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := "e:edit d:delete R:refresh"
		if cs := m.cacheStatus.View(); cs != "" {
			footer = cs + " " + footer
		}
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) renderItemLine(item Item, isSelected bool) string {
	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Calculate available width for key and value
	// Fixed: selector(2) + " = "(3) = 5
	available := output.ContentWidth(m.width, 4+5)

	// Dynamic width allocation based on actual content
	keyWidth, valueWidth := m.calculateKeyValueWidths(available, item)

	// Key (truncate if too long)
	key := output.Truncate(item.Key, keyWidth)
	keyStr := m.styles.Key.Render(fmt.Sprintf("%-*s", keyWidth, key))

	// Value display
	valueStr := m.formatValueWithWidth(item.Value, valueWidth)

	line := fmt.Sprintf("%s%s = %s", selector, keyStr, valueStr)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// calculateKeyValueWidths dynamically allocates width between key and value.
// It examines all items to find the actual max key width, then gives remaining space to values.
func (m Model) calculateKeyValueWidths(available int, _ Item) (keyW, valueW int) {
	const minKeyW, minValueW = 10, 15

	// Find the actual max key width from all items
	maxKeyLen := minKeyW
	for _, item := range m.items {
		if len(item.Key) > maxKeyLen {
			maxKeyLen = len(item.Key)
		}
	}

	// Key gets its actual needed width (plus 1 for spacing), capped by available - minValueW
	keyW = min(maxKeyLen+1, available-minValueW)
	if keyW < minKeyW {
		keyW = minKeyW
	}

	// Value gets remaining space
	valueW = available - keyW
	if valueW < minValueW {
		valueW = minValueW
	}

	return keyW, valueW
}

func (m Model) formatValueWithWidth(value any, maxWidth int) string {
	if value == nil {
		return m.styles.Null.Width(maxWidth).Render("null")
	}

	switch v := value.(type) {
	case string:
		// Quote first, then truncate - %q escapes add characters
		quoted := fmt.Sprintf("%q", v)
		display := output.Truncate(quoted, maxWidth)
		return m.styles.String.Width(maxWidth).Render(display)
	case float64:
		var numStr string
		if v == float64(int64(v)) {
			numStr = fmt.Sprintf("%d", int64(v))
		} else {
			numStr = fmt.Sprintf("%g", v)
		}
		return m.styles.Number.Width(maxWidth).Render(numStr)
	case bool:
		return m.styles.Bool.Width(maxWidth).Render(fmt.Sprintf("%v", v))
	case map[string]any, []any:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return m.styles.Object.Width(maxWidth).Render("{...}")
		}
		display := output.Truncate(string(jsonBytes), maxWidth)
		return m.styles.Object.Width(maxWidth).Render(display)
	default:
		return m.styles.Value.Width(maxWidth).Render(fmt.Sprintf("%v", v))
	}
}

// SelectedItem returns the currently selected item, if any.
func (m Model) SelectedItem() *Item {
	cursor := m.scroller.Cursor()
	if len(m.items) == 0 || cursor >= len(m.items) {
		return nil
	}
	return &m.items[cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.scroller.Cursor()
}

// ItemCount returns the number of items.
func (m Model) ItemCount() int {
	return len(m.items)
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

// Refresh triggers a refresh of the KVS items.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchItems())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll e:edit n:new d:delete r:refresh"
}

// IsEditing returns whether the edit modal is currently open.
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
