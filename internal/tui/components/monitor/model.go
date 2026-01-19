// Package monitor provides real-time device monitoring for the TUI.
package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/events"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// Deps holds the dependencies for the monitor component.
type Deps struct {
	Ctx             context.Context
	Svc             *shelly.Service
	IOS             *iostreams.IOStreams
	RefreshInterval time.Duration
	EventStream     *automation.EventStream // Shared event stream (optional - creates one if nil)
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

	// Sensor data
	Temperature *float64 // Temperature in Celsius
	Humidity    *float64 // Relative humidity percentage
	Illuminance *float64 // Illuminance in lux
	Battery     *int     // Battery percentage (if applicable)

	// Connection info
	ConnectionType string // "ws" for WebSocket, "poll" for HTTP polling
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

// ExportFormat represents the export file format.
type ExportFormat int

const (
	// ExportCSV exports data in CSV format.
	ExportCSV ExportFormat = iota
	// ExportJSON exports data in JSON format.
	ExportJSON
)

// ExportResultMsg is sent when an export completes.
type ExportResultMsg struct {
	Path   string
	Format ExportFormat
	Err    error
}

// Model holds the monitor state.
type Model struct {
	helpers.Sizable
	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	statuses        []DeviceStatus
	statusMap       map[string]*DeviceStatus // For O(1) updates by device name
	initialLoad     bool                     // True only on first load (shows loading screen)
	refreshing      bool                     // True during background refresh (shows indicator, keeps data)
	useWebSocket    bool                     // True if using WebSocket for updates
	eventStream     *automation.EventStream  // WebSocket event stream (may be shared)
	ownsEventStream bool                     // True if we created the event stream (so we should stop it)
	eventChan       chan events.Event        // Channel for WebSocket events
	err             error
	styles          Styles
	refreshInterval time.Duration

	// Energy settings
	costRate float64 // Cost per kWh
	currency string  // Currency symbol
}

// Styles for the monitor component.
type Styles struct {
	Container      lipgloss.Style
	Header         lipgloss.Style
	Row            lipgloss.Style
	SelectedRow    lipgloss.Style
	OnlineIcon     lipgloss.Style
	OfflineIcon    lipgloss.Style
	UpdatingIcon   lipgloss.Style
	DeviceName     lipgloss.Style
	Address        lipgloss.Style
	Power          lipgloss.Style
	Voltage        lipgloss.Style
	Current        lipgloss.Style
	Label          lipgloss.Style
	Value          lipgloss.Style
	LastUpdated    lipgloss.Style
	Timestamp      lipgloss.Style // Yellow timestamp for device updates
	Separator      lipgloss.Style
	SummaryCard    lipgloss.Style
	SummaryValue   lipgloss.Style
	Energy         lipgloss.Style
	Temperature    lipgloss.Style
	Humidity       lipgloss.Style
	Battery        lipgloss.Style
	ConnectionType lipgloss.Style
	Cost           lipgloss.Style
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
			Foreground(colors.Text).
			Width(8),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		LastUpdated: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Timestamp: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Separator: lipgloss.NewStyle().
			Foreground(colors.Muted),
		SummaryCard: styles.PanelBorder().
			Padding(0, 1).
			MarginRight(1),
		SummaryValue: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight),
		Energy: lipgloss.NewStyle().
			Foreground(colors.Success),
		Temperature: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Humidity: lipgloss.NewStyle().
			Foreground(colors.Primary),
		Battery: lipgloss.NewStyle().
			Foreground(colors.Success),
		ConnectionType: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Cost: lipgloss.NewStyle().
			Foreground(colors.Success).
			Bold(true),
	}
}

