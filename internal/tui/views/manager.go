// Package views provides view management for the TUI.
// It handles switching between different views (Dashboard, Automation, Config, Manage, Fleet)
// and provides a consistent interface for rendering view content.
package views

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// ViewID identifies a view.
type ViewID = tabs.TabID

const (
	// ViewDashboard is the main dashboard view showing devices, events, and status.
	ViewDashboard = tabs.TabDashboard
	// ViewAutomation is the automation view for scripts, schedules, and webhooks.
	ViewAutomation = tabs.TabAutomation
	// ViewConfig is the configuration view for WiFi, system, cloud, and inputs.
	ViewConfig = tabs.TabConfig
	// ViewManage is the local device management view for discovery, batch, firmware, and backup.
	ViewManage = tabs.TabManage
	// ViewMonitor is the real-time monitoring view.
	ViewMonitor = tabs.TabMonitor
	// ViewFleet is the Shelly Cloud Fleet management view.
	ViewFleet = tabs.TabFleet
)

// View represents a renderable view.
type View interface {
	// Init returns the initial command for the view.
	Init() tea.Cmd
	// Update handles messages for the view.
	Update(msg tea.Msg) (View, tea.Cmd)
	// View renders the view.
	View() string
	// SetSize sets the view dimensions.
	SetSize(width, height int) View
	// ID returns the view ID.
	ID() ViewID
}

// ViewChangedMsg is sent when the active view changes.
type ViewChangedMsg struct {
	Previous ViewID
	Current  ViewID
}

// ViewFocusChangedMsg is sent when the view's overall focus changes.
// When Focused is false, the device list sidebar has focus.
// When Focused is true, the view content has focus.
type ViewFocusChangedMsg struct {
	Focused bool
}

// ReturnFocusMsg is sent by views when Tab/Shift+Tab should return focus
// to the device list (i.e., when cycling past the first or last panel).
type ReturnFocusMsg struct{}

// Manager manages multiple views and their transitions.
type Manager struct {
	views   map[ViewID]View
	active  ViewID
	width   int
	height  int
	history []ViewID // Navigation history for back navigation
}

// New creates a new view manager.
func New() *Manager {
	return &Manager{
		views:   make(map[ViewID]View),
		active:  ViewDashboard,
		history: make([]ViewID, 0, 10),
	}
}

// Register registers a view with the manager.
func (m *Manager) Register(v View) *Manager {
	m.views[v.ID()] = v
	return m
}

// SetActive sets the active view by ID.
// Returns a command that emits ViewChangedMsg.
func (m *Manager) SetActive(id ViewID) tea.Cmd {
	if m.active == id {
		return nil
	}

	// Add current view to history
	if m.active != id {
		m.history = append(m.history, m.active)
		// Keep history limited
		if len(m.history) > 10 {
			m.history = m.history[1:]
		}
	}

	prev := m.active
	m.active = id

	return func() tea.Msg {
		return ViewChangedMsg{Previous: prev, Current: id}
	}
}

// Active returns the currently active view ID.
func (m *Manager) Active() ViewID {
	return m.active
}

// ActiveView returns the currently active view.
func (m *Manager) ActiveView() View {
	return m.views[m.active]
}

// Get returns a view by ID.
func (m *Manager) Get(id ViewID) View {
	return m.views[id]
}

// Back navigates to the previous view in history.
// Returns a command that emits ViewChangedMsg, or nil if no history.
func (m *Manager) Back() tea.Cmd {
	if len(m.history) == 0 {
		return nil
	}

	// Pop from history
	lastIdx := len(m.history) - 1
	prev := m.history[lastIdx]
	m.history = m.history[:lastIdx]

	curr := m.active
	m.active = prev

	return func() tea.Msg {
		return ViewChangedMsg{Previous: curr, Current: prev}
	}
}

// SetSize updates dimensions for all registered views.
func (m *Manager) SetSize(width, height int) {
	m.width = width
	m.height = height
	for id, v := range m.views {
		m.views[id] = v.SetSize(width, height)
	}
}

// Init initializes all registered views.
func (m *Manager) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, v := range m.views {
		if cmd := v.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// Update forwards messages to the active view.
func (m *Manager) Update(msg tea.Msg) tea.Cmd {
	if v := m.views[m.active]; v != nil {
		newV, cmd := v.Update(msg)
		m.views[m.active] = newV
		return cmd
	}
	return nil
}

// UpdateAll forwards messages to ALL views (not just active).
// Use this for async messages that may need to reach non-active views
// (e.g., StatusLoadedMsg for views that started a fetch before being deactivated).
func (m *Manager) UpdateAll(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for id, v := range m.views {
		if v != nil {
			newV, cmd := v.Update(msg)
			m.views[id] = newV
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// View renders the active view.
func (m *Manager) View() string {
	if v := m.views[m.active]; v != nil {
		return v.View()
	}
	return ""
}

// ViewCount returns the number of registered views.
func (m *Manager) ViewCount() int {
	return len(m.views)
}

// HasHistory returns true if there's navigation history.
func (m *Manager) HasHistory() bool {
	return len(m.history) > 0
}

// HistoryCount returns the number of items in navigation history.
func (m *Manager) HistoryCount() int {
	return len(m.history)
}

// Width returns the current width.
func (m *Manager) Width() int {
	return m.width
}

// Height returns the current height.
func (m *Manager) Height() int {
	return m.height
}

// DeviceSelectedMsg is emitted when a device is selected in the Dashboard view.
// Other views can respond to this to load data for the selected device.
type DeviceSelectedMsg struct {
	Device  string
	Address string
}

// DeviceSettable is an interface for views that can receive a device selection.
type DeviceSettable interface {
	SetDevice(device string) tea.Cmd
}

// PropagateDevice sends the selected device to all views that support it.
// This enables context propagation from Dashboard to Automation, Config, etc.
func (m *Manager) PropagateDevice(device string) tea.Cmd {
	var cmds []tea.Cmd

	for _, v := range m.views {
		if settable, ok := v.(DeviceSettable); ok {
			if cmd := settable.SetDevice(device); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}
