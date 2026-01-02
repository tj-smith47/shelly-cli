// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/shelly/monitoring"
)

// MonitoringOptions is an alias for monitoring.Options.
type MonitoringOptions = monitoring.Options

// MonitoringCallback is an alias for monitoring.Callback.
type MonitoringCallback = monitoring.Callback

// EventHandler is an alias for monitoring.EventHandler.
type EventHandler = monitoring.EventHandler

// DeviceSnapshot is an alias for monitoring.DeviceSnapshot.
type DeviceSnapshot = monitoring.DeviceSnapshot

// MonitoringDeviceInfo is an alias for monitoring.DeviceInfo.
type MonitoringDeviceInfo = monitoring.DeviceInfo

// MonitoringResolvedDevice is an alias for monitoring.ResolvedDevice.
type MonitoringResolvedDevice = monitoring.ResolvedDevice

// MonitoringDeviceStatus is an alias for monitoring.DeviceStatusResult.
type MonitoringDeviceStatus = monitoring.DeviceStatusResult

// Delegation methods - these delegate to the monitoring subpackage.
// This maintains backward compatibility for existing callers.

// GetEMStatus returns the status of an Energy Monitor (EM) component.
func (s *Service) GetEMStatus(ctx context.Context, device string, id int) (*model.EMStatus, error) {
	return s.Monitoring().GetEMStatus(ctx, device, id)
}

// GetEM1Status returns the status of a single-phase Energy Monitor (EM1) component.
func (s *Service) GetEM1Status(ctx context.Context, device string, id int) (*model.EM1Status, error) {
	return s.Monitoring().GetEM1Status(ctx, device, id)
}

// GetPMStatus returns the status of a Power Meter (PM) component.
func (s *Service) GetPMStatus(ctx context.Context, device string, id int) (*model.PMStatus, error) {
	return s.Monitoring().GetPMStatus(ctx, device, id)
}

// GetPM1Status returns the status of a Power Meter (PM1) component.
func (s *Service) GetPM1Status(ctx context.Context, device string, id int) (*model.PMStatus, error) {
	return s.Monitoring().GetPM1Status(ctx, device, id)
}

// ResetEMCounters resets energy counters on an EM component.
func (s *Service) ResetEMCounters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.Monitoring().ResetEMCounters(ctx, device, id, counterTypes)
}

// ResetPMCounters resets energy counters on a PM component.
func (s *Service) ResetPMCounters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.Monitoring().ResetPMCounters(ctx, device, id, counterTypes)
}

// ResetPM1Counters resets energy counters on a PM1 component.
func (s *Service) ResetPM1Counters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.Monitoring().ResetPM1Counters(ctx, device, id, counterTypes)
}

// ListEMComponents returns a list of EM component IDs on a device.
func (s *Service) ListEMComponents(ctx context.Context, device string) ([]int, error) {
	return s.Monitoring().ListEMComponents(ctx, device)
}

// ListEM1Components returns a list of EM1 component IDs on a device.
func (s *Service) ListEM1Components(ctx context.Context, device string) ([]int, error) {
	return s.Monitoring().ListEM1Components(ctx, device)
}

// ListPMComponents returns a list of PM component IDs on a device.
func (s *Service) ListPMComponents(ctx context.Context, device string) ([]int, error) {
	return s.Monitoring().ListPMComponents(ctx, device)
}

// ListPM1Components returns a list of PM1 component IDs on a device.
func (s *Service) ListPM1Components(ctx context.Context, device string) ([]int, error) {
	return s.Monitoring().ListPM1Components(ctx, device)
}

// MonitorDevice continuously monitors a device and calls the callback with updates.
func (s *Service) MonitorDevice(ctx context.Context, device string, opts MonitoringOptions, callback MonitoringCallback) error {
	return s.Monitoring().MonitorDevice(ctx, device, opts, callback)
}

// GetMonitoringSnapshot returns a single snapshot of all monitoring data for a device.
func (s *Service) GetMonitoringSnapshot(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	return s.Monitoring().GetMonitoringSnapshot(ctx, device)
}

// GetMonitoringSnapshotAuto returns monitoring data for a device, auto-detecting generation.
func (s *Service) GetMonitoringSnapshotAuto(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	return s.Monitoring().GetMonitoringSnapshotAuto(ctx, device)
}

// GetGen1StatusJSON returns Gen1 device status as JSON for event streaming.
func (s *Service) GetGen1StatusJSON(ctx context.Context, identifier string) (json.RawMessage, error) {
	return s.Monitoring().GetGen1StatusJSON(ctx, identifier)
}

// FetchAllSnapshots fetches device info and monitoring snapshots for all devices concurrently.
func (s *Service) FetchAllSnapshots(ctx context.Context, devices map[string]string, snapshots map[string]*DeviceSnapshot, mu *sync.Mutex) {
	s.Monitoring().FetchAllSnapshots(ctx, devices, snapshots, mu)
}

// SubscribeEvents subscribes to real-time events from a device via WebSocket.
func (s *Service) SubscribeEvents(ctx context.Context, device string, handler EventHandler) error {
	return s.Monitoring().SubscribeEvents(ctx, device, handler)
}

