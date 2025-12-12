// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// EMStatus represents the status of an Energy Monitor (EM) component.
type EMStatus struct {
	ID               int      `json:"id"`
	AVoltage         float64  `json:"a_voltage"`
	ACurrent         float64  `json:"a_current"`
	AActivePower     float64  `json:"a_act_power"`
	AApparentPower   float64  `json:"a_aprt_power"`
	APowerFactor     *float64 `json:"a_pf,omitempty"`
	AFreq            *float64 `json:"a_freq,omitempty"`
	BVoltage         float64  `json:"b_voltage"`
	BCurrent         float64  `json:"b_current"`
	BActivePower     float64  `json:"b_act_power"`
	BApparentPower   float64  `json:"b_aprt_power"`
	BPowerFactor     *float64 `json:"b_pf,omitempty"`
	BFreq            *float64 `json:"b_freq,omitempty"`
	CVoltage         float64  `json:"c_voltage"`
	CCurrent         float64  `json:"c_current"`
	CActivePower     float64  `json:"c_act_power"`
	CApparentPower   float64  `json:"c_aprt_power"`
	CPowerFactor     *float64 `json:"c_pf,omitempty"`
	CFreq            *float64 `json:"c_freq,omitempty"`
	NCurrent         *float64 `json:"n_current,omitempty"`
	TotalCurrent     float64  `json:"total_current"`
	TotalActivePower float64  `json:"total_act_power"`
	TotalAprtPower   float64  `json:"total_aprt_power"`
	Errors           []string `json:"errors,omitempty"`
}

