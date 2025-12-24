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
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cmdmode"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/events"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/help"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/search"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/statusbar"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
	"github.com/tj-smith47/shelly-cli/internal/tui/views"
)

// DeviceActionMsg reports the result of a device action.
type DeviceActionMsg struct {
	Device string
	Action string
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

	// Dimensions
	width  int
	height int

	// Components
	statusBar statusbar.Model
	search    search.Model
	cmdMode   cmdmode.Model
	help      help.Model
	toast     toast.Model
	events    events.Model // Event stream component

	// Settings
	refreshInterval time.Duration
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
	case cache.DeviceUpdateMsg, cache.AllDevicesLoadedMsg, cache.RefreshTickMsg:
		cmd := m.cache.Update(msg)
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
		m.help = m.help.SetContext(keys.ContextDevices)
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
	// Global actions
	if newModel, cmd, handled := m.handleGlobalKeys(msg); handled {
		return newModel, cmd, true
	}

	// Panel switching with Tab/Shift+Tab
	if newModel, handled := m.handlePanelSwitch(msg); handled {
		return newModel, nil, true
	}

	// Navigation (based on focused panel)
	if newModel, handled := m.handleNavigation(msg); handled {
		return newModel, nil, true
	}

	// Device actions
	if newModel, cmd, handled := m.handleDeviceActionKey(msg); handled {
		return newModel, cmd, true
	}

	return m, nil, false
}

// handlePanelSwitch handles Tab/Shift+Tab for switching panels.
func (m Model) handlePanelSwitch(msg tea.KeyPressMsg) (Model, bool) {
	switch msg.String() {
	case "tab":
		// Cycle through panels: DeviceList -> Detail -> Endpoints -> DeviceList
		switch m.focusedPanel {
		case PanelDeviceList:
			m.focusedPanel = PanelDetail
		case PanelDetail:
			m.focusedPanel = PanelEndpoints
		case PanelEndpoints:
			m.focusedPanel = PanelDeviceList
		}
		return m, true
	case "shift+tab":
		// Reverse cycle
		switch m.focusedPanel {
		case PanelDeviceList:
			m.focusedPanel = PanelEndpoints
		case PanelDetail:
			m.focusedPanel = PanelDeviceList
		case PanelEndpoints:
			m.focusedPanel = PanelDetail
		}
		return m, true
	}
	return m, false
}

// handleGlobalKeys handles quit, help, filter, command, tab switching, and escape keys.
func (m Model) handleGlobalKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.ForceQuit), key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit, true
	case key.Matches(msg, m.keys.Help):
		m.help = m.help.SetSize(m.width, m.height)
		m.help = m.help.SetContext(keys.ContextDevices)
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
	// Tab switching (1-5)
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

