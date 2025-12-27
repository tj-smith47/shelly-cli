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
	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cmdmode"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/confirm"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/devicedetail"
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
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
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
	// PanelEvents is the events panel.
	PanelEvents
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

// horizontalPadding is the left/right padding for tab content.
const horizontalPadding = 1

// PanelWidths holds calculated panel widths for the dashboard layout.
type PanelWidths struct {
	Events     int // Events panel width
	DeviceList int // Device list panel width
	DeviceInfo int // Device info/detail panel width
}

// PanelConstraints defines min/max widths for each panel.
type PanelConstraints struct {
	MinPercent int // Minimum width as percentage of total
	MaxPercent int // Maximum width as percentage of total
	MinChars   int // Minimum absolute character width
}

// EventStreamStartedMsg is sent when the event stream starts successfully.
type EventStreamStartedMsg struct{}

// EventStreamErrorMsg is sent when the event stream fails to start.
type EventStreamErrorMsg struct {
	Err error
}

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

	// Shared event stream (WebSocket for Gen2+, polling for Gen1)
	eventStream *shelly.EventStream

	// Unified focus and keybinding management
	focusManager *focus.Manager
	contextMap   *keys.ContextMap

	// State
	ready                   bool
	quitting                bool
	cursor                  int        // Selected device index
	componentCursor         int        // Selected component within device (-1 = all)
	filter                  string     // Device name filter
	focusedPanel            PanelFocus // Which panel is currently focused
	endpointCursor          int        // Selected endpoint in JSON panel
	initialSelectionEmitted bool       // Whether initial device selection has been emitted

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
	deviceDetail  devicedetail.Model

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

	// Create shared event stream (WebSocket for Gen2+, polling for Gen1)
	eventStream := shelly.NewEventStream(f.ShellyService())

	// Subscribe cache to event stream for real-time updates
	deviceCache.SubscribeToEvents(eventStream)

	// Create focus manager for Dashboard panels
	focusMgr := focus.New(
		focus.PanelEvents,
		focus.PanelDeviceList,
		focus.PanelDeviceInfo,
		focus.PanelEnergy,
	)

	// Create context-sensitive keybinding map
	contextKeyMap := keys.NewContextMap()

	// Create search component with initial filter
	searchModel := search.NewWithFilter(opts.Filter)

	// Create events component for real-time event stream
	eventsModel := events.New(events.Deps{
		Ctx:         ctx,
		Svc:         f.ShellyService(),
		IOS:         f.IOStreams(),
		EventStream: eventStream,
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

	// Create device detail overlay component
	deviceDetailModel := devicedetail.New(devicedetail.Deps{
		Ctx: ctx,
		Svc: f.ShellyService(),
	})

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

	// Register Monitor view
	monitorView := views.NewMonitor(views.MonitorDeps{
		Ctx:         ctx,
		Svc:         f.ShellyService(),
		IOS:         f.IOStreams(),
		EventStream: eventStream,
	})
	vm.Register(monitorView)

	// Register Fleet view
	fleet := views.NewFleet(views.FleetDeps{
		Ctx: ctx,
		Svc: f.ShellyService(),
		IOS: f.IOStreams(),
		Cfg: cfg,
	})
	vm.Register(fleet)

	// Create tab bar
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
		eventStream:     eventStream,
		focusManager:    focusMgr,
		contextMap:      contextKeyMap,
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
		deviceDetail:    deviceDetailModel,
		debugLogger:     debug.New(), // nil if SHELLY_TUI_DEBUG not set
	}
}

// Close cleans up resources used by the TUI.
func (m Model) Close() {
	// Stop the event stream and close all WebSocket connections
	if m.eventStream != nil {
		m.eventStream.Stop()
	}
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

// calculateOptimalWidths determines panel widths based on content.
// New layout: Device List (left) | Device Info + Events stacked (right)
// Events and DeviceInfo share the right column width (stacked vertically).
func (m Model) calculateOptimalWidths() PanelWidths {
	totalWidth := m.contentWidth()

	// Get content-based optimal width for device list
	deviceListOptimal := m.deviceList.OptimalWidth()

	// Define constraints as percentages
	deviceListMin := totalWidth * 15 / 100
	deviceListMax := totalWidth * 35 / 100 // Don't let device list take more than 35%

	// Apply constraints to device list width
	deviceListWidth := deviceListOptimal
	if deviceListWidth < deviceListMin {
		deviceListWidth = deviceListMin
	}
	if deviceListWidth > deviceListMax {
		deviceListWidth = deviceListMax
	}

	// Minimum absolute width for device list
	if deviceListWidth < 25 {
		deviceListWidth = 25
	}

	// Right column gets the rest (device info + events are stacked vertically)
	gap := 1
	rightColWidth := totalWidth - deviceListWidth - gap

	// Events and DeviceInfo share the rightColWidth (they're stacked, not side by side)
	// So we set Events = DeviceInfo = rightColWidth for the PanelWidths struct
	return PanelWidths{
		Events:     rightColWidth, // Events uses full right column width
		DeviceList: deviceListWidth,
		DeviceInfo: rightColWidth, // DeviceInfo uses full right column width
	}
}

// Init initializes the TUI and returns the first command to run.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.cache.Init(),
		m.startEventStream(),
		m.statusBar.Init(),
		m.toast.Init(),
		m.events.Init(),
		m.viewManager.Init(),
	)
}

