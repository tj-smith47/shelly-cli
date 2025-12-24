package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/branding"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cmdmode"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/confirm"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/deviceinfo"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/devicelist"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/energybars"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/energyhistory"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/events"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/help"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/jsonviewer"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/search"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/statusbar"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
	"github.com/tj-smith47/shelly-cli/internal/tui/views"
)

// DeviceActionMsg reports the result of a device action.
type DeviceActionMsg struct {
	Device string
	Action string
	Err    error
}

// PanelFocus tracks which panel is focused.
type PanelFocus int

const (
	// PanelDeviceList is the device list column.
	PanelDeviceList PanelFocus = iota
	// PanelDetail is the device info column.
	PanelDetail
	// PanelEndpoints means endpoint is selected, JSON overlay visible.
	PanelEndpoints
)

// LayoutMode represents the terminal layout mode based on width.
type LayoutMode int

const (
	// LayoutNarrow for terminals < 80 cols - vertical stack.
	LayoutNarrow LayoutMode = iota
	// LayoutStandard for 80-120 cols - standard 3-panel.
	LayoutStandard
	// LayoutWide for > 120 cols - extra detail space.
	LayoutWide
)

// Model is the main TUI application model.
type Model struct {
	// Core
	ctx     context.Context
	factory *cmdutil.Factory
	cfg     *config.Config
	keys    KeyMap
	styles  Styles

	// View management (5-tab system)
	viewManager *views.Manager
	tabBar      tabs.Model

	// Shared device cache
	cache *cache.Cache

	// State
	ready           bool
	quitting        bool
	cursor          int        // Selected device index
	componentCursor int        // Selected component within device (-1 = all)
	filter          string     // Device name filter
	focusedPanel    PanelFocus // Which panel is currently focused
	endpointCursor  int        // Selected endpoint in JSON panel

	// Dimensions
	width  int
	height int

	// Components
	statusBar     statusbar.Model
	search        search.Model
	cmdMode       cmdmode.Model
	help          help.Model
	toast         toast.Model
	events        events.Model
	energyBars    energybars.Model
	energyHistory energyhistory.Model
	jsonViewer    jsonviewer.Model
	confirm       confirm.Model
	deviceInfo    deviceinfo.Model
	deviceList    devicelist.Model

	// Settings
	refreshInterval time.Duration

	// Debug logging (enabled by SHELLY_TUI_DEBUG=1)
	debugLogger *debug.Logger
}

// Options configures the TUI.
type Options struct {
	RefreshInterval time.Duration
	Filter          string
}

// DefaultOptions returns default TUI options.
func DefaultOptions() Options {
	return Options{
		RefreshInterval: 5 * time.Second,
	}
}

// New creates a new TUI application.
func New(ctx context.Context, f *cmdutil.Factory, opts Options) Model {
	cfg, err := f.Config()
	if err != nil {
		cfg = nil
	}

	// Apply TUI-specific theme if configured
	if cfg != nil {
		if tc := cfg.GetTUIThemeConfig(); tc != nil {
			if err := theme.ApplyConfig(tc.Name, tc.Colors, tc.Semantic, tc.File); err != nil {
				f.IOStreams().DebugErr("tui theme", err)
			}
		}
	}

	// Create shared cache
	deviceCache := cache.New(ctx, f.ShellyService(), f.IOStreams(), opts.RefreshInterval)

	// Create search component with initial filter
	searchModel := search.NewWithFilter(opts.Filter)

	// Create events component for real-time event stream
	eventsModel := events.New(events.Deps{
		Ctx: ctx,
		Svc: f.ShellyService(),
		IOS: f.IOStreams(),
	})

	// Create energy bars component
	energyBarsModel := energybars.New(deviceCache)

	// Create energy history sparklines component
	energyHistoryModel := energyhistory.New(deviceCache)

	// Create JSON viewer component
	jsonViewerModel := jsonviewer.New(ctx, f.ShellyService())

	// Create device info component
	deviceInfoModel := deviceinfo.New()

	// Create device list component (uses shared cache)
	deviceListModel := devicelist.New(deviceCache)

	// Create view manager and register all views
	vm := views.New()

	// Register Dashboard view (delegates rendering to app.go)
	dashboard := views.NewDashboard(views.DashboardDeps{Ctx: ctx})
	vm.Register(dashboard)

	// Register Automation view
	automation := views.NewAutomation(views.AutomationDeps{
		Ctx: ctx,
		Svc: f.ShellyService(),
	})
	vm.Register(automation)

	// Register Config view
	configView := views.NewConfig(views.ConfigDeps{
		Ctx: ctx,
		Svc: f.ShellyService(),
	})
	vm.Register(configView)

	// Register Manage view
	manage := views.NewManage(views.ManageDeps{
		Ctx: ctx,
		Svc: f.ShellyService(),
	})
	vm.Register(manage)

	// Register Fleet view
	fleet := views.NewFleet(views.FleetDeps{
		Ctx: ctx,
		Svc: f.ShellyService(),
		IOS: f.IOStreams(),
		Cfg: cfg,
	})
	vm.Register(fleet)

	// Create 5-tab bar
	tabBar := tabs.New()

	// Load keybindings from config or use defaults
	keymap := KeyMapFromConfig(cfg)

	return Model{
		ctx:             ctx,
		factory:         f,
		cfg:             cfg,
		keys:            keymap,
		styles:          DefaultStyles(),
		viewManager:     vm,
		tabBar:          tabBar,
		cache:           deviceCache,
		refreshInterval: opts.RefreshInterval,
		filter:          opts.Filter,
		componentCursor: -1, // -1 means "all components"
		focusedPanel:    PanelDeviceList,
		statusBar:       statusbar.New(),
		search:          searchModel,
		cmdMode:         cmdmode.New(),
		help:            help.New(),
		toast:           toast.New(),
		events:          eventsModel,
		energyBars:      energyBarsModel,
		energyHistory:   energyHistoryModel,
		jsonViewer:      jsonViewerModel,
		confirm:         confirm.New(),
		deviceInfo:      deviceInfoModel,
		deviceList:      deviceListModel,
		debugLogger:     debug.New(), // nil if SHELLY_TUI_DEBUG not set
	}
}

