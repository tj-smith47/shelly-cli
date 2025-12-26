// Package cache provides a shared device data cache for the TUI.
// All views share this cache to avoid redundant network requests.
package cache

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/tj-smith47/shelly-go/events"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// DeviceData holds cached data for a single device.
type DeviceData struct {
	Device  model.Device
	Online  bool
	Fetched bool // True once we've attempted to fetch status
	Error   error

	// Status data (populated once online)
	Info     *shelly.DeviceInfo
	Snapshot *shelly.MonitoringSnapshot

	// Derived metrics (for quick access)
	Power       float64
	Voltage     float64
	Current     float64
	TotalEnergy float64
	Temperature float64

	// Switch states
	Switches []SwitchState

	UpdatedAt time.Time

	// lastRequestID tracks the most recent request for stale response handling.
	// Responses with older request IDs are discarded.
	lastRequestID uint64
}

// SwitchState holds the state of a switch component.
type SwitchState struct {
	ID     int
	On     bool
	Source string
}

// DeviceUpdateMsg is sent when a single device's data is updated.
type DeviceUpdateMsg struct {
	Name      string
	Data      *DeviceData
	RequestID uint64 // For stale response handling
}

// AllDevicesLoadedMsg is sent when all devices have been fetched at least once.
type AllDevicesLoadedMsg struct{}

// RefreshTickMsg triggers periodic refresh (deprecated, kept for compatibility).
type RefreshTickMsg struct{}

// DeviceRefreshMsg triggers refresh for a single device.
type DeviceRefreshMsg struct {
	Name string
}

// WaveMsg represents a wave of devices to load.
type WaveMsg struct {
	Wave      int
	Devices   []deviceFetch
	Remaining [][]deviceFetch
}

// WaveCompleteMsg signals a wave has completed loading.
type WaveCompleteMsg struct {
	Wave int
}

// deviceFetch holds info for fetching a single device.
type deviceFetch struct {
	Name   string
	Device model.Device
}

// RefreshConfig holds adaptive refresh intervals.
type RefreshConfig struct {
	Gen1Online   time.Duration // Refresh interval for online Gen1 devices
	Gen1Offline  time.Duration // Refresh interval for offline Gen1 devices
	Gen2Online   time.Duration // Refresh interval for online Gen2 devices
	Gen2Offline  time.Duration // Refresh interval for offline Gen2 devices
	FocusedBoost time.Duration // Faster refresh for currently focused device
}

// DefaultRefreshConfig returns sensible defaults for refresh intervals.
func DefaultRefreshConfig() RefreshConfig {
	return RefreshConfig{
		Gen1Online:   15 * time.Second, // Gen1 responds well but is fragile
		Gen1Offline:  60 * time.Second, // Don't hammer offline Gen1
		Gen2Online:   5 * time.Second,  // Gen2 handles faster polling
		Gen2Offline:  30 * time.Second, // Back off for offline Gen2
		FocusedBoost: 3 * time.Second,  // Fast refresh for focused device
	}
}

// Cache holds shared device data for all TUI components.
type Cache struct {
	mu      sync.RWMutex
	devices map[string]*DeviceData
	order   []string // Sorted device names for consistent display

	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	refreshInterval time.Duration // Base interval (still used for compatibility)
	refreshConfig   RefreshConfig // Adaptive refresh intervals

	// Track fetch progress
	pendingCount int
	initialLoad  bool

	// Wave loading state
	currentWave int

	// Adaptive refresh state
	focusedDevice      string               // Currently focused device gets faster refresh
	deviceRefreshTimes map[string]time.Time // Track last refresh per device

	// Request ID tracking for stale response handling
	requestCounter uint64

	// Event stream for publishing synthetic events (e.g., offline from HTTP failure)
	eventStream *shelly.EventStream
}

// New creates a new shared cache.
func New(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, refreshInterval time.Duration) *Cache {
	return &Cache{
		ctx:                ctx,
		svc:                svc,
		ios:                ios,
		refreshInterval:    refreshInterval,
		refreshConfig:      DefaultRefreshConfig(),
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		initialLoad:        true,
	}
}

// NewWithRefreshConfig creates a cache with custom refresh configuration.
func NewWithRefreshConfig(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, cfg RefreshConfig) *Cache {
	return &Cache{
		ctx:                ctx,
		svc:                svc,
		ios:                ios,
		refreshInterval:    cfg.Gen2Online, // Use Gen2 online as base
		refreshConfig:      cfg,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		initialLoad:        true,
	}
}

