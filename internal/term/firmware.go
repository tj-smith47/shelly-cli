package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayFirmwareStatus prints the firmware status.
func DisplayFirmwareStatus(ios *iostreams.IOStreams, status *shelly.FirmwareStatus) {
	ios.Println(theme.Bold().Render("Firmware Status"))
	ios.Println("")

	// Status
	statusStr := status.Status
	if statusStr == "" {
		statusStr = "idle"
	}
	ios.Printf("  Status:      %s\n", statusStr)

	// Update available
	ios.Printf("  Update:      %s\n", output.RenderAvailableState(status.HasUpdate, "up to date"))
	if status.HasUpdate && status.NewVersion != "" {
		ios.Printf("  New Version: %s\n", status.NewVersion)
	}

	// Progress (if updating)
	if status.Progress > 0 {
		ios.Printf("  Progress:    %d%%\n", status.Progress)
	}

	// Rollback
	ios.Printf("  Rollback:    %s\n", output.RenderAvailableState(status.CanRollback, "not available"))
}

// DisplayFirmwareInfo prints firmware check information.
func DisplayFirmwareInfo(ios *iostreams.IOStreams, info *shelly.FirmwareInfo) {
	ios.Println(theme.Bold().Render("Firmware Information"))
	ios.Println("")

	// Device info
	ios.Printf("  Device:     %s (%s)\n", info.DeviceID, info.DeviceModel)
	ios.Printf("  Generation: Gen%d\n", info.Generation)
	ios.Println("")

	// Version info
	ios.Printf("  Current:    %s\n", info.Current)

	switch {
	case info.HasUpdate:
		ios.Printf("  Available:  %s %s\n",
			info.Available,
			output.RenderUpdateStatus(true))
	case info.Available != "":
		ios.Printf("  Available:  %s %s\n",
			info.Available,
			output.RenderUpdateStatus(false))
	default:
		ios.Printf("  Available:  %s\n", output.RenderUpdateStatus(false))
	}

	if info.Beta != "" {
		ios.Printf("  Beta:       %s\n", info.Beta)
	}
}
