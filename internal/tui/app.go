package tui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cmdmode"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/devicedetail"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/devicelist"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/energy"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/events"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/help"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/search"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/statusbar"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/tabs"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
)

// DeviceActionMsg reports the result of a device action.
type DeviceActionMsg struct {
	Device string
	Action string
	Err    error
}

// ViewMode represents the current view in the TUI.
type ViewMode int

const (
	// ViewDevices shows the device list.
	ViewDevices ViewMode = iota
	// ViewMonitor shows real-time monitoring.
	ViewMonitor
	// ViewEvents shows the event stream.
	ViewEvents
	// ViewEnergy shows energy dashboard.
	ViewEnergy
)

// String returns the display name of the view.
func (v ViewMode) String() string {
	switch v {
	case ViewDevices:
		return "Devices"
	case ViewMonitor:
		return "Monitor"
	case ViewEvents:
		return "Events"
	case ViewEnergy:
		return "Energy"
	default:
		return "Unknown"
	}
}

// Model is the main TUI application model.
type Model struct {
	// Core
	ctx     context.Context
	factory *cmdutil.Factory
	cfg     *config.Config
	keys    KeyMap
	styles  Styles

	// State
	currentView ViewMode
	ready       bool
	quitting    bool

	// Dimensions
	width  int
	height int

	// Components
	deviceList   devicelist.Model
	monitor      monitor.Model
	events       events.Model
	energy       energy.Model
	statusBar    statusbar.Model
	tabs         tabs.Model
	search       search.Model
	cmdMode      cmdmode.Model
	help         help.Model
	deviceDetail devicedetail.Model
	toast        toast.Model

	// Settings
	refreshInterval time.Duration
}

// Options configures the TUI.
type Options struct {
	RefreshInterval time.Duration
	InitialView     ViewMode
	Filter          string
}

// DefaultOptions returns default TUI options.
func DefaultOptions() Options {
	return Options{
		RefreshInterval: 5 * time.Second,
		InitialView:     ViewDevices,
	}
}

// New creates a new TUI application.
func New(ctx context.Context, f *cmdutil.Factory, opts Options) Model {
	cfg, err := f.Config()
	if err != nil {
		cfg = nil
	}

	tabNames := []string{"Devices", "Monitor", "Events", "Energy"}
	svc := f.ShellyService()
	ios := f.IOStreams()

	// Create search component with initial filter
	searchModel := search.NewWithFilter(opts.Filter)

	// Load keybindings from config or use defaults
	keys := KeyMapFromConfig(cfg)

	return Model{
		ctx:             ctx,
		factory:         f,
		cfg:             cfg,
		keys:            keys,
		styles:          DefaultStyles(),
		currentView:     opts.InitialView,
		refreshInterval: opts.RefreshInterval,
		deviceList: devicelist.New(devicelist.Deps{
			Ctx:             ctx,
			Svc:             svc,
			IOS:             ios,
			RefreshInterval: opts.RefreshInterval,
		}),
		monitor: monitor.New(monitor.Deps{
			Ctx:             ctx,
			Svc:             svc,
			IOS:             ios,
			RefreshInterval: opts.RefreshInterval,
		}),
		events: events.New(events.Deps{
			Ctx: ctx,
			Svc: svc,
			IOS: ios,
		}),
		energy: energy.New(energy.Deps{
			Ctx:             ctx,
			Svc:             svc,
			IOS:             ios,
			RefreshInterval: opts.RefreshInterval,
		}),
		statusBar: statusbar.New(),
		tabs:      tabs.New(tabNames, int(opts.InitialView)),
		search:    searchModel,
		cmdMode:   cmdmode.New(),
		help:      help.New(),
		deviceDetail: devicedetail.New(devicedetail.Deps{
			Ctx: ctx,
			Svc: svc,
		}),
		toast: toast.New(),
	}
}

// Init initializes the TUI and returns the first command to run.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.deviceList.Init(),
		m.monitor.Init(),
		m.events.Init(),
		m.energy.Init(),
		m.statusBar.Init(),
		m.toast.Init(),
	)
}

