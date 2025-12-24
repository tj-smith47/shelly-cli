// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Energy component type constants for auto-detection.
const (
	ComponentTypeAuto = "auto"
	ComponentTypeEM   = "em"
	ComponentTypeEM1  = "em1"
)

// DetectEnergyComponentByID auto-detects the energy component type by checking
// which component list contains the given ID. Returns ComponentTypeAuto if no match found.
// If detection fails, a warning is logged via ios.
func (s *Service) DetectEnergyComponentByID(ctx context.Context, ios *iostreams.IOStreams, device string, id int) string {
	emIDs, emErr := s.ListEMComponents(ctx, device)
	if emErr != nil {
		ios.DebugErr("list EM components", emErr)
	}
	em1IDs, em1Err := s.ListEM1Components(ctx, device)
	if em1Err != nil {
		ios.DebugErr("list EM1 components", em1Err)
	}

	// Check if ID matches EM component
	for _, emID := range emIDs {
		if emID == id {
			return ComponentTypeEM
		}
	}

	// Check if ID matches EM1 component
	for _, em1ID := range em1IDs {
		if em1ID == id {
			return ComponentTypeEM1
		}
	}

	// Default to first available type
	if len(emIDs) > 0 {
		return ComponentTypeEM
	}
	if len(em1IDs) > 0 {
		return ComponentTypeEM1
	}

	// Detection failed - warn if both list operations returned errors
	if emErr != nil && em1Err != nil {
		ios.Warning("Could not detect energy component type: device may be offline or have no energy monitoring")
	}

	return ComponentTypeAuto
}

// DetectEnergyComponentType auto-detects whether a device uses EM or EM1 data components.
// It probes the device for EMData and EM1Data records and returns the appropriate type.
// Returns an error if no energy data components are found.
func (s *Service) DetectEnergyComponentType(ctx context.Context, ios *iostreams.IOStreams, device string, id int) (string, error) {
	// Try EMData first
	emRecords, err := s.GetEMDataRecords(ctx, device, id, nil)
	if err == nil && emRecords != nil && len(emRecords.Records) > 0 {
		return ComponentTypeEM, nil
	}
	ios.DebugErr("get EMData records", err)

	// Try EM1Data
	em1Records, err := s.GetEM1DataRecords(ctx, device, id, nil)
	if err == nil && em1Records != nil && len(em1Records.Records) > 0 {
		return ComponentTypeEM1, nil
	}
	ios.DebugErr("get EM1Data records", err)

	return "", fmt.Errorf("no energy data components found")
}

// CalculateTimeRange converts period/from/to flags to Unix timestamps.
// It supports predefined periods (hour, day, week, month) or explicit from/to times.
// Returns nil pointers if no time range is specified (empty period and no from/to).
func CalculateTimeRange(period, from, to string) (startTS, endTS *int64, err error) {
	// If explicit from/to provided, use those
	if from != "" || to != "" {
		return parseExplicitTimeRange(from, to)
	}

	// Calculate based on period
	now := time.Now()
	var start time.Time

	switch period {
	case "hour":
		start = now.Add(-1 * time.Hour)
	case "day", "":
		start = now.Add(-24 * time.Hour)
	case "week":
		start = now.Add(-7 * 24 * time.Hour)
	case "month":
		start = now.Add(-30 * 24 * time.Hour)
	default:
		return nil, nil, fmt.Errorf("invalid period: %s (use: hour, day, week, month)", period)
	}

	startUnix := start.Unix()
	endUnix := now.Unix()
	return &startUnix, &endUnix, nil
}

// parseExplicitTimeRange parses explicit from/to time strings into Unix timestamps.
func parseExplicitTimeRange(from, to string) (startTS, endTS *int64, err error) {
	if from != "" {
		t, err := ParseTime(from)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --from time: %w", err)
		}
		ts := t.Unix()
		startTS = &ts
	}
	if to != "" {
		t, err := ParseTime(to)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --to time: %w", err)
		}
		ts := t.Unix()
		endTS = &ts
	}
	return startTS, endTS, nil
}

// ParseTime parses a time string in various formats.
// Supported formats: RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS.
func ParseTime(s string) (time.Time, error) {
	formats := []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time (use RFC3339, YYYY-MM-DD, or 'YYYY-MM-DD HH:MM:SS')")
}

// energyDataBlock represents a block of energy data with period and power values.
type energyDataBlock interface {
	GetPeriod() int
	GetPowerValues() []float64
}

type emDataBlockAdapter struct{ b components.EMDataBlock }

func (a emDataBlockAdapter) GetPeriod() int { return a.b.Period }
func (a emDataBlockAdapter) GetPowerValues() []float64 {
	powers := make([]float64, len(a.b.Values))
	for i, v := range a.b.Values {
		powers[i] = v.TotalActivePower
	}
	return powers
}

