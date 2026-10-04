package monitoring

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

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
	readings, err := s.ReadPowerReadings(ctx, device)
	if err != nil {
		status.Online = false
		status.Error = "device unreachable"
		return status
	}
	for _, r := range readings {
		comp := model.ComponentPower{Type: r.Type, ID: r.ID, Power: r.Power}
		switch {
		case r.EM != nil:
			comp.Voltage, comp.Current = r.EM.AVoltage, r.EM.TotalCurrent
		case r.EM1 != nil:
			comp.Voltage, comp.Current = r.EM1.Voltage, r.EM1.Current
		case r.Meter != nil:
			comp.Voltage, comp.Current = r.Meter.Voltage, r.Meter.Current
			if r.Meter.AEnergy != nil {
				comp.Energy = r.Meter.AEnergy.Total
			}
		}
		status.Components = append(status.Components, comp)
		status.TotalPower += comp.Power
		status.TotalEnergy += comp.Energy
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

	var mu sync.Mutex

	for i, device := range devices {
		idx, dev := i, device
		g.Go(func() error {
			result := s.collectDeviceEnergy(ctx, dev, startTS, endTS)
			mu.Lock()
			comparison.Devices[idx] = result
			mu.Unlock()
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

	// A meter without stored history still gives the device's live power.
	readings, err := s.ReadPowerReadings(ctx, device)
	if err != nil {
		result.Online, result.Error = false, "device unreachable"
		return result
	}
	if len(readings) > 0 {
		var power float64
		for _, r := range readings {
			power += r.Power
		}
		result.AvgPower, result.PeakPower, result.DataPoints = power, power, 1
		result.Error = "no historical data"
		return result
	}

	result.Online, result.Error = false, ErrNoPowerMeter.Error()
	return result
}
