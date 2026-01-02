// Package term provides terminal display functions for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-go/gen1"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
)

// DisplayGen1Actions displays action URLs for a Gen1 device.
func DisplayGen1Actions(ios *iostreams.IOStreams, actions *gen1.ActionSettings) {
	if actions == nil || len(actions.Actions) == 0 {
		ios.Info("No actions configured")
		return
	}

	builder := table.NewBuilder("INDEX", "EVENT", "ENABLED", "URLS")
	for _, action := range actions.Actions {
		enabled := "no"
		if action.Enabled {
			enabled = "yes"
		}
		urls := output.LabelPlaceholder
		if len(action.URLs) > 0 {
			urls = action.URLs[0]
			if len(action.URLs) > 1 {
				urls += fmt.Sprintf(" (+%d more)", len(action.URLs)-1)
			}
		}
		builder.AddRow(
			fmt.Sprintf("%d", action.Index),
			string(action.Event),
			enabled,
			urls,
		)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print actions table", err)
	}
}

// DisplayGen1ActionURLs displays all URLs for a specific action.
func DisplayGen1ActionURLs(ios *iostreams.IOStreams, action gen1.Action) {
	ios.Printf("Action: %s (index %d)\n", action.Event, action.Index)
	ios.Printf("Enabled: %t\n", action.Enabled)
	if len(action.URLs) == 0 {
		ios.Printf("URLs: none\n")
		return
	}
	ios.Printf("URLs:\n")
	for i, url := range action.URLs {
		ios.Printf("  [%d] %s\n", i, url)
	}
}