// Close cleans up resources used by the TUI.
func (m Model) Close() {
	if err := m.debugLogger.Close(); err != nil {
		iostreams.DebugErr("close debug logger", err)
	}
}

// layoutMode returns the current layout mode based on terminal width.
func (m Model) layoutMode() LayoutMode {
	if m.width < 80 {
		return LayoutNarrow
	}
	if m.width > 120 {
		return LayoutWide
	}
	return LayoutStandard
}

// Init initializes the TUI and returns the first command to run.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.cache.Init(),
		m.statusBar.Init(),
		m.toast.Init(),
		m.events.Init(),
		m.viewManager.Init(),
	)
}

// Update handles messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle specific message types first
	if newModel, cmd, handled := m.handleSpecificMsg(msg); handled {
		return newModel, cmd
	}

	// Forward and update components
	newModel, cmds := m.updateComponents(msg)
	return newModel, tea.Batch(cmds...)
}

// handleSpecificMsg handles specific message types that return early.
func (m Model) handleSpecificMsg(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg), nil, true
	case tea.KeyPressMsg:
		return m.handleKeyPressMsg(msg)
	case DeviceActionMsg:
		newModel, cmd := m.handleDeviceAction(msg)
		return newModel, cmd, true
	case cache.DeviceUpdateMsg:
		// Forward to cache AND energyHistory (to record power history)
		cacheCmd := m.cache.Update(msg)
		var historyCmd tea.Cmd
		m.energyHistory, historyCmd = m.energyHistory.Update(msg)
		return m, tea.Batch(cacheCmd, historyCmd), true
	case cache.RefreshTickMsg, cache.WaveMsg, cache.DeviceRefreshMsg:
		return m, m.cache.Update(msg), true
	case cache.AllDevicesLoadedMsg:
		return m.handleAllDevicesLoaded(msg)
	case events.EventMsg, events.SubscriptionStatusMsg:
		var cmd tea.Cmd
		m.events, cmd = m.events.Update(msg)
		return m, cmd, true
	case search.FilterChangedMsg:
		m.filter = msg.Filter
		m.cursor = 0
		m.deviceList = m.deviceList.SetFilter(msg.Filter)
		return m, nil, true
	case search.ClosedMsg, cmdmode.ClosedMsg, confirm.CancelledMsg:
		return m, nil, true
	case cmdmode.CommandMsg:
		newModel, cmd := m.handleCommand(msg)
		return newModel, cmd, true
	case cmdmode.ErrorMsg:
		return m, statusbar.SetMessage(msg.Message, statusbar.MessageError), true
	default:
		return m.handleViewAndComponentMsgs(msg)
	}
}

// handleViewAndComponentMsgs handles view-related and component messages.
func (m Model) handleViewAndComponentMsgs(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tabs.TabChangedMsg:
		return m, m.viewManager.SetActive(msg.Current), true
	case views.ViewChangedMsg:
		m.tabBar, _ = m.tabBar.SetActive(msg.Current)
		return m, nil, true
	case views.DeviceSelectedMsg:
		return m, m.viewManager.PropagateDevice(msg.Device), true
	case devicelist.DeviceSelectedMsg:
		// Sync cursor from deviceList to app.go
		m.cursor = m.deviceList.Cursor()
		return m, m.viewManager.PropagateDevice(msg.Name), true
	case jsonviewer.CloseMsg:
		m.focusedPanel = PanelDetail
		return m, nil, true
	case jsonviewer.FetchedMsg:
		var cmd tea.Cmd
		m.jsonViewer, cmd = m.jsonViewer.Update(msg)
		return m, cmd, true
	case confirm.ConfirmedMsg:
		return m, toast.Success("Action confirmed: " + msg.Operation), true
	case deviceinfo.RequestJSONMsg:
		return m.handleRequestJSON(msg)
	}
	return m, nil, false
}