// New creates a new monitor model.
func New(deps Deps) Model {
	if err := deps.validate(); err != nil {
		iostreams.DebugErr("monitor component init", err)
		panic(fmt.Sprintf("monitor: invalid deps: %v", err))
	}

	refreshInterval := deps.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 10 * time.Second // Fallback polling interval for Gen1
	}

	// Use shared EventStream if provided, otherwise create our own
	eventStream := deps.EventStream
	ownsEventStream := false
	if eventStream == nil {
		eventStream = automation.NewEventStream(deps.Svc)
		ownsEventStream = true
	}

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

	// Load energy config for cost calculation
	energyCfg := config.DefaultEnergyConfig()
	if cfg, err := config.Load(); err == nil {
		energyCfg = cfg.GetEnergyConfig()
	}

	m := Model{
		Sizable:         helpers.NewSizable(11, panel.NewScroller(0, 10)),
		ctx:             deps.Ctx,
		svc:             deps.Svc,
		ios:             deps.IOS,
		statuses:        []DeviceStatus{},
		statusMap:       make(map[string]*DeviceStatus),
		initialLoad:     true,
		refreshing:      false,
		useWebSocket:    true,
		eventStream:     eventStream,
		ownsEventStream: ownsEventStream,
		eventChan:       eventChan,
		styles:          DefaultStyles(),
		refreshInterval: refreshInterval,
		costRate:        energyCfg.CostRate,
		currency:        energyCfg.Currency,
	}
	m.Loader = m.Loader.SetMessage("Fetching device statuses...")
	return m
}

// Init returns the initial command for the monitor.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.fetchStatuses(), // Initial HTTP fetch for immediate data
		m.Loader.Init(),   // Start spinner animation
	}

	if m.useWebSocket {
		// Only start event stream if we own it (not shared from app.go)
		if m.ownsEventStream {
			cmds = append(cmds, m.startEventStream())
		}
		// Always listen for events (shared or owned)
		cmds = append(cmds, m.listenForEvents())
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
		Name:           device.Name,
		Address:        device.Address,
		Type:           device.Type,
		Online:         false,
		UpdatedAt:      time.Now(),
		ConnectionType: "poll", // Default to polling, updated if WebSocket connected
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

	// Fetch sensor data via full status
	m.fetchSensorData(ctx, device.Address, &status)

	// Check WebSocket connection status
	if m.eventStream != nil {
		if info := m.eventStream.GetConnectionInfo(device.Name); info.Type == automation.ConnectionWebSocket {
			status.ConnectionType = "ws"
		}
	}

	return status
}

// fetchSensorData fetches sensor readings for a device.
func (m Model) fetchSensorData(ctx context.Context, address string, status *DeviceStatus) {
	// Use WithConnection to make a Shelly.GetStatus call
	err := m.svc.WithConnection(ctx, address, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
		if err != nil {
			return err
		}

		// Re-marshal the result to JSON bytes for parsing
		resultBytes, err := json.Marshal(result)
		if err != nil {
			return err
		}

		// Parse result as map
		var statusMap map[string]json.RawMessage
		if err := json.Unmarshal(resultBytes, &statusMap); err != nil {
			return err
		}

		// Collect sensor data
		sensorData := shelly.CollectSensorData(statusMap)

		// Extract first temperature reading if available
		if len(sensorData.Temperature) > 0 && sensorData.Temperature[0].TC != nil {
			status.Temperature = sensorData.Temperature[0].TC
		}

		// Extract first humidity reading if available
		if len(sensorData.Humidity) > 0 && sensorData.Humidity[0].RH != nil {
			status.Humidity = sensorData.Humidity[0].RH
		}

		// Extract first illuminance reading if available
		if len(sensorData.Illuminance) > 0 && sensorData.Illuminance[0].Lux != nil {
			status.Illuminance = sensorData.Illuminance[0].Lux
		}

		// Check for devicepower (battery) component
		for key, raw := range statusMap {
			if strings.HasPrefix(key, "devicepower:") {
				var dp model.DevicePowerReading
				if err := json.Unmarshal(raw, &dp); err == nil {
					status.Battery = &dp.Battery.Percent
				}
				break
			}
		}

		return nil
	})
	if err != nil {
		m.ios.DebugErr("fetch sensor data", err)
	}
}

