// Package inputs provides TUI components for managing device input settings.
package inputs

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
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the Inputs component.
type Deps struct {
	Ctx       context.Context
	Svc       *shelly.Service
	FileCache *cache.FileCache
}

// CachedInputsData holds inputs data for caching.
type CachedInputsData struct {
	Inputs []shelly.InputInfo `json:"inputs"`
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

// LoadedMsg signals that inputs were loaded.
type LoadedMsg struct {
	Inputs []shelly.InputInfo
	Err    error
}

// ActionsLoadedMsg signals that input actions (webhooks) were loaded.
type ActionsLoadedMsg struct {
	Actions map[int][]InputAction // input ID → linked actions
	Err     error
}

// ActionMsg signals an input action completed.
type ActionMsg struct {
	Action  string
	InputID int
	Err     error
}

// TriggerResultMsg signals the result of triggering an input event.
type TriggerResultMsg struct {
	InputID   int
	EventType string
	Err       error
}

// EditRequestMsg signals that the edit modal should be opened.
type EditRequestMsg struct {
	Device  string
	InputID int
}

// InputAction represents a webhook linked to an input event.
type InputAction struct {
	WebhookID int
	Event     string // e.g., "input.single_push"
	URLs      []string
	Enable    bool
}

// Model displays input settings for a device.
type Model struct {
	helpers.Sizable
	ctx           context.Context
	svc           *shelly.Service
	fileCache     *cache.FileCache
	device        string
	inputs        []shelly.InputInfo
	actions       map[int][]InputAction // input ID → linked actions
	loading       bool
	editing       bool
	configActions bool // True when action modal is visible
	err           error
	focused       bool
	panelIndex    int // 1-based panel index for Shift+N hotkey hint
	styles        Styles
	editModal     EditModel
	actionModal   ActionModal
	cacheStatus   cachestatus.Model
}

// Styles holds styles for the Inputs component.
type Styles struct {
	StateOn  lipgloss.Style
	StateOff lipgloss.Style
	Type     lipgloss.Style
	Name     lipgloss.Style
	ID       lipgloss.Style
	Label    lipgloss.Style
	Selected lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
}

// DefaultStyles returns the default styles for the Inputs component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		StateOn: lipgloss.NewStyle().
			Foreground(colors.Online),
		StateOff: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Type: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Name: lipgloss.NewStyle().
			Foreground(colors.Text),
		ID: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text),
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

// New creates a new Inputs model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("inputs component init", err)
		panic(fmt.Sprintf("inputs: invalid deps: %v", err))
	}

	m := Model{
		Sizable:     helpers.NewSizable(6, panel.NewScroller(0, 10)),
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		loading:     false,
		actions:     make(map[int][]InputAction),
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		editModal:   NewEditModel(deps.Ctx, deps.Svc),
		actionModal: NewActionModal(deps.Ctx, deps.Svc),
	}
	m.Loader = m.Loader.SetMessage("Loading inputs...")
	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display input settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.inputs = nil
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
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeInputs)
}

func (m Model) fetchInputs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		inputs, err := m.svc.InputList(ctx, m.device)
		return LoadedMsg{Inputs: inputs, Err: err}
	}
}

// fetchActions fetches webhooks and filters them for input-related events.
func (m Model) fetchActions() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		hooks, err := m.svc.ListWebhooks(ctx, m.device)
		if err != nil {
			return ActionsLoadedMsg{Err: err}
		}

		// Filter webhooks for input events and group by input ID (Cid)
		actions := make(map[int][]InputAction)
		for _, h := range hooks {
			// Input events start with "input." (e.g., "input.single_push")
			if !strings.HasPrefix(h.Event, "input.") {
				continue
			}
			action := InputAction{
				WebhookID: h.ID,
				Event:     h.Event,
				URLs:      h.URLs,
				Enable:    h.Enable,
			}
			actions[h.Cid] = append(actions[h.Cid], action)
		}

		return ActionsLoadedMsg{Actions: actions}
	}
}