// handleAllDevicesLoaded handles the AllDevicesLoadedMsg and emits initial selection.
func (m Model) handleAllDevicesLoaded(msg cache.AllDevicesLoadedMsg) (tea.Model, tea.Cmd, bool) {
	cmd := m.cache.Update(msg)
	devices := m.getFilteredDevices()
	if len(devices) == 0 || m.cursor >= len(devices) {
		return m, cmd, true
	}
	d := devices[m.cursor]
	return m, tea.Batch(cmd, func() tea.Msg {
		return views.DeviceSelectedMsg{
			Device:  d.Device.Name,
			Address: d.Device.Address,
		}
	}), true
}

// handleRequestJSON opens the JSON viewer for a requested endpoint.
func (m Model) handleRequestJSON(msg deviceinfo.RequestJSONMsg) (tea.Model, tea.Cmd, bool) {
	devices := m.getFilteredDevices()
	for _, d := range devices {
		if d.Device.Name != msg.DeviceName {
			continue
		}
		m.focusedPanel = PanelEndpoints
		endpoints := m.getDeviceMethods(d)
		var cmd tea.Cmd
		m.jsonViewer, cmd = m.jsonViewer.Open(d.Device.Address, msg.Endpoint, endpoints)
		return m, cmd, true
	}
	return m, nil, true
}

// updateComponents forwards messages to components and collects commands.
func (m Model) updateComponents(msg tea.Msg) (Model, []tea.Cmd) {
	var cmds []tea.Cmd

	// Update status bar, toast, and events
	var statusCmd, toastCmd, eventsCmd tea.Cmd
	m.statusBar, statusCmd = m.statusBar.Update(msg)
	m.toast, toastCmd = m.toast.Update(msg)
	m.events, eventsCmd = m.events.Update(msg)
	cmds = append(cmds, statusCmd, toastCmd, eventsCmd)

	// Forward non-key messages to ALL views so async results can be processed
	// (e.g., Config's wifi.StatusLoadedMsg needs to reach Config even if Dashboard is active)
	if _, isKey := msg.(tea.KeyPressMsg); !isKey {
		viewCmd := m.viewManager.UpdateAll(msg)
		cmds = append(cmds, viewCmd)
	}

	// Update deviceInfo with current device selection
	m = m.syncDeviceInfo()

	return m, cmds
}

// handleKeyPressMsg handles keyboard input.
func (m Model) handleKeyPressMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	// If help is visible, close on dismiss keys
	if m.help.Visible() {
		if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Escape) || key.Matches(msg, m.keys.Quit) {
			m.help = m.help.Hide()
			return m, nil, true
		}
		var cmd tea.Cmd
		m.help, cmd = m.help.Update(msg)
		return m, cmd, true
	}

	// If command mode is active, forward all keys to it
	if m.cmdMode.IsActive() {
		var cmd tea.Cmd
		m.cmdMode, cmd = m.cmdMode.Update(msg)
		return m, cmd, true
	}

	// If search is active, forward all keys to it
	if m.search.IsActive() {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd, true
	}

	// If JSON viewer is visible, forward all keys to it
	if m.jsonViewer.Visible() {
		var cmd tea.Cmd
		m.jsonViewer, cmd = m.jsonViewer.Update(msg)
		return m, cmd, true
	}

	// If confirm dialog is visible, forward all keys to it
	if m.confirm.Visible() {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd, true
	}

	return m.handleKeyPress(msg)
}

// handleWindowSize handles window resize events.
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height
	m.ready = true

	m.statusBar = m.statusBar.SetWidth(m.width)
	m.search = m.search.SetWidth(m.width)
	m.cmdMode = m.cmdMode.SetWidth(m.width)
	m.toast = m.toast.SetSize(m.width, m.height)
	m.help = m.help.SetSize(m.width, m.height)
	m.tabBar = m.tabBar.SetWidth(m.width)

	// Calculate content height (banner + tab bar + status bar)
	tabBarHeight := 1
	contentHeight := m.height - branding.BannerHeight() - tabBarHeight - 2
	m.viewManager.SetSize(m.width, contentHeight)

	// Calculate panel sizes for events
	panelHeight := (m.height - branding.BannerHeight() - 3) / 2
	panelWidth := m.width / 2
	m.events = m.events.SetSize(panelWidth, panelHeight)

	// Set JSON viewer size (centered overlay)
	jsonWidth := m.width * 2 / 3
	jsonHeight := m.height * 2 / 3
	if jsonWidth > 100 {
		jsonWidth = 100
	}
	m.jsonViewer = m.jsonViewer.SetSize(jsonWidth, jsonHeight)

	// Set confirm dialog size
	m.confirm = m.confirm.SetSize(m.width, m.height)

	// Set deviceList size (events column + device list/detail columns)
	deviceListWidth := m.width - m.width*25/100 - 2 // Full width minus events column
	m.deviceList = m.deviceList.SetSize(deviceListWidth, m.height-branding.BannerHeight()-5)

	return m
}

// handleDeviceAction handles device action results.
func (m Model) handleDeviceAction(msg DeviceActionMsg) (tea.Model, tea.Cmd) {
	var statusCmd tea.Cmd
	if msg.Err != nil {
		statusCmd = statusbar.SetMessage(
			msg.Device+": "+msg.Action+" failed - "+msg.Err.Error(),
			statusbar.MessageError,
		)
	} else {
		statusCmd = statusbar.SetMessage(
			msg.Device+": "+msg.Action+" success",
			statusbar.MessageSuccess,
		)
	}
	return m, statusCmd
}

