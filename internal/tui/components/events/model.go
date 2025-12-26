// Package events provides the event stream view for the TUI.
package events

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/events"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Deps holds the dependencies for the events component.
type Deps struct {
	Ctx         context.Context
	Svc         *shelly.Service
	IOS         *iostreams.IOStreams
	EventStream *shelly.EventStream // Shared event stream (WebSocket for Gen2+, polling for Gen1)
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
	if d.EventStream == nil {
		return fmt.Errorf("event stream is required")
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
	lastEventTime time.Time // Track when last event was added for refresh detection
}

// RefreshTickMsg triggers a check for new events.
type RefreshTickMsg struct{}

// Model holds the events state.
type Model struct {
	ctx          context.Context
	svc          *shelly.Service
	ios          *iostreams.IOStreams
	eventStream  *shelly.EventStream
	state        *sharedState
	maxItems     int
	width        int
	height       int
	styles       Styles
	scrollOffset int
	cursor       int
	paused       bool
	autoScroll   bool // Auto-scroll to bottom when new events arrive

	// Filtering
	filterByDevice bool   // When true, only show events for selectedDevice
	selectedDevice string // Device name to filter by (when filterByDevice is true)
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
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			Padding(0, 1),
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
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		ios:         deps.IOS,
		eventStream: deps.EventStream,
		maxItems:    100,
		autoScroll:  true, // Start with auto-scroll enabled (tail-f behavior)
		styles:      DefaultStyles(),
		state: &sharedState{
			events:        []Event{},
			subscriptions: make(map[string]context.CancelFunc),
			connStatus:    make(map[string]bool),
		},
	}
}

// Init returns the initial command for events.
func (m Model) Init() tea.Cmd {
	// Subscribe to the shared EventStream
	m.eventStream.Subscribe(m.handleStreamEvent)

	// Add initial "subscribed" event
	m.addEvent(Event{
		Timestamp:   time.Now(),
		Device:      "system",
		Type:        "info",
		Description: "Subscribed to event stream",
	})

	return m.scheduleRefresh()
}

// scheduleRefresh schedules the next refresh tick to check for new events.
func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// handleStreamEvent converts shelly-go events to our internal Event type.
// It is called by the shared EventStream when events arrive.
func (m Model) handleStreamEvent(evt events.Event) {
	// Apply device filter if enabled
	if m.filterByDevice && evt.DeviceID() != m.selectedDevice {
		return
	}

	var event Event
	event.Timestamp = evt.Timestamp()
	event.Device = evt.DeviceID()

	switch e := evt.(type) {
	case *events.StatusChangeEvent:
		event.Component = e.Component
		event.Type = "status"
		event.Description = fmt.Sprintf("%s changed", e.Component)
	case *events.NotifyEvent:
		event.Component = e.Component
		event.Type = categorizeEvent(e.Event)
		event.Description = fmt.Sprintf("%s: %s", e.Component, e.Event)
	case *events.FullStatusEvent:
		event.Type = "full_status"
		event.Description = "Full status update"
	case *events.DeviceOnlineEvent:
		event.Type = "info"
		event.Description = "Device online"
		// Update connection status
		m.state.mu.Lock()
		m.state.connStatus[evt.DeviceID()] = true
		m.state.mu.Unlock()
	case *events.DeviceOfflineEvent:
		event.Type = "warning"
		event.Description = "Device offline"
		// Update connection status
		m.state.mu.Lock()
		m.state.connStatus[evt.DeviceID()] = false
		m.state.mu.Unlock()
	case *events.ScriptEvent:
		event.Type = "script"
		event.Description = fmt.Sprintf("Script output: %s", e.Output)
	case *events.ErrorEvent:
		event.Type = "error"
		event.Description = e.Message
	default:
		event.Type = string(evt.Type())
		event.Description = fmt.Sprintf("Event: %s", evt.Type())
	}

	m.addEvent(event)
	m.ios.DebugCat(iostreams.CategoryNetwork, "events: received %s event from %s",
		event.Type, event.Device)
}

// SetSelectedDevice sets the device to filter events by.
func (m Model) SetSelectedDevice(name string) Model {
	m.selectedDevice = name
	return m
}

// ToggleFilter toggles the device filter on/off.
func (m Model) ToggleFilter() Model {
	m.filterByDevice = !m.filterByDevice
	return m
}

