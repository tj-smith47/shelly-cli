// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/gen1"
	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/transport"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// maxComponentID is the maximum component ID to probe when discovering components.
const maxComponentID = 10

// MonitoringOptions configures real-time monitoring behavior.
type MonitoringOptions struct {
	Interval      time.Duration // Refresh interval for polling
	Count         int           // Number of updates (0 = unlimited)
	IncludePower  bool          // Include power meter data
	IncludeEnergy bool          // Include energy meter data
}

// MonitoringCallback is called with each status update during monitoring.
type MonitoringCallback func(model.MonitoringSnapshot) error

// EventHandler handles device events received via WebSocket.
type EventHandler func(model.DeviceEvent) error

// DeviceSnapshot holds the latest status for a device in multi-device monitoring.
type DeviceSnapshot struct {
	Device   string
	Address  string
	Info     *DeviceInfo
	Snapshot *model.MonitoringSnapshot
	Error    error
}

// GetEMStatus returns the status of an Energy Monitor (EM) component.
func (s *Service) GetEMStatus(ctx context.Context, device string, id int) (*model.EMStatus, error) {
	var result *model.EMStatus
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		em := components.NewEM(conn.RPCClient(), id)
		status, err := em.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = emStatusFromComponent(status)
		return nil
	})
	return result, err
}

// GetEM1Status returns the status of a single-phase Energy Monitor (EM1) component.
func (s *Service) GetEM1Status(ctx context.Context, device string, id int) (*model.EM1Status, error) {
	var result *model.EM1Status
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		em1 := components.NewEM1(conn.RPCClient(), id)
		status, err := em1.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = em1StatusFromComponent(status)
		return nil
	})
	return result, err
}

// GetPMStatus returns the status of a Power Meter (PM) component.
func (s *Service) GetPMStatus(ctx context.Context, device string, id int) (*model.PMStatus, error) {
	var result *model.PMStatus
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		pm := components.NewPM(conn.RPCClient(), id)
		status, err := pm.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = pmStatusFromComponent(status)
		return nil
	})
	return result, err
}

// GetPM1Status returns the status of a Power Meter (PM1) component.
func (s *Service) GetPM1Status(ctx context.Context, device string, id int) (*model.PMStatus, error) {
	var result *model.PMStatus
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		pm1 := components.NewPM1(conn.RPCClient(), id)
		status, err := pm1.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = pm1StatusFromComponent(status)
		return nil
	})
	return result, err
}

// ResetEMCounters resets energy counters on an EM component.
func (s *Service) ResetEMCounters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.WithConnection(ctx, device, func(conn *client.Client) error {
		em := components.NewEM(conn.RPCClient(), id)
		return em.ResetCounters(ctx, counterTypes)
	})
}

// ResetPMCounters resets energy counters on a PM component.
func (s *Service) ResetPMCounters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.WithConnection(ctx, device, func(conn *client.Client) error {
		pm := components.NewPM(conn.RPCClient(), id)
		return pm.ResetCounters(ctx, counterTypes)
	})
}

// ResetPM1Counters resets energy counters on a PM1 component.
func (s *Service) ResetPM1Counters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.WithConnection(ctx, device, func(conn *client.Client) error {
		pm1 := components.NewPM1(conn.RPCClient(), id)
		return pm1.ResetCounters(ctx, counterTypes)
	})
}

// ListEMComponents returns a list of EM component IDs on a device.
func (s *Service) ListEMComponents(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverEMComponents(ctx, conn)
		return nil
	})
	return ids, err
}

func discoverEMComponents(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		em := components.NewEM(conn.RPCClient(), id)
		if _, err := em.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

// ListEM1Components returns a list of EM1 component IDs on a device.
func (s *Service) ListEM1Components(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverEM1Components(ctx, conn)
		return nil
	})
	return ids, err
}

func discoverEM1Components(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		em1 := components.NewEM1(conn.RPCClient(), id)
		if _, err := em1.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

// ListPMComponents returns a list of PM component IDs on a device.
func (s *Service) ListPMComponents(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverPMComponents(ctx, conn)
		return nil
	})
	return ids, err
}

func discoverPMComponents(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		pm := components.NewPM(conn.RPCClient(), id)
		if _, err := pm.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

// ListPM1Components returns a list of PM1 component IDs on a device.
func (s *Service) ListPM1Components(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverPM1Components(ctx, conn)
		return nil
	})
	return ids, err
}

