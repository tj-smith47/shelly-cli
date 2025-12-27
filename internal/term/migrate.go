package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// DisplayMigrationPreview shows what would be migrated in a dry run.
func DisplayMigrationPreview(ios *iostreams.IOStreams, source, sourceType, target string, diff *model.BackupDiff) {
	ios.Title("Migration Preview (dry run)")
	ios.Println()
	ios.Printf("Source: %s (%s)\n", source, sourceType)
	ios.Printf("Target: %s\n", target)
	ios.Println()

	if !diff.HasDifferences() {
		ios.Info("No differences found - target already matches source")
		return
	}

	DisplayMigrationDiff(ios, diff)
}

// DisplayMigrationDiff shows the differences that would be applied.
func DisplayMigrationDiff(ios *iostreams.IOStreams, d *model.BackupDiff) {
	// Use consolidated display functions with verbose=false for concise output
	DisplayConfigDiffs(ios, d.ConfigDiffs, false)
	DisplayScriptDiffs(ios, d.ScriptDiffs, false)
	DisplayScheduleDiffs(ios, d.ScheduleDiffs, false)
	DisplayWebhookDiffs(ios, d.WebhookDiffs, false)
}

// DisplayMigrationResult shows the result of a migration.
func DisplayMigrationResult(ios *iostreams.IOStreams, result *backup.RestoreResult) {
	ios.Println()
	if result.ConfigRestored {
		ios.Printf("  Config:    migrated\n")
	}
	if result.ScriptsRestored > 0 {
		ios.Printf("  Scripts:   %d migrated\n", result.ScriptsRestored)
	}
	if result.SchedulesRestored > 0 {
		ios.Printf("  Schedules: %d migrated\n", result.SchedulesRestored)
	}
	if result.WebhooksRestored > 0 {
		ios.Printf("  Webhooks:  %d migrated\n", result.WebhooksRestored)
	}

	if len(result.Warnings) > 0 {
		ios.Println()
		ios.Warning("Warnings:")
		for _, w := range result.Warnings {
			ios.Printf("  - %s\n", w)
		}
	}
}
