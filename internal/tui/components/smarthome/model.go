// Package smarthome provides TUI components for managing smart home protocol settings.
// This includes Matter, Zigbee, and LoRa configuration.
package smarthome

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
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the SmartHome component.
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

// CachedSmartHomeData holds smart home status for caching.
type CachedSmartHomeData struct {
	Matter *shelly.TUIMatterStatus `json:"matter"`
	Zigbee *shelly.TUIZigbeeStatus `json:"zigbee"`
	LoRa   *shelly.TUILoRaStatus   `json:"lora"`
}

// Protocol identifies which protocol section is focused.
type Protocol int

// Protocol constants.
const (
	ProtocolMatter Protocol = iota // Matter protocol section
	ProtocolZigbee                 // Zigbee protocol section
	ProtocolLoRa                   // LoRa protocol section
)

// Zigbee network state constants.
const (
	zigbeeStateJoined   = "joined"
	zigbeeStateSteering = "steering"
	zigbeeStateReady    = "ready"
)

// StatusLoadedMsg signals that smart home statuses were loaded.
type StatusLoadedMsg struct {
	Matter *shelly.TUIMatterStatus
	Zigbee *shelly.TUIZigbeeStatus
	LoRa   *shelly.TUILoRaStatus
	Err    error
}

// Model displays smart home protocol settings for a device.
type Model struct {
	helpers.Sizable
	ctx            context.Context
	svc            *shelly.Service
	fileCache      *cache.FileCache
	device         string
	matter         *shelly.TUIMatterStatus
	zigbee         *shelly.TUIZigbeeStatus
	lora           *shelly.TUILoRaStatus
	activeProtocol Protocol
	loading        bool
	toggling       bool // Matter/Zigbee enable/disable toggle in progress
	pendingReset   bool // Awaiting Matter reset confirmation (double-press)
	pendingLeave   bool // Awaiting Zigbee leave network confirmation (double-press)
	editing        bool // Matter edit modal visible
	zigbeeEditing  bool // Zigbee edit modal visible
	err            error
	focused        bool
	panelIndex     int // 1-based panel index for Shift+N hotkey hint
	styles         Styles
	cacheStatus    cachestatus.Model
	editModal      MatterEditModel
	zigbeeModal    ZigbeeEditModel
}

// Styles holds styles for the SmartHome component.
type Styles struct {
	Enabled   lipgloss.Style
	Disabled  lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Highlight lipgloss.Style
	Error     lipgloss.Style
	Muted     lipgloss.Style
	Section   lipgloss.Style
	Active    lipgloss.Style
	Warning   lipgloss.Style
}

// DefaultStyles returns the default styles for the SmartHome component.
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
		Active: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
	}
}

// New creates a new SmartHome model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("smarthome component init", err)
		panic(fmt.Sprintf("smarthome: invalid deps: %v", err))
	}

	m := Model{
		Sizable:     helpers.NewSizableLoaderOnly(),
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		fileCache:   deps.FileCache,
		styles:      DefaultStyles(),
		cacheStatus: cachestatus.New(),
		editModal:   NewMatterEditModel(deps.Ctx, deps.Svc),
		zigbeeModal: NewZigbeeEditModel(deps.Ctx, deps.Svc),
	}
	m.Loader = m.Loader.SetMessage("Loading smart home protocols...")
	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display smart home settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.matter = nil
	m.zigbee = nil
	m.lora = nil
	m.activeProtocol = ProtocolMatter
	m.err = nil
	m.loading = true
	m.toggling = false
	m.pendingReset = false
	m.pendingLeave = false
	m.editing = false
	m.zigbeeEditing = false
	m.editModal = m.editModal.Hide()
	m.zigbeeModal = m.zigbeeModal.Hide()
	m.cacheStatus = cachestatus.New()

	if device == "" {
		m.loading = false
		return m, nil
	}

	// Try to load from cache first
	return m, panelcache.LoadWithCache(m.fileCache, device, cache.TypeMatter)
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		var msg StatusLoadedMsg
		msg.Matter = m.fetchMatter(ctx)
		msg.Zigbee = m.fetchZigbee(ctx)
		msg.LoRa = m.fetchLoRa(ctx)

		return msg
	}
}

func (m Model) fetchMatter(ctx context.Context) *shelly.TUIMatterStatus {
	status, err := m.svc.GetTUIMatterStatus(ctx, m.device)
	if err != nil {
		return nil
	}
	return status
}

