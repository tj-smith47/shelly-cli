// Package events provides the event stream view for the TUI.
package events

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Deps holds the dependencies for the events component.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
	IOS *iostreams.IOStreams
}

// validate ensures all required dependencies are set.
func (d Deps) validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	if d.IOS == nil {
		return fmt.Errorf("iostreams is required")
	}
	return nil
}

// Event represents a device event.
type Event struct {
	Timestamp   time.Time
	Device      string
	Component   string
	ComponentID int
	Type        string
	Description string
	Data        map[string]any
}

// EventMsg is sent when new events are received.
type EventMsg struct {
	Events []Event
}

// SubscriptionStatusMsg indicates WebSocket subscription status.
type SubscriptionStatusMsg struct {
	Device    string
	Connected bool
	Error     error
}

// sharedState holds mutex-protected state that persists across model copies.
type sharedState struct {
	mu            sync.Mutex
	events        []Event
	subscriptions map[string]context.CancelFunc
	connStatus    map[string]bool
}

// Model holds the events state.
type Model struct {
	ctx          context.Context
	svc          *shelly.Service
	ios          *iostreams.IOStreams
	state        *sharedState
	maxItems     int
	width        int
	height       int
	styles       Styles
	scrollOffset int
	cursor       int
	paused       bool
}

// Styles for the events component.
type Styles struct {
	Container    lipgloss.Style
	Header       lipgloss.Style
	Event        lipgloss.Style
	SelectedRow  lipgloss.Style
	Time         lipgloss.Style
	Device       lipgloss.Style
	Component    lipgloss.Style
	Type         lipgloss.Style
	Description  lipgloss.Style
	Info         lipgloss.Style
	Warning      lipgloss.Style
	Error        lipgloss.Style
	Connected    lipgloss.Style
	Disconnected lipgloss.Style
	Footer       lipgloss.Style
}

// DefaultStyles returns default styles for events.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(1, 2),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight).
			MarginBottom(1),
		Event: lipgloss.NewStyle().
			MarginBottom(0),
		SelectedRow: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Bold(true),
		Time: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(10),
		Device: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Width(16),
		Component: lipgloss.NewStyle().
			Foreground(colors.Primary).
			Width(12),
		Type: lipgloss.NewStyle().
			Width(14),
		Description: lipgloss.NewStyle().
			Foreground(colors.Text),
		Info: lipgloss.NewStyle().
			Foreground(colors.Secondary),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Connected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disconnected: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Footer: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// New creates a new events model.
func New(deps Deps) Model {
	if err := deps.validate(); err != nil {
		panic(fmt.Sprintf("events: invalid deps: %v", err))
	}

	return Model{
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		ios:      deps.IOS,
		maxItems: 100,
		styles:   DefaultStyles(),
		state: &sharedState{
			events:        []Event{},
			subscriptions: make(map[string]context.CancelFunc),
			connStatus:    make(map[string]bool),
		},
	}
}

// Init returns the initial command for events.
func (m Model) Init() tea.Cmd {
	return m.subscribeToAllDevices()
}

// subscribeToAllDevices creates WebSocket subscriptions for all registered devices.
// Note: Gen1 devices are skipped as they don't support WebSocket events.
func (m Model) subscribeToAllDevices() tea.Cmd {
	return func() tea.Msg {
		deviceMap := config.ListDevices()
		if len(deviceMap) == 0 {
			return nil
		}

		var events []Event
		for _, d := range deviceMap {
			// Skip Gen1 devices - they don't support WebSocket events
			if d.Generation == 1 {
				continue
			}
			// Start subscription in background
			go m.subscribeToDevice(d)
			// Add connection event
			events = append(events, Event{
				Timestamp:   time.Now(),
				Device:      d.Name,
				Type:        "info",
				Description: "Connecting to WebSocket...",
			})
		}

		return EventMsg{Events: events}
	}
}

