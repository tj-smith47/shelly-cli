// Package tui provides the terminal user interface for the shelly CLI.
package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/branding"
	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	devmodel "github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cmdmode"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/confirm"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/control"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/devicedetail"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/deviceinfo"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/devicelist"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/energybars"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/energyhistory"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/events"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/help"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/jsonviewer"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/search"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/statusbar"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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

// Device action constants.
const (
	actionToggle = "toggle"
	actionOn     = "on"
	actionOff    = "off"
	actionReboot = "reboot"
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
	styles  Styles

	// View management (5-tab system)
	viewManager *views.Manager
	tabBar      tabs.Model

	// Shared device cache
	cache *cache.Cache

	// Shared event stream (WebSocket for Gen2+, polling for Gen1)
	eventStream *automation.EventStream

	// Unified focus and keybinding management
	focusState *focus.State
	contextMap *keys.ContextMap

	// State
	ready                   bool
	quitting                bool
	cursor                  int    // Selected device index
	componentCursor         int    // Selected component within device (-1 = all)
	filter                  string // Device name filter
	endpointCursor          int    // Selected endpoint in JSON panel
	initialSelectionEmitted bool   // Whether initial device selection has been emitted
	lastCacheVersion        uint64 // Last observed cache version for detecting WebSocket updates
	confirmingReboot        bool   // True when showing reboot confirmation
	pendingRebootDevice     string // Device name pending reboot

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
	energyHistory *energyhistory.Model
	jsonViewer    jsonviewer.Model
	confirm       confirm.Model
	deviceInfo    deviceinfo.Model
	deviceList    devicelist.Model
	deviceDetail  devicedetail.Model
	controlPanel  control.Panel

	// Debug logging (enabled by SHELLY_TUI_DEBUG=1)
	debugLogger *debug.Logger
}

// Options configures the TUI.
type Options struct {
	Filter string
}

// DefaultOptions returns default TUI options.
func DefaultOptions() Options {
	return Options{}
}

// applyTUITheme loads and applies the TUI-specific theme from configuration.
func applyTUITheme(cfg *config.Config, ios *iostreams.IOStreams) {
	if cfg == nil {
		return
	}
	tc := cfg.GetTUIThemeConfig()
	if tc == nil {
		return
	}

	if tc.File != "" {
		expanded := theme.ExpandPath(tc.File)
		data, err := afero.ReadFile(config.Fs(), expanded)
		if err != nil {
			ios.DebugErr("read tui theme file", err)
			return
		}
		if err := theme.ApplyThemeFromData(data, tc.Semantic); err != nil {
			ios.DebugErr("tui theme", err)
		}
		return
	}

	if err := theme.ApplyConfig(tc.Name, tc.Colors, tc.Semantic); err != nil {
		ios.DebugErr("tui theme", err)
	}
}

// getEventsConfig returns the TUI events config with defaults applied.
func getEventsConfig(cfg *config.Config) config.TUIEventsConfig {
	if cfg == nil {
		return config.DefaultTUIEventsConfig()
	}
	return cfg.GetTUIEventsConfig()
}

// New creates a new TUI application.
func New(ctx context.Context, f *cmdutil.Factory, opts Options) Model {
	cfg, err := f.Config()
	if err != nil {
		cfg = nil
	}

	applyTUITheme(cfg, f.IOStreams())

	// Create shared cache with FileCache for persistent static data caching
	deviceCache := cache.New(ctx, f.ShellyService(), f.IOStreams(), f.FileCache())

	// Create shared event stream (WebSocket for Gen2+, polling for Gen1)
	eventStream := automation.NewEventStream(f.ShellyService())

	// Subscribe cache to event stream for real-time updates
	deviceCache.SubscribeToEvents(eventStream)

	// Create focus state - single source of truth for all focus management
	focusStateInst := focus.NewState()

	// Create context-sensitive keybinding map
	contextKeyMap := keys.NewContextMap()

	// Create search component with initial filter
	searchModel := search.NewWithFilter(opts.Filter)

	// Create events component for real-time event stream with config
	eventsConfig := getEventsConfig(cfg)
	eventsModel := events.New(events.Deps{
		Ctx:                ctx,
		Svc:                f.ShellyService(),
		IOS:                f.IOStreams(),
		EventStream:        eventStream,
		FilteredEvents:     eventsConfig.FilteredEvents,
		FilteredComponents: eventsConfig.FilteredComponents,
		MaxItems:           eventsConfig.MaxItems,
	})

	// Create energy bars component
	energyBarsModel := energybars.New(deviceCache)

	// Create energy history sparklines component
	ehm := energyhistory.New(deviceCache)
	energyHistoryModel := &ehm

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

	// Create control panel overlay component
	controlSvc := control.NewServiceAdapter(f.ShellyService())
	controlPanelModel := control.NewPanel(ctx, controlSvc, controlSvc)

	// Create view manager and register all views
	vm := views.New()

	// Register Dashboard view (delegates rendering to app.go)
	dashboard := views.NewDashboard(views.DashboardDeps{Ctx: ctx})
	vm.Register(dashboard)

	// Register Automation view
	automationView := views.NewAutomation(views.AutomationDeps{
		Ctx:         ctx,
		Svc:         f.ShellyService(),
		AutoSvc:     f.AutomationService(),
		KVSSvc:      f.KVSService(),
		EventStream: eventStream,
		FocusState:  focusStateInst,
	})
	vm.Register(automationView)

	// Register Config view
	configView := views.NewConfig(views.ConfigDeps{
		Ctx:        ctx,
		Svc:        f.ShellyService(),
		FocusState: focusStateInst,
	})
	vm.Register(configView)

	// Register Manage view
	manage := views.NewManage(views.ManageDeps{
		Ctx:        ctx,
		Svc:        f.ShellyService(),
		FileCache:  f.FileCache(),
		FocusState: focusStateInst,
	})
	vm.Register(manage)

	// Register Monitor view
	monitorView := views.NewMonitor(views.MonitorDeps{
		Ctx:         ctx,
		Svc:         f.ShellyService(),
		IOS:         f.IOStreams(),
		EventStream: eventStream,
		FocusState:  focusStateInst,
	})
	vm.Register(monitorView)

	// Register Fleet view
	fleet := views.NewFleet(views.FleetDeps{
		Ctx:        ctx,
		Svc:        f.ShellyService(),
		IOS:        f.IOStreams(),
		Cfg:        cfg,
		FocusState: focusStateInst,
	})
	vm.Register(fleet)

	// Create tab bar
	tabBar := tabs.New()

	m := Model{
		ctx:             ctx,
		factory:         f,
		cfg:             cfg,
		styles:          DefaultStyles(),
		viewManager:     vm,
		tabBar:          tabBar,
		cache:           deviceCache,
		eventStream:     eventStream,
		focusState:      focusStateInst,
		contextMap:      contextKeyMap,
		filter:          opts.Filter,
		componentCursor: -1, // -1 means "all components"
		statusBar:       statusbar.New(),
		search:          searchModel,
		cmdMode:         cmdmode.New(),
		help:            help.New(),
		toast:           toast.New(),
		events:          eventsModel,
		energyBars:      energyBarsModel,
		energyHistory:   energyHistoryModel,
		jsonViewer:      jsonViewerModel,
		confirm:         confirm.New(confirm.WithModalOverlay()),
		deviceInfo:      deviceInfoModel,
		deviceList:      deviceListModel,
		deviceDetail:    deviceDetailModel,
		controlPanel:    controlPanelModel,
		debugLogger:     debug.New(), // nil if SHELLY_TUI_DEBUG not set
	}

	// Set global debug logger for trace logging from components
	debug.SetGlobal(m.debugLogger)

	// Initialize statusbar debug indicator based on logger state
	if m.debugLogger.Enabled() {
		m.statusBar = m.statusBar.SetDebugActive(true)
	}

	return m
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
// Note: EventStream is started AFTER initial wave loading completes (in handleAllDevicesLoaded)
// to avoid concurrent HTTP requests during startup which trip the circuit breaker.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.cache.Init(),
		m.statusBar.Init(),
		m.toast.Init(),
		m.events.Init(),
		m.viewManager.Init(),
		m.energyBars.Init(),
		m.energyHistory.Init(),
		m.deviceList.Init(),
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
		if model, ok := newModel.(Model); ok {
			return model.syncCacheVersion(), cmd
		}
		return newModel, cmd
	}

	// Forward and update components
	newModel, cmds := m.updateComponents(msg)
	return newModel.syncCacheVersion(), tea.Batch(cmds...)
}

// syncCacheVersion checks if cache version changed and updates status bar if needed.
// This ensures WebSocket updates (which modify cache directly) refresh the status bar.
func (m Model) syncCacheVersion() Model {
	currentVersion := m.cache.Version()
	if currentVersion != m.lastCacheVersion {
		m.lastCacheVersion = currentVersion
		m = m.updateStatusBarContext()
	}
	return m
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
	case EventStreamStartedMsg, EventStreamErrorMsg:
		return m.handleEventStreamMsg(msg)
	case events.EventMsg, events.SubscriptionStatusMsg, events.RefreshTickMsg:
		var cmd tea.Cmd
		m.events, cmd = m.events.Update(msg)
		return m, cmd, true
	case search.FilterChangedMsg:
		return m.handleFilterChanged(msg)
	case search.ClosedMsg, cmdmode.ClosedMsg, confirm.CancelledMsg, messages.ModalClosedMsg:
		return m.handleOverlayClosed(msg)
	case cmdmode.CommandMsg:
		newModel, cmd := m.handleCommand(msg)
		return newModel, cmd, true
	case cmdmode.ErrorMsg:
		return m, toast.Error(msg.Message), true
	case messages.ModalOpenedMsg:
		m.focusState.PushOverlay(msg.ID, msg.Mode)
		return m, nil, true
	default:
		return m.handleViewAndComponentMsgs(msg)
	}
}

func (m Model) handleEventStreamMsg(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case EventStreamStartedMsg:
		m.factory.IOStreams().Debug("event stream started successfully")
	case EventStreamErrorMsg:
		m.factory.IOStreams().DebugErr("event stream start", msg.Err)
		return m, toast.Error("Event stream failed: " + msg.Err.Error()), true
	}
	return m, nil, true
}

func (m Model) handleFilterChanged(msg search.FilterChangedMsg) (tea.Model, tea.Cmd, bool) {
	m.filter = msg.Filter
	m.cursor = 0
	m.deviceList = m.deviceList.SetFilter(msg.Filter)
	return m, nil, true
}

func (m Model) handleOverlayClosed(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg.(type) {
	case search.ClosedMsg, cmdmode.ClosedMsg:
		m = m.updateViewManagerSize()
		m.focusState.PopOverlay()
	case messages.ModalClosedMsg:
		m.focusState.PopOverlay()
	case confirm.CancelledMsg:
		// No additional action needed
	}
	return m, nil, true
}