func (m Model) fetchZigbee(ctx context.Context) *shelly.TUIZigbeeStatus {
	status, err := m.svc.GetTUIZigbeeStatus(ctx, m.device)
	if err != nil {
		return nil
	}
	return status
}

func (m Model) fetchLoRa(ctx context.Context) *shelly.TUILoRaStatus {
	status, err := m.svc.GetTUILoRaStatus(ctx, m.device)
	if err != nil {
		return nil
	}
	return status
}

// fetchAndCacheStatus fetches fresh data and caches it.
func (m Model) fetchAndCacheStatus() tea.Cmd {
	return panelcache.FetchAndCache(m.fileCache, m.device, cache.TypeMatter, cache.TTLSmartHome, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		data := CachedSmartHomeData{
			Matter: m.fetchMatter(ctx),
			Zigbee: m.fetchZigbee(ctx),
			LoRa:   m.fetchLoRa(ctx),
		}

		return data, nil
	})
}

// backgroundRefresh refreshes data in the background without blocking.
func (m Model) backgroundRefresh() tea.Cmd {
	return panelcache.BackgroundRefresh(m.fileCache, m.device, cache.TypeMatter, cache.TTLSmartHome, func() (any, error) {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		data := CachedSmartHomeData{
			Matter: m.fetchMatter(ctx),
			Zigbee: m.fetchZigbee(ctx),
			LoRa:   m.fetchLoRa(ctx),
		}

		return data, nil
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
		return m.handleMatterEditModalUpdate(msg)
	}
	if m.zigbeeEditing {
		return m.handleZigbeeEditModalUpdate(msg)
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
	case MatterToggleResultMsg:
		return m.handleMatterToggleResult(msg)
	case MatterResetResultMsg:
		return m.handleMatterResetResult(msg)
	case ZigbeeToggleResultMsg:
		return m.handleZigbeeToggleResult(msg)
	case ZigbeeSteeringResultMsg:
		return m.handleZigbeeSteeringResult(msg)
	case ZigbeeLeaveResultMsg:
		return m.handleZigbeeLeaveResult(msg)
	default:
		if m.focused {
			return m.handleFocusedAction(msg)
		}
	}
	return m, nil
}

func (m Model) handleFocusedAction(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.RefreshRequestMsg:
		return m.handleRefresh()
	case messages.EditRequestMsg:
		return m.handleEditKey()
	case messages.ResetRequestMsg:
		return m.handleMatterResetKey()
	case messages.ViewRequestMsg:
		return m.handleEditKey() // 'c' for codes also opens edit modal
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		m = m.prevProtocol()
	case messages.NavDown:
		m = m.nextProtocol()
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m Model) handleRefresh() (Model, tea.Cmd) {
	if !m.loading && m.device != "" {
		m.loading = true
		// Invalidate cache and fetch fresh data
		return m, tea.Batch(
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
			m.fetchAndCacheStatus(),
		)
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
	if msg.Device != m.device || msg.DataType != cache.TypeMatter {
		return m, nil
	}

	data, err := panelcache.Unmarshal[CachedSmartHomeData](msg.Data)
	if err == nil {
		m.matter = data.Matter
		m.zigbee = data.Zigbee
		m.lora = data.LoRa
	}
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)

	// Emit StatusLoadedMsg with cached data so sequential loading can advance
	loadedCmd := func() tea.Msg { return StatusLoadedMsg{Matter: m.matter, Zigbee: m.zigbee, LoRa: m.lora} }

	if msg.NeedsRefresh {
		m.cacheStatus, _ = m.cacheStatus.StartRefresh()
		return m, tea.Batch(loadedCmd, m.cacheStatus.Tick(), m.backgroundRefresh())
	}
	return m, loadedCmd
}

func (m Model) handleCacheMiss(msg panelcache.CacheMissMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeMatter {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchAndCacheStatus())
}

func (m Model) handleRefreshComplete(msg panelcache.RefreshCompleteMsg) (Model, tea.Cmd) {
	if msg.Device != m.device || msg.DataType != cache.TypeMatter {
		return m, nil
	}
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		iostreams.DebugErr("smarthome background refresh", msg.Err)
		m.err = msg.Err
		// Emit StatusLoadedMsg with error so sequential loading can advance
		return m, func() tea.Msg { return StatusLoadedMsg{Err: msg.Err} }
	}
	if data, ok := msg.Data.(CachedSmartHomeData); ok {
		m.matter = data.Matter
		m.zigbee = data.Zigbee
		m.lora = data.LoRa
	}
	// Emit StatusLoadedMsg so sequential loading can advance
	return m, func() tea.Msg { return StatusLoadedMsg{Matter: m.matter, Zigbee: m.zigbee, LoRa: m.lora} }
}