// handleCommand handles command mode commands.
func (m Model) handleCommand(msg cmdmode.CommandMsg) (tea.Model, tea.Cmd) {
	switch msg.Command {
	case cmdmode.CmdQuit:
		m.quitting = true
		return m, tea.Quit

	case cmdmode.CmdDevice, cmdmode.CmdFilter:
		m.filter = msg.Args
		m.cursor = 0
		if msg.Args == "" {
			return m, statusbar.SetMessage("Filter cleared", statusbar.MessageSuccess)
		}
		return m, statusbar.SetMessage("Filter: "+msg.Args, statusbar.MessageSuccess)

	case cmdmode.CmdTheme:
		if !theme.SetTheme(msg.Args) {
			return m, statusbar.SetMessage("Invalid theme: "+msg.Args, statusbar.MessageError)
		}
		m.styles = DefaultStyles()
		return m, statusbar.SetMessage("Theme: "+msg.Args, statusbar.MessageSuccess)

	case cmdmode.CmdView:
		// Views are collapsed - just acknowledge command
		return m, statusbar.SetMessage("Single unified view", statusbar.MessageSuccess)

	case cmdmode.CmdHelp:
		m.help = m.help.SetSize(m.width, m.height)
		m.help = m.help.SetContext(m.getHelpContext())
		m.help = m.help.Toggle()
		return m, nil

	case cmdmode.CmdToggle:
		if cmd := m.executeDeviceAction("toggle"); cmd != nil {
			return m, cmd
		}
		return m, statusbar.SetMessage("No device selected or device offline", statusbar.MessageError)

	default:
		return m, statusbar.SetMessage("Unknown command", statusbar.MessageError)
	}
}

// getHelpContext returns the help context based on the current active tab.
func (m Model) getHelpContext() keys.Context {
	switch m.tabBar.ActiveTabID() {
	case tabs.TabDashboard:
		return keys.ContextDevices
	case tabs.TabAutomation:
		return keys.ContextAutomation
	case tabs.TabConfig:
		return keys.ContextConfig
	case tabs.TabManage:
		return keys.ContextManage
	case tabs.TabFleet:
		return keys.ContextFleet
	default:
		return keys.ContextDevices
	}
}

// getFilteredDevices returns devices matching the current filter.
func (m Model) getFilteredDevices() []*cache.DeviceData {
	all := m.cache.GetAllDevices()
	if m.filter == "" {
		return all
	}

	filterLower := strings.ToLower(m.filter)
	filtered := make([]*cache.DeviceData, 0, len(all))
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Device.Name), filterLower) ||
			strings.Contains(strings.ToLower(d.Device.Type), filterLower) ||
			strings.Contains(d.Device.Address, filterLower) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// handleKeyPress handles global key presses.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	// Global actions (view switching, quit, help, etc.)
	if newModel, cmd, handled := m.handleGlobalKeys(msg); handled {
		return newModel, cmd, true
	}

	// If NOT on Dashboard, forward keys to the active view
	if !m.isDashboardActive() {
		cmd := m.viewManager.Update(msg)
		return m, cmd, true
	}

	// Dashboard-specific handling below

	// Panel switching with Tab/Shift+Tab
	if newModel, cmd, handled := m.handlePanelSwitch(msg); handled {
		return newModel, cmd, true
	}

	// Navigation (based on focused panel)
	if newModel, cmd, handled := m.handleNavigation(msg); handled {
		return newModel, cmd, true
	}

	// Device actions
	if newModel, cmd, handled := m.handleDeviceActionKey(msg); handled {
		return newModel, cmd, true
	}

	return m, nil, false
}

// handlePanelSwitch handles Tab/Shift+Tab for switching panels and Enter for JSON overlay.
func (m Model) handlePanelSwitch(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		m.focusedPanel = m.nextPanel()
		return m, nil, true
	case "shift+tab":
		m.focusedPanel = m.prevPanel()
		return m, nil, true
	case "enter":
		return m.openJSONViewer()
	case "escape":
		if m.jsonViewer.Visible() {
			m.jsonViewer = m.jsonViewer.Close()
			m.focusedPanel = PanelDetail
			return m, nil, true
		}
	}
	return m, nil, false
}

// nextPanel returns the next panel in the cycle: DeviceList -> Detail -> DeviceList.
func (m Model) nextPanel() PanelFocus {
	if m.focusedPanel == PanelDeviceList {
		return PanelDetail
	}
	return PanelDeviceList
}

// prevPanel returns the previous panel in the cycle: Detail -> DeviceList -> Detail.
func (m Model) prevPanel() PanelFocus {
	if m.focusedPanel == PanelDetail {
		return PanelDeviceList
	}
	return PanelDetail
}

