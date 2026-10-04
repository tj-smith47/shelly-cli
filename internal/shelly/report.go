package shelly

import (
	"context"
	"maps"
	"math"
	"slices"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// auditDeviceTimeout bounds the whole audit of one device. Reachability is
// bounded separately by DefaultTimeout (see AuditDevice), so an offline
// device costs DefaultTimeout; the rest is for the cloud and firmware checks,
// which on a Gen1 device are rate limited to one request every two seconds.
const auditDeviceTimeout = 3 * DefaultTimeout

// ForEachDevice calls fn once for every device, all at the same time, and
// returns when every call has returned. Each call gets its own context with
// the given timeout, so one slow or offline device does not use up the time
// of the others. i is the device's index in devices, so fn can write its
// result to its own slot of a slice without locking.
func ForEachDevice(ctx context.Context, devices []string, timeout time.Duration, fn func(ctx context.Context, i int, device string)) {
	var wg sync.WaitGroup
	for i, device := range devices {
		wg.Go(func() {
			devCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			fn(devCtx, i, device)
		})
	}
	wg.Wait()
}

// AuditDevices audits every device concurrently and returns the results in
// the order of devices.
func (s *Service) AuditDevices(ctx context.Context, devices []string, checks model.AuditChecks) []*model.AuditResult {
	results := make([]*model.AuditResult, len(devices))
	ForEachDevice(ctx, devices, auditDeviceTimeout, func(ctx context.Context, i int, device string) {
		results[i] = s.AuditDevice(ctx, device, checks)
	})
	return results
}

// GenerateDevicesReport asks every registered device for its identity and
// firmware, concurrently, and returns the inventory sorted by device name.
func (s *Service) GenerateDevicesReport(ctx context.Context, devices map[string]model.Device) model.DevicesReport {
	names := slices.Sorted(maps.Keys(devices))
	report := model.DevicesReport{
		Timestamp:  time.Now(),
		ReportType: model.ReportTypeDevices,
		Devices:    make([]model.DeviceReportRow, len(names)),
	}
	// devices is usually the live registry, which DeviceInfo updates as
	// devices answer, so the registered values are copied before any call.
	for i, name := range names {
		registered := devices[name]
		report.Devices[i] = model.DeviceReportRow{
			Name:       name,
			IP:         registered.Address,
			Model:      registered.Model,
			Generation: registered.Generation,
			MAC:        model.NormalizeMAC(registered.MAC),
		}
	}
	ForEachDevice(ctx, names, DefaultTimeout, func(ctx context.Context, i int, name string) {
		row := &report.Devices[i]
		info, err := s.DeviceInfo(ctx, name)
		if err != nil {
			row.Error = err.Error()
			return
		}
		row.Online = true
		row.Model = info.Model
		row.Generation = info.Generation
		row.Firmware = info.Firmware
		if mac := model.NormalizeMAC(info.MAC); mac != "" {
			row.MAC = mac
		}
	})

	report.Summary.Total = len(report.Devices)
	for _, row := range report.Devices {
		if row.Online {
			report.Summary.Online++
		} else {
			report.Summary.Offline++
		}
	}
	return report
}

// GenerateEnergyReport reads the power meters of every registered device,
// concurrently, Gen1 and Gen2+ alike, and returns each device's active power
// and the total, sorted by device name.
func (s *Service) GenerateEnergyReport(ctx context.Context, devices map[string]model.Device) model.EnergyReport {
	names := slices.Sorted(maps.Keys(devices))
	report := model.EnergyReport{
		Timestamp:  time.Now(),
		ReportType: model.ReportTypeEnergy,
		Devices:    make([]model.EnergyReportRow, len(names)),
	}
	ForEachDevice(ctx, names, 2*DefaultTimeout, func(ctx context.Context, i int, name string) {
		report.Devices[i] = s.energyReportRow(ctx, name)
	})

	report.Summary.Total = len(report.Devices)
	for _, row := range report.Devices {
		if !row.Online {
			report.Summary.Offline++
			continue
		}
		report.Summary.Online++
		if row.Reporting {
			report.Summary.DevicesReporting++
			report.Summary.TotalPowerW += row.PowerW
		}
	}
	report.Summary.TotalPowerW = roundWatts(report.Summary.TotalPowerW)
	return report
}

// energyReportRow checks that the device answers before reading its meters,
// because a monitoring snapshot of a device that does not answer is empty
// rather than an error, which would read as "online, no meters".
func (s *Service) energyReportRow(ctx context.Context, name string) model.EnergyReportRow {
	row := model.EnergyReportRow{Name: name}
	if _, err := s.probeDevice(ctx, name); err != nil {
		row.Error = err.Error()
		return row
	}
	row.Online = true
	snapshot, err := s.GetMonitoringSnapshotAuto(ctx, name)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	row.Reporting = len(snapshot.EM)+len(snapshot.EM1)+len(snapshot.PM) > 0
	power, _ := output.CalculateSnapshotTotals(snapshot)
	row.PowerW = roundWatts(power)
	return row
}

// roundWatts rounds to 0.01 W, so sums of meter readings print as 48.15
// rather than 48.150000000000006.
func roundWatts(w float64) float64 {
	return math.Round(w*100) / 100
}

// GenerateAuditReport runs every audit check (the same checks as `shelly
// audit`) on every registered device, concurrently, and returns one row per
// device sorted by name, with the counts in the summary.
func (s *Service) GenerateAuditReport(ctx context.Context, devices map[string]model.Device) model.AuditReport {
	names := slices.Sorted(maps.Keys(devices))
	report := model.AuditReport{
		Timestamp:  time.Now(),
		ReportType: model.ReportTypeAudit,
		Devices:    make([]model.AuditReportRow, 0, len(names)),
	}
	sum := &report.Summary
	for _, result := range s.AuditDevices(ctx, names, model.AllAuditChecks()) {
		row := model.NewAuditReportRow(result)
		report.Devices = append(report.Devices, row)
		sum.DevicesScanned++
		sum.Issues += len(row.Issues)
		sum.Warnings += len(row.Warnings)
		if !row.Reachable {
			sum.Unreachable++
			continue
		}
		sum.Reachable++
		if row.AuthEnabled != nil {
			if *row.AuthEnabled {
				sum.AuthEnabled++
			} else {
				sum.AuthDisabled++
			}
		}
		if row.CloudConnected != nil && *row.CloudConnected {
			sum.CloudConnected++
		}
		if row.FirmwareOutdated != nil && *row.FirmwareOutdated {
			sum.OutdatedFirmware++
		}
	}
	return report
}

// probeDevice is DeviceInfo bounded by DefaultTimeout, for a reachability
// check that must leave the rest of the caller's time for the calls after it.
func (s *Service) probeDevice(ctx context.Context, identifier string) (*DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()
	return s.DeviceInfo(ctx, identifier)
}
