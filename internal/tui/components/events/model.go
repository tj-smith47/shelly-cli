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
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// Event level constants.
const (
	levelStatus     = "status"
	levelFullStatus = "full_status"
	levelInfo       = "info"
	levelWarning    = "warning"
	levelError      = "error"
	levelScript     = "script"

	// Minimum width to show dual columns.
	dualColumnMinWidth = 100
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
	ctx         context.Context
	svc         *shelly.Service
	ios         *iostreams.IOStreams
	eventStream *shelly.EventStream
	state       *sharedState
	scroller    *panel.Scroller
	maxItems    int
	width       int
	height      int
	styles      Styles
	paused      bool
	autoScroll  bool // Auto-scroll to top when new events arrive (newest at top)

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
		scroller:    panel.NewScroller(0, 10), // Will be updated by SetSize
		maxItems:    100,
		autoScroll:  true, // Start with auto-scroll enabled (cursor stays at top/newest)
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
		event.Type = levelStatus
		event.Description = fmt.Sprintf("%s changed", e.Component)
	case *events.NotifyEvent:
		event.Component = e.Component
		event.Type = categorizeEvent(e.Event)
		event.Description = fmt.Sprintf("%s: %s", e.Component, e.Event)
	case *events.FullStatusEvent:
		event.Type = levelFullStatus
		event.Description = "Full status update"
	case *events.DeviceOnlineEvent:
		event.Type = levelInfo
		event.Description = "Device online"
		// Update connection status
		m.state.mu.Lock()
		m.state.connStatus[evt.DeviceID()] = true
		m.state.mu.Unlock()
	case *events.DeviceOfflineEvent:
		event.Type = levelWarning
		event.Description = "Device offline"
		// Update connection status
		m.state.mu.Lock()
		m.state.connStatus[evt.DeviceID()] = false
		m.state.mu.Unlock()
	case *events.ScriptEvent:
		event.Type = levelScript
		event.Description = fmt.Sprintf("Script output: %s", e.Output)
	case *events.ErrorEvent:
		event.Type = levelError
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
	// Prepend new events at the beginning (newest at top)
	m.state.events = append([]Event{e}, m.state.events...)
	if len(m.state.events) > m.maxItems {
		// Remove oldest events from the end
		m.state.events = m.state.events[:m.maxItems]
	}
	m.state.lastEventTime = time.Now()
}

