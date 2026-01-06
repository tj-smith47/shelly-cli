// Package scripts provides TUI components for managing device scripts.
package scripts

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Script represents a script on a device.
type Script struct {
	ID      int
	Name    string
	Enabled bool
	Running bool
}

// ListDeps holds the dependencies for the scripts list component.
type ListDeps struct {
	Ctx       context.Context
	Svc       *automation.Service
	FileCache *cache.FileCache
}

// CachedScriptsData holds scripts data for caching.
type CachedScriptsData struct {
	Scripts []Script `json:"scripts"`
}

// Validate ensures all required dependencies are set.
func (d ListDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// LoadedMsg signals that scripts were loaded.
type LoadedMsg struct {
	Scripts []Script
	Err     error
}

// ActionMsg signals a script action result.
type ActionMsg struct {
	Action   string // "start", "stop", "delete"
	ScriptID int
	Err      error
}

// SelectScriptMsg signals that a script was selected for viewing.
type SelectScriptMsg struct {
	Script Script
}

// EditScriptMsg signals that a script should be edited in external editor.
type EditScriptMsg struct {
	Script Script
}

// CreateScriptMsg signals that a new script should be created.
type CreateScriptMsg struct {
	Device string
}

// ListModel displays scripts for a device.
type ListModel struct {
	helpers.Sizable
	ctx         context.Context
	svc         *automation.Service
	fileCache   *cache.FileCache
	device      string
	scripts     []Script
	loading     bool
	err         error
	focused     bool
	panelIndex  int // 1-based panel index for Shift+N hotkey hint
	styles      ListStyles
	cacheStatus cachestatus.Model
}

// ListStyles holds styles for the list component.
type ListStyles struct {
	Running  lipgloss.Style
	Stopped  lipgloss.Style
	Disabled lipgloss.Style
	Name     lipgloss.Style
	Selected lipgloss.Style
	Status   lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
}

// DefaultListStyles returns the default styles for the script list.
func DefaultListStyles() ListStyles {
	colors := theme.GetSemanticColors()
	return ListStyles{
		Running: lipgloss.NewStyle().
			Foreground(colors.Online),
		Stopped: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Name: lipgloss.NewStyle().
			Foreground(colors.Text),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Status: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewList creates a new scripts list model.
func NewList(deps ListDeps) ListModel {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("scripts list component init", err)
		panic(fmt.Sprintf("scripts: invalid deps: %v", err))
	}

	m := ListModel{
		Sizable:     helpers.NewSizable(4, panel.NewScroller(0, 1)),
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		loading:     false,
		styles:      DefaultListStyles(),
		cacheStatus: cachestatus.New(),
	}
	m.Loader = m.Loader.SetMessage("Loading scripts...")
	return m
}

// Init returns the initial command.
func (m ListModel) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to list scripts for and triggers a fetch.
func (m ListModel) SetDevice(device string) (ListModel, tea.Cmd) {
	m.device = device
	m.scripts = nil
	m.Scroller.SetItemCount(0)
	m.err = nil
	m.cacheStatus = cachestatus.New()

	if device == "" {
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeScripts)
}

// fetchScripts creates a command to fetch scripts from the device.
func (m ListModel) fetchScripts() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		scripts, err := m.svc.ListScripts(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Script, len(scripts))
		for i, s := range scripts {
			result[i] = Script{
				ID:      s.ID,
				Name:    s.Name,
				Enabled: s.Enable,
				Running: s.Running,
			}
		}

		return LoadedMsg{Scripts: result}
	}
}

// fetchAndCacheScripts fetches fresh data and caches it.
func (m ListModel) fetchAndCacheScripts() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeScripts, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		scripts, err := m.svc.ListScripts(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Script, len(scripts))
		for i, s := range scripts {
			result[i] = Script{
				ID:      s.ID,
				Name:    s.Name,
				Enabled: s.Enable,
				Running: s.Running,
			}
		}

		return CachedScriptsData{Scripts: result}, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m ListModel) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeScripts, cache.TTLAutomation, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		scripts, err := m.svc.ListScripts(ctx, m.device)
		if err != nil {
			return nil, err
		}

		result := make([]Script, len(scripts))
		for i, s := range scripts {
			result[i] = Script{
				ID:      s.ID,
				Name:    s.Name,
				Enabled: s.Enable,
				Running: s.Running,
			}
		}

		return CachedScriptsData{Scripts: result}, nil
	})
}

// SetSize sets the component dimensions.
func (m ListModel) SetSize(width, height int) ListModel {
	m.ApplySize(width, height)
	return m
}

// SetFocused sets the focus state.
func (m ListModel) SetFocused(focused bool) ListModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m ListModel) SetPanelIndex(index int) ListModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
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