// openJSONViewer opens the JSON viewer for the selected device if on Detail panel.
func (m Model) openJSONViewer() (Model, tea.Cmd, bool) {
	if m.focusedPanel != PanelDetail {
		return m, nil, false
	}

	devices := m.getFilteredDevices()
	if len(devices) == 0 || m.cursor >= len(devices) {
		return m, nil, false
	}

	d := devices[m.cursor]
	if !d.Online {
		return m, nil, false
	}

	m.focusedPanel = PanelEndpoints
	methods := m.getDeviceMethods(d)
	endpoint := m.selectEndpoint(methods)

	var cmd tea.Cmd
	m.jsonViewer, cmd = m.jsonViewer.Open(d.Device.Address, endpoint, methods)
	return m, cmd, true
}

// selectEndpoint returns the appropriate endpoint based on cursor position.
func (m Model) selectEndpoint(methods []string) string {
	if m.endpointCursor < len(methods) {
		return methods[m.endpointCursor]
	}
	if len(methods) > 0 {
		return methods[0]
	}
	return ""
}

// handleGlobalKeys handles quit, help, filter, command, tab switching, and escape keys.
func (m Model) handleGlobalKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.ForceQuit), key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit, true
	case key.Matches(msg, m.keys.Help):
		m.help = m.help.SetSize(m.width, m.height)
		m.help = m.help.SetContext(m.getHelpContext())
		m.help = m.help.Toggle()
		return m, nil, true
	case key.Matches(msg, m.keys.Filter):
		var cmd tea.Cmd
		m.search, cmd = m.search.Activate()
		return m, cmd, true
	case key.Matches(msg, m.keys.Command):
		var cmd tea.Cmd
		m.cmdMode, cmd = m.cmdMode.Activate()
		return m, cmd, true
	case key.Matches(msg, m.keys.Escape):
		if m.filter != "" {
			m.filter = ""
			m.cursor = 0
			return m, nil, true
		}
	case key.Matches(msg, m.keys.View1):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabDashboard)
		return m, m.viewManager.SetActive(views.ViewDashboard), true
	case key.Matches(msg, m.keys.View2):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabAutomation)
		return m, m.viewManager.SetActive(views.ViewAutomation), true
	case key.Matches(msg, m.keys.View3):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabConfig)
		return m, m.viewManager.SetActive(views.ViewConfig), true
	case key.Matches(msg, m.keys.View4):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabManage)
		return m, m.viewManager.SetActive(views.ViewManage), true
	case key.Matches(msg, m.keys.View5):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabFleet)
		return m, m.viewManager.SetActive(views.ViewFleet), true
	}
	return m, nil, false
}

// navDirection represents a navigation direction for cursor movement.
type navDirection int

const (
	navDown navDirection = iota
	navUp
	navFirst
	navLast
)

// parseNavDirection converts a key string to a navigation direction.
func parseNavDirection(keyStr string) (navDirection, bool) {
	switch keyStr {
	case "j", "down":
		return navDown, true
	case "k", "up":
		return navUp, true
	case "g":
		return navFirst, true
	case "G":
		return navLast, true
	default:
		return 0, false
	}
}

// handleNavigation handles j/k/g/G/h/l navigation keys based on focused panel.
func (m Model) handleNavigation(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	// When PanelDeviceList is focused, forward navigation to deviceList component
	if m.focusedPanel == PanelDeviceList {
		var cmd tea.Cmd
		m.deviceList, cmd = m.deviceList.Update(msg)
		// Sync cursor from deviceList
		m.cursor = m.deviceList.Cursor()
		return m, cmd, true
	}

	// When PanelDetail is focused, forward navigation to deviceInfo component
	if m.focusedPanel == PanelDetail {
		var cmd tea.Cmd
		m.deviceInfo, cmd = m.deviceInfo.Update(msg)
		return m, cmd, true
	}

	// Handle endpoints panel navigation
	devices := m.getFilteredDevices()
	keyStr := msg.String()

	if m.focusedPanel == PanelEndpoints {
		if dir, ok := parseNavDirection(keyStr); ok {
			m = m.navEndpoints(dir, devices)
			return m, nil, true
		}
	}

	return m, nil, false
}

// navEndpoints handles vertical navigation within the endpoints panel.
func (m Model) navEndpoints(dir navDirection, devices []*cache.DeviceData) Model {
	deviceCount := len(devices)
	switch dir {
	case navDown:
		if m.cursor >= 0 && m.cursor < deviceCount {
			methods := m.getDeviceMethods(devices[m.cursor])
			if m.endpointCursor < len(methods)-1 {
				m.endpointCursor++
			}
		}
	case navUp:
		if m.endpointCursor > 0 {
			m.endpointCursor--
		}
	case navFirst:
		m.endpointCursor = 0
	case navLast:
		if m.cursor >= 0 && m.cursor < deviceCount {
			methods := m.getDeviceMethods(devices[m.cursor])
			if len(methods) > 0 {
				m.endpointCursor = len(methods) - 1
			}
		}
	}
	return m
}

// syncDeviceInfo syncs the deviceInfo component with current selection and focus.
func (m Model) syncDeviceInfo() Model {
	devices := m.getFilteredDevices()

	// Update selected device
	if len(devices) > 0 && m.cursor < len(devices) {
		m.deviceInfo = m.deviceInfo.SetDevice(devices[m.cursor])
	} else {
		m.deviceInfo = m.deviceInfo.SetDevice(nil)
	}

	// Update focus state
	m.deviceInfo = m.deviceInfo.SetFocused(m.focusedPanel == PanelDetail)

	return m
}

