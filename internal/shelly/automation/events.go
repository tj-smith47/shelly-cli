// Package automation provides script, schedule, and event automation for Shelly devices.
package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/events"
	"github.com/tj-smith47/shelly-go/gen1"
	"github.com/tj-smith47/shelly-go/notifications"
	"github.com/tj-smith47/shelly-go/rpc"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// EventStream manages WebSocket connections to multiple devices for real-time updates.
type EventStream struct {
	svc         EventStreamProvider
	bus         *events.EventBus
	connections map[string]*deviceConnection
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	pollerWg    sync.WaitGroup

	// CoIoT listener for Gen1 devices (optional, reduces HTTP polling)
	coiotListener    *gen1.CoIoTListener
	coiotDevices     map[string]time.Time // device MAC -> last CoIoT update time
	coiotDevicesMu   sync.RWMutex
	coiotMACToName   map[string]string // MAC -> device name mapping
	coiotMACToNameMu sync.RWMutex
}

type deviceConnection struct {
	name       string
	address    string
	ws         *transport.WebSocket
	cancel     context.CancelFunc
	generation int
	online     bool   // tracks last known online state for Gen1 polling
	coiotID    string // CoIoT device ID for Gen1 devices (e.g., "shelly1-AABBCC")
}

// NewEventStream creates a new event stream manager.
func NewEventStream(svc EventStreamProvider) *EventStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &EventStream{
		svc:            svc,
		bus:            events.NewEventBus(),
		connections:    make(map[string]*deviceConnection),
		ctx:            ctx,
		cancel:         cancel,
		coiotDevices:   make(map[string]time.Time),
		coiotMACToName: make(map[string]string),
	}
}

// Subscribe registers a handler to receive all device events.
func (es *EventStream) Subscribe(handler func(events.Event)) {
	es.bus.Subscribe(handler)
}

// SubscribeFiltered registers a handler with filtering.
func (es *EventStream) SubscribeFiltered(filter events.Filter, handler func(events.Event)) {
	es.bus.SubscribeFiltered(filter, handler)
}

// Start begins streaming events from all registered devices.
// It connects to Gen2+ devices via WebSocket, listens for CoIoT from Gen1 devices,
// and falls back to HTTP polling for Gen1 devices without CoIoT.
func (es *EventStream) Start() error {
	devices := config.ListDevices()
	if len(devices) == 0 {
		return nil
	}

	// Start CoIoT listener for Gen1 devices (multicast UDP)
	es.startCoIoTListener()

	var wg sync.WaitGroup
	for _, d := range devices {
		device := d
		wg.Go(func() {
			es.connectDevice(device.Name, device.Address)
		})
	}
	wg.Wait()

	return nil
}

