// Package output provides pure formatters (data → string).
package output

import (
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// FormatTextReport formats a report as human-readable text.
func FormatTextReport(report model.DeviceReport) string {
	var result string

	result += fmt.Sprintf("Shelly %s Report\n", report.ReportType)
	result += fmt.Sprintf("Generated: %s\n\n", report.Timestamp.Format(time.RFC3339))

	if len(report.Devices) > 0 {
		result += "Devices:\n"
		for _, d := range report.Devices {
			status := "offline"
			if d.Online {
				status = "online"
			}
			result += fmt.Sprintf("  - %s (%s): %s\n", d.Name, d.IP, status)
			if d.Model != "" {
				result += fmt.Sprintf("    Model: %s, Firmware: %s\n", d.Model, d.Firmware)
			}
		}
		result += "\n"
	}

	result += "Summary:\n"
	for k, v := range report.Summary {
		result += fmt.Sprintf("  %s: %v\n", k, v)
	}

	return result
}