// Update handles messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case DeviceActionMsg:
		return m.handleDeviceAction(msg)

	case search.FilterChangedMsg:
		m.deviceList = m.deviceList.SetFilter(msg.Filter)
		return m, nil

	case search.ClosedMsg:
		// Search was closed, nothing special to do
		return m, nil

	case devicedetail.Msg:
		var cmd tea.Cmd
		m.deviceDetail, cmd = m.deviceDetail.Update(msg)
		return m, cmd

	case devicedetail.ClosedMsg:
		// Device detail was closed, nothing special to do
		return m, nil

	case cmdmode.CommandMsg:
		return m.handleCommand(msg)

	case cmdmode.ErrorMsg:
		return m, toast.Error(msg.Message)

	case cmdmode.ClosedMsg:
		// Command mode was closed without executing
		return m, nil

	case tea.KeyPressMsg:
		if newModel, cmd, handled := m.handleKeyPressMsg(msg); handled {
			return newModel, cmd
		}
	}

	// Forward to active component
	var componentCmd tea.Cmd
	m, componentCmd = m.updateActiveComponent(msg)
	if componentCmd != nil {
		cmds = append(cmds, componentCmd)
	}

	// Update status bar
	var statusCmd tea.Cmd
	m.statusBar, statusCmd = m.statusBar.Update(msg)
	cmds = append(cmds, statusCmd)

	// Update toast notifications
	var toastCmd tea.Cmd
	m.toast, toastCmd = m.toast.Update(msg)
	cmds = append(cmds, toastCmd)

	return m, tea.Batch(cmds...)
}

// handleKeyPressMsg handles keyboard input, routing to overlays or main handlers.
func (m Model) handleKeyPressMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	// If device detail is visible, forward all keys to it
	if m.deviceDetail.Visible() {
		var cmd tea.Cmd
		m.deviceDetail, cmd = m.deviceDetail.Update(msg)
		return m, cmd, true
	}

	// If help is visible, close on dismiss keys or forward for scrolling
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
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.ready = true

	headerHeight := 3 // tabs + border
	footerHeight := 2 // status bar + border
	contentHeight := m.height - headerHeight - footerHeight

	m.deviceList = m.deviceList.SetSize(m.width, contentHeight)
	m.monitor = m.monitor.SetSize(m.width, contentHeight)
	m.events = m.events.SetSize(m.width, contentHeight)
	m.energy = m.energy.SetSize(m.width, contentHeight)
	m.statusBar = m.statusBar.SetWidth(m.width)
	m.search = m.search.SetWidth(m.width)
	m.cmdMode = m.cmdMode.SetWidth(m.width)
	m.toast = m.toast.SetSize(m.width, m.height)
	return m, nil
}

// handleDeviceActionKey handles device action key presses (t/o/O/R).
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

// handleDeviceAction handles device action results and updates the status bar and toast.
func (m Model) handleDeviceAction(msg DeviceActionMsg) (tea.Model, tea.Cmd) {
	var statusCmd, toastCmd tea.Cmd
	if msg.Err != nil {
		statusCmd = statusbar.SetMessage(
			msg.Device+": "+msg.Action+" failed - "+msg.Err.Error(),
			statusbar.MessageError,
		)
		toastCmd = toast.Error(msg.Device + ": " + msg.Action + " failed")
	} else {
		statusCmd = statusbar.SetMessage(
			msg.Device+": "+msg.Action+" success",
			statusbar.MessageSuccess,
		)
		toastCmd = toast.Success(msg.Device + ": " + msg.Action + " success")
	}

	// Also refresh the current view to show updated state
	refreshCmd := m.refreshCurrentView()

	return m, tea.Batch(statusCmd, toastCmd, refreshCmd)
}

