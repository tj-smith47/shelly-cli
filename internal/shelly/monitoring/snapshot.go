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

	if opts.IncludeEnergy {
		status.EM = s.collectEMStatus(ctx, device)
		status.EM1 = s.collectEM1Status(ctx, device)
	}

	if opts.IncludePower {
		status.PM = s.collectPMStatus(ctx, device)
	}

	return status
}

// GetMonitoringSnapshot returns a single snapshot of all monitoring data for a device.
func (s *Service) GetMonitoringSnapshot(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	opts := Options{
		IncludeEnergy: true,
		IncludePower:  true,
	}
	status := s.collectDeviceStatus(ctx, device, opts)
	return &status, nil
}

// GetMonitoringSnapshotAuto returns monitoring data for a device, auto-detecting generation.
func (s *Service) GetMonitoringSnapshotAuto(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	resolvedDevice, err := s.connector.ResolveWithGeneration(ctx, device)

	// If we know it's Gen1, try Gen1 first
	if err == nil && resolvedDevice.Generation == 1 {
		gen1Snapshot, gen1Err := s.getGen1MonitoringSnapshot(ctx, device)
		if gen1Err == nil {
			return gen1Snapshot, nil
		}
		snapshot, err := s.GetMonitoringSnapshot(ctx, device)
		if err == nil && (len(snapshot.PM) > 0 || len(snapshot.EM) > 0 || len(snapshot.EM1) > 0) {
			return snapshot, nil
		}
		return &model.MonitoringSnapshot{Device: device, Timestamp: time.Now(), Online: true}, nil
	}

	// Gen2+ or unknown: Try Gen2 first
	snapshot, err := s.GetMonitoringSnapshot(ctx, device)
	if err == nil && (len(snapshot.PM) > 0 || len(snapshot.EM) > 0 || len(snapshot.EM1) > 0) {
		return snapshot, nil
	}

	gen1Snapshot, gen1Err := s.getGen1MonitoringSnapshot(ctx, device)
	if gen1Err == nil {
		return gen1Snapshot, nil
	}

	if err == nil {
		return snapshot, nil
	}
	return nil, err
}

func (s *Service) getGen1MonitoringSnapshot(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	var snapshot *model.MonitoringSnapshot
	err := s.connector.WithGen1Connection(ctx, device, func(conn *client.Gen1Client) error {
		status, err := conn.GetStatus(ctx)
		if err != nil {
			return err
		}

		snapshot = &model.MonitoringSnapshot{
			Device:    device,
			Timestamp: time.Now(),
			Online:    true,
		}

		snapshot.PM = convertGen1Meters(status.Meters)
		snapshot.PM = append(snapshot.PM, convertGen1EMeters(status.EMeters, len(status.Meters))...)

		return nil
	})
	return snapshot, err
}

// GetGen1StatusJSON returns Gen1 device status as JSON for event streaming.
func (s *Service) GetGen1StatusJSON(ctx context.Context, identifier string) (json.RawMessage, error) {
	snapshot, err := s.getGen1MonitoringSnapshot(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return json.Marshal(snapshot)
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
	ws := transport.NewWebSocket(wsURL,
		transport.WithReconnect(true),
		transport.WithPingInterval(30*time.Second),
	)

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

	<-ctx.Done()
	return ctx.Err()
}

func closeWebSocket(ws *transport.WebSocket) {
	if err := ws.Close(); err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "closing websocket", err)
	}
}

// Collection helpers

func (s *Service) collectEMStatus(ctx context.Context, device string) []model.EMStatus {
	var result []model.EMStatus
	emIDs, err := s.ListEMComponents(ctx, device)
	if err != nil {
		return result
	}
	for _, id := range emIDs {
		if emStatus, err := s.GetEMStatus(ctx, device, id); err == nil {
			result = append(result, *emStatus)
		}
	}
	return result
}

func (s *Service) collectEM1Status(ctx context.Context, device string) []model.EM1Status {
	var result []model.EM1Status
	em1IDs, err := s.ListEM1Components(ctx, device)
	if err != nil {
		return result
	}
	for _, id := range em1IDs {
		if em1Status, err := s.GetEM1Status(ctx, device, id); err == nil {
			result = append(result, *em1Status)
		}
	}
	return result
}

func (s *Service) collectPMStatus(ctx context.Context, device string) []model.PMStatus {
	var result []model.PMStatus
	pmIDs, err := s.ListPMComponents(ctx, device)
	if err == nil {
		for _, id := range pmIDs {
			if pmStatus, err := s.GetPMStatus(ctx, device, id); err == nil {
				result = append(result, *pmStatus)
			}
		}
	}
	pm1IDs, err := s.ListPM1Components(ctx, device)
	if err == nil {
		for _, id := range pm1IDs {
			if pm1Status, err := s.GetPM1Status(ctx, device, id); err == nil {
				result = append(result, *pm1Status)
			}
		}
	}
	result = append(result, s.collectSwitchPowerStatus(ctx, device)...)
	return result
}

func (s *Service) collectSwitchPowerStatus(ctx context.Context, device string) []model.PMStatus {
	var result []model.PMStatus
	connErr := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		comps, err := conn.FilterComponents(ctx, model.ComponentSwitch)
		if err != nil {
			iostreams.DebugErrCat(iostreams.CategoryDevice, "list switches for power", err)
			return err
		}
		for _, comp := range comps {
			status, err := conn.Switch(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			if status.Power != nil {
				pm := model.PMStatus{
					ID:     100 + comp.ID,
					APower: *status.Power,
				}
				if status.Voltage != nil {
					pm.Voltage = *status.Voltage
				}
				if status.Current != nil {
					pm.Current = *status.Current
				}
				result = append(result, pm)
			}
		}
		return nil
	})
	if connErr != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "switch power collection", connErr)
	}
	return result
}

// Gen1 conversion helpers

// convertGen1Meters converts Gen1 meters to model.PMStatus.
func convertGen1Meters(meters []gen1.MeterStatus) []model.PMStatus {
	result := make([]model.PMStatus, len(meters))
	for i, m := range meters {
		result[i] = model.PMStatus{
			ID:      i,
			APower:  m.Power,
			Voltage: 0, // Gen1 meters don't expose voltage directly
			Current: 0, // Gen1 meters don't expose current directly
		}
		if m.Total > 0 {
			result[i].AEnergy = &model.PMEnergyCounters{Total: float64(m.Total)}
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
