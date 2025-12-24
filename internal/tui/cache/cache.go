// Package cache provides a shared device data cache for the TUI.
// All views share this cache to avoid redundant network requests.
package cache

import (
	"context"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"

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
}

// SwitchState holds the state of a switch component.
type SwitchState struct {
	ID     int
	On     bool
	Source string
}

// DeviceUpdateMsg is sent when a single device's data is updated.
type DeviceUpdateMsg struct {
	Name string
	Data *DeviceData
}

// AllDevicesLoadedMsg is sent when all devices have been fetched at least once.
type AllDevicesLoadedMsg struct{}

// RefreshTickMsg triggers periodic refresh.
type RefreshTickMsg struct{}

// Cache holds shared device data for all TUI components.
type Cache struct {
	mu      sync.RWMutex
	devices map[string]*DeviceData
	order   []string // Sorted device names for consistent display

	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	refreshInterval time.Duration

	// Track fetch progress
	pendingCount int
	initialLoad  bool
}

// New creates a new shared cache.
func New(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, refreshInterval time.Duration) *Cache {
	return &Cache{
		ctx:             ctx,
		svc:             svc,
		ios:             ios,
		refreshInterval: refreshInterval,
		devices:         make(map[string]*DeviceData),
		initialLoad:     true,
	}
}

// Init returns the initial command to start fetching devices.
func (c *Cache) Init() tea.Cmd {
	return tea.Batch(
		c.loadDevices(),
		c.scheduleRefresh(),
	)
}

// scheduleRefresh schedules the next refresh tick.
func (c *Cache) scheduleRefresh() tea.Cmd {
	return tea.Tick(c.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// loadDevices loads all devices from config and starts fetching their status.
func (c *Cache) loadDevices() tea.Cmd {
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

		// Return a batch of commands to fetch each device
		cmds := make([]tea.Cmd, 0, len(deviceMap))
		for name, dev := range deviceMap {
			cmds = append(cmds, c.fetchDevice(name, dev))
		}
		return tea.BatchMsg(cmds)
	}
}

// fetchDevice fetches status for a single device.
func (c *Cache) fetchDevice(name string, device model.Device) tea.Cmd {
	return func() tea.Msg {
		data := &DeviceData{
			Device:    device,
			Fetched:   true,
			UpdatedAt: time.Now(),
		}

		// Per-device timeout
		ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
		defer cancel()

		// Get device info first (auto-detects Gen1 vs Gen2)
		info, err := c.svc.DeviceInfoAuto(ctx, device.Address)
		if err != nil {
			data.Error = err
			data.Online = false
			return DeviceUpdateMsg{Name: name, Data: data}
		}

		data.Info = info
		data.Online = true

		// Gen1 devices use different APIs - skip Gen2-specific calls
		if info.Generation == 1 {
			// For Gen1, we're online but don't have switch/monitoring data yet
			// TODO: Add Gen1-specific relay status collection
			return DeviceUpdateMsg{Name: name, Data: data}
		}

		// Gen2+ device - get switch states
		switches, err := c.svc.SwitchList(ctx, device.Address)
		if err == nil {
			for _, sw := range switches {
				data.Switches = append(data.Switches, SwitchState{
					ID: sw.ID,
					On: sw.Output,
				})
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

		return DeviceUpdateMsg{Name: name, Data: data}
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
	case DeviceUpdateMsg:
		c.mu.Lock()
		c.devices[msg.Name] = msg.Data
		c.pendingCount--
		allDone := c.pendingCount <= 0 && c.initialLoad
		if allDone {
			c.initialLoad = false
		}
		c.mu.Unlock()

		if allDone {
			return func() tea.Msg { return AllDevicesLoadedMsg{} }
		}
		return nil

	case RefreshTickMsg:
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
