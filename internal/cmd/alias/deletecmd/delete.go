// Package deletecmd provides the alias delete command.
package deletecmd

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the alias delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:         "alias",
		ValidArgsFunc:    completion.AliasNames(),
		SkipConfirmation: true,
		ExistsFunc: func(name string) (any, bool) {
			return nil, config.IsAlias(name)
		},
		DeleteFunc: config.RemoveAlias,
	})
}
