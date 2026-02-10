// Package monitor provides real-time device monitoring for the TUI.
package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/afero"
	"github.com/tj-smith47/shelly-go/events"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
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

	// Sensor data (first-value shortcuts for metrics line)
	Temperature *float64 // Temperature in Celsius
	Humidity    *float64 // Relative humidity percentage
	Illuminance *float64 // Illuminance in lux
	Battery     *int     // Battery percentage (if applicable)

	// Full sensor data (all readings from all sensors for Environment panel)
	Sensors *model.SensorData

	// Health indicators (from sys/wifi/component status)
	ChipTemp  *float64 // Component temperature in °C (from switch/cover/light)
	WiFiRSSI  *float64 // WiFi signal strength in dBm
	FSFree    int      // Filesystem free space in bytes
	FSSize    int      // Filesystem total size in bytes
	HasUpdate bool     // Firmware update available

	// Connection info
	ConnectionType string // "ws" for WebSocket, "poll" for HTTP polling

	// Link info (for devices linked to a parent switch)
	LinkState string // Derived state from parent switch (e.g., "Off", "On", "Unknown")
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
	panel.Sizable
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
	refreshInterval time.Duration

	// Energy settings
	costRate float64 // Cost per kWh
	currency string  // Currency symbol
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
		Sizable:         panel.NewSizable(11, panel.NewScroller(0, 10)),
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

		// Global timeout for the full polling cycle. With a concurrency limit of 4
		// and per-device timeout of 5s, 18 devices take ~25s worst case.
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		// Use errgroup for concurrent status fetching with concurrency limit.
		// Without a limit, all devices are polled simultaneously which saturates
		// the home network and causes false "offline" results.
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(4)

		results := make(chan DeviceStatus, len(deviceMap))

		for _, d := range deviceMap {
			device := d
			if device.DisplayName() == "" || device.Address == "" {
				continue // Skip devices with no name or address
			}
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
		Name:           device.DisplayName(),
		Address:        device.Address,
		Type:           device.Type,
		Online:         false,
		UpdatedAt:      time.Now(),
		ConnectionType: "poll", // Default to polling, updated if WebSocket connected
	}

	// Per-device timeout to prevent single slow device from blocking others
	pollCtx, pollCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pollCancel()

	snapshot, err := m.svc.GetMonitoringSnapshotAuto(pollCtx, device.Address)
	if err != nil {
		status.Error = err
		// For linked devices, resolve parent switch state instead of showing "offline".
		// Use a fresh timeout — the poll timeout may have expired reaching this device,
		// but the parent device may still be reachable.
		linkCtx, linkCancel := context.WithTimeout(ctx, 5*time.Second)
		defer linkCancel()
		if ls, linkErr := m.svc.ResolveLinkStatus(linkCtx, device.Name); linkErr == nil {
			status.LinkState = ls.State
			status.Error = nil // Clear the connectivity error — link state is authoritative
		}
		return status
	}

	status.Online = true
	aggregateMetrics(&status, snapshot.PM, false)
	aggregateMetrics(&status, snapshot.EM, true)
	aggregateMetrics(&status, snapshot.EM1, false)

	// Fetch sensor data via full status
	m.fetchSensorData(pollCtx, device.Address, &status)

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

		// Collect all sensor data
		sensorData := shelly.CollectSensorData(statusMap)
		status.Sensors = sensorData

		// Extract first readings as shortcuts for the power ranking metrics line
		if len(sensorData.Temperature) > 0 && sensorData.Temperature[0].TC != nil {
			status.Temperature = sensorData.Temperature[0].TC
		}
		if len(sensorData.Humidity) > 0 && sensorData.Humidity[0].RH != nil {
			status.Humidity = sensorData.Humidity[0].RH
		}
		if len(sensorData.Illuminance) > 0 && sensorData.Illuminance[0].Lux != nil {
			status.Illuminance = sensorData.Illuminance[0].Lux
		}
		if len(sensorData.DevicePower) > 0 {
			status.Battery = &sensorData.DevicePower[0].Battery.Percent
		}

		// Extract health indicators
		extractHealthData(statusMap, status)

		return nil
	})
	if err != nil {
		m.ios.DebugErr("fetch sensor data", err)
	}
}

// Health data extraction structs — minimal shapes for JSON parsing.
type sysHealthData struct {
	FSFree           int  `json:"fs_free"`
	FSSize           int  `json:"fs_size"`
	RestartRequired  bool `json:"restart_required"`
	AvailableUpdates *struct {
		Stable *struct {
			Version string `json:"version"`
		} `json:"stable,omitempty"`
	} `json:"available_updates,omitempty"`
}

