// Package events provides the event stream view for the TUI.
package events

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/paginator"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/events"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
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
	dualColumnMinWidth = 80 // Lowered to fit in dashboard right column
)

// Column represents which column has focus in dual-column mode.
type Column int

const (
	// ColumnUser is the left column showing user-relevant events.
	ColumnUser Column = iota
	// ColumnSystem is the right column showing system events.
	ColumnSystem
)

// Deps holds the dependencies for the events component.
type Deps struct {
	Ctx         context.Context
	Svc         *shelly.Service
	IOS         *iostreams.IOStreams
	EventStream *automation.EventStream // Shared event stream (WebSocket for Gen2+, polling for Gen1)
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
	mu            sync.RWMutex
	userEvents    []Event // User-relevant events (status changes, errors, etc.)
	systemEvents  []Event // System events (full_status polling noise)
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
	eventStream *automation.EventStream
	state       *sharedState
	scroller    *panel.Scroller
	maxItems    int
	width       int
	height      int
	styles      Styles
	paused      bool
	autoScroll  bool // Auto-scroll to top when new events arrive (newest at top)
	focused     bool // Whether this component has focus

	// Dual column navigation
	focusedColumn Column // Which column has focus (ColumnUser or ColumnSystem)
	userCursor    int    // Cursor position in user events list
	systemCursor  int    // Cursor position in system events list

	// Paginators for dual column view
	userPaginator   paginator.Model
	systemPaginator paginator.Model
	perPage         int // Events per page (calculated from height)

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
			Foreground(theme.Yellow()).
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

	// Create paginators with Dots style (filled/empty dots)
	// Use bright colors for visibility
	userPag := paginator.New()
	userPag.Type = paginator.Dots
	userPag.PerPage = 10 // Will be updated by SetSize
	userPag.ActiveDot = lipgloss.NewStyle().Foreground(theme.Purple()).Bold(true).Render("●")
	userPag.InactiveDot = lipgloss.NewStyle().Foreground(theme.Purple()).Render("○")

	sysPag := paginator.New()
	sysPag.Type = paginator.Dots
	sysPag.PerPage = 10
	sysPag.ActiveDot = lipgloss.NewStyle().Foreground(theme.Red()).Bold(true).Render("●")
	sysPag.InactiveDot = lipgloss.NewStyle().Foreground(theme.Red()).Render("○")

	return Model{
		ctx:             deps.Ctx,
		svc:             deps.Svc,
		ios:             deps.IOS,
		eventStream:     deps.EventStream,
		scroller:        panel.NewScroller(0, 10), // Will be updated by SetSize
		maxItems:        50,                       // Per list - each category capped at 50
		autoScroll:      true,                     // Start with auto-scroll enabled (cursor stays at top/newest)
		focused:         false,
		focusedColumn:   ColumnUser, // Start with user column focused
		userCursor:      0,
		systemCursor:    0,
		perPage:         10, // Will be updated by SetSize
		userPaginator:   userPag,
		systemPaginator: sysPag,
		styles:          DefaultStyles(),
		state: &sharedState{
			userEvents:    []Event{},
			systemEvents:  []Event{},
			subscriptions: make(map[string]context.CancelFunc),
			connStatus:    make(map[string]bool),
		},
	}
}

// SetFocused sets whether the events component has focus.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
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
	debug.TraceLock("events", "Lock", "addEvent")
	m.state.mu.Lock()
	defer func() {
		m.state.mu.Unlock()
		debug.TraceUnlock("events", "Lock", "addEvent")
	}()

	// Add to appropriate list based on event type
	if isUserEvent(e) {
		m.state.userEvents = append([]Event{e}, m.state.userEvents...)
		if len(m.state.userEvents) > m.maxItems {
			m.state.userEvents = m.state.userEvents[:m.maxItems]
		}
	} else {
		m.state.systemEvents = append([]Event{e}, m.state.systemEvents...)
		if len(m.state.systemEvents) > m.maxItems {
			m.state.systemEvents = m.state.systemEvents[:m.maxItems]
		}
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

// handleKeyPress handles keyboard input for scrolling, pagination, and clearing.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "h", "left":
		m.focusedColumn = ColumnUser
		m.autoScroll = false
	case "l", "right":
		m.focusedColumn = ColumnSystem
		m.autoScroll = false
	case "j", "down":
		m = m.moveCursorDown()
	case "k", "up":
		m = m.moveCursorUp()
	case "n", "pgdown", "ctrl+d":
		m = m.pageDown()
	case "N", "pgup", "ctrl+u":
		m = m.pageUp()
	case "g":
		m = m.jumpToStart()
	case "G":
		m = m.jumpToEnd()
	case "c":
		m = m.clearEvents()
	case "p", "space":
		m.paused = !m.paused
	case "f":
		m.filterByDevice = !m.filterByDevice
	}
	return m
}

