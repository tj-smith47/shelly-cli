// Package list provides the alias list command.
package list

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the alias list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all command aliases",
		Long:    `Display all configured command aliases with their expansion.`,
		Example: `  # List all aliases
  shelly alias list

  # List as JSON
  shelly alias list -o json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	aliases := cfg.ListAliases()

	if len(aliases) == 0 {
		ios.Info("No aliases configured. Use 'shelly alias set <name> <command>' to create one.")
		return nil
	}

	// Handle JSON/YAML output
	if cmdutil.WantsStructured() {
		return cmdutil.FormatOutput(ios, aliases)
	}

	// Table output
	printTable(ios, aliases)
	return nil
}

func printTable(ios *iostreams.IOStreams, aliases []config.Alias) {
	table := output.NewTable("Name", "Command", "Type")

	for _, alias := range aliases {
		aliasType := "command"
		if alias.Shell {
			aliasType = "shell"
		}
		table.AddRow(alias.Name, alias.Command, aliasType)
	}

	table.Print()
	ios.Count("alias", len(aliases))
}
