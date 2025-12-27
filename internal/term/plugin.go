package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// PluginUpgradeResult contains the result of a plugin upgrade attempt.
// This mirrors plugins.UpgradeResult to avoid import cycles.
type PluginUpgradeResult struct {
	Name       string
	OldVersion string
	NewVersion string
	Upgraded   bool
	Skipped    bool
	Error      error
}

// DisplayPluginUpgradeResults prints the results of plugin upgrade operations.
func DisplayPluginUpgradeResults(ios *iostreams.IOStreams, results []PluginUpgradeResult) {
	var upgraded, skipped, failed int

	for _, r := range results {
		switch {
		case r.Error != nil && r.Skipped:
			ios.Warning("  %s: Skipped - %v", r.Name, r.Error)
			skipped++
		case r.Error != nil:
			ios.Error("  %s: Failed - %v", r.Name, r.Error)
			failed++
		case r.Upgraded:
			ios.Success("  %s: Upgraded from %s to %s", r.Name, r.OldVersion, r.NewVersion)
			upgraded++
		default:
			ios.Info("  %s: Already at latest version %s", r.Name, r.OldVersion)
		}
	}

	ios.Println("")
	ios.Printf("Upgrade complete: %d upgraded, %d skipped, %d failed\n", upgraded, skipped, failed)
}

// DisplayPluginUpgradeResult prints the result of a single plugin upgrade.
func DisplayPluginUpgradeResult(ios *iostreams.IOStreams, result PluginUpgradeResult) {
	switch {
	case result.Error != nil:
		ios.Error("Failed to upgrade %s: %v", result.Name, result.Error)
	case result.Upgraded:
		ios.Success("Upgraded %s from %s to %s", result.Name, result.OldVersion, result.NewVersion)
	default:
		ios.Info("%s is already at latest version %s", result.Name, result.OldVersion)
	}
}