// Init returns the initial command to start fetching devices.
func (c *Cache) Init() tea.Cmd {
	return c.loadDevicesWave()
}

// SubscribeToEvents subscribes to a shared EventStream for real-time updates.
// This allows the cache to update device status in real-time when events arrive
// via WebSocket (Gen2+) or polling (Gen1).
// Also stores the event stream to publish synthetic events (e.g., offline from HTTP failure).
func (c *Cache) SubscribeToEvents(es *shelly.EventStream) {
	c.eventStream = es
	es.Subscribe(func(evt events.Event) {
		c.handleDeviceEvent(evt)
	})
}

// handleDeviceEvent processes an event from the shared EventStream.
func (c *Cache) handleDeviceEvent(evt events.Event) {
	deviceID := evt.DeviceID()

	c.mu.Lock()
	defer c.mu.Unlock()

	data, exists := c.devices[deviceID]
	if !exists {
		// Device not in cache, ignore event
		return
	}

	// Update based on event type
	switch e := evt.(type) {
	case *events.StatusChangeEvent:
		c.updateFromStatusChange(data, e)
	case *events.FullStatusEvent:
		c.updateFromFullStatus(data, e)
	case *events.DeviceOnlineEvent:
		data.Online = true
		data.UpdatedAt = evt.Timestamp()
	case *events.DeviceOfflineEvent:
		data.Online = false
		data.UpdatedAt = evt.Timestamp()
	}

	c.ios.DebugCat(iostreams.CategoryDevice, "cache: event update for %s, type=%s", deviceID, evt.Type())
}

// statusChangeData holds parsed data from a status change event.
type statusChangeData struct {
	Apower  *float64                 `json:"apower"`
	Voltage *float64                 `json:"voltage"`
	Current *float64                 `json:"current"`
	AEnergy *struct{ Total float64 } `json:"aenergy"`
	Output  *bool                    `json:"output"`
	Temp    *struct{ TC float64 }    `json:"temperature"`
}

// updateFromStatusChange updates cache from a status change event.
func (c *Cache) updateFromStatusChange(data *DeviceData, evt *events.StatusChangeEvent) {
	data.Online = true
	data.UpdatedAt = evt.Timestamp()

	if evt.Status == nil {
		return
	}

	var status statusChangeData
	if err := json.Unmarshal(evt.Status, &status); err != nil {
		return
	}

	// Extract switch ID from component name (e.g., "switch:0" -> 0)
	switchID := -1
	if strings.HasPrefix(evt.Component, "switch:") {
		idStr := strings.TrimPrefix(evt.Component, "switch:")
		if id, err := strconv.Atoi(idStr); err == nil {
			switchID = id
		}
	}

	c.applyStatusChangeMetrics(data, &status, switchID)
}

// applyStatusChangeMetrics applies parsed metrics to device data.
// switchID is -1 if not a switch component, otherwise the switch ID (0, 1, etc.).
func (c *Cache) applyStatusChangeMetrics(data *DeviceData, status *statusChangeData, switchID int) {
	if status.Apower != nil {
		data.Power = *status.Apower
	}
	if status.Voltage != nil {
		data.Voltage = *status.Voltage
	}
	if status.Current != nil {
		data.Current = *status.Current
	}
	if status.AEnergy != nil {
		data.TotalEnergy = status.AEnergy.Total
	}
	if status.Temp != nil {
		data.Temperature = status.Temp.TC
	}
	// Update switch states if this is a switch component
	if status.Output != nil && switchID >= 0 {
		// Find and update the correct switch by ID
		for i := range data.Switches {
			if data.Switches[i].ID == switchID {
				data.Switches[i].On = *status.Output
				return
			}
		}
		// Switch not found - add it (shouldn't normally happen, but handles edge cases)
		data.Switches = append(data.Switches, SwitchState{ID: switchID, On: *status.Output})
	}
}

// parsedComponents holds parsed power component data from a full status event.
type parsedComponents struct {
	pm  []shelly.PMStatus
	em  []shelly.EMStatus
	em1 []shelly.EM1Status
}

