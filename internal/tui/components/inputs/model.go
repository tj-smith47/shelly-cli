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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
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

// ActionMsg signals an input action completed.
type ActionMsg struct {
	Action  string
	InputID int
	Err     error
}

// EditRequestMsg signals that the edit modal should be opened.
type EditRequestMsg struct {
	Device  string
	InputID int
}

// Model displays input settings for a device.
type Model struct {
	ctx         context.Context
	svc         *shelly.Service
	fileCache   *cache.FileCache
	device      string
	inputs      []shelly.InputInfo
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

	return Model{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		scroller:    panel.NewScroller(0, 10),
		loading:     false,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		loader: loading.New(
			loading.WithMessage("Loading inputs..."),
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

// SetDevice sets the device to display input settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.inputs = nil
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
	m.width = width
	m.height = height
	m.loader = m.loader.SetSize(width-4, height-4)
	visibleRows := height - 6 // Reserve space for header and footer
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
	var cmd tea.Cmd
	m.loader, cmd = m.loader.Update(msg)
	// Continue processing these messages even during loading
	switch msg.(type) {
	case LoadedMsg, panelcache.CacheHitMsg, panelcache.CacheMissMsg, panelcache.RefreshCompleteMsg:
		return m, nil, false
	default:
		if cmd != nil {
			return m, cmd, true
		}
	}
	return m, nil, false
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
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
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
		m.scroller.SetItemCount(len(m.inputs))
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
	if msg.Device != m.device || msg.DataType != cache.TypeInputs {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchAndCacheInputs())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeInputs {
		return m, nil
	}
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("inputs background refresh", msg.Err)
		return m, nil
	}
	if data, ok := msg.Data.(CachedInputsData); ok {
		m.inputs = data.Inputs
		m.scroller.SetItemCount(len(m.inputs))
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
	m.inputs = msg.Inputs
	m.scroller.SetItemCount(len(m.inputs))
	m.scroller.CursorToStart()
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
			m.loader.Tick(),
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
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeInputs),
				m.fetchAndCacheInputs(),
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
	case "r", "R":
		// Refresh list - invalidate cache and fetch fresh data
		if !m.loading && m.device != "" {
			m.loading = true
			return m, tea.Batch(
				m.loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeInputs),
				m.fetchAndCacheInputs(),
			)
		}
	case "e", "enter":
		// Open edit modal for selected input
		if len(m.inputs) > 0 && !m.loading {
			input := m.inputs[m.scroller.Cursor()]
			m.editing = true
			m.editModal = m.editModal.SetSize(m.width, m.height)
			var cmd tea.Cmd
			m.editModal, cmd = m.editModal.Show(m.device, input.ID)
			return m, tea.Batch(cmd, func() tea.Msg {
				return EditOpenedMsg{}
			})
		}
	}

	return m, nil
}

// View renders the Inputs component.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Inputs").
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

	if len(m.inputs) == 0 {
		r.SetContent(m.styles.Muted.Render("No inputs found"))
		return r.Render()
	}

	var content strings.Builder

	// Header
	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Inputs (%d):", len(m.inputs))))
	content.WriteString("\n\n")

	// Input list
	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		input := m.inputs[i]
		isSelected := m.scroller.IsCursorAt(i)
		content.WriteString(m.renderInputLine(input, isSelected))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := "e:edit R:refresh"
		if cs := m.cacheStatus.View(); cs != "" {
			footer = cs + " " + footer
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

	line := fmt.Sprintf("%s%s %s %s %s", selector, stateStr, idStr, nameStr, typeStr)

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
	return m.scroller.Cursor()
}

// Refresh triggers a refresh of the inputs.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchInputs())
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
	cursor := m.scroller.Cursor()
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
