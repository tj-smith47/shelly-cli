// Package deletecmd provides the scene delete subcommand.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the scene delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:      "scene",
		ValidArgsFunc: completion.SceneNames(),
		ExistsFunc: func(name string) (any, bool) {
			return config.GetScene(name)
		},
		DeleteFunc: config.DeleteScene,
		InfoFunc: func(resource any, name string) string {
			scene, ok := resource.(config.Scene)
			if !ok || len(scene.Actions) == 0 {
				return fmt.Sprintf("Delete scene %q?", name)
			}
			return fmt.Sprintf("Delete scene %q with %d action(s)?", name, len(scene.Actions))
		},
	})
}