// aggregateStatusFromPM aggregates status data from PM components.
func aggregateStatusFromPM(status *DeviceStatus, pms []model.PMStatus) {
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
func aggregateStatusFromEM(status *DeviceStatus, ems []model.EMStatus) {
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
func aggregateStatusFromEM1(status *DeviceStatus, em1s []model.EM1Status) {
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
	// Update loader for spinner animation during initial load
	if m.initialLoad {
		var loaderCmd tea.Cmd
		m.Loader, loaderCmd = m.Loader.Update(msg)
		if loaderCmd != nil {
			// Continue to process StatusUpdateMsg even during loading
			if statusMsg, ok := msg.(StatusUpdateMsg); ok {
				return m.handleStatusUpdate(statusMsg, loaderCmd)
			}
			return m, loaderCmd
		}
	}

	switch msg := msg.(type) {
	case StatusUpdateMsg:
		return m.handleStatusUpdate(msg, nil)

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

	// Action messages from context-based keybindings
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.RefreshRequestMsg:
		if m.refreshing {
			return m, nil
		}
		m.refreshing = true
		return m, m.fetchStatuses()
	case messages.ExportRequestMsg:
		return m, m.exportData(ExportCSV)

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleDeviceEvent processes a WebSocket event and updates device status.
// handleStatusUpdate processes status update messages.
func (m Model) handleStatusUpdate(msg StatusUpdateMsg, additionalCmd tea.Cmd) (Model, tea.Cmd) {
	m.initialLoad = false
	m.refreshing = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, additionalCmd
	}
	m.statuses = msg.Statuses
	m.Scroller.SetItemCount(len(m.statuses))
	// Build status map for O(1) updates
	m.statusMap = make(map[string]*DeviceStatus, len(m.statuses))
	for i := range m.statuses {
		m.statusMap[m.statuses[i].Name] = &m.statuses[i]
	}
	return m, additionalCmd
}

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
	var fullStatus model.MonitoringSnapshot
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

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not in context system
	if msg.String() == "X" {
		// Export to JSON (uppercase X is component-specific)
		return m, m.exportData(ExportJSON)
	}

	return m, nil
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		m.Scroller.CursorUp()
	case messages.NavDown:
		m.Scroller.CursorDown()
	case messages.NavPageUp:
		m.Scroller.PageUp()
	case messages.NavPageDown:
		m.Scroller.PageDown()
	case messages.NavHome:
		m.Scroller.CursorToStart()
	case messages.NavEnd:
		m.Scroller.CursorToEnd()
	case messages.NavLeft, messages.NavRight:
		// Not applicable for this component
	}
	return m, nil
}

// exportData exports the current monitoring data to a file.
func (m Model) exportData(format ExportFormat) tea.Cmd {
	return func() tea.Msg {
		if len(m.statuses) == 0 {
			return ExportResultMsg{Err: fmt.Errorf("no data to export")}
		}

		// Create export directory
		configDir, err := config.Dir()
		if err != nil {
			return ExportResultMsg{Err: fmt.Errorf("get config dir: %w", err)}
		}
		exportDir := filepath.Join(configDir, "exports", "monitor")
		if err := os.MkdirAll(exportDir, 0o750); err != nil {
			return ExportResultMsg{Err: fmt.Errorf("create export dir: %w", err)}
		}

		// Generate filename
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		ext := "csv"
		if format == ExportJSON {
			ext = "json"
		}
		filename := fmt.Sprintf("monitor_%s.%s", timestamp, ext)
		path := filepath.Join(exportDir, filename)

		// Export based on format
		switch format {
		case ExportCSV:
			err = m.exportCSV(path)
		case ExportJSON:
			err = m.exportJSON(path)
		}

		if err != nil {
			return ExportResultMsg{Err: err}
		}

		return ExportResultMsg{Path: path, Format: format}
	}
}

func (m Model) exportCSV(path string) (retErr error) {
	f, err := os.Create(path) // #nosec G304 -- path is constructed from config dir
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && retErr == nil {
			retErr = fmt.Errorf("close file: %w", closeErr)
		}
	}()

	// Write header
	if _, err := fmt.Fprintln(f, "device,address,type,online,power_w,voltage_v,current_a,frequency_hz,energy_wh,temperature_c,humidity_pct,illuminance_lux,battery_pct,connection,updated"); err != nil {
		return err
	}

	// Write device data
	for _, s := range m.statuses {
		temp := ""
		if s.Temperature != nil {
			temp = fmt.Sprintf("%.1f", *s.Temperature)
		}
		humidity := ""
		if s.Humidity != nil {
			humidity = fmt.Sprintf("%.0f", *s.Humidity)
		}
		lux := ""
		if s.Illuminance != nil {
			lux = fmt.Sprintf("%.0f", *s.Illuminance)
		}
		battery := ""
		if s.Battery != nil {
			battery = fmt.Sprintf("%d", *s.Battery)
		}

		if _, err := fmt.Fprintf(f, "%s,%s,%s,%t,%.2f,%.2f,%.3f,%.1f,%.2f,%s,%s,%s,%s,%s,%s\n",
			s.Name, s.Address, s.Type, s.Online,
			s.Power, s.Voltage, s.Current, s.Frequency, s.TotalEnergy,
			temp, humidity, lux, battery,
			s.ConnectionType, s.UpdatedAt.Format(time.RFC3339),
		); err != nil {
			return err
		}
	}

	return nil
}

