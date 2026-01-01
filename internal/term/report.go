package term

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// OutputReport outputs a report in the specified format.
func OutputReport(ios *iostreams.IOStreams, report model.DeviceReport, format, outputPath string) error {
	var content string

	switch format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
		content = string(data)

	case "text":
		content = output.FormatTextReport(report)

	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	if outputPath != "" {
		if err := afero.WriteFile(config.Fs(), outputPath, []byte(content), 0o600); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		ios.Success("Report saved to: %s", outputPath)
		return nil
	}

	ios.Println(content)
	return nil
}