// handleNavigation handles j/k/g/G/h/l navigation keys.
func (m Model) handleNavigation(msg tea.KeyPressMsg) (Model, bool) {
	devices := m.getFilteredDevices()
	deviceCount := len(devices)

	switch msg.String() {
	case "j", "down":
		if m.cursor < deviceCount-1 {
			m.cursor++
			m.componentCursor = -1 // Reset component selection when changing device
		}
		return m, true
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.componentCursor = -1 // Reset component selection when changing device
		}
		return m, true
	case "g":
		m.cursor = 0
		m.componentCursor = -1
		return m, true
	case "G":
		if deviceCount > 0 {
			m.cursor = deviceCount - 1
			m.componentCursor = -1
		}
		return m, true
	case "h", "left":
		// Navigate to previous component or back to "all"
		if m.componentCursor > -1 {
			m.componentCursor--
		}
		return m, true
	case "l", "right":
		// Navigate to next component
		if m.cursor < deviceCount && m.cursor >= 0 {
			d := devices[m.cursor]
			maxComponent := len(d.Switches) - 1
			if m.componentCursor < maxComponent {
				m.componentCursor++
			}
		}
		return m, true
	}
	return m, false
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
		inputHeight = 1
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

	v := tea.NewView(result)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
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

	// Calculate column widths based on layout mode
	var col1Width, col2Width, col3Width int
	if m.layoutMode() == LayoutWide {
		// Wide mode: 15% events, 35% list, 50% info
		col1Width = m.width * 15 / 100
		col2Width = m.width * 35 / 100
		col3Width = m.width - col1Width - col2Width - 2 // -2 for gaps
	} else {
		// Standard mode: 20% events, 40% list, 40% info
		col1Width = m.width * 20 / 100
		col2Width = m.width * 40 / 100
		col3Width = m.width - col1Width - col2Width - 2 // -2 for gaps
	}

	if col1Width < 15 {
		col1Width = 15
	}

	// Get focused panel border color
	focusBorder := colors.Highlight
	unfocusBorder := colors.TableBorder

	// Column 1: Events panel (ORANGE border, full height)
	eventsCol := m.renderEventsColumn(col1Width, height)

	// Column 2: Device List
	listBorder := unfocusBorder
	if m.focusedPanel == PanelDeviceList {
		listBorder = focusBorder
	}
	listCol := m.renderDeviceListColumn(devices, col2Width, height, listBorder)

	// Column 3: Device Info OR JSON overlay
	var infoCol string
	if m.focusedPanel == PanelEndpoints {
		// JSON overlay mode
		infoCol = m.renderJSONOverlay(devices, col3Width, height)
	} else {
		// Normal device info
		infoBorder := unfocusBorder
		if m.focusedPanel == PanelDetail {
			infoBorder = focusBorder
		}
		infoCol = m.renderDeviceInfoColumn(devices, col3Width, height, infoBorder)
	}

	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, eventsCol, " ", listCol, " ", infoCol)
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

	// Row 3: Device Info OR JSON overlay
	var infoRow string
	if m.focusedPanel == PanelEndpoints {
		infoRow = m.renderJSONOverlay(devices, panelWidth, infoHeight)
	} else {
		infoBorder := unfocusBorder
		if m.focusedPanel == PanelDetail {
			infoBorder = focusBorder
		}
		infoRow = m.renderDeviceInfoColumn(devices, panelWidth, infoHeight, infoBorder)
	}

	// Stack panels vertically
	return lipgloss.JoinVertical(lipgloss.Left, eventsRow, listRow, infoRow)
}

// renderEventsColumn renders the events column with ORANGE border, full height.
func (m Model) renderEventsColumn(width, height int) string {
	colors := theme.GetSemanticColors()

	// ORANGE border for events column
	orangeBorder := theme.Orange()

	colStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(orangeBorder).
		Width(width - 2).
		Height(height - 2)

	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(orangeBorder).Bold(true)
	content.WriteString(titleStyle.Render(" Events") + "\n")

	eventCount := m.events.EventCount()
	countStyle := lipgloss.NewStyle().Foreground(theme.Green())
	content.WriteString(countStyle.Render(fmt.Sprintf(" %d total", eventCount)) + "\n\n")

	// Get events from the component's internal state
	eventsView := m.events.View()
	if eventsView == "" {
		content.WriteString(lipgloss.NewStyle().Foreground(colors.Muted).Italic(true).Render(" Waiting..."))
	} else {
		m.writeEventLines(&content, eventsView, width, height)
	}

	return colStyle.Render(content.String())
}

func (m Model) writeEventLines(content *strings.Builder, eventsView string, width, height int) {
	lines := strings.Split(eventsView, "\n")
	maxLines := max(height-6, 3)
	for i, line := range lines {
		if i >= maxLines {
			break
		}
		if len(line) > width-6 {
			line = line[:width-9] + "..."
		}
		content.WriteString(line + "\n")
	}
}

// renderDeviceListColumn renders just the device list.
func (m Model) renderDeviceListColumn(devices []*cache.DeviceData, width, height int, borderColor color.Color) string {
	colors := theme.GetSemanticColors()

	colStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2)

	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(colors.Highlight).Bold(true)
	content.WriteString(titleStyle.Render(fmt.Sprintf(" Devices (%d)", len(devices))) + "\n\n")

	if len(devices) == 0 {
		if m.cache.IsLoading() {
			content.WriteString(lipgloss.NewStyle().Foreground(theme.Cyan()).Render(" Loading..."))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Render(" No devices"))
		}
		return colStyle.Render(content.String())
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
			Render(fmt.Sprintf("\n [%d/%d]", m.cursor+1, len(devices))))
	}

	return colStyle.Render(content.String())
}