// handleCommand handles command mode commands.
func (m Model) handleCommand(msg cmdmode.CommandMsg) (tea.Model, tea.Cmd) {
	switch msg.Command {
	case cmdmode.CmdQuit:
		m.quitting = true
		return m, tea.Quit

	case cmdmode.CmdDevice:
		// Jump to device by name - set filter and switch to devices view
		m.deviceList = m.deviceList.SetFilter(msg.Args)
		m.currentView = ViewDevices
		m.tabs = m.tabs.SetActive(int(m.currentView))
		return m, tea.Batch(
			toast.Success("Filter set: "+msg.Args),
			m.refreshCurrentView(),
		)

	case cmdmode.CmdFilter:
		// Set filter on device list
		m.deviceList = m.deviceList.SetFilter(msg.Args)
		m.currentView = ViewDevices
		m.tabs = m.tabs.SetActive(int(m.currentView))
		if msg.Args == "" {
			return m, toast.Success("Filter cleared")
		}
		return m, toast.Success("Filter: " + msg.Args)

	case cmdmode.CmdTheme:
		// Set theme at runtime
		if !theme.SetTheme(msg.Args) {
			return m, toast.Error("Invalid theme: " + msg.Args)
		}
		// Refresh styles
		m.styles = DefaultStyles()
		return m, toast.Success("Theme: " + msg.Args)

	case cmdmode.CmdView:
		// Switch to a specific view
		newView := m.parseView(msg.Args)
		if newView < 0 {
			return m, toast.Error("Unknown view: " + msg.Args)
		}
		m.currentView = ViewMode(newView)
		m.tabs = m.tabs.SetActive(int(m.currentView))
		return m, nil

	case cmdmode.CmdHelp:
		m.help = m.help.SetSize(m.width, m.height)
		m.help = m.help.SetContext(m.currentHelpContext())
		m.help = m.help.Toggle()
		return m, nil

	case cmdmode.CmdToggle:
		// Toggle the selected device
		if cmd := m.executeDeviceAction("toggle"); cmd != nil {
			return m, cmd
		}
		return m, toast.Error("No device selected or device offline")

	default:
		return m, toast.Error("Unknown command")
	}
}

// parseView parses a view name to a ViewMode index.
func (m Model) parseView(name string) int {
	switch name {
	case "devices", "device", "dev", "d", "1":
		return int(ViewDevices)
	case "monitor", "mon", "m", "2":
		return int(ViewMonitor)
	case "events", "event", "e", "3":
		return int(ViewEvents)
	case "energy", "power", "p", "4":
		return int(ViewEnergy)
	default:
		return -1
	}
}

// handleKeyPress handles global key presses and returns whether it was handled.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.ForceQuit):
		m.quitting = true
		return m, tea.Quit, true

	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit, true

	case key.Matches(msg, m.keys.Help):
		m.help = m.help.SetSize(m.width, m.height)
		m.help = m.help.SetContext(m.currentHelpContext())
		m.help = m.help.Toggle()
		return m, nil, true

	case key.Matches(msg, m.keys.Escape):
		// Escape does nothing when no overlay is open
		return m, nil, false

	case key.Matches(msg, m.keys.Refresh):
		return m, m.refreshCurrentView(), true

	case key.Matches(msg, m.keys.Enter):
		// Show device detail in Devices or Monitor view
		if m.currentView == ViewDevices || m.currentView == ViewMonitor {
			if selected := m.deviceList.SelectedDevice(); selected != nil {
				var cmd tea.Cmd
				m.deviceDetail = m.deviceDetail.SetSize(m.width, m.height)
				m.deviceDetail, cmd = m.deviceDetail.Show(selected.Device)
				return m, cmd, true
			}
		}

	case key.Matches(msg, m.keys.Filter):
		// Only allow filter in Devices view
		if m.currentView == ViewDevices {
			var cmd tea.Cmd
			m.search, cmd = m.search.Activate()
			return m, cmd, true
		}

	case key.Matches(msg, m.keys.Command):
		var cmd tea.Cmd
		m.cmdMode, cmd = m.cmdMode.Activate()
		return m, cmd, true
	}

	// Handle device actions (extracted for complexity)
	if newModel, cmd, handled := m.handleDeviceActionKey(msg); handled {
		return newModel, cmd, true
	}

	// Handle view switching
	if newView, ok := m.handleViewSwitch(msg); ok {
		m.currentView = newView
		m.tabs = m.tabs.SetActive(int(m.currentView))
		return m, nil, true
	}

	return m, nil, false
}