// handleViewAndComponentMsgs handles view-related and component messages.
func (m Model) handleViewAndComponentMsgs(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	// Handle device selection messages
	if model, cmd, handled := m.handleDeviceSelectionMsgs(msg); handled {
		return model, cmd, true
	}
	// Handle component/viewer messages
	if model, cmd, handled := m.handleComponentMsgs(msg); handled {
		return model, cmd, true
	}
	// Handle tab/view navigation
	switch msg := msg.(type) {
	case tabs.TabChangedMsg:
		return m, m.viewManager.SetActive(msg.Current), true
	case views.ViewChangedMsg:
		m.tabBar, _ = m.tabBar.SetActive(msg.Current)
		return m, nil, true
	}
	// Route non-key messages to visible overlays (e.g., spinner ticks for device detail)
	if _, isKey := msg.(tea.KeyPressMsg); !isKey {
		if model, cmd, handled := m.routeToVisibleOverlay(msg); handled {
			return model, cmd, true
		}
	}
	return m, nil, false
}

func (m Model) handleDeviceSelectionMsgs(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case views.DeviceSelectedMsg:
		debug.TraceEvent("DeviceSelectedMsg(views): device=%s (debounced)", msg.Device)
		// Sync device info panel with new selection
		m = m.syncDeviceInfo()
		// Use debounced cache fetch to prevent overwhelming devices during rapid scrolling
		return m, tea.Batch(
			m.viewManager.PropagateDevice(msg.Device),
			m.cache.SetFocusedDevice(msg.Device),
		), true
	case devicelist.DeviceSelectedMsg:
		m.cursor = m.deviceList.Cursor()
		debug.TraceEvent("DeviceSelectedMsg(devicelist): device=%s (debounced)", msg.Name)
		// Sync device info panel with new selection
		m = m.syncDeviceInfo()
		// Use debounced cache fetch to prevent overwhelming devices during rapid scrolling
		return m, tea.Batch(
			m.viewManager.PropagateDevice(msg.Name),
			m.cache.SetFocusedDevice(msg.Name),
		), true
	case cache.ExtendedStatusDebounceMsg:
		debug.TraceEvent("ExtendedStatusDebounceMsg: triggering FetchExtendedStatus for %s", msg.Name)
		return m, m.cache.FetchExtendedStatus(msg.Name), true
	case cache.FetchExtendedStatusMsg:
		wifiOK := msg.WiFi != nil
		sysOK := msg.Sys != nil
		debug.TraceEvent("FetchExtendedStatusMsg for %s: wifi=%v sys=%v", msg.Name, wifiOK, sysOK)
		if wifiOK {
			debug.TraceEvent("WiFi data: SSID=%s RSSI=%d", msg.WiFi.SSID, msg.WiFi.RSSI)
		}
		if sysOK {
			debug.TraceEvent("Sys data: Uptime=%d RAM=%d/%d", msg.Sys.Uptime, msg.Sys.RAMFree, msg.Sys.RAMSize)
		}
		m.cache.HandleExtendedStatus(msg)
		m = m.syncDeviceInfo()
		return m, nil, true
	case devicelist.OpenBrowserMsg:
		return m, m.openDeviceBrowser(msg.Address), true
	}
	return m, nil, false
}

func (m Model) handleComponentMsgs(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case jsonviewer.CloseMsg:
		m = m.setFocus(focus.PanelDashboardInfo)
		m.focusState.PopOverlay()
		return m, nil, true
	case jsonviewer.FetchedMsg:
		var cmd tea.Cmd
		m.jsonViewer, cmd = m.jsonViewer.Update(msg)
		return m, cmd, true
	case confirm.ConfirmedMsg:
		return m, toast.Success("Action confirmed: " + msg.Operation), true
	case deviceinfo.RequestJSONMsg:
		return m.handleRequestJSON(msg)
	case deviceinfo.RequestToggleMsg:
		return m.handleRequestToggle(msg)
	case devicedetail.ClosedMsg:
		m.deviceDetail, _ = m.deviceDetail.Update(msg)
		m.focusState.PopOverlay()
		return m, nil, true
	case devicedetail.Msg:
		var cmd tea.Cmd
		m.deviceDetail, cmd = m.deviceDetail.Update(msg)
		return m, cmd, true
	case control.PanelCloseMsg:
		m.controlPanel = m.controlPanel.Hide()
		m.focusState.PopOverlay()
		return m, nil, true
	case control.ActionMsg:
		var cmd tea.Cmd
		m.controlPanel, cmd = m.controlPanel.Update(msg)
		return m, cmd, true
	}
	return m, nil, false
}

// handleDeviceUpdate handles cache.DeviceUpdateMsg by forwarding to cache and energy history.
func (m Model) handleDeviceUpdate(msg cache.DeviceUpdateMsg) (tea.Model, tea.Cmd, bool) {
	// Forward to cache AND energyHistory (to record power history)
	cacheCmd := m.cache.Update(msg)
	var historyCmd tea.Cmd
	*m.energyHistory, historyCmd = m.energyHistory.Update(msg)
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
// Also starts the EventStream now that wave loading is complete - this prevents
// concurrent HTTP requests during startup which trip the circuit breaker.
func (m Model) handleAllDevicesLoaded(msg cache.AllDevicesLoadedMsg) (tea.Model, tea.Cmd, bool) {
	cmd := m.cache.Update(msg)
	m = m.updateStatusBarContext()

	// Forward to energyHistory to trigger initial data collection
	var historyCmd tea.Cmd
	*m.energyHistory, historyCmd = m.energyHistory.Update(msg)
	cmd = tea.Batch(cmd, historyCmd)

	// Start EventStream now that initial load is complete
	// This avoids concurrent HTTP requests with wave loading
	eventStreamCmd := m.startEventStream()

	devices := m.getFilteredDevices()
	if len(devices) == 0 || m.cursor >= len(devices) {
		return m, tea.Batch(cmd, eventStreamCmd), true
	}
	d := devices[m.cursor]
	return m, tea.Batch(cmd, eventStreamCmd, func() tea.Msg {
		return views.DeviceSelectedMsg{
			Device:  d.Device.Name,
			Address: d.Device.Address,
		}
	}), true
}

// handleRequestToggle toggles a specific component from the deviceinfo panel.
func (m Model) handleRequestToggle(msg deviceinfo.RequestToggleMsg) (tea.Model, tea.Cmd, bool) {
	// Find the device
	data := m.cache.GetDevice(msg.DeviceName)
	if data == nil || !data.Online {
		return m, nil, true
	}

	svc := m.factory.ShellyService()
	compID := msg.ComponentID

	return m, func() tea.Msg {
		var err error
		ctx := m.ctx

		// Plugin devices use PluginControl for all component types
		if data.Device.IsPluginManaged() {
			_, err = svc.PluginControl(ctx, data.Device.Address, "toggle", msg.ComponentType, compID)
		} else {
			switch msg.ComponentType {
			case "switch":
				_, err = svc.SwitchToggle(ctx, data.Device.Address, compID)
			case "light":
				_, err = svc.LightToggle(ctx, data.Device.Address, compID)
			case "cover":
				// Cover toggle uses stop/open logic
				_, err = svc.QuickToggle(ctx, data.Device.Address, &compID)
			}
		}

		return DeviceActionMsg{
			Device: msg.DeviceName,
			Action: fmt.Sprintf("toggle %s:%d", msg.ComponentType, compID),
			Err:    err,
		}
	}, true
}

// handleRequestJSON opens the JSON viewer for a requested endpoint.
func (m Model) handleRequestJSON(msg deviceinfo.RequestJSONMsg) (tea.Model, tea.Cmd, bool) {
	devices := m.getFilteredDevices()
	for _, d := range devices {
		if d.Device.Name != msg.DeviceName {
			continue
		}
		m = m.setFocus(focus.PanelDashboardJSON)
		endpoints := m.getDeviceMethods(d)
		var cmd tea.Cmd
		m.jsonViewer, cmd = m.jsonViewer.Open(d.Device.Address, msg.Endpoint, endpoints)
		m.focusState.PushOverlay(focus.OverlayJSONViewer, focus.ModeOverlay)
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

	// Update energy panels for spinner animation (they need tick messages)
	var energyBarsCmd, energyHistoryCmd tea.Cmd
	m.energyBars, energyBarsCmd = m.energyBars.Update(msg)
	*m.energyHistory, energyHistoryCmd = m.energyHistory.Update(msg)
	cmds = append(cmds, energyBarsCmd, energyHistoryCmd)

	// Update deviceList for spinner animation during loading
	var deviceListCmd tea.Cmd
	m.deviceList, deviceListCmd = m.deviceList.Update(msg)
	cmds = append(cmds, deviceListCmd)

	// Forward non-key messages to ALL views so async results can be processed
	// (e.g., Config's wifi.StatusLoadedMsg needs to reach Config even if Dashboard is active)
	if _, isKey := msg.(tea.KeyPressMsg); !isKey {
		viewCmd := m.viewManager.UpdateAll(msg)
		cmds = append(cmds, viewCmd)
	}

	return m, cmds
}

// handleKeyPressMsg handles keyboard input using the FocusState system.
// It routes keys based on the current focus mode (normal/overlay/modal/input).
func (m Model) handleKeyPressMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	// Handle reboot confirmation
	if m.confirmingReboot {
		return m.handleRebootConfirmation(msg)
	}

	// STEP 1: Determine effective mode from focusState or visibility checks
	effectiveMode := m.getEffectiveFocusMode()

	// STEP 2: Get context and convert key to action
	ctx := m.getContextForMode(effectiveMode)
	action := m.contextMap.Match(ctx, msg)

	// STEP 3: Route based on focus mode
	switch effectiveMode {
	case focus.ModeInput:
		return m.handleInputModeKey(msg, action)
	case focus.ModeOverlay:
		return m.handleOverlayModeKey(msg, action)
	case focus.ModeModal:
		return m.handleModalModeKey(msg, action)
	default:
		return m.handleNormalModeKey(msg, action)
	}
}

// getEffectiveFocusMode determines the current focus mode.
// If focusState has overlays, use its mode. Otherwise, check component visibility.
func (m Model) getEffectiveFocusMode() focus.Mode {
	// If focusState has registered overlays, use its mode
	if m.focusState.HasOverlay() {
		return m.focusState.Mode()
	}

	// Check visibility of components that don't yet push to focusState
	// Input mode takes priority
	if m.cmdMode.IsActive() || m.search.IsActive() {
		return focus.ModeInput
	}

	// Then check for modals (blocking)
	if m.viewManager.HasActiveModal() || m.confirm.Visible() {
		return focus.ModeModal
	}

	// Then check for overlays (non-blocking)
	if m.help.Visible() || m.jsonViewer.Visible() || m.deviceDetail.Visible() || m.controlPanel.Visible() {
		return focus.ModeOverlay
	}

	return focus.ModeNormal
}

// handleRebootConfirmation handles Y/N key presses for reboot confirmation.
func (m Model) handleRebootConfirmation(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "y", "Y":
		m.confirmingReboot = false
		deviceName := m.pendingRebootDevice
		m.pendingRebootDevice = ""
		return m, m.executeReboot(deviceName), true
	case "n", "N", keyconst.KeyEsc:
		m.confirmingReboot = false
		m.pendingRebootDevice = ""
		return m, nil, true
	}
	return m, nil, true
}