// updateFromFullStatus updates cache from a full status event.
// Parses the complete device status and extracts all power/energy metrics.
func (c *Cache) updateFromFullStatus(data *DeviceData, evt *events.FullStatusEvent) {
	data.Online = true
	data.UpdatedAt = evt.Timestamp()

	if evt.Status == nil {
		return
	}

	var statusMap map[string]json.RawMessage
	if err := json.Unmarshal(evt.Status, &statusMap); err != nil {
		c.ios.DebugErr("cache: parse full status", err)
		return
	}

	c.resetDeviceMetrics(data)
	components := c.parseComponentsFromStatus(data, statusMap)
	c.aggregateComponentMetrics(data, components)
	c.updateDeviceSnapshot(data, components)
}

// resetDeviceMetrics clears aggregated values before re-calculating.
func (c *Cache) resetDeviceMetrics(data *DeviceData) {
	data.Power = 0
	data.Voltage = 0
	data.Current = 0
	data.TotalEnergy = 0
	data.Temperature = 0
}

// parseComponentsFromStatus parses component statuses from the status map.
func (c *Cache) parseComponentsFromStatus(data *DeviceData, statusMap map[string]json.RawMessage) parsedComponents {
	var components parsedComponents

	for key, rawStatus := range statusMap {
		c.parseComponentByType(data, key, rawStatus, &components)
	}
	return components
}

// parseComponentByType routes parsing to the appropriate handler based on component type.
func (c *Cache) parseComponentByType(data *DeviceData, key string, rawStatus json.RawMessage, components *parsedComponents) {
	switch {
	case strings.HasPrefix(key, "switch:"):
		c.parseSwitchStatus(data, rawStatus)
	case strings.HasPrefix(key, "pm:"), strings.HasPrefix(key, "pm1:"):
		if pm := c.parsePMStatus(rawStatus); pm != nil {
			components.pm = append(components.pm, *pm)
		}
	case strings.HasPrefix(key, "em:"):
		if em := c.parseEMStatus(rawStatus); em != nil {
			components.em = append(components.em, *em)
		}
	case strings.HasPrefix(key, "em1:"):
		if em1 := c.parseEM1Status(rawStatus); em1 != nil {
			components.em1 = append(components.em1, *em1)
		}
	case strings.HasPrefix(key, "temperature:"):
		c.parseTemperatureStatus(data, rawStatus)
	case key == "sys":
		c.parseSysStatus(data, rawStatus)
	}
}

// parsePMStatus parses a PM component status.
func (c *Cache) parsePMStatus(rawStatus json.RawMessage) *shelly.PMStatus {
	var pm shelly.PMStatus
	if err := json.Unmarshal(rawStatus, &pm); err != nil {
		return nil
	}
	return &pm
}

// parseEMStatus parses an EM component status.
func (c *Cache) parseEMStatus(rawStatus json.RawMessage) *shelly.EMStatus {
	var em shelly.EMStatus
	if err := json.Unmarshal(rawStatus, &em); err != nil {
		return nil
	}
	return &em
}

// parseEM1Status parses an EM1 component status.
func (c *Cache) parseEM1Status(rawStatus json.RawMessage) *shelly.EM1Status {
	var em1 shelly.EM1Status
	if err := json.Unmarshal(rawStatus, &em1); err != nil {
		return nil
	}
	return &em1
}

// aggregateComponentMetrics aggregates metrics from parsed components into device data.
func (c *Cache) aggregateComponentMetrics(data *DeviceData, components parsedComponents) {
	c.aggregatePMMetrics(data, components.pm)
	c.aggregateEMMetrics(data, components.em)
	c.aggregateEM1Metrics(data, components.em1)
}

// aggregatePMMetrics aggregates PM component metrics.
func (c *Cache) aggregatePMMetrics(data *DeviceData, pms []shelly.PMStatus) {
	for _, pm := range pms {
		data.Power += pm.APower
		if data.Voltage == 0 && pm.Voltage > 0 {
			data.Voltage = pm.Voltage
		}
		if data.Current == 0 && pm.Current > 0 {
			data.Current = pm.Current
		}
		if pm.AEnergy != nil {
			data.TotalEnergy += pm.AEnergy.Total
		}
	}
}

// aggregateEMMetrics aggregates EM component metrics.
func (c *Cache) aggregateEMMetrics(data *DeviceData, ems []shelly.EMStatus) {
	for _, em := range ems {
		data.Power += em.TotalActivePower
		data.Current += em.TotalCurrent
		if data.Voltage == 0 && em.AVoltage > 0 {
			data.Voltage = em.AVoltage
		}
	}
}

