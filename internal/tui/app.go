package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/energybars"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/events"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/help"
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

// EndpointFetchMsg reports the result of an endpoint RPC call.
type EndpointFetchMsg struct {
	Method string
	Data   any
	Err    error
}

// UI constants.
const (
	selectedPrefix   = "▶ "
	unselectedPrefix = "  "
)

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
	endpointData    any        // Fetched endpoint data
	endpointMethod  string     // Currently fetched endpoint method
	endpointLoading bool       // Whether endpoint data is being fetched
	endpointError   error      // Error from endpoint fetch

	// Dimensions
	width  int
	height int

	// Components
	statusBar  statusbar.Model
	search     search.Model
	cmdMode    cmdmode.Model
	help       help.Model
	toast      toast.Model
	events     events.Model
	energyBars energybars.Model

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
	case DeviceActionMsg:
		newModel, cmd := m.handleDeviceAction(msg)
		return newModel, cmd, true
	case cache.DeviceUpdateMsg, cache.RefreshTickMsg:
		cmd := m.cache.Update(msg)
		return m, cmd, true
	case cache.AllDevicesLoadedMsg:
		cmd := m.cache.Update(msg)
		// Emit initial device selection for the first device
		devices := m.getFilteredDevices()
		if len(devices) > 0 && m.cursor < len(devices) {
			d := devices[m.cursor]
			return m, tea.Batch(cmd, func() tea.Msg {
				return views.DeviceSelectedMsg{
					Device:  d.Device.Name,
					Address: d.Device.Address,
				}
			}), true
		}
		return m, cmd, true
	case events.EventMsg, events.SubscriptionStatusMsg:
		var cmd tea.Cmd
		m.events, cmd = m.events.Update(msg)
		return m, cmd, true
	case search.FilterChangedMsg:
		m.filter = msg.Filter
		m.cursor = 0
		return m, nil, true
	case search.ClosedMsg, cmdmode.ClosedMsg:
		return m, nil, true
	case cmdmode.CommandMsg:
		newModel, cmd := m.handleCommand(msg)
		return newModel, cmd, true
	case cmdmode.ErrorMsg:
		return m, statusbar.SetMessage(msg.Message, statusbar.MessageError), true
	case tabs.TabChangedMsg:
		// Sync view manager with tab bar
		cmd := m.viewManager.SetActive(msg.Current)
		return m, cmd, true
	case views.ViewChangedMsg:
		// View changed, sync tab bar
		m.tabBar, _ = m.tabBar.SetActive(msg.Current)
		return m, nil, true
	case views.DeviceSelectedMsg:
		// Propagate device selection to all views that support it
		return m, m.viewManager.PropagateDevice(msg.Device), true
	case EndpointFetchMsg:
		m.endpointLoading = false
		if msg.Err != nil {
			m.endpointError = msg.Err
			m.endpointData = nil
		} else {
			m.endpointData = msg.Data
			m.endpointMethod = msg.Method
			m.endpointError = nil
		}
		return m, nil, true
	case tea.KeyPressMsg:
		return m.handleKeyPressMsg(msg)
	}
	return m, nil, false
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
		// Cycle between panels: DeviceList <-> Detail
		switch m.focusedPanel {
		case PanelDeviceList:
			m.focusedPanel = PanelDetail
		case PanelDetail:
			m.focusedPanel = PanelDeviceList
		case PanelEndpoints:
			// When JSON overlay is visible, go to DeviceList
			m.focusedPanel = PanelDeviceList
		}
		return m, nil, true
	case "shift+tab":
		// Reverse cycle between panels: Detail <-> DeviceList
		switch m.focusedPanel {
		case PanelDeviceList:
			m.focusedPanel = PanelDetail
		case PanelDetail:
			m.focusedPanel = PanelDeviceList
		case PanelEndpoints:
			// When JSON overlay is visible, go to Detail
			m.focusedPanel = PanelDetail
		}
		return m, nil, true
	case "enter":
		// Toggle JSON overlay - Enter shows it when on Detail, Escape hides it
		if m.focusedPanel == PanelDetail {
			m.focusedPanel = PanelEndpoints
			// Trigger endpoint fetch for the selected endpoint
			cmd := m.fetchSelectedEndpoint()
			return m, cmd, true
		}
	case "escape":
		// Close JSON overlay and clear endpoint data
		if m.focusedPanel == PanelEndpoints {
			m.focusedPanel = PanelDetail
			m.endpointData = nil
			m.endpointError = nil
			m.endpointLoading = false
			return m, nil, true
		}
	}
	return m, nil, false
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

