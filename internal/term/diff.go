package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// displayDiffSection is a generic helper for printing diff sections.
func displayDiffSection[T any](
	ios *iostreams.IOStreams,
	diffs []T,
	verbose bool,
	header, addedMsg string,
	getLabel func(T) string,
	getDiffType func(T) string,
) {
	if len(diffs) == 0 {
		return
	}
	if verbose {
		ios.Printf("%s:\n", header)
	} else {
		ios.Printf("%s changes:\n", header)
	}
	for _, d := range diffs {
		label := getLabel(d)
		switch getDiffType(d) {
		case model.DiffAdded:
			if verbose {
				ios.Printf("  + %s (%s)\n", label, addedMsg)
			} else {
				ios.Printf("  + %s\n", label)
			}
		case model.DiffRemoved:
			if verbose {
				ios.Printf("  - %s (exists on device, not in backup)\n", label)
			} else {
				ios.Printf("  - %s\n", label)
			}
		case model.DiffChanged:
			if verbose {
				ios.Printf("  ~ %s (will be updated)\n", label)
			} else {
				ios.Printf("  ~ %s\n", label)
			}
		}
	}
	ios.Println()
}

// DisplayConfigDiffs prints configuration differences.
func DisplayConfigDiffs(ios *iostreams.IOStreams, diffs []model.ConfigDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Configuration", "will be added from backup",
		func(d model.ConfigDiff) string { return d.Path },
		func(d model.ConfigDiff) string { return d.DiffType })
}

// DisplayScriptDiffs prints script differences.
func DisplayScriptDiffs(ios *iostreams.IOStreams, diffs []model.ScriptDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Script", "will be created",
		func(d model.ScriptDiff) string { return d.Name },
		func(d model.ScriptDiff) string { return d.DiffType })
}

// DisplayScheduleDiffs prints schedule differences.
func DisplayScheduleDiffs(ios *iostreams.IOStreams, diffs []model.ScheduleDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Schedule", "will be created",
		func(d model.ScheduleDiff) string { return d.Timespec },
		func(d model.ScheduleDiff) string { return d.DiffType })
}

// DisplayWebhookDiffs prints webhook differences.
func DisplayWebhookDiffs(ios *iostreams.IOStreams, diffs []model.WebhookDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Webhook", "will be created",
		func(d model.WebhookDiff) string {
			if d.Name != "" {
				return d.Name
			}
			return d.Event
		},
		func(d model.WebhookDiff) string { return d.DiffType })
}

// DisplayConfigDiffsSummary prints configuration differences with grouped sections and summary.
func DisplayConfigDiffsSummary(ios *iostreams.IOStreams, diffs []model.ConfigDiff) {
	var added, removed, changed []model.ConfigDiff
	for _, d := range diffs {
		switch d.DiffType {
		case model.DiffAdded:
			added = append(added, d)
		case model.DiffRemoved:
			removed = append(removed, d)
		case model.DiffChanged:
			changed = append(changed, d)
		}
	}

	if len(removed) > 0 {
		ios.Println(output.RenderDiffRemoved())
		for _, d := range removed {
			ios.Printf("  - %s: %v\n", d.Path, output.FormatDisplayValue(d.OldValue))
		}
		ios.Println("")
	}

	if len(added) > 0 {
		ios.Println(output.RenderDiffAdded())
		for _, d := range added {
			ios.Printf("  + %s: %v\n", d.Path, output.FormatDisplayValue(d.NewValue))
		}
		ios.Println("")
	}

	if len(changed) > 0 {
		ios.Println(output.RenderDiffChanged())
		for _, d := range changed {
			ios.Printf("  ~ %s:\n", d.Path)
			ios.Printf("    - %v\n", output.FormatDisplayValue(d.OldValue))
			ios.Printf("    + %v\n", output.FormatDisplayValue(d.NewValue))
		}
		ios.Println("")
	}

	ios.Printf("Summary: %d added, %d removed, %d changed\n", len(added), len(removed), len(changed))
}