type wifiHealthData struct {
	RSSI *float64 `json:"rssi,omitempty"`
}

type componentTempData struct {
	Temperature *struct {
		TC *float64 `json:"tC,omitempty"`
	} `json:"temperature,omitempty"`
}

// extractHealthData extracts sys/wifi/component health indicators from a Shelly.GetStatus response.
func extractHealthData(statusMap map[string]json.RawMessage, status *DeviceStatus) {
	// Extract sys health
	if raw, ok := statusMap["sys"]; ok {
		var sys sysHealthData
		if json.Unmarshal(raw, &sys) == nil {
			status.FSFree = sys.FSFree
			status.FSSize = sys.FSSize
			status.HasUpdate = sys.AvailableUpdates != nil && sys.AvailableUpdates.Stable != nil
		}
	}

	// Extract WiFi RSSI
	if raw, ok := statusMap["wifi"]; ok {
		var wifi wifiHealthData
		if json.Unmarshal(raw, &wifi) == nil {
			status.WiFiRSSI = wifi.RSSI
		}
	}

	// Extract highest chip temp from power components (switch, cover, light, rgb, rgbw)
	for key, raw := range statusMap {
		if !isComponentKey(key) {
			continue
		}
		var comp componentTempData
		if json.Unmarshal(raw, &comp) != nil || comp.Temperature == nil || comp.Temperature.TC == nil {
			continue
		}
		tc := *comp.Temperature.TC
		if status.ChipTemp == nil || tc > *status.ChipTemp {
			status.ChipTemp = &tc
		}
	}
}