func (m ListModel) handleMessage(msg tea.Msg) (ListModel, tea.Cmd) {
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

func (m ListModel) updateLoading(msg tea.Msg) (ListModel, tea.Cmd, bool) {
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

func (m ListModel) handleCacheHit(msg panelcache.CacheHitMsg) (ListModel, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeScripts {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedScriptsData](msg.Data)
	if err == nil {
		m.scripts = data.Scripts
		m.Scroller.SetItemCount(len(m.scripts))
		m.Scroller.CursorToStart()
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit LoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return LoadedMsg{Scripts: m.scripts} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m ListModel) handleCacheMiss(msg panelcache.CacheMissMsg) (ListModel, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeScripts {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheScripts())
}

func (m ListModel) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (ListModel, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeScripts {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("scripts background refresh", msg.Err)
		m.err = msg.Err
		// Emit LoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return LoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedScriptsData); ok {
		m.scripts = data.Scripts
		m.Scroller.SetItemCount(len(m.scripts))
	}
	// Emit LoadedMsg so sequential loading can advance
	return m, func() tea.Msg { return LoadedMsg{Scripts: m.scripts} }
}

func (m ListModel) handleLoaded(msg LoadedMsg) (ListModel, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.scripts = msg.Scripts
	m.Scroller.SetItemCount(len(m.scripts))
	m.Scroller.CursorToStart()
	return m, nil
}

func (m ListModel) handleAction(msg ActionMsg) (ListModel, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh after action
	m.loading = true
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeScripts),
		m.fetchAndCacheScripts(),
	)
}

func (m ListModel) handleKey(msg tea.KeyPressMsg) (ListModel, tea.Cmd) {
	// Handle navigation keys first
	if keys.HandleScrollNavigation(msg.String(), m.Scroller) {
		return m, nil
	}

	// Handle action keys
	switch msg.String() {
	case "enter":
		// View script (open in viewer)
		return m, m.selectScript()
	case "e":
		// Edit script (open in external editor)
		return m, m.editScript()
	case "r":
		// Run/start script
		return m, m.startScript()
	case "s":
		// Stop script
		return m, m.stopScript()
	case "d":
		// Delete script
		return m, m.deleteScript()
	case "n":
		// New script - will be handled by parent
		return m, m.createScript()
	case "R":
		// Refresh list - invalidate cache and fetch fresh data
		m.loading = true
		return m, tea.Batch(
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeScripts),
			m.fetchAndCacheScripts(),
		)
	}

	return m, nil
}

func (m ListModel) selectScript() tea.Cmd {
	cursor := m.Scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	return func() tea.Msg {
		return SelectScriptMsg{Script: script}
	}
}

func (m ListModel) editScript() tea.Cmd {
	cursor := m.Scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	return func() tea.Msg {
		return EditScriptMsg{Script: script}
	}
}

func (m ListModel) startScript() tea.Cmd {
	cursor := m.Scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	if script.Running {
		return nil // Already running
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.StartScript(ctx, m.device, script.ID)
		return ActionMsg{Action: "start", ScriptID: script.ID, Err: err}
	}
}

func (m ListModel) stopScript() tea.Cmd {
	cursor := m.Scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	if !script.Running {
		return nil // Not running
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.StopScript(ctx, m.device, script.ID)
		return ActionMsg{Action: "stop", ScriptID: script.ID, Err: err}
	}
}

func (m ListModel) deleteScript() tea.Cmd {
	cursor := m.Scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteScript(ctx, m.device, script.ID)
		return ActionMsg{Action: "delete", ScriptID: script.ID, Err: err}
	}
}

func (m ListModel) createScript() tea.Cmd {
	if m.device == "" {
		return nil
	}
	return func() tea.Msg {
		return CreateScriptMsg{Device: m.device}
	}
}

// View renders the scripts list.
func (m ListModel) View() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Scripts").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings and cache status when focused
	if m.focused && m.device != "" && len(m.scripts) > 0 {
		footer := "e:edit r:run s:stop d:del n:new"
		if cs := m.cacheStatus.View(); cs != "" {
			footer = cs + " " + footer
		}
		r.SetFooter(footer)
	}

	r.SetContent(m.renderContent())
	return r.Render()
}

func (m ListModel) renderContent() string {
	if m.device == "" {
		return m.styles.Muted.Render("No device selected")
	}

	if m.loading {
		return m.Loader.View()
	}

	if m.err != nil {
		return m.renderError()
	}

	if len(m.scripts) == 0 {
		return m.styles.Muted.Render("No scripts on device")
	}

	return m.renderScriptsList()
}

func (m ListModel) renderError() string {
	errMsg := m.err.Error()
	// Detect Gen1 or unsupported device errors and show a friendly message
	if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "unknown method") ||
		strings.Contains(errMsg, "not found") {
		return m.styles.Muted.Render("Scripts not supported on this device")
	}
	return m.styles.Error.Render("Error: " + errMsg)
}

func (m ListModel) renderScriptsList() string {
	return generics.RenderScrollableList(generics.ListRenderConfig[Script]{
		Items:    m.scripts,
		Scroller: m.Scroller,
		RenderItem: func(script Script, _ int, isCursor bool) string {
			return m.renderScriptLine(script, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoWhenNeeded,
	})
}

func (m ListModel) renderScriptLine(script Script, isSelected bool) string {
	// Status icon
	var icon, status string
	switch {
	case !script.Enabled:
		icon = m.styles.Disabled.Render("-")
		status = m.styles.Status.Render("(disabled)")
	case script.Running:
		icon = m.styles.Running.Render("●")
		status = m.styles.Status.Render("(running)")
	default:
		icon = m.styles.Stopped.Render("○")
		status = m.styles.Status.Render("(stopped)")
	}

	// Name
	name := script.Name
	if name == "" {
		name = fmt.Sprintf("script_%d", script.ID)
	}

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Build line
	line := fmt.Sprintf("%s%s %s %s", selector, icon, name, status)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// SelectedScript returns the currently selected script, if any.
func (m ListModel) SelectedScript() *Script {
	cursor := m.Scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	return &m.scripts[cursor]
}

// ScriptCount returns the number of scripts.
func (m ListModel) ScriptCount() int {
	return len(m.scripts)
}

// Device returns the current device address.
func (m ListModel) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m ListModel) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m ListModel) Error() error {
	return m.err
}

// Refresh triggers a refresh of the script list.
func (m ListModel) Refresh() (ListModel, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchScripts())
}

// FooterText returns keybinding hints for the footer.
func (m ListModel) FooterText() string {
	return "j/k:scroll g/G:top/bottom space:toggle enter:edit"
}