// fetchAndCacheInputs fetches fresh data and caches it.
func (m Model) fetchAndCacheInputs() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeInputs, cache.TTLInputs, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		inputs, err := m.svc.InputList(ctx, m.device)
		if err != nil {
			return nil, err
		}

		return CachedInputsData{Inputs: inputs}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeInputs, cache.TTLInputs, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		inputs, err := m.svc.InputList(ctx, m.device)
		if err != nil {
			return nil, err
		}

		return CachedInputsData{Inputs: inputs}, nil
	})
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.ApplySize(width, height)
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

	// Handle action modal if visible
	if m.configActions {
		return m.handleActionModalUpdate(msg)
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
		if _, ok := msg.(LoadedMsg); ok {
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
	case ActionsLoadedMsg:
		return m.handleActionsLoaded(msg)
	case TriggerResultMsg:
		// Re-emit so parent view can show toast
		return m, func() tea.Msg { return msg }
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.EditRequestMsg, messages.RefreshRequestMsg:
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
		return m.handleEditOpenKey()
	case messages.RefreshRequestMsg:
		if !m.loading && m.device != "" {
			m.loading = true
			return m, tea.Batch(
				m.Loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeInputs),
				m.fetchAndCacheInputs(),
			)
		}
	}
	return m, nil
}

