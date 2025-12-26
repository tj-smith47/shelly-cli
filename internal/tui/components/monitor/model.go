// Package monitor provides real-time device monitoring for the TUI.
package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/events"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Deps holds the dependencies for the monitor component.
type Deps struct {
	Ctx             context.Context
	Svc             *shelly.Service
	IOS             *iostreams.IOStreams
	RefreshInterval time.Duration
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

// DeviceStatus represents the real-time status of a device.
type DeviceStatus struct {
	Name        string
	Address     string
	Type        string
	Online      bool
	Power       float64
	Voltage     float64
	Current     float64
	Frequency   float64
	TotalEnergy float64 // Total energy consumption in Wh
	UpdatedAt   time.Time
	Error       error
}

// StatusUpdateMsg is sent when device status is updated.
type StatusUpdateMsg struct {
	Statuses []DeviceStatus
	Err      error
}

// RefreshTickMsg triggers periodic refresh (fallback for Gen1 devices).
type RefreshTickMsg struct{}

// DeviceEventMsg wraps WebSocket events from devices.
type DeviceEventMsg struct {
	Event events.Event
}

// Model holds the monitor state.
type Model struct {
	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	statuses        []DeviceStatus
	statusMap       map[string]*DeviceStatus // For O(1) updates by device name
	initialLoad     bool                     // True only on first load (shows loading screen)
	refreshing      bool                     // True during background refresh (shows indicator, keeps data)
	useWebSocket    bool                     // True if using WebSocket for updates
	eventStream     *shelly.EventStream      // WebSocket event stream
	eventChan       chan events.Event        // Channel for WebSocket events
	err             error
	width           int
	height          int
	styles          Styles
	refreshInterval time.Duration
	scrollOffset    int
	cursor          int // Currently selected row
}

// Styles for the monitor component.
type Styles struct {
	Container    lipgloss.Style
	Header       lipgloss.Style
	Row          lipgloss.Style
	SelectedRow  lipgloss.Style
	OnlineIcon   lipgloss.Style
	OfflineIcon  lipgloss.Style
	UpdatingIcon lipgloss.Style
	DeviceName   lipgloss.Style
	Address      lipgloss.Style
	Power        lipgloss.Style
	Voltage      lipgloss.Style
	Current      lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	LastUpdated  lipgloss.Style
	Separator    lipgloss.Style
	SummaryCard  lipgloss.Style
	SummaryValue lipgloss.Style
	Energy       lipgloss.Style
}

// DefaultStyles returns default styles for the monitor.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			Padding(0, 1), // No border - wrapper adds border with title
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight).
			MarginBottom(1),
		Row: lipgloss.NewStyle().
			MarginBottom(1),
		SelectedRow: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		OnlineIcon: lipgloss.NewStyle().
			Foreground(colors.Online),
		OfflineIcon: lipgloss.NewStyle().
			Foreground(colors.Offline),
		UpdatingIcon: lipgloss.NewStyle().
			Foreground(colors.Updating),
		DeviceName: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text),
		Address: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Power: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Voltage: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Current: lipgloss.NewStyle().
			Foreground(colors.Primary),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(8),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		LastUpdated: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Separator: lipgloss.NewStyle().
			Foreground(colors.Muted),
		SummaryCard: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(0, 1).
			MarginRight(1),
		SummaryValue: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight),
		Energy: lipgloss.NewStyle().
			Foreground(colors.Success),
	}
}

// New creates a new monitor model.
func New(deps Deps) Model {
	if err := deps.validate(); err != nil {
		panic(fmt.Sprintf("monitor: invalid deps: %v", err))
	}

	refreshInterval := deps.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 10 * time.Second // Fallback polling interval for Gen1
	}

	// Create event stream for WebSocket connections
	eventStream := shelly.NewEventStream(deps.Svc)
	eventChan := make(chan events.Event, 100)

	// Subscribe to all events and forward to channel
	eventStream.Subscribe(func(evt events.Event) {
		select {
		case eventChan <- evt:
		default:
			// Channel full, drop event (shouldn't happen with buffer)
			deps.IOS.DebugErr("monitor event channel full", nil)
		}
	})

	return Model{
		ctx:             deps.Ctx,
		svc:             deps.Svc,
		ios:             deps.IOS,
		statuses:        []DeviceStatus{},
		statusMap:       make(map[string]*DeviceStatus),
		initialLoad:     true,
		refreshing:      false,
		useWebSocket:    true,
		eventStream:     eventStream,
		eventChan:       eventChan,
		styles:          DefaultStyles(),
		refreshInterval: refreshInterval,
	}
}