type em1DataBlockAdapter struct{ b components.EM1DataBlock }

func (a em1DataBlockAdapter) GetPeriod() int { return a.b.Period }
func (a em1DataBlockAdapter) GetPowerValues() []float64 {
	powers := make([]float64, len(a.b.Values))
	for i, v := range a.b.Values {
		powers[i] = v.ActivePower
	}
	return powers
}

// calculateMetrics is the generic implementation for energy metrics calculation.
func calculateMetrics(blocks []energyDataBlock) (energy, avgPower, peakPower float64, dataPoints int) {
	var totalPower float64
	for _, block := range blocks {
		for _, power := range block.GetPowerValues() {
			totalPower += power
			if power > peakPower {
				peakPower = power
			}
			energy += power * float64(block.GetPeriod()) / 3600.0
			dataPoints++
		}
	}
	if dataPoints > 0 {
		avgPower = totalPower / float64(dataPoints)
	}
	energy /= 1000.0 // Wh to kWh
	return
}

// CalculateEMMetrics calculates energy metrics from EM data history.
func CalculateEMMetrics(data *components.EMDataGetDataResult) (energy, avgPower, peakPower float64, dataPoints int) {
	blocks := make([]energyDataBlock, len(data.Data))
	for i, b := range data.Data {
		blocks[i] = emDataBlockAdapter{b}
	}
	return calculateMetrics(blocks)
}

// CalculateEM1Metrics calculates energy metrics from EM1 data history.
func CalculateEM1Metrics(data *components.EM1DataGetDataResult) (energy, avgPower, peakPower float64, dataPoints int) {
	blocks := make([]energyDataBlock, len(data.Data))
	for i, b := range data.Data {
		blocks[i] = em1DataBlockAdapter{b}
	}
	return calculateMetrics(blocks)
}

// componentCollector defines how to collect power data from a component type.
type componentCollector[T any] struct {
	compType  string
	listIDs   func(ctx context.Context, device string) ([]int, error)
	getStatus func(ctx context.Context, device string, id int) (T, error)
	toPower   func(status T, id int) (model.ComponentPower, float64, float64) // returns comp, power, energy
}

// collectComponents is a generic helper for collecting component power data.
func collectComponents[T any](ctx context.Context, device string, c componentCollector[T], status *model.DashboardDeviceEntry) {
	ids, err := c.listIDs(ctx, device)
	if err != nil {
		return
	}
	for _, id := range ids {
		compStatus, err := c.getStatus(ctx, device, id)
		if err != nil {
			continue
		}
		comp, power, energy := c.toPower(compStatus, id)
		comp.Type = c.compType
		comp.ID = id
		status.Components = append(status.Components, comp)
		status.TotalPower += power
		status.TotalEnergy += energy
	}
}