func (m Model) handleCacheHit(msg panelcache.CacheHitMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeInputs {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedInputsData](msg.Data)
	if err == nil {
		m.inputs = data.Inputs
		m.Scroller.SetItemCount(len(m.inputs))
		m.Scroller.CursorToStart()
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit LoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return LoadedMsg{Inputs: m.inputs} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeInputs {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheInputs())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeInputs {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("inputs background refresh", msg.Err)
		m.err = msg.Err
		// Emit LoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return LoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedInputsData); ok {
		m.inputs = data.Inputs
		m.Scroller.SetItemCount(len(m.inputs))
	}
	// Emit LoadedMsg so sequential loading can advance
	return m, func() tea.Msg { return LoadedMsg{Inputs: m.inputs} }
}

func (m Model) handleLoaded(msg LoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.inputs = msg.Inputs
	m.Scroller.SetItemCount(len(m.inputs))
	m.Scroller.CursorToStart()
	// Fetch actions (webhooks linked to inputs) in background
	return m, m.fetchActions()
}

func (m Model) handleActionsLoaded(msg ActionsLoadedMsg) (Model, tea.Cmd) {
	if msg.Err != nil {
		// Log but don't block - actions are supplementary
		iostreams.DebugErr("load input actions", msg.Err)
		return m, nil
	}
	m.actions = msg.Actions
	return m, nil
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
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeInputs),
			m.fetchAndCacheInputs(),
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
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeInputs),
				m.fetchAndCacheInputs(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleActionModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.actionModal, cmd = m.actionModal.Update(msg)

	// Check if modal was closed
	if !m.actionModal.Visible() {
		m.configActions = false
		// Re-emit close message for parent to handle
		if closeMsg, ok := msg.(ActionModalClosedMsg); ok {
			if closeMsg.Changed {
				// Refresh actions if changes were made
				return m, tea.Batch(cmd, m.fetchActions())
			}
		}
		return m, cmd
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not in context system
	switch msg.String() {
	case "a":
		// Open action configuration modal for selected input
		return m.openActionModal()
	case "T":
		// Trigger single_push event on selected input (button type only)
		return m, m.triggerInput()
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

func (m Model) handleEditOpenKey() (Model, tea.Cmd) {
	// Open edit modal for selected input
	if len(m.inputs) > 0 && !m.loading {
		input := m.inputs[m.Scroller.Cursor()]
		m.editing = true
		m.editModal = m.editModal.SetSize(m.Width, m.Height)
		var cmd tea.Cmd
		m.editModal, cmd = m.editModal.Show(m.device, input.ID)
		return m, tea.Batch(cmd, func() tea.Msg {
			return EditOpenedMsg{}
		})
	}
	return m, nil
}

func (m Model) triggerInput() tea.Cmd {
	if m.device == "" || m.loading || len(m.inputs) == 0 {
		return nil
	}
	cursor := m.Scroller.Cursor()
	if cursor >= len(m.inputs) {
		return nil
	}
	input := m.inputs[cursor]
	// Trigger only works for button type inputs
	if input.Type != inputTypeButton {
		return func() tea.Msg {
			return TriggerResultMsg{
				InputID:   input.ID,
				EventType: "single_push",
				Err:       fmt.Errorf("trigger only works for button type inputs"),
			}
		}
	}

	device := m.device
	inputID := input.ID
	svc := m.svc
	ctx := m.ctx

	return func() tea.Msg {
		triggerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := svc.InputTrigger(triggerCtx, device, inputID, "single_push")
		return TriggerResultMsg{
			InputID:   inputID,
			EventType: "single_push",
			Err:       err,
		}
	}
}

func (m Model) openActionModal() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.inputs) == 0 {
		return m, nil
	}
	cursor := m.Scroller.Cursor()
	if cursor >= len(m.inputs) {
		return m, nil
	}
	input := m.inputs[cursor]

	// Get actions for this input
	inputActions := m.actions[input.ID]

	m.configActions = true
	m.actionModal = m.actionModal.SetSize(m.Width, m.Height)
	m.actionModal = m.actionModal.Show(m.device, input.ID, inputActions)

	return m, func() tea.Msg { return ActionModalOpenedMsg{} }
}

// View renders the Inputs component.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	// Render action modal if configuring actions
	if m.configActions {
		return m.actionModal.View()
	}

	r := rendering.New(m.Width, m.Height).
		SetTitle("Inputs").
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

	if len(m.inputs) == 0 {
		r.SetContent(styles.NoItemsFound("inputs", m.Width, m.Height))
		return r.Render()
	}

	var content strings.Builder

	// Header
	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Inputs (%d):", len(m.inputs))))
	content.WriteString("\n\n")

	// Input list with scroll indicator
	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[shelly.InputInfo]{
		Items:    m.inputs,
		Scroller: m.Scroller,
		RenderItem: func(input shelly.InputInfo, _ int, isCursor bool) string {
			return m.renderInputLine(input, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	}))

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := "e:edit a:actions T:test R:refresh"
		if cs := m.cacheStatus.View(); cs != "" {
			footer += " | " + cs
		}
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) renderInputLine(input shelly.InputInfo, isSelected bool) string {
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// State indicator
	var stateStr string
	if input.State {
		stateStr = m.styles.StateOn.Render("●")
	} else {
		stateStr = m.styles.StateOff.Render("○")
	}

	// ID
	idStr := m.styles.ID.Render(fmt.Sprintf("Input:%d", input.ID))

	// Name
	name := input.Name
	if name == "" {
		name = "(unnamed)"
	}
	nameStr := m.styles.Name.Render(name)

	// Type
	typeStr := m.styles.Type.Render(fmt.Sprintf("[%s]", input.Type))

	// Action count
	actionStr := ""
	if actions, ok := m.actions[input.ID]; ok && len(actions) > 0 {
		actionStr = m.styles.Muted.Render(fmt.Sprintf(" (%d actions)", len(actions)))
	}

	line := fmt.Sprintf("%s%s %s %s %s%s", selector, stateStr, idStr, nameStr, typeStr, actionStr)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// Inputs returns the current list of inputs.
func (m Model) Inputs() []shelly.InputInfo {
	return m.inputs
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

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.Scroller.Cursor()
}

// Refresh triggers a refresh of the inputs.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchInputs())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll e:edit r:refresh"
}

// SelectedInput returns the currently selected input, if any.
func (m Model) SelectedInput() *shelly.InputInfo {
	if len(m.inputs) == 0 {
		return nil
	}
	cursor := m.Scroller.Cursor()
	if cursor < 0 || cursor >= len(m.inputs) {
		return nil
	}
	return &m.inputs[cursor]
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