func (m Model) handleStatusLoaded(msg StatusLoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	m.cacheStatus = m.cacheStatus.StopRefresh()
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.matter = msg.Matter
	m.zigbee = msg.Zigbee
	m.lora = msg.LoRa
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not covered by action messages
	switch msg.String() {
	case "1", "2", "3":
		return m.handleProtocolSelect(msg.String()), nil
	case "t":
		return m.handleProtocolToggle()
	case "c":
		if m.activeProtocol == ProtocolMatter {
			return m.handleMatterEditKey()
		}
	case "p":
		if m.activeProtocol == ProtocolZigbee {
			return m.handleZigbeePairKey()
		}
	case "R":
		return m.handleProtocolDestructive()
	case keyconst.KeyEsc, keyconst.KeyCtrlOpenBracket:
		if m.pendingReset || m.pendingLeave {
			m.pendingReset = false
			m.pendingLeave = false
			return m, nil
		}
	}

	return m, nil
}

func (m Model) handleProtocolSelect(key string) Model {
	m.pendingReset = false
	m.pendingLeave = false
	switch key {
	case "1":
		m.activeProtocol = ProtocolMatter
	case "2":
		m.activeProtocol = ProtocolZigbee
	case "3":
		m.activeProtocol = ProtocolLoRa
	}
	return m
}

func (m Model) handleProtocolToggle() (Model, tea.Cmd) {
	switch m.activeProtocol {
	case ProtocolMatter:
		return m.handleMatterToggle()
	case ProtocolZigbee:
		return m.handleZigbeeToggle()
	case ProtocolLoRa:
		return m, nil
	}
	return m, nil
}

func (m Model) handleProtocolDestructive() (Model, tea.Cmd) {
	switch m.activeProtocol {
	case ProtocolMatter:
		return m.handleMatterResetKey()
	case ProtocolZigbee:
		return m.handleZigbeeLeaveKey()
	case ProtocolLoRa:
		return m, nil
	}
	return m, nil
}

func (m Model) nextProtocol() Model {
	switch m.activeProtocol {
	case ProtocolMatter:
		m.activeProtocol = ProtocolZigbee
	case ProtocolZigbee:
		m.activeProtocol = ProtocolLoRa
	case ProtocolLoRa:
		m.activeProtocol = ProtocolMatter
	}
	return m
}

func (m Model) prevProtocol() Model {
	switch m.activeProtocol {
	case ProtocolMatter:
		m.activeProtocol = ProtocolLoRa
	case ProtocolZigbee:
		m.activeProtocol = ProtocolMatter
	case ProtocolLoRa:
		m.activeProtocol = ProtocolZigbee
	}
	return m
}

func (m Model) handleMatterEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
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
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
			m.fetchAndCacheStatus(),
		)
	}

	// Handle save result message - close modal and refresh
	if saveMsg, ok := msg.(MatterEditSaveResultMsg); ok {
		if saveMsg.Err == nil {
			m.editing = false
			m.editModal = m.editModal.Hide()
			m.loading = true
			return m, tea.Batch(
				m.Loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	// Handle reset result in modal
	if resetMsg, ok := msg.(MatterResetResultMsg); ok {
		if resetMsg.Err == nil {
			m.editing = false
			m.editModal = m.editModal.Hide()
			m.loading = true
			return m, tea.Batch(
				m.Loader.Tick(),
				panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
				m.fetchAndCacheStatus(),
				func() tea.Msg { return EditClosedMsg{Saved: true} },
			)
		}
	}

	return m, cmd
}

func (m Model) handleZigbeeEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.zigbeeModal, cmd = m.zigbeeModal.Update(msg)

	// Check if modal was closed
	if !m.zigbeeModal.Visible() {
		m.zigbeeEditing = false
		// Invalidate cache and refresh data after edit
		m.loading = true
		return m, tea.Batch(
			cmd,
			m.Loader.Tick(),
			panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
			m.fetchAndCacheStatus(),
		)
	}

	return m, cmd
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	switch m.activeProtocol {
	case ProtocolMatter:
		return m.handleMatterEditKey()
	case ProtocolZigbee:
		return m.handleZigbeeEditKey()
	default:
		return m, nil
	}
}

func (m Model) handleMatterEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.matter == nil || m.activeProtocol != ProtocolMatter {
		return m, nil
	}
	m.editing = true
	m.pendingReset = false
	m.editModal = m.editModal.SetSize(m.Width, m.Height)
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Show(m.device, m.matter)
	return m, tea.Batch(cmd, func() tea.Msg { return EditOpenedMsg{} })
}