// EM1Status represents the status of a single-phase Energy Monitor (EM1) component.
type EM1Status struct {
	ID        int      `json:"id"`
	Voltage   float64  `json:"voltage"`
	Current   float64  `json:"current"`
	ActPower  float64  `json:"act_power"`
	AprtPower float64  `json:"aprt_power"`
	PF        *float64 `json:"pf,omitempty"`
	Freq      *float64 `json:"freq,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}

// PMStatus represents the status of a Power Meter (PM/PM1) component.
type PMStatus struct {
	ID         int             `json:"id"`
	Voltage    float64         `json:"voltage"`
	Current    float64         `json:"current"`
	APower     float64         `json:"apower"`
	Freq       *float64        `json:"freq,omitempty"`
	AEnergy    *EnergyCounters `json:"aenergy,omitempty"`
	RetAEnergy *EnergyCounters `json:"ret_aenergy,omitempty"`
	Errors     []string        `json:"errors,omitempty"`
}

// EnergyCounters represents accumulated energy measurements.
type EnergyCounters struct {
	Total    float64   `json:"total"`
	ByMinute []float64 `json:"by_minute,omitempty"`
	MinuteTs *int64    `json:"minute_ts,omitempty"`
}

// EnergyHistory represents historical energy data.
type EnergyHistory struct {
	Period   string         `json:"period"`
	From     time.Time      `json:"from"`
	To       time.Time      `json:"to"`
	DeviceID string         `json:"device_id"`
	Data     []EnergyRecord `json:"data"`
}

// EnergyRecord represents a single energy measurement record.
type EnergyRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Energy    float64   `json:"energy"`  // Wh
	Power     float64   `json:"power"`   // W (average)
	Voltage   float64   `json:"voltage"` // V (average)
	Current   float64   `json:"current"` // A (average)
}

// DeviceEvent represents a real-time event from a device.
type DeviceEvent struct {
	Device      string         `json:"device"`
	Timestamp   time.Time      `json:"timestamp"`
	Event       string         `json:"event"`
	Component   string         `json:"component"`
	ComponentID int            `json:"component_id"`
	Data        map[string]any `json:"data,omitempty"`
}

// MonitorOptions configures real-time monitoring behavior.
type MonitorOptions struct {
	Interval      time.Duration // Refresh interval for polling
	Count         int           // Number of updates (0 = unlimited)
	IncludePower  bool          // Include power meter data
	IncludeEnergy bool          // Include energy meter data
}

// GetEMStatus returns the status of an Energy Monitor (EM) component.
func (s *Service) GetEMStatus(ctx context.Context, device string, id int) (*EMStatus, error) {
	var result *EMStatus
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		em := components.NewEM(conn.RPCClient(), id)
		status, err := em.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &EMStatus{
			ID:               status.ID,
			AVoltage:         status.AVoltage,
			ACurrent:         status.ACurrent,
			AActivePower:     status.AActivePower,
			AApparentPower:   status.AApparentPower,
			APowerFactor:     status.APowerFactor,
			AFreq:            status.AFreq,
			BVoltage:         status.BVoltage,
			BCurrent:         status.BCurrent,
			BActivePower:     status.BActivePower,
			BApparentPower:   status.BApparentPower,
			BPowerFactor:     status.BPowerFactor,
			BFreq:            status.BFreq,
			CVoltage:         status.CVoltage,
			CCurrent:         status.CCurrent,
			CActivePower:     status.CActivePower,
			CApparentPower:   status.CApparentPower,
			CPowerFactor:     status.CPowerFactor,
			CFreq:            status.CFreq,
			NCurrent:         status.NCurrent,
			TotalCurrent:     status.TotalCurrent,
			TotalActivePower: status.TotalActivePower,
			TotalAprtPower:   status.TotalApparentPower,
			Errors:           status.Errors,
		}
		return nil
	})
	return result, err
}

// GetEM1Status returns the status of a single-phase Energy Monitor (EM1) component.
func (s *Service) GetEM1Status(ctx context.Context, device string, id int) (*EM1Status, error) {
	var result *EM1Status
	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		em1 := components.NewEM1(conn.RPCClient(), id)
		status, err := em1.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &EM1Status{
			ID:        status.ID,
			Voltage:   status.Voltage,
			Current:   status.Current,
			ActPower:  status.ActPower,
			AprtPower: status.AprtPower,
			PF:        status.PF,
			Freq:      status.Freq,
			Errors:    status.Errors,
		}
		return nil
	})
	return result, err
}

// GetPMStatus returns the status of a Power Meter (PM) component.
func (s *Service) GetPMStatus(ctx context.Context, device string, id int) (*PMStatus, error) {
	var result *PMStatus
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
func (s *Service) GetPM1Status(ctx context.Context, device string, id int) (*PMStatus, error) {
	var result *PMStatus
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

func pmStatusFromComponent(status *components.PMStatus) *PMStatus {
	result := &PMStatus{
		ID:      status.ID,
		Voltage: status.Voltage,
		Current: status.Current,
		APower:  status.APower,
		Freq:    status.Freq,
		Errors:  status.Errors,
	}
	if status.AEnergy != nil {
		result.AEnergy = &EnergyCounters{
			Total:    status.AEnergy.Total,
			ByMinute: status.AEnergy.ByMinute,
			MinuteTs: status.AEnergy.MinuteTs,
		}
	}
	if status.RetAEnergy != nil {
		result.RetAEnergy = &EnergyCounters{
			Total:    status.RetAEnergy.Total,
			ByMinute: status.RetAEnergy.ByMinute,
			MinuteTs: status.RetAEnergy.MinuteTs,
		}
	}
	return result
}

func pm1StatusFromComponent(status *components.PM1Status) *PMStatus {
	result := &PMStatus{
		ID:      status.ID,
		Voltage: status.Voltage,
		Current: status.Current,
		APower:  status.APower,
		Freq:    status.Freq,
		Errors:  status.Errors,
	}
	if status.AEnergy != nil {
		result.AEnergy = &EnergyCounters{
			Total:    status.AEnergy.Total,
			ByMinute: status.AEnergy.ByMinute,
			MinuteTs: status.AEnergy.MinuteTs,
		}
	}
	if status.RetAEnergy != nil {
		result.RetAEnergy = &EnergyCounters{
			Total:    status.RetAEnergy.Total,
			ByMinute: status.RetAEnergy.ByMinute,
			MinuteTs: status.RetAEnergy.MinuteTs,
		}
	}
	return result
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

// maxComponentID is the maximum component ID to probe when discovering components.
const maxComponentID = 10

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

// MonitorCallback is called with each status update during monitoring.
type MonitorCallback func(MonitoringSnapshot) error

// MonitoringSnapshot represents a complete device monitoring snapshot.
type MonitoringSnapshot struct {
	Device    string      `json:"device"`
	Timestamp time.Time   `json:"timestamp"`
	EM        []EMStatus  `json:"em,omitempty"`
	EM1       []EM1Status `json:"em1,omitempty"`
	PM        []PMStatus  `json:"pm,omitempty"`
	Online    bool        `json:"online"`
	Error     string      `json:"error,omitempty"`
}

// MonitorDevice continuously monitors a device and calls the callback with updates.
// Runs until the context is cancelled or Count updates are received.
func (s *Service) MonitorDevice(ctx context.Context, device string, opts MonitorOptions, callback MonitorCallback) error {
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

func (s *Service) collectDeviceStatus(ctx context.Context, device string, opts MonitorOptions) MonitoringSnapshot {
	status := MonitoringSnapshot{
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
// This is a convenience method that collects all available energy and power data.
func (s *Service) GetMonitoringSnapshot(ctx context.Context, device string) (*MonitoringSnapshot, error) {
	opts := MonitorOptions{
		IncludeEnergy: true,
		IncludePower:  true,
	}
	status := s.collectDeviceStatus(ctx, device, opts)
	return &status, nil
}

func (s *Service) collectEMStatus(ctx context.Context, device string) []EMStatus {
	var result []EMStatus
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

func (s *Service) collectEM1Status(ctx context.Context, device string) []EM1Status {
	var result []EM1Status
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

func (s *Service) collectPMStatus(ctx context.Context, device string) []PMStatus {
	var result []PMStatus
	// Collect PM components
	pmIDs, err := s.ListPMComponents(ctx, device)
	if err == nil {
		for _, id := range pmIDs {
			if pmStatus, err := s.GetPMStatus(ctx, device, id); err == nil {
				result = append(result, *pmStatus)
			}
		}
	}
	// Collect PM1 components
	pm1IDs, err := s.ListPM1Components(ctx, device)
	if err == nil {
		for _, id := range pm1IDs {
			if pm1Status, err := s.GetPM1Status(ctx, device, id); err == nil {
				result = append(result, *pm1Status)
			}
		}
	}
	return result
}

// EventHandler handles device events received via WebSocket.
type EventHandler func(DeviceEvent) error

// SubscribeEvents subscribes to real-time events from a device via WebSocket.
// Runs until the context is cancelled.
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

	// Subscribe to notifications
	notifyHandler := func(msg json.RawMessage) {
		var notif struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params,omitempty"`
		}
		if err := json.Unmarshal(msg, &notif); err != nil {
			return
		}

		event := parseNotification(device, notif.Method, notif.Params)

		// Handler errors are logged but don't stop event streaming
		if err := handler(event); err != nil {
			iostreams.DebugErr("event handler error", err)
		}
	}

	if err := ws.Subscribe(notifyHandler); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