// executeReboot executes reboot on the named device.
func (m Model) executeReboot(deviceName string) tea.Cmd {
	cfg := config.Get()
	if cfg == nil {
		return toast.Error("Config not loaded")
	}
	device, ok := cfg.Devices[deviceName]
	if !ok {
		return toast.Error("Device not found: " + deviceName)
	}
	svc := m.factory.ShellyService()
	return func() tea.Msg {
		err := svc.DeviceReboot(m.ctx, device.Address, 0)
		return DeviceActionMsg{
			Device: deviceName,
			Action: actionReboot,
			Err:    err,
		}
	}
}

// getContextForMode returns the appropriate key context for the focus mode.
func (m Model) getContextForMode(mode focus.Mode) keys.Context {
	switch mode {
	case focus.ModeInput:
		return keys.ContextInput
	case focus.ModeModal:
		return keys.ContextModal
	case focus.ModeOverlay:
		// Use specific context if focusState knows the overlay
		if m.focusState.HasOverlay() {
			switch m.focusState.TopOverlay() {
			case focus.OverlayHelp:
				return keys.ContextHelp
			case focus.OverlayJSONViewer:
				return keys.ContextJSON
			case focus.OverlayNone, focus.OverlayConfirm, focus.OverlayDeviceDetail,
				focus.OverlayControlPanel, focus.OverlayEditModal, focus.OverlayCmdMode, focus.OverlaySearch,
				focus.OverlayScriptEditor, focus.OverlayScriptCreate, focus.OverlayScriptConsole,
				focus.OverlayScriptEval, focus.OverlayScheduleCreate, focus.OverlayAlertForm,
				focus.OverlayTemplateSelect, focus.OverlayProvisioning, focus.OverlayMigration:
				// Use global context for these (they don't have specific key contexts)
			}
		}
		// Fallback based on visible component
		if m.help.Visible() {
			return keys.ContextHelp
		}
		if m.jsonViewer.Visible() {
			return keys.ContextJSON
		}
		return keys.ContextGlobal
	default:
		return m.getCurrentKeyContext()
	}
}

// handleInputModeKey handles keys when in input mode (search, command mode).
func (m Model) handleInputModeKey(msg tea.KeyPressMsg, _ keys.Action) (tea.Model, tea.Cmd, bool) {
	// Command mode gets all keys
	if m.cmdMode.IsActive() {
		var cmd tea.Cmd
		m.cmdMode, cmd = m.cmdMode.Update(msg)
		return m, cmd, true
	}

	// Search gets all keys
	if m.search.IsActive() {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd, true
	}

	return m, nil, false
}

// handleOverlayModeKey handles keys when an overlay is visible (help, JSON viewer, etc.).
func (m Model) handleOverlayModeKey(msg tea.KeyPressMsg, action keys.Action) (tea.Model, tea.Cmd, bool) {
	// When help is searching, route all keys (including Escape) to the help overlay
	// so the search input can handle them (Escape exits search, not the overlay)
	if m.help.Visible() && m.help.Searching() {
		return m.routeToVisibleOverlay(msg)
	}

	// Handle escape to close overlay
	if action == keys.ActionEscape {
		return m.closeTopOverlay()
	}

	// Route to the visible overlay
	return m.routeToVisibleOverlay(msg)
}

// handleModalModeKey handles keys when a modal dialog is active.
// All keys pass through raw to the modal - the modal handles its own field
// navigation (Tab/Shift+Tab) and arrow keys for cursor movement in text inputs.
func (m Model) handleModalModeKey(msg tea.KeyPressMsg, action keys.Action) (tea.Model, tea.Cmd, bool) {
	// For modals managed by focusState, handle escape at app level
	if action == keys.ActionEscape && m.focusState.HasOverlay() {
		return m.closeTopOverlay()
	}

	// Pass all keys (including escape) raw to the modal
	// ViewManager modals handle their own escape key
	return m.routeToVisibleOverlay(msg)
}

// handleNormalModeKey handles keys when no overlay/modal is active.
func (m Model) handleNormalModeKey(msg tea.KeyPressMsg, _ keys.Action) (tea.Model, tea.Cmd, bool) {
	return m.handleKeyPress(msg)
}

// routeToVisibleOverlay routes a message to the currently visible overlay/modal.
func (m Model) routeToVisibleOverlay(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	// Check focusState first
	if m.focusState.HasOverlay() {
		switch m.focusState.TopOverlay() {
		case focus.OverlayHelp:
			var cmd tea.Cmd
			m.help, cmd = m.help.Update(msg)
			return m, cmd, true
		case focus.OverlayJSONViewer:
			var cmd tea.Cmd
			m.jsonViewer, cmd = m.jsonViewer.Update(msg)
			return m, cmd, true
		case focus.OverlayConfirm:
			var cmd tea.Cmd
			m.confirm, cmd = m.confirm.Update(msg)
			return m, cmd, true
		case focus.OverlayDeviceDetail:
			var cmd tea.Cmd
			m.deviceDetail, cmd = m.deviceDetail.Update(msg)
			return m, cmd, true
		case focus.OverlayControlPanel:
			var cmd tea.Cmd
			m.controlPanel, cmd = m.controlPanel.Update(msg)
			return m, cmd, true
		case focus.OverlayEditModal, focus.OverlayScriptEditor, focus.OverlayScriptCreate,
			focus.OverlayScriptConsole, focus.OverlayScriptEval, focus.OverlayScheduleCreate,
			focus.OverlayAlertForm, focus.OverlayTemplateSelect, focus.OverlayProvisioning,
			focus.OverlayMigration:
			// Route these to the view manager
			cmd := m.viewManager.Update(msg)
			return m, cmd, true
		case focus.OverlayNone, focus.OverlayCmdMode, focus.OverlaySearch:
			// These are handled by input mode, fall through to visibility check
		}
	}

	// Fallback: route based on visibility
	if m.help.Visible() {
		var cmd tea.Cmd
		m.help, cmd = m.help.Update(msg)
		return m, cmd, true
	}
	if m.jsonViewer.Visible() {
		var cmd tea.Cmd
		m.jsonViewer, cmd = m.jsonViewer.Update(msg)
		return m, cmd, true
	}
	if m.confirm.Visible() {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd, true
	}
	if m.deviceDetail.Visible() {
		var cmd tea.Cmd
		m.deviceDetail, cmd = m.deviceDetail.Update(msg)
		return m, cmd, true
	}
	if m.controlPanel.Visible() {
		var cmd tea.Cmd
		m.controlPanel, cmd = m.controlPanel.Update(msg)
		return m, cmd, true
	}
	if m.viewManager.HasActiveModal() {
		cmd := m.viewManager.Update(msg)
		return m, cmd, true
	}

	return m, nil, false
}

