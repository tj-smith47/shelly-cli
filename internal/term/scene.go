package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplaySceneDetails prints detailed scene information.
func DisplaySceneDetails(ios *iostreams.IOStreams, scene config.Scene) {
	ios.Info("Scene: %s", theme.Bold().Render(scene.Name))
	if scene.Description != "" {
		ios.Info("Description: %s", scene.Description)
	}

	if len(scene.Actions) == 0 {
		ios.Info("")
		ios.Info("%s", theme.Dim().Render("No actions defined"))
		ios.Info("Add actions with: shelly scene add-action %s <device> <method> [params]", scene.Name)
		return
	}

	ios.Info("")
	ios.Info("Actions (%d):", len(scene.Actions))

	builder := table.NewBuilder("#", "Device", "Method", "Parameters")

	for i, action := range scene.Actions {
		params := "-"
		if len(action.Params) > 0 {
			params = output.FormatParamsInline(action.Params)
		}
		builder.AddRow(
			theme.Dim().Render(fmt.Sprintf("%d", i+1)),
			theme.Bold().Render(action.Device),
			theme.Highlight().Render(action.Method),
			params,
		)
	}

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print scene actions table", err)
	}
}