// startEventStream starts the shared WebSocket event stream.
func (m Model) startEventStream() tea.Cmd {
	return func() tea.Msg {
		if err := m.eventStream.Start(); err != nil {
			return EventStreamErrorMsg{Err: err}
		}
		return EventStreamStartedMsg{}
	}
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
		return m.handleDeviceUpdate(msg)
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
	case search.ClosedMsg, cmdmode.ClosedMsg:
		m = m.updateViewManagerSize()
		return m, nil, true
	case confirm.CancelledMsg:
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
	case devicelist.OpenBrowserMsg:
		// Open device web UI in browser
		return m, m.openDeviceBrowser(msg.Address), true
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
	case devicedetail.Msg, devicedetail.ClosedMsg:
		var cmd tea.Cmd
		m.deviceDetail, cmd = m.deviceDetail.Update(msg)
		return m, cmd, true
	case views.ReturnFocusMsg:
		// View is returning focus to device list (Tab/Shift+Tab past first/last panel)
		m.focusedPanel = PanelDeviceList
		return m, nil, true
	}
	return m, nil, false
}

// handleDeviceUpdate handles cache.DeviceUpdateMsg by forwarding to cache and energy history.
func (m Model) handleDeviceUpdate(msg cache.DeviceUpdateMsg) (tea.Model, tea.Cmd, bool) {
	// Forward to cache AND energyHistory (to record power history)
	cacheCmd := m.cache.Update(msg)
	var historyCmd tea.Cmd
	m.energyHistory, historyCmd = m.energyHistory.Update(msg)
	m = m.updateStatusBarContext()

	// Sync device info panel with updated cache data
	// This ensures the panel reflects any state changes from toggle/on/off actions
	m = m.syncDeviceInfo()

	cmds := []tea.Cmd{cacheCmd, historyCmd}

	// Emit initial selection when first device becomes available
	if !m.initialSelectionEmitted {
		if selectionCmd, emitted := m.emitInitialSelection(); emitted {
			m.initialSelectionEmitted = true
			cmds = append(cmds, selectionCmd)
		}
	}
	return m, tea.Batch(cmds...), true
}

// emitInitialSelection emits a DeviceSelectedMsg for the first device if available.
// Returns the command and true if selection was emitted, nil and false otherwise.
func (m Model) emitInitialSelection() (tea.Cmd, bool) {
	devices := m.getFilteredDevices()
	if len(devices) == 0 || m.cursor >= len(devices) {
		return nil, false
	}
	d := devices[m.cursor]
	return func() tea.Msg {
		return views.DeviceSelectedMsg{
			Device:  d.Device.Name,
			Address: d.Device.Address,
		}
	}, true
}

