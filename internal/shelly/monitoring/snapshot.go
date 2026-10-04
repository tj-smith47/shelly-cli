package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/gen1"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// MonitorDevice continuously monitors a device and calls the callback with updates.
// Runs until the context is cancelled or Count updates are received.
func (s *Service) MonitorDevice(ctx context.Context, device string, opts Options, callback Callback) error {
	if opts.Interval == 0 {
		opts.Interval = 2 * time.Second
	}

	// Default to include all data if nothing specified
	if !opts.IncludeEnergy && !opts.IncludePower {
		opts.IncludeEnergy = true
		opts.IncludePower = true
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	updates := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			status := s.collectDeviceStatus(ctx, device, opts)
			if err := callback(status); err != nil {
				return err
			}
			updates++
			if opts.Count > 0 && updates >= opts.Count {
				return nil
			}
		}
	}
}

func (s *Service) collectDeviceStatus(ctx context.Context, device string, opts Options) model.MonitoringSnapshot {
	status := model.MonitoringSnapshot{
		Device:    device,
		Timestamp: time.Now(),
		Online:    true,
	}
	readings, err := s.ReadPowerReadings(ctx, device)
	if err != nil {
		status.Online = false
		status.Error = err.Error()
		return status
	}
	addReadingsToSnapshot(&status, readings, opts)
	return status
}

// addReadingsToSnapshot files each reading under EM, EM1 or PM. A switch,
// cover, light or Gen1 meter reading goes under PM, the shape they share.
func addReadingsToSnapshot(snapshot *model.MonitoringSnapshot, readings []model.PowerReading, opts Options) {
	for _, r := range readings {
		switch {
		case r.EM != nil:
			if opts.IncludeEnergy {
				snapshot.EM = append(snapshot.EM, *r.EM)
			}
		case r.EM1 != nil:
			if opts.IncludeEnergy {
				snapshot.EM1 = append(snapshot.EM1, *r.EM1)
			}
		case r.Meter != nil:
			if opts.IncludePower {
				snapshot.PM = append(snapshot.PM, *r.Meter)
			}
		}
	}
}

// GetMonitoringSnapshot returns a single snapshot of all monitoring data for
// a device, Gen1 or Gen2+.
func (s *Service) GetMonitoringSnapshot(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	readings, err := s.ReadPowerReadings(ctx, device)
	if err != nil {
		return nil, err
	}
	snapshot := &model.MonitoringSnapshot{Device: device, Timestamp: time.Now(), Online: true}
	addReadingsToSnapshot(snapshot, readings, Options{IncludeEnergy: true, IncludePower: true})
	return snapshot, nil
}

// GetGen1StatusJSON returns Gen1 device status as JSON for event streaming.
func (s *Service) GetGen1StatusJSON(ctx context.Context, identifier string) (json.RawMessage, error) {
	snapshot, err := s.GetMonitoringSnapshot(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return json.Marshal(snapshot)
}

// Gen1DeviceInfo holds Gen1 device identification for CoIoT registration.
type Gen1DeviceInfo struct {
	Type string // Device type (e.g., "SHSW-PM")
	MAC  string // MAC address (e.g., "C45BBE6C2D3A")
}

// GetGen1DeviceInfo returns the Gen1 device type and MAC for CoIoT registration.
func (s *Service) GetGen1DeviceInfo(ctx context.Context, identifier string) (*Gen1DeviceInfo, error) {
	var info *Gen1DeviceInfo
	err := s.connector.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		devInfo := conn.Info()
		if devInfo == nil {
			return fmt.Errorf("device info not available")
		}
		info = &Gen1DeviceInfo{
			Type: devInfo.Model, // Model holds the Gen1 Type (e.g., "SHSW-PM")
			MAC:  devInfo.MAC,
		}
		return nil
	})
	return info, err
}

// FetchAllSnapshots fetches device info and monitoring snapshots for all devices concurrently.
func (s *Service) FetchAllSnapshots(ctx context.Context, devices map[string]string, snapshots map[string]*DeviceSnapshot, mu *sync.Mutex) {
	var wg sync.WaitGroup
	for name, address := range devices {
		wg.Go(func() {
			snapshot := &DeviceSnapshot{
				Device:  name,
				Address: address,
			}

			info, err := s.connector.DeviceInfo(ctx, address)
			if err != nil {
				snapshot.Error = err
			} else {
				snapshot.Info = info
			}

			if snapshot.Error == nil {
				snap, err := s.GetMonitoringSnapshot(ctx, address)
				if err != nil {
					snapshot.Error = err
				} else {
					snapshot.Snapshot = snap
				}
			}

			mu.Lock()
			snapshots[name] = snapshot
			mu.Unlock()
		})
	}
	wg.Wait()
}