// Init returns the initial command for the monitor.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.fetchStatuses(), // Initial HTTP fetch for immediate data
	}

	if m.useWebSocket {
		// Start WebSocket connections and event listener
		cmds = append(cmds, m.startEventStream(), m.listenForEvents())
	} else {
		// Fallback to polling
		cmds = append(cmds, m.scheduleRefresh())
	}

	return tea.Batch(cmds...)
}

// startEventStream starts WebSocket connections to all devices.
func (m Model) startEventStream() tea.Cmd {
	return func() tea.Msg {
		if err := m.eventStream.Start(); err != nil {
			m.ios.DebugErr("start event stream", err)
		}
		return nil
	}
}

// listenForEvents returns a command that listens for events from the channel.
func (m Model) listenForEvents() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-m.ctx.Done():
			return nil
		case evt := <-m.eventChan:
			return DeviceEventMsg{Event: evt}
		}
	}
}

// scheduleRefresh schedules the next refresh tick.
func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// fetchStatuses returns a command that fetches device statuses.
func (m Model) fetchStatuses() tea.Cmd {
	return func() tea.Msg {
		deviceMap := config.ListDevices()
		if len(deviceMap) == 0 {
			return StatusUpdateMsg{Statuses: nil}
		}

		// Add timeout to prevent slow/offline devices from blocking UI
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		// Use errgroup for concurrent status fetching
		// Rate limiting is handled at the service layer
		g, gctx := errgroup.WithContext(ctx)

		results := make(chan DeviceStatus, len(deviceMap))

		for _, d := range deviceMap {
			device := d
			g.Go(func() error {
				status := m.checkDeviceStatus(gctx, device)
				results <- status
				return nil
			})
		}

		go func() {
			if err := g.Wait(); err != nil {
				m.ios.DebugErr("monitor errgroup wait", err)
			}
			close(results)
		}()

		statuses := make([]DeviceStatus, 0, len(deviceMap))
		for status := range results {
			statuses = append(statuses, status)
		}

		// Sort by name for consistent display
		sort.Slice(statuses, func(i, j int) bool {
			return statuses[i].Name < statuses[j].Name
		})

		return StatusUpdateMsg{Statuses: statuses}
	}
}

// checkDeviceStatus fetches the real-time status of a single device.
func (m Model) checkDeviceStatus(ctx context.Context, device model.Device) DeviceStatus {
	status := DeviceStatus{
		Name:      device.Name,
		Address:   device.Address,
		Type:      device.Type,
		Online:    false,
		UpdatedAt: time.Now(),
	}

	// Per-device timeout to prevent single slow device from blocking others
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	snapshot, err := m.svc.GetMonitoringSnapshotAuto(ctx, device.Address)
	if err != nil {
		status.Error = err
		return status
	}

	status.Online = true
	aggregateStatusFromPM(&status, snapshot.PM)
	aggregateStatusFromEM(&status, snapshot.EM)
	aggregateStatusFromEM1(&status, snapshot.EM1)

	return status
}

// aggregateStatusFromPM aggregates status data from PM components.
func aggregateStatusFromPM(status *DeviceStatus, pms []shelly.PMStatus) {
	for _, pm := range pms {
		status.Power += pm.APower
		if status.Voltage == 0 && pm.Voltage > 0 {
			status.Voltage = pm.Voltage
		}
		if status.Current == 0 && pm.Current > 0 {
			status.Current = pm.Current
		}
		if pm.Freq != nil && status.Frequency == 0 {
			status.Frequency = *pm.Freq
		}
		if pm.AEnergy != nil {
			status.TotalEnergy += pm.AEnergy.Total
		}
	}
}

// aggregateStatusFromEM aggregates status data from EM components (3-phase meters).
func aggregateStatusFromEM(status *DeviceStatus, ems []shelly.EMStatus) {
	for _, em := range ems {
		status.Power += em.TotalActivePower
		status.Current += em.TotalCurrent
		if status.Voltage == 0 && em.AVoltage > 0 {
			status.Voltage = em.AVoltage
		}
		if em.AFreq != nil && status.Frequency == 0 {
			status.Frequency = *em.AFreq
		}
	}
}