// getEventCounts returns the number of user and system events.
func (m Model) getEventCounts() (userCount, sysCount int) {
	m.state.mu.RLock()
	userCount = len(m.state.userEvents)
	sysCount = len(m.state.systemEvents)
	m.state.mu.RUnlock()
	return userCount, sysCount
}

// moveCursorDown moves the cursor down in the focused column.
func (m Model) moveCursorDown() Model {
	userCount, sysCount := m.getEventCounts()

	if m.focusedColumn == ColumnUser {
		if m.userCursor < userCount-1 {
			m.userCursor++
		}
		m.userPaginator.Page = m.userCursor / m.perPage
	} else {
		if m.systemCursor < sysCount-1 {
			m.systemCursor++
		}
		m.systemPaginator.Page = m.systemCursor / m.perPage
	}
	m.scroller.CursorDown()
	m.autoScroll = false
	return m
}

// moveCursorUp moves the cursor up in the focused column.
func (m Model) moveCursorUp() Model {
	if m.focusedColumn == ColumnUser {
		if m.userCursor > 0 {
			m.userCursor--
		}
		m.userPaginator.Page = m.userCursor / m.perPage
	} else {
		if m.systemCursor > 0 {
			m.systemCursor--
		}
		m.systemPaginator.Page = m.systemCursor / m.perPage
	}
	m.scroller.CursorUp()
	m.autoScroll = m.userCursor == 0 && m.systemCursor == 0
	return m
}

// pageDown moves to the next page in the focused column.
func (m Model) pageDown() Model {
	if m.focusedColumn == ColumnUser {
		m.userPaginator.NextPage()
		m.userCursor = m.userPaginator.Page * m.perPage
	} else {
		m.systemPaginator.NextPage()
		m.systemCursor = m.systemPaginator.Page * m.perPage
	}
	m.scroller.PageDown()
	m.autoScroll = false
	return m
}

// pageUp moves to the previous page in the focused column.
func (m Model) pageUp() Model {
	if m.focusedColumn == ColumnUser {
		m.userPaginator.PrevPage()
		m.userCursor = m.userPaginator.Page * m.perPage
	} else {
		m.systemPaginator.PrevPage()
		m.systemCursor = m.systemPaginator.Page * m.perPage
	}
	m.scroller.PageUp()
	m.autoScroll = m.userCursor == 0 && m.systemCursor == 0
	return m
}

// jumpToStart moves to the first item in both columns.
func (m Model) jumpToStart() Model {
	m.userPaginator.Page = 0
	m.systemPaginator.Page = 0
	m.userCursor = 0
	m.systemCursor = 0
	m.scroller.CursorToStart()
	m.autoScroll = true
	return m
}

// jumpToEnd moves to the last item in the focused column.
func (m Model) jumpToEnd() Model {
	userCount, sysCount := m.getEventCounts()

	if m.focusedColumn == ColumnUser {
		m.userPaginator.Page = m.userPaginator.TotalPages - 1
		m.userCursor = max(0, userCount-1)
	} else {
		m.systemPaginator.Page = m.systemPaginator.TotalPages - 1
		m.systemCursor = max(0, sysCount-1)
	}
	m.scroller.CursorToEnd()
	m.autoScroll = false
	return m
}

// clearEvents clears all events and resets cursors.
func (m Model) clearEvents() Model {
	m.Clear()
	m.userPaginator.Page = 0
	m.systemPaginator.Page = 0
	m.userCursor = 0
	m.systemCursor = 0
	m.scroller.SetItemCount(0)
	m.autoScroll = true
	return m
}

func (m Model) togglePause() Model {
	m.paused = !m.paused
	return m
}

// eventCount returns the total number of events (thread-safe).
func (m Model) eventCount() int {
	m.state.mu.RLock()
	defer m.state.mu.RUnlock()
	return len(m.state.userEvents) + len(m.state.systemEvents)
}

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height

	// For dual columns: 4 reserved lines (title + column headers + separator + pagination dots)
	// Calculate events per page
	m.perPage = height - 4
	if m.perPage < 1 {
		m.perPage = 1
	}

	// Update paginators
	m.userPaginator.PerPage = m.perPage
	m.systemPaginator.PerPage = m.perPage

	// Reserve 1 row for header (single column mode)
	visibleRows := height - 1
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	return m
}