// connectDevice establishes a WebSocket connection to a device.
func (es *EventStream) connectDevice(name, address string) {
	ctx, cancel := context.WithCancel(es.ctx)

	// Check device generation - use name (not address) to look up in config
	resolvedDevice, err := es.svc.ResolveWithGeneration(ctx, name)
	if err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, fmt.Sprintf("resolve device %s", name), err)
		// Publish offline event
		es.bus.Publish(events.NewDeviceOfflineEvent(name).WithReason(err.Error()))
		cancel()
		return
	}

	// Gen1 devices don't support WebSocket - use polling fallback
	if resolvedDevice.Generation == 1 {
		iostreams.DebugCat(iostreams.CategoryNetwork, "Device %s is Gen1, using polling", name)
		es.pollerWg.Go(func() { es.pollGen1Device(ctx, name, address) })

		es.mu.Lock()
		es.connections[name] = &deviceConnection{
			name:       name,
			address:    address,
			cancel:     cancel,
			generation: 1,
		}
		es.mu.Unlock()
		return
	}

	// Connect via WebSocket for Gen2+ devices
	wsURL := fmt.Sprintf("ws://%s/rpc", address)
	ws := transport.NewWebSocket(wsURL,
		transport.WithReconnect(true),
		transport.WithPingInterval(30*time.Second),
	)

	// Pre-register connection so state callback can find it
	es.mu.Lock()
	es.connections[name] = &deviceConnection{
		name:       name,
		address:    address,
		ws:         ws,
		cancel:     cancel,
		generation: resolvedDevice.Generation,
		online:     false, // Will be set true on successful connect
	}
	es.mu.Unlock()

	// Register state change callback to handle reconnection events
	// This ensures DeviceOnlineEvent is published on both initial connect AND reconnections
	ws.OnStateChange(func(state transport.ConnectionState) {
		es.handleWebSocketStateChange(name, address, state)
	})

	if err := ws.Connect(ctx); err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, fmt.Sprintf("connect websocket %s", name), err)
		es.bus.Publish(events.NewDeviceOfflineEvent(name).WithReason(err.Error()))
		// Clean up pre-registered connection
		es.mu.Lock()
		delete(es.connections, name)
		es.mu.Unlock()
		cancel()
		return
	}

	// Register notification handler
	notifyHandler := func(msg json.RawMessage) {
		debug.TraceEvent("ws: %s received message: %s", name, string(msg))

		if !notifications.IsNotification(msg) {
			debug.TraceEvent("ws: %s not a notification, ignoring", name)
			return
		}

		evts, err := notifications.ParseGen2NotificationJSON(name, msg)
		if err != nil {
			debug.TraceEvent("ws: %s parse error: %v", name, err)
			return
		}

		debug.TraceEvent("ws: %s parsed %d events", name, len(evts))
		for _, evt := range evts {
			debug.TraceEvent("ws: %s publishing event type=%s", name, evt.Type())
			es.bus.Publish(evt)
		}
	}

	if err := ws.Subscribe(notifyHandler); err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, fmt.Sprintf("subscribe %s", name), err)
		closeWS(ws)
		// Clean up pre-registered connection
		es.mu.Lock()
		delete(es.connections, name)
		es.mu.Unlock()
		cancel()
		return
	}

	// Fetch initial status via WebSocket and enable notifications.
	// Per Shelly docs: "To start receiving notifications over websocket,
	// you have to send at least one request frame with a valid source (src)."
	// We use GetStatus instead of GetDeviceInfo to also get initial state,
	// eliminating the need for a separate HTTP request during initial load.
	rb := rpc.NewRequestBuilder()
	req, err := rb.Build("Shelly.GetStatus", nil)
	if err == nil {
		if statusJSON, err := ws.Call(ctx, req); err != nil {
			debug.TraceEvent("ws: %s initial GetStatus failed: %v", name, err)
		} else {
			debug.TraceEvent("ws: %s notifications enabled, publishing initial status", name)
			// Publish as FullStatusEvent so cache can use it
			es.bus.Publish(events.NewFullStatusEvent(name, statusJSON).
				WithSource(events.EventSourceWebSocket))
		}
	}

	// Connection was pre-registered before Connect() to enable state callback
	// DeviceOnlineEvent is published by handleWebSocketStateChange on StateConnected
	iostreams.DebugCat(iostreams.CategoryNetwork, "WebSocket connected for %s, waiting for state callback", name)
}

// handleWebSocketStateChange handles WebSocket connection state changes.
// This ensures DeviceOnlineEvent is published on both initial connect AND reconnections,
// and DeviceOfflineEvent is published when the connection is lost.
func (es *EventStream) handleWebSocketStateChange(name, address string, state transport.ConnectionState) {
	debug.TraceEvent("ws: %s state changed to %s", name, state)

	switch state {
	case transport.StateConnected:
		es.mu.Lock()
		conn := es.connections[name]
		if conn == nil {
			// Connection was removed (e.g., during cleanup)
			es.mu.Unlock()
			return
		}
		wasOnline := conn.online
		conn.online = true
		es.mu.Unlock()

		// Publish online event for both initial connection and reconnection
		es.bus.Publish(events.NewDeviceOnlineEvent(name).WithAddress(address))
		if wasOnline {
			// This is a reconnection
			iostreams.DebugCat(iostreams.CategoryNetwork, "Reconnected to %s via WebSocket", name)
		} else {
			// Initial connection
			iostreams.DebugCat(iostreams.CategoryNetwork, "Connected to %s via WebSocket", name)
		}

	case transport.StateDisconnected:
		es.mu.Lock()
		conn := es.connections[name]
		if conn == nil {
			es.mu.Unlock()
			return
		}
		wasOnline := conn.online
		conn.online = false
		es.mu.Unlock()

		// Only publish offline event if device was previously online
		// This avoids spurious offline events during initial connection failures
		if wasOnline {
			es.bus.Publish(events.NewDeviceOfflineEvent(name).WithReason("websocket disconnected"))
			iostreams.DebugCat(iostreams.CategoryNetwork, "Disconnected from %s", name)
		}

	case transport.StateReconnecting:
		debug.TraceEvent("ws: %s attempting reconnection", name)

	case transport.StateConnecting:
		debug.TraceEvent("ws: %s connecting", name)

	case transport.StateClosed:
		debug.TraceEvent("ws: %s closed", name)
	}
}