// IsFiltering returns whether device filtering is active.
func (m Model) IsFiltering() bool {
	return m.filterByDevice
}

// FilteredDevice returns the device being filtered by.
func (m Model) FilteredDevice() string {
	return m.selectedDevice
}

func (m Model) addEvent(e Event) {
	if m.paused {
		return
	}
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	// Append new events at the end (tail-f style: newest at bottom)
	m.state.events = append(m.state.events, e)
	if len(m.state.events) > m.maxItems {
		// Remove oldest events from the beginning
		m.state.events = m.state.events[len(m.state.events)-m.maxItems:]
	}
	m.state.lastEventTime = time.Now()
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

// Update handles messages for events.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EventMsg:
		for _, e := range msg.Events {
			m.addEvent(e)
		}
		return m, nil

	case RefreshTickMsg:
		// Auto-scroll to bottom if enabled (tail-f behavior)
		if m.autoScroll {
			count := m.eventCount()
			visibleRows := m.visibleRows()
			if count > visibleRows {
				m.scrollOffset = count - visibleRows
				m.cursor = count - 1
			} else {
				m.scrollOffset = 0
				m.cursor = max(0, count-1)
			}
		}
		// Reschedule next refresh - the tick itself triggers a re-render
		// which will pick up any events added by background goroutines
		return m, m.scheduleRefresh()

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
		m.autoScroll = false // User scrolled to top, disable auto-scroll
	case "G":
		m = m.cursorToBottom()
		m.autoScroll = true // User went to bottom, enable auto-scroll
	case "c":
		m.Clear()
		m.cursor = 0
		m.scrollOffset = 0
		m.autoScroll = true // After clear, re-enable auto-scroll
	case "p":
		m = m.togglePause()
	case "f":
		m = m.ToggleFilter()
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
	// Re-enable auto-scroll if at bottom
	m.autoScroll = m.cursor >= count-1
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		m.autoScroll = false // User is scrolling up, disable auto-scroll
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
	m.autoScroll = true // Going to bottom enables auto-scroll
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
	// Re-enable auto-scroll if at bottom
	m.autoScroll = m.cursor >= count-1
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
	m.autoScroll = false // User is scrolling up, disable auto-scroll
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
// Overhead: header(1) + header margin(1) + footer(1) = 3 lines.
func (m Model) visibleRows() int {
	availableHeight := m.height - 3
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
	eventList := make([]Event, len(m.state.events))
	copy(eventList, m.state.events)
	connStatus := make(map[string]bool)
	for k, v := range m.state.connStatus {
		connStatus[k] = v
	}
	m.state.mu.Unlock()

	if len(eventList) == 0 {
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

	filterStr := ""
	if m.filterByDevice {
		filterStr = "  " + m.styles.Info.Render(fmt.Sprintf("[Filter: %s]", m.selectedDevice))
	}

	header := m.styles.Header.Render("Event Stream") + "  " + statusStr + pauseStr + filterStr

	// Calculate visible events
	visibleRows := m.visibleRows()
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleRows
	if endIdx > len(eventList) {
		endIdx = len(eventList)
	}

	eventsToShow := eventList[startIdx:endIdx]

	// Column widths for fixed-width layout (compact to maximize description)
	const (
		colTime   = 9  // HH:MM:SS + space
		colDevice = 12 // Truncate long device names
		colComp   = 10 // Switch:0, etc.
		colType   = 8  // info/warning/error
	)

	var rows string
	for i, e := range eventsToShow {
		actualIdx := startIdx + i
		isSelected := actualIdx == m.cursor
		rows += m.renderEventRow(e, isSelected, colTime, colDevice, colComp, colType)
	}

	// Footer with scroll info
	scrollInfo := ""
	if len(eventList) > visibleRows {
		scrollInfo = fmt.Sprintf(" [%d-%d of %d]", startIdx+1, endIdx, len(eventList))
	}
	footer := m.styles.Footer.Render(fmt.Sprintf("j/k: scroll  PgUp/PgDn: page  g/G: top/bottom  p: pause  c: clear%s", scrollInfo))

	// Trim trailing newline from rows to avoid extra blank line before footer
	rows = strings.TrimSuffix(rows, "\n")

	content := lipgloss.JoinVertical(lipgloss.Left, header, rows, footer)
	return m.styles.Container.Width(m.width - 4).Render(content)
}

// renderEventRow renders a single event row with fixed-width columns.
func (m Model) renderEventRow(e Event, isSelected bool, colTime, colDevice, colComp, colType int) string {
	// Build each column with fixed width and proper padding
	timeVal := e.Timestamp.Format("15:04:05")
	timeStr := m.styles.Time.Render(output.PadRight(timeVal, colTime))

	deviceVal := output.Truncate(e.Device, colDevice-2)
	deviceStr := m.styles.Device.Render(output.PadRight(deviceVal, colDevice))

	compVal := m.formatComponent(e)
	compVal = output.Truncate(compVal, colComp-2)
	compStr := m.styles.Component.Render(output.PadRight(compVal, colComp))

	typeStyle := m.getTypeStyle(e.Type)
	typeVal := output.Truncate(e.Type, colType-2)
	typeStr := m.styles.Type.Inherit(typeStyle).Render(output.PadRight(typeVal, colType))

	// Calculate remaining width for description (panel width minus other columns and padding)
	descWidth := m.width - colTime - colDevice - colComp - colType - 6 // 6 for borders/padding
	if descWidth < 20 {
		descWidth = 20 // Minimum readable width
	}
	descStr := m.styles.Description.Render(output.Truncate(e.Description, descWidth))

	row := timeStr + deviceStr + compStr + typeStr + descStr

	if isSelected {
		return m.styles.SelectedRow.Render(row) + "\n"
	}
	return m.styles.Event.Render(row) + "\n"
}

// formatComponent formats the component field for display.
func (m Model) formatComponent(e Event) string {
	if e.Component == "" {
		return "-"
	}
	if e.ComponentID > 0 {
		return fmt.Sprintf("%s:%d", e.Component, e.ComponentID)
	}
	return e.Component
}

// getTypeStyle returns the style for an event type.
func (m Model) getTypeStyle(eventType string) lipgloss.Style {
	switch eventType {
	case "error":
		return m.styles.Error
	case "warning":
		return m.styles.Warning
	default:
		return m.styles.Info
	}
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

// OptimalWidth calculates the minimum width needed to display events
// without excessive truncation. Considers fixed columns plus description.
func (m Model) OptimalWidth() int {
	// Fixed column widths from the View function
	const (
		colTime   = 9  // HH:MM:SS + space
		colDevice = 12 // Device name
		colComp   = 10 // Component
		colType   = 8  // Event type
		overhead  = 10 // Borders, padding, separators
	)

	fixedWidth := colTime + colDevice + colComp + colType + overhead

	// Get events safely with locking
	m.state.mu.Lock()
	eventList := make([]Event, len(m.state.events))
	copy(eventList, m.state.events)
	m.state.mu.Unlock()

	if len(eventList) == 0 {
		return fixedWidth + 30 // Default description width
	}

	// Use median description length for optimal width
	totalDescLen := 0
	maxDescLen := 0
	for _, e := range eventList {
		descLen := len(e.Description)
		totalDescLen += descLen
		if descLen > maxDescLen {
			maxDescLen = descLen
		}
	}

	avgDescLen := totalDescLen / len(eventList)

	// Use 75th percentile: average + 25% of difference to max
	optimalDescWidth := avgDescLen + (maxDescLen-avgDescLen)/4

	// Apply constraints on description width
	if optimalDescWidth < 20 {
		optimalDescWidth = 20
	}
	if optimalDescWidth > 60 {
		optimalDescWidth = 60 // Don't let events panel get too wide
	}

	return fixedWidth + optimalDescWidth
}

// MinWidth returns the minimum usable width for the events panel.
func (m Model) MinWidth() int {
	// Minimum to show truncated but readable content
	return 50
}

// MaxDescriptionLen returns the length of the longest event description.
func (m Model) MaxDescriptionLen() int {
	m.state.mu.Lock()
	eventList := make([]Event, len(m.state.events))
	copy(eventList, m.state.events)
	m.state.mu.Unlock()

	maxLen := 0
	for _, e := range eventList {
		if len(e.Description) > maxLen {
			maxLen = len(e.Description)
		}
	}
	return maxLen
}