// CollectDashboardData collects energy data from multiple devices concurrently.
func (s *Service) CollectDashboardData(ctx context.Context, ios *iostreams.IOStreams, devices []string) model.DashboardData {
	dashboard := model.DashboardData{
		Timestamp:   time.Now(),
		DeviceCount: len(devices),
		Devices:     make([]model.DashboardDeviceEntry, len(devices)),
	}

	g, ctx := errgroup.WithContext(ctx)
	// Use global rate limit for concurrency (service layer also enforces this)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for i, device := range devices {
		idx, dev := i, device
		g.Go(func() error {
			dashboard.Devices[idx] = s.collectDashboardDeviceStatus(ctx, dev)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("collecting dashboard data", err)
	}

	// Aggregate totals
	for _, dev := range dashboard.Devices {
		if dev.Online {
			dashboard.OnlineCount++
			dashboard.TotalPower += dev.TotalPower
			dashboard.TotalEnergy += dev.TotalEnergy
		} else {
			dashboard.OfflineCount++
		}
	}

	return dashboard
}

func (s *Service) collectDashboardDeviceStatus(ctx context.Context, device string) model.DashboardDeviceEntry {
	status := model.DashboardDeviceEntry{Device: device, Online: true}

	// Collect each component type using the generic collector
	collectComponents(ctx, device, componentCollector[*EMStatus]{
		compType: "EM", listIDs: s.ListEMComponents, getStatus: s.GetEMStatus,
		toPower: func(st *EMStatus, id int) (model.ComponentPower, float64, float64) {
			return model.ComponentPower{Voltage: st.AVoltage, Current: st.TotalCurrent, Power: st.TotalActivePower}, st.TotalActivePower, 0
		},
	}, &status)

	collectComponents(ctx, device, componentCollector[*EM1Status]{
		compType: "EM1", listIDs: s.ListEM1Components, getStatus: s.GetEM1Status,
		toPower: func(st *EM1Status, id int) (model.ComponentPower, float64, float64) {
			return model.ComponentPower{Voltage: st.Voltage, Current: st.Current, Power: st.ActPower}, st.ActPower, 0
		},
	}, &status)

	collectComponents(ctx, device, componentCollector[*PMStatus]{
		compType: "PM", listIDs: s.ListPMComponents, getStatus: s.GetPMStatus,
		toPower: func(st *PMStatus, id int) (model.ComponentPower, float64, float64) {
			energy := 0.0
			if st.AEnergy != nil {
				energy = st.AEnergy.Total
			}
			return model.ComponentPower{Voltage: st.Voltage, Current: st.Current, Power: st.APower, Energy: energy}, st.APower, energy
		},
	}, &status)

	collectComponents(ctx, device, componentCollector[*PMStatus]{
		compType: "PM1", listIDs: s.ListPM1Components, getStatus: s.GetPM1Status,
		toPower: func(st *PMStatus, id int) (model.ComponentPower, float64, float64) {
			energy := 0.0
			if st.AEnergy != nil {
				energy = st.AEnergy.Total
			}
			return model.ComponentPower{Voltage: st.Voltage, Current: st.Current, Power: st.APower, Energy: energy}, st.APower, energy
		},
	}, &status)

	// Mark offline if no components found
	if len(status.Components) == 0 {
		if _, pingErr := s.ListPMComponents(ctx, device); pingErr != nil {
			status.Online = false
			status.Error = "device unreachable"
		}
	}

	return status
}

// CollectComparisonData collects energy comparison data from multiple devices.
func (s *Service) CollectComparisonData(ctx context.Context, ios *iostreams.IOStreams, devices []string, period string, startTS, endTS *int64) model.ComparisonData {
	comparison := model.ComparisonData{
		Period:  period,
		Devices: make([]model.DeviceEnergy, len(devices)),
	}

	if startTS != nil {
		comparison.From = time.Unix(*startTS, 0)
	}
	if endTS != nil {
		comparison.To = time.Unix(*endTS, 0)
	}

	g, ctx := errgroup.WithContext(ctx)
	// Use global rate limit for concurrency (service layer also enforces this)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for i, device := range devices {
		idx, dev := i, device
		g.Go(func() error {
			comparison.Devices[idx] = s.collectDeviceEnergy(ctx, dev, startTS, endTS)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("collecting comparison data", err)
	}

	// Calculate totals and find min/max
	comparison.MinEnergy = -1
	for _, dev := range comparison.Devices {
		if dev.Online {
			comparison.TotalEnergy += dev.Energy
			if dev.Energy > comparison.MaxEnergy {
				comparison.MaxEnergy = dev.Energy
			}
			if comparison.MinEnergy < 0 || dev.Energy < comparison.MinEnergy {
				comparison.MinEnergy = dev.Energy
			}
		}
	}
	if comparison.MinEnergy < 0 {
		comparison.MinEnergy = 0
	}

	return comparison
}

func (s *Service) collectDeviceEnergy(ctx context.Context, device string, startTS, endTS *int64) model.DeviceEnergy {
	result := model.DeviceEnergy{Device: device, Online: true}

	// Try EM data first
	if emData, err := s.GetEMDataHistory(ctx, device, 0, startTS, endTS); err == nil && emData != nil && len(emData.Data) > 0 {
		result.Energy, result.AvgPower, result.PeakPower, result.DataPoints = CalculateEMMetrics(emData)
		return result
	}

	// Try EM1 data
	if em1Data, err := s.GetEM1DataHistory(ctx, device, 0, startTS, endTS); err == nil && em1Data != nil && len(em1Data.Data) > 0 {
		result.Energy, result.AvgPower, result.PeakPower, result.DataPoints = CalculateEM1Metrics(em1Data)
		return result
	}

	// Try current power from PM/PM1 for devices without history
	if power := s.collectCurrentPower(ctx, device); power > 0 {
		result.AvgPower, result.PeakPower, result.DataPoints = power, power, 1
		result.Error = "no historical data"
		return result
	}

	result.Online, result.Error = false, "no data available"
	return result
}

func (s *Service) collectCurrentPower(ctx context.Context, device string) float64 {
	var totalPower float64
	for _, list := range []func(context.Context, string) ([]int, error){s.ListPMComponents, s.ListPM1Components} {
		if ids, err := list(ctx, device); err == nil {
			for _, id := range ids {
				if pm, err := s.GetPMStatus(ctx, device, id); err == nil {
					totalPower += pm.APower
				} else if pm1, err := s.GetPM1Status(ctx, device, id); err == nil {
					totalPower += pm1.APower
				}
			}
		}
	}
	return totalPower
}