// aggregateEM1Metrics aggregates EM1 component metrics.
func (c *Cache) aggregateEM1Metrics(data *DeviceData, em1s []shelly.EM1Status) {
	for _, em1 := range em1s {
		data.Power += em1.ActPower
		if data.Voltage == 0 && em1.Voltage > 0 {
			data.Voltage = em1.Voltage
		}
		if data.Current == 0 && em1.Current > 0 {
			data.Current = em1.Current
		}
	}
}

// updateDeviceSnapshot updates the device snapshot with parsed component data.
func (c *Cache) updateDeviceSnapshot(data *DeviceData, components parsedComponents) {
	if len(components.pm) == 0 && len(components.em) == 0 && len(components.em1) == 0 {
		return
	}
	if data.Snapshot == nil {
		data.Snapshot = &shelly.MonitoringSnapshot{}
	}
	data.Snapshot.PM = components.pm
	data.Snapshot.EM = components.em
	data.Snapshot.EM1 = components.em1
	data.Snapshot.Online = true
}

// parseSwitchStatus extracts power metrics from a switch component status.
func (c *Cache) parseSwitchStatus(data *DeviceData, rawStatus json.RawMessage) {
	var sw struct {
		ID      int      `json:"id"`
		Output  bool     `json:"output"`
		APower  *float64 `json:"apower"`
		Voltage *float64 `json:"voltage"`
		Current *float64 `json:"current"`
		AEnergy *struct {
			Total float64 `json:"total"`
		} `json:"aenergy"`
		Temperature *struct {
			TC float64 `json:"tC"`
		} `json:"temperature"`
	}
	if err := json.Unmarshal(rawStatus, &sw); err != nil {
		return
	}

	// Update switch states
	found := false
	for i := range data.Switches {
		if data.Switches[i].ID == sw.ID {
			data.Switches[i].On = sw.Output
			found = true
			break
		}
	}
	if !found {
		data.Switches = append(data.Switches, SwitchState{ID: sw.ID, On: sw.Output})
	}

	// Accumulate power metrics (switches without PM components report these directly)
	if sw.APower != nil {
		data.Power += *sw.APower
	}
	if sw.Voltage != nil && data.Voltage == 0 {
		data.Voltage = *sw.Voltage
	}
	if sw.Current != nil && data.Current == 0 {
		data.Current = *sw.Current
	}
	if sw.AEnergy != nil {
		data.TotalEnergy += sw.AEnergy.Total
	}
	if sw.Temperature != nil && data.Temperature == 0 {
		data.Temperature = sw.Temperature.TC
	}
}

// parseTemperatureStatus extracts temperature from a temperature component.
func (c *Cache) parseTemperatureStatus(data *DeviceData, rawStatus json.RawMessage) {
	var temp struct {
		TC float64 `json:"tC"`
		TF float64 `json:"tF"`
	}
	if err := json.Unmarshal(rawStatus, &temp); err == nil && data.Temperature == 0 {
		data.Temperature = temp.TC
	}
}

// parseSysStatus extracts system info like uptime and temperature.
func (c *Cache) parseSysStatus(data *DeviceData, rawStatus json.RawMessage) {
	var sys struct {
		DeviceTemp *struct {
			TC float64 `json:"tC"`
		} `json:"device_temp"`
	}
	if err := json.Unmarshal(rawStatus, &sys); err == nil && sys.DeviceTemp != nil && data.Temperature == 0 {
		data.Temperature = sys.DeviceTemp.TC
	}
}