// handleNavigation handles j/k/g/G/h/l navigation keys based on focused panel.
func (m Model) handleNavigation(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	devices := m.getFilteredDevices()
	deviceCount := len(devices)
	oldCursor := m.cursor

	switch msg.String() {
	case "j", "down":
		switch m.focusedPanel {
		case PanelDeviceList:
			if m.cursor < deviceCount-1 {
				m.cursor++
				m.componentCursor = -1
				m.endpointCursor = 0
			}
		case PanelDetail, PanelEndpoints:
			if m.cursor >= 0 && m.cursor < deviceCount {
				methods := m.getDeviceMethods(devices[m.cursor])
				if m.endpointCursor < len(methods)-1 {
					m.endpointCursor++
				}
			}
		}
		return m, m.emitDeviceSelection(devices, oldCursor), true
	case "k", "up":
		switch m.focusedPanel {
		case PanelDeviceList:
			if m.cursor > 0 {
				m.cursor--
				m.componentCursor = -1
				m.endpointCursor = 0
			}
		case PanelDetail, PanelEndpoints:
			if m.endpointCursor > 0 {
				m.endpointCursor--
			}
		}
		return m, m.emitDeviceSelection(devices, oldCursor), true
	case "g":
		switch m.focusedPanel {
		case PanelDeviceList:
			m.cursor = 0
			m.componentCursor = -1
			m.endpointCursor = 0
		case PanelDetail, PanelEndpoints:
			m.endpointCursor = 0
		}
		return m, m.emitDeviceSelection(devices, oldCursor), true
	case "G":
		switch m.focusedPanel {
		case PanelDeviceList:
			if deviceCount > 0 {
				m.cursor = deviceCount - 1
				m.componentCursor = -1
				m.endpointCursor = 0
			}
		case PanelDetail, PanelEndpoints:
			if m.cursor >= 0 && m.cursor < deviceCount {
				methods := m.getDeviceMethods(devices[m.cursor])
				if len(methods) > 0 {
					m.endpointCursor = len(methods) - 1
				}
			}
		}
		return m, m.emitDeviceSelection(devices, oldCursor), true
	case "h", "left":
		if m.componentCursor > -1 {
			m.componentCursor--
		}
		return m, nil, true
	case "l", "right":
		if m.cursor < deviceCount && m.cursor >= 0 {
			d := devices[m.cursor]
			maxComponent := len(d.Switches) - 1
			if m.componentCursor < maxComponent {
				m.componentCursor++
			}
		}
		return m, nil, true
	}
	return m, nil, false
}

