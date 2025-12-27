package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// DisplayBulkProvisionDryRun shows what would be provisioned without making changes.
func DisplayBulkProvisionDryRun(ios *iostreams.IOStreams, cfg *model.BulkProvisionConfig) {
	ios.Info("Dry run - validating configuration:")
	for _, d := range cfg.Devices {
		wifi := cfg.WiFi
		if d.WiFi != nil {
			wifi = d.WiFi
		}
		if wifi == nil {
			ios.Warning("  %s: no WiFi config", d.Name)
		} else {
			ios.Info("  %s: SSID=%s", d.Name, wifi.SSID)
		}
	}
}

// DisplayBulkProvisionResults shows the results of bulk provisioning.
// Returns the number of failed devices.
func DisplayBulkProvisionResults(ios *iostreams.IOStreams, results []model.ProvisionResult, totalDevices int) int {
	var failed int
	for _, r := range results {
		if r.Err != nil {
			ios.Error("Failed to provision %s: %v", r.Device, r.Err)
			failed++
		} else {
			ios.Success("Provisioned %s", r.Device)
		}
	}

	if failed == 0 {
		ios.Success("All %d devices provisioned successfully", totalDevices)
	}

	return failed
}