func discoverPM1Components(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		pm1 := components.NewPM1(conn.RPCClient(), id)
		if _, err := pm1.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

// MonitorDevice continuously monitors a device and calls the callback with updates.
// Runs until the context is cancelled or Count updates are received.
func (s *Service) MonitorDevice(ctx context.Context, device string, opts MonitoringOptions, callback MonitoringCallback) error {
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

func (s *Service) collectDeviceStatus(ctx context.Context, device string, opts MonitoringOptions) model.MonitoringSnapshot {
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
	opts := MonitoringOptions{
		IncludeEnergy: true,
		IncludePower:  true,
	}
	status := s.collectDeviceStatus(ctx, device, opts)
	return &status, nil
}

// GetMonitoringSnapshotAuto returns monitoring data for a device, auto-detecting generation.
func (s *Service) GetMonitoringSnapshotAuto(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	resolvedDevice, err := s.ResolveWithGeneration(ctx, device)

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
	err := s.WithGen1Connection(ctx, device, func(conn *client.Gen1Client) error {
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

			info, err := s.DeviceInfo(ctx, address)
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
	connErr := s.WithConnection(ctx, device, func(conn *client.Client) error {
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

// SubscribeEvents subscribes to real-time events from a device via WebSocket.
func (s *Service) SubscribeEvents(ctx context.Context, device string, handler EventHandler) error {
	resolved, err := s.resolver.Resolve(device)
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

// CollectPrometheusMetrics collects metrics from a device in Prometheus format.
func (s *Service) CollectPrometheusMetrics(ctx context.Context, device string) (*export.PrometheusMetrics, error) {
	metrics := &export.PrometheusMetrics{}
	online := true

	deviceStatus, err := s.DeviceStatus(ctx, device)
	if err != nil {
		online = false
	} else if deviceStatus.Status != nil {
		metrics.Metrics = append(metrics.Metrics, export.CollectSystemPrometheusMetrics(device, deviceStatus.Status)...)
	}

	pmIDs, pmErr := s.ListPMComponents(ctx, device)
	if pmErr != nil {
		online = false
	} else {
		pmMetrics := export.CollectMeterPrometheusMetrics(device, "pm", pmIDs, func(id int) (*model.PMStatus, error) {
			return s.GetPMStatus(ctx, device, id)
		})
		metrics.Metrics = append(metrics.Metrics, pmMetrics...)
	}

	if pm1IDs, err := s.ListPM1Components(ctx, device); err == nil {
		pm1Metrics := export.CollectMeterPrometheusMetrics(device, "pm1", pm1IDs, func(id int) (*model.PMStatus, error) {
			return s.GetPM1Status(ctx, device, id)
		})
		metrics.Metrics = append(metrics.Metrics, pm1Metrics...)
	}

	if em1IDs, err := s.ListEM1Components(ctx, device); err == nil {
		em1Metrics := export.CollectMeterPrometheusMetrics(device, "em1", em1IDs, func(id int) (*model.EM1Status, error) {
			return s.GetEM1Status(ctx, device, id)
		})
		metrics.Metrics = append(metrics.Metrics, em1Metrics...)
	}

	onlineVal := 0.0
	if online {
		onlineVal = 1.0
	}
	metrics.Metrics = append(metrics.Metrics, export.PrometheusMetric{
		Name: "shelly_device_online", Help: "Device online status (1=online, 0=offline)",
		Type: "gauge", Labels: map[string]string{"device": device}, Value: onlineVal,
	})

	metrics.Metrics = append(metrics.Metrics, s.collectEMPrometheusMetrics(ctx, device)...)

	return metrics, nil
}

func (s *Service) collectEMPrometheusMetrics(ctx context.Context, device string) []export.PrometheusMetric {
	readings := s.collectEMReadings(ctx, device)
	return export.ReadingsToPrometheusMetrics(readings)
}

// CollectComponentReadings collects all meter readings from a device.
func (s *Service) CollectComponentReadings(ctx context.Context, device string) []model.ComponentReading {
	var readings []model.ComponentReading

	if pmIDs, err := s.ListPMComponents(ctx, device); err == nil {
		readings = append(readings, export.CollectMeterReadings(device, "pm", pmIDs, func(id int) (*model.PMStatus, error) {
			return s.GetPMStatus(ctx, device, id)
		})...)
	}

	if pm1IDs, err := s.ListPM1Components(ctx, device); err == nil {
		readings = append(readings, export.CollectMeterReadings(device, "pm1", pm1IDs, func(id int) (*model.PMStatus, error) {
			return s.GetPM1Status(ctx, device, id)
		})...)
	}

	if em1IDs, err := s.ListEM1Components(ctx, device); err == nil {
		readings = append(readings, export.CollectMeterReadings(device, "em1", em1IDs, func(id int) (*model.EM1Status, error) {
			return s.GetEM1Status(ctx, device, id)
		})...)
	}

	readings = append(readings, s.collectEMReadings(ctx, device)...)
	return readings
}

func (s *Service) collectEMReadings(ctx context.Context, device string) []model.ComponentReading {
	emIDs, err := s.ListEMComponents(ctx, device)
	if err != nil {
		return nil
	}
	emStatuses := make([]*model.EMStatus, 0, len(emIDs))
	for _, id := range emIDs {
		status, err := s.GetEMStatus(ctx, device, id)
		if err != nil {
			continue
		}
		emStatuses = append(emStatuses, status)
	}
	return export.CollectEMReadings(device, emStatuses)
}

// CollectInfluxDBPoints collects metrics from a device in InfluxDB line protocol format.
func (s *Service) CollectInfluxDBPoints(ctx context.Context, device string) ([]export.InfluxDBPoint, error) {
	var points []export.InfluxDBPoint
	now := time.Now()

	pm1IDs, err := s.ListPM1Components(ctx, device)
	if err == nil {
		for _, id := range pm1IDs {
			status, err := s.GetPM1Status(ctx, device, id)
			if err != nil {
				continue
			}

			fields := map[string]float64{
				"power":   status.APower,
				"voltage": status.Voltage,
				"current": status.Current,
			}

			if status.AEnergy != nil {
				fields["energy"] = status.AEnergy.Total
			}
			if status.Freq != nil {
				fields["frequency"] = *status.Freq
			}

			points = append(points, export.InfluxDBPoint{
				Measurement: "shelly_power",
				Tags: map[string]string{
					"device":       device,
					"component":    "pm1",
					"component_id": fmt.Sprintf("%d", id),
				},
				Fields:    fields,
				Timestamp: now,
			})
		}
	}

	emIDs, err := s.ListEMComponents(ctx, device)
	if err == nil {
		for _, id := range emIDs {
			status, err := s.GetEMStatus(ctx, device, id)
			if err != nil {
				continue
			}
			points = append(points, export.EMReadingsToInfluxDBPoints([]*model.EMStatus{status}, device, now)...)
		}
	}

	return points, nil
}

// GetEMDataRecords retrieves available time intervals with stored EMData.
func (s *Service) GetEMDataRecords(ctx context.Context, device string, id int, fromTS *int64) (*components.EMDataRecordsResult, error) {
	var result *components.EMDataRecordsResult
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		emdata := components.NewEMData(conn.RPCClient(), id)
		var err error
		result, err = emdata.GetRecords(ctx, fromTS)
		return err
	})
	return result, err
}

// GetEMDataHistory retrieves historical EMData measurements for a time range.
func (s *Service) GetEMDataHistory(ctx context.Context, device string, id int, startTS, endTS *int64) (*components.EMDataGetDataResult, error) {
	var result *components.EMDataGetDataResult
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		emdata := components.NewEMData(conn.RPCClient(), id)
		var err error
		result, err = emdata.GetData(ctx, startTS, endTS)
		return err
	})
	return result, err
}

// DeleteEMData deletes all stored historical EMData.
func (s *Service) DeleteEMData(ctx context.Context, device string, id int) error {
	return s.WithConnection(ctx, device, func(conn *client.Client) error {
		emdata := components.NewEMData(conn.RPCClient(), id)
		return emdata.DeleteAllData(ctx)
	})
}

// GetEMDataCSVURL returns the HTTP URL for downloading EMData as CSV.
func (s *Service) GetEMDataCSVURL(device string, id int, startTS, endTS *int64, addKeys bool) (string, error) {
	addr, err := s.resolver.Resolve(device)
	if err != nil {
		return "", err
	}
	return components.EMDataCSVURL(addr.Address, id, startTS, endTS, addKeys), nil
}

// GetEM1DataRecords retrieves available time intervals with stored EM1Data.
func (s *Service) GetEM1DataRecords(ctx context.Context, device string, id int, fromTS *int64) (*components.EM1DataRecordsResult, error) {
	var result *components.EM1DataRecordsResult
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		em1data := components.NewEM1Data(conn.RPCClient(), id)
		var err error
		result, err = em1data.GetRecords(ctx, fromTS)
		return err
	})
	return result, err
}

// GetEM1DataHistory retrieves historical EM1Data measurements for a time range.
func (s *Service) GetEM1DataHistory(ctx context.Context, device string, id int, startTS, endTS *int64) (*components.EM1DataGetDataResult, error) {
	var result *components.EM1DataGetDataResult
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		em1data := components.NewEM1Data(conn.RPCClient(), id)
		var err error
		result, err = em1data.GetData(ctx, startTS, endTS)
		return err
	})
	return result, err
}

