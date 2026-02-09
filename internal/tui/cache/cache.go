// Package cache provides a shared device data cache for the TUI.
// All views share this cache to avoid redundant network requests.
package cache

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"sort"
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
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
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
	Switches    []SwitchState
	Lights      []LightState
	Covers      []CoverState
	Inputs      []InputState
	RGBs        []RGBState
	Thermostats []ThermostatState

	// Per-component power tracking for accurate aggregation
	SwitchPowers map[int]float64
	LightPowers  map[int]float64 // Light component powers
	CoverPowers  map[int]float64
	PMPowers     map[int]float64 // Dedicated PM component powers
	EMPowers     map[int]float64 // Energy meter powers
	EM1Powers    map[int]float64 // Single-phase EM powers

	UpdatedAt time.Time

	// Link-derived state (populated when device is offline and has a parent link)
	LinkedTo  *config.Link // Non-nil if this device has a parent link configured
	LinkState string       // Derived state from parent switch (e.g., "Off (via bedroom-2pm:0)")

	// NeedsRefresh is set when WebSocket reports a state change without power data.
	// The cache handler should trigger an HTTP refresh to get accurate power readings.
	NeedsRefresh bool

	// lastRequestID tracks the most recent request for stale response handling.
	// Responses with older request IDs are discarded.
	lastRequestID uint64
}

// SwitchState holds the state of a switch component.
type SwitchState struct {
	ID     int
	Name   string
	On     bool
	Source string
}

// LightState holds the state of a light component.
type LightState struct {
	ID   int
	Name string
	On   bool
}

// CoverState holds the state of a cover component.
type CoverState struct {
	ID    int
	Name  string
	State string // "open", "closed", "opening", "closing", "stopped"
}

// InputState holds the state of an input component.
// Input components are physical button/toggle terminals that sense HIGH/LOW states.
type InputState struct {
	ID    int
	State bool   // HIGH (true) / LOW (false)
	Type  string // "button", "switch", "analog"
	Name  string // User-configured name
}

// RGBState holds the state of an RGB/RGBW component.
type RGBState struct {
	ID         int
	Name       string
	Output     bool
	Brightness int // 0-100
	Red        int // 0-255
	Green      int // 0-255
	Blue       int // 0-255
	White      int // 0-255 (for RGBW devices)
	Power      float64
	Source     string
}