// categorizeEvent determines the event type category.
func categorizeEvent(eventName string) string {
	switch eventName {
	case "NotifyStatus", "NotifyEvent":
		return levelStatus
	case "NotifyFullStatus":
		return levelFullStatus
	default:
		if eventName != "" {
			return eventName
		}
		return levelInfo
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
		// Update scroller with current event count
		m.scroller.SetItemCount(m.eventCount())
		// Auto-scroll to top if enabled (newest events appear at top)
		if m.autoScroll {
			m.scroller.CursorToStart()
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
		m.scroller.CursorDown()
		m.autoScroll = m.scroller.Cursor() == 0
	case "k", "up":
		m.scroller.CursorUp()
		m.autoScroll = m.scroller.Cursor() == 0
	case "pgdown", "ctrl+d":
		m.scroller.PageDown()
		m.autoScroll = false
	case "pgup", "ctrl+u":
		m.scroller.PageUp()
		m.autoScroll = m.scroller.Cursor() == 0
	case "g":
		m.scroller.CursorToStart()
		m.autoScroll = true // User went to top (newest), enable auto-scroll
	case "G":
		m.scroller.CursorToEnd()
		m.autoScroll = false // User went to bottom (oldest), disable auto-scroll
	case "c":
		m.Clear()
		m.scroller.SetItemCount(0)
		m.autoScroll = true // After clear, re-enable auto-scroll
	case "p":
		m = m.togglePause()
	case "f":
		m = m.ToggleFilter()
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

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Reserve 1 row for header
	visibleRows := height - 1
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	return m
}

// View renders the events (content only - wrapper handles title/footer).
func (m Model) View() string {
	m.state.mu.Lock()
	eventList := make([]Event, len(m.state.events))
	copy(eventList, m.state.events)
	m.state.mu.Unlock()

	if len(eventList) == 0 {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(m.styles.Footer.GetForeground()).
			Render("Waiting for events...")
	}

	// Use dual column layout when width permits
	if m.width >= dualColumnMinWidth {
		return m.renderDualColumns(eventList)
	}

	// Single column layout (original behavior)
	return m.renderSingleColumn(eventList)
}

// renderSingleColumn renders events in a single column (original layout).
func (m Model) renderSingleColumn(eventList []Event) string {
	// Get visible range from scroller
	startIdx, endIdx := m.scroller.VisibleRange()
	if endIdx > len(eventList) {
		endIdx = len(eventList)
	}

	eventsToShow := eventList[startIdx:endIdx]

	// Column widths - account for header text + content + padding
	// Header text: TIME(4), DEVICE(6), COMPONENT(9), LEVEL(5), DESCRIPTION
	colTime := 10  // HH:MM:SS (8) + 2 padding
	colLevel := 13 // "full_status" (11) + 2 padding

	// Measure actual device and component widths from visible events
	maxDevice := 6 // Min for "DEVICE" header
	maxComp := 9   // Min for "COMPONENT" header
	for _, e := range eventsToShow {
		if len(e.Device) > maxDevice {
			maxDevice = len(e.Device)
		}
		compStr := m.formatComponent(e)
		if len(compStr) > maxComp {
			maxComp = len(compStr)
		}
	}
	// Add padding and cap
	colDevice := min(maxDevice+2, 20) // Cap at 20
	colComp := min(maxComp+2, 16)     // Cap at 16 (needs to fit "COMPONENT")

	// Render header row
	header := m.renderHeaderRow(colTime, colDevice, colComp, colLevel)

	var rows string
	for i, e := range eventsToShow {
		actualIdx := startIdx + i
		isSelected := m.scroller.IsCursorAt(actualIdx)
		rows += m.renderEventRow(e, isSelected, colTime, colDevice, colComp, colLevel)
	}

	// Trim trailing newline
	rows = strings.TrimSuffix(rows, "\n")

	return header + rows
}

// renderEventRow renders a single event row with dynamic column widths.
func (m Model) renderEventRow(e Event, isSelected bool, colTime, colDevice, colComp, colLevel int) string {
	// Build each column with proper width
	timeVal := e.Timestamp.Format("15:04:05")
	timeStr := m.styles.Time.Render(output.PadRight(timeVal, colTime))

	deviceStr := m.styles.Device.Render(output.PadRight(e.Device, colDevice))

	compVal := m.formatComponent(e)
	compStr := m.styles.Component.Render(output.PadRight(compVal, colComp))

	// Level column with color
	levelStyle := m.getTypeStyle(e.Type)
	levelStr := levelStyle.Render(output.PadRight(e.Type, colLevel))

	// Description gets all remaining width
	fixedWidth := colTime + colDevice + colComp + colLevel
	descWidth := m.width - fixedWidth - 2 // 2 for minimal padding
	if descWidth < 20 {
		descWidth = 20
	}

	descStr := m.styles.Description.Render(output.Truncate(e.Description, descWidth))

	row := timeStr + deviceStr + compStr + levelStr + descStr

	if isSelected {
		return m.styles.SelectedRow.Render(row) + "\n"
	}
	return m.styles.Event.Render(row) + "\n"
}

// renderHeaderRow renders the column header row.
func (m Model) renderHeaderRow(colTime, colDevice, colComp, colLevel int) string {
	// Use only bold+foreground for header cells, no margin (margin is per-cell)
	colors := theme.GetSemanticColors()
	cellStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Highlight)

	timeStr := cellStyle.Render(output.PadRight("TIME", colTime))
	deviceStr := cellStyle.Render(output.PadRight("DEVICE", colDevice))
	compStr := cellStyle.Render(output.PadRight("COMPONENT", colComp))
	levelStr := cellStyle.Render(output.PadRight("LEVEL", colLevel))
	descStr := cellStyle.Render("DESCRIPTION")

	return timeStr + deviceStr + compStr + levelStr + descStr + "\n"
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
	colors := theme.GetSemanticColors()
	switch eventType {
	case levelError:
		return lipgloss.NewStyle().Foreground(colors.Error)
	case levelWarning:
		return lipgloss.NewStyle().Foreground(colors.Warning)
	case levelStatus:
		return lipgloss.NewStyle().Foreground(colors.Success)
	case levelFullStatus:
		return lipgloss.NewStyle().Foreground(colors.Primary)
	case levelScript:
		return lipgloss.NewStyle().Foreground(colors.Secondary)
	case levelInfo:
		return lipgloss.NewStyle().Foreground(colors.Online)
	default:
		return lipgloss.NewStyle().Foreground(colors.Text)
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

// FooterText returns the keybindings for the wrapper to display.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:new/old p:pause c:clear"
}

// StatusBadge returns status indicators for the wrapper to display.
// Returns event count and any active flags (paused, filtered).
func (m Model) StatusBadge() string {
	m.state.mu.Lock()
	eventCount := len(m.state.events)
	m.state.mu.Unlock()

	parts := []string{fmt.Sprintf("%d", eventCount)}

	if m.paused {
		parts = append(parts, "PAUSED")
	}
	if m.filterByDevice {
		parts = append(parts, m.selectedDevice)
	}

	return strings.Join(parts, " | ")
}

// ScrollInfo returns scroll position info for the wrapper.
func (m Model) ScrollInfo() string {
	return m.scroller.ScrollInfo()
}

// IsPaused returns whether event collection is paused.
func (m Model) IsPaused() bool {
	return m.paused
}

// isUserEvent returns true if the event is user-relevant (not system noise).
// User events: status changes, online/offline, errors, script output, warnings.
// System events: full_status updates (frequent polling noise).
func isUserEvent(e Event) bool {
	switch e.Type {
	case levelFullStatus:
		return false // System noise - exclude from user column
	case levelStatus, levelInfo, levelWarning, levelError, levelScript:
		return true // User-relevant events
	default:
		return true // Unknown types go to user column
	}
}

// splitEventsByType splits events into user and system categories.
func splitEventsByType(eventList []Event) (user, system []Event) {
	user = make([]Event, 0, len(eventList)/2)
	system = make([]Event, 0, len(eventList)/2)
	for _, e := range eventList {
		if isUserEvent(e) {
			user = append(user, e)
		} else {
			system = append(system, e)
		}
	}
	return user, system
}

// renderDualColumns renders events in two columns: User (left) and System (right).
func (m Model) renderDualColumns(eventList []Event) string {
	colors := theme.GetSemanticColors()

	// Split events by type
	userEvents, systemEvents := splitEventsByType(eventList)

	// Calculate column widths
	colWidth := (m.width - 3) / 2 // -3 for separator (│) and spacing
	visibleRows := m.height - 2   // -2 for headers

	if visibleRows < 1 {
		visibleRows = 1
	}

	// Limit events to visible rows
	if len(userEvents) > visibleRows {
		userEvents = userEvents[:visibleRows]
	}
	if len(systemEvents) > visibleRows {
		systemEvents = systemEvents[:visibleRows]
	}

	// Build headers
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Highlight)
	leftHeader := headerStyle.Render("Events")
	rightHeader := lipgloss.NewStyle().Bold(true).Foreground(colors.Muted).Render("System")

	// Pad headers to column width
	leftHeader = output.PadRight(leftHeader, colWidth)
	rightHeader = output.PadRight(rightHeader, colWidth)

	separator := lipgloss.NewStyle().Foreground(colors.TableBorder).Render("│")

	// Build rows
	lines := make([]string, 0, visibleRows+1)
	lines = append(lines, leftHeader+" "+separator+" "+rightHeader)

	for i := range visibleRows {
		leftCell := ""
		rightCell := ""

		if i < len(userEvents) {
			leftCell = m.renderCompactEvent(userEvents[i], colWidth)
		}
		if i < len(systemEvents) {
			rightCell = m.renderCompactEvent(systemEvents[i], colWidth)
		}

		// Pad cells to column width
		leftCell = output.PadRight(leftCell, colWidth)
		rightCell = output.PadRight(rightCell, colWidth)

		lines = append(lines, leftCell+" "+separator+" "+rightCell)
	}

	return strings.Join(lines, "\n")
}

// renderCompactEvent renders a single event in compact format for dual columns.
func (m Model) renderCompactEvent(e Event, maxWidth int) string {
	// Format: HH:MM:SS device level desc
	timeStr := m.styles.Time.Render(e.Timestamp.Format("15:04:05"))

	// Truncate device name if needed
	device := e.Device
	if len(device) > 10 {
		device = device[:9] + "…"
	}
	deviceStr := m.styles.Device.Render(device)

	// Level indicator (short form)
	levelStyle := m.getTypeStyle(e.Type)
	levelShort := e.Type
	if len(levelShort) > 6 {
		levelShort = levelShort[:6]
	}
	levelStr := levelStyle.Render(levelShort)

	prefix := timeStr + " " + deviceStr + " " + levelStr + " "
	prefixLen := 8 + 1 + min(len(e.Device), 10) + 1 + min(len(e.Type), 6) + 1

	// Description with remaining width
	descWidth := maxWidth - prefixLen
	if descWidth < 5 {
		descWidth = 5
	}
	desc := output.Truncate(e.Description, descWidth)
	descStr := m.styles.Description.Render(desc)

	return prefix + descStr
}
