// Package list provides the scene list subcommand.
package list

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the scene list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[config.Scene]{
		Resource: "scene",
		FetchFunc: func() []config.Scene {
			scenes := config.ListScenes()
			result := make([]config.Scene, 0, len(scenes))
			for _, scene := range scenes {
				result = append(result, scene)
			}
			sort.Slice(result, func(i, j int) bool {
				return result[i].Name < result[j].Name
			})
			return result
		},
		DisplayFunc: displayScenes,
	})
}

func displayScenes(ios *iostreams.IOStreams, scenes []config.Scene) {
	table := output.NewTable("Name", "Actions", "Description")
	for _, scene := range scenes {
		actions := formatActionCount(len(scene.Actions))
		description := scene.Description
		if description == "" {
			description = "-"
		}
		table.AddRow(scene.Name, actions, description)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Println()
	ios.Count("scene", len(scenes))
}

func formatActionCount(count int) string {
	return output.FormatActionCount(count)
}