// emitDeviceSelection emits a DeviceSelectedMsg if the cursor changed.
func (m Model) emitDeviceSelection(devices []*cache.DeviceData, oldCursor int) tea.Cmd {
	if m.cursor == oldCursor || len(devices) == 0 || m.cursor >= len(devices) {
		return nil
	}
	d := devices[m.cursor]
	return func() tea.Msg {
		return views.DeviceSelectedMsg{
			Device:  d.Device.Name,
			Address: d.Device.Address,
		}
	}
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

// renderMultiPanelLayout renders 3 panels, horizontally or vertically based on width.
func (m Model) renderMultiPanelLayout(height int) string {
	// Narrow mode: stack panels vertically
	if m.layoutMode() == LayoutNarrow {
		return m.renderNarrowLayout(height)
	}

	colors := theme.GetSemanticColors()
	devices := m.getFilteredDevices()

	// Split height: top 75% for panels, bottom 25% for energy bars
	topHeight := height * 75 / 100
	energyHeight := height - topHeight - 1 // -1 for gap

	if topHeight < 10 {
		topHeight = 10
	}
	if energyHeight < 5 {
		energyHeight = 5
	}

	// Calculate column widths based on layout mode
	var col1Width, col2Width, col3Width int
	if m.layoutMode() == LayoutWide {
		col1Width = m.width * 25 / 100
		col2Width = m.width * 20 / 100
		col3Width = m.width - col1Width - col2Width - 2
	} else {
		col1Width = m.width * 30 / 100
		col2Width = m.width * 25 / 100
		col3Width = m.width - col1Width - col2Width - 2
	}

	if col1Width < 15 {
		col1Width = 15
	}

	focusBorder := colors.Highlight
	unfocusBorder := colors.TableBorder

	eventsCol := m.renderEventsColumn(col1Width, topHeight)

	listBorder := unfocusBorder
	if m.focusedPanel == PanelDeviceList {
		listBorder = focusBorder
	}
	listCol := m.renderDeviceListColumn(devices, col2Width, topHeight, listBorder)

	infoBorder := unfocusBorder
	if m.focusedPanel == PanelDetail || m.focusedPanel == PanelEndpoints {
		infoBorder = focusBorder
	}
	infoCol := m.renderDeviceInfoColumn(devices, col3Width, topHeight, infoBorder)

	topContent := lipgloss.JoinHorizontal(lipgloss.Top, eventsCol, " ", listCol, " ", infoCol)

	// Ensure topContent fills the full width
	topContentStyle := lipgloss.NewStyle().Width(m.width)
	topContent = topContentStyle.Render(topContent)

	// Energy bars panel at bottom (full width)
	energyPanel := m.renderEnergyPanel(m.width, energyHeight)

	// Combine top and bottom sections
	content := lipgloss.JoinVertical(lipgloss.Left, topContent, energyPanel)

	// JSON overlay (centered on top of content when active)
	if m.focusedPanel == PanelEndpoints {
		jsonOverlay := m.renderJSONOverlay(devices, m.width*2/3, height*2/3)
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

// renderEnergyPanel renders the energy consumption bars panel.
func (m Model) renderEnergyPanel(width, height int) string {
	m.energyBars = m.energyBars.SetSize(width, height)
	return m.energyBars.View()
}

// renderNarrowLayout renders panels stacked vertically for narrow terminals.
func (m Model) renderNarrowLayout(height int) string {
	colors := theme.GetSemanticColors()
	devices := m.getFilteredDevices()
	panelWidth := m.width - 2 // Full width minus borders

	// Divide height into 3 parts: events (20%), list (40%), info (40%)
	eventsHeight := height * 20 / 100
	listHeight := height * 40 / 100
	infoHeight := height - eventsHeight - listHeight - 2 // -2 for gaps

	// Minimum heights
	if eventsHeight < 4 {
		eventsHeight = 4
	}
	if listHeight < 6 {
		listHeight = 6
	}
	if infoHeight < 6 {
		infoHeight = 6
	}

	// Get focused panel border color
	focusBorder := colors.Highlight
	unfocusBorder := colors.TableBorder

	// Row 1: Events panel (compact)
	eventsRow := m.renderEventsColumn(panelWidth, eventsHeight)

	// Row 2: Device List
	listBorder := unfocusBorder
	if m.focusedPanel == PanelDeviceList {
		listBorder = focusBorder
	}
	listRow := m.renderDeviceListColumn(devices, panelWidth, listHeight, listBorder)

	// Row 3: Device Info (always rendered)
	infoBorder := unfocusBorder
	if m.focusedPanel == PanelDetail || m.focusedPanel == PanelEndpoints {
		infoBorder = focusBorder
	}
	infoRow := m.renderDeviceInfoColumn(devices, panelWidth, infoHeight, infoBorder)

	// Stack panels vertically
	content := lipgloss.JoinVertical(lipgloss.Left, eventsRow, listRow, infoRow)

	// JSON overlay (centered on top of content when active)
	if m.focusedPanel == PanelEndpoints {
		jsonOverlay := m.renderJSONOverlay(devices, m.width-4, height*2/3)
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

// renderDeviceListColumn renders just the device list.
func (m Model) renderDeviceListColumn(devices []*cache.DeviceData, width, height int, borderColor color.Color) string {
	colors := theme.GetSemanticColors()

	title := fmt.Sprintf("Devices (%d)", len(devices))
	r := rendering.New(width, height).
		SetTitle(title).
		SetFocused(m.focusedPanel == PanelDeviceList).
		SetFocusColor(borderColor).
		SetBlurColor(borderColor)

	var content strings.Builder

	if len(devices) == 0 {
		if m.cache.IsLoading() {
			content.WriteString(lipgloss.NewStyle().Foreground(theme.Cyan()).Render("Loading..."))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Render("No devices"))
		}
		return r.SetContent(content.String()).Render()
	}

	// Ensure cursor is valid
	if m.cursor >= len(devices) {
		m.cursor = len(devices) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// Show device list
	visibleItems := height - 5
	if visibleItems < 1 {
		visibleItems = 1
	}

	scrollOffset := 0
	if m.cursor >= visibleItems {
		scrollOffset = m.cursor - visibleItems + 1
	}

	endIdx := min(scrollOffset+visibleItems, len(devices))
	for i := scrollOffset; i < endIdx; i++ {
		d := devices[i]
		selected := i == m.cursor
		prefix := unselectedPrefix
		if selected {
			prefix = selectedPrefix
		}

		status := lipgloss.NewStyle().Foreground(colors.Offline).Render("○")
		if !d.Fetched {
			status = lipgloss.NewStyle().Foreground(theme.Yellow()).Render("◐")
		} else if d.Online {
			status = lipgloss.NewStyle().Foreground(theme.Green()).Render("●")
		}

		name := d.Device.Name
		maxLen := width - 10
		if len(name) > maxLen && maxLen > 3 {
			name = name[:maxLen-3] + "..."
		}

		line := fmt.Sprintf("%s%s %s", prefix, status, name)
		if selected {
			content.WriteString(lipgloss.NewStyle().
				Background(colors.AltBackground).
				Foreground(theme.Pink()).
				Bold(true).
				Render(line) + "\n")
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(colors.Text).Render(line) + "\n")
		}
	}

	if len(devices) > visibleItems {
		content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).
			Render(fmt.Sprintf("\n[%d/%d]", m.cursor+1, len(devices))))
	}

	return r.SetContent(content.String()).Render()
}

// renderDeviceInfoColumn renders device info, power metrics, and endpoints.
func (m Model) renderDeviceInfoColumn(devices []*cache.DeviceData, width, height int, borderColor color.Color) string {
	colors := theme.GetSemanticColors()

	r := rendering.New(width, height).
		SetTitle("Device Info").
		SetFocused(m.focusedPanel == PanelDetail).
		SetFocusColor(borderColor).
		SetBlurColor(borderColor)

	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(colors.Highlight).Bold(true)

	if len(devices) == 0 || m.cursor >= len(devices) {
		content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Render("Select a device"))
		return r.SetContent(content.String()).Render()
	}

	d := devices[m.cursor]
	labelStyle := lipgloss.NewStyle().Foreground(theme.Purple())
	valueStyle := lipgloss.NewStyle().Foreground(theme.Cyan()).Bold(true)

	if !d.Fetched {
		content.WriteString(lipgloss.NewStyle().Foreground(theme.Yellow()).Render("Connecting..."))
		return r.SetContent(content.String()).Render()
	}

	if !d.Online {
		content.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render("✗ Offline"))
		return r.SetContent(content.String()).Render()
	}

	// Device info
	if d.Info != nil {
		modelName := m.factory.ShellyService().ModelDisplayName(d.Info.Model)
		content.WriteString(labelStyle.Render(" Model:    ") + valueStyle.Render(modelName) + "\n")
		content.WriteString(labelStyle.Render(" Firmware: ") + valueStyle.Render(d.Info.Firmware) + "\n")
		content.WriteString(labelStyle.Render(" MAC:      ") + valueStyle.Render(d.Info.MAC) + "\n")
	}
	content.WriteString(labelStyle.Render(" Address:  ") + valueStyle.Render(d.Device.Address) + "\n")

	// Power metrics (if available)
	writePowerMetrics(&content, d, titleStyle, labelStyle, valueStyle)

	// Switch states
	if len(d.Switches) > 0 {
		content.WriteString("\n" + titleStyle.Render(" Components") + "\n")
		for i, sw := range d.Switches {
			state := lipgloss.NewStyle().Foreground(colors.Offline).Render("OFF")
			if sw.On {
				state = lipgloss.NewStyle().Foreground(theme.Green()).Bold(true).Render("ON")
			}
			prefix := "  "
			if i == m.componentCursor {
				prefix = "▶ "
			}
			content.WriteString(fmt.Sprintf("%sSwitch %d: %s\n", prefix, sw.ID, state))
		}
		content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Italic(true).
			Render("  h/l: select, t: toggle") + "\n")
	}

	// Endpoints section
	content.WriteString("\n" + titleStyle.Render(" Endpoints") + "\n")
	methods := m.getDeviceMethods(d)
	methodStyle := lipgloss.NewStyle().Foreground(theme.Cyan())

	maxMethods := height - 20
	if maxMethods < 3 {
		maxMethods = 3
	}

	for i, method := range methods {
		if i >= maxMethods {
			content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).
				Render(fmt.Sprintf("  +%d more\n", len(methods)-maxMethods)))
			break
		}
		prefix := "  "
		if i == m.endpointCursor && m.focusedPanel == PanelDetail {
			prefix = "▶ "
		}
		content.WriteString(prefix + methodStyle.Render(method) + "\n")
	}
	content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Italic(true).
		Render("  Enter: view JSON") + "\n")

	return r.SetContent(content.String()).Render()
}

