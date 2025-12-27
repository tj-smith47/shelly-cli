package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayDeviceList renders a device list as a table.
func DisplayDeviceList(ios *iostreams.IOStreams, devices []model.DeviceListItem, showPlatform, showVersion bool) {
	// Build headers based on options
	headers := []string{"#", "Name", "Address"}
	if showPlatform {
		headers = append(headers, "Platform")
	}
	headers = append(headers, "Type", "Model", "Generation")
	if showVersion {
		headers = append(headers, "Version", "Update")
	}
	headers = append(headers, "Auth")

	table := output.NewStyledTable(ios, headers...)

	for i, dev := range devices {
		gen := output.RenderGeneration(dev.Generation)
		auth := output.RenderAuthRequired(dev.Auth)

		// Build row based on options
		row := []string{fmt.Sprintf("%d", i+1), dev.Name, dev.Address}
		if showPlatform {
			row = append(row, dev.Platform)
		}
		row = append(row, dev.Type, dev.Model, gen)
		if showVersion {
			version := dev.CurrentVersion
			if version == "" {
				version = "-"
			}
			update := "-"
			if dev.HasUpdate && dev.AvailableVersion != "" {
				update = theme.StatusOK().Render("↑ " + dev.AvailableVersion)
			}
			row = append(row, version, update)
		}
		row = append(row, auth)

		table.AddRow(row...)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	if _, err := fmt.Fprintln(ios.Out); err != nil {
		ios.DebugErr("print newline", err)
	}
}