// handleAllDevicesLoaded handles the AllDevicesLoadedMsg and emits initial selection.
func (m Model) handleAllDevicesLoaded(msg cache.AllDevicesLoadedMsg) (tea.Model, tea.Cmd, bool) {
	cmd := m.cache.Update(msg)
	m = m.updateStatusBarContext()
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

	// If device detail overlay is visible, forward all keys to it
	if m.deviceDetail.Visible() {
		var cmd tea.Cmd
		m.deviceDetail, cmd = m.deviceDetail.Update(msg)
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

	// Set deviceList size - needed for proper scroll calculations
	// The actual layout may adjust this in render, but we need a valid
	// size for navigation to work correctly (visibleRows() calculation)
	deviceListHeight := contentHeight - 2 // Account for borders
	if deviceListHeight < 5 {
		deviceListHeight = 5
	}
	deviceListWidth := m.width * 40 / 100 // 40% for list panel
	if deviceListWidth < 20 {
		deviceListWidth = 20
	}
	m.deviceList = m.deviceList.SetSize(deviceListWidth, deviceListHeight)

	m = m.updateStatusBarContext()
	return m
}

// updateViewManagerSize recalculates and sets the view manager size
// based on current input bar state (search or command mode).
func (m Model) updateViewManagerSize() Model {
	tabBarHeight := 1
	footerHeight := 2
	inputHeight := 0
	if m.search.IsActive() || m.cmdMode.IsActive() {
		inputHeight = 3 // Input bar has top border, content, bottom border
	}
	contentHeight := m.height - branding.BannerHeight() - tabBarHeight - footerHeight - inputHeight
	m.viewManager.SetSize(m.width, contentHeight)
	return m
}

// handleDeviceAction handles device action results.
func (m Model) handleDeviceAction(msg DeviceActionMsg) (tea.Model, tea.Cmd) {
	var statusCmd tea.Cmd
	var eventLevel, eventDesc string

	if msg.Err != nil {
		statusCmd = statusbar.SetMessage(
			msg.Device+": "+msg.Action+" failed - "+msg.Err.Error(),
			statusbar.MessageError,
		)
		eventLevel = "error"
		eventDesc = msg.Action + " failed: " + msg.Err.Error()
	} else {
		statusCmd = statusbar.SetMessage(
			msg.Device+": "+msg.Action+" success",
			statusbar.MessageSuccess,
		)
		eventLevel = "info"
		eventDesc = msg.Action + " executed successfully"
	}

	// Emit event to events panel for visibility
	evt := events.Event{
		Timestamp:   time.Now(),
		Device:      msg.Device,
		Component:   "command",
		Type:        eventLevel,
		Description: eventDesc,
	}
	var evtCmd tea.Cmd
	m.events, evtCmd = m.events.Update(events.EventMsg{Events: []events.Event{evt}})

	return m, tea.Batch(statusCmd, evtCmd)
}

// updateStatusBarContext updates the status bar with context-specific items.
func (m Model) updateStatusBarContext() Model {
	total := m.cache.DeviceCount()
	online := m.cache.OnlineCount()

	m.statusBar = m.statusBar.ClearItems()

	if total > 0 {
		m.statusBar = m.statusBar.AddItem(
			fmt.Sprintf("%d devices (%d online)", total, online),
			fmt.Sprintf("%d (%d on)", total, online),
			fmt.Sprintf("%d/%d", total, online),
		)
	}

	if m.filter != "" {
		m.statusBar = m.statusBar.AddItem(
			fmt.Sprintf("Filter: %s", m.filter),
			fmt.Sprintf("F: %s", m.filter),
			"F",
		)
	}

	return m
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
		m.deviceList = m.deviceList.SetFilter(msg.Args)
		m.deviceList = m.deviceList.SetCursor(0)
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

// handleKeyPress handles global key presses using the context-sensitive keybinding system.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	// Get current context based on focused panel and active view
	ctx := m.getCurrentKeyContext()

	// Try context-sensitive keybinding system first
	action := m.contextMap.Match(ctx, msg)
	if action != keys.ActionNone {
		newModel, cmd, handled := m.dispatchAction(action)
		if handled {
			return newModel, cmd, true
		}
	}

	// Fall back to legacy handlers for actions not handled by dispatchAction
	// Global actions (view switching, quit, help, etc.)
	if newModel, cmd, handled := m.handleGlobalKeys(msg); handled {
		return newModel, cmd, true
	}

	// Device actions work on any tab with a device list (Dashboard, Monitor, Automation, Config)
	// Try device actions BEFORE forwarding to views so t/o/f/r keys work everywhere
	if m.hasDeviceList() {
		if newModel, cmd, handled := m.handleDeviceActionKey(msg); handled {
			return newModel, cmd, true
		}
	}

	// For tabs with device list (Automation, Config, Monitor), handle focus management
	if m.hasDeviceList() && !m.isDashboardActive() {
		newModel, cmd := m.handleDeviceListTabKeyPress(msg)
		return newModel, cmd, true
	}

	// If NOT on Dashboard and NOT a device-list tab, forward all keys to view
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

	return m, nil, false
}

// getCurrentKeyContext returns the appropriate keybinding context based on current state.
func (m Model) getCurrentKeyContext() keys.Context {
	// Check view-specific context first
	switch m.tabBar.ActiveTabID() {
	case tabs.TabDashboard:
		// Dashboard uses panel-based context
		return m.panelToContext(m.focusedPanel)
	case tabs.TabAutomation:
		return keys.ContextAutomation
	case tabs.TabConfig:
		return keys.ContextConfig
	case tabs.TabManage:
		return keys.ContextManage
	case tabs.TabFleet:
		return keys.ContextFleet
	case tabs.TabMonitor:
		return keys.ContextDevices // Monitor uses device-like navigation
	default:
		return m.panelToContext(m.focusedPanel)
	}
}

// panelToContext converts a PanelFocus to a keys.Context.
func (m Model) panelToContext(panel PanelFocus) keys.Context {
	switch panel {
	case PanelEvents:
		return keys.ContextEvents
	case PanelDeviceList:
		return keys.ContextDevices
	case PanelDetail:
		return keys.ContextInfo
	case PanelEndpoints:
		return keys.ContextJSON
	default:
		return keys.ContextGlobal
	}
}

// handleDeviceListTabKeyPress handles key presses for tabs with device list (Automation, Config, Monitor).
// Always handles the key (caller returns handled=true).
func (m Model) handleDeviceListTabKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	keyStr := msg.String()

	// Shift+1 (!) ALWAYS returns focus to device list, regardless of current focus
	if keyStr == keyconst.Shift1 {
		m.focusedPanel = PanelDeviceList
		cmd := m.viewManager.Update(views.ViewFocusChangedMsg{Focused: false})
		return m, cmd
	}

	// Escape ALWAYS returns focus to device list
	if keyStr == "esc" {
		m.focusedPanel = PanelDeviceList
		cmd := m.viewManager.Update(views.ViewFocusChangedMsg{Focused: false})
		return m, cmd
	}

	// When device list is focused, try to handle navigation keys here
	if m.focusedPanel == PanelDeviceList {
		newModel, cmd, handled := m.handleDeviceListFocusedKeys(msg)
		if handled {
			return newModel, cmd
		}
		// Non-navigation keys (Enter, e, d, etc.) should go to the view
		// so actions like "edit script" work even when device list is focused
		cmd = m.viewManager.Update(msg)
		return m, cmd
	}

	// When focused on view (PanelDetail), forward keys to the view
	// The view handles its own internal panel cycling via Tab/Shift+Tab
	cmd := m.viewManager.Update(msg)
	return m, cmd
}

// handleDeviceListFocusedKeys handles keys when device list panel is focused.
func (m Model) handleDeviceListFocusedKeys(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	keyStr := msg.String()
	switch keyStr {
	case "tab", "shift+tab":
		// Both Tab and Shift+Tab move focus to the view
		m.focusedPanel = PanelDetail
		// Tell the view it now has focus
		cmd := m.viewManager.Update(views.ViewFocusChangedMsg{Focused: true})
		return m, cmd, true
	}
	// Shift+1 (!) keeps focus on device list (panel 1 on all tabs with device list)
	if keyStr == keyconst.Shift1 {
		m.focusedPanel = PanelDeviceList
		// Tell the view it no longer has focus
		cmd := m.viewManager.Update(views.ViewFocusChangedMsg{Focused: false})
		return m, cmd, true
	}
	// Shift+2-9 (@, #, $, %, ^, &, *, () switch focus to view and forward the key for panel selection
	if isShiftNumberKey(keyStr) && keyStr != keyconst.Shift1 {
		m.focusedPanel = PanelDetail
		cmd := m.viewManager.Update(msg)
		return m, cmd, true
	}
	// Navigation keys for device list
	if newModel, cmd, handled := m.handleNavigation(msg); handled {
		return newModel, cmd, true
	}
	// Other keys still go to device list
	return m, nil, false
}