// scheduleRefresh schedules the next refresh tick (deprecated - use scheduleDeviceRefresh).
func (c *Cache) scheduleRefresh() tea.Cmd {
	return tea.Tick(c.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// scheduleDeviceRefresh schedules refresh for a single device based on its state.
// Adds random jitter (0-50% of interval) to prevent all devices refreshing simultaneously.
func (c *Cache) scheduleDeviceRefresh(name string, data *DeviceData) tea.Cmd {
	interval := c.getRefreshInterval(data)
	// Add 0-50% jitter to stagger requests and prevent rate limiter congestion
	// #nosec G404 - cryptographic randomness not needed for request jittering
	jitter := time.Duration(rand.Int64N(int64(interval / 2)))
	return tea.Tick(interval+jitter, func(time.Time) tea.Msg {
		return DeviceRefreshMsg{Name: name}
	})
}

// getRefreshInterval returns the appropriate refresh interval for a device.
func (c *Cache) getRefreshInterval(data *DeviceData) time.Duration {
	// Focused device gets faster refresh
	if data != nil && data.Device.Name != "" && data.Device.Name == c.focusedDevice {
		return c.refreshConfig.FocusedBoost
	}

	// No data or info - use default Gen2 online interval
	if data == nil {
		return c.refreshConfig.Gen2Online
	}

	// Unknown generation - check online status only
	if data.Info == nil {
		if data.Online {
			return c.refreshConfig.Gen2Online
		}
		return c.refreshConfig.Gen2Offline
	}

	// Gen1 device intervals
	if data.Info.Generation == 1 {
		if data.Online {
			return c.refreshConfig.Gen1Online
		}
		return c.refreshConfig.Gen1Offline
	}

	// Gen2+ device intervals
	if data.Online {
		return c.refreshConfig.Gen2Online
	}
	return c.refreshConfig.Gen2Offline
}

// SetFocusedDevice sets the currently focused device for faster refresh.
func (c *Cache) SetFocusedDevice(name string) tea.Cmd {
	c.mu.Lock()
	old := c.focusedDevice
	c.focusedDevice = name
	c.mu.Unlock()

	if name != old && name != "" {
		// Trigger immediate refresh of newly focused device
		c.mu.RLock()
		data := c.devices[name]
		c.mu.RUnlock()
		if data != nil {
			return c.fetchDeviceWithID(name, data.Device)
		}
	}
	return nil
}

// GetFocusedDevice returns the currently focused device name.
func (c *Cache) GetFocusedDevice() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.focusedDevice
}

// loadDevicesWave initiates wave-based device loading.
func (c *Cache) loadDevicesWave() tea.Cmd {
	return func() tea.Msg {
		deviceMap := config.ListDevices()
		if len(deviceMap) == 0 {
			return AllDevicesLoadedMsg{}
		}

		c.mu.Lock()
		// Initialize all devices with "fetching" state
		c.order = make([]string, 0, len(deviceMap))
		for name, dev := range deviceMap {
			c.devices[name] = &DeviceData{
				Device:  dev,
				Fetched: false,
			}
			c.order = append(c.order, name)
		}
		// Sort for consistent display
		sortStrings(c.order)
		c.pendingCount = len(deviceMap)
		c.mu.Unlock()

		// Create waves (Gen2 first for resilience)
		waves := createWaves(deviceMap)
		if len(waves) == 0 {
			return AllDevicesLoadedMsg{}
		}

		// Start first wave immediately
		return WaveMsg{
			Wave:      0,
			Devices:   waves[0],
			Remaining: waves[1:],
		}
	}
}

// createWaves creates device loading waves with Gen2 devices first.
// Gen2 devices (ESP32) are more resilient and can handle concurrent requests better.
// Wave sizing: first wave is 3 devices for quick UI feedback, then 2 per wave.
func createWaves(devices map[string]model.Device) [][]deviceFetch {
	if len(devices) == 0 {
		return nil
	}

	// Convert to slice for sorting
	fetches := make([]deviceFetch, 0, len(devices))
	for name, dev := range devices {
		fetches = append(fetches, deviceFetch{Name: name, Device: dev})
	}

	// Sort: Gen2 first (generation > 1 or unknown), then Gen1, then by name
	sort.Slice(fetches, func(i, j int) bool {
		iGen := fetches[i].Device.Generation
		jGen := fetches[j].Device.Generation
		// Gen2+ first (generation 0 = unknown, treat as Gen2)
		iGen1 := iGen == 1
		jGen1 := jGen == 1
		if iGen1 != jGen1 {
			return !iGen1 // Gen2 first
		}
		return fetches[i].Name < fetches[j].Name
	})

	// Determine wave sizes
	// First wave: 3 devices for quick feedback
	// Subsequent waves: 2 devices to avoid overloading
	firstWaveSize := 3
	subsequentWaveSize := 2

	if len(fetches) <= firstWaveSize {
		return [][]deviceFetch{fetches}
	}

	waves := make([][]deviceFetch, 0)
	waves = append(waves, fetches[:firstWaveSize])
	remaining := fetches[firstWaveSize:]

	for len(remaining) > 0 {
		size := subsequentWaveSize
		if size > len(remaining) {
			size = len(remaining)
		}
		waves = append(waves, remaining[:size])
		remaining = remaining[size:]
	}

	return waves
}

// processWave processes a wave of device fetches.
func (c *Cache) processWave(msg WaveMsg) tea.Cmd {
	// Fetch all devices in this wave
	cmds := make([]tea.Cmd, 0, len(msg.Devices)+1)
	for _, df := range msg.Devices {
		cmds = append(cmds, c.fetchDeviceWithID(df.Name, df.Device))
	}

	// Schedule next wave with delay if there are more
	if len(msg.Remaining) > 0 {
		cmds = append(cmds, tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
			return WaveMsg{
				Wave:      msg.Wave + 1,
				Devices:   msg.Remaining[0],
				Remaining: msg.Remaining[1:],
			}
		}))
	}

	return tea.Batch(cmds...)
}