// handleViewSwitch checks if the key switches views and returns the new view.
func (m Model) handleViewSwitch(msg tea.KeyPressMsg) (ViewMode, bool) {
	switch {
	case key.Matches(msg, m.keys.Tab):
		return (m.currentView + 1) % 4, true
	case key.Matches(msg, m.keys.ShiftTab):
		return (m.currentView + 3) % 4, true
	case key.Matches(msg, m.keys.View1):
		return ViewDevices, true
	case key.Matches(msg, m.keys.View2):
		return ViewMonitor, true
	case key.Matches(msg, m.keys.View3):
		return ViewEvents, true
	case key.Matches(msg, m.keys.View4):
		return ViewEnergy, true
	default:
		return m.currentView, false
	}
}

// updateActiveComponent forwards messages to the active component.
func (m Model) updateActiveComponent(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentView {
	case ViewDevices:
		m.deviceList, cmd = m.deviceList.Update(msg)
	case ViewMonitor:
		m.monitor, cmd = m.monitor.Update(msg)
	case ViewEvents:
		m.events, cmd = m.events.Update(msg)
	case ViewEnergy:
		m.energy, cmd = m.energy.Update(msg)
	}
	return m, cmd
}

// executeDeviceAction executes a device action on the selected device.
func (m Model) executeDeviceAction(action string) tea.Cmd {
	// Only allow device actions in Devices or Monitor view
	if m.currentView != ViewDevices && m.currentView != ViewMonitor {
		return nil
	}

	// Get selected device
	selected := m.deviceList.SelectedDevice()
	if selected == nil || !selected.Online {
		return nil
	}

	device := selected.Device
	svc := m.factory.ShellyService()

	return func() tea.Msg {
		var err error
		switch action {
		case "toggle":
			_, err = svc.SwitchToggle(m.ctx, device.Address, 0)
		case "on":
			err = svc.SwitchOn(m.ctx, device.Address, 0)
		case "off":
			err = svc.SwitchOff(m.ctx, device.Address, 0)
		case "reboot":
			err = svc.DeviceReboot(m.ctx, device.Address, 0)
		}
		return DeviceActionMsg{
			Device: device.Name,
			Action: action,
			Err:    err,
		}
	}
}

// currentHelpContext returns the help context for the current view.
func (m Model) currentHelpContext() help.ViewContext {
	switch m.currentView {
	case ViewDevices:
		return help.ContextDevices
	case ViewMonitor:
		return help.ContextMonitor
	case ViewEvents:
		return help.ContextEvents
	case ViewEnergy:
		return help.ContextEnergy
	default:
		return help.ContextDevices
	}
}

// refreshCurrentView returns a command to refresh the current view's data.
func (m Model) refreshCurrentView() tea.Cmd {
	switch m.currentView {
	case ViewDevices:
		return m.deviceList.Refresh()
	case ViewMonitor:
		return m.monitor.Refresh()
	case ViewEvents:
		return nil // Events are real-time, no manual refresh
	case ViewEnergy:
		return m.energy.Refresh()
	default:
		return nil
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

	// Header (tabs)
	header := m.tabs.View()

	// Input bar (search or command mode)
	var inputBar string
	if m.search.IsActive() {
		inputBar = m.search.View()
	} else if m.cmdMode.IsActive() {
		inputBar = m.cmdMode.View()
	}

	// Content based on current view
	var content string
	switch m.currentView {
	case ViewDevices:
		content = m.deviceList.View()
	case ViewMonitor:
		content = m.monitor.View()
	case ViewEvents:
		content = m.events.View()
	case ViewEnergy:
		content = m.energy.View()
	}

	// Overlays (help or device detail)
	if m.help.Visible() {
		content = m.help.View()
	} else if m.deviceDetail.Visible() {
		content = m.deviceDetail.View()
	}

	// Footer (status bar)
	footer := m.statusBar.View()

	// Compose the layout
	var result string
	if inputBar != "" {
		result = lipgloss.JoinVertical(lipgloss.Left,
			header,
			inputBar,
			content,
			footer,
		)
	} else {
		result = lipgloss.JoinVertical(lipgloss.Left,
			header,
			content,
			footer,
		)
	}

	// Add toast overlay if there are toasts
	if m.toast.HasToasts() {
		result = m.toast.Overlay(result)
	}

	v := tea.NewView(result)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
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