// renderJSONOverlay renders the JSON pop-over that covers the device info column.
func (m Model) renderJSONOverlay(devices []*cache.DeviceData, width, height int) string {
	colors := theme.GetSemanticColors()

	title := "JSON Response"
	if m.endpointMethod != "" {
		title = m.endpointMethod
	}

	r := rendering.New(width, height).
		SetTitle(title).
		SetFocused(true).
		SetFocusColor(theme.Pink()).
		SetBlurColor(theme.Pink())

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Italic(true).
		Render("Esc: close | j/k: scroll") + "\n\n")

	if len(devices) == 0 || m.cursor >= len(devices) {
		return r.SetContent(content.String()).Render()
	}

	d := devices[m.cursor]
	if !d.Online {
		content.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render("Device offline"))
		return r.SetContent(content.String()).Render()
	}

	jsonStyle := lipgloss.NewStyle().Foreground(theme.Green())
	labelStyle := lipgloss.NewStyle().Foreground(theme.Cyan()).Bold(true)
	errorStyle := lipgloss.NewStyle().Foreground(colors.Error)

	// Show loading state
	if m.endpointLoading {
		content.WriteString(lipgloss.NewStyle().Foreground(colors.Warning).Render("Loading..."))
		return r.SetContent(content.String()).Render()
	}

	// Show error if any
	if m.endpointError != nil {
		content.WriteString(errorStyle.Render("Error: " + m.endpointError.Error()))
		return r.SetContent(content.String()).Render()
	}

	// Show fetched endpoint data
	if m.endpointData != nil {
		content.WriteString(labelStyle.Render(m.endpointMethod+":") + "\n")
		if jsonBytes, err := json.MarshalIndent(m.endpointData, "", "  "); err == nil {
			for _, line := range strings.Split(string(jsonBytes), "\n") {
				content.WriteString(jsonStyle.Render(line) + "\n")
			}
		} else {
			content.WriteString(errorStyle.Render("Failed to format JSON: " + err.Error()))
		}
		return r.SetContent(content.String()).Render()
	}

	// Fallback: show cached data if no endpoint fetch in progress
	if d.Info != nil {
		content.WriteString(labelStyle.Render("DeviceInfo (cached):") + "\n")
		jsonData := map[string]any{
			"id":       d.Info.ID,
			"model":    d.Info.Model,
			"firmware": d.Info.Firmware,
			"mac":      d.Info.MAC,
		}
		if jsonBytes, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
			for _, line := range strings.Split(string(jsonBytes), "\n") {
				content.WriteString(jsonStyle.Render(line) + "\n")
			}
		}
	}

	// Show switch states as JSON
	if len(d.Switches) > 0 {
		content.WriteString("\n" + labelStyle.Render("SwitchStatus (cached):") + "\n")
		switchData := make([]map[string]any, len(d.Switches))
		for i, sw := range d.Switches {
			switchData[i] = map[string]any{"id": sw.ID, "output": sw.On}
		}
		if jsonBytes, err := json.MarshalIndent(switchData, "", "  "); err == nil {
			for _, line := range strings.Split(string(jsonBytes), "\n") {
				content.WriteString(jsonStyle.Render(line) + "\n")
			}
		}
	}

	// Show power metrics as JSON
	writePowerMetricsJSON(&content, d, labelStyle, jsonStyle)

	return r.SetContent(content.String()).Render()
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