// fetchDevice fetches status for a single device (deprecated - use fetchDeviceWithID).
func (c *Cache) fetchDevice(name string, device model.Device) tea.Cmd {
	return c.fetchDeviceWithID(name, device)
}

// fetchDeviceWithID fetches status for a single device with request ID tracking.
func (c *Cache) fetchDeviceWithID(name string, device model.Device) tea.Cmd {
	requestID := atomic.AddUint64(&c.requestCounter, 1)

	return func() tea.Msg {
		data := &DeviceData{
			Device:        device,
			Fetched:       true,
			UpdatedAt:     time.Now(),
			lastRequestID: requestID,
		}

		// Per-device timeout based on known generation.
		// Gen1 (ESP8266) needs more time than Gen2 (ESP32).
		// Unknown generation uses 15s to allow for Gen2->Gen1 fallback in DeviceInfoAuto.
		timeout := 15 * time.Second
		if device.Generation == 1 {
			timeout = 20 * time.Second // Gen1 is slower
		} else if device.Generation >= 2 {
			timeout = 10 * time.Second // Gen2+ is faster
		}
		ctx, cancel := context.WithTimeout(c.ctx, timeout)
		defer cancel()

		// Get device info first (auto-detects Gen1 vs Gen2)
		// Pass device name (not address) so resolver finds stored generation
		info, err := c.svc.DeviceInfoAuto(ctx, name)
		if err != nil {
			data.Error = err
			data.Online = false
			return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
		}

		data.Info = info
		data.Online = true

		// Populate missing device info from DeviceInfo response.
		// When a device is added with just name+address, we discover Type/Model/Generation
		// on first fetch and persist them so subsequent sessions don't re-detect.
		// Model = human-readable product name (e.g., "Shelly Pro 1PM")
		// Type = model code/SKU (e.g., "SPSW-001PE16EU")
		var updates config.DeviceUpdates
		if info.Model != "" {
			displayName := types.ModelDisplayName(info.Model)
			// Populate Model if empty or if Model==Type (old bug where both had same value)
			if data.Device.Model == "" || data.Device.Model == data.Device.Type {
				data.Device.Model = displayName
				updates.Model = displayName
			}
			// Populate Type if empty
			if data.Device.Type == "" {
				data.Device.Type = info.Model
				updates.Type = info.Model
			}
		}
		// Populate generation if not yet known (0 = unset)
		if data.Device.Generation == 0 && info.Generation > 0 {
			data.Device.Generation = info.Generation
			updates.Generation = info.Generation
		}
		// Persist discovered info to config so it's available immediately on next session
		if updates.Type != "" || updates.Model != "" || updates.Generation > 0 {
			if err := config.UpdateDeviceInfo(name, updates); err != nil {
				c.ios.DebugErr("persist device info", err)
			}
		}

		// Get switch states (Gen2+ only - Gen1 uses different relay API)
		if info.Generation > 1 {
			switches, err := c.svc.SwitchList(ctx, name)
			if err == nil {
				for _, sw := range switches {
					data.Switches = append(data.Switches, SwitchState{
						ID: sw.ID,
						On: sw.Output,
					})
				}
			}
		}

		// Get monitoring snapshot for power metrics (auto-detects Gen1 vs Gen2)
		snapshot, err := c.svc.GetMonitoringSnapshotAuto(ctx, name)
		if err != nil {
			// Device is online but couldn't get snapshot - that's OK for non-metering devices
			c.ios.DebugErr("cache snapshot "+name, err)
		} else {
			data.Snapshot = snapshot
			c.aggregateMetrics(data, snapshot)
		}

		return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
	}
}

