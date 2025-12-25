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

// firmwareCheckRow holds display values for a single firmware check result.
type firmwareCheckRow struct {
	name, platform, current, stable, beta, status string
	hasUpdate                                     bool
}

// buildFirmwareCheckRow builds display values from a firmware check result.
func buildFirmwareCheckRow(r shelly.FirmwareCheckResult) firmwareCheckRow {
	row := firmwareCheckRow{name: r.Name}

	if r.Err != nil {
		row.status = output.RenderErrorState()
		row.platform = output.LabelPlaceholder
		row.current = output.LabelPlaceholder
		row.stable = r.Err.Error()
		row.beta = output.LabelPlaceholder
		return row
	}

	// Get platform, defaulting to "shelly" if empty
	row.platform = r.Info.Platform
	if row.platform == "" {
		row.platform = "shelly"
	}

	row.current = r.Info.Current
	row.stable = r.Info.Available
	if row.stable == "" {
		row.stable = output.LabelPlaceholder
	}
	row.beta = r.Info.Beta
	if row.beta == "" {
		row.beta = output.LabelPlaceholder
	}

	if r.Info.HasUpdate {
		row.status = output.RenderBoolState(true, "update available", "")
		row.hasUpdate = true
	} else {
		row.status = output.FormatPlaceholder("up to date")
	}

	return row
}

// DisplayFirmwareCheckAll displays the results of checking firmware on all devices.
func DisplayFirmwareCheckAll(ios *iostreams.IOStreams, results []shelly.FirmwareCheckResult) {
	table := output.NewTable("Device", "Platform", "Current", "Stable", "Beta", "Status")
	updatesAvailable := 0

	for _, r := range results {
		row := buildFirmwareCheckRow(r)
		if row.hasUpdate {
			updatesAvailable++
		}
		table.AddRow(row.name, row.platform, row.current, row.stable, row.beta, row.status)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	ios.Println("")
	if updatesAvailable > 0 {
		ios.Success("%d device(s) have updates available", updatesAvailable)
	} else {
		ios.Info("All devices are up to date")
	}
}

// UpdateTarget contains information about a firmware update target.
type UpdateTarget struct {
	DeviceID    string
	DeviceModel string
	Current     string
	Available   string
	Beta        string
	CustomURL   string
	UseBeta     bool
}

// DisplayUpdateTarget prints information about a firmware update target.
func DisplayUpdateTarget(ios *iostreams.IOStreams, target UpdateTarget) {
	ios.Println("")
	ios.Printf("  Device:  %s (%s)\n", target.DeviceID, target.DeviceModel)
	ios.Printf("  Current: %s\n", target.Current)
	switch {
	case target.CustomURL != "":
		ios.Printf("  Target:  %s\n", theme.StatusWarn().Render("custom URL"))
	case target.UseBeta:
		ios.Printf("  Target:  %s %s\n", target.Beta, theme.StatusWarn().Render("(beta)"))
	default:
		ios.Printf("  Target:  %s\n", target.Available)
	}
	ios.Println("")
}

// UpdateResult contains the result of a firmware update operation.
type UpdateResult struct {
	Name    string
	Success bool
	Err     error
}

// DisplayUpdateResults prints the results of batch firmware updates.
func DisplayUpdateResults(ios *iostreams.IOStreams, results []UpdateResult) {
	ios.Println("")
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			ios.Success("Updated: %s", r.Name)
			successCount++
		} else {
			ios.Error("Failed: %s - %v", r.Name, r.Err)
			failCount++
		}
	}

	ios.Println("")
	if failCount > 0 {
		ios.Warning("Updated %d device(s), %d failed", successCount, failCount)
	} else {
		ios.Success("Updated %d device(s)", successCount)
	}
}

// DisplayDevicesToUpdate prints a table of devices that will be updated.
func DisplayDevicesToUpdate(ios *iostreams.IOStreams, devices []shelly.DeviceUpdateStatus) {
	ios.Println("")
	ios.Printf("%s\n", theme.Bold().Render("Devices to update:"))
	table := output.NewTable("Device", "Current", "Available")
	for _, s := range devices {
		table.AddRow(s.Name, s.Info.Current, s.Info.Available)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Println("")
}
