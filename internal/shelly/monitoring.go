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
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
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

// ResetPowerCounters resets the energy counters of one component that meters
// power. See monitoring.Service.ResetPowerCounters.
func (s *Service) ResetPowerCounters(ctx context.Context, device, typ string, id int, counterTypes []string) (model.PowerReading, error) {
	return s.Monitoring().ResetPowerCounters(ctx, device, typ, id, counterTypes)
}

// ReadPowerReadings returns every power reading the device reports.
// See monitoring.Service.ReadPowerReadings.
func (s *Service) ReadPowerReadings(ctx context.Context, device string) ([]model.PowerReading, error) {
	return s.Monitoring().ReadPowerReadings(ctx, device)
}

// ReadPowerReading returns one power reading of the device.
// See monitoring.Service.ReadPowerReading.
func (s *Service) ReadPowerReading(ctx context.Context, device, typ string, id int) (model.PowerReading, error) {
	return s.Monitoring().ReadPowerReading(ctx, device, typ, id)
}

// CollectPowerReadings reads every power reading on each device.
// See monitoring.Service.CollectPowerReadings.
func (s *Service) CollectPowerReadings(ctx context.Context, devices []string) ([]model.PowerReading, []model.EnergyStatusSkip) {
	return s.Monitoring().CollectPowerReadings(ctx, devices)
}

// MonitorDevice continuously monitors a device and calls the callback with updates.
func (s *Service) MonitorDevice(ctx context.Context, device string, opts MonitoringOptions, callback MonitoringCallback) error {
	return s.Monitoring().MonitorDevice(ctx, device, opts, callback)
}

// GetMonitoringSnapshot returns a single snapshot of all monitoring data for a device.
func (s *Service) GetMonitoringSnapshot(ctx context.Context, device string) (*model.MonitoringSnapshot, error) {
	return s.Monitoring().GetMonitoringSnapshot(ctx, device)
}

// GetGen1StatusJSON returns Gen1 device status as JSON for event streaming.
func (s *Service) GetGen1StatusJSON(ctx context.Context, identifier string) (json.RawMessage, error) {
	return s.Monitoring().GetGen1StatusJSON(ctx, identifier)
}

// Gen1DeviceInfo is an alias for automation.Gen1DeviceInfo.
type Gen1DeviceInfo = automation.Gen1DeviceInfo

// GetGen1DeviceInfo returns the Gen1 device type and MAC for CoIoT registration.
func (s *Service) GetGen1DeviceInfo(ctx context.Context, identifier string) (*Gen1DeviceInfo, error) {
	info, err := s.Monitoring().GetGen1DeviceInfo(ctx, identifier)
	if err != nil || info == nil {
		return nil, err
	}
	return &Gen1DeviceInfo{
		Type: info.Type,
		MAC:  info.MAC,
	}, nil
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
func NewPrometheusCollector(svc *Service, devices []string) *monitoring.PrometheusCollector {
	return monitoring.NewPrometheusCollector(svc.Monitoring(), devices)
}