// handleDeviceUpdate processes a device update message.
func (c *Cache) handleDeviceUpdate(msg DeviceUpdateMsg) tea.Cmd {
	c.mu.Lock()
	existing := c.devices[msg.Name]

	// Discard stale responses - only accept newer request IDs
	if existing != nil && msg.RequestID > 0 && msg.RequestID < existing.lastRequestID {
		c.mu.Unlock()
		return nil // Stale response, discard
	}

	// Apply update with preservation logic
	c.applyDeviceUpdate(msg, existing)
	c.deviceRefreshTimes[msg.Name] = time.Now()
	c.pendingCount--
	allDone := c.pendingCount <= 0 && c.initialLoad
	if allDone {
		c.initialLoad = false
	}
	c.mu.Unlock()

	// Schedule next refresh for this device (adaptive interval)
	var cmds []tea.Cmd
	cmds = append(cmds, c.scheduleDeviceRefresh(msg.Name, msg.Data))

	if allDone {
		cmds = append(cmds, func() tea.Msg { return AllDevicesLoadedMsg{} })
	}

	if len(cmds) == 1 {
		return cmds[0]
	}
	return tea.Batch(cmds...)
}

// applyDeviceUpdate applies the update, preserving existing data where appropriate.
// Must be called with c.mu held.
func (c *Cache) applyDeviceUpdate(msg DeviceUpdateMsg, existing *DeviceData) {
	// If refresh failed but we have existing good data, preserve it
	if shouldPreserveExisting(msg.Data, existing) {
		wasOnline := existing.Online
		existing.Online = false
		existing.Error = msg.Data.Error
		existing.UpdatedAt = msg.Data.UpdatedAt
		existing.lastRequestID = msg.RequestID
		c.devices[msg.Name] = existing

		// Emit synthetic offline event if device was previously online
		if wasOnline && c.eventStream != nil {
			reason := "HTTP polling failed"
			if msg.Data.Error != nil {
				reason = msg.Data.Error.Error()
			}
			offlineEvt := events.NewDeviceOfflineEvent(msg.Name).
				WithReason(reason).
				WithSource(events.EventSourceLocal)
			c.eventStream.Publish(offlineEvt)
		}
		return
	}

	// Normal update - preserve snapshot if new one is nil
	preserveSnapshotFromExisting(msg.Data, existing)
	c.devices[msg.Name] = msg.Data
}

// shouldPreserveExisting returns true if the existing data should be preserved.
func shouldPreserveExisting(newData, existing *DeviceData) bool {
	return newData.Error != nil && existing != nil && existing.Fetched && existing.Info != nil
}

// preserveSnapshotFromExisting copies snapshot data from existing to new if new is nil or empty.
func preserveSnapshotFromExisting(newData, existing *DeviceData) {
	if existing == nil || existing.Snapshot == nil {
		return
	}
	// Preserve if new snapshot is nil OR empty (no PM/EM/EM1 data)
	newIsEmpty := newData.Snapshot == nil || (len(newData.Snapshot.PM) == 0 && len(newData.Snapshot.EM) == 0 && len(newData.Snapshot.EM1) == 0)
	existingHasData := len(existing.Snapshot.PM) > 0 || len(existing.Snapshot.EM) > 0 || len(existing.Snapshot.EM1) > 0

	if newIsEmpty && existingHasData {
		newData.Snapshot = existing.Snapshot
		if newData.Power == 0 && existing.Power != 0 {
			newData.Power = existing.Power
		}
	}
}

// aggregateMetrics extracts metrics from the monitoring snapshot.
func (c *Cache) aggregateMetrics(data *DeviceData, snapshot *shelly.MonitoringSnapshot) {
	// Power meters (PM) - most common for switches with power monitoring
	for _, pm := range snapshot.PM {
		data.Power += pm.APower
		if data.Voltage == 0 && pm.Voltage > 0 {
			data.Voltage = pm.Voltage
		}
		if data.Current == 0 && pm.Current > 0 {
			data.Current = pm.Current
		}
		if pm.AEnergy != nil {
			data.TotalEnergy += pm.AEnergy.Total
		}
	}

	// Energy meters (EM - 3-phase meters like Pro 3EM)
	for _, em := range snapshot.EM {
		data.Power += em.TotalActivePower
		data.Current += em.TotalCurrent
		if data.Voltage == 0 && em.AVoltage > 0 {
			data.Voltage = em.AVoltage
		}
	}

	// Single-phase energy meters (EM1)
	for _, em1 := range snapshot.EM1 {
		data.Power += em1.ActPower
		if data.Voltage == 0 && em1.Voltage > 0 {
			data.Voltage = em1.Voltage
		}
		if data.Current == 0 && em1.Current > 0 {
			data.Current = em1.Current
		}
	}
}

