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

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// DeviceData holds cached data for a single device.
type DeviceData struct {
	Device  model.Device
	Online  bool
	Fetched bool // True once we've attempted to fetch status
	Error   error

	// Status data (populated once online)
	Info     *shelly.DeviceInfo
	Snapshot *model.MonitoringSnapshot

	// Extended status (lazy-loaded on device focus for Gen2+)
	WiFi *shelly.WiFiStatus // WiFi status (SSID, RSSI, IP, etc.)
	Sys  *shelly.SysStatus  // System status (uptime, RAM, FS, etc.)

	// Derived metrics (for quick access)
	Power       float64
	Voltage     float64
	Current     float64
	TotalEnergy float64
	Temperature float64

	// Component states
	Switches []SwitchState
	Lights   []LightState
	Covers   []CoverState

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

// LightState holds the state of a light component.
type LightState struct {
	ID int
	On bool
}

// CoverState holds the state of a cover component.
type CoverState struct {
	ID    int
	State string // "open", "closed", "opening", "closing", "stopped"
}

// DeviceUpdateMsg is sent when a single device's data is updated.
type DeviceUpdateMsg struct {
	Name      string
	Data      *DeviceData
	RequestID uint64 // For stale response handling
}

// AllDevicesLoadedMsg is sent when all devices have been fetched at least once.
type AllDevicesLoadedMsg struct{}

// RefreshTickMsg triggers a one-time refresh of all devices (user-initiated only).
// This is NOT scheduled automatically - automatic refresh uses DeviceRefreshMsg
// with per-device adaptive intervals and jitter to prevent overwhelming devices.
type RefreshTickMsg struct{}

// DeviceRefreshMsg triggers refresh for a single device.
type DeviceRefreshMsg struct {
	Name string
}

// FocusDebounceMsg triggers a debounced fetch for the focused device.
// This prevents rapid scrolling from overwhelming devices with HTTP requests.
type FocusDebounceMsg struct {
	Name string
}

// ExtendedStatusDebounceMsg signals that extended status should be fetched after debounce.
// Handled by app.go to trigger FetchExtendedStatus only after user stops scrolling.
type ExtendedStatusDebounceMsg struct {
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
	version uint64   // Incremented on every device data change for cache invalidation

	ctx           context.Context
	svc           *shelly.Service
	ios           *iostreams.IOStreams
	refreshConfig RefreshConfig // Adaptive refresh intervals

	// FileCache for persisting static device data (model, gen, MAC, components)
	// This data never changes, so we cache it with 24h TTL to avoid redundant HTTP calls
	fileCache *cache.FileCache

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
	eventStream *automation.EventStream

	// MAC-to-IP mapping for instant IP remapping (updated on each device fetch)
	macToIP map[string]string

	// WebSocket connection tracking - devices with active WebSocket don't need HTTP polling
	// Gen2+ devices connected via WebSocket receive real-time updates, so HTTP polling is redundant
	wsConnected map[string]bool
}

// New creates a new shared cache with default refresh configuration.
func New(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, fc *cache.FileCache) *Cache {
	return &Cache{
		ctx:                ctx,
		svc:                svc,
		ios:                ios,
		refreshConfig:      DefaultRefreshConfig(),
		fileCache:          fc,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		wsConnected:        make(map[string]bool),
		initialLoad:        true,
	}
}

// NewWithRefreshConfig creates a cache with custom refresh configuration.
func NewWithRefreshConfig(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, fc *cache.FileCache, cfg RefreshConfig) *Cache {
	return &Cache{
		ctx:                ctx,
		svc:                svc,
		ios:                ios,
		refreshConfig:      cfg,
		fileCache:          fc,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		wsConnected:        make(map[string]bool),
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
// Tracks WebSocket connection state to skip HTTP polling for WebSocket-connected devices.
func (c *Cache) SubscribeToEvents(es *automation.EventStream) {
	c.eventStream = es

	// Subscribe to events with connection state tracking
	es.Subscribe(func(evt events.Event) {
		// Track connection state changes
		deviceID := evt.DeviceID()
		switch evt.(type) {
		case *events.DeviceOnlineEvent:
			// Check if this device is connected via WebSocket
			info := es.GetConnectionInfo(deviceID)
			c.SetWebSocketConnected(deviceID, info.Type == automation.ConnectionWebSocket)
		case *events.DeviceOfflineEvent:
			// Device went offline - no longer has WebSocket connection
			c.SetWebSocketConnected(deviceID, false)
		}

		// Process the event normally
		c.handleDeviceEvent(evt)
	})

	// Initialize connection state for all currently connected devices
	for name, info := range es.GetAllConnectionInfo() {
		c.SetWebSocketConnected(name, info.Type == automation.ConnectionWebSocket)
	}
}

// handleDeviceEvent processes an event from the shared EventStream.
func (c *Cache) handleDeviceEvent(evt events.Event) {
	deviceID := evt.DeviceID()

	debug.TraceLock("cache", "Lock", "handleDeviceEvent:"+deviceID)
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		debug.TraceUnlock("cache", "Lock", "handleDeviceEvent:"+deviceID)
	}()

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
	c.version++ // Increment version for cache invalidation

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
	pm  []model.PMStatus
	em  []model.EMStatus
	em1 []model.EM1Status
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
func (c *Cache) parsePMStatus(rawStatus json.RawMessage) *model.PMStatus {
	var pm model.PMStatus
	if err := json.Unmarshal(rawStatus, &pm); err != nil {
		return nil
	}
	return &pm
}

// parseEMStatus parses an EM component status.
func (c *Cache) parseEMStatus(rawStatus json.RawMessage) *model.EMStatus {
	var em model.EMStatus
	if err := json.Unmarshal(rawStatus, &em); err != nil {
		return nil
	}
	return &em
}

// parseEM1Status parses an EM1 component status.
func (c *Cache) parseEM1Status(rawStatus json.RawMessage) *model.EM1Status {
	var em1 model.EM1Status
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
func (c *Cache) aggregatePMMetrics(data *DeviceData, pms []model.PMStatus) {
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
func (c *Cache) aggregateEMMetrics(data *DeviceData, ems []model.EMStatus) {
	for _, em := range ems {
		data.Power += em.TotalActivePower
		data.Current += em.TotalCurrent
		if data.Voltage == 0 && em.AVoltage > 0 {
			data.Voltage = em.AVoltage
		}
	}
}

// aggregateEM1Metrics aggregates EM1 component metrics.
func (c *Cache) aggregateEM1Metrics(data *DeviceData, em1s []model.EM1Status) {
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
		data.Snapshot = &model.MonitoringSnapshot{}
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

// scheduleDeviceRefresh schedules refresh for a single device based on its state.
// Adds random jitter (0-50% of interval) to prevent all devices refreshing simultaneously.
// Skips scheduling for devices with active WebSocket connections (they receive real-time updates).
func (c *Cache) scheduleDeviceRefresh(name string, data *DeviceData) tea.Cmd {
	// Skip scheduling HTTP refresh if device has active WebSocket
	// WebSocket provides real-time updates, no polling needed
	if c.IsWebSocketConnected(name) {
		debug.TraceEvent("scheduleDeviceRefresh: skipping %s (WebSocket connected)", name)
		return nil
	}

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

// FocusDebounceInterval is the delay before triggering a fetch for a focused device.
// This prevents rapid scrolling from overwhelming devices with HTTP requests.
const FocusDebounceInterval = 250 * time.Millisecond

// SetFocusedDevice sets the currently focused device for faster refresh.
// Uses debouncing to prevent rapid scrolling from triggering many HTTP requests.
func (c *Cache) SetFocusedDevice(name string) tea.Cmd {
	c.mu.Lock()
	old := c.focusedDevice
	c.focusedDevice = name
	c.mu.Unlock()

	if name != old && name != "" {
		// Schedule debounced fetch instead of immediate - prevents scroll spam
		return tea.Tick(FocusDebounceInterval, func(time.Time) tea.Msg {
			return FocusDebounceMsg{Name: name}
		})
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
		// Use Device.Name as key for consistent lookup (not config key)
		c.order = make([]string, 0, len(deviceMap))
		for _, dev := range deviceMap {
			c.devices[dev.Name] = &DeviceData{
				Device:  dev,
				Fetched: false,
			}
			c.order = append(c.order, dev.Name)
		}
		// Sort for consistent display
		sortStrings(c.order)
		c.pendingCount = len(deviceMap)
		c.version++ // Increment version for cache invalidation
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
	// Use Device.Name for consistent lookup (not config key)
	fetches := make([]deviceFetch, 0, len(devices))
	for _, dev := range devices {
		fetches = append(fetches, deviceFetch{Name: dev.Name, Device: dev})
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

		ctx, cancel := context.WithTimeout(c.ctx, c.deviceTimeout(device.Generation))
		defer cancel()

		// Try to get device info from FileCache first (it never changes, 24h TTL)
		info, fromCache := c.tryGetCachedDeviceInfo(name)

		if info == nil {
			// Not cached - fetch from device
			var err error
			info, err = c.svc.DeviceInfoAuto(ctx, name)
			if err != nil {
				data.Error = err
				data.Online = false
				return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
			}
			// Cache for 24 hours - device info never changes
			c.cacheDeviceInfo(name, info)
		} else {
			debug.TraceEvent("cache: using cached device info for %s", name)
		}

		data.Info = info
		data.Online = true

		// Populate device info (applies cached data to device model)
		c.populateDeviceInfo(name, data, info)

		// Get component states (Gen2+ only - Gen1 uses different relay API)
		if info.Generation > 1 {
			// Try to use cached component availability first
			cachedComponents := c.tryGetCachedComponents(name)
			if cachedComponents != nil {
				debug.TraceEvent("cache: using cached components for %s", name)
				c.applyComponentsFromCache(data, cachedComponents)
				// Fetch current states for known components
				c.refreshComponentStates(ctx, name, data, cachedComponents)
			} else {
				// First time - discover components and cache them
				c.ios.DebugCat(iostreams.CategoryDevice, "cache: fetching component states for %s (gen %d)", name, info.Generation)
				c.fetchSwitchStates(ctx, name, data)
				c.fetchLightStates(ctx, name, data)
				c.fetchCoverStates(ctx, name, data)
				c.ios.DebugCat(iostreams.CategoryDevice, "cache: %s components: %d switches, %d lights, %d covers", name, len(data.Switches), len(data.Lights), len(data.Covers))
				// Cache component availability for next time
				c.cacheComponents(name, data)
			}
		} else if !fromCache {
			c.ios.DebugCat(iostreams.CategoryDevice, "cache: skipping component fetch for %s (gen %d)", name, info.Generation)
		}

		// Get monitoring snapshot for power metrics (always needed, it's dynamic data)
		c.fetchMonitoringSnapshot(ctx, name, data)

		return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
	}
}

// deviceTimeout returns the appropriate timeout for a device generation.
func (c *Cache) deviceTimeout(generation int) time.Duration {
	switch generation {
	case 1:
		return 20 * time.Second // Gen1 (ESP8266) is slower
	case 2, 3:
		return 10 * time.Second // Gen2+ (ESP32) is faster
	default:
		return 15 * time.Second // Unknown - allow for fallback detection
	}
}

// populateDeviceInfo fills in missing device info and persists to config.
// Skips entirely if device already has all required info populated.
func (c *Cache) populateDeviceInfo(name string, data *DeviceData, info *shelly.DeviceInfo) {
	// Skip if device already has all info - no need to check or persist
	if data.Device.Model != "" && data.Device.Generation > 0 && data.Device.MAC != "" {
		return
	}

	updates := c.buildDeviceUpdates(data, info)
	if updates.Type != "" || updates.Model != "" || updates.Generation > 0 || updates.MAC != "" {
		if err := config.UpdateDeviceInfo(name, updates); err != nil {
			c.ios.DebugErr("persist device info", err)
		}
	}
}

// buildDeviceUpdates creates device updates from info, updating data in place.
func (c *Cache) buildDeviceUpdates(data *DeviceData, info *shelly.DeviceInfo) config.DeviceUpdates {
	var updates config.DeviceUpdates

	if info.Model != "" {
		displayName := types.ModelDisplayName(info.Model)
		if data.Device.Model == "" || data.Device.Model == data.Device.Type {
			data.Device.Model = displayName
			updates.Model = displayName
		}
		if data.Device.Type == "" {
			data.Device.Type = info.Model
			updates.Type = info.Model
		}
	}
	if data.Device.Generation == 0 && info.Generation > 0 {
		data.Device.Generation = info.Generation
		updates.Generation = info.Generation
	}
	if data.Device.MAC == "" && info.MAC != "" {
		data.Device.MAC = model.NormalizeMAC(info.MAC)
		updates.MAC = info.MAC
	}

	return updates
}

// CachedComponents holds component availability that can be cached.
// Component availability never changes for a device, so we cache it with 24h TTL.
type CachedComponents struct {
	SwitchIDs []int `json:"switch_ids,omitempty"`
	LightIDs  []int `json:"light_ids,omitempty"`
	CoverIDs  []int `json:"cover_ids,omitempty"`
}

// tryGetCachedDeviceInfo attempts to retrieve device info from FileCache.
// Returns nil if not cached or expired.
func (c *Cache) tryGetCachedDeviceInfo(name string) (*shelly.DeviceInfo, bool) {
	if c.fileCache == nil {
		return nil, false
	}

	entry, err := c.fileCache.Get(name, cache.TypeDeviceInfo)
	if err != nil {
		c.ios.DebugErr("cache: get device info from file cache", err)
		return nil, false
	}
	if entry == nil {
		return nil, false
	}

	var info shelly.DeviceInfo
	if err := entry.Unmarshal(&info); err != nil {
		c.ios.DebugErr("cache: unmarshal cached device info", err)
		return nil, false
	}

	return &info, true
}

// cacheDeviceInfo stores device info in FileCache with 24h TTL.
func (c *Cache) cacheDeviceInfo(name string, info *shelly.DeviceInfo) {
	if c.fileCache == nil || info == nil {
		return
	}

	if err := c.fileCache.Set(name, cache.TypeDeviceInfo, info, cache.TTLDeviceInfo); err != nil {
		c.ios.DebugErr("cache: set device info in file cache", err)
	} else {
		debug.TraceEvent("cache: stored device info for %s in file cache", name)
	}
}

// tryGetCachedComponents attempts to retrieve component availability from FileCache.
// Returns nil if not cached or expired.
func (c *Cache) tryGetCachedComponents(name string) *CachedComponents {
	if c.fileCache == nil {
		return nil
	}

	entry, err := c.fileCache.Get(name, cache.TypeComponents)
	if err != nil {
		c.ios.DebugErr("cache: get components from file cache", err)
		return nil
	}
	if entry == nil {
		return nil
	}

	var components CachedComponents
	if err := entry.Unmarshal(&components); err != nil {
		c.ios.DebugErr("cache: unmarshal cached components", err)
		return nil
	}

	return &components
}

// cacheComponents stores component availability in FileCache with 24h TTL.
func (c *Cache) cacheComponents(name string, data *DeviceData) {
	if c.fileCache == nil {
		return
	}

	// Only cache if device has any components
	if len(data.Switches) == 0 && len(data.Lights) == 0 && len(data.Covers) == 0 {
		return
	}

	components := CachedComponents{
		SwitchIDs: make([]int, len(data.Switches)),
		LightIDs:  make([]int, len(data.Lights)),
		CoverIDs:  make([]int, len(data.Covers)),
	}

	for i, sw := range data.Switches {
		components.SwitchIDs[i] = sw.ID
	}
	for i, lt := range data.Lights {
		components.LightIDs[i] = lt.ID
	}
	for i, cv := range data.Covers {
		components.CoverIDs[i] = cv.ID
	}

	if err := c.fileCache.Set(name, cache.TypeComponents, components, cache.TTLComponents); err != nil {
		c.ios.DebugErr("cache: set components in file cache", err)
	} else {
		debug.TraceEvent("cache: stored components for %s in file cache (%d sw, %d lt, %d cv)",
			name, len(components.SwitchIDs), len(components.LightIDs), len(components.CoverIDs))
	}
}

// applyComponentsFromCache initializes component slices from cached availability.
func (c *Cache) applyComponentsFromCache(data *DeviceData, cached *CachedComponents) {
	data.Switches = make([]SwitchState, len(cached.SwitchIDs))
	for i, id := range cached.SwitchIDs {
		data.Switches[i] = SwitchState{ID: id}
	}

	data.Lights = make([]LightState, len(cached.LightIDs))
	for i, id := range cached.LightIDs {
		data.Lights[i] = LightState{ID: id}
	}

	data.Covers = make([]CoverState, len(cached.CoverIDs))
	for i, id := range cached.CoverIDs {
		data.Covers[i] = CoverState{ID: id}
	}
}

// refreshComponentStates fetches current state for known components.
// This avoids the discovery HTTP calls (List methods) by directly getting states.
func (c *Cache) refreshComponentStates(ctx context.Context, name string, data *DeviceData, cached *CachedComponents) {
	// For now, we still need to call the List methods to get current state.
	// Future optimization: call individual GetStatus for each component ID.
	// This is still an improvement because we skip discovery when components haven't changed.

	if len(cached.SwitchIDs) > 0 {
		switches, err := c.svc.SwitchList(ctx, name)
		if err == nil {
			for i, sw := range switches {
				if i < len(data.Switches) {
					data.Switches[i].On = sw.Output
				}
			}
		}
	}

	if len(cached.LightIDs) > 0 {
		lights, err := c.svc.LightList(ctx, name)
		if err == nil {
			for i, lt := range lights {
				if i < len(data.Lights) {
					data.Lights[i].On = lt.Output
				}
			}
		}
	}

	if len(cached.CoverIDs) > 0 {
		covers, err := c.svc.CoverList(ctx, name)
		if err == nil {
			for i, cv := range covers {
				if i < len(data.Covers) {
					data.Covers[i].State = cv.State
				}
			}
		}
	}
}

// fetchSwitchStates fetches switch states for Gen2+ devices.
func (c *Cache) fetchSwitchStates(ctx context.Context, name string, data *DeviceData) {
	switches, err := c.svc.SwitchList(ctx, name)
	if err != nil {
		c.ios.DebugCat(iostreams.CategoryDevice, "cache: switch list for %s failed: %v", name, err)
		return
	}
	c.ios.DebugCat(iostreams.CategoryDevice, "cache: switch list for %s returned %d switches", name, len(switches))
	for _, sw := range switches {
		data.Switches = append(data.Switches, SwitchState{
			ID: sw.ID,
			On: sw.Output,
		})
	}
}

// fetchLightStates fetches light states for Gen2+ devices.
func (c *Cache) fetchLightStates(ctx context.Context, name string, data *DeviceData) {
	lights, err := c.svc.LightList(ctx, name)
	if err != nil {
		c.ios.DebugCat(iostreams.CategoryDevice, "cache: light list for %s failed: %v", name, err)
		return
	}
	c.ios.DebugCat(iostreams.CategoryDevice, "cache: light list for %s returned %d lights", name, len(lights))
	for _, lt := range lights {
		data.Lights = append(data.Lights, LightState{
			ID: lt.ID,
			On: lt.Output,
		})
	}
}

// fetchCoverStates fetches cover states for Gen2+ devices.
func (c *Cache) fetchCoverStates(ctx context.Context, name string, data *DeviceData) {
	covers, err := c.svc.CoverList(ctx, name)
	if err != nil {
		return
	}
	for _, cv := range covers {
		data.Covers = append(data.Covers, CoverState{
			ID:    cv.ID,
			State: cv.State,
		})
	}
}

// fetchMonitoringSnapshot fetches power metrics snapshot.
func (c *Cache) fetchMonitoringSnapshot(ctx context.Context, name string, data *DeviceData) {
	snapshot, err := c.svc.GetMonitoringSnapshotAuto(ctx, name)
	if err != nil {
		c.ios.DebugErr("cache snapshot "+name, err)
		return
	}
	data.Snapshot = snapshot
	c.aggregateMetrics(data, snapshot)
}

// FetchExtendedStatusMsg is sent when extended status fetch completes.
type FetchExtendedStatusMsg struct {
	Name string
	WiFi *shelly.WiFiStatus
	Sys  *shelly.SysStatus
}

// FetchExtendedStatus fetches WiFi and Sys status for a device on-demand.
// This is called lazily when a device is focused in the device info panel.
// Requests are made in parallel to avoid timeout issues.
func (c *Cache) FetchExtendedStatus(name string) tea.Cmd {
	debug.TraceEvent("cache: FetchExtendedStatus called for %s", name)
	return func() tea.Msg {
		msg := FetchExtendedStatusMsg{Name: name}

		// Fetch WiFi and Sys in parallel with independent timeouts
		var wg sync.WaitGroup
		var wifiMu, sysMu sync.Mutex

		wg.Go(func() {
			ctx, cancel := context.WithTimeout(c.ctx, 15*time.Second)
			defer cancel()
			if wifi, err := c.svc.GetWiFiStatus(ctx, name); err == nil {
				wifiMu.Lock()
				msg.WiFi = wifi
				wifiMu.Unlock()
				debug.TraceEvent("cache: wifi status for %s succeeded: SSID=%s RSSI=%d", name, wifi.SSID, wifi.RSSI)
			} else {
				debug.TraceEvent("cache: wifi status for %s failed: %v", name, err)
			}
		})

		wg.Go(func() {
			ctx, cancel := context.WithTimeout(c.ctx, 15*time.Second)
			defer cancel()
			if sys, err := c.svc.GetSysStatus(ctx, name); err == nil {
				sysMu.Lock()
				msg.Sys = sys
				sysMu.Unlock()
				debug.TraceEvent("cache: sys status for %s succeeded: uptime=%d", name, sys.Uptime)
			} else {
				debug.TraceEvent("cache: sys status for %s failed: %v", name, err)
			}
		})

		wg.Wait()
		debug.TraceEvent("cache: FetchExtendedStatus for %s complete: wifi=%v sys=%v", name, msg.WiFi != nil, msg.Sys != nil)
		return msg
	}
}

// HandleExtendedStatus updates device data with extended status.
func (c *Cache) HandleExtendedStatus(msg FetchExtendedStatusMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, exists := c.devices[msg.Name]
	if !exists || data == nil {
		debug.TraceEvent("HandleExtendedStatus: device %s not found in cache (exists=%v)", msg.Name, exists)
		return
	}

	debug.TraceEvent("HandleExtendedStatus: updating %s ptr=%p", msg.Name, data)
	if msg.WiFi != nil {
		data.WiFi = msg.WiFi
		debug.TraceEvent("HandleExtendedStatus: stored WiFi for %s (SSID=%s)", msg.Name, msg.WiFi.SSID)
	}
	if msg.Sys != nil {
		data.Sys = msg.Sys
		debug.TraceEvent("HandleExtendedStatus: stored Sys for %s (uptime=%d)", msg.Name, msg.Sys.Uptime)
	}
	// Verify it's stored
	debug.TraceEvent("HandleExtendedStatus: after store for %s: WiFi=%v Sys=%v", msg.Name, data.WiFi != nil, data.Sys != nil)
	c.version++
}

// handleDeviceUpdate processes a device update message.
func (c *Cache) handleDeviceUpdate(msg DeviceUpdateMsg) tea.Cmd {
	debug.TraceLock("cache", "Lock", "handleDeviceUpdate:"+msg.Name)
	c.mu.Lock()
	existing := c.devices[msg.Name]

	// Discard stale responses - only accept newer request IDs
	if existing != nil && msg.RequestID > 0 && msg.RequestID < existing.lastRequestID {
		c.mu.Unlock()
		debug.TraceUnlock("cache", "Lock", "handleDeviceUpdate:"+msg.Name+":stale")
		return nil // Stale response, discard
	}

	// Apply update with preservation logic - returns deferred event to publish after unlock
	deferredEvent := c.applyDeviceUpdate(msg, existing)
	c.deviceRefreshTimes[msg.Name] = time.Now()
	c.version++ // Increment version for cache invalidation
	c.pendingCount--
	allDone := c.pendingCount <= 0 && c.initialLoad
	if allDone {
		c.initialLoad = false
	}
	c.mu.Unlock()
	debug.TraceUnlock("cache", "Lock", "handleDeviceUpdate:"+msg.Name)

	// Publish deferred event AFTER releasing lock to avoid deadlock
	// (eventStream.Publish is synchronous and may call handleDeviceEvent)
	if deferredEvent != nil && c.eventStream != nil {
		c.eventStream.Publish(deferredEvent)
	}

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
// Returns a deferred event to publish AFTER releasing the lock (to avoid deadlock).
func (c *Cache) applyDeviceUpdate(msg DeviceUpdateMsg, existing *DeviceData) events.Event {
	// If refresh failed but we have existing good data, preserve it
	if shouldPreserveExisting(msg.Data, existing) {
		wasOnline := existing.Online
		existing.Online = false
		existing.Error = msg.Data.Error
		existing.UpdatedAt = msg.Data.UpdatedAt
		existing.lastRequestID = msg.RequestID
		c.devices[msg.Name] = existing

		// Return synthetic offline event if device was previously online
		// (caller will publish after releasing lock to avoid deadlock)
		if wasOnline {
			reason := "HTTP polling failed"
			if msg.Data.Error != nil {
				reason = msg.Data.Error.Error()
			}
			return events.NewDeviceOfflineEvent(msg.Name).
				WithReason(reason).
				WithSource(events.EventSourceLocal)
		}
		return nil
	}

	// Normal update - preserve snapshot and extended status if new one is nil
	preserveSnapshotFromExisting(msg.Data, existing)
	preserveExtendedStatusFromExisting(msg.Data, existing)
	c.devices[msg.Name] = msg.Data

	// Update MAC-to-IP mapping when device is online and has a MAC
	if msg.Data.Online && msg.Data.Device.MAC != "" {
		mac := msg.Data.Device.NormalizedMAC()
		if mac != "" && msg.Data.Device.Address != "" {
			c.macToIP[mac] = msg.Data.Device.Address
		}
	}
	return nil
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

// preserveExtendedStatusFromExisting copies WiFi and Sys status from existing data.
// These are lazily fetched on device selection and must be preserved across refresh cycles.
func preserveExtendedStatusFromExisting(newData, existing *DeviceData) {
	if existing == nil {
		return
	}
	if newData.WiFi == nil && existing.WiFi != nil {
		newData.WiFi = existing.WiFi
	}
	if newData.Sys == nil && existing.Sys != nil {
		newData.Sys = existing.Sys
	}
}

// aggregateMetrics extracts metrics from the monitoring snapshot.
func (c *Cache) aggregateMetrics(data *DeviceData, snapshot *model.MonitoringSnapshot) {
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
		// Skip refresh for devices with active WebSocket (real-time updates)
		if c.IsWebSocketConnected(msg.Name) {
			debug.TraceEvent("DeviceRefreshMsg: skipping %s (WebSocket connected)", msg.Name)
			return nil
		}

		// Refresh a single device
		c.mu.RLock()
		data := c.devices[msg.Name]
		c.mu.RUnlock()

		if data == nil {
			return nil
		}
		return c.fetchDeviceWithID(msg.Name, data.Device)

	case FocusDebounceMsg:
		// Debounced fetch for focused device - only trigger if still focused
		c.mu.RLock()
		currentFocus := c.focusedDevice
		data := c.devices[msg.Name]
		wsConnected := c.wsConnected[msg.Name]
		c.mu.RUnlock()

		// Only fetch if this device is still focused (user stopped scrolling)
		if msg.Name != currentFocus || data == nil {
			return nil
		}

		deviceName := msg.Name // Capture for closure

		// WebSocket provides real-time updates - skip HTTP fetch for device status
		// But still fetch extended status (WiFi/Sys) if not yet loaded
		if wsConnected {
			debug.TraceEvent("FocusDebounceMsg: skipping HTTP for %s (WebSocket connected)", msg.Name)
			// Only fetch extended status if missing
			if data.WiFi == nil || data.Sys == nil {
				return func() tea.Msg { return ExtendedStatusDebounceMsg{Name: deviceName} }
			}
			return nil
		}

		debug.TraceEvent("FocusDebounceMsg: triggering fetch for %s", msg.Name)
		// No WebSocket - need HTTP fetch for device status and extended status
		return tea.Batch(
			c.fetchDeviceWithID(deviceName, data.Device),
			func() tea.Msg { return ExtendedStatusDebounceMsg{Name: deviceName} },
		)

	case RefreshTickMsg:
		// Manual refresh all devices (user-initiated only, no auto-rescheduling)
		// Per-device adaptive refresh (DeviceRefreshMsg) handles automatic polling.
		return c.refreshAllDevices()
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
			cmds = append(cmds, c.fetchDeviceWithID(name, dev))
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
	debug.TraceLock("cache", "RLock", "GetAllDevices")
	c.mu.RLock()
	defer func() {
		c.mu.RUnlock()
		debug.TraceUnlock("cache", "RLock", "GetAllDevices")
	}()

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
	debug.TraceLock("cache", "RLock", "GetOnlineDevices")
	c.mu.RLock()
	defer func() {
		c.mu.RUnlock()
		debug.TraceUnlock("cache", "RLock", "GetOnlineDevices")
	}()

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

// Version returns the current cache version for change detection.
// The version increments on every device data change.
func (c *Cache) Version() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.version
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

// ComponentCountsResult holds all component counts.
type ComponentCountsResult struct {
	SwitchesOn   int
	SwitchesOff  int
	LightsOn     int
	LightsOff    int
	CoversOpen   int
	CoversClosed int
	CoversMoving int
}

// ComponentCounts returns all component counts in a single lock acquisition.
// This is more efficient than calling SwitchCounts, LightCounts, CoverCounts separately.
func (c *Cache) ComponentCounts() ComponentCountsResult {
	debug.TraceLock("cache", "RLock", "ComponentCounts")
	c.mu.RLock()
	defer func() {
		c.mu.RUnlock()
		debug.TraceUnlock("cache", "RLock", "ComponentCounts")
	}()

	var result ComponentCountsResult
	for _, data := range c.devices {
		if !data.Online {
			continue
		}
		for _, sw := range data.Switches {
			if sw.On {
				result.SwitchesOn++
			} else {
				result.SwitchesOff++
			}
		}
		for _, lt := range data.Lights {
			if lt.On {
				result.LightsOn++
			} else {
				result.LightsOff++
			}
		}
		for _, cv := range data.Covers {
			switch cv.State {
			case "open":
				result.CoversOpen++
			case "closed":
				result.CoversClosed++
			default: // opening, closing, stopped
				result.CoversMoving++
			}
		}
	}
	return result
}

// SwitchCounts returns the count of switches that are on and off across all online devices.
func (c *Cache) SwitchCounts() (on, off int) {
	counts := c.ComponentCounts()
	return counts.SwitchesOn, counts.SwitchesOff
}

// LightCounts returns the count of lights that are on and off across all online devices.
func (c *Cache) LightCounts() (on, off int) {
	counts := c.ComponentCounts()
	return counts.LightsOn, counts.LightsOff
}

// CoverCounts returns the count of covers by state across all online devices:
// open, closed, moving (opening/closing/stopped).
func (c *Cache) CoverCounts() (open, closed, moving int) {
	counts := c.ComponentCounts()
	return counts.CoversOpen, counts.CoversClosed, counts.CoversMoving
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
		macToIP:            make(map[string]string),
		wsConnected:        make(map[string]bool),
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

	// Update MAC-to-IP mapping if device has a MAC
	if mac := device.NormalizedMAC(); mac != "" && device.Address != "" {
		c.macToIP[mac] = device.Address
	}
}

// GetIPByMAC returns the cached IP address for a device with the given MAC address.
// Returns empty string if the MAC is not in the cache or invalid.
// This enables instant IP remapping without network discovery for TUI operations.
func (c *Cache) GetIPByMAC(mac string) string {
	if mac == "" {
		return ""
	}
	normalized := model.NormalizeMAC(mac)
	if normalized == "" {
		return ""
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.macToIP[normalized]
}

// UpdateMACMapping updates the MAC-to-IP mapping for a device.
// Called when a device's IP changes (e.g., DHCP reassignment detected via discovery).
func (c *Cache) UpdateMACMapping(mac, ip string) {
	if mac == "" || ip == "" {
		return
	}
	normalized := model.NormalizeMAC(mac)
	if normalized == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.macToIP[normalized] = ip
}

// SetWebSocketConnected updates the WebSocket connection state for a device.
// When a device has an active WebSocket connection, HTTP polling is skipped
// because real-time updates are received via WebSocket.
func (c *Cache) SetWebSocketConnected(name string, connected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if connected {
		c.wsConnected[name] = true
		debug.TraceEvent("cache: %s WebSocket connected, HTTP polling will be skipped", name)
	} else {
		delete(c.wsConnected, name)
		debug.TraceEvent("cache: %s WebSocket disconnected, HTTP polling resumed", name)
	}
}

// IsWebSocketConnected returns whether a device has an active WebSocket connection.
// Devices with WebSocket connections receive real-time updates and don't need HTTP polling.
func (c *Cache) IsWebSocketConnected(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.wsConnected[name]
}
