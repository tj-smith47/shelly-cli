package output

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

var reportTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

func TestFormatDevicesReportText(t *testing.T) {
	t.Parallel()

	got := FormatDevicesReportText(model.DevicesReport{
		Timestamp: reportTime,
		Devices: []model.DeviceReportRow{
			{Name: "bedroom", IP: "192.168.1.101", Model: "Shelly Duo", Generation: 1},
			{Name: "kitchen", IP: "192.168.1.100", Model: "Shelly Plus 1PM", Generation: 2,
				Firmware: "1.4.4", MAC: "AABBCCDDEEFF", Online: true},
		},
		Summary: model.DevicesReportSummary{Total: 2, Online: 1, Offline: 1},
	})

	for _, want := range []string{
		"Shelly Device inventory report", "Generated: 2024-01-15T10:30:00Z",
		"NAME", "bedroom", "offline", "kitchen", "online", "Shelly Plus 1PM", "1.4.4",
		"Summary: 2 devices, 1 online, 1 offline",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report lacks %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "bedroom") > strings.Index(got, "kitchen") {
		t.Errorf("rows are not in the given order:\n%s", got)
	}
}

func TestFormatEnergyReportText(t *testing.T) {
	t.Parallel()

	got := FormatEnergyReportText(model.EnergyReport{
		Timestamp: reportTime,
		Devices: []model.EnergyReportRow{
			{Name: "attic"},
			{Name: "bulb", Online: true, Reporting: true, PowerW: 7.5},
			{Name: "button", Online: true},
		},
		Summary: model.EnergyReportSummary{Total: 3, Online: 2, Offline: 1, DevicesReporting: 1, TotalPowerW: 7.5},
	})

	for _, want := range []string{"7.5 W", "no meter", "Total power: 7.5 W from 1 reporting devices (3 devices, 2 online, 1 offline)"} {
		if !strings.Contains(got, want) {
			t.Errorf("report lacks %q:\n%s", want, got)
		}
	}
}

func TestFormatAuditReportText(t *testing.T) {
	t.Parallel()

	no, yes := false, true
	got := FormatAuditReportText(model.AuditReport{
		Timestamp: reportTime,
		Devices: []model.AuditReportRow{
			{Name: "attic", Issues: []string{"Device unreachable"}},
			{Name: "plug", Reachable: true, AuthEnabled: &no, CloudConnected: &yes,
				FirmwareCurrent: "1.4.4", FirmwareAvailable: "1.5.0", FirmwareOutdated: &yes,
				Warnings: []string{"Firmware update available: 1.4.4 -> 1.5.0"}},
		},
		Summary: model.AuditReportSummary{DevicesScanned: 2, Reachable: 1, Unreachable: 1, AuthDisabled: 1,
			CloudConnected: 1, OutdatedFirmware: 1, Issues: 1, Warnings: 1},
	})

	for _, want := range []string{
		"disabled", "connected", "1.5.0",
		"attic: issue: Device unreachable", "plug: warning: Firmware update available",
		"Summary: 2 devices scanned, 1 reachable, 1 unreachable", "Outdated firmware: 1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report lacks %q:\n%s", want, got)
		}
	}
}
