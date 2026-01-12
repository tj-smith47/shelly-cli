// Package list provides the alias list command.
package list

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the alias list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[config.NamedAlias]{
		Resource:    "alias",
		FetchFunc:   config.ListAliasesSorted,
		DisplayFunc: term.DisplayAliasList,
		HintMsg:     "Use 'shelly alias set <name> <command>' to create one.",
	})
}