// closeTopOverlay closes the topmost overlay or modal.
func (m Model) closeTopOverlay() (tea.Model, tea.Cmd, bool) {
	// If focusState has overlays, pop from stack
	if m.focusState.HasOverlay() {
		id := m.focusState.PopOverlay()
		switch id {
		case focus.OverlayHelp:
			m.help = m.help.Hide()
		case focus.OverlayJSONViewer:
			m.jsonViewer = m.jsonViewer.Close()
		case focus.OverlayConfirm:
			m.confirm = m.confirm.Hide()
		case focus.OverlayDeviceDetail:
			m.deviceDetail = m.deviceDetail.Hide()
		case focus.OverlayControlPanel:
			m.controlPanel = m.controlPanel.Hide()
		case focus.OverlayNone:
			// Nothing to close
			return m, nil, false
		case focus.OverlayEditModal, focus.OverlayCmdMode, focus.OverlaySearch,
			focus.OverlayScriptEditor, focus.OverlayScriptCreate, focus.OverlayScriptConsole,
			focus.OverlayScriptEval, focus.OverlayScheduleCreate, focus.OverlayAlertForm,
			focus.OverlayTemplateSelect, focus.OverlayProvisioning, focus.OverlayMigration:
			// These are managed by their components via escape key handling
		}
		return m, nil, true
	}

	// Fallback: close based on visibility (priority order)
	if m.help.Visible() {
		m.help = m.help.Hide()
		return m, nil, true
	}
	if m.jsonViewer.Visible() {
		m.jsonViewer = m.jsonViewer.Close()
		return m, nil, true
	}
	if m.confirm.Visible() {
		m.confirm = m.confirm.Hide()
		return m, nil, true
	}
	if m.deviceDetail.Visible() {
		m.deviceDetail = m.deviceDetail.Hide()
		return m, nil, true
	}
	if m.controlPanel.Visible() {
		m.controlPanel = m.controlPanel.Hide()
		return m, nil, true
	}

	return m, nil, false
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
// based on current input bar state (search, command mode, or toast).
func (m Model) updateViewManagerSize() Model {
	tabBarHeight := 1
	footerHeight := 2
	inputHeight := 0
	if m.search.IsActive() || m.cmdMode.IsActive() || m.toast.HasToasts() {
		inputHeight = 3 // Input bar has top border, content, bottom border
	}
	contentHeight := m.height - branding.BannerHeight() - tabBarHeight - footerHeight - inputHeight
	m.viewManager.SetSize(m.width, contentHeight)
	return m
}

// handleDeviceAction handles device action results.
func (m Model) handleDeviceAction(msg DeviceActionMsg) (tea.Model, tea.Cmd) {
	var toastCmd tea.Cmd
	var eventLevel, eventDesc string

	if msg.Err != nil {
		toastCmd = toast.Error(msg.Device + ": " + msg.Action + " failed - " + msg.Err.Error())
		eventLevel = "error"
		eventDesc = msg.Action + " failed: " + msg.Err.Error()
	} else {
		toastCmd = toast.Success(msg.Device + ": " + msg.Action + " success")
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

	cmds := []tea.Cmd{toastCmd, evtCmd}

	// On successful state-changing actions, trigger a cache refresh to update UI
	// This ensures the UI reflects the new state even if WebSocket notification is delayed
	if msg.Err == nil && isStateChangingAction(msg.Action) {
		cmds = append(cmds, func() tea.Msg {
			return cache.DeviceRefreshMsg{Name: msg.Device}
		})
	}

	return m, tea.Batch(cmds...)
}

// isStateChangingAction returns true for actions that change device state.
func isStateChangingAction(action string) bool {
	switch action {
	case actionToggle, actionOn, actionOff:
		return true
	default:
		return false
	}
}

// updateStatusBarContext updates the status bar with context-specific items.
func (m Model) updateStatusBarContext() Model {
	// Get all component counts in a single lock acquisition
	counts := m.cache.ComponentCounts()

	m.statusBar = m.statusBar.SetComponentCounts(statusbar.ComponentCounts{
		SwitchesOn:   counts.SwitchesOn,
		SwitchesOff:  counts.SwitchesOff,
		LightsOn:     counts.LightsOn,
		LightsOff:    counts.LightsOff,
		CoversOpen:   counts.CoversOpen,
		CoversClosed: counts.CoversClosed,
		CoversMoving: counts.CoversMoving,
	})

	// Set view-specific context based on active tab
	activeTab := m.tabBar.ActiveTabID()
	panelName := m.focusedPanelName()

	switch activeTab {
	case tabs.TabDashboard, tabs.TabAutomation, tabs.TabConfig:
		// Device-based views: show selected device info
		devices := m.getFilteredDevices()
		if len(devices) > 0 && m.cursor >= 0 && m.cursor < len(devices) {
			d := devices[m.cursor]
			m.statusBar = m.statusBar.SetDeviceContext(d.Device.DisplayName(), d.Device.Address, panelName)
		} else {
			m.statusBar = m.statusBar.ClearContext()
		}

	case tabs.TabMonitor:
		// Monitor view: show WebSocket connection status and refresh interval
		wsConnected, wsTotal := m.events.ConnectionCounts()
		refreshInterval := 5 * time.Second // Events refresh interval
		m.statusBar = m.statusBar.SetMonitorContext(wsConnected, wsTotal, refreshInterval, panelName)

	case tabs.TabManage:
		// Manage view: show firmware update count from firmware component
		firmwareUpdates := 0
		if manageView, ok := m.viewManager.Get(views.ViewManage).(*views.Manage); ok {
			firmwareUpdates = manageView.Firmware().UpdateCount()
		}
		m.statusBar = m.statusBar.SetManageContext(firmwareUpdates, panelName)

	case tabs.TabFleet:
		// Fleet view: show selected group from groups model
		groupName := ""
		if fleetView, ok := m.viewManager.Get(views.ViewFleet).(*views.Fleet); ok {
			if group := fleetView.Groups().SelectedGroup(); group != nil {
				groupName = group.Name
			}
		}
		m.statusBar = m.statusBar.SetFleetContext(groupName, panelName)

	default:
		m.statusBar = m.statusBar.ClearContext()
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
			return m, toast.Success("Filter cleared")
		}
		return m, toast.Success("Filter: " + msg.Args)

	case cmdmode.CmdTheme:
		if !theme.SetTheme(msg.Args) {
			return m, toast.Error("Invalid theme: " + msg.Args)
		}
		m.styles = DefaultStyles()
		return m, toast.Success("Theme: " + msg.Args)

	case cmdmode.CmdView:
		// Views are collapsed - just acknowledge command
		return m, toast.Info("Single unified view")

	case cmdmode.CmdHelp:
		m.help = m.help.SetSize(m.width, m.height)
		m.help = m.help.SetContext(m.getHelpContext())
		m.help = m.help.Toggle()
		return m, nil

	case cmdmode.CmdToggle:
		if cmd := m.executeDeviceAction(actionToggle); cmd != nil {
			return m, cmd
		}
		return m, toast.Error("No device selected or device offline")

	default:
		return m, toast.Error("Unknown command")
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
	case tabs.TabMonitor:
		return keys.ContextMonitor
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

	// Handle Tab/Shift+Tab at app level for ALL tabs - this is the unified focus system
	keyStr := msg.String()
	if keyStr == keyconst.KeyTab || keyStr == keyconst.KeyShiftTab {
		prevTab := m.focusState.ActiveTab()
		prevPanel := m.focusState.ActivePanel()
		if keyStr == keyconst.KeyTab {
			m.focusState.NextPanel()
		} else {
			m.focusState.PrevPanel()
		}
		m = m.syncComponentFocus()
		cmd := m.focusState.EmitChanged(prevTab, prevPanel, false, true, false)
		return m, cmd, true
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

	// Panel switching with Tab/Shift+Tab (already handled above, this is fallback)
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
// When device list is focused, always returns ContextDevices so device actions (d, c, t, etc.) work.
func (m Model) getCurrentKeyContext() keys.Context {
	// When device list is focused on any tab with device list, use Devices context
	// This ensures device actions (detail, control, toggle, etc.) work regardless of tab
	if m.hasDeviceList() && m.focusState.IsPanelFocused(focus.PanelDeviceList) {
		return keys.ContextDevices
	}

	// Otherwise use tab/panel-specific context
	switch m.tabBar.ActiveTabID() {
	case tabs.TabDashboard:
		// Dashboard uses panel-based context
		return m.panelToContext(m.focusState.ActivePanel())
	case tabs.TabAutomation:
		return keys.ContextAutomation
	case tabs.TabConfig:
		return keys.ContextConfig
	case tabs.TabManage:
		return keys.ContextManage
	case tabs.TabFleet:
		return keys.ContextFleet
	case tabs.TabMonitor:
		return keys.ContextMonitor
	default:
		return m.panelToContext(m.focusState.ActivePanel())
	}
}

// panelToContext converts a focus.GlobalPanelID to a keys.Context.
func (m Model) panelToContext(panel focus.GlobalPanelID) keys.Context {
	switch panel {
	case focus.PanelDashboardEvents:
		return keys.ContextEvents
	case focus.PanelDeviceList:
		return keys.ContextDevices
	case focus.PanelDashboardInfo:
		return keys.ContextInfo
	case focus.PanelDashboardJSON:
		return keys.ContextJSON
	case focus.PanelDashboardEnergyBars, focus.PanelDashboardEnergyHistory:
		return keys.ContextEnergy
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
		m = m.setFocus(focus.PanelDeviceList)
		return m, nil
	}

	// Escape ALWAYS returns focus to device list
	if keyStr == "esc" {
		m = m.setFocus(focus.PanelDeviceList)
		return m, nil
	}

	// When device list is focused, try to handle navigation keys here
	if m.focusState.IsPanelFocused(focus.PanelDeviceList) {
		newModel, cmd, handled := m.handleDeviceListFocusedKeys(msg)
		if handled {
			return newModel, cmd
		}
		// Non-navigation keys (Enter, e, d, etc.) should go to the view
		// so actions like "edit script" work even when device list is focused
		cmd = m.viewManager.Update(msg)
		return m, cmd
	}

	// When focused on view (PanelDetail), forward non-Tab keys to the view
	// Tab/Shift+Tab is handled in handleKeyPress before this is reached
	cmd := m.viewManager.Update(msg)
	return m, cmd
}

// handleDeviceListFocusedKeys handles keys when device list panel is focused.
// Tab/Shift+Tab is handled in handleKeyPress before this is reached.
func (m Model) handleDeviceListFocusedKeys(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	keyStr := msg.String()
	// Shift+1 (!) keeps focus on device list (panel 1 on all tabs with device list)
	if keyStr == keyconst.Shift1 {
		m = m.setFocus(focus.PanelDeviceList)
		return m, nil, true
	}
	// Shift+2-9 (@, #, $, %, ^, &, *, () switch focus to view and forward the key for panel selection
	if isShiftNumberKey(keyStr) && keyStr != keyconst.Shift1 {
		m = m.setFocus(focus.PanelDashboardInfo)
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

// handlePanelSwitch handles Shift+N for direct jump and Enter for JSON overlay.
// Tab/Shift+Tab is handled in handleKeyPress before tab-specific routing.
func (m Model) handlePanelSwitch(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch msg.String() {
	case keyconst.Shift1:
		m = m.setFocus(focus.PanelDeviceList)
		return m, nil, true
	case keyconst.Shift2:
		m = m.setFocus(focus.PanelDashboardInfo)
		return m, nil, true
	case keyconst.Shift3:
		m = m.setFocus(focus.PanelDashboardEvents)
		return m, nil, true
	case keyconst.Shift4:
		m = m.setFocus(focus.PanelDashboardEnergyBars)
		return m, nil, true
	case keyconst.Shift5:
		m = m.setFocus(focus.PanelDashboardEnergyHistory)
		return m, nil, true
	case keyconst.KeyEnter:
		return m.openJSONViewer()
	case keyconst.KeyEsc:
		if m.jsonViewer.Visible() {
			m.jsonViewer = m.jsonViewer.Close()
			m = m.setFocus(focus.PanelDashboardInfo)
			return m, nil, true
		}
	}
	return m, nil, false
}

// syncComponentFocus updates all component focus states to match focusState.
func (m Model) syncComponentFocus() Model {
	current := m.focusState.ActivePanel()
	m.deviceInfo = m.deviceInfo.SetFocused(current == focus.PanelDashboardInfo)
	m.deviceList = m.deviceList.SetFocused(current == focus.PanelDeviceList)
	m.events = m.events.SetFocused(current == focus.PanelDashboardEvents)
	m.energyBars = m.energyBars.SetFocused(current == focus.PanelDashboardEnergyBars)
	*m.energyHistory = m.energyHistory.SetFocused(current == focus.PanelDashboardEnergyHistory)
	return m
}

// setFocus sets focus to the specified panel using focusState as the single source of truth.
// Updates all component focus states to match.
func (m Model) setFocus(panel focus.GlobalPanelID) Model {
	m.focusState.SetActivePanel(panel)
	// Update component focus states to match focusState
	m.deviceInfo = m.deviceInfo.SetFocused(panel == focus.PanelDashboardInfo)
	m.deviceList = m.deviceList.SetFocused(panel == focus.PanelDeviceList)
	m.events = m.events.SetFocused(panel == focus.PanelDashboardEvents)
	m.energyBars = m.energyBars.SetFocused(panel == focus.PanelDashboardEnergyBars)
	*m.energyHistory = m.energyHistory.SetFocused(panel == focus.PanelDashboardEnergyHistory)
	return m
}

// openJSONViewer opens the JSON viewer for the selected device if on Detail panel.
func (m Model) openJSONViewer() (Model, tea.Cmd, bool) {
	if !m.focusState.IsPanelFocused(focus.PanelDashboardInfo) {
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

	m = m.setFocus(focus.PanelDashboardJSON)
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

	// Handle component-level actions (edit, new, delete, etc.)
	// These are forwarded to views as action messages
	if newModel, cmd, handled := m.dispatchComponentAction(action); handled {
		return newModel, cmd, true
	}

	// Handle navigation actions - convert to NavigationMsg
	if newModel, cmd, handled := m.dispatchNavigationAction(action); handled {
		return newModel, cmd, true
	}

	// Actions not directly handled - let existing handlers deal with them
	return m, nil, false
}

// dispatchGlobalAction handles global actions like quit, help, filter, command mode.
func (m Model) dispatchGlobalAction(action keys.Action) (Model, tea.Cmd, bool) {
	switch action {
	case keys.ActionQuit:
		m.quitting = true
		return m, tea.Quit, true
	case keys.ActionHelp:
		return m.dispatchHelpAction()
	case keys.ActionFilter:
		return m.dispatchFilterAction()
	case keys.ActionCommand:
		return m.dispatchCommandAction()
	case keys.ActionRefresh:
		return m.dispatchRefreshAction()
	case keys.ActionRefreshAll:
		return m, tea.Batch(
			func() tea.Msg { return cache.RefreshTickMsg{} },
			toast.Show("Refreshing all devices", toast.LevelInfo),
		), true
	case keys.ActionNextPanel, keys.ActionPrevPanel:
		return m.dispatchPanelCycleAction(action)
	case keys.ActionEscape:
		return m.dispatchEscapeAction()
	case keys.ActionDebug:
		return m.dispatchDebugAction()
	default:
		return m, nil, false
	}
}

func (m Model) dispatchHelpAction() (Model, tea.Cmd, bool) {
	m.help = m.help.SetSize(m.width, m.height)
	m.help = m.help.SetContext(m.getHelpContext())
	wasVisible := m.help.Visible()
	m.help = m.help.Toggle()
	// Update focusState based on visibility change
	if m.help.Visible() && !wasVisible {
		m.focusState.PushOverlay(focus.OverlayHelp, focus.ModeOverlay)
	} else if !m.help.Visible() && wasVisible {
		m.focusState.PopOverlay()
	}
	return m, nil, true
}

func (m Model) dispatchFilterAction() (Model, tea.Cmd, bool) {
	var cmd tea.Cmd
	m.search, cmd = m.search.Activate()
	m = m.updateViewManagerSize()
	m.focusState.PushOverlay(focus.OverlaySearch, focus.ModeInput)
	return m, cmd, true
}

func (m Model) dispatchCommandAction() (Model, tea.Cmd, bool) {
	var cmd tea.Cmd
	m.cmdMode, cmd = m.cmdMode.Activate()
	m = m.updateViewManagerSize()
	m.focusState.PushOverlay(focus.OverlayCmdMode, focus.ModeInput)
	return m, cmd, true
}

func (m Model) dispatchRefreshAction() (Model, tea.Cmd, bool) {
	device := m.deviceList.SelectedDevice()
	if device == nil {
		return m, toast.Show("No device selected", toast.LevelWarning), true
	}
	return m, tea.Batch(
		func() tea.Msg { return cache.DeviceRefreshMsg{Name: device.Device.Name} },
		toast.Show(fmt.Sprintf("Refreshing %s", device.Device.Name), toast.LevelInfo),
	), true
}

func (m Model) dispatchPanelCycleAction(action keys.Action) (Model, tea.Cmd, bool) {
	// Only handle panel cycling for Dashboard - other tabs handle it in their views
	if !m.isDashboardActive() {
		return m, nil, false
	}
	if action == keys.ActionNextPanel {
		m.focusState.NextPanel()
	} else {
		m.focusState.PrevPanel()
	}
	m = m.syncComponentFocus()
	return m, nil, true
}

func (m Model) dispatchEscapeAction() (Model, tea.Cmd, bool) {
	if m.filter != "" {
		m.filter = ""
		m.cursor = 0
		m.deviceList = m.deviceList.SetFilter("")
		m.deviceList = m.deviceList.SetCursor(0)
		return m, nil, true
	}
	return m, nil, false
}

// dispatchDebugAction handles the debug toggle action.
func (m Model) dispatchDebugAction() (Model, tea.Cmd, bool) {
	enabled, sessionDir := m.debugLogger.Toggle()
	// Update global logger for trace logging from components
	if enabled {
		debug.SetGlobal(m.debugLogger)
	} else {
		debug.SetGlobal(nil)
	}
	var desc, toastMsg string
	if enabled {
		desc = "Debug session started: " + sessionDir
		toastMsg = "Debug ON: " + sessionDir
	} else {
		desc = "Debug session ended"
		toastMsg = "Debug OFF: session saved"
	}
	debugEvent := events.EventMsg{
		Events: []events.Event{{
			Timestamp:   time.Now(),
			Device:      "system",
			Component:   "debug",
			Type:        "info",
			Description: desc,
		}},
	}
	m.statusBar = m.statusBar.SetDebugActive(enabled)
	return m, tea.Batch(toast.Success(toastMsg), func() tea.Msg { return debugEvent }), true
}

// dispatchTabAction handles tab switching actions (1-6).
func (m Model) dispatchTabAction(action keys.Action) (Model, tea.Cmd, bool) {
	switch action {
	case keys.ActionTab1:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabDashboard)
		m.focusState.SetActiveTab(tabs.TabDashboard)
		m = m.setFocus(focus.PanelDeviceList)
		// Fetch extended status for current device and sync device info
		var extCmd tea.Cmd
		if dev := m.deviceList.SelectedDevice(); dev != nil && dev.Online {
			debug.TraceEvent("Tab1 switch: device=%s online, triggering FetchExtendedStatus", dev.Device.Name)
			extCmd = m.cache.FetchExtendedStatus(dev.Device.Name)
		} else {
			debug.TraceEvent("Tab1 switch: no device selected or offline")
		}
		m = m.syncDeviceInfo()
		return m, tea.Batch(m.viewManager.SetActive(views.ViewDashboard), extCmd), true

	case keys.ActionTab2:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabAutomation)
		m.focusState.SetActiveTab(tabs.TabAutomation)
		m = m.setFocus(focus.PanelDeviceList)
		// Propagate current device if one is selected
		var deviceName string
		if dev := m.deviceList.SelectedDevice(); dev != nil {
			deviceName = dev.Device.Name
		}
		return m, tea.Batch(
			m.viewManager.SetActive(views.ViewAutomation),
			m.viewManager.PropagateDevice(deviceName),
		), true

	case keys.ActionTab3:
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabConfig)
		m.focusState.SetActiveTab(tabs.TabConfig)
		m = m.setFocus(focus.PanelDeviceList)
		// Propagate current device if one is selected
		var deviceName string
		if dev := m.deviceList.SelectedDevice(); dev != nil {
			deviceName = dev.Device.Name
		}
		return m, tea.Batch(
			m.viewManager.SetActive(views.ViewConfig),
			m.viewManager.PropagateDevice(deviceName),
		), true

	case keys.ActionTab4:
		prevTab := m.focusState.ActiveTab()
		prevPanel := m.focusState.ActivePanel()
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabManage)
		// Manage tab has no device list - SetActiveTab sets first panel (PanelManageDiscovery)
		m.focusState.SetActiveTab(tabs.TabManage)
		cmd := m.focusState.EmitChanged(prevTab, prevPanel, true, true, false)
		return m, tea.Batch(m.viewManager.SetActive(views.ViewManage), cmd), true

	case keys.ActionTab5:
		prevTab := m.focusState.ActiveTab()
		prevPanel := m.focusState.ActivePanel()
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabMonitor)
		// Monitor tab has no device list - SetActiveTab sets first panel (PanelMonitorPowerRanking)
		m.focusState.SetActiveTab(tabs.TabMonitor)
		cmd := m.focusState.EmitChanged(prevTab, prevPanel, true, true, false)
		return m, tea.Batch(m.viewManager.SetActive(views.ViewMonitor), cmd), true

	case keys.ActionTab6:
		prevTab := m.focusState.ActiveTab()
		prevPanel := m.focusState.ActivePanel()
		m.tabBar, _ = m.tabBar.SetActive(tabs.TabFleet)
		// Fleet tab has no device list - SetActiveTab sets first panel (PanelFleetDevices)
		m.focusState.SetActiveTab(tabs.TabFleet)
		cmd := m.focusState.EmitChanged(prevTab, prevPanel, true, true, false)
		return m, tea.Batch(m.viewManager.SetActive(views.ViewFleet), cmd), true

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
		m = m.setFocus(focus.PanelDeviceList)
		return m, nil, true
	case keys.ActionPanel2:
		m = m.setFocus(focus.PanelDashboardInfo)
		return m, nil, true
	case keys.ActionPanel3:
		m = m.setFocus(focus.PanelDashboardEvents)
		return m, nil, true
	case keys.ActionPanel4:
		m = m.setFocus(focus.PanelDashboardEnergyBars)
		return m, nil, true
	case keys.ActionPanel5:
		m = m.setFocus(focus.PanelDashboardEnergyHistory)
		return m, nil, true
	case keys.ActionPanel6, keys.ActionPanel7, keys.ActionPanel8, keys.ActionPanel9:
		// Panel 6-9 not mapped in Dashboard layout
		return m, nil, false
	default:
		return m, nil, false
	}
}

// dispatchDeviceKeyAction handles device control actions (toggle, on, off, reboot, enter, browser, control, detail).
func (m Model) dispatchDeviceKeyAction(action keys.Action) (Model, tea.Cmd, bool) {
	// Monitor tab has its own device selection (power ranking), not the device list
	if m.isMonitorTabActive() {
		return m.dispatchMonitorDeviceAction(action)
	}

	switch action {
	case keys.ActionToggle:
		// When focused on deviceInfo panel, let it handle toggle for individual component control
		if m.focusState.IsPanelFocused(focus.PanelDashboardInfo) {
			return m, nil, false // Fall through to handleNavigation -> deviceInfo.Update
		}
		return m.dispatchDeviceAction(actionToggle)
	case keys.ActionOn:
		return m.dispatchDeviceAction(actionOn)
	case keys.ActionOff:
		return m.dispatchDeviceAction(actionOff)
	case keys.ActionReboot:
		return m.dispatchDeviceAction(actionReboot)
	case keys.ActionEnter:
		return m.dispatchEnterAction()
	case keys.ActionBrowser:
		return m.dispatchBrowserAction()
	case keys.ActionControl:
		return m.dispatchControlAction()
	case keys.ActionDetail:
		return m.dispatchDetailAction()
	case keys.ActionPlatformFilter:
		return m.dispatchPlatformFilterAction()
	default:
		return m, nil, false
	}
}

// dispatchPlatformFilterAction cycles the platform filter on the device list.
func (m Model) dispatchPlatformFilterAction() (Model, tea.Cmd, bool) {
	if !m.hasDeviceList() {
		return m, nil, false
	}
	m.deviceList, _ = m.deviceList.Update(messages.PlatformFilterMsg{})
	pf := m.deviceList.PlatformFilter()
	if pf != "" {
		return m, toast.Show("Platform: "+pf, toast.LevelInfo), true
	}
	return m, toast.Show("Platform: all", toast.LevelInfo), true
}

// dispatchBrowserAction opens the selected device's web UI in the browser.
func (m Model) dispatchBrowserAction() (Model, tea.Cmd, bool) {
	if !m.hasDeviceList() {
		return m, nil, false
	}
	device := m.deviceList.SelectedDevice()
	if device == nil || device.Device.Address == "" {
		return m, nil, false
	}
	return m, m.openDeviceBrowser(device.Device.Address), true
}

// dispatchDeviceAction executes a device action on the selected device.
func (m Model) dispatchDeviceAction(action string) (Model, tea.Cmd, bool) {
	if !m.hasDeviceList() {
		return m, nil, false
	}

	// Handle reboot with confirmation
	if action == actionReboot {
		devices := m.getFilteredDevices()
		if m.cursor >= len(devices) {
			return m, nil, false
		}
		selected := devices[m.cursor]
		if !selected.Online {
			return m, nil, false
		}
		m.confirmingReboot = true
		m.pendingRebootDevice = selected.Device.Name
		return m, nil, true
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
	if m.focusState.IsPanelFocused(focus.PanelDashboardInfo) {
		return m.openJSONViewer()
	}
	// Default: forward to view
	return m, nil, false
}

// dispatchControlAction opens the control panel for the selected device.
func (m Model) dispatchControlAction() (Model, tea.Cmd, bool) {
	if !m.hasDeviceList() {
		return m, nil, false
	}
	return m.showControlPanel()
}

// dispatchDetailAction opens the device detail overlay for the selected device.
func (m Model) dispatchDetailAction() (Model, tea.Cmd, bool) {
	if !m.hasDeviceList() {
		return m, nil, false
	}
	return m.showDeviceDetail()
}

// dispatchComponentAction handles component-level actions by converting them to
// action messages and forwarding them to the view manager.
// This allows components to handle semantic actions instead of raw key strings.
func (m Model) dispatchComponentAction(action keys.Action) (Model, tea.Cmd, bool) {
	var msg tea.Msg

	switch action {
	case keys.ActionEdit:
		msg = messages.EditRequestMsg{}
	case keys.ActionNew:
		msg = messages.NewRequestMsg{}
	case keys.ActionDelete:
		msg = messages.DeleteRequestMsg{}
	case keys.ActionCopy:
		msg = messages.CopyRequestMsg{}
	case keys.ActionPause:
		msg = messages.PauseRequestMsg{}
	case keys.ActionClear:
		msg = messages.ClearRequestMsg{}
	case keys.ActionExpand:
		msg = messages.ExpandRequestMsg{}
	case keys.ActionFilterToggle:
		msg = messages.FilterToggleRequestMsg{}
	case keys.ActionSort:
		msg = messages.SortRequestMsg{}
	default:
		return m, nil, false
	}

	// Forward the action message to the view manager
	cmd := m.viewManager.Update(msg)
	return m, cmd, true
}

// dispatchNavigationAction converts navigation actions to NavigationMsg and forwards to views.
// This provides a unified navigation interface for components.
func (m Model) dispatchNavigationAction(action keys.Action) (Model, tea.Cmd, bool) {
	var dir messages.NavDirection

	switch action {
	case keys.ActionUp:
		dir = messages.NavUp
	case keys.ActionDown:
		dir = messages.NavDown
	case keys.ActionLeft:
		dir = messages.NavLeft
	case keys.ActionRight:
		dir = messages.NavRight
	case keys.ActionPageUp:
		dir = messages.NavPageUp
	case keys.ActionPageDown:
		dir = messages.NavPageDown
	case keys.ActionHome:
		dir = messages.NavHome
	case keys.ActionEnd:
		dir = messages.NavEnd
	default:
		return m, nil, false
	}

	navMsg := messages.NavigationMsg{Direction: dir}

	// Dashboard components are managed directly in app.go, not in viewManager
	if m.isDashboardActive() {
		return m.routeDashboardNavigation(navMsg)
	}

	// Forward navigation to the view manager for other tabs
	cmd := m.viewManager.Update(navMsg)
	return m, cmd, true
}

// routeDashboardNavigation routes NavigationMsg to the focused Dashboard component.
func (m Model) routeDashboardNavigation(navMsg messages.NavigationMsg) (Model, tea.Cmd, bool) {
	panel := m.focusState.ActivePanel()

	var cmd tea.Cmd
	switch panel {
	case focus.PanelDeviceList:
		m.deviceList, cmd = m.deviceList.Update(navMsg)
		return m, cmd, true
	case focus.PanelDashboardInfo:
		m.deviceInfo, cmd = m.deviceInfo.Update(navMsg)
		return m, cmd, true
	case focus.PanelDashboardEvents:
		m.events, cmd = m.events.Update(navMsg)
		return m, cmd, true
	case focus.PanelDashboardEnergyBars:
		m.energyBars, cmd = m.energyBars.Update(navMsg)
		return m, cmd, true
	case focus.PanelDashboardEnergyHistory:
		*m.energyHistory, cmd = m.energyHistory.Update(navMsg)
		return m, cmd, true
	default:
		return m, nil, false
	}
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
	case "j", keyconst.KeyDown:
		return navDown, true
	case "k", keyconst.KeyUp:
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
	case "j", "k", keyconst.KeyUp, keyconst.KeyDown, "g", "G",
		keyconst.KeyPgDown, keyconst.KeyPgUp, keyconst.KeyCtrlD, keyconst.KeyCtrlU:
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
	if m.focusState.IsPanelFocused(focus.PanelDeviceList) {
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

	// When PanelDashboardInfo is focused, forward navigation to deviceInfo component
	if m.focusState.IsPanelFocused(focus.PanelDashboardInfo) {
		var cmd tea.Cmd
		m.deviceInfo, cmd = m.deviceInfo.Update(msg)
		return m, cmd, true
	}

	// When PanelDashboardEvents is focused, forward navigation to events component
	if m.focusState.IsPanelFocused(focus.PanelDashboardEvents) {
		var cmd tea.Cmd
		m.events, cmd = m.events.Update(msg)
		return m, cmd, true
	}

	// When PanelDashboardEnergyBars or PanelDashboardEnergyHistory is focused, handle scrolling
	if m.focusState.IsPanelFocused(focus.PanelDashboardEnergyBars) ||
		m.focusState.IsPanelFocused(focus.PanelDashboardEnergyHistory) {
		return m.handleEnergyPanelNavigation(msg)
	}

	// Handle endpoints panel navigation
	devices := m.getFilteredDevices()
	keyStr := msg.String()

	if m.focusState.IsPanelFocused(focus.PanelDashboardJSON) {
		if dir, ok := parseNavDirection(keyStr); ok {
			m = m.navEndpoints(dir, devices)
			return m, nil, true
		}
	}

	return m, nil, false
}

// handleEnergyPanelNavigation handles scroll navigation for energy panels.
func (m Model) handleEnergyPanelNavigation(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	keyStr := msg.String()
	isBars := m.focusState.IsPanelFocused(focus.PanelDashboardEnergyBars)

	switch keyStr {
	case "j", keyconst.KeyDown:
		if isBars {
			m.energyBars = m.energyBars.ScrollDown()
		} else {
			*m.energyHistory = m.energyHistory.ScrollDown()
		}
		return m, nil, true
	case "k", keyconst.KeyUp:
		if isBars {
			m.energyBars = m.energyBars.ScrollUp()
		} else {
			*m.energyHistory = m.energyHistory.ScrollUp()
		}
		return m, nil, true
	case keyconst.KeyCtrlD, keyconst.KeyPgDown:
		if isBars {
			m.energyBars = m.energyBars.PageDown()
		} else {
			*m.energyHistory = m.energyHistory.PageDown()
		}
		return m, nil, true
	case keyconst.KeyCtrlU, keyconst.KeyPgUp:
		if isBars {
			m.energyBars = m.energyBars.PageUp()
		} else {
			*m.energyHistory = m.energyHistory.PageUp()
		}
		return m, nil, true
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
		dev := devices[m.cursor]
		// Debug: log what we're syncing - include pointer address to track identity
		wifiOK := dev.WiFi != nil
		sysOK := dev.Sys != nil
		debug.TraceEvent("syncDeviceInfo: %s wifi=%v sys=%v ptr=%p", dev.Device.Name, wifiOK, sysOK, dev)
		if wifiOK {
			debug.TraceEvent("syncDeviceInfo: WiFi SSID=%s RSSI=%d", dev.WiFi.SSID, dev.WiFi.RSSI)
		}
		m.deviceInfo = m.deviceInfo.SetDevice(dev)
	} else {
		m.deviceInfo = m.deviceInfo.SetDevice(nil)
	}

	// Update focus state and panel index
	m.deviceInfo = m.deviceInfo.SetFocused(m.focusState.IsPanelFocused(focus.PanelDashboardInfo)).SetPanelIndex(2)

	return m
}

// showDeviceDetail shows the device detail overlay for the selected device.
func (m Model) showDeviceDetail() (Model, tea.Cmd, bool) {
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
	m.focusState.PushOverlay(focus.OverlayDeviceDetail, focus.ModeOverlay)
	return m, cmd, true
}

// showControlPanel shows the control panel overlay for the selected device's component.
func (m Model) showControlPanel() (Model, tea.Cmd, bool) {
	devices := m.getFilteredDevices()
	if m.cursor >= len(devices) || m.cursor < 0 {
		return m, nil, false
	}

	selected := devices[m.cursor]
	if !selected.Online {
		return m, nil, false
	}

	data := m.cache.GetDevice(selected.Device.Name)
	if data == nil {
		return m, nil, false
	}

	m.controlPanel = m.controlPanel.SetSize(m.width*2/3, m.height*2/3)

	// Plugin devices use the generic plugin control panel
	if selected.Device.IsPluginManaged() {
		return m.showPluginControlPanel(data, selected.Device)
	}

	selectedIdx := m.deviceInfo.SelectedComponent()

	// Try to show component at selected index
	if result, ok := m.showComponentAtIndex(data, selected.Device.Address, selectedIdx); ok {
		return result, nil, true
	}

	// Fall back to first available controllable component
	if result, ok := m.showFirstAvailableComponent(data, selected.Device.Address); ok {
		return result, nil, true
	}

	return m, nil, false
}

// showComponentAtIndex attempts to show the control panel for the component at the given index.
func (m Model) showComponentAtIndex(data *cache.DeviceData, address string, selectedIdx int) (Model, bool) {
	idx := 0

	// Switches (index -1 means first switch)
	for _, sw := range data.Switches {
		if selectedIdx == -1 || selectedIdx == idx {
			return m.activateSwitchControl(address, sw), true
		}
		idx++
	}

	// Lights
	for _, lt := range data.Lights {
		if selectedIdx == idx {
			return m.activateLightControl(address, lt), true
		}
		idx++
	}

	// Covers
	for _, cv := range data.Covers {
		if selectedIdx == idx {
			return m.activateCoverControl(address, cv), true
		}
		idx++
	}

	// RGBs
	for _, rgb := range data.RGBs {
		if selectedIdx == idx {
			return m.activateRGBControl(address, rgb), true
		}
		idx++
	}

	// Thermostats
	for _, th := range data.Thermostats {
		if selectedIdx == idx {
			return m.activateThermostatControl(address, th), true
		}
		idx++
	}

	return m, false
}

// showFirstAvailableComponent shows the first available controllable component.
func (m Model) showFirstAvailableComponent(data *cache.DeviceData, address string) (Model, bool) {
	if len(data.Switches) > 0 {
		return m.activateSwitchControl(address, data.Switches[0]), true
	}
	if len(data.Lights) > 0 {
		return m.activateLightControl(address, data.Lights[0]), true
	}
	if len(data.Covers) > 0 {
		return m.activateCoverControl(address, data.Covers[0]), true
	}
	if len(data.RGBs) > 0 {
		return m.activateRGBControl(address, data.RGBs[0]), true
	}
	if len(data.Thermostats) > 0 {
		return m.activateThermostatControl(address, data.Thermostats[0]), true
	}
	return m, false
}

func (m Model) activateSwitchControl(address string, sw cache.SwitchState) Model {
	state := control.SwitchState{ID: sw.ID, Output: sw.On, Source: sw.Source}
	m.controlPanel = m.controlPanel.ShowSwitch(address, state)
	m.focusState.PushOverlay(focus.OverlayControlPanel, focus.ModeOverlay)
	return m
}

func (m Model) activateLightControl(address string, lt cache.LightState) Model {
	state := control.LightState{ID: lt.ID, Output: lt.On}
	m.controlPanel = m.controlPanel.ShowLight(address, state)
	m.focusState.PushOverlay(focus.OverlayControlPanel, focus.ModeOverlay)
	return m
}

func (m Model) activateCoverControl(address string, cv cache.CoverState) Model {
	state := control.CoverState{ID: cv.ID, State: cv.State, Position: 50}
	m.controlPanel = m.controlPanel.ShowCover(address, state)
	m.focusState.PushOverlay(focus.OverlayControlPanel, focus.ModeOverlay)
	return m
}

func (m Model) activateRGBControl(address string, rgb cache.RGBState) Model {
	state := control.RGBState{
		ID:         rgb.ID,
		Name:       rgb.Name,
		Output:     rgb.Output,
		Brightness: rgb.Brightness,
		Red:        rgb.Red,
		Green:      rgb.Green,
		Blue:       rgb.Blue,
		White:      rgb.White,
		Power:      rgb.Power,
		Source:     rgb.Source,
	}
	m.controlPanel = m.controlPanel.ShowRGB(address, state)
	m.focusState.PushOverlay(focus.OverlayControlPanel, focus.ModeOverlay)
	return m
}

func (m Model) activateThermostatControl(address string, th cache.ThermostatState) Model {
	state := control.ThermostatState{
		ID:              th.ID,
		Name:            th.Name,
		Enabled:         th.Enabled,
		Mode:            th.Mode,
		TargetC:         th.TargetC,
		CurrentC:        th.CurrentC,
		CurrentHumidity: th.CurrentHumidity,
		ValvePosition:   th.ValvePosition,
		BoostActive:     th.BoostActive,
		BoostRemaining:  th.BoostRemaining,
		OverrideActive:  th.OverrideActive,
		Source:          th.Source,
	}
	m.controlPanel = m.controlPanel.ShowThermostat(address, state)
	m.focusState.PushOverlay(focus.OverlayControlPanel, focus.ModeOverlay)
	return m
}

// showPluginControlPanel shows the plugin-specific control panel.
func (m Model) showPluginControlPanel(data *cache.DeviceData, device devmodel.Device) (Model, tea.Cmd, bool) {
	components := buildPluginComponents(data)
	if len(components) == 0 {
		return m, nil, false
	}

	m.controlPanel = m.controlPanel.ShowPlugin(device.Address, device.GetPlatform(), components)
	m.focusState.PushOverlay(focus.OverlayControlPanel, focus.ModeOverlay)
	return m, nil, true
}

// buildPluginComponents converts cache data into plugin control components.
func buildPluginComponents(data *cache.DeviceData) []control.PluginComponent {
	components := make([]control.PluginComponent, 0, len(data.Switches)+len(data.Lights)+len(data.Covers))

	for _, sw := range data.Switches {
		components = append(components, control.PluginComponent{
			Type: "switch",
			ID:   sw.ID,
			Name: sw.Name,
			On:   sw.On,
		})
	}
	for _, lt := range data.Lights {
		components = append(components, control.PluginComponent{
			Type: "light",
			ID:   lt.ID,
			Name: lt.Name,
			On:   lt.On,
		})
	}
	for _, cv := range data.Covers {
		components = append(components, control.PluginComponent{
			Type:  "cover",
			ID:    cv.ID,
			Name:  cv.Name,
			State: cv.State,
		})
	}

	return components
}

// executeDeviceAction executes a device action on the selected device.
// When called from the device list, this toggles ALL controllable components (componentID=nil).
// For individual component control, use the deviceInfo panel which routes through handleRequestToggle.
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

	return func() tea.Msg {
		var err error

		switch action {
		case actionToggle:
			// QuickToggle with nil toggles all controllable components (switch, light, rgb, cover)
			_, err = svc.QuickToggle(m.ctx, device.Address, nil)
		case actionOn:
			_, err = svc.QuickOn(m.ctx, device.Address, nil)
		case actionOff:
			_, err = svc.QuickOff(m.ctx, device.Address, nil)
		case actionReboot:
			err = svc.DeviceReboot(m.ctx, device.Address, 0)
		}

		return DeviceActionMsg{
			Device: device.Name,
			Action: action,
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

	// Update Monitor tab badge with triggered alert count
	m.tabBar = m.updateMonitorTabBadge()

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

	// Render view modal overlay (edit modals from Config, Automation views)
	if m.viewManager.HasActiveModal() {
		content = m.renderWithModalOverlay(content, contentHeight)
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

// renderInputBar renders the search, command mode, or toast input bar.
// Priority: search > cmdmode > toast (only one shows at a time).
func (m Model) renderInputBar() string {
	if m.search.IsActive() {
		return m.search.View()
	}
	if m.cmdMode.IsActive() {
		return m.cmdMode.View()
	}
	// Show toast in input bar area when no other input is active
	if m.toast.HasToasts() {
		return m.toast.ViewAsInputBar()
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
// Ensures every line is exactly m.width visible characters to prevent ghosting
// when switching tabs (old content showing through).
func (m Model) padContent(content string, contentHeight int) string {
	cw := m.contentWidth()
	pad := strings.Repeat(" ", horizontalPadding)
	contentLines := strings.Split(content, "\n")
	paddedLines := make([]string, contentHeight)
	emptyLine := strings.Repeat(" ", m.width)
	for i := range contentHeight {
		if i < len(contentLines) {
			line := contentLines[i]
			lineWidth := ansi.StringWidth(line)
			if lineWidth > cw {
				// Truncate lines that are too wide
				line = ansi.Truncate(line, cw, "")
				// Re-measure after truncation as ansi.Truncate may return shorter string
				lineWidth = ansi.StringWidth(line)
			}
			if lineWidth < cw {
				// Pad lines that are too narrow to fill content width
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

// applyOverlays applies confirm, device detail, and control panel overlays.
// Note: Toast is now rendered in the input bar area (renderInputBar), not as an overlay.
func (m Model) applyOverlays(result string) string {
	if m.confirmingReboot {
		result = m.centerOverlay(m.renderRebootConfirmation())
	}
	if m.confirm.Visible() {
		result = m.confirm.Overlay(result)
	}
	if m.deviceDetail.Visible() {
		result = m.centerOverlay(m.deviceDetail.View())
	}
	if m.controlPanel.Visible() {
		result = m.centerOverlay(m.controlPanel.View())
	}
	return result
}

// renderRebootConfirmation renders the reboot confirmation dialog.
func (m Model) renderRebootConfirmation() string {
	colors := theme.GetSemanticColors()
	confirmStyle := lipgloss.NewStyle().
		Foreground(colors.Error).
		Bold(true)
	warningStyle := lipgloss.NewStyle().
		Foreground(colors.Warning)
	mutedStyle := lipgloss.NewStyle().
		Foreground(colors.Muted)
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Error).
		Padding(1, 3)

	var content strings.Builder
	content.WriteString(confirmStyle.Render("Reboot device: " + m.pendingRebootDevice + "?"))
	content.WriteString("\n\n")
	content.WriteString(warningStyle.Render("The device will be temporarily unavailable."))
	content.WriteString("\n\n")
	content.WriteString(mutedStyle.Render("Press Y to confirm, N or Esc to cancel"))

	return borderStyle.Render(content.String())
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
	switch m.focusState.ActivePanel() {
	case focus.PanelDeviceList:
		return "DeviceList"
	case focus.PanelDashboardInfo:
		return "Detail"
	case focus.PanelDashboardEvents:
		return "Events"
	case focus.PanelDashboardJSON:
		return "JSON"
	case focus.PanelDashboardEnergyBars:
		return "EnergyBars"
	case focus.PanelDashboardEnergyHistory:
		return "EnergyHistory"
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
// This includes Dashboard, Automation, and Config tabs (not Monitor — it uses power ranking).
func (m Model) hasDeviceList() bool {
	switch m.tabBar.ActiveTabID() {
	case tabs.TabDashboard, tabs.TabAutomation, tabs.TabConfig:
		return true
	default:
		return false
	}
}

// updateMonitorTabBadge updates the Monitor tab badge with the triggered alert count.
func (m Model) updateMonitorTabBadge() tabs.Model {
	v := m.viewManager.Get(views.ViewMonitor)
	if v == nil {
		return m.tabBar
	}
	monitorView, ok := v.(*views.Monitor)
	if !ok {
		return m.tabBar
	}
	count := monitorView.TriggeredAlertCount()
	if count > 0 {
		return m.tabBar.SetBadge(tabs.TabMonitor, fmt.Sprintf("(%d)", count))
	}
	return m.tabBar.SetBadge(tabs.TabMonitor, "")
}

// isMonitorTabActive returns true if the Monitor tab is currently active.
func (m Model) isMonitorTabActive() bool {
	return m.tabBar.ActiveTabID() == tabs.TabMonitor
}

// getMonitorSelectedDevice returns the currently selected device from the Monitor view's power ranking.
func (m Model) getMonitorSelectedDevice() *monitor.DeviceStatus {
	v := m.viewManager.Get(views.ViewMonitor)
	if v == nil {
		return nil
	}
	monitorView, ok := v.(*views.Monitor)
	if !ok {
		return nil
	}
	return monitorView.SelectedDevice()
}

// dispatchMonitorDeviceAction handles device actions for the Monitor tab.
// Uses the power ranking's selected device instead of the device list.
// Device-specific actions only apply when Power Ranking panel is focused.
func (m Model) dispatchMonitorDeviceAction(action keys.Action) (Model, tea.Cmd, bool) {
	isPowerRankingFocused := m.focusState.IsPanelFocused(focus.PanelMonitorPowerRanking)

	// Device-specific actions: only when Power Ranking panel is focused
	if isPowerRankingFocused {
		switch action {
		case keys.ActionToggle:
			return m.toggleMonitorDevice()
		case keys.ActionOn:
			return m.monitorDeviceQuickAction(actionOn)
		case keys.ActionOff:
			return m.monitorDeviceQuickAction(actionOff)
		case keys.ActionControl:
			return m.showMonitorControlPanel()
		case keys.ActionDetail, keys.ActionEnter:
			return m.showMonitorDeviceDetail()
		case keys.ActionBrowser:
			return m.openMonitorDeviceBrowser()
		case keys.ActionHistory:
			return m.showMonitorEnergyHistory()
		case keys.ActionPhaseDetail:
			return m.showMonitorPhaseDetail()
		default:
			// Not a power-ranking action; fall through to global actions
		}
	}

	// Global monitor actions: work from any panel
	switch action {
	case keys.ActionExport:
		cmd := m.viewManager.Update(messages.ExportRequestMsg{})
		return m, cmd, true
	case keys.ActionExportJSON:
		cmd := m.viewManager.Update(messages.ExportRequestMsg{Format: "json"})
		return m, cmd, true
	case keys.ActionViewJSON:
		return m.openMonitorJSONViewer()
	default:
		return m, nil, false
	}
}

// toggleMonitorDevice toggles the selected device from the Monitor power ranking.
func (m Model) toggleMonitorDevice() (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || !dev.Online {
		return m, nil, false
	}

	svc := m.factory.ShellyService()
	name := dev.Name
	addr := dev.Address

	return m, func() tea.Msg {
		_, err := svc.QuickToggle(m.ctx, addr, nil)
		return DeviceActionMsg{Device: name, Action: actionToggle, Err: err}
	}, true
}

// monitorDeviceQuickAction performs a quick on/off action on the Monitor's selected device.
func (m Model) monitorDeviceQuickAction(action string) (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || !dev.Online {
		return m, nil, false
	}

	svc := m.factory.ShellyService()
	name := dev.Name
	addr := dev.Address

	return m, func() tea.Msg {
		var err error
		switch action {
		case actionOn:
			_, err = svc.QuickOn(m.ctx, addr, nil)
		case actionOff:
			_, err = svc.QuickOff(m.ctx, addr, nil)
		}
		return DeviceActionMsg{Device: name, Action: action, Err: err}
	}, true
}

// showMonitorControlPanel opens the control panel for the Monitor's selected device.
func (m Model) showMonitorControlPanel() (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || !dev.Online {
		return m, nil, false
	}

	data := m.cache.GetDevice(dev.Name)
	if data == nil {
		return m, nil, false
	}

	m.controlPanel = m.controlPanel.SetSize(m.width*2/3, m.height*2/3)

	// Show first available controllable component
	if result, ok := m.showFirstAvailableComponent(data, dev.Address); ok {
		return result, nil, true
	}

	return m, nil, false
}

// showMonitorDeviceDetail opens the device detail overlay for the Monitor's selected device.
func (m Model) showMonitorDeviceDetail() (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || !dev.Online {
		return m, nil, false
	}

	m.deviceDetail = m.deviceDetail.SetSize(m.width*2/3, m.height*2/3)

	var cmd tea.Cmd
	m.deviceDetail, cmd = m.deviceDetail.Show(devmodel.Device{Name: dev.Name, Address: dev.Address})
	m.focusState.PushOverlay(focus.OverlayDeviceDetail, focus.ModeOverlay)
	return m, cmd, true
}

// openMonitorDeviceBrowser opens the web UI for the Monitor's selected device.
func (m Model) openMonitorDeviceBrowser() (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || dev.Address == "" {
		return m, nil, false
	}
	return m, m.openDeviceBrowser(dev.Address), true
}

// showMonitorEnergyHistory opens the energy history overlay for the selected device.
func (m Model) showMonitorEnergyHistory() (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || !dev.Online {
		return m, nil, false
	}
	cmd := m.viewManager.Update(views.EnergyHistoryRequestMsg{DeviceName: dev.Name, Address: dev.Address, Type: dev.Type})
	return m, cmd, true
}

// showMonitorPhaseDetail opens the 3-phase detail overlay for the selected device.
func (m Model) showMonitorPhaseDetail() (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || !dev.Online {
		return m, nil, false
	}
	cmd := m.viewManager.Update(views.PhaseDetailRequestMsg{DeviceName: dev.Name, Address: dev.Address})
	return m, cmd, true
}

// openMonitorJSONViewer opens the JSON viewer for the Monitor's selected device.
func (m Model) openMonitorJSONViewer() (Model, tea.Cmd, bool) {
	dev := m.getMonitorSelectedDevice()
	if dev == nil || !dev.Online {
		return m, nil, false
	}

	data := m.cache.GetDevice(dev.Name)
	if data == nil {
		return m, nil, false
	}

	methods := m.getDeviceMethods(data)
	endpoint := "Shelly.GetStatus"
	if data.Device.Generation == 1 {
		endpoint = "/status"
	}

	var cmd tea.Cmd
	m.jsonViewer, cmd = m.jsonViewer.Open(dev.Address, endpoint, methods)
	m.focusState.PushOverlay(focus.OverlayJSONViewer, focus.ModeOverlay)
	return m, cmd, true
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
	energyHeight := height - topHeight // JoinVertical stacks directly, no gap needed

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
	m.deviceInfo = m.deviceInfo.SetFocused(m.focusState.IsPanelFocused(focus.PanelDashboardInfo)).SetPanelIndex(2)
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

// renderMonitorLayout renders the Monitor tab with the multi-panel layout.
// The Monitor view handles its own layout (summary bar + 2-column panels).
func (m Model) renderMonitorLayout(height int) string {
	cw := m.contentWidth()

	// The monitor view handles its own multi-panel layout
	m.viewManager.SetSize(cw, height)
	return lipgloss.NewStyle().Width(cw).Render(m.viewManager.View())
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

	// Split width evenly between bars and history (50/50)
	// Account for 1 char gap between panels
	halfWidth := (width - 1) / 2
	barsWidth := halfWidth
	historyWidth := width - barsWidth - 1 // Use remaining to avoid off-by-one

	// Ensure minimum widths for each panel (title needs at least ~25 chars)
	if barsWidth < 30 || historyWidth < 30 {
		// Not enough space for both, render only bars at full width
		barsIdx := focus.PanelDashboardEnergyBars.PanelIndex()
		m.energyBars = m.energyBars.SetSize(width, height).
			SetFocused(m.focusState.IsPanelFocused(focus.PanelDashboardEnergyBars)).
			SetPanelIndex(barsIdx)
		return m.energyBars.View()
	}

	// Set size and focus state for both panels
	barsIdx := focus.PanelDashboardEnergyBars.PanelIndex()
	historyIdx := focus.PanelDashboardEnergyHistory.PanelIndex()

	m.energyBars = m.energyBars.SetSize(barsWidth, height).
		SetFocused(m.focusState.IsPanelFocused(focus.PanelDashboardEnergyBars)).
		SetPanelIndex(barsIdx)
	*m.energyHistory = m.energyHistory.SetSize(historyWidth, height)
	*m.energyHistory = m.energyHistory.SetFocused(m.focusState.IsPanelFocused(focus.PanelDashboardEnergyHistory))
	*m.energyHistory = m.energyHistory.SetPanelIndex(historyIdx)

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
	m.deviceList = m.deviceList.SetFocused(m.focusState.IsPanelFocused(focus.PanelDeviceList)).SetPanelIndex(1)
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
	focused := m.focusState.IsPanelFocused(focus.PanelDeviceList)

	// Account for border when setting component size
	innerWidth := width - 2   // left + right border
	innerHeight := height - 2 // top + bottom border

	m.deviceList = m.deviceList.SetSize(innerWidth, innerHeight)
	m.deviceList = m.deviceList.SetFocused(focused).SetPanelIndex(1)
	m.deviceList = m.deviceList.SetListOnly(true)

	deviceCount := m.cache.DeviceCount()
	onlineCount := m.cache.OnlineCount()

	// Format count in yellow
	badge := theme.SemanticWarning().Render(fmt.Sprintf("%d/%d", onlineCount, deviceCount))

	r := rendering.New(width, height).
		SetTitle("Devices").
		SetBadge(badge).
		SetFocused(focused).
		SetPanelIndex(1)

	// Add component counts footer with emojis
	footer := m.buildDeviceListFooter()
	if footer != "" {
		r.SetFooter(footer)
	}

	return r.SetContent(m.deviceList.View()).Render()
}

// buildDeviceListFooter builds a footer with emoji-only component counts.
// Format: on(green)/off(red) for each component type.
func (m Model) buildDeviceListFooter() string {
	counts := m.cache.ComponentCounts()
	var parts []string

	colors := theme.GetSemanticColors()
	onStyle := lipgloss.NewStyle().Foreground(colors.Online)
	offStyle := lipgloss.NewStyle().Foreground(colors.Error)

	// Switches: ⚡️ on/off
	if counts.SwitchesOn > 0 || counts.SwitchesOff > 0 {
		parts = append(parts, fmt.Sprintf("⚡️%s/%s",
			onStyle.Render(fmt.Sprintf("%d", counts.SwitchesOn)),
			offStyle.Render(fmt.Sprintf("%d", counts.SwitchesOff))))
	}

	// Lights: 💡 on/off
	if counts.LightsOn > 0 || counts.LightsOff > 0 {
		parts = append(parts, fmt.Sprintf("💡%s/%s",
			onStyle.Render(fmt.Sprintf("%d", counts.LightsOn)),
			offStyle.Render(fmt.Sprintf("%d", counts.LightsOff))))
	}

	// Covers: 🪟 open/closed
	if counts.CoversOpen > 0 || counts.CoversClosed > 0 || counts.CoversMoving > 0 {
		parts = append(parts, fmt.Sprintf("🪟%s/%s",
			onStyle.Render(fmt.Sprintf("%d", counts.CoversOpen)),
			offStyle.Render(fmt.Sprintf("%d", counts.CoversClosed+counts.CoversMoving))))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " │ ")
}

// renderEventsColumn renders the events column with embedded title.
func (m Model) renderEventsColumn(width, height int) string {
	focused := m.focusState.IsPanelFocused(focus.PanelDashboardEvents)

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
		SetPanelIndex(3)

	// Resize events model to fit inside the border box
	// Footer is embedded in border, not a separate line
	m.events = m.events.SetSize(r.ContentWidth(), r.ContentHeight()).SetFocused(focused)
	eventsView := m.events.View()

	return r.SetContent(eventsView).Render()
}

func (m Model) renderHeader() string {
	colors := theme.GetSemanticColors()

	// Banner on the right - using Highlight color (cyan in Dracula theme)
	bannerStyle := lipgloss.NewStyle().
		Foreground(colors.Highlight).
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
	metaLines[1] = labelStyle.Render("Devices: ") +
		onlineStyle.Render(fmt.Sprintf("%d online", online))
	if offline > 0 {
		metaLines[1] += labelStyle.Render(" / ") + offlineStyle.Render(fmt.Sprintf("%d offline", offline))
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

// renderWithModalOverlay renders a view modal as a centered overlay on top of main content.
// This handles edit modals from Config and Automation views.
func (m Model) renderWithModalOverlay(content string, contentHeight int) string {
	modalView := m.viewManager.RenderActiveModal()
	if modalView == "" {
		return content
	}

	// Center the modal overlay on top of the content (same as help overlay)
	return lipgloss.Place(
		m.width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalView,
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
// If the browser cannot be opened, the URL is copied to clipboard.
func (m Model) openDeviceBrowser(address string) tea.Cmd {
	return func() tea.Msg {
		url := "http://" + address
		b := browser.New()
		if err := b.Browse(m.ctx, url); err != nil {
			// Check if URL was copied to clipboard as fallback
			var clipErr *browser.ClipboardFallbackError
			if errors.As(err, &clipErr) {
				return toast.ShowMsg{
					Level:   toast.LevelInfo,
					Message: "Could not open browser. URL copied to clipboard.",
				}
			}
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