// subscribeToDevice creates a WebSocket subscription for a single device.
func (m Model) subscribeToDevice(device model.Device) {
	m.state.mu.Lock()
	// Cancel existing subscription if any
	if cancel, exists := m.state.subscriptions[device.Address]; exists {
		cancel()
	}
	// Create new cancellable context
	ctx, cancel := context.WithCancel(m.ctx)
	m.state.subscriptions[device.Address] = cancel
	m.state.mu.Unlock()

	handler := func(evt shelly.DeviceEvent) error {
		event := Event{
			Timestamp:   evt.Timestamp,
			Device:      device.Name,
			Component:   evt.Component,
			ComponentID: evt.ComponentID,
			Type:        categorizeEvent(evt.Event),
			Description: formatEventDescription(evt),
			Data:        evt.Data,
		}

		// Update connection status
		m.state.mu.Lock()
		m.state.connStatus[device.Address] = true
		m.state.mu.Unlock()

		// Add event to the list
		m.addEvent(event)
		return nil
	}

	// Mark as connected when subscription starts successfully
	m.state.mu.Lock()
	m.state.connStatus[device.Address] = true
	m.state.mu.Unlock()

	m.addEvent(Event{
		Timestamp:   time.Now(),
		Device:      device.Name,
		Type:        "info",
		Description: "WebSocket connected",
	})

	// This blocks until context is cancelled or error occurs
	err := m.svc.SubscribeEvents(ctx, device.Address, handler)

	// Mark as disconnected on exit
	m.state.mu.Lock()
	m.state.connStatus[device.Address] = false
	m.state.mu.Unlock()

	if err != nil && !errors.Is(err, context.Canceled) {
		m.addEvent(Event{
			Timestamp:   time.Now(),
			Device:      device.Name,
			Type:        "error",
			Description: fmt.Sprintf("WebSocket disconnected: %v", err),
		})
	}
}

func (m Model) addEvent(e Event) {
	if m.paused {
		return
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	m.state.events = append([]Event{e}, m.state.events...)
	if len(m.state.events) > m.maxItems {
		m.state.events = m.state.events[:m.maxItems]
	}
}

// categorizeEvent determines the event type category.
func categorizeEvent(eventName string) string {
	switch eventName {
	case "NotifyStatus", "NotifyEvent":
		return "status"
	case "NotifyFullStatus":
		return "full_status"
	default:
		if eventName != "" {
			return eventName
		}
		return "info"
	}
}

// formatEventDescription creates a human-readable description of an event.
func formatEventDescription(evt shelly.DeviceEvent) string {
	if evt.Component != "" {
		if evt.ComponentID > 0 {
			return fmt.Sprintf("%s:%d %s", evt.Component, evt.ComponentID, evt.Event)
		}
		return fmt.Sprintf("%s %s", evt.Component, evt.Event)
	}
	return evt.Event
}

// Update handles messages for events.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EventMsg:
		for _, e := range msg.Events {
			m.addEvent(e)
		}
		return m, nil

	case SubscriptionStatusMsg:
		return m.handleSubscriptionStatus(msg), nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg), nil
	}

	return m, nil
}

// handleSubscriptionStatus handles WebSocket subscription status updates.
func (m Model) handleSubscriptionStatus(msg SubscriptionStatusMsg) Model {
	m.state.mu.Lock()
	m.state.connStatus[msg.Device] = msg.Connected
	m.state.mu.Unlock()

	if msg.Error != nil {
		m.addEvent(Event{
			Timestamp:   time.Now(),
			Device:      msg.Device,
			Type:        "error",
			Description: msg.Error.Error(),
		})
	}
	return m
}

// handleKeyPress handles keyboard input for scrolling and clearing.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "pgdown", "ctrl+d":
		m = m.pageDown()
	case "pgup", "ctrl+u":
		m = m.pageUp()
	case "g":
		m.cursor = 0
		m.scrollOffset = 0
	case "G":
		m = m.cursorToBottom()
	case "c":
		m.Clear()
		m.cursor = 0
		m.scrollOffset = 0
	case "p":
		m = m.togglePause()
	}
	return m
}

