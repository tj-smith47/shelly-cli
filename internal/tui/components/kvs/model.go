// Package kvs provides TUI components for browsing device key-value store.
package kvs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
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

// ExportedMsg signals that KVS was exported to a file.
type ExportedMsg struct {
	Path string
	Err  error
}

// ImportedMsg signals that KVS was imported from a file.
type ImportedMsg struct {
	Path  string
	Count int
	Err   error
}

// Model displays KVS items for a device.
type Model struct {
	panel.Sizable    // Embeds Width, Height, Loader, Scroller
	ctx              context.Context
	svc              *shellykvs.Service
	fileCache        *cache.FileCache
	device           string
	items            []Item
	loading          bool
	editing          bool
	confirmingDelete bool
	deleteKey        string
	err              error
	focused          bool
	panelIndex       int // 1-based panel index for Shift+N hotkey hint
	styles           Styles
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

	m := Model{
		Sizable:     panel.NewSizable(4, panel.NewScroller(0, 10)),
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		editModal:   NewEditModel(deps.Ctx, deps.Svc),
	}
	m.Loader = m.Loader.SetMessage("Loading KVS...")
	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to browse KVS for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.items = nil
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
	case ExportedMsg:
		return m.handleExported(msg)
	case ImportedMsg:
		return m.handleImported(msg)
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.EditRequestMsg, messages.NewRequestMsg, messages.DeleteRequestMsg,
		messages.RefreshRequestMsg, messages.ExportRequestMsg, messages.ImportRequestMsg:
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
	case messages.EditRequestMsg:
		return m.handleEditKey()
	case messages.NewRequestMsg:
		return m.handleNewKey()
	case messages.DeleteRequestMsg:
		return m.handleDeleteKey()
	case messages.RefreshRequestMsg:
		m.loading = true
		return m, tea.Batch(
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeKVS),
			m.fetchAndCacheItems(),
		)
	case messages.ExportRequestMsg:
		return m, m.exportKVS()
	case messages.ImportRequestMsg:
		return m, m.importKVS()
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
		m.Scroller.SetItemCount(len(m.items))
		m.Scroller.CursorToStart()
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
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheItems())
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
		m.Scroller.SetItemCount(len(m.items))
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
	m.Scroller.SetItemCount(len(m.items))
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
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeKVS),
		m.fetchAndCacheItems(),
		func() tea.Msg { return EditClosedMsg{Saved: true} },
	)
}

func (m Model) handleExported(msg ExportedMsg) (Model, tea.Cmd) {
	// Re-emit so parent view can show toast
	return m, func() tea.Msg { return msg }
}

func (m Model) handleImported(msg ImportedMsg) (Model, tea.Cmd) {
	// If import was successful, refresh the list
	if msg.Err == nil {
		m.loading = true
		return m, tea.Batch(
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeKVS),
			m.fetchAndCacheItems(),
			func() tea.Msg { return msg },
		)
	}
	// Re-emit so parent view can show toast
	return m, func() tea.Msg { return msg }
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
				m.Loader.Tick(),
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
	// Component-specific keys not in context system
	if msg.String() == "enter" {
		return m, m.selectItem()
	}

	return m, nil
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

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.items) == 0 {
		return m, nil
	}
	cursor := m.Scroller.Cursor()
	if cursor >= len(m.items) {
		return m, nil
	}
	item := m.items[cursor]
	m.editing = true
	w, h := m.EditModalDims()
	m.editModal = m.editModal.SetSize(w, h)
	m.editModal = m.editModal.ShowEdit(m.device, &item)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleNewKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading {
		return m, nil
	}
	m.editing = true
	w, h := m.EditModalDims()
	m.editModal = m.editModal.SetSize(w, h)
	m.editModal = m.editModal.ShowNew(m.device)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleDeleteKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.items) == 0 {
		return m, nil
	}
	cursor := m.Scroller.Cursor()
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
	cursor := m.Scroller.Cursor()
	if len(m.items) == 0 || cursor >= len(m.items) {
		return nil
	}
	item := m.items[cursor]
	return func() tea.Msg {
		return SelectMsg{Item: item}
	}
}

// kvsExportData represents the JSON structure for KVS export.
type kvsExportData struct {
	Device string          `json:"device"`
	Items  []kvsExportItem `json:"items"`
}