// View renders the events (content only - wrapper handles title/footer).
func (m Model) View() string {
	debug.TraceLock("events", "RLock", "View")
	m.state.mu.RLock()
	userEvents := make([]Event, len(m.state.userEvents))
	copy(userEvents, m.state.userEvents)
	systemEvents := make([]Event, len(m.state.systemEvents))
	copy(systemEvents, m.state.systemEvents)
	m.state.mu.RUnlock()
	debug.TraceUnlock("events", "RLock", "View")

	if len(userEvents) == 0 && len(systemEvents) == 0 {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(m.styles.Footer.GetForeground()).
			Render("Waiting for events...")
	}

	// Use dual column layout when width permits
	if m.width >= dualColumnMinWidth {
		return m.renderDualColumnsDirect(userEvents, systemEvents)
	}

	// Single column layout - merge events for display
	eventList := mergeEventsByTime(userEvents, systemEvents)
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

// ConnectionCounts returns the number of connected and total devices via WebSocket.
func (m Model) ConnectionCounts() (connected, total int) {
	debug.TraceLock("events", "RLock", "ConnectionCounts")
	m.state.mu.RLock()
	defer func() {
		m.state.mu.RUnlock()
		debug.TraceUnlock("events", "RLock", "ConnectionCounts")
	}()
	total = len(m.state.connStatus)
	for _, isConnected := range m.state.connStatus {
		if isConnected {
			connected++
		}
	}
	return connected, total
}

// Clear clears all events.
func (m Model) Clear() {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()
	m.state.userEvents = []Event{}
	m.state.systemEvents = []Event{}
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

	// Get events safely with read locking
	m.state.mu.RLock()
	eventList := mergeEventsByTime(m.state.userEvents, m.state.systemEvents)
	m.state.mu.RUnlock()

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
	m.state.mu.RLock()
	eventList := mergeEventsByTime(m.state.userEvents, m.state.systemEvents)
	m.state.mu.RUnlock()

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
	return "n/N:page g/G:first/last p:pause c:clear"
}

// StatusBadge returns status indicators for the wrapper to display.
// Returns event count and any active flags (paused, filtered).
func (m Model) StatusBadge() string {
	m.state.mu.RLock()
	eventCount := len(m.state.userEvents) + len(m.state.systemEvents)
	m.state.mu.RUnlock()

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

// mergeEventsByTime merges user and system events sorted by timestamp (newest first).
func mergeEventsByTime(user, system []Event) []Event {
	result := make([]Event, 0, len(user)+len(system))
	i, j := 0, 0

	for i < len(user) || j < len(system) {
		if i >= len(user) {
			result = append(result, system[j:]...)
			break
		}
		if j >= len(system) {
			result = append(result, user[i:]...)
			break
		}
		// Both slices are sorted newest-first, so compare timestamps
		if user[i].Timestamp.After(system[j].Timestamp) {
			result = append(result, user[i])
			i++
		} else {
			result = append(result, system[j])
			j++
		}
	}
	return result
}

// columnLayout holds width calculations for event columns.
type columnLayout struct {
	colWidth  int
	sysColW   int
	timeW     int
	levelW    int
	userDevW  int
	userCompW int
	userDescW int
	sysDevW   int
	sysDescW  int
}

// computeColumnLayout calculates column widths with dynamic sizing.
// System events can shrink below 50% to give user events more room.
// Component column collapses when empty and expands as needed.
func computeColumnLayout(totalWidth int, userEvents, systemEvents []Event) columnLayout {
	const timeW = 8
	const levelW = 12
	const minDeviceW, minDescW = 10, 8
	const headerCompW = 4 // "COMP"
	const spacesUser, spacesSys = 4, 3
	const separatorW = 3 // " │ "

	// Calculate available width for both columns
	availableWidth := totalWidth - separatorW
	maxSysColWidth := availableWidth / 2 // System column caps at 50%
	minSysColWidth := 30                 // Minimum system column width

	// Calculate actual system content width needed
	sysContentWidth := calculateSystemContentWidth(systemEvents, timeW, levelW, spacesSys)

	// System column: use content width, capped at 50%, with minimum
	sysColWidth := max(minSysColWidth, min(sysContentWidth, maxSysColWidth))

	// User column gets the rest
	userColWidth := max(30, availableWidth-sysColWidth)

	// Calculate component column width dynamically
	// Collapse to header width when no events have components
	maxCompLen := 0
	for _, e := range userEvents {
		compLen := len(e.Component)
		if e.ComponentID > 0 {
			compLen += 2 // ":N" suffix
		}
		if compLen > maxCompLen {
			maxCompLen = compLen
		}
	}
	// Use header width as minimum, cap at reasonable max
	userCompW := max(headerCompW, min(maxCompLen, 12))

	// User column layout
	remainingUser := userColWidth - timeW - levelW - spacesUser - userCompW
	userDeviceW := max(minDeviceW, remainingUser*30/100)
	userDescW := max(minDescW, remainingUser-userDeviceW)

	// System column layout (no COMP)
	remainingSys := sysColWidth - timeW - levelW - spacesSys
	sysDeviceW := max(minDeviceW, remainingSys*25/100)
	sysDescW := max(minDescW, remainingSys-sysDeviceW)

	return columnLayout{
		colWidth:  userColWidth,
		timeW:     timeW,
		levelW:    levelW,
		userDevW:  userDeviceW,
		userCompW: userCompW,
		userDescW: userDescW,
		sysDevW:   sysDeviceW,
		sysDescW:  sysDescW,
		sysColW:   sysColWidth,
	}
}

// calculateSystemContentWidth determines the width needed for system events.
func calculateSystemContentWidth(sysEvents []Event, timeW, levelW, spaces int) int {
	if len(sysEvents) == 0 {
		return 30 // Minimum default
	}

	maxDeviceLen := 0
	maxDescLen := 0

	for _, e := range sysEvents {
		deviceLen := len(e.Device)
		if deviceLen > maxDeviceLen {
			maxDeviceLen = deviceLen
		}
		descLen := len(e.Description)
		if descLen > maxDescLen {
			maxDescLen = descLen
		}
	}

	// Total: TIME + DEVICE + LEVEL + DESC + spaces
	// Cap device and desc at reasonable maximums
	maxDeviceLen = min(maxDeviceLen, 16)
	maxDescLen = min(maxDescLen, 40)

	return timeW + maxDeviceLen + levelW + maxDescLen + spaces
}

// dualColumnRenderState holds state for rendering dual columns.
type dualColumnRenderState struct {
	layout           columnLayout
	separator        string
	cursorStyle      lipgloss.Style
	emptyLeft        string
	emptyRight       string
	userCursorInPage int
	sysCursorInPage  int
	maxLinesPerEvent int
}

// renderDualColumnsDirect renders events in two columns: User (left) and System (right).
// Takes pre-split user and system event lists.
func (m Model) renderDualColumnsDirect(userEvents, systemEvents []Event) string {
	colors := theme.GetSemanticColors()
	layout := computeColumnLayout(m.width, userEvents, systemEvents)

	m.userPaginator.SetTotalPages(len(userEvents))
	m.systemPaginator.SetTotalPages(len(systemEvents))

	userStart, userEnd := m.userPaginator.GetSliceBounds(len(userEvents))
	sysStart, sysEnd := m.systemPaginator.GetSliceBounds(len(systemEvents))

	userPage := userEvents[userStart:userEnd]
	sysPage := systemEvents[sysStart:sysEnd]

	separator := lipgloss.NewStyle().Foreground(colors.TableBorder).Render("│")

	// Build header lines
	lines := m.buildDualColumnHeader(layout, separator)

	// Create render state
	rs := dualColumnRenderState{
		layout:           layout,
		separator:        separator,
		cursorStyle:      lipgloss.NewStyle().Background(colors.AltBackground).Bold(true),
		emptyLeft:        output.PadRight("", layout.colWidth),
		emptyRight:       output.PadRight("", layout.sysColW),
		userCursorInPage: m.userCursor - userStart,
		sysCursorInPage:  m.systemCursor - sysStart,
		maxLinesPerEvent: m.calcMaxLinesPerUserEvent(len(userPage)),
	}

	// Render event rows
	lines = m.renderEventRows(lines, userPage, sysPage, rs)

	// Pagination dots
	leftDots := lipgloss.NewStyle().Width(layout.colWidth).Align(lipgloss.Center).Render(m.userPaginator.View())
	rightDots := lipgloss.NewStyle().Width(layout.sysColW).Align(lipgloss.Center).Render(m.systemPaginator.View())
	lines = append(lines, leftDots+" "+separator+" "+rightDots)

	return strings.Join(lines, "\n")
}

// buildDualColumnHeader builds the header section for dual column view.
func (m Model) buildDualColumnHeader(layout columnLayout, separator string) []string {
	colors := theme.GetSemanticColors()

	userTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Purple())
	sysTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Red())
	if m.focused && m.focusedColumn == ColumnUser {
		userTitleStyle = userTitleStyle.Underline(true)
	} else if m.focused {
		sysTitleStyle = sysTitleStyle.Underline(true)
	}

	leftTitle := output.PadRight(userTitleStyle.Render("User Events"), layout.colWidth)
	rightTitle := output.PadRight(sysTitleStyle.Render("System Events"), layout.sysColW)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Purple())
	leftColHeader := headerStyle.Render(output.PadRight("TIME", layout.timeW)) + " " +
		headerStyle.Render(output.PadRight("DEVICE", layout.userDevW)) + " " +
		headerStyle.Render(output.PadRight("COMP", layout.userCompW)) + " " +
		headerStyle.Render(output.PadRight("LEVEL", layout.levelW)) + " " +
		headerStyle.Render("DESC")

	sysHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Red())
	rightColHeader := sysHeaderStyle.Render(output.PadRight("TIME", layout.timeW)) + " " +
		sysHeaderStyle.Render(output.PadRight("DEVICE", layout.sysDevW)) + " " +
		sysHeaderStyle.Render(output.PadRight("LEVEL", layout.levelW)) + " " +
		sysHeaderStyle.Render("DESC")

	leftColHeader = output.PadRight(leftColHeader, layout.colWidth)
	rightColHeader = output.PadRight(rightColHeader, layout.sysColW)

	leftSepLine := lipgloss.NewStyle().Foreground(colors.TableBorder).Render(strings.Repeat("-", layout.colWidth))
	rightSepLine := lipgloss.NewStyle().Foreground(colors.TableBorder).Render(strings.Repeat("-", layout.sysColW))

	return []string{
		leftTitle + " " + separator + " " + rightTitle,
		leftColHeader + " " + separator + " " + rightColHeader,
		leftSepLine + " " + separator + " " + rightSepLine,
	}
}