func (m Model) exportJSON(path string) (retErr error) {
	// Build export structure
	type exportDevice struct {
		Name        string   `json:"name"`
		Address     string   `json:"address"`
		Type        string   `json:"type"`
		Online      bool     `json:"online"`
		Power       float64  `json:"power_w"`
		Voltage     float64  `json:"voltage_v"`
		Current     float64  `json:"current_a"`
		Frequency   float64  `json:"frequency_hz"`
		TotalEnergy float64  `json:"energy_wh"`
		Temperature *float64 `json:"temperature_c,omitempty"`
		Humidity    *float64 `json:"humidity_pct,omitempty"`
		Illuminance *float64 `json:"illuminance_lux,omitempty"`
		Battery     *int     `json:"battery_pct,omitempty"`
		Connection  string   `json:"connection"`
		UpdatedAt   string   `json:"updated_at"`
		Error       string   `json:"error,omitempty"`
	}

	type exportData struct {
		ExportTime  string         `json:"export_time"`
		TotalPower  float64        `json:"total_power_w"`
		TotalEnergy float64        `json:"total_energy_wh"`
		TotalCost   *float64       `json:"total_cost,omitempty"`
		Currency    string         `json:"currency,omitempty"`
		CostRate    float64        `json:"cost_rate_per_kwh,omitempty"`
		Devices     []exportDevice `json:"devices"`
	}

	// Calculate totals
	var totalPower, totalEnergy float64
	for _, s := range m.statuses {
		if s.Online {
			totalPower += s.Power
			totalEnergy += s.TotalEnergy
		}
	}

	export := exportData{
		ExportTime:  time.Now().Format(time.RFC3339),
		TotalPower:  totalPower,
		TotalEnergy: totalEnergy,
		CostRate:    m.costRate,
		Currency:    m.currency,
		Devices:     make([]exportDevice, 0, len(m.statuses)),
	}

	if m.costRate > 0 && totalEnergy > 0 {
		cost := (totalEnergy / 1000) * m.costRate
		export.TotalCost = &cost
	}

	for _, s := range m.statuses {
		d := exportDevice{
			Name:        s.Name,
			Address:     s.Address,
			Type:        s.Type,
			Online:      s.Online,
			Power:       s.Power,
			Voltage:     s.Voltage,
			Current:     s.Current,
			Frequency:   s.Frequency,
			TotalEnergy: s.TotalEnergy,
			Temperature: s.Temperature,
			Humidity:    s.Humidity,
			Illuminance: s.Illuminance,
			Battery:     s.Battery,
			Connection:  s.ConnectionType,
			UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
		}
		if s.Error != nil {
			d.Error = s.Error.Error()
		}
		export.Devices = append(export.Devices, d)
	}

	// Write JSON
	f, err := os.Create(path) // #nosec G304 -- path is constructed from config dir
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && retErr == nil {
			retErr = fmt.Errorf("close file: %w", closeErr)
		}
	}()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(export); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.Width = width
	m.Height = height
	m.Loader = m.Loader.SetSize(width-helpers.LoaderBorderOffset, height-helpers.LoaderBorderOffset)
	// Account for: header (2), summary (4), empty line (1), footer (2), container padding (2) = 11 lines overhead
	// Each card: 2 lines content + 1 margin + 1 separator = 4 lines
	availableHeight := height - 11
	visibleRows := (availableHeight + 1) / 4
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.Scroller.SetVisibleRows(visibleRows)
	return m
}