// aggregateStatusFromEM1 aggregates status data from EM1 components (single-phase meters).
func aggregateStatusFromEM1(status *DeviceStatus, em1s []shelly.EM1Status) {
	for _, em1 := range em1s {
		status.Power += em1.ActPower
		if status.Voltage == 0 && em1.Voltage > 0 {
			status.Voltage = em1.Voltage
		}
		if status.Current == 0 && em1.Current > 0 {
			status.Current = em1.Current
		}
		if em1.Freq != nil && status.Frequency == 0 {
			status.Frequency = *em1.Freq
		}
	}
}

// Refresh returns a command to refresh the monitor.
func (m Model) Refresh() tea.Cmd {
	return m.fetchStatuses()
}

// Update handles messages for the monitor.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusUpdateMsg:
		m.initialLoad = false
		m.refreshing = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.statuses = msg.Statuses
		// Build status map for O(1) updates
		m.statusMap = make(map[string]*DeviceStatus, len(m.statuses))
		for i := range m.statuses {
			m.statusMap[m.statuses[i].Name] = &m.statuses[i]
		}
		return m, nil

	case DeviceEventMsg:
		// Handle WebSocket event - update status in place
		m.handleDeviceEvent(msg.Event)
		// Continue listening for more events
		return m, m.listenForEvents()

	case RefreshTickMsg:
		// Skip refresh if using WebSocket (except for initial load)
		if m.useWebSocket && !m.initialLoad {
			return m, m.scheduleRefresh()
		}
		// Skip refresh if already refreshing to prevent overlap
		if m.refreshing {
			return m, m.scheduleRefresh()
		}
		m.refreshing = true
		return m, tea.Batch(
			m.fetchStatuses(),
			m.scheduleRefresh(),
		)

	case tea.KeyPressMsg:
		m = m.handleKeyPress(msg)
	}

	return m, nil
}

// handleDeviceEvent processes a WebSocket event and updates device status.
func (m Model) handleDeviceEvent(evt events.Event) {
	deviceID := evt.DeviceID()

	switch e := evt.(type) {
	case *events.StatusChangeEvent:
		// Update power/energy data from status change
		if status, ok := m.statusMap[deviceID]; ok {
			status.UpdatedAt = e.Timestamp()
			m.parseStatusChange(status, e.Component, e.Status)
		}

	case *events.FullStatusEvent:
		// Full status update - parse all data
		if status, ok := m.statusMap[deviceID]; ok {
			status.UpdatedAt = e.Timestamp()
			m.parseFullStatus(status, e.Status)
		}

	case *events.DeviceOnlineEvent:
		if status, ok := m.statusMap[deviceID]; ok {
			status.Online = true
			status.Error = nil
			status.UpdatedAt = e.Timestamp()
		}

	case *events.DeviceOfflineEvent:
		if status, ok := m.statusMap[deviceID]; ok {
			status.Online = false
			if e.Reason != "" {
				status.Error = fmt.Errorf("%s", e.Reason)
			}
			status.UpdatedAt = e.Timestamp()
		}
	}
}

// parseStatusChange parses a status change event for a component.
func (m Model) parseStatusChange(status *DeviceStatus, component string, data json.RawMessage) {
	// Parse component status (switch, pm, em, etc.)
	var statusData map[string]any
	if err := json.Unmarshal(data, &statusData); err != nil {
		m.ios.DebugErr(fmt.Sprintf("parse status change for %s", component), err)
		return
	}

	// Extract power data from switch/pm components
	if power, ok := statusData["apower"].(float64); ok {
		status.Power = power
	}
	if voltage, ok := statusData["voltage"].(float64); ok {
		status.Voltage = voltage
	}
	if current, ok := statusData["current"].(float64); ok {
		status.Current = current
	}
}

// parseFullStatus parses a full device status event.
func (m Model) parseFullStatus(status *DeviceStatus, data json.RawMessage) {
	var fullStatus shelly.MonitoringSnapshot
	if err := json.Unmarshal(data, &fullStatus); err != nil {
		return
	}

	status.Online = fullStatus.Online
	status.Power = 0
	status.Voltage = 0
	status.Current = 0
	status.TotalEnergy = 0

	aggregateStatusFromPM(status, fullStatus.PM)
	aggregateStatusFromEM(status, fullStatus.EM)
	aggregateStatusFromEM1(status, fullStatus.EM1)
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "g":
		m.cursor = 0
		m.scrollOffset = 0
	case "G":
		m = m.cursorToEnd()
	case "pgdown", "ctrl+d":
		m = m.pageDown()
	case "pgup", "ctrl+u":
		m = m.pageUp()
	}
	return m
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.statuses)-1 {
		m.cursor++
		visibleRows := m.visibleRows()
		if m.cursor >= m.scrollOffset+visibleRows {
			m.scrollOffset = m.cursor - visibleRows + 1
		}
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.scrollOffset {
			m.scrollOffset = m.cursor
		}
	}
	return m
}