type kvsExportItem struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func (m Model) exportKVS() tea.Cmd {
	if m.device == "" || len(m.items) == 0 {
		return nil
	}

	device := m.device
	items := m.items

	return func() tea.Msg {
		// Create kvs directory if it doesn't exist
		dir := "kvs"
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return ExportedMsg{Err: err}
		}

		// Generate filename: kvs/<device>_kvs.json
		safeDevice := strings.ReplaceAll(device, ".", "_")
		filename := fmt.Sprintf("%s/%s_kvs.json", dir, safeDevice)

		// Build export data
		exportItems := make([]kvsExportItem, len(items))
		for i, item := range items {
			exportItems[i] = kvsExportItem{
				Key:   item.Key,
				Value: item.Value,
			}
		}
		data := kvsExportData{
			Device: device,
			Items:  exportItems,
		}

		// Marshal to JSON with indentation
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return ExportedMsg{Err: err}
		}

		// Write file
		if err := os.WriteFile(filename, jsonData, 0o600); err != nil {
			return ExportedMsg{Err: err}
		}

		return ExportedMsg{Path: filename}
	}
}

func (m Model) importKVS() tea.Cmd {
	if m.device == "" {
		return nil
	}

	device := m.device
	svc := m.svc
	ctx := m.ctx

	return func() tea.Msg {
		// Generate filename: kvs/<device>_kvs.json
		safeDevice := strings.ReplaceAll(device, ".", "_")
		filename := fmt.Sprintf("kvs/%s_kvs.json", safeDevice)

		// Read file
		data, err := os.ReadFile(filename) //nolint:gosec // G304: User controls filename via device selection
		if err != nil {
			return ImportedMsg{Path: filename, Err: err}
		}

		// Parse JSON
		var exportData kvsExportData
		if err := json.Unmarshal(data, &exportData); err != nil {
			return ImportedMsg{Path: filename, Err: fmt.Errorf("invalid JSON: %w", err)}
		}

		// Import each item
		importCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		importedCount := 0
		for _, item := range exportData.Items {
			if err := svc.Set(importCtx, device, item.Key, item.Value); err != nil {
				return ImportedMsg{Path: filename, Count: importedCount, Err: err}
			}
			importedCount++
		}

		return ImportedMsg{Path: filename, Count: importedCount}
	}
}

// View renders the KVS browser.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.Width, m.Height).
		SetTitle("Key-Value Store").
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
		if tuierrors.IsUnsupportedFeature(m.err) {
			r.SetContent(styles.EmptyStateWithBorder(tuierrors.UnsupportedMessage("KVS"), m.Width, m.Height))
		} else {
			msg, _ := tuierrors.FormatError(m.err)
			r.SetContent(m.styles.Error.Render(msg))
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
		r.SetContent(styles.EmptyStateWithBorder("No KVS entries", m.Width, m.Height))
		return r.Render()
	}

	var content strings.Builder

	// KVS items with scroll indicator
	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[Item]{
		Items:    m.items,
		Scroller: m.Scroller,
		RenderItem: func(item Item, _ int, isCursor bool) string {
			return m.renderItemLine(item, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	}))

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := theme.StyledKeybindings(keys.FormatHints([]keys.Hint{
			{Key: "e", Desc: "edit"},
			{Key: "d", Desc: "del"},
			{Key: "n", Desc: "new"},
			{Key: "X", Desc: "export"},
			{Key: "I", Desc: "import"},
			{Key: "R", Desc: "refresh"},
		}, keys.FooterHintWidth(m.Width)))
		if cs := m.cacheStatus.View(); cs != "" {
			footer += " | " + cs
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
	available := output.ContentWidth(m.Width, 4+5)

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
	cursor := m.Scroller.Cursor()
	if len(m.items) == 0 || cursor >= len(m.items) {
		return nil
	}
	return &m.items[cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.Scroller.Cursor()
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
	return m, tea.Batch(m.Loader.Tick(), m.fetchItems())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return keys.FormatHints([]keys.Hint{
		{Key: "j/k", Desc: "scroll"},
		{Key: "e", Desc: "edit"},
		{Key: "n", Desc: "new"},
		{Key: "d", Desc: "delete"},
		{Key: "r", Desc: "refresh"},
	}, keys.FooterHintWidth(m.Width))
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
