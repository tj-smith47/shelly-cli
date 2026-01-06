// Package virtuals provides TUI components for managing virtual components.
package virtuals

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
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Virtual represents a virtual component on a device.
type Virtual struct {
	Key       string
	Type      shelly.VirtualComponentType
	ID        int
	Name      string
	BoolValue *bool
	NumValue  *float64
	StrValue  *string
	Options   []string
	Min       *float64
	Max       *float64
	Unit      *string
}

// Deps holds the dependencies for the virtuals component.
type Deps struct {
	Ctx       context.Context
	Svc       *shelly.Service
	FileCache *cache.FileCache
}

// CachedVirtualsData holds the data for the virtuals cache.
type CachedVirtualsData struct {
	Virtuals []Virtual `json:"virtuals"`
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

// LoadedMsg signals that virtual components were loaded.
type LoadedMsg struct {
	Virtuals []Virtual
	Err      error
}

// ActionMsg signals a virtual component action result.
type ActionMsg struct {
	Action string // "toggle", "set", "trigger", "delete"
	Key    string
	Err    error
}

// Model displays virtual components for a device.
type Model struct {
	ctx              context.Context
	svc              *shelly.Service
	device           string
	virtuals         []Virtual
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
	fileCache        *cache.FileCache
	cacheStatus      cachestatus.Model
}

// Styles holds styles for the virtual components list.
type Styles struct {
	TypeBoolean lipgloss.Style
	TypeNumber  lipgloss.Style
	TypeText    lipgloss.Style
	TypeEnum    lipgloss.Style
	TypeButton  lipgloss.Style
	TypeGroup   lipgloss.Style
	Value       lipgloss.Style
	Name        lipgloss.Style
	Selected    lipgloss.Style
	Error       lipgloss.Style
	Muted       lipgloss.Style
	Warning     lipgloss.Style
	Confirm     lipgloss.Style
}

// DefaultStyles returns the default styles for the virtual components list.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		TypeBoolean: lipgloss.NewStyle().
			Foreground(colors.Online),
		TypeNumber: lipgloss.NewStyle().
			Foreground(colors.Warning),
		TypeText: lipgloss.NewStyle().
			Foreground(colors.Info),
		TypeEnum: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		TypeButton: lipgloss.NewStyle().
			Foreground(colors.Error),
		TypeGroup: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
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
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Confirm: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
	}
}

// New creates a new virtuals model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("virtuals component init", err)
		panic(fmt.Sprintf("virtuals: invalid deps: %v", err))
	}

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		scroller:    panel.NewScroller(0, 10),
		loading:     false,
		styles:      DefaultStyles(),
		fileCache:   deps.FileCache,
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading virtual components..."),
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

// SetDevice sets the device to list virtual components for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.virtuals = nil
	m.scroller.SetItemCount(0)
	m.scroller.CursorToStart()
	m.err = nil
	m.cacheStatus = cachestatus.New()

	if device == "" {
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeVirtuals)
}

// fetchVirtuals creates a command to fetch virtual components from the device.
func (m Model) fetchVirtuals() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		components, err := m.svc.ListVirtualComponents(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Virtual, len(components))
		for i, c := range components {
			result[i] = Virtual{
				Key:       c.Key,
				Type:      c.Type,
				ID:        c.ID,
				Name:      c.Name,
				BoolValue: c.BoolValue,
				NumValue:  c.NumValue,
				StrValue:  c.StrValue,
				Options:   c.Options,
				Min:       c.Min,
				Max:       c.Max,
				Unit:      c.Unit,
			}
		}

		return LoadedMsg{Virtuals: result}
	}
}