// renderEventRows renders the event data rows for dual column view.
func (m Model) renderEventRows(lines []string, userPage, sysPage []Event, rs dualColumnRenderState) []string {
	linesUsed, userIdx, sysIdx := 0, 0, 0

	for linesUsed < m.perPage {
		switch {
		case userIdx < len(userPage):
			newLines, newLinesUsed, newSysIdx := m.renderUserEventRow(
				userPage[userIdx], sysPage, userIdx, sysIdx, linesUsed, rs)
			lines = append(lines, newLines...)
			linesUsed += newLinesUsed
			sysIdx = newSysIdx
			userIdx++
		case sysIdx < len(sysPage):
			line := m.renderSystemOnlyRow(sysPage[sysIdx], sysIdx, rs)
			lines = append(lines, line)
			sysIdx++
			linesUsed++
		default:
			lines = append(lines, rs.emptyLeft+" "+rs.separator+" "+rs.emptyRight)
			linesUsed++
		}
	}
	return lines
}

// renderUserEventRow renders a user event row with optional corresponding system event.
func (m Model) renderUserEventRow(userEvt Event, sysPage []Event, userIdx, sysIdx, linesUsed int, rs dualColumnRenderState) (resultLines []string, linesAdded, newSysIdx int) {
	isUserSelected := m.focused && m.focusedColumn == ColumnUser && userIdx == rs.userCursorInPage

	userLines := m.renderUserEventWrapped(userEvt, rs.layout.timeW, rs.layout.userDevW,
		rs.layout.userCompW, rs.layout.levelW, rs.layout.userDescW, rs.maxLinesPerEvent)

	sysLine := ""
	isSysSelected := false
	newSysIdx = sysIdx
	if newSysIdx < len(sysPage) {
		isSysSelected = m.focused && m.focusedColumn == ColumnSystem && newSysIdx == rs.sysCursorInPage
		sysLine = m.renderSystemEvent(sysPage[newSysIdx], rs.layout.timeW, rs.layout.sysDevW,
			rs.layout.levelW, rs.layout.sysDescW)
		newSysIdx++
	}

	resultLines = make([]string, 0, len(userLines))
	for i, uline := range userLines {
		if linesUsed+linesAdded >= m.perPage {
			break
		}
		leftCell := output.PadRight(uline, rs.layout.colWidth)
		rightCell := rs.emptyRight
		if i == 0 {
			rightCell = output.PadRight(sysLine, rs.layout.sysColW)
			if isUserSelected {
				leftCell = rs.cursorStyle.Render(leftCell)
			}
			if isSysSelected {
				rightCell = rs.cursorStyle.Render(rightCell)
			}
		}
		resultLines = append(resultLines, leftCell+" "+rs.separator+" "+rightCell)
		linesAdded++
	}
	return resultLines, linesAdded, newSysIdx
}