func (m Model) handleMatterToggle() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.toggling || m.matter == nil {
		return m, nil
	}
	m.toggling = true
	m.pendingReset = false
	m.err = nil

	newEnabled := !m.matter.Enabled
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var err error
		if newEnabled {
			err = m.svc.Wireless().MatterEnable(ctx, m.device)
		} else {
			err = m.svc.Wireless().MatterDisable(ctx, m.device)
		}
		return MatterToggleResultMsg{Enabled: newEnabled, Err: err}
	}
}

func (m Model) handleMatterToggleResult(msg MatterToggleResultMsg) (Model, tea.Cmd) {
	m.toggling = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh to get fresh status
	m.loading = true
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleMatterResetKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.matter == nil || !m.matter.Enabled || m.activeProtocol != ProtocolMatter {
		return m, nil
	}

	// Double-press confirmation pattern
	if m.pendingReset {
		// Second press - execute reset
		m.pendingReset = false
		return m.executeMatterReset()
	}

	// First press - request confirmation
	m.pendingReset = true
	return m, nil
}

func (m Model) executeMatterReset() (Model, tea.Cmd) {
	m.toggling = true // Reuse toggling flag for busy state
	m.err = nil
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Wireless().MatterReset(ctx, m.device)
		return MatterResetResultMsg{Err: err}
	}
}

func (m Model) handleMatterResetResult(msg MatterResetResultMsg) (Model, tea.Cmd) {
	m.toggling = false
	m.pendingReset = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh to get fresh status
	m.loading = true
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
		m.fetchAndCacheStatus(),
	)
}

// --- Zigbee handlers ---

func (m Model) handleZigbeeEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.zigbee == nil || m.activeProtocol != ProtocolZigbee {
		return m, nil
	}
	m.zigbeeEditing = true
	m.pendingLeave = false
	m.zigbeeModal = m.zigbeeModal.SetSize(m.Width, m.Height)
	var cmd tea.Cmd
	m.zigbeeModal, cmd = m.zigbeeModal.Show(m.device, m.zigbee)
	return m, tea.Batch(cmd, func() tea.Msg { return EditOpenedMsg{} })
}

func (m Model) handleZigbeeToggle() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.toggling || m.zigbee == nil {
		return m, nil
	}
	m.toggling = true
	m.pendingLeave = false
	m.err = nil

	newEnabled := !m.zigbee.Enabled
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var err error
		if newEnabled {
			err = m.svc.Wireless().ZigbeeEnable(ctx, m.device)
		} else {
			err = m.svc.Wireless().ZigbeeDisable(ctx, m.device)
		}
		return ZigbeeToggleResultMsg{Enabled: newEnabled, Err: err}
	}
}

func (m Model) handleZigbeeToggleResult(msg ZigbeeToggleResultMsg) (Model, tea.Cmd) {
	m.toggling = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh to get fresh status
	m.loading = true
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleZigbeePairKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.toggling || m.zigbee == nil || !m.zigbee.Enabled {
		return m, nil
	}
	m.toggling = true // Reuse toggling flag for busy state
	m.err = nil
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Wireless().ZigbeeStartNetworkSteering(ctx, m.device)
		return ZigbeeSteeringResultMsg{Err: err}
	}
}

func (m Model) handleZigbeeSteeringResult(msg ZigbeeSteeringResultMsg) (Model, tea.Cmd) {
	m.toggling = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh to get fresh status
	m.loading = true
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
		m.fetchAndCacheStatus(),
	)
}

func (m Model) handleZigbeeLeaveKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || m.zigbee == nil || !m.zigbee.Enabled {
		return m, nil
	}
	if m.zigbee.NetworkState != zigbeeStateJoined {
		return m, nil
	}

	// Double-press confirmation pattern
	if m.pendingLeave {
		// Second press - execute leave
		m.pendingLeave = false
		return m.executeZigbeeLeave()
	}

	// First press - request confirmation
	m.pendingLeave = true
	return m, nil
}

func (m Model) executeZigbeeLeave() (Model, tea.Cmd) {
	m.toggling = true // Reuse toggling flag for busy state
	m.err = nil
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Wireless().ZigbeeLeaveNetwork(ctx, m.device)
		return ZigbeeLeaveResultMsg{Err: err}
	}
}

func (m Model) handleZigbeeLeaveResult(msg ZigbeeLeaveResultMsg) (Model, tea.Cmd) {
	m.toggling = false
	m.pendingLeave = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Invalidate cache and refresh to get fresh status
	m.loading = true
	return m, tea.Batch(
		m.Loader.Tick(),
		panelcache.Invalidate(m.fileCache, m.device, cache.TypeMatter),
		m.fetchAndCacheStatus(),
	)
}

// IsEditing returns whether any edit modal is currently visible.
func (m Model) IsEditing() bool {
	return m.editing || m.zigbeeEditing
}

// RenderEditModal returns the active edit modal view for full-screen overlay rendering.
func (m Model) RenderEditModal() string {
	if m.editing {
		return m.editModal.View()
	}
	if m.zigbeeEditing {
		return m.zigbeeModal.View()
	}
	return ""
}

// SetEditModalSize sets the active edit modal dimensions.
// This should be called with screen-based dimensions when a modal is visible.
func (m Model) SetEditModalSize(width, height int) Model {
	if m.editing {
		m.editModal = m.editModal.SetSize(width, height)
	}
	if m.zigbeeEditing {
		m.zigbeeModal = m.zigbeeModal.SetSize(width, height)
	}
	return m
}

// View renders the SmartHome component.
func (m Model) View() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Smart Home").
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

	var content strings.Builder

	// Matter Section
	content.WriteString(m.renderMatter())
	content.WriteString("\n\n")

	// Zigbee Section
	content.WriteString(m.renderZigbee())
	content.WriteString("\n\n")

	// LoRa Section
	content.WriteString(m.renderLoRa())

	r.SetContent(content.String())

	// Footer with keybindings and cache status (shown when focused)
	if m.focused {
		footer := m.buildFooter()
		r.SetFooter(footer)
	}
	return r.Render()
}

func (m Model) buildFooter() string {
	// Show confirmation prompts
	if m.pendingReset {
		return m.styles.Enabled.Render("Press 'R' again to confirm factory reset, Esc to cancel")
	}
	if m.pendingLeave {
		return m.styles.Enabled.Render("Press 'R' again to confirm leaving network, Esc to cancel")
	}

	if m.toggling {
		return "Processing..."
	}

	var footer string
	switch {
	case m.activeProtocol == ProtocolMatter && m.matter != nil:
		if m.matter.Enabled {
			footer = "e:edit t:toggle c:codes R:reset r:refresh"
		} else {
			footer = "e:edit t:toggle r:refresh"
		}
	case m.activeProtocol == ProtocolZigbee && m.zigbee != nil:
		footer = m.buildZigbeeFooter()
	default:
		footer = "1-3:sel j/k:nav r:refresh"
	}

	if cs := m.cacheStatus.View(); cs != "" {
		footer += " | " + cs
	}
	return footer
}

func (m Model) buildZigbeeFooter() string {
	if !m.zigbee.Enabled {
		return "e:edit t:toggle r:refresh"
	}
	if m.zigbee.NetworkState == zigbeeStateJoined {
		return "e:edit t:toggle p:pair R:leave r:refresh"
	}
	return "e:edit t:toggle p:pair r:refresh"
}

func (m Model) renderMatter() string {
	var content strings.Builder

	// Section header
	header := "Matter"
	if m.activeProtocol == ProtocolMatter {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.matter == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported (Gen4+ only)"))
		return content.String()
	}

	// Enabled status
	if m.matter.Enabled {
		content.WriteString("    " + m.styles.Enabled.Render("● Enabled") + "\n")
	} else {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled") + "\n")
		return content.String()
	}

	// Commissionable status
	content.WriteString("    " + m.styles.Label.Render("Commission: "))
	if m.matter.Commissionable {
		content.WriteString(m.styles.Warning.Render("Ready to pair"))
	} else {
		content.WriteString(m.styles.Muted.Render("Already paired"))
	}
	content.WriteString("\n")

	// Fabrics count
	content.WriteString("    " + m.styles.Label.Render("Fabrics:    "))
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.matter.FabricsCount)))

	return content.String()
}

