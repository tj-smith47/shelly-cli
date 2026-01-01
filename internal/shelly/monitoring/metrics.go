package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// CollectPrometheusMetrics collects metrics from a device in Prometheus format.
func (s *Service) CollectPrometheusMetrics(ctx context.Context, device string) (*export.PrometheusMetrics, error) {
	metrics := &export.PrometheusMetrics{}
	online := true

	deviceStatus, err := s.connector.DeviceStatus(ctx, device)
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