// renderSystemOnlyRow renders a row with only a system event (no user event).
func (m Model) renderSystemOnlyRow(sysEvt Event, sysIdx int, rs dualColumnRenderState) string {
	isSysSelected := m.focused && m.focusedColumn == ColumnSystem && sysIdx == rs.sysCursorInPage
	rightCell := output.PadRight(m.renderSystemEvent(sysEvt, rs.layout.timeW, rs.layout.sysDevW,
		rs.layout.levelW, rs.layout.sysDescW), rs.layout.sysColW)
	if isSysSelected {
		rightCell = rs.cursorStyle.Render(rightCell)
	}
	return rs.emptyLeft + " " + rs.separator + " " + rightCell
}

// calcMaxLinesPerUserEvent calculates how many lines each user event can use.
func (m Model) calcMaxLinesPerUserEvent(userEventCount int) int {
	if userEventCount > 0 && m.perPage > userEventCount {
		maxLines := m.perPage / userEventCount
		return min(3, maxLines) // Cap at 3 lines per event
	}
	return 1
}

// renderUserEventWrapped renders a user event with word-wrapped description.
// Returns multiple lines if the description needs wrapping, up to maxLines.
func (m Model) renderUserEventWrapped(e Event, timeW, deviceW, compW, levelW, descW, maxLines int) []string {
	// Time column
	timeStr := m.styles.Time.Width(timeW).Render(e.Timestamp.Format("15:04:05"))

	// Device column - truncate if needed
	device := e.Device
	if len(device) > deviceW {
		device = device[:deviceW-1] + "…"
	}
	deviceStr := m.styles.Device.Width(deviceW).Render(device)

	// Component column
	comp := m.formatComponent(e)
	if len(comp) > compW {
		comp = comp[:compW-1] + "…"
	}
	compStr := m.styles.Component.Width(compW).Render(comp)

	// Level column
	levelStyle := m.getTypeStyle(e.Type)
	levelStr := levelStyle.Width(levelW).Render(e.Type)

	// Build prefix for first line and continuation indent
	prefix := timeStr + " " + deviceStr + " " + compStr + " " + levelStr + " "
	prefixWidth := timeW + 1 + deviceW + 1 + compW + 1 + levelW + 1
	contIndent := strings.Repeat(" ", prefixWidth)

	// Wrap description using lipgloss word-wrap
	wrappedDesc := lipgloss.NewStyle().Width(descW).Render(e.Description)
	descLines := strings.Split(wrappedDesc, "\n")

	// Cap at maxLines and add truncation indicator if needed
	truncated := false
	if len(descLines) > maxLines {
		descLines = descLines[:maxLines]
		truncated = true
	}

	// Build result lines
	result := make([]string, len(descLines))
	for i, descLine := range descLines {
		// Add truncation indicator on last line if we truncated
		if truncated && i == len(descLines)-1 {
			if len(descLine) > descW-1 {
				descLine = descLine[:descW-2] + "…"
			} else {
				descLine += "…"
			}
		}
		descStr := m.styles.Description.Render(descLine)

		if i == 0 {
			result[i] = prefix + descStr
		} else {
			result[i] = contIndent + descStr
		}
	}

	return result
}

// renderSystemEvent renders a system event row (no COMP column).
func (m Model) renderSystemEvent(e Event, timeW, deviceW, levelW, descW int) string {
	// Time column
	timeStr := m.styles.Time.Width(timeW).Render(e.Timestamp.Format("15:04:05"))

	// Device column - truncate if needed
	device := e.Device
	if len(device) > deviceW {
		device = device[:deviceW-1] + "…"
	}
	deviceStr := m.styles.Device.Width(deviceW).Render(device)

	// Level column - no truncation needed now that levelW is 12
	levelStyle := m.getTypeStyle(e.Type)
	levelStr := levelStyle.Width(levelW).Render(e.Type)

	// Description column
	desc := e.Description
	if len(desc) > descW {
		desc = desc[:descW-1] + "…"
	}
	descStr := m.styles.Description.Render(desc)

	return timeStr + " " + deviceStr + " " + levelStr + " " + descStr
}