func closeWebSocket(ws *transport.WebSocket) {
	if err := ws.Close(); err != nil {
		iostreams.DebugErr("closing websocket", err)
	}
}

// parseNotification converts a raw WebSocket notification into a DeviceEvent.
func parseNotification(device, method string, params map[string]any) DeviceEvent {
	event := DeviceEvent{
		Device:    device,
		Timestamp: time.Now(),
		Event:     method,
		Data:      params,
	}

	// Try to extract component info from params
	// Shelly notifications often have a format like "switch:0" in the source
	if src, ok := params["src"].(string); ok {
		event.Component, event.ComponentID = parseComponentSource(src)
	}

	// Or component info might be in the method name
	if event.Component == "" {
		event.Component, event.ComponentID = parseComponentSource(method)
	}

	return event
}

// parseComponentSource extracts component type and ID from a source string.
// Examples: "switch:0" -> ("switch", 0), "cover:1" -> ("cover", 1).
func parseComponentSource(src string) (component string, id int) {
	_, err := fmt.Sscanf(src, "%[^:]:%d", &component, &id)
	if err != nil {
		return src, 0
	}
	return component, id
}

// PrometheusMetrics represents metrics in Prometheus exposition format.
type PrometheusMetrics struct {
	Metrics []PrometheusMetric `json:"metrics"`
}

// PrometheusMetric represents a single Prometheus metric.
type PrometheusMetric struct {
	Name   string            `json:"name"`
	Help   string            `json:"help"`
	Type   string            `json:"type"` // gauge, counter
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
}

