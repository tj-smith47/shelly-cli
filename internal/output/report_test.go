package output

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestFormatTextReport(t *testing.T) {
	t.Parallel()

	t.Run("basic report", func(t *testing.T) {
		t.Parallel()
		report := model.DeviceReport{
			ReportType: "Status",
			Timestamp:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Devices: []model.DeviceReportInfo{
				{
					Name:     "Kitchen Light",
					IP:       "192.168.1.100",
					Online:   true,
					Model:    "SHSW-1",
					Firmware: "1.0.0",
				},
				{
					Name:   "Bedroom Switch",
					IP:     "192.168.1.101",
					Online: false,
				},
			},
			Summary: map[string]any{
				"total":  2,
				"online": 1,
			},
		}

		got := FormatTextReport(report)

		// Verify key elements
		if !strings.Contains(got, "Status Report") {
			t.Error("expected report to contain 'Status Report'")
		}
		if !strings.Contains(got, "Kitchen Light") {
			t.Error("expected report to contain 'Kitchen Light'")
		}
		if !strings.Contains(got, "online") {
			t.Error("expected report to contain 'online'")
		}
		if !strings.Contains(got, "offline") {
			t.Error("expected report to contain 'offline'")
		}
		if !strings.Contains(got, "SHSW-1") {
			t.Error("expected report to contain model")
		}
	})

	t.Run("empty devices", func(t *testing.T) {
		t.Parallel()
		report := model.DeviceReport{
			ReportType: "Test",
			Timestamp:  time.Now(),
			Summary: map[string]any{
				"count": 0,
			},
		}

		got := FormatTextReport(report)

		if !strings.Contains(got, "Test Report") {
			t.Error("expected report to contain 'Test Report'")
		}
		if !strings.Contains(got, "Summary") {
			t.Error("expected report to contain 'Summary'")
		}
	})
}