// handleDeviceActionKey handles device action key presses.
func (m Model) handleDeviceActionKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	var action string
	switch {
	case key.Matches(msg, m.keys.Toggle):
		action = "toggle"
	case key.Matches(msg, m.keys.TurnOn):
		action = "on"
	case key.Matches(msg, m.keys.TurnOff):
		action = "off"
	case key.Matches(msg, m.keys.Reboot):
		action = "reboot"
	default:
		return m, nil, false
	}

	if cmd := m.executeDeviceAction(action); cmd != nil {
		return m, cmd, true
	}
	return m, nil, false
}

// executeDeviceAction executes a device action on the selected device.
func (m Model) executeDeviceAction(action string) tea.Cmd {
	devices := m.getFilteredDevices()
	if m.cursor >= len(devices) {
		return nil
	}

	selected := devices[m.cursor]
	if !selected.Online {
		return nil
	}

	device := selected.Device
	svc := m.factory.ShellyService()
	componentID := m.componentCursor // -1 means all, >=0 means specific component

	return func() tea.Msg {
		var err error
		var compIDPtr *int
		if componentID >= 0 {
			compIDPtr = &componentID
		}

		switch action {
		case "toggle":
			// QuickToggle handles all controllable components (switch, light, rgb, cover)
			_, err = svc.QuickToggle(m.ctx, device.Address, compIDPtr)
		case "on":
			_, err = svc.QuickOn(m.ctx, device.Address, compIDPtr)
		case "off":
			_, err = svc.QuickOff(m.ctx, device.Address, compIDPtr)
		case "reboot":
			err = svc.DeviceReboot(m.ctx, device.Address, 0)
		}

		// Build action description
		actionDesc := action
		if componentID >= 0 {
			actionDesc = fmt.Sprintf("%s (component %d)", action, componentID)
		}

		return DeviceActionMsg{
			Device: device.Name,
			Action: actionDesc,
			Err:    err,
		}
	}
}

