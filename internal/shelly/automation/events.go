// Package automation provides script, schedule, and event automation for Shelly devices.
package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/events"
	"github.com/tj-smith47/shelly-go/notifications"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
}

type deviceConnection struct {
	name       string
	address    string
	ws         *transport.WebSocket
	cancel     context.CancelFunc
	generation int
}

// NewEventStream creates a new event stream manager.
func NewEventStream(svc EventStreamProvider) *EventStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &EventStream{
		svc:         svc,
		bus:         events.NewEventBus(),
		connections: make(map[string]*deviceConnection),
		ctx:         ctx,
		cancel:      cancel,
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
// It connects to Gen2+ devices via WebSocket and polls Gen1 devices.
func (es *EventStream) Start() error {
	devices := config.ListDevices()
	if len(devices) == 0 {
		return nil
	}

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

	// Check device generation
	resolvedDevice, err := es.svc.ResolveWithGeneration(ctx, address)
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

	if err := ws.Connect(ctx); err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, fmt.Sprintf("connect websocket %s", name), err)
		es.bus.Publish(events.NewDeviceOfflineEvent(name).WithReason(err.Error()))
		cancel()
		return
	}

	// Register notification handler
	notifyHandler := func(msg json.RawMessage) {
		if !notifications.IsNotification(msg) {
			return
		}

		evts, err := notifications.ParseGen2NotificationJSON(name, msg)
		if err != nil {
			iostreams.DebugErrCat(iostreams.CategoryNetwork, "parse notification", err)
			return
		}

		for _, evt := range evts {
			es.bus.Publish(evt)
		}
	}

	if err := ws.Subscribe(notifyHandler); err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, fmt.Sprintf("subscribe %s", name), err)
		closeWS(ws)
		cancel()
		return
	}

	// Store connection
	es.mu.Lock()
	es.connections[name] = &deviceConnection{
		name:       name,
		address:    address,
		ws:         ws,
		cancel:     cancel,
		generation: resolvedDevice.Generation,
	}
	es.mu.Unlock()

	// Publish online event
	es.bus.Publish(events.NewDeviceOnlineEvent(name).WithAddress(address))

	iostreams.DebugCat(iostreams.CategoryNetwork, "Connected to %s via WebSocket", name)
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
	statusJSON, err := es.svc.GetGen1StatusJSON(ctx, address)
	if err != nil {
		es.bus.Publish(events.NewDeviceOfflineEvent(name).WithReason(err.Error()))
		return
	}

	es.bus.Publish(events.NewFullStatusEvent(name, statusJSON).
		WithSource(events.EventSourceLocal))
}

// Stop closes all WebSocket connections and stops event streaming.
func (es *EventStream) Stop() {
	es.cancel()
	es.pollerWg.Wait()

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