// handlePanelSwitch handles Tab/Shift+Tab for switching panels, Shift+N for direct jump, and Enter for JSON overlay.
func (m Model) handlePanelSwitch(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		m.focusedPanel = m.nextPanel()
		return m, nil, true
	case "shift+tab":
		m.focusedPanel = m.prevPanel()
		return m, nil, true
	case keyconst.Shift1:
		m.focusedPanel = PanelDeviceList
		return m, nil, true
	case keyconst.Shift2:
		m.focusedPanel = PanelDetail
		return m, nil, true
	case keyconst.Shift3:
		m.focusedPanel = PanelEvents
		return m, nil, true
	case "enter":
		return m.openJSONViewer()
	case "esc":
		if m.jsonViewer.Visible() {
			m.jsonViewer = m.jsonViewer.Close()
			m.focusedPanel = PanelDetail
			return m, nil, true
		}
	}
	return m, nil, false
}

// nextPanel returns the next panel in the cycle: DeviceList -> Detail -> Events -> DeviceList.
func (m Model) nextPanel() PanelFocus {
	switch m.focusedPanel {
	case PanelDeviceList:
		return PanelDetail
	case PanelDetail:
		return PanelEvents
	default:
		return PanelDeviceList
	}
}

// prevPanel returns the previous panel in the cycle: Events -> Detail -> DeviceList -> Events.
func (m Model) prevPanel() PanelFocus {
	switch m.focusedPanel {
	case PanelEvents:
		return PanelDetail
	case PanelDetail:
		return PanelDeviceList
	default:
		return PanelEvents
	}
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
		m = m.updateViewManagerSize()
		return m, cmd, true
	case key.Matches(msg, m.keys.Command):
		var cmd tea.Cmd
		m.cmdMode, cmd = m.cmdMode.Activate()
		m = m.updateViewManagerSize()
		return m, cmd, true
	case key.Matches(msg, m.keys.Refresh):
		// Trigger a full cache refresh by sending RefreshTickMsg
		return m, func() tea.Msg { return cache.RefreshTickMsg{} }, true
	case key.Matches(msg, m.keys.Escape):
		if m.filter != "" {
			m.filter = ""
			m.cursor = 0
			m.deviceList = m.deviceList.SetFilter("")
			m.deviceList = m.deviceList.SetCursor(0)
			return m, nil, true
		}
	case key.Matches(msg, m.keys.View1):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabDashboard)
		m.focusedPanel = PanelDeviceList // Reset focus to device list
		return m, m.viewManager.SetActive(views.ViewDashboard), true
	case key.Matches(msg, m.keys.View2):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabAutomation)
		m.focusedPanel = PanelDeviceList // Reset focus to device list for sidebar layout
		return m, m.viewManager.SetActive(views.ViewAutomation), true
	case key.Matches(msg, m.keys.View3):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabConfig)
		m.focusedPanel = PanelDeviceList // Reset focus to device list for sidebar layout
		return m, m.viewManager.SetActive(views.ViewConfig), true
	case key.Matches(msg, m.keys.View4):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabManage)
		return m, m.viewManager.SetActive(views.ViewManage), true
	case key.Matches(msg, m.keys.View5):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabMonitor)
		m.focusedPanel = PanelDeviceList // Reset focus to device list
		return m, m.viewManager.SetActive(views.ViewMonitor), true
	case key.Matches(msg, m.keys.View6):
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabFleet)
		return m, m.viewManager.SetActive(views.ViewFleet), true
	}
	return m, nil, false
}

// dispatchAction handles an action from the context-sensitive keybinding system.
// Returns the updated model, command, and whether the action was handled.
func (m Model) dispatchAction(action keys.Action) (Model, tea.Cmd, bool) {
	// Handle global actions first
	if newModel, cmd, handled := m.dispatchGlobalAction(action); handled {
		return newModel, cmd, true
	}

	// Handle tab switching actions
	if newModel, cmd, handled := m.dispatchTabAction(action); handled {
		return newModel, cmd, true
	}

	// Handle panel jump actions
	if newModel, cmd, handled := m.dispatchPanelAction(action); handled {
		return newModel, cmd, true
	}

	// Handle device actions
	if newModel, cmd, handled := m.dispatchDeviceKeyAction(action); handled {
		return newModel, cmd, true
	}

	// Actions not directly handled - let existing handlers deal with them
	// This includes: ActionNone, navigation, events panel actions, etc.
	return m, nil, false
}

// dispatchGlobalAction handles global actions like quit, help, filter, command mode.
func (m Model) dispatchGlobalAction(action keys.Action) (Model, tea.Cmd, bool) {
	switch action {
	case keys.ActionQuit:
		m.quitting = true
		return m, tea.Quit, true

	case keys.ActionHelp:
		m.help = m.help.SetSize(m.width, m.height)
		m.help = m.help.SetContext(m.getHelpContext())
		m.help = m.help.Toggle()
		return m, nil, true

	case keys.ActionFilter:
		var cmd tea.Cmd
		m.search, cmd = m.search.Activate()
		m = m.updateViewManagerSize()
		return m, cmd, true

	case keys.ActionCommand:
		var cmd tea.Cmd
		m.cmdMode, cmd = m.cmdMode.Activate()
		m = m.updateViewManagerSize()
		return m, cmd, true

	case keys.ActionRefresh:
		return m, func() tea.Msg { return cache.RefreshTickMsg{} }, true

	case keys.ActionNextPanel:
		// Only handle panel cycling for Dashboard - other tabs handle it in their views
		if m.isDashboardActive() {
			m.focusedPanel = m.nextPanel()
			return m, nil, true
		}
		return m, nil, false // Let views handle Tab for their internal panels

	case keys.ActionPrevPanel:
		// Only handle panel cycling for Dashboard - other tabs handle it in their views
		if m.isDashboardActive() {
			m.focusedPanel = m.prevPanel()
			return m, nil, true
		}
		return m, nil, false // Let views handle Shift+Tab for their internal panels

	case keys.ActionEscape:
		if m.filter != "" {
			m.filter = ""
			m.cursor = 0
			m.deviceList = m.deviceList.SetFilter("")
			m.deviceList = m.deviceList.SetCursor(0)
			return m, nil, true
		}
		return m, nil, false

	default:
		return m, nil, false
	}
}