// fetchAndCacheVirtuals fetches fresh data and caches it.
func (m Model) fetchAndCacheVirtuals() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeVirtuals, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		components, err := m.svc.ListVirtualComponents(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Virtual, len(components))
		for i, c := range components {
			result[i] = Virtual{
				Key:       c.Key,
				Type:      c.Type,
				ID:        c.ID,
				Name:      c.Name,
				BoolValue: c.BoolValue,
				NumValue:  c.NumValue,
				StrValue:  c.StrValue,
				Options:   c.Options,
				Min:       c.Min,
				Max:       c.Max,
				Unit:      c.Unit,
			}
		}

		return CachedVirtualsData{Virtuals: result}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeVirtuals, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		components, err := m.svc.ListVirtualComponents(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Virtual, len(components))
		for i, c := range components {
			result[i] = Virtual{
				Key:       c.Key,
				Type:      c.Type,
				ID:        c.ID,
				Name:      c.Name,
				BoolValue: c.BoolValue,
				NumValue:  c.NumValue,
				StrValue:  c.StrValue,
				Options:   c.Options,
				Min:       c.Min,
				Max:       c.Max,
				Unit:      c.Unit,
			}
		}

		return CachedVirtualsData{Virtuals: result}, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
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
	if msg.Device != m.device || msg.DataType != cache.TypeVirtuals {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedVirtualsData](msg.Data)
	if err == nil {
		m.virtuals = data.Virtuals
		m.scroller.SetItemCount(len(m.virtuals))
		m.scroller.CursorToStart()
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit LoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return LoadedMsg{Virtuals: m.virtuals} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeVirtuals {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheVirtuals())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeVirtuals {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("virtuals background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedVirtualsData); ok {
		m.virtuals = data.Virtuals
		m.scroller.SetItemCount(len(m.virtuals))
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
	m.virtuals = msg.Virtuals
	m.scroller.SetItemCount(len(m.virtuals))
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
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeVirtuals),
		m.fetchAndCacheVirtuals(),
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
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeVirtuals),
			m.fetchAndCacheVirtuals(),
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
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeVirtuals),
				m.fetchAndCacheVirtuals(),
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
	case "t", "enter":
		return m, m.toggleOrTrigger()
	case "e":
		return m.handleEditKey()
	case "n":
		return m.handleNewKey()
	case "h", "left":
		return m, m.adjustValue(-1)
	case "l", "right":
		return m, m.adjustValue(1)
	case "d":
		return m.handleDeleteKey()
	case "R":
		// Refresh list - invalidate cache and fetch fresh data
		m.loading = true
		return m, tea.Batch(
			m.loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeVirtuals),
			m.fetchAndCacheVirtuals(),
		)
	}

	return m, nil
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.virtuals) == 0 {
		return m, nil
	}
	cursor := m.scroller.Cursor()
	if cursor >= len(m.virtuals) {
		return m, nil
	}
	v := m.virtuals[cursor]

	// Can't edit groups
	if v.Type == shelly.VirtualGroup {
		return m, nil
	}

	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowEdit(m.device, &v)
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
	if m.device == "" || m.loading || len(m.virtuals) == 0 {
		return m, nil
	}
	cursor := m.scroller.Cursor()
	if cursor >= len(m.virtuals) {
		return m, nil
	}
	// Start delete confirmation
	m.confirmingDelete = true
	m.deleteKey = m.virtuals[cursor].Key
	return m, nil
}

func (m Model) executeDelete(key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteVirtualComponent(ctx, m.device, key)
		return ActionMsg{Action: "delete", Key: key, Err: err}
	}
}

func (m Model) toggleOrTrigger() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.virtuals) == 0 || cursor >= len(m.virtuals) {
		return nil
	}
	v := m.virtuals[cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var err error
		action := "toggle"

		switch v.Type {
		case shelly.VirtualBoolean:
			err = m.svc.ToggleVirtualBoolean(ctx, m.device, v.ID)
		case shelly.VirtualButton:
			err = m.svc.TriggerVirtualButton(ctx, m.device, v.ID)
			action = "trigger"
		default:
			return nil // No action for other types via toggle
		}

		return ActionMsg{Action: action, Key: v.Key, Err: err}
	}
}

func (m Model) adjustValue(delta int) tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.virtuals) == 0 || cursor >= len(m.virtuals) {
		return nil
	}
	v := m.virtuals[cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var err error
		switch v.Type {
		case shelly.VirtualNumber:
			if v.NumValue != nil {
				newVal := *v.NumValue + float64(delta)
				if v.Min != nil && newVal < *v.Min {
					newVal = *v.Min
				}
				if v.Max != nil && newVal > *v.Max {
					newVal = *v.Max
				}
				err = m.svc.SetVirtualNumber(ctx, m.device, v.ID, newVal)
			}
		case shelly.VirtualEnum:
			if len(v.Options) > 0 && v.StrValue != nil {
				idx := findIndex(v.Options, *v.StrValue)
				if idx >= 0 {
					newIdx := (idx + delta + len(v.Options)) % len(v.Options)
					err = m.svc.SetVirtualEnum(ctx, m.device, v.ID, v.Options[newIdx])
				}
			}
		default:
			return nil
		}

		return ActionMsg{Action: "set", Key: v.Key, Err: err}
	}
}