// DeleteEM1Data deletes all stored historical EM1Data.
func (s *Service) DeleteEM1Data(ctx context.Context, device string, id int) error {
	return s.WithConnection(ctx, device, func(conn *client.Client) error {
		em1data := components.NewEM1Data(conn.RPCClient(), id)
		return em1data.DeleteAllData(ctx)
	})
}

// GetEM1DataCSVURL returns the HTTP URL for downloading EM1Data as CSV.
func (s *Service) GetEM1DataCSVURL(device string, id int, startTS, endTS *int64, addKeys bool) (string, error) {
	addr, err := s.resolver.Resolve(device)
	if err != nil {
		return "", err
	}
	return components.EM1DataCSVURL(addr.Address, id, startTS, endTS, addKeys), nil
}

// CollectJSONMetrics collects metrics from multiple devices for JSON output.
func (s *Service) CollectJSONMetrics(ctx context.Context, devices []string) export.JSONMetricsOutput {
	output := export.JSONMetricsOutput{
		Timestamp: time.Now(),
		Devices:   make([]export.JSONMetricsDevice, len(devices)),
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	var mu sync.Mutex

	for i, device := range devices {
		idx := i
		dev := device
		g.Go(func() error {
			readings := s.CollectComponentReadings(ctx, dev)
			mu.Lock()
			output.Devices[idx] = export.JSONMetricsDevice{
				Device:     dev,
				Online:     len(readings) > 0,
				Components: readings,
			}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return output
	}

	return output
}

// PrometheusCollector collects and caches Prometheus metrics from multiple devices.
type PrometheusCollector struct {
	svc     *Service
	devices []string

	mu      sync.RWMutex
	metrics map[string]*export.PrometheusMetrics
	errors  map[string]error
}

// NewPrometheusCollector creates a new Prometheus metrics collector.
func NewPrometheusCollector(svc *Service, devices []string) *PrometheusCollector {
	return &PrometheusCollector{
		svc:     svc,
		devices: devices,
		metrics: make(map[string]*export.PrometheusMetrics),
		errors:  make(map[string]error),
	}
}

// Collect fetches metrics from all configured devices concurrently.
func (c *PrometheusCollector) Collect(ctx context.Context) {
	newMetrics := make(map[string]*export.PrometheusMetrics)
	newErrors := make(map[string]error)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	var mu sync.Mutex

	for _, device := range c.devices {
		dev := device
		g.Go(func() error {
			m, err := c.svc.CollectPrometheusMetrics(ctx, dev)
			mu.Lock()
			if err != nil {
				newErrors[dev] = err
			} else {
				newMetrics[dev] = m
			}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		panic("unexpected error from errgroup: " + err.Error())
	}

	c.mu.Lock()
	c.metrics = newMetrics
	c.errors = newErrors
	c.mu.Unlock()
}

// Errors returns any collection errors from the last Collect call.
func (c *PrometheusCollector) Errors() map[string]error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]error, len(c.errors))
	for k, v := range c.errors {
		result[k] = v
	}
	return result
}

// FormatMetrics returns the collected metrics in Prometheus exposition format.
func (c *PrometheusCollector) FormatMetrics() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	combined := &export.PrometheusMetrics{}
	for _, m := range c.metrics {
		if m != nil {
			combined.Metrics = append(combined.Metrics, m.Metrics...)
		}
	}

	return export.FormatPrometheusMetrics(combined)
}

// StreamInfluxDBPoints continuously collects and outputs InfluxDB points at the given interval.
// It calls the writePoints function with collected points on each tick.
// Returns when context is cancelled.
func (s *Service) StreamInfluxDBPoints(ctx context.Context, devices []string, measurement string, tags map[string]string, interval time.Duration, writePoints func([]export.InfluxDBPoint)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			points := s.CollectInfluxDBPointsMulti(ctx, devices, measurement, tags)
			writePoints(points)
		}
	}
}

