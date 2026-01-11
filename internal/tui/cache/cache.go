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
	Inputs   []InputState

	// Per-component power tracking for accurate aggregation
	SwitchPowers map[int]float64
	CoverPowers  map[int]float64

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

// InputState holds the state of an input component.
// Input components are physical button/toggle terminals that sense HIGH/LOW states.
type InputState struct {
	ID    int
	State bool   // HIGH (true) / LOW (false)
	Type  string // "button", "switch", "analog"
	Name  string // User-configured name
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
		waveConfig:         DefaultWaveConfig(),
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
		waveConfig:         DefaultWaveConfig(),
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

	switch e := evt.(type) {
	case *events.StatusChangeEvent:
		c.handleStatusChange(data, e)
	case *events.FullStatusEvent:
		c.handleFullStatus(data, e)
	case *events.DeviceOnlineEvent:
		data.Online = true
		data.UpdatedAt = evt.Timestamp()
	case *events.DeviceOfflineEvent:
		data.Online = false
		data.UpdatedAt = evt.Timestamp()
	}
	c.version++

	c.ios.DebugCat(iostreams.CategoryDevice, "cache: event update for %s, type=%s", deviceID, evt.Type())
}

// handleStatusChange processes incremental status updates from WebSocket.
func (c *Cache) handleStatusChange(data *DeviceData, evt *events.StatusChangeEvent) {
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
}

// scheduleDeviceRefresh schedules refresh for a single device based on its state.
// Adds random jitter (0-50% of interval) to prevent all devices refreshing simultaneously.
// Skips scheduling for devices with active WebSocket connections (they receive real-time updates).
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

		ctx, cancel := context.WithTimeout(c.ctx, c.deviceTimeout(device.Generation))
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
		data.Online = true

		// Populate device info (applies cached data to device model + persists to config)
		c.populateDeviceInfo(name, data, info)

		// Single HTTP call to get all status (switches, lights, covers, power, WiFi, Sys)
		statusMap, err := c.svc.GetFullStatusAuto(ctx, name)
		if err != nil {
			// Status fetch failed but device info succeeded - still mark as online with error
			data.Error = err
			c.ios.DebugErr("cache: get full status "+name, err)
			return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
		}

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
		var wg sync.WaitGroup
		var wifiMu, sysMu sync.Mutex

		if !hasWiFi {
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
		}

		if !hasSys {
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

		// Refresh a single device - copy Device while holding lock to avoid race
		c.mu.RLock()
		data := c.devices[msg.Name]
		var device model.Device
		if data != nil {
			device = data.Device // Copy struct while holding lock
		}
		c.mu.RUnlock()

		if data == nil {
			return nil
		}
		return c.fetchDeviceWithID(msg.Name, device)

	case FocusDebounceMsg:
		// Debounced fetch for focused device - only trigger if still focused
		// Copy all needed values while holding lock to avoid race conditions
		c.mu.RLock()
		currentFocus := c.focusedDevice
		data := c.devices[msg.Name]
		wsConnected := c.wsConnected[msg.Name]
		var hasWiFi, hasSys bool
		var device model.Device
		if data != nil {
			hasWiFi = data.WiFi != nil
			hasSys = data.Sys != nil
			device = data.Device // Copy struct while holding lock
		}
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
			if !hasWiFi || !hasSys {
				return func() tea.Msg { return ExtendedStatusDebounceMsg{Name: deviceName} }
			}
			return nil
		}

		debug.TraceEvent("FocusDebounceMsg: triggering fetch for %s", msg.Name)
		// No WebSocket - need HTTP fetch for device status and extended status
		return tea.Batch(
			c.fetchDeviceWithID(deviceName, device),
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