func (m Model) renderZigbee() string {
	var content strings.Builder

	// Section header
	header := "Zigbee"
	if m.activeProtocol == ProtocolZigbee {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.zigbee == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported"))
		return content.String()
	}

	// Enabled status
	if m.zigbee.Enabled {
		content.WriteString("    " + m.styles.Enabled.Render("● Enabled") + "\n")
	} else {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled") + "\n")
		return content.String()
	}

	// Network state
	content.WriteString("    " + m.styles.Label.Render("State:   "))
	switch m.zigbee.NetworkState {
	case zigbeeStateJoined:
		content.WriteString(m.styles.Enabled.Render("Joined"))
	case zigbeeStateSteering:
		content.WriteString(m.styles.Warning.Render("Searching..."))
	case zigbeeStateReady:
		content.WriteString(m.styles.Muted.Render("Ready"))
	default:
		content.WriteString(m.styles.Muted.Render(m.zigbee.NetworkState))
	}
	content.WriteString("\n")

	// Network details (when joined)
	if m.zigbee.NetworkState == zigbeeStateJoined {
		content.WriteString("    " + m.styles.Label.Render("Channel: "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.zigbee.Channel)))
		content.WriteString("\n")

		content.WriteString("    " + m.styles.Label.Render("PAN ID:  "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("0x%04X", m.zigbee.PANID)))

		if m.zigbee.EUI64 != "" {
			content.WriteString("\n")
			content.WriteString("    " + m.styles.Label.Render("EUI-64:  "))
			content.WriteString(m.styles.Value.Render(m.zigbee.EUI64))
		}

		if m.zigbee.CoordinatorEUI64 != "" {
			content.WriteString("\n")
			content.WriteString("    " + m.styles.Label.Render("Coord:   "))
			content.WriteString(m.styles.Value.Render(m.zigbee.CoordinatorEUI64))
		}
	}

	return content.String()
}

func (m Model) renderLoRa() string {
	var content strings.Builder

	// Section header
	header := "LoRa"
	if m.activeProtocol == ProtocolLoRa {
		header = "▶ " + header
		content.WriteString(m.styles.Active.Render(header))
	} else {
		header = "  " + header
		content.WriteString(m.styles.Section.Render(header))
	}
	content.WriteString("\n")

	if m.lora == nil {
		content.WriteString(m.styles.Muted.Render("    Not supported (add-on required)"))
		return content.String()
	}

	// Enabled status
	if m.lora.Enabled {
		content.WriteString("    " + m.styles.Enabled.Render("● Enabled") + "\n")
	} else {
		content.WriteString("    " + m.styles.Disabled.Render("○ Disabled") + "\n")
		return content.String()
	}

	// Frequency
	content.WriteString("    " + m.styles.Label.Render("Freq:   "))
	freqMHz := float64(m.lora.Frequency) / 1000000
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%.2f MHz", freqMHz)))
	content.WriteString("\n")

	// TX Power
	content.WriteString("    " + m.styles.Label.Render("Power:  "))
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d dBm", m.lora.TxPower)))

	// Last RSSI/SNR if available
	if m.lora.RSSI != 0 {
		content.WriteString("\n    " + m.styles.Label.Render("Last RSSI: "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d dBm", m.lora.RSSI)))
		content.WriteString(" | SNR: ")
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%.1f dB", m.lora.SNR)))
	}

	return content.String()
}

// Matter returns the current Matter status.
func (m Model) Matter() *shelly.TUIMatterStatus {
	return m.matter
}

// Zigbee returns the current Zigbee status.
func (m Model) Zigbee() *shelly.TUIZigbeeStatus {
	return m.zigbee
}

// LoRa returns the current LoRa status.
func (m Model) LoRa() *shelly.TUILoRaStatus {
	return m.lora
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

// ActiveProtocol returns the currently selected protocol.
func (m Model) ActiveProtocol() Protocol {
	return m.activeProtocol
}

// Refresh triggers a refresh of the smart home data.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchStatus())
}