func (m Model) cursorToEnd() Model {
	if len(m.statuses) > 0 {
		m.cursor = len(m.statuses) - 1
		maxOffset := len(m.statuses) - m.visibleRows()
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageDown() Model {
	if len(m.statuses) == 0 {
		return m
	}
	visibleRows := m.visibleRows()
	m.cursor += visibleRows
	if m.cursor >= len(m.statuses) {
		m.cursor = len(m.statuses) - 1
	}
	if m.cursor >= m.scrollOffset+visibleRows {
		m.scrollOffset = m.cursor - visibleRows + 1
	}
	maxOffset := len(m.statuses) - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageUp() Model {
	if len(m.statuses) == 0 {
		return m
	}
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

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// visibleRows calculates how many device rows can be displayed.
// Each device card takes 3 lines (2 content + margin) + 1 separator = 4 lines total.
func (m Model) visibleRows() int {
	// Account for: header (2), summary (4), empty line (1), footer (2), container padding (2) = 11 lines overhead
	availableHeight := m.height - 11
	// Each card: 2 lines content + 1 margin + 1 separator = 4 lines
	// Last card has no separator, so for N cards: 4N - 1 lines
	// Solving for N: N = (availableHeight + 1) / 4
	rowHeight := 4
	visibleRows := (availableHeight + 1) / rowHeight
	if visibleRows < 1 {
		return 1
	}
	return visibleRows
}

// View renders the monitor.
func (m Model) View() string {
	if m.initialLoad {
		loadingText := m.styles.UpdatingIcon.Render("◐ ") + "Fetching device statuses..."
		return m.styles.Container.
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(loadingText)
	}

	if m.err != nil {
		return m.styles.Container.
			Width(m.width).
			Render(theme.StatusError().Render("Error: " + m.err.Error()))
	}

	if len(m.statuses) == 0 {
		return m.styles.Container.
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No devices to monitor.\nUse 'shelly device add' to add devices.")
	}

	// Calculate totals for summary
	var totalPower, totalEnergy float64
	var onlineCount, offlineCount int
	for _, s := range m.statuses {
		if s.Online {
			onlineCount++
			totalPower += s.Power
			totalEnergy += s.TotalEnergy
		} else {
			offlineCount++
		}
	}

	// Summary cards row
	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderSummaryCard("Power", m.formatPower(totalPower)),
		m.renderSummaryCard("Energy", m.formatEnergy(totalEnergy)),
		m.renderSummaryCard("Devices", fmt.Sprintf("%d/%d online", onlineCount, len(m.statuses))),
	)

	// Calculate visible rows
	visibleRows := m.visibleRows()

	// Apply scroll offset
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleRows
	if endIdx > len(m.statuses) {
		endIdx = len(m.statuses)
	}

	// Calculate content width (container width minus padding)
	contentWidth := m.width - 6 // 2 padding + 2 border on each side

	var rows string
	for i := startIdx; i < endIdx; i++ {
		s := m.statuses[i]
		isSelected := i == m.cursor
		rows += m.renderDeviceCard(s, isSelected)
		if i < endIdx-1 {
			separator := m.styles.Separator.Render(strings.Repeat("─", contentWidth))
			rows += separator + "\n"
		}
	}

	// Scroll indicator
	scrollInfo := ""
	if len(m.statuses) > visibleRows {
		scrollInfo = m.styles.LastUpdated.Render(
			fmt.Sprintf(" [%d-%d of %d] ", startIdx+1, endIdx, len(m.statuses)),
		)
	}

	footer := m.styles.LastUpdated.Render(
		fmt.Sprintf("Last updated: %s", time.Now().Format("15:04:05")),
	) + scrollInfo

	content := lipgloss.JoinVertical(lipgloss.Left, summary, "", rows, footer)
	return m.styles.Container.Width(m.width).Height(m.height).Render(content)
}

// renderSummaryCard renders a summary card with label and value.
func (m Model) renderSummaryCard(label, value string) string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		m.styles.Label.Render(label),
		m.styles.SummaryValue.Render(value),
	)
	return m.styles.SummaryCard.Width(20).Render(content)
}

// formatPower formats watts for display.
func (m Model) formatPower(watts float64) string {
	if watts >= 1000 || watts <= -1000 {
		return fmt.Sprintf("%.2f kW", watts/1000)
	}
	return fmt.Sprintf("%.1f W", watts)
}

// formatEnergy formats watt-hours for display.
func (m Model) formatEnergy(wh float64) string {
	if wh >= 1000000 {
		return fmt.Sprintf("%.2f MWh", wh/1000000)
	}
	if wh >= 1000 {
		return fmt.Sprintf("%.2f kWh", wh/1000)
	}
	return fmt.Sprintf("%.1f Wh", wh)
}

// renderDeviceCard renders a single device status card.
func (m Model) renderDeviceCard(s DeviceStatus, isSelected bool) string {
	// Choose row style based on selection
	rowStyle := m.styles.Row
	if isSelected {
		rowStyle = m.styles.SelectedRow
	}

	// Selection indicator
	selIndicator := "  "
	if isSelected {
		selIndicator = "▶ "
	}

	statusIcon := m.styles.OfflineIcon.Render("○")
	if s.Online {
		statusIcon = m.styles.OnlineIcon.Render("●")
	}

	// First line: selection indicator, icon, name, address, type
	line1 := fmt.Sprintf("%s%s %s  %s  %s",
		selIndicator,
		statusIcon,
		m.styles.DeviceName.Render(s.Name),
		m.styles.Address.Render(s.Address),
		m.styles.Address.Render(s.Type),
	)

	if !s.Online {
		errMsg := "unreachable"
		if s.Error != nil {
			errMsg = s.Error.Error()
			if len(errMsg) > 40 {
				errMsg = errMsg[:40] + "..."
			}
		}
		line2 := "    " + theme.StatusError().Render(errMsg)
		return rowStyle.Render(line1+"\n"+line2) + "\n"
	}

	// Second line: metrics
	line2 := "    " + m.buildMetricsLine(s)

	return rowStyle.Render(line1+"\n"+line2) + "\n"
}

// buildMetricsLine builds the metrics display line for a device.
func (m Model) buildMetricsLine(s DeviceStatus) string {
	metrics := m.collectMetrics(s)
	if len(metrics) == 0 {
		return m.styles.LastUpdated.Render("no power data")
	}

	result := ""
	for i, metric := range metrics {
		if i > 0 {
			result += m.styles.Separator.Render(" │ ")
		}
		result += metric
	}
	return result
}

// collectMetrics collects all available metrics for a device.
func (m Model) collectMetrics(s DeviceStatus) []string {
	metrics := []string{}

	if s.Power > 0 || s.Power < 0 {
		metrics = append(metrics, m.styles.Power.Render(fmt.Sprintf("%.1fW", s.Power)))
	}
	if s.Voltage > 0 {
		metrics = append(metrics, m.styles.Voltage.Render(fmt.Sprintf("%.1fV", s.Voltage)))
	}
	if s.Current > 0 {
		metrics = append(metrics, m.styles.Current.Render(fmt.Sprintf("%.2fA", s.Current)))
	}
	if s.Frequency > 0 {
		metrics = append(metrics, m.styles.Value.Render(fmt.Sprintf("%.1fHz", s.Frequency)))
	}
	if s.TotalEnergy > 0 {
		metrics = append(metrics, m.styles.Energy.Render(m.formatEnergy(s.TotalEnergy)))
	}

	return metrics
}

// StatusCount returns the count of online/offline devices.
func (m Model) StatusCount() (online, offline int) {
	for _, s := range m.statuses {
		if s.Online {
			online++
		} else {
			offline++
		}
	}
	return online, offline
}

// SelectedDevice returns the currently selected device, if any.
func (m Model) SelectedDevice() *DeviceStatus {
	if len(m.statuses) == 0 {
		return nil
	}
	if m.cursor < 0 || m.cursor >= len(m.statuses) {
		return nil
	}
	return &m.statuses[m.cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.cursor
}

// IsRefreshing returns true if the monitor is currently refreshing.
func (m Model) IsRefreshing() bool {
	return m.refreshing
}

// IsLoading returns true if the initial load is in progress.
func (m Model) IsLoading() bool {
	return m.initialLoad
}
