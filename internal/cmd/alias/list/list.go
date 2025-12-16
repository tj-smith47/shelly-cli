// Package list provides the alias list command.
package list

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the alias list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[config.Alias]{
		Resource:    "alias",
		FetchFunc:   config.ListAliasesSorted,
		DisplayFunc: displayAliases,
		HintMsg:     "Use 'shelly alias set <name> <command>' to create one.",
	})
}

func displayAliases(ios *iostreams.IOStreams, aliases []config.Alias) {
	table := output.NewTable("Name", "Command", "Type")

	for _, alias := range aliases {
		aliasType := "command"
		if alias.Shell {
			aliasType = "shell"
		}
		table.AddRow(alias.Name, alias.Command, aliasType)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Println()
	ios.Count("alias", len(aliases))
}