func findIndex(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i
		}
	}
	return -1
}

// View renders the virtual components list.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Virtual Components").
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
			r.SetContent(m.styles.Muted.Render("Virtual components not supported on this device"))
		} else {
			r.SetContent(m.styles.Error.Render("Error: " + errMsg))
		}
		return r.Render()
	}

	// Show delete confirmation
	if m.confirmingDelete {
		var content strings.Builder
		content.WriteString(m.styles.Confirm.Render("Delete virtual component: " + m.deleteKey + "?"))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Warning.Render("Press Y to confirm, N or Esc to cancel"))
		r.SetContent(content.String())
		return r.Render()
	}

	if len(m.virtuals) == 0 {
		r.SetContent(m.styles.Muted.Render("No virtual components"))
		return r.Render()
	}

	var content strings.Builder

	// Virtual list with scroll indicator
	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[Virtual]{
		Items:    m.virtuals,
		Scroller: m.scroller,
		RenderItem: func(v Virtual, _ int, isCursor bool) string {
			return m.renderVirtualLine(v, isCursor)
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

func (m Model) renderVirtualLine(v Virtual, isSelected bool) string {
	// Type indicator
	var typeStr string
	switch v.Type {
	case shelly.VirtualBoolean:
		typeStr = m.styles.TypeBoolean.Render("BOOL")
	case shelly.VirtualNumber:
		typeStr = m.styles.TypeNumber.Render("NUM ")
	case shelly.VirtualText:
		typeStr = m.styles.TypeText.Render("TEXT")
	case shelly.VirtualEnum:
		typeStr = m.styles.TypeEnum.Render("ENUM")
	case shelly.VirtualButton:
		typeStr = m.styles.TypeButton.Render("BTN ")
	case shelly.VirtualGroup:
		typeStr = m.styles.TypeGroup.Render("GRP ")
	default:
		typeStr = m.styles.Muted.Render("??? ")
	}

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Calculate available width for name and value
	// Fixed: selector(2) + type(4) + spaces(2) = 8
	available := output.ContentWidth(m.width, 4+8)
	nameWidth, valueWidth := output.SplitWidth(available, 40, 10, 15)

	// Name or ID
	name := v.Name
	if name == "" {
		name = fmt.Sprintf("#%d", v.ID)
	}
	name = output.Truncate(name, nameWidth)
	nameStr := m.styles.Name.Render(fmt.Sprintf("%-*s", nameWidth, name))

	// Value display
	valueStr := m.formatValueWithWidth(v, valueWidth)

	line := fmt.Sprintf("%s%s %s %s", selector, typeStr, nameStr, valueStr)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

func (m Model) formatValueWithWidth(v Virtual, maxWidth int) string {
	switch v.Type {
	case shelly.VirtualBoolean:
		if v.BoolValue != nil {
			if *v.BoolValue {
				return m.styles.TypeBoolean.Render("ON")
			}
			return m.styles.Muted.Render("OFF")
		}
	case shelly.VirtualNumber:
		if v.NumValue != nil {
			val := fmt.Sprintf("%.1f", *v.NumValue)
			if v.Unit != nil {
				val += *v.Unit
			}
			return m.styles.Value.Render(val)
		}
	case shelly.VirtualText:
		if v.StrValue != nil {
			textWidth := maxWidth - 2 // Account for quotes
			if textWidth < 10 {
				textWidth = 10
			}
			text := output.Truncate(*v.StrValue, textWidth)
			return m.styles.Value.Render(fmt.Sprintf("%q", text))
		}
	case shelly.VirtualEnum:
		if v.StrValue != nil {
			return m.styles.Value.Render(*v.StrValue)
		}
	case shelly.VirtualButton:
		return m.styles.Muted.Render("[Press]")
	case shelly.VirtualGroup:
		return m.styles.Muted.Render("(group)")
	}
	return m.styles.Muted.Render("-")
}

// SelectedVirtual returns the currently selected virtual component, if any.
func (m Model) SelectedVirtual() *Virtual {
	cursor := m.scroller.Cursor()
	if len(m.virtuals) == 0 || cursor >= len(m.virtuals) {
		return nil
	}
	return &m.virtuals[cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.scroller.Cursor()
}

// VirtualCount returns the number of virtual components.
func (m Model) VirtualCount() int {
	return len(m.virtuals)
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

// Refresh triggers a refresh of the virtual components list.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchVirtuals())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom enter:details"
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