// View renders the TUI.
func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	if !m.ready {
		v := tea.NewView("Initializing...")
		v.AltScreen = true
		return v
	}

	// Header with metadata on left, banner on right
	headerBanner := m.renderHeader()

	// Tab bar
	tabBarView := m.tabBar.View()

	// Input bar (search or command mode)
	var inputBar string
	if m.search.IsActive() {
		inputBar = m.search.View()
	} else if m.cmdMode.IsActive() {
		inputBar = m.cmdMode.View()
	}

	// Calculate content area
	bannerHeight := branding.BannerHeight()
	tabBarHeight := 1
	footerHeight := 2
	inputHeight := 0
	if inputBar != "" {
		inputHeight = 3 // Input bar has top border, content, bottom border
	}
	contentHeight := m.height - bannerHeight - tabBarHeight - footerHeight - inputHeight

	// Render content based on active view
	var content string
	if m.isDashboardActive() {
		// Dashboard uses the multi-panel layout rendered by app.go
		content = m.renderMultiPanelLayout(contentHeight)
	} else {
		// Other views render themselves
		content = m.viewManager.View()
	}

	// Help overlay - overlay on content when visible
	if m.help.Visible() {
		content = m.renderWithHelpOverlay(content, contentHeight)
	}

	// Footer (status bar)
	footer := m.statusBar.View()

	// Compose the layout
	var result string
	if inputBar != "" {
		result = lipgloss.JoinVertical(lipgloss.Left,
			headerBanner, tabBarView, inputBar, content, footer,
		)
	} else {
		result = lipgloss.JoinVertical(lipgloss.Left,
			headerBanner, tabBarView, content, footer,
		)
	}

	// Add toast overlay (now a no-op)
	if m.toast.HasToasts() {
		result = m.toast.Overlay(result)
	}

	// Add confirm dialog overlay
	if m.confirm.Visible() {
		confirmView := m.confirm.View()
		result = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			confirmView,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	// Debug logging (if enabled)
	m.debugLogger.Log(m.tabBar.ActiveTabID().String(), m.focusedPanelName(), m.width, m.height, result)

	v := tea.NewView(result)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// focusedPanelName returns a human-readable name for the currently focused panel.
func (m Model) focusedPanelName() string {
	switch m.focusedPanel {
	case PanelDeviceList:
		return "DeviceList"
	case PanelDetail:
		return "Detail"
	case PanelEndpoints:
		return "Endpoints"
	default:
		return "Unknown"
	}
}

// isDashboardActive returns true if the Dashboard view is active.
func (m Model) isDashboardActive() bool {
	activeView := m.viewManager.ActiveView()
	if activeView == nil {
		return true // Default to dashboard
	}
	// Check if active view is a Dashboard view (which has empty View())
	if d, ok := activeView.(*views.Dashboard); ok {
		return d.IsDashboardView()
	}
	return false
}

// renderMultiPanelLayout renders panels horizontally or vertically based on width.
func (m Model) renderMultiPanelLayout(height int) string {
	// Narrow mode: stack panels vertically
	if m.layoutMode() == LayoutNarrow {
		return m.renderNarrowLayout(height)
	}

	// Split height: top 75% for panels, bottom 25% for energy bars
	topHeight := height * 75 / 100
	energyHeight := height - topHeight - 1 // -1 for gap

	if topHeight < 10 {
		topHeight = 10
	}
	if energyHeight < 5 {
		energyHeight = 5
	}

	// Calculate column widths: events column + deviceList (which has internal 40/60 split)
	var eventsWidth, deviceListWidth int
	if m.layoutMode() == LayoutWide {
		eventsWidth = m.width * 25 / 100
	} else {
		eventsWidth = m.width * 30 / 100
	}
	deviceListWidth = m.width - eventsWidth - 1 // -1 for gap

	if eventsWidth < 15 {
		eventsWidth = 15
	}

	// Render events column
	eventsCol := m.renderEventsColumn(eventsWidth, topHeight)

	// Render device list component (includes list + detail in split pane)
	m.deviceList = m.deviceList.SetSize(deviceListWidth, topHeight)
	m.deviceList = m.deviceList.SetFocused(m.focusedPanel == PanelDeviceList)
	deviceListCol := m.deviceList.View()

	topContent := lipgloss.JoinHorizontal(lipgloss.Top, eventsCol, " ", deviceListCol)

	// Ensure topContent fills the full width
	topContentStyle := lipgloss.NewStyle().Width(m.width)
	topContent = topContentStyle.Render(topContent)

	// Energy bars panel at bottom (full width)
	energyPanel := m.renderEnergyPanel(m.width, energyHeight)

	// Combine top and bottom sections
	content := lipgloss.JoinVertical(lipgloss.Left, topContent, energyPanel)

	// JSON viewer overlay (centered on top of content when active)
	if m.jsonViewer.Visible() {
		jsonOverlay := m.jsonViewer.View()
		return lipgloss.Place(
			m.width,
			height,
			lipgloss.Center,
			lipgloss.Center,
			jsonOverlay,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return content
}

// renderEnergyPanel renders the energy consumption bars with sparkline history.
func (m Model) renderEnergyPanel(width, height int) string {
	// Split width between bars and history
	barsWidth := width * 60 / 100
	historyWidth := width - barsWidth - 1

	m.energyBars = m.energyBars.SetSize(barsWidth, height)
	m.energyHistory = m.energyHistory.SetSize(historyWidth, height)

	barsView := m.energyBars.View()
	historyView := m.energyHistory.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, barsView, " ", historyView)
}

// renderNarrowLayout renders panels stacked vertically for narrow terminals.
func (m Model) renderNarrowLayout(height int) string {
	panelWidth := m.width - 2 // Full width minus borders

	// Divide height into 2 parts: events (25%), deviceList (75%)
	eventsHeight := height * 25 / 100
	deviceListHeight := height - eventsHeight - 1 // -1 for gap

	// Minimum heights
	if eventsHeight < 4 {
		eventsHeight = 4
	}
	if deviceListHeight < 10 {
		deviceListHeight = 10
	}

	// Row 1: Events panel (compact)
	eventsRow := m.renderEventsColumn(panelWidth, eventsHeight)

	// Row 2: Device List (with internal split pane)
	m.deviceList = m.deviceList.SetSize(panelWidth, deviceListHeight)
	m.deviceList = m.deviceList.SetFocused(m.focusedPanel == PanelDeviceList)
	deviceListRow := m.deviceList.View()

	// Stack panels vertically
	content := lipgloss.JoinVertical(lipgloss.Left, eventsRow, deviceListRow)

	// JSON viewer overlay (centered on top of content when active)
	if m.jsonViewer.Visible() {
		jsonOverlay := m.jsonViewer.View()
		return lipgloss.Place(
			m.width,
			height,
			lipgloss.Center,
			lipgloss.Center,
			jsonOverlay,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return content
}

// renderEventsColumn renders the events column with embedded title.
func (m Model) renderEventsColumn(width, height int) string {
	colors := theme.GetSemanticColors()
	orangeBorder := theme.Orange()

	eventCount := m.events.EventCount()
	title := fmt.Sprintf("Events (%d)", eventCount)

	r := rendering.New(width, height).
		SetTitle(title).
		SetFocused(false).
		SetFocusColor(orangeBorder).
		SetBlurColor(orangeBorder)

	eventsView := m.events.View()
	if eventsView == "" {
		eventsView = lipgloss.NewStyle().Foreground(colors.Muted).Italic(true).Render("Waiting...")
	}

	return r.SetContent(eventsView).Render()
}

func (m Model) renderHeader() string {
	colors := theme.GetSemanticColors()

	// Banner on the right - using Cyan (bright blue in Dracula theme)
	// theme.Blue() is actually the muted comment color, Cyan is the vibrant blue
	bannerStyle := lipgloss.NewStyle().
		Foreground(theme.Cyan()).
		Bold(true)
	bannerLines := branding.BannerLines()
	bannerWidth := branding.BannerWidth()

	// Calculate left panel width (what's left after banner)
	leftWidth := m.width - bannerWidth - 4
	if leftWidth < 20 {
		leftWidth = 20
	}

	// Build metadata lines to fill the banner height
	online := m.cache.OnlineCount()
	total := m.cache.DeviceCount()
	offline := total - online
	totalPower := m.cache.TotalPower()

	onlineStyle := lipgloss.NewStyle().Foreground(colors.Online).Bold(true)
	offlineStyle := lipgloss.NewStyle().Foreground(colors.Offline)
	powerStyle := lipgloss.NewStyle().Foreground(colors.Warning).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(colors.Muted)
	valueStyle := lipgloss.NewStyle().Foreground(colors.Text).Bold(true)
	titleStyle := lipgloss.NewStyle().Foreground(colors.Highlight).Bold(true)

	// Create metadata lines to match banner height
	metaLines := make([]string, len(bannerLines))

	// Line 0: Title
	metaLines[0] = titleStyle.Render("Shelly Dashboard")

	// Line 1: Empty for spacing
	metaLines[1] = ""

	// Line 2: Device counts
	if m.cache.IsLoading() {
		fetched := m.cache.FetchedCount()
		metaLines[2] = labelStyle.Render("Devices: ") + fmt.Sprintf("Loading... %d/%d", fetched, total)
	} else {
		metaLines[2] = labelStyle.Render("Devices: ") +
			onlineStyle.Render(fmt.Sprintf("%d online", online))
		if offline > 0 {
			metaLines[2] += labelStyle.Render(" / ") + offlineStyle.Render(fmt.Sprintf("%d offline", offline))
		}
	}

	// Line 3: Power consumption
	if totalPower != 0 {
		powerStr := fmt.Sprintf("%.1fW", totalPower)
		if totalPower >= 1000 || totalPower <= -1000 {
			powerStr = fmt.Sprintf("%.2fkW", totalPower/1000)
		}
		metaLines[3] = labelStyle.Render("Power:   ") + powerStyle.Render(powerStr)
	} else {
		metaLines[3] = labelStyle.Render("Power:   ") + valueStyle.Render("--")
	}

	// Line 4: Current filter
	if m.filter != "" {
		filterStyle := lipgloss.NewStyle().Foreground(colors.Highlight).Italic(true)
		metaLines[4] = labelStyle.Render("Filter:  ") + filterStyle.Render(m.filter)
	} else {
		metaLines[4] = ""
	}

	// Line 5: Theme
	themeName := theme.CurrentThemeName()
	if themeName != "" && len(metaLines) > 5 {
		themeStyle := lipgloss.NewStyle().Foreground(colors.Secondary)
		metaLines[5] = labelStyle.Render("Theme:   ") + themeStyle.Render(themeName)
	}

	// Fill remaining lines if banner is taller
	for i := 6; i < len(metaLines); i++ {
		metaLines[i] = ""
	}

	// Join metadata and banner side by side
	var result strings.Builder
	for i, bannerLine := range bannerLines {
		// Left side: metadata with fixed width
		metaLine := ""
		if i < len(metaLines) {
			metaLine = metaLines[i]
		}
		leftPart := lipgloss.NewStyle().Width(leftWidth).Render(metaLine)

		// Right side: banner line
		rightPart := bannerStyle.Render(bannerLine)

		result.WriteString(leftPart)
		result.WriteString(rightPart)
		if i < len(bannerLines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// renderWithHelpOverlay renders help as a centered overlay on top of main content.
func (m Model) renderWithHelpOverlay(content string, contentHeight int) string {
	helpView := m.help.View()
	if helpView == "" {
		return content
	}

	// Center the help overlay on top of the content
	return lipgloss.Place(
		m.width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		helpView,
		lipgloss.WithWhitespaceChars(" "),
	)
}

// Run starts the TUI application.
func Run(ctx context.Context, f *cmdutil.Factory, opts Options) error {
	m := New(ctx, f, opts)
	defer m.Close()
	p := tea.NewProgram(m,
		tea.WithContext(ctx),
	)
	_, err := p.Run()
	return err
}

func (m Model) getDeviceMethods(d *cache.DeviceData) []string {
	if d == nil || !d.Online {
		return nil
	}
	if d.Device.Generation == 1 || (d.Info != nil && d.Info.Generation == 1) {
		// Gen1 uses REST API paths, not RPC methods
		return []string{
			"/shelly",
			"/status",
			"/settings",
			"/relay/0",
			"/meter/0",
		}
	}

	// Base methods available on all Gen2 devices
	methods := []string{
		"Shelly.GetDeviceInfo",
		"Shelly.GetStatus",
		"Shelly.GetConfig",
		"Sys.GetStatus",
		"Wifi.GetStatus",
		"Cloud.GetStatus",
		"Script.List",
	}

	// Add Switch methods if device has switches
	if len(d.Switches) > 0 {
		methods = append(methods, "Switch.GetStatus", "Switch.GetConfig")
	}

	// Add PM/EM methods only if device has those components
	if d.Snapshot != nil {
		if len(d.Snapshot.PM) > 0 {
			methods = append(methods, "PM.GetStatus")
		}
		if len(d.Snapshot.EM) > 0 {
			methods = append(methods, "EM.GetStatus")
		}
		if len(d.Snapshot.EM1) > 0 {
			methods = append(methods, "EM1.GetStatus")
		}
	}

	return methods
}
