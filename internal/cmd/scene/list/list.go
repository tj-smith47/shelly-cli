// Package list provides the scene list subcommand.
package list

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/term"
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
		DisplayFunc: term.DisplaySceneList,
	})
}