// CollectInfluxDBPointsMulti collects InfluxDB points from multiple devices concurrently.
func (s *Service) CollectInfluxDBPointsMulti(ctx context.Context, devices []string, measurement string, tags map[string]string) []export.InfluxDBPoint {
	now := time.Now()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	var mu sync.Mutex
	var allPoints []export.InfluxDBPoint

	for _, device := range devices {
		dev := device
		g.Go(func() error {
			readings := s.CollectComponentReadings(ctx, dev)
			points := export.ReadingsToInfluxDBPoints(readings, now)
			for i := range points {
				points[i].Measurement = measurement
				for k, v := range tags {
					points[i].Tags[k] = v
				}
			}
			mu.Lock()
			allPoints = append(allPoints, points...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return allPoints
	}

	return allPoints
}

// Component conversion helpers (previously in monitoring subpackage)

// emStatusFromComponent converts shelly-go EMStatus to model.EMStatus.
func emStatusFromComponent(c *components.EMStatus) *model.EMStatus {
	return &model.EMStatus{
		ID:               c.ID,
		AVoltage:         c.AVoltage,
		ACurrent:         c.ACurrent,
		AActivePower:     c.AActivePower,
		AApparentPower:   c.AApparentPower,
		APowerFactor:     c.APowerFactor,
		AFreq:            c.AFreq,
		BVoltage:         c.BVoltage,
		BCurrent:         c.BCurrent,
		BActivePower:     c.BActivePower,
		BApparentPower:   c.BApparentPower,
		BPowerFactor:     c.BPowerFactor,
		BFreq:            c.BFreq,
		CVoltage:         c.CVoltage,
		CCurrent:         c.CCurrent,
		CActivePower:     c.CActivePower,
		CApparentPower:   c.CApparentPower,
		CPowerFactor:     c.CPowerFactor,
		CFreq:            c.CFreq,
		NCurrent:         c.NCurrent,
		TotalCurrent:     c.TotalCurrent,
		TotalActivePower: c.TotalActivePower,
		TotalAprtPower:   c.TotalApparentPower,
		Errors:           c.Errors,
	}
}

// em1StatusFromComponent converts shelly-go EM1Status to model.EM1Status.
func em1StatusFromComponent(c *components.EM1Status) *model.EM1Status {
	return &model.EM1Status{
		ID:        c.ID,
		Voltage:   c.Voltage,
		Current:   c.Current,
		ActPower:  c.ActPower,
		AprtPower: c.AprtPower,
		PF:        c.PF,
		Freq:      c.Freq,
		Errors:    c.Errors,
	}
}

// pmStatusFromComponent converts shelly-go PMStatus to model.PMStatus.
func pmStatusFromComponent(c *components.PMStatus) *model.PMStatus {
	result := &model.PMStatus{
		ID:      c.ID,
		Voltage: c.Voltage,
		Current: c.Current,
		APower:  c.APower,
		Freq:    c.Freq,
		Errors:  c.Errors,
	}
	if c.AEnergy != nil {
		result.AEnergy = &model.PMEnergyCounters{
			Total:    c.AEnergy.Total,
			ByMinute: c.AEnergy.ByMinute,
			MinuteTs: c.AEnergy.MinuteTs,
		}
	}
	if c.RetAEnergy != nil {
		result.RetAEnergy = &model.PMEnergyCounters{
			Total:    c.RetAEnergy.Total,
			ByMinute: c.RetAEnergy.ByMinute,
			MinuteTs: c.RetAEnergy.MinuteTs,
		}
	}
	return result
}

// pm1StatusFromComponent converts shelly-go PM1Status to model.PMStatus.
func pm1StatusFromComponent(c *components.PM1Status) *model.PMStatus {
	result := &model.PMStatus{
		ID:      c.ID,
		Voltage: c.Voltage,
		Current: c.Current,
		APower:  c.APower,
		Freq:    c.Freq,
		Errors:  c.Errors,
	}
	if c.AEnergy != nil {
		result.AEnergy = &model.PMEnergyCounters{
			Total:    c.AEnergy.Total,
			ByMinute: c.AEnergy.ByMinute,
			MinuteTs: c.AEnergy.MinuteTs,
		}
	}
	if c.RetAEnergy != nil {
		result.RetAEnergy = &model.PMEnergyCounters{
			Total:    c.RetAEnergy.Total,
			ByMinute: c.RetAEnergy.ByMinute,
			MinuteTs: c.RetAEnergy.MinuteTs,
		}
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
