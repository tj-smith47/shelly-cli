// Package cache provides a shared device data cache for the TUI.
// All views share this cache to avoid redundant network requests.
package cache

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
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

// scheduleRefresh schedules the next refresh tick (deprecated - use scheduleDeviceRefresh).
func (c *Cache) scheduleRefresh() tea.Cmd {
	return tea.Tick(c.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// scheduleDeviceRefresh schedules refresh for a single device based on its state.
func (c *Cache) scheduleDeviceRefresh(name string, data *DeviceData) tea.Cmd {
	interval := c.getRefreshInterval(data)
	return tea.Tick(interval, func(time.Time) tea.Msg {
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
		info, err := c.svc.DeviceInfoAuto(ctx, device.Address)
		if err != nil {
			data.Error = err
			data.Online = false
			return DeviceUpdateMsg{Name: name, Data: data, RequestID: requestID}
		}

		data.Info = info
		data.Online = true

		// Populate device model from DeviceInfo if not already set
		// Use human-readable product name if available (e.g., "Shelly Pro 1PM")
		if data.Device.Model == "" && info.Model != "" {
			data.Device.Model = types.ModelDisplayName(info.Model)
		}
		// Type is the model code/SKU (e.g., "SPSW-001PE16EU") for reference
		if data.Device.Type == "" && info.Model != "" {
			data.Device.Type = info.Model
		}
		// Update generation if not set
		if data.Device.Generation == 0 && info.Generation > 0 {
			data.Device.Generation = info.Generation
		}

		// Get switch states (Gen2+ only - Gen1 uses different relay API)
		if info.Generation > 1 {
			switches, err := c.svc.SwitchList(ctx, device.Address)
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
		snapshot, err := c.svc.GetMonitoringSnapshotAuto(ctx, device.Address)
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

	// If refresh failed but we have existing good data, preserve it
	// Only update the error/online status, keep cached info
	if msg.Data.Error != nil && existing != nil && existing.Fetched && existing.Info != nil {
		// Preserve the existing cached data, just mark offline
		existing.Online = false
		existing.Error = msg.Data.Error
		existing.UpdatedAt = msg.Data.UpdatedAt
		existing.lastRequestID = msg.RequestID
		// Keep existing.Info, Snapshot, Power, etc. - don't clear good data
		c.devices[msg.Name] = existing
	} else {
		// Normal update - new data or first fetch
		c.devices[msg.Name] = msg.Data
	}
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