// isComponentKey returns true if the key is a power component that may have chip temperature.
func isComponentKey(key string) bool {
	for _, prefix := range []string{"switch:", "cover:", "light:", "rgb:", "rgbw:"} {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// aggregateMetrics aggregates power metrics from any meter type (PM, EM, EM1) into
// a DeviceStatus. The accumulateCurrent flag controls whether current values are
// summed (true, for 3-phase EM meters) or set-first-non-zero (false, for PM/EM1).
// Uses the model.MeterReading interface already implemented by all meter status types.
func aggregateMetrics[T any, PT interface {
	*T
	model.MeterReading
}](status *DeviceStatus, items []T, accumulateCurrent bool) {
	for i := range items {
		m := PT(&items[i])
		status.Power += m.GetPower()
		if status.Voltage == 0 && m.GetVoltage() > 0 {
			status.Voltage = m.GetVoltage()
		}
		if accumulateCurrent {
			status.Current += m.GetCurrent()
		} else if status.Current == 0 && m.GetCurrent() > 0 {
			status.Current = m.GetCurrent()
		}
		if freq := m.GetFreq(); freq != nil && status.Frequency == 0 {
			status.Frequency = *freq
		}
		if energy := m.GetEnergy(); energy != nil {
			status.TotalEnergy += *energy
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
		return m, m.exportData(m.resolveExportFormat(msg.Format))

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

	aggregateMetrics(status, fullStatus.PM, false)
	aggregateMetrics(status, fullStatus.EM, true)
	aggregateMetrics(status, fullStatus.EM1, false)
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not in context system
	if msg.String() == "X" {
		// Export to JSON (uppercase X is component-specific)
		return m, m.exportData(ExportJSON)
	}

	return m, nil
}

// resolveExportFormat converts a string format to ExportFormat.
func (m Model) resolveExportFormat(format string) ExportFormat {
	if format == "json" {
		return ExportJSON
	}
	return ExportCSV
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	m.Scroller.HandleNavigation(msg)
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
		fs := config.Fs()
		if err := fs.MkdirAll(exportDir, 0o750); err != nil {
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
			err = m.exportCSV(fs, path)
		case ExportJSON:
			err = m.exportJSON(fs, path)
		}

		if err != nil {
			return ExportResultMsg{Err: err}
		}

		return ExportResultMsg{Path: path, Format: format}
	}
}

// optionalFloat formats a *float64 for CSV, returning "" if nil.
func optionalFloat(v *float64, format string) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf(format, *v)
}

// optionalInt formats a *int for CSV, returning "" if nil.
func optionalInt(v *int) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

func (m Model) exportCSV(fs afero.Fs, path string) (retErr error) {
	f, err := fs.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && retErr == nil {
			retErr = fmt.Errorf("close file: %w", closeErr)
		}
	}()

	// Write header
	header := "device,address,type,online,power_w,voltage_v,current_a,frequency_hz,energy_wh," +
		"temperature_c,humidity_pct,illuminance_lux,battery_pct," +
		"chip_temp_c,wifi_rssi_dbm,fs_used_pct,has_update," +
		"connection,updated"
	if _, err := fmt.Fprintln(f, header); err != nil {
		return err
	}

	// Write device data
	for _, s := range m.statuses {
		// Health indicators
		chipTemp := optionalFloat(s.ChipTemp, "%.1f")
		wifiRSSI := optionalFloat(s.WiFiRSSI, "%.0f")
		fsUsed := ""
		if s.FSSize > 0 {
			fsUsed = fmt.Sprintf("%d", 100-s.FSFree*100/s.FSSize)
		}

		if _, err := fmt.Fprintf(f, "%s,%s,%s,%t,%.2f,%.2f,%.3f,%.1f,%.2f,%s,%s,%s,%s,%s,%s,%s,%t,%s,%s\n",
			s.Name, s.Address, s.Type, s.Online,
			s.Power, s.Voltage, s.Current, s.Frequency, s.TotalEnergy,
			optionalFloat(s.Temperature, "%.1f"),
			optionalFloat(s.Humidity, "%.0f"),
			optionalFloat(s.Illuminance, "%.0f"),
			optionalInt(s.Battery),
			chipTemp, wifiRSSI, fsUsed, s.HasUpdate,
			s.ConnectionType, s.UpdatedAt.Format(time.RFC3339),
		); err != nil {
			return err
		}
	}

	return nil
}

// exportHealth holds health indicators for JSON export.
type exportHealth struct {
	ChipTemp  *float64 `json:"chip_temp_c,omitempty"`
	WiFiRSSI  *float64 `json:"wifi_rssi_dbm,omitempty"`
	FSUsedPct *int     `json:"fs_used_pct,omitempty"`
	HasUpdate bool     `json:"has_update,omitempty"`
}

// buildExportHealth builds health data for export, returning nil if no health indicators present.
func buildExportHealth(s DeviceStatus) *exportHealth {
	if s.ChipTemp == nil && s.WiFiRSSI == nil && s.FSSize == 0 && !s.HasUpdate {
		return nil
	}
	h := &exportHealth{
		ChipTemp:  s.ChipTemp,
		WiFiRSSI:  s.WiFiRSSI,
		HasUpdate: s.HasUpdate,
	}
	if s.FSSize > 0 {
		pct := 100 - s.FSFree*100/s.FSSize
		h.FSUsedPct = &pct
	}
	return h
}

func (m Model) exportJSON(fs afero.Fs, path string) (retErr error) {
	// Build export structure
	type exportDevice struct {
		Name        string        `json:"name"`
		Address     string        `json:"address"`
		Type        string        `json:"type"`
		Online      bool          `json:"online"`
		Power       float64       `json:"power_w"`
		Voltage     float64       `json:"voltage_v"`
		Current     float64       `json:"current_a"`
		Frequency   float64       `json:"frequency_hz"`
		TotalEnergy float64       `json:"energy_wh"`
		Temperature *float64      `json:"temperature_c,omitempty"`
		Humidity    *float64      `json:"humidity_pct,omitempty"`
		Illuminance *float64      `json:"illuminance_lux,omitempty"`
		Battery     *int          `json:"battery_pct,omitempty"`
		Health      *exportHealth `json:"health,omitempty"`
		Connection  string        `json:"connection"`
		UpdatedAt   string        `json:"updated_at"`
		Error       string        `json:"error,omitempty"`
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

		d.Health = buildExportHealth(s)
		export.Devices = append(export.Devices, d)
	}

	// Write JSON
	f, err := fs.Create(path)
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
	m.Loader = m.Loader.SetSize(width-panel.LoaderBorderOffset, height-panel.LoaderBorderOffset)
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

// formatPower delegates to the shared output.FormatPower formatter.
func formatPower(watts float64) string {
	return output.FormatPower(watts)
}

// formatEnergy delegates to the shared output.FormatEnergy formatter.
func formatEnergy(wh float64) string {
	return output.FormatEnergy(wh)
}

// Statuses returns the current device statuses.
func (m Model) Statuses() []DeviceStatus {
	return m.statuses
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
	return keys.FormatHints([]keys.Hint{
		{Key: "j/k", Desc: "scroll"},
		{Key: "g/G", Desc: "top/btm"},
		{Key: "r", Desc: "refresh"},
		{Key: "x", Desc: "csv"},
		{Key: "X", Desc: "json"},
	}, keys.FooterHintWidth(m.Width))
}