// renderDeviceInfoColumn renders device info, power metrics, and endpoints.
func (m Model) renderDeviceInfoColumn(devices []*cache.DeviceData, width, height int, borderColor color.Color) string {
	colors := theme.GetSemanticColors()

	colStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2)

	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(colors.Highlight).Bold(true)
	content.WriteString(titleStyle.Render(" Device Info") + "\n\n")

	if len(devices) == 0 || m.cursor >= len(devices) {
		content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Render(" Select a device"))
		return colStyle.Render(content.String())
	}

	d := devices[m.cursor]
	labelStyle := lipgloss.NewStyle().Foreground(theme.Purple())
	valueStyle := lipgloss.NewStyle().Foreground(theme.Cyan()).Bold(true)

	if !d.Fetched {
		content.WriteString(lipgloss.NewStyle().Foreground(theme.Yellow()).Render(" Connecting..."))
		return colStyle.Render(content.String())
	}

	if !d.Online {
		content.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render(" ✗ Offline"))
		return colStyle.Render(content.String())
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

	return colStyle.Render(content.String())
}

// renderJSONOverlay renders the JSON pop-over that covers the device info column.
func (m Model) renderJSONOverlay(devices []*cache.DeviceData, width, height int) string {
	colors := theme.GetSemanticColors()

	// Use pink border for JSON overlay to make it pop
	colStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Pink()).
		Width(width - 2).
		Height(height - 2)

	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(theme.Pink()).Bold(true)
	content.WriteString(titleStyle.Render(" JSON Response") + "\n")
	content.WriteString(lipgloss.NewStyle().Foreground(theme.Purple()).Italic(true).
		Render(" Esc: close | j/k: scroll") + "\n\n")

	if len(devices) == 0 || m.cursor >= len(devices) {
		return colStyle.Render(content.String())
	}

	d := devices[m.cursor]
	if !d.Online {
		content.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render(" Device offline"))
		return colStyle.Render(content.String())
	}

	jsonStyle := lipgloss.NewStyle().Foreground(theme.Green())
	labelStyle := lipgloss.NewStyle().Foreground(theme.Cyan()).Bold(true)

	// Show device info as JSON
	if d.Info != nil {
		content.WriteString(labelStyle.Render("DeviceInfo:") + "\n")
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
		content.WriteString("\n" + labelStyle.Render("SwitchStatus:") + "\n")
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

	return colStyle.Render(content.String())
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

// renderWithHelpOverlay renders main content stacked with help popup at bottom.
func (m Model) renderWithHelpOverlay(content string, contentHeight int) string {
	helpView := m.help.View()
	if helpView == "" {
		return content
	}

	// Get help popup height
	helpHeight := m.help.ViewHeight()

	// Reduce main content height to make room for help
	reducedHeight := contentHeight - helpHeight
	if reducedHeight < 5 {
		reducedHeight = 5
	}

	// Re-render main content at reduced height with multi-panel layout
	mainContent := m.renderMultiPanelLayout(reducedHeight)

	// Stack main content above help popup
	return lipgloss.JoinVertical(lipgloss.Left, mainContent, helpView)
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
	if d.Device.Generation == 1 {
		return []string{
			"Shelly.GetInfo",
			"Shelly.GetStatus",
			"Relay.GetStatus",
			"Meter.GetInfo",
		}
	}
	return []string{
		"Shelly.GetDeviceInfo",
		"Shelly.GetStatus",
		"Shelly.GetConfig",
		"Switch.GetStatus",
		"Switch.GetConfig",
		"PM.GetStatus",
		"EM.GetStatus",
		"Sys.GetStatus",
		"Wifi.GetStatus",
		"Cloud.GetStatus",
		"Script.List",
	}
}