func formatEnergy(wh float64) string {
	if wh >= 1000000 {
		return fmt.Sprintf("%.2f MWh", wh/1000000)
	}
	if wh >= 1000 {
		return fmt.Sprintf("%.2f kWh", wh/1000)
	}
	return fmt.Sprintf("%.1f Wh", wh)
}

func writePowerMetrics(content *strings.Builder, d *cache.DeviceData, titleStyle, labelStyle, valueStyle lipgloss.Style) {
	if d.Power == 0 && d.Voltage == 0 {
		return
	}
	content.WriteString("\n" + titleStyle.Render(" Power Metrics") + "\n")
	powerStyle := lipgloss.NewStyle().Foreground(theme.Orange()).Bold(true)
	if d.Power != 0 {
		content.WriteString(labelStyle.Render(" Power:   ") + powerStyle.Render(fmt.Sprintf("%.1fW", d.Power)) + "\n")
	}
	if d.Voltage != 0 {
		content.WriteString(labelStyle.Render(" Voltage: ") + valueStyle.Render(fmt.Sprintf("%.1fV", d.Voltage)) + "\n")
	}
	if d.Current != 0 {
		content.WriteString(labelStyle.Render(" Current: ") + valueStyle.Render(fmt.Sprintf("%.2fA", d.Current)) + "\n")
	}
	if d.TotalEnergy != 0 {
		content.WriteString(labelStyle.Render(" Total:   ") + valueStyle.Render(formatEnergy(d.TotalEnergy)) + "\n")
	}
}