// pollGen1Device polls a Gen1 device for status updates.
func (es *EventStream) pollGen1Device(ctx context.Context, name, address string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Immediately get first status
	es.fetchGen1Status(ctx, name, address)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			es.fetchGen1Status(ctx, name, address)
		}
	}
}

func (es *EventStream) fetchGen1Status(ctx context.Context, name, address string) {
	es.mu.RLock()
	conn := es.connections[name]
	if conn == nil {
		es.mu.RUnlock()
		return
	}
	wasOnline := conn.online
	coiotID := conn.coiotID
	es.mu.RUnlock()

	// Skip HTTP polling if CoIoT is providing updates for this device
	if coiotID != "" && es.HasRecentCoIoTUpdate(coiotID) {
		debug.TraceEvent("poll: skipping HTTP for %s (CoIoT active)", name)
		return
	}

	statusJSON, err := es.svc.GetGen1StatusJSON(ctx, address)
	if err != nil {
		// Only emit offline event on state change (was online, now offline)
		if wasOnline {
			es.mu.Lock()
			conn.online = false
			es.mu.Unlock()
			es.bus.Publish(events.NewDeviceOfflineEvent(name).WithReason(err.Error()))
		}
		return
	}

	// Device responded - mark as online and emit event if state changed
	if !wasOnline {
		es.mu.Lock()
		conn.online = true
		es.mu.Unlock()
		es.bus.Publish(events.NewDeviceOnlineEvent(name).WithAddress(address))
	}

	es.bus.Publish(events.NewFullStatusEvent(name, statusJSON).
		WithSource(events.EventSourceLocal))
}

// Stop closes all WebSocket connections and stops event streaming.
func (es *EventStream) Stop() {
	es.cancel()
	es.pollerWg.Wait()

	// Stop CoIoT listener
	if es.coiotListener != nil {
		if err := es.coiotListener.Stop(); err != nil {
			iostreams.DebugErrCat(iostreams.CategoryNetwork, "stop coiot listener", err)
		}
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	for name, conn := range es.connections {
		conn.cancel()
		if conn.ws != nil {
			closeWS(conn.ws)
		}
		delete(es.connections, name)
	}

	es.bus.Close()
}

// AddDevice adds a new device to the event stream.
func (es *EventStream) AddDevice(name, address string) {
	es.mu.RLock()
	_, exists := es.connections[name]
	es.mu.RUnlock()

	if exists {
		return
	}

	go es.connectDevice(name, address)
}

// RemoveDevice removes a device from the event stream.
func (es *EventStream) RemoveDevice(name string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if conn, ok := es.connections[name]; ok {
		conn.cancel()
		if conn.ws != nil {
			closeWS(conn.ws)
		}
		delete(es.connections, name)
	}
}

// IsConnected returns whether a device is connected.
func (es *EventStream) IsConnected(name string) bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
	_, ok := es.connections[name]
	return ok
}

// ConnectedDevices returns the names of all connected devices.
func (es *EventStream) ConnectedDevices() []string {
	es.mu.RLock()
	defer es.mu.RUnlock()

	names := make([]string, 0, len(es.connections))
	for name := range es.connections {
		names = append(names, name)
	}
	return names
}

// ConnectionType represents how a device is connected to the event stream.
type ConnectionType int

const (
	// ConnectionNone means no connection to the device.
	ConnectionNone ConnectionType = iota
	// ConnectionWebSocket means the device is connected via WebSocket (Gen2+).
	ConnectionWebSocket
	// ConnectionPolling means the device is polled via HTTP (Gen1).
	ConnectionPolling
)

// String returns a display string for the connection type.
func (ct ConnectionType) String() string {
	switch ct {
	case ConnectionWebSocket:
		return "WS"
	case ConnectionPolling:
		return "HTTP"
	default:
		return "—"
	}
}

// Symbol returns a single character symbol for the connection type.
func (ct ConnectionType) Symbol() string {
	switch ct {
	case ConnectionWebSocket:
		return "⚡"
	case ConnectionPolling:
		return "↻"
	default:
		return "○"
	}
}

// ConnectionInfo holds connection details for a device.
type ConnectionInfo struct {
	Type       ConnectionType
	Generation int
}

// GetConnectionInfo returns connection info for a device.
func (es *EventStream) GetConnectionInfo(name string) ConnectionInfo {
	es.mu.RLock()
	defer es.mu.RUnlock()

	conn, ok := es.connections[name]
	if !ok {
		return ConnectionInfo{Type: ConnectionNone}
	}

	connType := ConnectionPolling
	if conn.ws != nil {
		connType = ConnectionWebSocket
	}

	return ConnectionInfo{
		Type:       connType,
		Generation: conn.generation,
	}
}