// dispatchTabAction handles tab switching actions (1-5).
func (m Model) dispatchTabAction(action keys.Action) (Model, tea.Cmd, bool) {
	switch action {
	case keys.ActionTab1:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabDashboard)
		m.focusedPanel = PanelDeviceList
		return m, m.viewManager.SetActive(views.ViewDashboard), true

	case keys.ActionTab2:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabAutomation)
		m.focusedPanel = PanelDeviceList
		// View starts unfocused (device list has focus)
		return m, tea.Batch(
			m.viewManager.SetActive(views.ViewAutomation),
			func() tea.Msg { return views.ViewFocusChangedMsg{Focused: false} },
		), true

	case keys.ActionTab3:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabConfig)
		m.focusedPanel = PanelDeviceList
		// View starts unfocused (device list has focus)
		return m, tea.Batch(
			m.viewManager.SetActive(views.ViewConfig),
			func() tea.Msg { return views.ViewFocusChangedMsg{Focused: false} },
		), true

	case keys.ActionTab4:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabManage)
		return m, m.viewManager.SetActive(views.ViewManage), true

	case keys.ActionTab5:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabMonitor)
		m.focusedPanel = PanelDeviceList
		return m, m.viewManager.SetActive(views.ViewMonitor), true

	default:
		return m, nil, false
	}
}

// dispatchPanelAction handles panel jumping actions (Shift+1-9).
// Only handles for Dashboard - other tabs have their own panel numbering.
//
//nolint:unparam // tea.Cmd is nil but kept for consistent dispatch function signature
func (m Model) dispatchPanelAction(action keys.Action) (Model, tea.Cmd, bool) {
	// Only handle panel jumps for Dashboard - other tabs handle Shift+N in their views
	if !m.isDashboardActive() {
		return m, nil, false
	}

	switch action {
	case keys.ActionPanel1:
		m.focusedPanel = PanelDeviceList
		return m, nil, true
	case keys.ActionPanel2:
		m.focusedPanel = PanelDetail
		return m, nil, true
	case keys.ActionPanel3:
		m.focusedPanel = PanelEvents
		return m, nil, true
	case keys.ActionPanel4, keys.ActionPanel5, keys.ActionPanel6,
		keys.ActionPanel7, keys.ActionPanel8, keys.ActionPanel9:
		// Panel 4-9 not mapped in Dashboard layout
		return m, nil, false
	default:
		return m, nil, false
	}
}

// dispatchDeviceKeyAction handles device control actions (toggle, on, off, reboot, enter).
func (m Model) dispatchDeviceKeyAction(action keys.Action) (Model, tea.Cmd, bool) {
	switch action {
	case keys.ActionToggle:
		return m.dispatchDeviceAction("toggle")
	case keys.ActionOn:
		return m.dispatchDeviceAction("on")
	case keys.ActionOff:
		return m.dispatchDeviceAction("off")
	case keys.ActionReboot:
		return m.dispatchDeviceAction("reboot")
	case keys.ActionEnter:
		return m.dispatchEnterAction()
	default:
		return m, nil, false
	}
}

// dispatchDeviceAction executes a device action on the selected device.
func (m Model) dispatchDeviceAction(action string) (Model, tea.Cmd, bool) {
	if !m.hasDeviceList() {
		return m, nil, false
	}
	cmd := m.executeDeviceAction(action)
	return m, cmd, true
}

