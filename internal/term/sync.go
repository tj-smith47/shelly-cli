package term

import "github.com/tj-smith47/shelly-cli/internal/iostreams"

// DisplaySyncProgress displays the sync status for a device.
func DisplaySyncProgress(ios *iostreams.IOStreams, device, status string) {
	ios.Printf("  %s: %s\n", device, status)
}

// DisplaySyncSummary displays the overall sync summary.
func DisplaySyncSummary(ios *iostreams.IOStreams, success, failed int, dryRun bool, syncDir string) {
	ios.Println()
	if failed > 0 {
		ios.Warning("Completed: %d succeeded, %d failed", success, failed)
	} else {
		ios.Success("Completed: %d device(s) synced", success)
	}

	if !dryRun {
		ios.Info("Configs saved to: %s", syncDir)
	}
}

// DisplayPushSummary displays the overall push summary.
func DisplayPushSummary(ios *iostreams.IOStreams, success, failed, skipped int) {
	ios.Println()
	if failed > 0 {
		ios.Warning("Completed: %d succeeded, %d failed, %d skipped", success, failed, skipped)
	} else {
		ios.Success("Completed: %d device(s) updated", success)
	}
}

// DisplaySyncHeader displays the sync operation header.
func DisplaySyncHeader(ios *iostreams.IOStreams, op string, count int, dryRun bool) {
	ios.Info("%s configurations %s %d device(s)...", op, direction(op), count)
	if dryRun {
		ios.Warning("[DRY RUN] No changes will be made")
	}
	ios.Println()
}

func direction(op string) string {
	if op == "Pulling" {
		return "from"
	}
	return "to"
}