// CollectPrometheusMetrics collects metrics from a device in Prometheus format.
func (s *Service) CollectPrometheusMetrics(ctx context.Context, device string) (*export.PrometheusMetrics, error) {
	return s.Monitoring().CollectPrometheusMetrics(ctx, device)
}

// CollectComponentReadings collects all meter readings from a device.
func (s *Service) CollectComponentReadings(ctx context.Context, device string) []model.ComponentReading {
	return s.Monitoring().CollectComponentReadings(ctx, device)
}

// CollectInfluxDBPoints collects metrics from a device in InfluxDB line protocol format.
func (s *Service) CollectInfluxDBPoints(ctx context.Context, device string) ([]export.InfluxDBPoint, error) {
	return s.Monitoring().CollectInfluxDBPoints(ctx, device)
}

// GetEMDataRecords retrieves available time intervals with stored EMData.
func (s *Service) GetEMDataRecords(ctx context.Context, device string, id int, fromTS *int64) (*components.EMDataRecordsResult, error) {
	return s.Monitoring().GetEMDataRecords(ctx, device, id, fromTS)
}

// GetEMDataHistory retrieves historical EMData measurements for a time range.
func (s *Service) GetEMDataHistory(ctx context.Context, device string, id int, startTS, endTS *int64) (*components.EMDataGetDataResult, error) {
	return s.Monitoring().GetEMDataHistory(ctx, device, id, startTS, endTS)
}

// DeleteEMData deletes all stored historical EMData.
func (s *Service) DeleteEMData(ctx context.Context, device string, id int) error {
	return s.Monitoring().DeleteEMData(ctx, device, id)
}

// GetEMDataCSVURL returns the HTTP URL for downloading EMData as CSV.
func (s *Service) GetEMDataCSVURL(device string, id int, startTS, endTS *int64, addKeys bool) (string, error) {
	return s.Monitoring().GetEMDataCSVURL(device, id, startTS, endTS, addKeys)
}

// GetEM1DataRecords retrieves available time intervals with stored EM1Data.
func (s *Service) GetEM1DataRecords(ctx context.Context, device string, id int, fromTS *int64) (*components.EM1DataRecordsResult, error) {
	return s.Monitoring().GetEM1DataRecords(ctx, device, id, fromTS)
}

// GetEM1DataHistory retrieves historical EM1Data measurements for a time range.
func (s *Service) GetEM1DataHistory(ctx context.Context, device string, id int, startTS, endTS *int64) (*components.EM1DataGetDataResult, error) {
	return s.Monitoring().GetEM1DataHistory(ctx, device, id, startTS, endTS)
}

// DeleteEM1Data deletes all stored historical EM1Data.
func (s *Service) DeleteEM1Data(ctx context.Context, device string, id int) error {
	return s.Monitoring().DeleteEM1Data(ctx, device, id)
}

// GetEM1DataCSVURL returns the HTTP URL for downloading EM1Data as CSV.
func (s *Service) GetEM1DataCSVURL(device string, id int, startTS, endTS *int64, addKeys bool) (string, error) {
	return s.Monitoring().GetEM1DataCSVURL(device, id, startTS, endTS, addKeys)
}

// CollectJSONMetrics collects metrics from multiple devices for JSON output.
func (s *Service) CollectJSONMetrics(ctx context.Context, devices []string) export.JSONMetricsOutput {
	return s.Monitoring().CollectJSONMetrics(ctx, devices)
}

// StreamInfluxDBPoints continuously collects and outputs InfluxDB points at the given interval.
func (s *Service) StreamInfluxDBPoints(ctx context.Context, devices []string, measurement string, tags map[string]string, interval time.Duration, writePoints func([]export.InfluxDBPoint)) error {
	return s.Monitoring().StreamInfluxDBPoints(ctx, devices, measurement, tags, interval, writePoints)
}

// CollectInfluxDBPointsMulti collects InfluxDB points from multiple devices concurrently.
func (s *Service) CollectInfluxDBPointsMulti(ctx context.Context, devices []string, measurement string, tags map[string]string) []export.InfluxDBPoint {
	return s.Monitoring().CollectInfluxDBPointsMulti(ctx, devices, measurement, tags)
}

// CollectDashboardData collects energy data from multiple devices concurrently.
func (s *Service) CollectDashboardData(ctx context.Context, ios *iostreams.IOStreams, devices []string) model.DashboardData {
	return s.Monitoring().CollectDashboardData(ctx, ios, devices)
}

// CollectComparisonData collects energy comparison data from multiple devices.
func (s *Service) CollectComparisonData(ctx context.Context, ios *iostreams.IOStreams, devices []string, period string, startTS, endTS *int64) model.ComparisonData {
	return s.Monitoring().CollectComparisonData(ctx, ios, devices, period, startTS, endTS)
}

// NewPrometheusCollector creates a new Prometheus metrics collector.
// This is a package-level factory function for backward compatibility.
func NewPrometheusCollector(svc *Service, devices []string) *monitoring.PrometheusCollector {
	return monitoring.NewPrometheusCollector(svc.Monitoring(), devices)
}