// SubscribeEvents subscribes to real-time events from a device via WebSocket.
func (s *Service) SubscribeEvents(ctx context.Context, device string, handler EventHandler) error {
	resolved, err := s.connector.Resolve(device)
	if err != nil {
		return fmt.Errorf("failed to resolve device: %w", err)
	}

	wsURL := fmt.Sprintf("ws://%s/rpc", resolved.Address)
	ws, err := client.NewDeviceWebSocket(wsURL, resolved.Auth,
		transport.WithReconnect(true),
		transport.WithPingInterval(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if err := ws.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer closeWebSocket(ws)

	notifyHandler := func(msg json.RawMessage) {
		iostreams.DebugCat(iostreams.CategoryNetwork, "WebSocket notification received: %s", string(msg))

		var notif struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params,omitempty"`
		}
		if err := json.Unmarshal(msg, &notif); err != nil {
			iostreams.DebugErrCat(iostreams.CategoryNetwork, "failed to parse notification", err)
			return
		}

		iostreams.DebugCat(iostreams.CategoryNetwork, "Parsed notification method: %s", notif.Method)
		event := parseNotification(device, notif.Method, notif.Params)

		if err := handler(event); err != nil {
			iostreams.DebugErrCat(iostreams.CategoryNetwork, "event handler error", err)
		}
	}

	if err := ws.Subscribe(notifyHandler); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// A device sends no notifications until it has answered one request frame,
	// and on a password-protected device that frame must authenticate.
	if _, err := client.StartDeviceNotifications(ctx, ws); err != nil {
		return fmt.Errorf("start event stream: %w", err)
	}

	<-ctx.Done()
	return ctx.Err()
}

func closeWebSocket(ws *client.DeviceWebSocket) {
	if err := ws.Close(); err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "closing websocket", err)
	}
}

// Gen1 conversion helpers

// convertGen1Meters converts Gen1 meters to model.PMStatus.
func convertGen1Meters(meters []gen1.MeterStatus) []model.PMStatus {
	result := make([]model.PMStatus, len(meters))
	for i, m := range meters {
		result[i] = model.PMStatus{ID: i, APower: m.Power}
		// A Gen1 meter counts its total in watt-minutes; every other meter,
		// and model.PMEnergyCounters, counts watt-hours.
		if m.Total > 0 {
			result[i].AEnergy = &model.PMEnergyCounters{Total: float64(m.Total) / 60}
		}
	}
	return result
}

// convertGen1EMeters converts Gen1 emeters to model.PMStatus.
func convertGen1EMeters(emeters []gen1.EMeterStatus, startID int) []model.PMStatus {
	result := make([]model.PMStatus, len(emeters))
	for i, em := range emeters {
		result[i] = model.PMStatus{
			ID:      startID + i,
			APower:  em.Power,
			Voltage: em.Voltage,
			Current: em.Current,
		}
		if em.Total > 0 {
			result[i].AEnergy = &model.PMEnergyCounters{Total: em.Total}
		}
	}
	return result
}

// Event parsing helpers

// parseNotification converts a WebSocket notification to a DeviceEvent.
func parseNotification(device, method string, params map[string]any) model.DeviceEvent {
	event := model.DeviceEvent{
		Device:    device,
		Timestamp: time.Now(),
		Event:     method,
		Data:      params,
	}

	// Parse component source from notification params
	if src, ok := params["src"].(string); ok {
		compType, compID := parseComponentSource(src)
		event.Component = compType
		event.ComponentID = compID
	}

	return event
}

// parseComponentSource extracts component type and ID from a source string.
// Example: "switch:0" -> ("switch", 0).
func parseComponentSource(src string) (compType string, compID int) {
	if _, err := fmt.Sscanf(src, "%[^:]:%d", &compType, &compID); err != nil {
		return src, 0
	}
	return compType, compID
}