func writePowerMetricsJSON(content *strings.Builder, d *cache.DeviceData, labelStyle, jsonStyle lipgloss.Style) {
	if d.Power == 0 && d.Voltage == 0 {
		return
	}
	content.WriteString("\n" + labelStyle.Render("PowerMetrics:") + "\n")
	powerData := map[string]any{}
	if d.Power != 0 {
		powerData["power"] = d.Power
	}
	if d.Voltage != 0 {
		powerData["voltage"] = d.Voltage
	}
	if d.Current != 0 {
		powerData["current"] = d.Current
	}
	if jsonBytes, err := json.MarshalIndent(powerData, "", "  "); err == nil {
		for _, line := range strings.Split(string(jsonBytes), "\n") {
			content.WriteString(jsonStyle.Render(line) + "\n")
		}
	}
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

// fetchSelectedEndpoint returns a command to fetch the currently selected endpoint.
func (m Model) fetchSelectedEndpoint() tea.Cmd {
	devices := m.getFilteredDevices()
	if len(devices) == 0 || m.cursor >= len(devices) {
		return nil
	}

	d := devices[m.cursor]
	if !d.Online {
		return nil
	}

	methods := m.getDeviceMethods(d)
	if m.endpointCursor >= len(methods) {
		return nil
	}

	method := methods[m.endpointCursor]
	address := d.Device.Address
	svc := m.factory.ShellyService()
	ctx := m.ctx
	isGen1 := d.Device.Generation == 1 || (d.Info != nil && d.Info.Generation == 1)

	m.endpointLoading = true
	m.endpointData = nil
	m.endpointError = nil
	m.endpointMethod = method

	return func() tea.Msg {
		var data any
		var err error

		if isGen1 {
			// Gen1 uses REST API - method is actually a path like "/shelly"
			var bytes []byte
			bytes, err = svc.RawGen1Call(ctx, address, method)
			if err == nil {
				// Try to parse as JSON for display
				var jsonData any
				if jsonErr := json.Unmarshal(bytes, &jsonData); jsonErr == nil {
					data = jsonData
				} else {
					data = string(bytes)
				}
			}
		} else {
			// Gen2 uses RPC
			data, err = svc.RawRPC(ctx, address, method, nil)
		}

		return EndpointFetchMsg{
			Method: method,
			Data:   data,
			Err:    err,
		}
	}
}
