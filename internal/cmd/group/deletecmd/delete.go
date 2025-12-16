// Package deletecmd provides the group delete subcommand.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the group delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:      "group",
		ValidArgsFunc: completion.GroupNames(),
		ExistsFunc: func(name string) (any, bool) {
			return config.GetGroup(name)
		},
		DeleteFunc: config.DeleteGroup,
		InfoFunc: func(resource any, name string) string {
			group, ok := resource.(config.Group)
			if !ok || len(group.Devices) == 0 {
				return fmt.Sprintf("Delete group %q?", name)
			}
			return fmt.Sprintf("Delete group %q with %d device(s)?", name, len(group.Devices))
		},
	})
}
