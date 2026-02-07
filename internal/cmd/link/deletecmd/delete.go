// Package deletecmd provides the link delete subcommand.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the link delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:      "link",
		ValidArgsFunc: completion.LinkedDeviceNames(),
		ExistsFunc: func(name string) (any, bool) {
			return config.GetLink(name)
		},
		DeleteFunc: config.DeleteLink,
		InfoFunc: func(_ any, name string) string {
			return fmt.Sprintf("Delete link for %q?", name)
		},
	})
}