func (m Model) cursorDown() Model {
	count := m.eventCount()
	if count == 0 {
		return m
	}
	if m.cursor < count-1 {
		m.cursor++
	}
	visibleRows := m.visibleRows()
	if m.cursor >= m.scrollOffset+visibleRows {
		m.scrollOffset = m.cursor - visibleRows + 1
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	return m
}

func (m Model) cursorToBottom() Model {
	count := m.eventCount()
	if count > 0 {
		m.cursor = count - 1
		visibleRows := m.visibleRows()
		maxOffset := count - visibleRows
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageDown() Model {
	count := m.eventCount()
	if count == 0 {
		return m
	}
	visibleRows := m.visibleRows()
	m.cursor += visibleRows
	if m.cursor >= count {
		m.cursor = count - 1
	}
	if m.cursor >= m.scrollOffset+visibleRows {
		m.scrollOffset = m.cursor - visibleRows + 1
	}
	maxOffset := count - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageUp() Model {
	visibleRows := m.visibleRows()
	m.cursor -= visibleRows
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	return m
}

func (m Model) togglePause() Model {
	m.paused = !m.paused
	return m
}

// eventCount returns the number of events (thread-safe).
func (m Model) eventCount() int {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	return len(m.state.events)
}

// visibleRows calculates how many rows can be displayed.
func (m Model) visibleRows() int {
	availableHeight := m.height - 6
	if availableHeight < 1 {
		return 10
	}
	return availableHeight
}

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// View renders the events.
func (m Model) View() string {
	m.state.mu.Lock()
	events := make([]Event, len(m.state.events))
	copy(events, m.state.events)
	connStatus := make(map[string]bool)
	for k, v := range m.state.connStatus {
		connStatus[k] = v
	}
	m.state.mu.Unlock()

	if len(events) == 0 {
		return m.styles.Container.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No events yet.\nEvents will appear here as devices report state changes.\n\nPress 'c' to clear events when they appear.")
	}

	// Connection status summary
	connected, disconnected := 0, 0
	for _, status := range connStatus {
		if status {
			connected++
		} else {
			disconnected++
		}
	}

	statusStr := m.styles.Connected.Render(fmt.Sprintf("● %d connected", connected))
	if disconnected > 0 {
		statusStr += "  " + m.styles.Disconnected.Render(fmt.Sprintf("○ %d disconnected", disconnected))
	}

	pauseStr := ""
	if m.paused {
		pauseStr = "  " + m.styles.Warning.Render("[PAUSED]")
	}

	header := m.styles.Header.Render("Event Stream") + "  " + statusStr + pauseStr

	// Calculate visible events
	visibleRows := m.visibleRows()
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleRows
	if endIdx > len(events) {
		endIdx = len(events)
	}

	eventsToShow := events[startIdx:endIdx]

	// Column widths for fixed-width layout
	const (
		colTime   = 10 // HH:MM:SS + space
		colDevice = 16
		colComp   = 12
		colType   = 14
	)

	var rows string
	for i, e := range eventsToShow {
		actualIdx := startIdx + i
		isSelected := actualIdx == m.cursor

		// Build each column with fixed width and proper padding
		timeVal := e.Timestamp.Format("15:04:05")
		timeStr := m.styles.Time.Render(output.PadRight(timeVal, colTime))

		deviceVal := output.Truncate(e.Device, colDevice-2)
		deviceStr := m.styles.Device.Render(output.PadRight(deviceVal, colDevice))

		var compVal string
		if e.Component != "" {
			if e.ComponentID > 0 {
				compVal = fmt.Sprintf("%s:%d", e.Component, e.ComponentID)
			} else {
				compVal = e.Component
			}
		} else {
			compVal = "-"
		}
		compVal = output.Truncate(compVal, colComp-2)
		compStr := m.styles.Component.Render(output.PadRight(compVal, colComp))

		var typeStyle lipgloss.Style
		switch e.Type {
		case "error":
			typeStyle = m.styles.Error
		case "warning":
			typeStyle = m.styles.Warning
		default:
			typeStyle = m.styles.Info
		}
		typeVal := output.Truncate(e.Type, colType-2)
		typeStr := m.styles.Type.Inherit(typeStyle).Render(output.PadRight(typeVal, colType))

		descStr := m.styles.Description.Render(output.Truncate(e.Description, 40))

		row := timeStr + deviceStr + compStr + typeStr + descStr

		if isSelected {
			rows += m.styles.SelectedRow.Render(row) + "\n"
		} else {
			rows += m.styles.Event.Render(row) + "\n"
		}
	}

	// Footer with scroll info
	scrollInfo := ""
	if len(events) > visibleRows {
		scrollInfo = fmt.Sprintf(" [%d-%d of %d]", startIdx+1, endIdx, len(events))
	}
	footer := m.styles.Footer.Render(fmt.Sprintf("j/k: scroll  PgUp/PgDn: page  g/G: top/bottom  p: pause  c: clear%s", scrollInfo))

	content := lipgloss.JoinVertical(lipgloss.Left, header, rows, footer)
	return m.styles.Container.Width(m.width - 4).Render(content)
}

// EventCount returns the number of events.
func (m Model) EventCount() int {
	return m.eventCount()
}

// Clear clears all events.
func (m Model) Clear() {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	m.state.events = []Event{}
	// Note: scrollOffset is reset by the caller when the model is updated
}

// Cleanup cancels all WebSocket subscriptions.
func (m Model) Cleanup() {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	for _, cancel := range m.state.subscriptions {
		cancel()
	}
	m.state.subscriptions = make(map[string]context.CancelFunc)
}