// CollectPrometheusMetrics collects metrics from a device in Prometheus format.
func (s *Service) CollectPrometheusMetrics(ctx context.Context, device string) (*PrometheusMetrics, error) {
	metrics := &PrometheusMetrics{}

	// Collect PM1 metrics
	pm1IDs, err := s.ListPM1Components(ctx, device)
	if err == nil {
		for _, id := range pm1IDs {
			status, err := s.GetPM1Status(ctx, device, id)
			if err != nil {
				continue
			}

			labels := map[string]string{
				"device":       device,
				"component":    "pm1",
				"component_id": fmt.Sprintf("%d", id),
			}

			metrics.Metrics = append(metrics.Metrics,
				PrometheusMetric{
					Name:   "shelly_power_watts",
					Help:   "Current power consumption in watts",
					Type:   "gauge",
					Labels: labels,
					Value:  status.APower,
				},
				PrometheusMetric{
					Name:   "shelly_voltage_volts",
					Help:   "Current voltage in volts",
					Type:   "gauge",
					Labels: labels,
					Value:  status.Voltage,
				},
				PrometheusMetric{
					Name:   "shelly_current_amps",
					Help:   "Current in amps",
					Type:   "gauge",
					Labels: labels,
					Value:  status.Current,
				},
			)

			if status.AEnergy != nil {
				metrics.Metrics = append(metrics.Metrics, PrometheusMetric{
					Name:   "shelly_energy_wh_total",
					Help:   "Total energy consumption in watt-hours",
					Type:   "counter",
					Labels: labels,
					Value:  status.AEnergy.Total,
				})
			}

			if status.Freq != nil {
				metrics.Metrics = append(metrics.Metrics, PrometheusMetric{
					Name:   "shelly_frequency_hz",
					Help:   "AC frequency in hertz",
					Type:   "gauge",
					Labels: labels,
					Value:  *status.Freq,
				})
			}
		}
	}

	// Collect EM metrics
	emIDs, err := s.ListEMComponents(ctx, device)
	if err == nil {
		for _, id := range emIDs {
			status, err := s.GetEMStatus(ctx, device, id)
			if err != nil {
				continue
			}

			baseLabels := map[string]string{
				"device":       device,
				"component":    "em",
				"component_id": fmt.Sprintf("%d", id),
			}

			// Phase A
			labelsA := copyLabels(baseLabels)
			labelsA["phase"] = "a"
			metrics.Metrics = append(metrics.Metrics,
				PrometheusMetric{
					Name:   "shelly_power_watts",
					Help:   "Current power consumption in watts",
					Type:   "gauge",
					Labels: labelsA,
					Value:  status.AActivePower,
				},
				PrometheusMetric{
					Name:   "shelly_voltage_volts",
					Help:   "Current voltage in volts",
					Type:   "gauge",
					Labels: labelsA,
					Value:  status.AVoltage,
				},
				PrometheusMetric{
					Name:   "shelly_current_amps",
					Help:   "Current in amps",
					Type:   "gauge",
					Labels: labelsA,
					Value:  status.ACurrent,
				},
			)

			// Phase B
			labelsB := copyLabels(baseLabels)
			labelsB["phase"] = "b"
			metrics.Metrics = append(metrics.Metrics,
				PrometheusMetric{
					Name:   "shelly_power_watts",
					Help:   "Current power consumption in watts",
					Type:   "gauge",
					Labels: labelsB,
					Value:  status.BActivePower,
				},
				PrometheusMetric{
					Name:   "shelly_voltage_volts",
					Help:   "Current voltage in volts",
					Type:   "gauge",
					Labels: labelsB,
					Value:  status.BVoltage,
				},
				PrometheusMetric{
					Name:   "shelly_current_amps",
					Help:   "Current in amps",
					Type:   "gauge",
					Labels: labelsB,
					Value:  status.BCurrent,
				},
			)

			// Phase C
			labelsC := copyLabels(baseLabels)
			labelsC["phase"] = "c"
			metrics.Metrics = append(metrics.Metrics,
				PrometheusMetric{
					Name:   "shelly_power_watts",
					Help:   "Current power consumption in watts",
					Type:   "gauge",
					Labels: labelsC,
					Value:  status.CActivePower,
				},
				PrometheusMetric{
					Name:   "shelly_voltage_volts",
					Help:   "Current voltage in volts",
					Type:   "gauge",
					Labels: labelsC,
					Value:  status.CVoltage,
				},
				PrometheusMetric{
					Name:   "shelly_current_amps",
					Help:   "Current in amps",
					Type:   "gauge",
					Labels: labelsC,
					Value:  status.CCurrent,
				},
			)

			// Total
			labelsTotal := copyLabels(baseLabels)
			labelsTotal["phase"] = "total"
			metrics.Metrics = append(metrics.Metrics,
				PrometheusMetric{
					Name:   "shelly_power_watts",
					Help:   "Current power consumption in watts",
					Type:   "gauge",
					Labels: labelsTotal,
					Value:  status.TotalActivePower,
				},
				PrometheusMetric{
					Name:   "shelly_current_amps",
					Help:   "Current in amps",
					Type:   "gauge",
					Labels: labelsTotal,
					Value:  status.TotalCurrent,
				},
			)
		}
	}

	return metrics, nil
}