// View renders the monitor.
func (m Model) View() string {
	if m.initialLoad {
		return m.styles.Container.
			Width(m.Width).
			Height(m.Height).
			Render(m.Loader.SetSize(m.Width, m.Height).View())
	}

	if m.err != nil {
		return m.styles.Container.
			Width(m.Width).
			Render(theme.StatusError().Render("Error: " + m.err.Error()))
	}

	if len(m.statuses) == 0 {
		return m.styles.Container.
			Width(m.Width).
			Height(m.Height).
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

	// Calculate cost
	costStr := ""
	if m.costRate > 0 && totalEnergy > 0 {
		cost := (totalEnergy / 1000) * m.costRate // Convert Wh to kWh
		costStr = fmt.Sprintf("%s%.2f", m.currency, cost)
	}

	// Summary cards row
	summaryCards := []string{
		m.renderSummaryCard("Power", m.formatPower(totalPower)),
		m.renderSummaryCard("Energy", m.formatEnergy(totalEnergy)),
		m.renderSummaryCard("Devices", fmt.Sprintf("%d/%d online", onlineCount, len(m.statuses))),
	}
	if costStr != "" {
		summaryCards = append(summaryCards, m.renderSummaryCard("Cost", costStr))
	}
	summary := lipgloss.JoinHorizontal(lipgloss.Top, summaryCards...)

	// Get visible range from scroller
	startIdx, endIdx := m.Scroller.VisibleRange()
	if endIdx > len(m.statuses) {
		endIdx = len(m.statuses)
	}

	// Calculate content width (container width minus padding)
	contentWidth := m.Width - 6 // 2 padding + 2 border on each side

	var rows string
	for i := startIdx; i < endIdx; i++ {
		s := m.statuses[i]
		isSelected := m.Scroller.IsCursorAt(i)
		rows += m.renderDeviceCard(s, isSelected)
		if i < endIdx-1 {
			separator := m.styles.Separator.Render(strings.Repeat("─", contentWidth))
			rows += separator + "\n"
		}
	}

	// Scroll indicator
	scrollInfo := ""
	if info := m.Scroller.ScrollInfo(); info != "" {
		scrollInfo = m.styles.LastUpdated.Render(" " + info + " ")
	}

	footer := m.styles.LastUpdated.Render(
		fmt.Sprintf("Last updated: %s", time.Now().Format("15:04:05")),
	) + scrollInfo

	content := lipgloss.JoinVertical(lipgloss.Left, summary, "", rows, footer)
	return m.styles.Container.Width(m.Width).Height(m.Height).Render(content)
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

	// Yellow timestamp showing last update time
	timestamp := m.styles.Timestamp.Render(s.UpdatedAt.Format("15:04:05"))

	// Connection type indicator
	connIndicator := ""
	if s.ConnectionType == "ws" {
		connIndicator = m.styles.ConnectionType.Render(" [ws]")
	}

	// First line: selection indicator, icon, name, address, type, timestamp, connection
	line1 := fmt.Sprintf("%s%s %s  %s  %s  %s%s",
		selIndicator,
		statusIcon,
		m.styles.DeviceName.Render(s.Name),
		m.styles.Address.Render(s.Address),
		m.styles.Address.Render(s.Type),
		timestamp,
		connIndicator,
	)

	if !s.Online {
		errMsg := "unreachable"
		if s.Error != nil {
			errMsg = output.Truncate(s.Error.Error(), 40)
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

	// Sensor metrics
	if s.Temperature != nil {
		metrics = append(metrics, m.styles.Temperature.Render(fmt.Sprintf("%.1f°C", *s.Temperature)))
	}
	if s.Humidity != nil {
		metrics = append(metrics, m.styles.Humidity.Render(fmt.Sprintf("%.0f%%", *s.Humidity)))
	}
	if s.Illuminance != nil {
		metrics = append(metrics, m.styles.Value.Render(fmt.Sprintf("%.0flux", *s.Illuminance)))
	}
	if s.Battery != nil {
		style := m.styles.Battery
		if *s.Battery < 20 {
			style = theme.StatusError() // Low battery warning
		}
		metrics = append(metrics, style.Render(fmt.Sprintf("🔋%d%%", *s.Battery)))
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
	cursor := m.Scroller.Cursor()
	if len(m.statuses) == 0 || cursor < 0 || cursor >= len(m.statuses) {
		return nil
	}
	return &m.statuses[cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.Scroller.Cursor()
}

// IsRefreshing returns true if the monitor is currently refreshing.
func (m Model) IsRefreshing() bool {
	return m.refreshing
}

// IsLoading returns true if the initial load is in progress.
func (m Model) IsLoading() bool {
	return m.initialLoad
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom r:refresh x:csv X:json"
}
