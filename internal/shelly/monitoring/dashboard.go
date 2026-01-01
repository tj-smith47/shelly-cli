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
	collectComponents(ctx, device, componentCollector[*model.EMStatus]{
		compType: "EM", listIDs: s.ListEMComponents, getStatus: s.GetEMStatus,
		toPower: func(st *model.EMStatus, id int) (model.ComponentPower, float64, float64) {
			return model.ComponentPower{Voltage: st.AVoltage, Current: st.TotalCurrent, Power: st.TotalActivePower}, st.TotalActivePower, 0
		},
	}, &status)

	collectComponents(ctx, device, componentCollector[*model.EM1Status]{
		compType: "EM1", listIDs: s.ListEM1Components, getStatus: s.GetEM1Status,
		toPower: func(st *model.EM1Status, id int) (model.ComponentPower, float64, float64) {
			return model.ComponentPower{Voltage: st.Voltage, Current: st.Current, Power: st.ActPower}, st.ActPower, 0
		},
	}, &status)

	collectComponents(ctx, device, componentCollector[*model.PMStatus]{
		compType: "PM", listIDs: s.ListPMComponents, getStatus: s.GetPMStatus,
		toPower: func(st *model.PMStatus, id int) (model.ComponentPower, float64, float64) {
			energy := 0.0
			if st.AEnergy != nil {
				energy = st.AEnergy.Total
			}
			return model.ComponentPower{Voltage: st.Voltage, Current: st.Current, Power: st.APower, Energy: energy}, st.APower, energy
		},
	}, &status)

	collectComponents(ctx, device, componentCollector[*model.PMStatus]{
		compType: "PM1", listIDs: s.ListPM1Components, getStatus: s.GetPM1Status,
		toPower: func(st *model.PMStatus, id int) (model.ComponentPower, float64, float64) {
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