// Update handles cache-related messages.
func (c *Cache) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case WaveMsg:
		c.mu.Lock()
		c.currentWave = msg.Wave
		c.mu.Unlock()
		return c.processWave(msg)

	case DeviceUpdateMsg:
		return c.handleDeviceUpdate(msg)

	case DeviceRefreshMsg:
		// Refresh a single device
		c.mu.RLock()
		data := c.devices[msg.Name]
		c.mu.RUnlock()

		if data == nil {
			return nil
		}
		return c.fetchDeviceWithID(msg.Name, data.Device)

	case RefreshTickMsg:
		// Legacy: batch refresh all devices (kept for compatibility)
		return tea.Batch(
			c.refreshAllDevices(),
			c.scheduleRefresh(),
		)
	}
	return nil
}

// refreshAllDevices refreshes status for all devices.
func (c *Cache) refreshAllDevices() tea.Cmd {
	return func() tea.Msg {
		c.mu.RLock()
		devicesCopy := make(map[string]model.Device, len(c.devices))
		for name, data := range c.devices {
			devicesCopy[name] = data.Device
		}
		c.mu.RUnlock()

		if len(devicesCopy) == 0 {
			return nil
		}

		c.mu.Lock()
		c.pendingCount = len(devicesCopy)
		c.mu.Unlock()

		cmds := make([]tea.Cmd, 0, len(devicesCopy))
		for name, dev := range devicesCopy {
			cmds = append(cmds, c.fetchDevice(name, dev))
		}
		return tea.BatchMsg(cmds)
	}
}

// GetDevice returns cached data for a device.
func (c *Cache) GetDevice(name string) *DeviceData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.devices[name]
}

// GetAllDevices returns all cached devices in sorted order.
func (c *Cache) GetAllDevices() []*DeviceData {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*DeviceData, 0, len(c.order))
	for _, name := range c.order {
		if data, ok := c.devices[name]; ok {
			result = append(result, data)
		}
	}
	return result
}

// GetOnlineDevices returns only online devices.
func (c *Cache) GetOnlineDevices() []*DeviceData {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*DeviceData, 0, len(c.order))
	for _, name := range c.order {
		if data, ok := c.devices[name]; ok && data.Online {
			result = append(result, data)
		}
	}
	return result
}

// DeviceCount returns total device count.
func (c *Cache) DeviceCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.devices)
}

// OnlineCount returns count of online devices.
func (c *Cache) OnlineCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, data := range c.devices {
		if data.Online {
			count++
		}
	}
	return count
}

// TotalPower returns total power consumption across all devices.
func (c *Cache) TotalPower() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var total float64
	for _, data := range c.devices {
		if data.Online {
			total += data.Power
		}
	}
	return total
}

// IsLoading returns true if initial load is in progress.
func (c *Cache) IsLoading() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialLoad && c.pendingCount > 0
}

// FetchedCount returns the number of devices that have been fetched.
func (c *Cache) FetchedCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, data := range c.devices {
		if data.Fetched {
			count++
		}
	}
	return count
}

// sortStrings sorts a slice of strings in place.
func sortStrings(s []string) {
	for i := range len(s) - 1 {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// NewForTesting creates a cache instance for testing without network dependencies.
func NewForTesting() *Cache {
	return &Cache{
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
	}
}

// SetDeviceForTesting adds a device to the cache for testing purposes.
func (c *Cache) SetDeviceForTesting(device model.Device, online bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := &DeviceData{
		Device:    device,
		Online:    online,
		Fetched:   true,
		UpdatedAt: time.Now(),
	}

	c.devices[device.Name] = data

	// Maintain sorted order
	c.order = make([]string, 0, len(c.devices))
	for name := range c.devices {
		c.order = append(c.order, name)
	}
	sortStrings(c.order)
}