// dispatchEnterAction handles the Enter key action based on context.
// Only handles Enter on Dashboard - other tabs forward to their views.
func (m Model) dispatchEnterAction() (Model, tea.Cmd, bool) {
	// Only handle Enter for Dashboard tab
	if !m.isDashboardActive() {
		return m, nil, false
	}
	if m.focusedPanel == PanelDetail {
		return m.openJSONViewer()
	}
	// Default: forward to view
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

// isNavigationKey returns true if the key is a navigation key that should be
// handled by list components (device list, scripts list, etc.).
func isNavigationKey(keyStr string) bool {
	switch keyStr {
	case "j", "k", "up", "down", "g", "G", "pgdown", "pgup", "ctrl+d", "ctrl+u":
		return true
	default:
		return false
	}
}

// isShiftNumberKey returns true if the key is a Shift+Number key (!, @, #, etc.).
func isShiftNumberKey(keyStr string) bool {
	switch keyStr {
	case keyconst.Shift1, keyconst.Shift2, keyconst.Shift3,
		keyconst.Shift4, keyconst.Shift5, keyconst.Shift6,
		keyconst.Shift7, keyconst.Shift8, keyconst.Shift9:
		return true
	default:
		return false
	}
}

// handleNavigation handles j/k/g/G/h/l navigation keys based on focused panel.
func (m Model) handleNavigation(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	// When PanelDeviceList is focused, only forward NAVIGATION keys to deviceList
	if m.focusedPanel == PanelDeviceList {
		// Only handle actual navigation keys - other keys should fall through
		if !isNavigationKey(msg.String()) {
			return m, nil, false
		}
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

	// When PanelEvents is focused, forward navigation to events component
	if m.focusedPanel == PanelEvents {
		var cmd tea.Cmd
		m.events, cmd = m.events.Update(msg)
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

	// Update focus state and panel index
	m.deviceInfo = m.deviceInfo.SetFocused(m.focusedPanel == PanelDetail).SetPanelIndex(2)

	return m
}

// handleDeviceActionKey handles device action key presses.
func (m Model) handleDeviceActionKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	// Handle 'd' to show device detail overlay
	if msg.String() == "d" || msg.String() == "D" {
		return m.showDeviceDetail()
	}

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

// showDeviceDetail shows the device detail overlay for the selected device.
func (m Model) showDeviceDetail() (tea.Model, tea.Cmd, bool) {
	devices := m.getFilteredDevices()
	if m.cursor >= len(devices) || m.cursor < 0 {
		return m, nil, false
	}

	selected := devices[m.cursor]
	if !selected.Online {
		return m, nil, false
	}

	// Set size for the overlay
	m.deviceDetail = m.deviceDetail.SetSize(m.width*2/3, m.height*2/3)

	// Show the overlay with the device
	var cmd tea.Cmd
	m.deviceDetail, cmd = m.deviceDetail.Show(selected.Device)
	return m, cmd, true
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

	// Render all view components
	headerBanner := m.renderHeader()
	tabBarView := m.tabBar.View()
	inputBar := m.renderInputBar()
	contentHeight := m.calculateContentHeight(inputBar)
	content := m.renderTabContent(contentHeight)
	content = m.padContent(content, contentHeight)

	if m.help.Visible() {
		content = m.renderWithHelpOverlay(content, contentHeight)
	}

	// Compose the layout
	result := m.composeLayout(headerBanner, tabBarView, inputBar, content)
	result = m.applyOverlays(result)

	m.debugLogger.Log(m.tabBar.ActiveTabID().String(), m.focusedPanelName(), m.width, m.height, result)

	v := tea.NewView(result)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// renderInputBar renders the search or command mode input bar.
func (m Model) renderInputBar() string {
	if m.search.IsActive() {
		return m.search.View()
	}
	if m.cmdMode.IsActive() {
		return m.cmdMode.View()
	}
	return ""
}

// calculateContentHeight calculates the available height for main content.
func (m Model) calculateContentHeight(inputBar string) int {
	bannerHeight := branding.BannerHeight()
	tabBarHeight := 1
	footerHeight := 2
	inputHeight := 0
	if inputBar != "" {
		inputHeight = 3
	}
	return m.height - bannerHeight - tabBarHeight - footerHeight - inputHeight
}

// renderTabContent renders content based on active tab.
func (m Model) renderTabContent(contentHeight int) string {
	cw := m.contentWidth()
	switch m.tabBar.ActiveTabID() {
	case tabs.TabDashboard:
		return m.renderMultiPanelLayout(contentHeight)
	case tabs.TabMonitor:
		return m.renderMonitorLayout(contentHeight)
	case tabs.TabAutomation, tabs.TabConfig:
		return m.renderWithDeviceList(contentHeight)
	default:
		m.viewManager.SetSize(cw, contentHeight)
		return lipgloss.NewStyle().Width(cw).Render(m.viewManager.View())
	}
}

// padContent pads content to fill the entire content area with horizontal padding.
func (m Model) padContent(content string, contentHeight int) string {
	cw := m.contentWidth()
	pad := strings.Repeat(" ", horizontalPadding)
	contentLines := strings.Split(content, "\n")
	paddedLines := make([]string, contentHeight)
	emptyLine := strings.Repeat(" ", m.width)
	for i := range contentHeight {
		if i < len(contentLines) {
			line := contentLines[i]
			lineWidth := lipgloss.Width(line)
			if lineWidth < cw {
				line += strings.Repeat(" ", cw-lineWidth)
			}
			paddedLines[i] = pad + line + pad
		} else {
			paddedLines[i] = emptyLine
		}
	}
	return strings.Join(paddedLines, "\n")
}

// composeLayout joins header, tabs, input, content, and footer.
func (m Model) composeLayout(header, tabBar, inputBar, content string) string {
	footer := m.statusBar.View()
	if inputBar != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, inputBar, content, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, content, footer)
}

// applyOverlays applies toast, confirm, and device detail overlays.
func (m Model) applyOverlays(result string) string {
	if m.toast.HasToasts() {
		result = m.toast.Overlay(result)
	}
	if m.confirm.Visible() {
		result = m.centerOverlay(m.confirm.View())
	}
	if m.deviceDetail.Visible() {
		result = m.centerOverlay(m.deviceDetail.View())
	}
	return result
}

// centerOverlay centers an overlay on the screen.
func (m Model) centerOverlay(overlay string) string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}

// focusedPanelName returns a human-readable name for the currently focused panel.
func (m Model) focusedPanelName() string {
	switch m.focusedPanel {
	case PanelDeviceList:
		return "DeviceList"
	case PanelDetail:
		return "Detail"
	case PanelEvents:
		return "Events"
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

// contentWidth returns the width available for tab content (accounting for horizontal padding).
func (m Model) contentWidth() int {
	return m.width - (2 * horizontalPadding)
}

// hasDeviceList returns true if the current tab has a device list for device actions.
// This includes Dashboard, Monitor, Automation, and Config tabs.
func (m Model) hasDeviceList() bool {
	switch m.tabBar.ActiveTabID() {
	case tabs.TabDashboard, tabs.TabMonitor, tabs.TabAutomation, tabs.TabConfig:
		return true
	default:
		return false
	}
}

// renderMultiPanelLayout renders panels horizontally or vertically based on width.
// Layout: Device List (left) | Device Info (top-right) / Events (bottom-right)
// Energy bars at bottom spanning full width.
func (m Model) renderMultiPanelLayout(height int) string {
	// Narrow mode: stack panels vertically
	if m.layoutMode() == LayoutNarrow {
		return m.renderNarrowLayout(height)
	}

	// Split height: top 70% for panels, bottom 30% for energy bars
	topHeight := height * 70 / 100
	energyHeight := height - topHeight - 1 // -1 for gap

	if topHeight < 10 {
		topHeight = 10
	}
	if energyHeight < 5 {
		energyHeight = 5
	}

	// Calculate dynamic panel widths based on content
	cw := m.contentWidth()
	widths := m.calculateOptimalWidths()

	// Device list on the left (persists across views)
	deviceListCol := m.renderDeviceListColumn(widths.DeviceList, topHeight)

	// Right column width = total - device list - gap
	rightColWidth := cw - widths.DeviceList - 1

	// Split right column vertically: device info (top 30%) and events (bottom 70%)
	// Total must equal topHeight so right column matches device list height
	infoHeight := topHeight * 30 / 100
	eventsHeight := topHeight - infoHeight // No gap - JoinVertical stacks directly

	if infoHeight < 6 {
		infoHeight = 6
		eventsHeight = topHeight - infoHeight
	}
	if eventsHeight < 12 {
		eventsHeight = 12
	}

	// Render device info component (top-right)
	m.deviceInfo = m.deviceInfo.SetSize(rightColWidth, infoHeight)
	m.deviceInfo = m.deviceInfo.SetFocused(m.focusedPanel == PanelDetail).SetPanelIndex(2)
	deviceInfoPanel := m.deviceInfo.View()

	// Render events (bottom-right) - now gets much more horizontal space
	eventsPanel := m.renderEventsColumn(rightColWidth, eventsHeight)

	// Stack info and events vertically
	rightCol := lipgloss.JoinVertical(lipgloss.Left, deviceInfoPanel, eventsPanel)

	// Combine left and right columns
	topContent := lipgloss.JoinHorizontal(lipgloss.Top, deviceListCol, " ", rightCol)

	// Ensure topContent fills the full width
	topContentStyle := lipgloss.NewStyle().Width(cw)
	topContent = topContentStyle.Render(topContent)

	// Energy bars panel at bottom (full width)
	energyPanel := m.renderEnergyPanel(cw, energyHeight)

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

// renderMonitorLayout renders the Monitor tab with energy panels.
// Layout: Monitor component on top (60%), Energy panels on bottom (40%).
func (m Model) renderMonitorLayout(height int) string {
	cw := m.contentWidth()

	// Split height: top 60% for monitor, bottom 40% for energy panels
	monitorHeight := height * 60 / 100
	energyHeight := height - monitorHeight

	if monitorHeight < 10 {
		monitorHeight = 10
	}
	if energyHeight < 8 {
		energyHeight = 8
	}

	// Get monitor view and wrap with rendering.New() for embedded title
	// Content area is 2 less in each dimension to account for border
	contentWidth := cw - 2
	contentHeight := monitorHeight - 2
	m.viewManager.SetSize(contentWidth, contentHeight)
	monitorContent := m.viewManager.View()

	// Wrap with rendering.New() for superfile-style embedded title
	r := rendering.New(cw, monitorHeight).
		SetTitle("Monitor").
		SetContent(monitorContent).
		SetFocused(true) // Monitor panel is always focused when on Monitor tab
	monitorView := r.Render()

	// Render energy panels at bottom
	energyPanel := m.renderEnergyPanel(cw, energyHeight)

	// Combine top and bottom sections
	return lipgloss.JoinVertical(lipgloss.Left, monitorView, energyPanel)
}

// renderWithDeviceList wraps a view with the device list on the left for device persistence.
// Used for Automation and Config tabs to allow device switching while on those views.
func (m Model) renderWithDeviceList(height int) string {
	cw := m.contentWidth()
	widths := m.calculateOptimalWidths()

	// Device list on the left (persists across views)
	deviceListCol := m.renderDeviceListColumn(widths.DeviceList, height)

	// Right column width = total - device list - gap
	rightColWidth := cw - widths.DeviceList - 1

	// Get the view content from view manager
	// Resize the view to fit the available space
	m.viewManager.SetSize(rightColWidth, height)
	viewContent := m.viewManager.View()

	// Combine left device list and right view content
	content := lipgloss.JoinHorizontal(lipgloss.Top, deviceListCol, " ", viewContent)

	// Ensure content fills the full width
	contentStyle := lipgloss.NewStyle().Width(cw)
	content = contentStyle.Render(content)

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
	// Guard against too-small dimensions that would cause title overlap
	if width < 40 || height < 3 {
		return ""
	}

	// Split width between bars and history
	barsWidth := width * 60 / 100
	historyWidth := width - barsWidth - 1

	// Ensure minimum widths for each panel (title needs at least ~25 chars)
	if barsWidth < 30 {
		barsWidth = 30
	}
	if historyWidth < 30 {
		historyWidth = width - barsWidth - 1
		if historyWidth < 10 {
			// Not enough space, render only bars
			m.energyBars = m.energyBars.SetSize(width, height)
			return m.energyBars.View()
		}
	}

	m.energyBars = m.energyBars.SetSize(barsWidth, height)
	m.energyHistory = m.energyHistory.SetSize(historyWidth, height)

	barsView := m.energyBars.View()
	historyView := m.energyHistory.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, barsView, " ", historyView)
}

// renderNarrowLayout renders panels stacked vertically for narrow terminals.
func (m Model) renderNarrowLayout(height int) string {
	cw := m.contentWidth()
	panelWidth := cw - 2 // Content width minus borders

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

	// Row 2: Device List (list only) + Device Info stacked in narrow mode
	// In narrow mode, we keep devicelist's internal split pane for compactness
	m.deviceList = m.deviceList.SetSize(panelWidth, deviceListHeight)
	m.deviceList = m.deviceList.SetFocused(m.focusedPanel == PanelDeviceList).SetPanelIndex(1)
	m.deviceList = m.deviceList.SetListOnly(false) // Use split pane in narrow mode
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

// renderDeviceListColumn renders the device list with consistent border styling.
func (m Model) renderDeviceListColumn(width, height int) string {
	focused := m.focusedPanel == PanelDeviceList

	// Account for border when setting component size
	innerWidth := width - 2   // left + right border
	innerHeight := height - 2 // top + bottom border

	m.deviceList = m.deviceList.SetSize(innerWidth, innerHeight)
	m.deviceList = m.deviceList.SetFocused(focused).SetPanelIndex(1)
	m.deviceList = m.deviceList.SetListOnly(true)

	deviceCount := m.cache.DeviceCount()
	onlineCount := m.cache.OnlineCount()

	r := rendering.New(width, height).
		SetTitle("Devices").
		SetBadge(fmt.Sprintf("%d/%d", onlineCount, deviceCount)).
		SetFocused(focused).
		SetPanelIndex(1)

	return r.SetContent(m.deviceList.View()).Render()
}

// renderEventsColumn renders the events column with embedded title.
func (m Model) renderEventsColumn(width, height int) string {
	orangeBorder := theme.Orange()
	focused := m.focusedPanel == PanelEvents

	// Build badge with status info (count, paused, filtered)
	badge := m.events.StatusBadge()

	// Build footer with keybindings
	footer := m.events.FooterText()

	// Scroll info goes in its own section (footerBadge)
	scrollInfo := m.events.ScrollInfo()

	r := rendering.New(width, height).
		SetTitle("Events").
		SetBadge(badge).
		SetFooter(footer).
		SetFooterBadge(scrollInfo).
		SetFocused(focused).
		SetPanelIndex(3).
		SetFocusColor(orangeBorder).
		SetBlurColor(orangeBorder)

	// Resize events model to fit inside the border box
	// Footer is embedded in border, not a separate line
	m.events = m.events.SetSize(r.ContentWidth(), r.ContentHeight())
	eventsView := m.events.View()

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
	labelStyle := lipgloss.NewStyle().Foreground(colors.Highlight)
	valueStyle := lipgloss.NewStyle().Foreground(colors.Text)
	titleStyle := lipgloss.NewStyle().Foreground(colors.Highlight).Bold(true)

	// Create metadata lines to match banner height (condensed layout)
	metaLines := make([]string, len(bannerLines))

	// Line 0: Title
	metaLines[0] = titleStyle.Render("Shelly Dashboard")

	// Line 1: Device counts (moved up, no blank line after title)
	if m.cache.IsLoading() {
		fetched := m.cache.FetchedCount()
		metaLines[1] = labelStyle.Render("Devices: ") + valueStyle.Render(fmt.Sprintf("Loading... %d/%d", fetched, total))
	} else {
		metaLines[1] = labelStyle.Render("Devices: ") +
			onlineStyle.Render(fmt.Sprintf("%d online", online))
		if offline > 0 {
			metaLines[1] += labelStyle.Render(" / ") + offlineStyle.Render(fmt.Sprintf("%d offline", offline))
		}
	}

	// Line 2: Power consumption
	if totalPower != 0 {
		powerStr := fmt.Sprintf("%.1fW", totalPower)
		if totalPower >= 1000 || totalPower <= -1000 {
			powerStr = fmt.Sprintf("%.2fkW", totalPower/1000)
		}
		metaLines[2] = labelStyle.Render("Power:   ") + powerStyle.Render(powerStr)
	} else {
		metaLines[2] = labelStyle.Render("Power:   ") + valueStyle.Render("--")
	}

	// Line 3: Theme (moved up, no blank line before)
	themeName := theme.CurrentThemeName()
	if themeName != "" {
		metaLines[3] = labelStyle.Render("Theme:   ") + valueStyle.Render(themeName)
	}

	// Line 4: Current filter (if set)
	if m.filter != "" {
		filterStyle := lipgloss.NewStyle().Foreground(colors.Highlight).Italic(true)
		metaLines[4] = labelStyle.Render("Filter:  ") + filterStyle.Render(m.filter)
	}

	// Fill remaining lines with empty strings
	for i := 5; i < len(metaLines); i++ {
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

// openDeviceBrowser opens the device's web UI in the default browser.
func (m Model) openDeviceBrowser(address string) tea.Cmd {
	return func() tea.Msg {
		url := "http://" + address
		b := browser.New()
		if err := b.Browse(m.ctx, url); err != nil {
			return DeviceActionMsg{
				Device: address,
				Action: "open browser",
				Err:    err,
			}
		}
		return DeviceActionMsg{
			Device: address,
			Action: "open browser",
			Err:    nil,
		}
	}
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

	// Add Switch methods if device has switches (each switch requires id parameter)
	for _, sw := range d.Switches {
		methods = append(methods,
			fmt.Sprintf("Switch.GetStatus?id=%d", sw.ID),
			fmt.Sprintf("Switch.GetConfig?id=%d", sw.ID))
	}

	// Add PM/EM methods only if device has those components (each requires id parameter)
	if d.Snapshot != nil {
		for _, pm := range d.Snapshot.PM {
			methods = append(methods, fmt.Sprintf("PM.GetStatus?id=%d", pm.ID))
		}
		for _, em := range d.Snapshot.EM {
			methods = append(methods, fmt.Sprintf("EM.GetStatus?id=%d", em.ID))
		}
		for _, em1 := range d.Snapshot.EM1 {
			methods = append(methods, fmt.Sprintf("EM1.GetStatus?id=%d", em1.ID))
		}
	}

	return methods
}