func copyLabels(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// FormatPrometheusMetrics formats metrics as Prometheus exposition format.
func FormatPrometheusMetrics(metrics *PrometheusMetrics) string {
	var result string
	seen := make(map[string]bool)

	for _, m := range metrics.Metrics {
		// Print HELP and TYPE only once per metric name
		if !seen[m.Name] {
			result += fmt.Sprintf("# HELP %s %s\n", m.Name, m.Help)
			result += fmt.Sprintf("# TYPE %s %s\n", m.Name, m.Type)
			seen[m.Name] = true
		}

		// Format labels
		labels := ""
		if len(m.Labels) > 0 {
			first := true
			labels = "{"
			for k, v := range m.Labels {
				if !first {
					labels += ","
				}
				labels += fmt.Sprintf("%s=%q", k, v)
				first = false
			}
			labels += "}"
		}

		result += fmt.Sprintf("%s%s %g\n", m.Name, labels, m.Value)
	}

	return result
}

// InfluxDBPoint represents a single InfluxDB line protocol point.
type InfluxDBPoint struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]float64
	Timestamp   time.Time
}

// CollectInfluxDBPoints collects metrics from a device in InfluxDB line protocol format.
func (s *Service) CollectInfluxDBPoints(ctx context.Context, device string) ([]InfluxDBPoint, error) {
	var points []InfluxDBPoint
	now := time.Now()

	// Collect PM1 metrics
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

			points = append(points, InfluxDBPoint{
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

	// Collect EM metrics
	emIDs, err := s.ListEMComponents(ctx, device)
	if err == nil {
		for _, id := range emIDs {
			status, err := s.GetEMStatus(ctx, device, id)
			if err != nil {
				continue
			}

			compID := fmt.Sprintf("%d", id)
			points = append(points,
				// Phase A
				InfluxDBPoint{
					Measurement: "shelly_power",
					Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "a"},
					Fields:      map[string]float64{"power": status.AActivePower, "voltage": status.AVoltage, "current": status.ACurrent},
					Timestamp:   now,
				},
				// Phase B
				InfluxDBPoint{
					Measurement: "shelly_power",
					Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "b"},
					Fields:      map[string]float64{"power": status.BActivePower, "voltage": status.BVoltage, "current": status.BCurrent},
					Timestamp:   now,
				},
				// Phase C
				InfluxDBPoint{
					Measurement: "shelly_power",
					Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "c"},
					Fields:      map[string]float64{"power": status.CActivePower, "voltage": status.CVoltage, "current": status.CCurrent},
					Timestamp:   now,
				},
				// Total
				InfluxDBPoint{
					Measurement: "shelly_power",
					Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "total"},
					Fields:      map[string]float64{"power": status.TotalActivePower, "current": status.TotalCurrent},
					Timestamp:   now,
				},
			)
		}
	}

	return points, nil
}

// FormatInfluxDBLineProtocol formats points as InfluxDB line protocol.
func FormatInfluxDBLineProtocol(points []InfluxDBPoint) string {
	var result string
	for _, p := range points {
		// Measurement
		line := p.Measurement

		// Tags
		for k, v := range p.Tags {
			line += fmt.Sprintf(",%s=%s", k, v)
		}

		// Fields
		line += " "
		first := true
		for k, v := range p.Fields {
			if !first {
				line += ","
			}
			line += fmt.Sprintf("%s=%g", k, v)
			first = false
		}

		// Timestamp (nanoseconds)
		line += fmt.Sprintf(" %d\n", p.Timestamp.UnixNano())

		result += line
	}
	return result
}