// GetAllConnectionInfo returns connection info for all connected devices.
func (es *EventStream) GetAllConnectionInfo() map[string]ConnectionInfo {
	es.mu.RLock()
	defer es.mu.RUnlock()

	info := make(map[string]ConnectionInfo, len(es.connections))
	for name, conn := range es.connections {
		connType := ConnectionPolling
		if conn.ws != nil {
			connType = ConnectionWebSocket
		}
		info[name] = ConnectionInfo{
			Type:       connType,
			Generation: conn.generation,
		}
	}
	return info
}

// Publish publishes a synthetic event to the event stream.
// This is used to emit events not originating from device connections,
// such as offline events from HTTP polling failures.
func (es *EventStream) Publish(evt events.Event) {
	if es.bus != nil {
		es.bus.Publish(evt)
	}
}

func closeWS(ws *transport.WebSocket) {
	if err := ws.Close(); err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "closing websocket", err)
	}
}

// startCoIoTListener starts the CoIoT multicast listener for Gen1 devices.
// CoIoT allows Gen1 devices to push status updates instead of polling.
func (es *EventStream) startCoIoTListener() {
	es.coiotListener = gen1.NewCoIoTListener()

	es.coiotListener.OnStatus(func(deviceID string, status *gen1.CoIoTStatus) {
		es.handleCoIoTStatus(deviceID, status)
	})

	if err := es.coiotListener.Start(); err != nil {
		// CoIoT listener is optional - log and continue with HTTP polling
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "start coiot listener", err)
		es.coiotListener = nil
		return
	}

	iostreams.DebugCat(iostreams.CategoryNetwork, "CoIoT listener started on %s:%d",
		gen1.CoIoTMulticastAddr, gen1.CoIoTPort)
}

// handleCoIoTStatus processes a CoIoT status update from a Gen1 device.
func (es *EventStream) handleCoIoTStatus(deviceID string, status *gen1.CoIoTStatus) {
	// deviceID from CoIoT is the MAC address (e.g., "shelly1-AABBCC")
	debug.TraceEvent("coiot: received status from %s", deviceID)

	// Record that we received a CoIoT update for this device
	es.coiotDevicesMu.Lock()
	es.coiotDevices[deviceID] = time.Now()
	es.coiotDevicesMu.Unlock()

	// Try to find the device name from our connection map
	es.coiotMACToNameMu.RLock()
	name, found := es.coiotMACToName[deviceID]
	es.coiotMACToNameMu.RUnlock()

	if !found {
		// Try to match by iterating connections (first time we see this device)
		es.mu.RLock()
		for connName, conn := range es.connections {
			if conn.generation == 1 {
				// For now, we can't easily match MAC to name without additional info
				// The deviceID format is like "shelly1-AABBCC" where AABBCC is part of MAC
				// We'll need to resolve this via device info
				_ = connName
			}
		}
		es.mu.RUnlock()

		// If we still can't find it, skip (device not in our config)
		debug.TraceEvent("coiot: unknown device %s, skipping", deviceID)
		return
	}

	// Publish event with the CoIoT status data
	// Convert CoIoT status to JSON for FullStatusEvent
	statusJSON, err := json.Marshal(status)
	if err != nil {
		debug.TraceEvent("coiot: failed to marshal status: %v", err)
		return
	}

	debug.TraceEvent("coiot: publishing status for %s", name)
	es.bus.Publish(events.NewFullStatusEvent(name, statusJSON).
		WithSource(events.EventSourceLocal))
}

// RegisterCoIoTDevice registers a device MAC for CoIoT status matching.
// This is called when we discover a Gen1 device's MAC address.
func (es *EventStream) RegisterCoIoTDevice(name, mac string) {
	es.coiotMACToNameMu.Lock()
	es.coiotMACToName[mac] = name
	es.coiotMACToNameMu.Unlock()
	debug.TraceEvent("coiot: registered %s -> %s", mac, name)
}

// HasRecentCoIoTUpdate checks if a device has received a CoIoT update recently.
// Used to skip HTTP polling for devices that are pushing via CoIoT.
func (es *EventStream) HasRecentCoIoTUpdate(mac string) bool {
	es.coiotDevicesMu.RLock()
	lastUpdate, found := es.coiotDevices[mac]
	es.coiotDevicesMu.RUnlock()

	if !found {
		return false
	}

	// Consider CoIoT updates "recent" if within 2x the default period (30s)
	// Default CoIoT period is 15s, so this gives some buffer
	return time.Since(lastUpdate) < 30*time.Second
}
