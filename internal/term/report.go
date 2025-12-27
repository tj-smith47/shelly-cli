package term

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// OutputReport outputs a report in the specified format.
func OutputReport(ios *iostreams.IOStreams, report model.DeviceReport, format, outputPath string) error {
	var output string

	switch format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
		output = string(data)

	case "text":
		output = FormatTextReport(report)

	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0o600); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		ios.Success("Report saved to: %s", outputPath)
		return nil
	}

	ios.Println(output)
	return nil
}

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
