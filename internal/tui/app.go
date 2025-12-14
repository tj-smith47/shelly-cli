package tui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/devicelist"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/energy"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/events"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/statusbar"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/tabs"
)

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
	deviceList devicelist.Model
	monitor    monitor.Model
	events     events.Model
	energy     energy.Model
	statusBar  statusbar.Model
	tabs       tabs.Model

	// Settings
	refreshInterval time.Duration
	showHelp        bool
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

	return Model{
		ctx:             ctx,
		factory:         f,
		cfg:             cfg,
		keys:            DefaultKeyMap(),
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
	)
}

// Update handles messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyPressMsg:
		if newModel, cmd, handled := m.handleKeyPress(msg); handled {
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

	return m, tea.Batch(cmds...)
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
	return m, nil
}

// handleKeyPress handles global key presses and returns whether it was handled.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.ForceQuit):
		m.quitting = true
		return m, tea.Quit, true

	case key.Matches(msg, m.keys.Quit):
		if !m.showHelp {
			m.quitting = true
			return m, tea.Quit, true
		}
		m.showHelp = false
		return m, nil, true

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		return m, nil, true

	case key.Matches(msg, m.keys.Escape):
		if m.showHelp {
			m.showHelp = false
			return m, nil, true
		}

	case key.Matches(msg, m.keys.Refresh):
		return m, m.refreshCurrentView(), true
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

	// Help overlay
	if m.showHelp {
		content = m.renderHelp()
	}

	// Footer (status bar)
	footer := m.statusBar.View()

	// Compose the layout
	result := lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		footer,
	)

	v := tea.NewView(result)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m Model) renderHelp() string {
	help := m.styles.Title.Render("Keyboard Shortcuts") + "\n\n"

	bindings := m.keys.FullHelp()
	for _, group := range bindings {
		for _, b := range group {
			help += m.styles.HelpKey.Render(b.Help().Key) + " "
			help += m.styles.HelpDesc.Render(b.Help().Desc) + "  "
		}
		help += "\n"
	}

	help += "\n" + m.styles.HelpDesc.Render("Press ? or Esc to close")

	return m.styles.Border.
		Width(m.width-2).
		Height(m.height-5).
		Align(lipgloss.Center, lipgloss.Center).
		Render(help)
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