// ThermostatState holds the state of a thermostat component.
type ThermostatState struct {
	ID              int
	Name            string
	Enabled         bool
	Mode            string  // "heat", "cool", "auto", "off"
	TargetC         float64 // Target temperature in Celsius
	CurrentC        float64 // Current temperature in Celsius
	CurrentHumidity float64 // Current humidity percentage
	ValvePosition   int     // Valve position 0-100%
	BoostActive     bool
	BoostRemaining  int // Seconds remaining in boost
	OverrideActive  bool
	Source          string
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

// WaveConfig holds wave loading configuration for initial device loading.
// Wave loading prevents overwhelming devices during TUI startup by loading
// devices in small batches with delays between waves.
type WaveConfig struct {
	FirstWaveSize      int           // Devices in first wave (default: 2)
	SubsequentWaveSize int           // Devices in subsequent waves (default: 1)
	WaveDelay          time.Duration // Delay between waves (default: 500ms)
	InitialLoadDelay   time.Duration // Longer delay for initial load (default: 750ms)
}

// DefaultWaveConfig returns sensible defaults for wave loading.
// Conservative defaults to avoid overwhelming devices during startup.
func DefaultWaveConfig() WaveConfig {
	return WaveConfig{
		FirstWaveSize:      2,                      // Small first wave for quick UI feedback
		SubsequentWaveSize: 1,                      // Very conservative - one device at a time
		WaveDelay:          500 * time.Millisecond, // Standard delay between waves
		InitialLoadDelay:   750 * time.Millisecond, // Longer delay during initial load
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
	waveConfig    WaveConfig    // Wave loading configuration

	// FileCache for persisting static device data (model, gen, MAC, components)
	// This data never changes, so we cache it with 24h TTL to avoid redundant HTTP calls
	fileCache *cache.FileCache

	// Track fetch progress
	pendingCount int
	initialLoad  bool

	// initialLoadComplete tracks when all devices have been fetched at least once.
	// This is separate from initialLoad - initialLoad tracks whether we're in the initial
	// loading phase, while initialLoadComplete tracks whether that phase has finished.
	// Used to determine when to start per-device HTTP polling (only after initial load
	// AND for devices without WebSocket).
	initialLoadComplete bool

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

	// EventStream connection tracking - devices managed by EventStream don't need cache HTTP polling
	// Gen2+ devices receive real-time updates via WebSocket
	// Gen1 devices are polled by EventStream (10s interval), so cache polling is redundant
	esManaged map[string]bool

	// pendingRefreshes tracks devices needing HTTP refresh after WebSocket state change without power data.
	// This handles BUG-009 (WebSocket may not include power) and BUG-014 (no refresh after state change).
	// The timestamp is used for debouncing - we wait 500ms before triggering refresh.
	pendingRefreshes map[string]time.Time
}

// New creates a new shared cache with default refresh configuration.
func New(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, fc *cache.FileCache) *Cache {
	return &Cache{
		ctx:                ctx,
		svc:                svc,
		ios:                ios,
		refreshConfig:      DefaultRefreshConfig(),
		waveConfig:         DefaultWaveConfig(),
		fileCache:          fc,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		esManaged:          make(map[string]bool),
		pendingRefreshes:   make(map[string]time.Time),
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
		waveConfig:         DefaultWaveConfig(),
		fileCache:          fc,
		devices:            make(map[string]*DeviceData),
		deviceRefreshTimes: make(map[string]time.Time),
		macToIP:            make(map[string]string),
		esManaged:          make(map[string]bool),
		pendingRefreshes:   make(map[string]time.Time),
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
// Tracks EventStream connection state to skip cache HTTP polling for managed devices.
// Devices managed by EventStream (WebSocket OR polling) don't need cache polling.
func (c *Cache) SubscribeToEvents(es *automation.EventStream) {
	c.eventStream = es

	// Subscribe to events with connection state tracking
	es.Subscribe(func(evt events.Event) {
		// Track connection state changes
		deviceID := evt.DeviceID()
		switch evt.(type) {
		case *events.DeviceOnlineEvent:
			// Check if this device is managed by EventStream (WebSocket or polling)
			info := es.GetConnectionInfo(deviceID)
			isManaged := info.Type == automation.ConnectionWebSocket || info.Type == automation.ConnectionPolling
			c.SetEventStreamManaged(deviceID, isManaged)
		case *events.DeviceOfflineEvent:
			// Device went offline - no longer managed by EventStream
			c.SetEventStreamManaged(deviceID, false)
		}

		// Process the event normally
		c.handleDeviceEvent(evt)
	})

	// Initialize connection state for all currently connected devices
	for name, info := range es.GetAllConnectionInfo() {
		isManaged := info.Type == automation.ConnectionWebSocket || info.Type == automation.ConnectionPolling
		c.SetEventStreamManaged(name, isManaged)
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

	// Only increment version when the event type actually modifies cache data.
	// Unhandled event types (like NotifyEvent) should NOT increment version,
	// otherwise high-frequency events (e.g., ble.scan_result) cause unnecessary re-renders.
	var handled bool

	switch e := evt.(type) {
	case *events.StatusChangeEvent:
		c.handleStatusChange(data, e)
		handled = true
	case *events.FullStatusEvent:
		c.handleFullStatus(data, e)
		handled = true
	case *events.DeviceOnlineEvent:
		data.Online = true
		data.UpdatedAt = evt.Timestamp()
		handled = true
	case *events.DeviceOfflineEvent:
		data.Online = false
		data.UpdatedAt = evt.Timestamp()
		handled = true
	}

	// Only process refresh and increment version if event was actually handled
	if !handled {
		return
	}

	// Check if WebSocket update needs HTTP refresh (BUG-009/014)
	// This handles cases where state changed but power data wasn't included
	// Only for NON-EventStream devices - EventStream devices get updates via WebSocket/polling
	if data.NeedsRefresh {
		data.NeedsRefresh = false
		if !c.esManaged[deviceID] {
			c.pendingRefreshes[deviceID] = time.Now()
			debug.TraceEvent("cache: %s needs HTTP refresh (state change without power)", deviceID)
		}
	}

	c.version++

	c.ios.DebugCat(iostreams.CategoryDevice, "cache: event update for %s, type=%s", deviceID, evt.Type())
}

// handleStatusChange processes incremental status updates from WebSocket.
func (c *Cache) handleStatusChange(data *DeviceData, evt *events.StatusChangeEvent) {
	debug.TraceEvent("cache: StatusChangeEvent for %s, component=%s", data.Device.Name, evt.Component)
	data.Online = true
	data.UpdatedAt = evt.Timestamp()

	if evt.Status == nil {
		return
	}

	componentType, componentID := ParseComponentName(evt.Component)
	ApplyIncrementalUpdate(data, componentType, componentID, evt.Status)
}

// handleFullStatus processes full status events from WebSocket.
func (c *Cache) handleFullStatus(data *DeviceData, evt *events.FullStatusEvent) {
	data.Online = true
	data.UpdatedAt = evt.Timestamp()

	if evt.Status == nil {
		return
	}

	var statusMap map[string]json.RawMessage
	if err := json.Unmarshal(evt.Status, &statusMap); err != nil {
		c.ios.DebugErr("cache: parse full status event", err)
		return
	}

	parsed, err := ParseFullStatus(data.Device.Name, statusMap)
	if err != nil {
		c.ios.DebugErr("cache: parse full status", err)
		return
	}

	ApplyParsedStatus(data, parsed)
	data.Fetched = true
}

// scheduleDeviceRefresh schedules refresh for a single device based on its state.
// Adds random jitter (0-50% of interval) to prevent all devices refreshing simultaneously.
// Skips scheduling for devices managed by EventStream (WebSocket for Gen2+, polling for Gen1).
// Also skips scheduling during initial load - let wave loading complete first.
func (c *Cache) scheduleDeviceRefresh(name string, data *DeviceData) tea.Cmd {
	// Don't start per-device polling until initial load is complete
	// This allows wave loading to finish without interference
	c.mu.RLock()
	loadComplete := c.initialLoadComplete
	c.mu.RUnlock()

	if !loadComplete {
		debug.TraceEvent("scheduleDeviceRefresh: skipping %s (initial load not complete)", name)
		return nil
	}

	// Skip scheduling HTTP refresh if device is managed by EventStream
	// EventStream provides updates via WebSocket (Gen2+) or its own polling (Gen1)
	if c.IsEventStreamManaged(name) {
		debug.TraceEvent("scheduleDeviceRefresh: skipping %s (EventStream managed)", name)
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
		// Use Device.DisplayName() as key for consistent lookup (not config key)
		c.order = make([]string, 0, len(deviceMap))
		for _, dev := range deviceMap {
			displayName := dev.DisplayName()
			c.devices[displayName] = &DeviceData{
				Device:  dev,
				Fetched: false,
			}
			c.order = append(c.order, displayName)
		}
		// Sort for consistent display
		sortStrings(c.order)
		c.pendingCount = len(deviceMap)
		c.version++ // Increment version for cache invalidation
		c.mu.Unlock()

		// Create waves (Gen2 first for resilience)
		waves := c.createWaves(deviceMap)
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
// Wave sizing uses WaveConfig: default is 2 devices first wave, then 1 per wave.
func (c *Cache) createWaves(devices map[string]model.Device) [][]deviceFetch {
	if len(devices) == 0 {
		return nil
	}

	// Convert to slice for sorting
	// Use Device.DisplayName() for consistent lookup (not config key)
	fetches := make([]deviceFetch, 0, len(devices))
	for _, dev := range devices {
		fetches = append(fetches, deviceFetch{Name: dev.DisplayName(), Device: dev})
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

	// Use configurable wave sizes
	firstWaveSize := c.waveConfig.FirstWaveSize
	subsequentWaveSize := c.waveConfig.SubsequentWaveSize

	// Handle edge case where config might be 0 or less
	if firstWaveSize <= 0 {
		firstWaveSize = 2
	}
	if subsequentWaveSize <= 0 {
		subsequentWaveSize = 1
	}

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
		// Use longer delay during initial load to be more gentle on devices
		c.mu.RLock()
		isInitialLoad := c.initialLoad
		c.mu.RUnlock()

		delay := c.waveConfig.WaveDelay
		if isInitialLoad {
			delay = c.waveConfig.InitialLoadDelay
		}

		cmds = append(cmds, tea.Tick(delay, func(time.Time) tea.Msg {
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
// Uses a single GetFullStatus call instead of multiple component-specific calls.
func (c *Cache) fetchDeviceWithID(name string, device model.Device) tea.Cmd {
	requestID := atomic.AddUint64(&c.requestCounter, 1)

	return func() tea.Msg {
		data := &DeviceData{
			Device:        device,
			Fetched:       true,
			UpdatedAt:     time.Now(),
			lastRequestID: requestID,
		}

		// Plugin-managed devices use plugin hooks instead of Shelly API
		if device.IsPluginManaged() {
			return c.fetchPluginDevice(name, device, data, requestID)
		}

		// Mark context as polling - polling failures shouldn't trip circuit breaker (BUG-015)
		ctx, cancel := context.WithTimeout(ratelimit.MarkAsPolling(c.ctx), c.deviceTimeout(device.Generation))
		defer cancel()

		// Try to get device info from FileCache first (it never changes, 24h TTL)
		info, _ := c.tryGetCachedDeviceInfo(name)

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

		// Populate device info (applies cached data to device model + persists to config)
		c.populateDeviceInfo(name, data, info)

		// Skip HTTP status fetch if WebSocket already provided status (Gen2+ only).
		// WebSocket connections fetch Shelly.GetStatus on connect and publish FullStatusEvent.
		c.mu.RLock()
		existingData := c.devices[name]
		esManaged := c.esManaged[name]
		hasRecentStatus := existingData != nil && existingData.Fetched &&
			time.Since(existingData.UpdatedAt) < 5*time.Second
		c.mu.RUnlock()

		if esManaged && hasRecentStatus {
			debug.TraceEvent("cache: skipping HTTP status fetch for %s (WebSocket provided status)", name)
			// Use existing data from WebSocket, just update device info
			c.mu.Lock()
			existingData.Info = info
			existingData.Device = device
			c.mu.Unlock()
			return DeviceUpdateMsg{Name: name, Data: existingData, RequestID: requestID}
		}

		// Single HTTP call to get all status (switches, lights, covers, power, WiFi, Sys)
		statusMap, err := c.svc.GetFullStatusAuto(ctx, name)
		if err != nil {
			// Status fetch failed - device is unreachable, mark as offline
			data.Error = err
			data.Online = false
			c.ios.DebugErr("cache: get full status "+name, err)
			return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
		}

		// Device responded successfully - mark as online
		data.Online = true

		// Parse using unified parser based on generation
		var parsed *ParsedStatus
		var parseErr error
		if info.Generation == 1 {
			parsed, parseErr = ParseGen1Status(name, statusMap)
		} else {
			parsed, parseErr = ParseFullStatus(name, statusMap)
		}

		if parseErr != nil {
			c.ios.DebugErr("cache: parse full status "+name, parseErr)
			// Continue with partial data rather than failing
		}

		if parsed != nil {
			ApplyParsedStatus(data, parsed)

			debug.TraceEvent("cache: %s parsed: %d switches, %d lights, %d covers, power=%.1fW",
				name, len(data.Switches), len(data.Lights), len(data.Covers), data.Power)
		}

		// Fetch config to get component names (separate call, non-blocking failure)
		configMap, err := c.svc.GetFullConfigAuto(ctx, name)
		if err != nil {
			c.ios.DebugErr("cache: get full config "+name, err)
		} else {
			var names *ComponentNames
			if info.Generation == 1 {
				names = ParseGen1Config(configMap)
			} else {
				names = ParseFullConfig(configMap)
			}
			ApplyConfigNames(data, names)
		}

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

// fetchPluginDevice fetches status for a plugin-managed device via plugin hooks.
func (c *Cache) fetchPluginDevice(name string, device model.Device, data *DeviceData, requestID uint64) tea.Msg {
	ctx, cancel := context.WithTimeout(ratelimit.MarkAsPolling(c.ctx), 15*time.Second)
	defer cancel()

	// Build minimal DeviceInfo from config (plugins don't have Shelly DeviceInfo)
	data.Info = BuildPluginDeviceInfo(name, device)

	// Get status via plugin hook through the service layer
	result, err := c.svc.GetPluginDeviceStatus(ctx, device)
	if err != nil {
		data.Error = err
		data.Online = false
		c.ios.DebugErr("cache: plugin status "+name, err)
		return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
	}

	data.Online = result.Online

	// Parse plugin status into our unified format
	parsed := ParsePluginStatus(name, result)
	if parsed != nil {
		ApplyParsedStatus(data, parsed)
		debug.TraceEvent("cache: plugin %s parsed: %d switches, %d lights, %d covers, power=%.1fW",
			name, len(data.Switches), len(data.Lights), len(data.Covers), data.Power)
	}

	return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
}

// populateDeviceInfo reconciles device info between live data and stored config.
// Uses MAC as the source of truth for device identity. Updates config if data is
// missing or incorrect.
func (c *Cache) populateDeviceInfo(name string, data *DeviceData, info *shelly.DeviceInfo) {
	updates := c.reconcileDeviceInfo(data, info)
	if updates.Type != "" || updates.Model != "" || updates.Generation > 0 || updates.MAC != "" {
		if err := config.UpdateDeviceInfo(name, updates); err != nil {
			c.ios.DebugErr("persist device info", err)
		}
	}
}

// reconcileDeviceInfo compares live device info against stored config and returns updates.
// Uses MAC as the source of truth for device identity.
func (c *Cache) reconcileDeviceInfo(data *DeviceData, info *shelly.DeviceInfo) config.DeviceUpdates {
	var updates config.DeviceUpdates

	c.reconcileMAC(data, info, &updates)
	c.reconcileType(data, info, &updates)
	c.reconcileModel(data, info, &updates)
	c.reconcileGeneration(data, info, &updates)

	return updates
}

// reconcileMAC handles MAC address reconciliation and device replacement detection.
func (c *Cache) reconcileMAC(data *DeviceData, info *shelly.DeviceInfo, updates *config.DeviceUpdates) {
	if info.MAC == "" {
		return
	}

	liveMAC := model.NormalizeMAC(info.MAC)
	storedMAC := data.Device.NormalizedMAC()

	if storedMAC == "" {
		// MAC was missing, fill it in
		data.Device.MAC = liveMAC
		updates.MAC = info.MAC
		debug.TraceEvent("cache: reconcile %s - added missing MAC: %s", data.Device.Name, info.MAC)
		return
	}

	if liveMAC != storedMAC {
		// Device was replaced - warn user but continue with update
		c.ios.DebugCat(iostreams.CategoryDevice,
			"WARNING: Device %q appears to have been replaced (stored MAC: %s, live MAC: %s)",
			data.Device.Name, storedMAC, liveMAC)
		debug.TraceEvent("cache: device %s replaced - MAC changed from %s to %s",
			data.Device.Name, storedMAC, liveMAC)
		data.Device.MAC = liveMAC
		updates.MAC = info.MAC
	}
}

// reconcileType handles device type/SKU reconciliation.
func (c *Cache) reconcileType(data *DeviceData, info *shelly.DeviceInfo, updates *config.DeviceUpdates) {
	if info.Model == "" {
		return
	}

	if data.Device.Type == "" {
		data.Device.Type = info.Model
		updates.Type = info.Model
		debug.TraceEvent("cache: reconcile %s - added missing Type: %s", data.Device.Name, info.Model)
	} else if data.Device.Type != info.Model {
		debug.TraceEvent("cache: reconcile %s - corrected Type: %s -> %s",
			data.Device.Name, data.Device.Type, info.Model)
		data.Device.Type = info.Model
		updates.Type = info.Model
	}
}

// reconcileModel handles display name reconciliation.
func (c *Cache) reconcileModel(data *DeviceData, info *shelly.DeviceInfo, updates *config.DeviceUpdates) {
	if info.Model == "" {
		return
	}

	displayName := types.ModelDisplayName(info.Model)

	if data.Device.Model == "" || data.Device.Model == data.Device.Type {
		data.Device.Model = displayName
		updates.Model = displayName
		debug.TraceEvent("cache: reconcile %s - added missing Model: %s", data.Device.Name, displayName)
	} else if data.Device.Model != displayName {
		debug.TraceEvent("cache: reconcile %s - corrected Model: %s -> %s",
			data.Device.Name, data.Device.Model, displayName)
		data.Device.Model = displayName
		updates.Model = displayName
	}
}

// reconcileGeneration handles generation reconciliation.
func (c *Cache) reconcileGeneration(data *DeviceData, info *shelly.DeviceInfo, updates *config.DeviceUpdates) {
	if info.Generation <= 0 {
		return
	}

	if data.Device.Generation == 0 {
		data.Device.Generation = info.Generation
		updates.Generation = info.Generation
		debug.TraceEvent("cache: reconcile %s - added missing Generation: %d", data.Device.Name, info.Generation)
	} else if data.Device.Generation != info.Generation {
		debug.TraceEvent("cache: reconcile %s - corrected Generation: %d -> %d",
			data.Device.Name, data.Device.Generation, info.Generation)
		data.Device.Generation = info.Generation
		updates.Generation = info.Generation
	}
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

// FetchExtendedStatusMsg is sent when extended status fetch completes.
type FetchExtendedStatusMsg struct {
	Name string
	WiFi *shelly.WiFiStatus
	Sys  *shelly.SysStatus
}

// FetchExtendedStatus fetches WiFi and Sys status for a device on-demand.
// This is called lazily when a device is focused in the device info panel.
// If data is already present (from GetFullStatus parsing), it returns immediately.
// Only fetches what's missing, in parallel to avoid timeout issues.
func (c *Cache) FetchExtendedStatus(name string) tea.Cmd {
	debug.TraceEvent("cache: FetchExtendedStatus called for %s", name)

	// Check if we already have this data from GetFullStatus
	c.mu.RLock()
	data := c.devices[name]
	var hasWiFi, hasSys bool
	if data != nil {
		hasWiFi = data.WiFi != nil
		hasSys = data.Sys != nil
	}
	c.mu.RUnlock()

	// If we already have both, return immediately
	if hasWiFi && hasSys {
		debug.TraceEvent("cache: FetchExtendedStatus skipping %s (already have WiFi and Sys)", name)
		return func() tea.Msg {
			return FetchExtendedStatusMsg{Name: name}
		}
	}

	return func() tea.Msg {
		msg := FetchExtendedStatusMsg{Name: name}

		// Fetch only what's missing, in parallel with independent timeouts
		// Mark as polling - these are background requests that shouldn't trip circuit breaker (BUG-015)
		var wg sync.WaitGroup
		var wifiMu, sysMu sync.Mutex

		if !hasWiFi {
			wg.Go(func() {
				ctx, cancel := context.WithTimeout(ratelimit.MarkAsPolling(c.ctx), 15*time.Second)
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
		}

		if !hasSys {
			wg.Go(func() {
				ctx, cancel := context.WithTimeout(ratelimit.MarkAsPolling(c.ctx), 15*time.Second)
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
		}

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
		c.initialLoadComplete = true // Mark that initial load phase has finished
		debug.TraceEvent("cache: initial load complete")
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
		c.resolveLinkState(msg.Name, existing)
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

	// Resolve link state for offline devices
	c.resolveLinkState(msg.Name, msg.Data)

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

// resolveLinkState populates link-derived state for offline devices.
// Must be called with c.mu held.
func (c *Cache) resolveLinkState(name string, data *DeviceData) {
	// Look up link for this device
	link, ok := config.GetLink(name)
	if !ok {
		data.LinkedTo = nil
		data.LinkState = ""
		return
	}

	data.LinkedTo = &link

	// If device is online, don't override with link state
	if data.Online {
		data.LinkState = ""
		return
	}

	// Try to derive state from cached parent data
	parentData, parentExists := c.devices[link.ParentDevice]
	if !parentExists || parentData == nil || !parentData.Online {
		data.LinkState = "Unknown"
		return
	}

	// Find the switch state from parent's cached data
	for _, sw := range parentData.Switches {
		if sw.ID == link.SwitchID {
			if sw.On {
				data.LinkState = "On"
			} else {
				data.LinkState = "Off (switch off)"
			}
			return
		}
	}

	data.LinkState = "Unknown"
}

// RefreshDebounceInterval is the delay before triggering HTTP refresh for devices
// that had state changes without power data in WebSocket updates.
const RefreshDebounceInterval = 500 * time.Millisecond

// processPendingRefreshes checks for devices needing HTTP refresh and returns commands.
// This handles BUG-009/014 where WebSocket state changes don't include power data.
func (c *Cache) processPendingRefreshes() tea.Cmd {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.pendingRefreshes) == 0 {
		return nil
	}

	now := time.Now()
	cmds := make([]tea.Cmd, 0, len(c.pendingRefreshes))

	for name, requestTime := range c.pendingRefreshes {
		// Wait for debounce interval before triggering refresh
		if now.Sub(requestTime) < RefreshDebounceInterval {
			continue
		}

		delete(c.pendingRefreshes, name)

		// Skip HTTP refresh for EventStream-managed devices - they get updates via EventStream
		if c.esManaged[name] {
			debug.TraceEvent("cache: skipping HTTP refresh for %s (EventStream managed)", name)
			continue
		}

		// Get device data for the refresh command
		data := c.devices[name]
		if data == nil {
			continue
		}

		device := data.Device // Copy while holding lock
		debug.TraceEvent("cache: triggering HTTP refresh for %s (debounced)", name)
		cmds = append(cmds, c.fetchDeviceWithIDUnlocked(name, device))
	}

	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// fetchDeviceWithIDUnlocked is like fetchDeviceWithID but assumes the lock is NOT held.
// Used by processPendingRefreshes which needs to release the lock before returning.
func (c *Cache) fetchDeviceWithIDUnlocked(name string, device model.Device) tea.Cmd {
	return c.fetchDeviceWithID(name, device)
}

// handleFocusDebounce processes a debounced focus request for a device.
// Returns a command to fetch the device if still focused and not WebSocket-connected.
func (c *Cache) handleFocusDebounce(msg FocusDebounceMsg) tea.Cmd {
	// Debounced fetch for focused device - only trigger if still focused
	// Copy all needed values while holding lock to avoid race conditions
	c.mu.RLock()
	currentFocus := c.focusedDevice
	loadComplete := c.initialLoadComplete
	data := c.devices[msg.Name]
	esManaged := c.esManaged[msg.Name]
	var hasWiFi, hasSys bool
	var device model.Device
	if data != nil {
		hasWiFi = data.WiFi != nil
		hasSys = data.Sys != nil
		device = data.Device // Copy struct while holding lock
	}
	c.mu.RUnlock()

	// Skip during initial load - wave loading will fetch all device data
	// This prevents overwhelming devices with concurrent HTTP requests
	if !loadComplete {
		debug.TraceEvent("FocusDebounceMsg: skipping %s (initial load not complete)", msg.Name)
		return nil
	}

	// Only fetch if this device is still focused (user stopped scrolling)
	if msg.Name != currentFocus || data == nil {
		return nil
	}

	deviceName := msg.Name // Capture for closure

	// EventStream provides updates - skip HTTP fetch for device status
	// But still fetch extended status (WiFi/Sys) if not yet loaded
	if esManaged {
		debug.TraceEvent("FocusDebounceMsg: skipping HTTP for %s (EventStream managed)", msg.Name)
		// Only fetch extended status if missing
		if !hasWiFi || !hasSys {
			return func() tea.Msg { return ExtendedStatusDebounceMsg{Name: deviceName} }
		}
		return nil
	}

	debug.TraceEvent("FocusDebounceMsg: triggering fetch for %s", msg.Name)
	// Not EventStream-managed - need HTTP fetch for device status and extended status
	return tea.Batch(
		c.fetchDeviceWithID(deviceName, device),
		func() tea.Msg { return ExtendedStatusDebounceMsg{Name: deviceName} },
	)
}

// Update handles cache-related messages.
func (c *Cache) Update(msg tea.Msg) tea.Cmd {
	// Process any pending HTTP refreshes for devices with state changes without power data
	pendingCmd := c.processPendingRefreshes()

	// Helper to combine pending refresh commands with message command
	combineWithPending := func(cmd tea.Cmd) tea.Cmd {
		if pendingCmd == nil {
			return cmd
		}
		if cmd == nil {
			return pendingCmd
		}
		return tea.Batch(pendingCmd, cmd)
	}

	switch msg := msg.(type) {
	case WaveMsg:
		c.mu.Lock()
		c.currentWave = msg.Wave
		c.mu.Unlock()
		return combineWithPending(c.processWave(msg))

	case DeviceUpdateMsg:
		return combineWithPending(c.handleDeviceUpdate(msg))

	case DeviceRefreshMsg:
		// DeviceRefreshMsg is an explicit refresh request (from user action, keybinding, etc.)
		// Always honor it - esManaged only blocks automatic/scheduled polling, not explicit requests
		// Refresh a single device - copy Device while holding lock to avoid race
		c.mu.RLock()
		data := c.devices[msg.Name]
		var device model.Device
		if data != nil {
			device = data.Device // Copy struct while holding lock
		}
		c.mu.RUnlock()

		if data == nil {
			return combineWithPending(nil)
		}
		return combineWithPending(c.fetchDeviceWithID(msg.Name, device))

	case FocusDebounceMsg:
		return combineWithPending(c.handleFocusDebounce(msg))

	case RefreshTickMsg:
		// Manual refresh all devices (user-initiated only, no auto-rescheduling)
		// Per-device adaptive refresh (DeviceRefreshMsg) handles automatic polling.
		return combineWithPending(c.refreshAllDevices())
	}
	return combineWithPending(nil)
}

// refreshAllDevices refreshes status for all devices.
// Skips EventStream-managed devices since they receive updates via EventStream.
func (c *Cache) refreshAllDevices() tea.Cmd {
	return func() tea.Msg {
		c.mu.RLock()
		devicesCopy := make(map[string]model.Device, len(c.devices))
		for name, data := range c.devices {
			// Skip EventStream-managed devices - they get updates via EventStream
			if c.esManaged[name] {
				continue
			}
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
// Counts components from ALL devices - offline devices' components are counted as "off".
// For devices without parsed components, infers counts from device model.
func (c *Cache) ComponentCounts() ComponentCountsResult {
	debug.TraceLock("cache", "RLock", "ComponentCounts")
	c.mu.RLock()
	defer func() {
		c.mu.RUnlock()
		debug.TraceUnlock("cache", "RLock", "ComponentCounts")
	}()

	var result ComponentCountsResult
	for _, data := range c.devices {
		countDeviceComponents(data, &result)
	}
	return result
}

// countDeviceComponents adds the component counts for a single device to the result.
func countDeviceComponents(data *DeviceData, result *ComponentCountsResult) {
	hasParsedComponents := len(data.Switches) > 0 || len(data.Lights) > 0 || len(data.Covers) > 0

	if hasParsedComponents {
		countParsedComponents(data, result)
		return
	}
	// No parsed components - infer from device model/type
	countInferredComponents(data, result)
}

// countParsedComponents counts components from parsed device data.
func countParsedComponents(data *DeviceData, result *ComponentCountsResult) {
	// Count switches - offline devices' switches count as off
	for _, sw := range data.Switches {
		if data.Online && sw.On {
			result.SwitchesOn++
		} else {
			result.SwitchesOff++
		}
	}
	// Count lights - offline devices' lights count as off
	for _, lt := range data.Lights {
		if data.Online && lt.On {
			result.LightsOn++
		} else {
			result.LightsOff++
		}
	}
	// Count covers - offline devices' covers count as closed
	for _, cv := range data.Covers {
		countCover(data.Online, cv.State, result)
	}
}

// countCover counts a single cover based on online status and state.
func countCover(online bool, state string, result *ComponentCountsResult) {
	if !online {
		result.CoversClosed++
		return
	}
	switch state {
	case "open":
		result.CoversOpen++
	case "closed":
		result.CoversClosed++
	default:
		result.CoversMoving++
	}
}

// countInferredComponents infers component counts from device model/type.
func countInferredComponents(data *DeviceData, result *ComponentCountsResult) {
	modelStr := data.Device.Type
	if modelStr == "" {
		modelStr = data.Device.Model
	}
	if modelStr == "" {
		return
	}
	caps := DetectComponents(modelStr)
	// Inferred components are always "off" since we don't know actual state
	result.SwitchesOff += caps.NumSwitches
	result.LightsOff += caps.NumLights
	result.CoversClosed += caps.NumCovers
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

// SumPowers calculates the total power from all per-component power maps.
// This is used after WebSocket incremental updates to recalculate the aggregate power
// instead of overwriting it with a single component's value.
func (d *DeviceData) SumPowers() float64 {
	var total float64
	for _, v := range d.SwitchPowers {
		total += v
	}
	for _, v := range d.LightPowers {
		total += v
	}
	for _, v := range d.CoverPowers {
		total += v
	}
	for _, v := range d.PMPowers {
		total += v
	}
	for _, v := range d.EMPowers {
		total += v
	}
	for _, v := range d.EM1Powers {
		total += v
	}
	return total
}

// EnsurePowerMaps initializes all power maps if they are nil.
// This is needed for WebSocket incremental updates which may arrive before HTTP fetch.
func (d *DeviceData) EnsurePowerMaps() {
	if d.SwitchPowers == nil {
		d.SwitchPowers = make(map[int]float64)
	}
	if d.LightPowers == nil {
		d.LightPowers = make(map[int]float64)
	}
	if d.CoverPowers == nil {
		d.CoverPowers = make(map[int]float64)
	}
	if d.PMPowers == nil {
		d.PMPowers = make(map[int]float64)
	}
	if d.EMPowers == nil {
		d.EMPowers = make(map[int]float64)
	}
	if d.EM1Powers == nil {
		d.EM1Powers = make(map[int]float64)
	}
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
		esManaged:          make(map[string]bool),
		pendingRefreshes:   make(map[string]time.Time),
		refreshConfig:      DefaultRefreshConfig(),
		waveConfig:         DefaultWaveConfig(),
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

	c.devices[device.DisplayName()] = data

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

// SetEventStreamManaged updates the EventStream connection state for a device.
// When a device is managed by EventStream (WebSocket for Gen2+ or polling for Gen1),
// cache HTTP polling is skipped because updates are already being pushed/polled.
func (c *Cache) SetEventStreamManaged(name string, managed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if managed {
		c.esManaged[name] = true
		debug.TraceEvent("cache: %s EventStream managed, cache HTTP polling skipped", name)
	} else {
		delete(c.esManaged, name)
		debug.TraceEvent("cache: %s EventStream disconnected, cache HTTP polling resumed", name)
	}
}

// IsEventStreamManaged returns whether a device is managed by EventStream.
// Managed devices receive updates via WebSocket (Gen2+) or are polled by EventStream (Gen1),
// so cache HTTP polling is unnecessary and skipped.
func (c *Cache) IsEventStreamManaged(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.esManaged[name]
}
