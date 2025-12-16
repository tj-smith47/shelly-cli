// Package exportcmd provides the alias export command.
package exportcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the alias export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [file]",
		Aliases: []string{"save", "dump"},
		Short:   "Export aliases to a YAML file",
		Long: `Export all command aliases to a YAML file.

If no file is specified, outputs to stdout.

The output format is:
  aliases:
    name1: "command1"
    name2: "command2"
    shellalias: "!shell command"`,
		Example: `  # Export to file
  shelly alias export aliases.yaml

  # Export to stdout
  shelly alias export

  # Export and pipe to another command
  shelly alias export | cat`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := ""
			if len(args) > 0 {
				filename = args[0]
			}
			return run(f, filename)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, filename string) error {
	ios := f.IOStreams()

	aliases := config.ListAliases()
	if len(aliases) == 0 {
		ios.Warning("No aliases to export")
		return nil
	}

	output, err := config.ExportAliases(filename)
	if err != nil {
		return fmt.Errorf("failed to export aliases: %w", err)
	}

	if filename == "" {
		// Print YAML to stdout
		fmt.Print(output)
	} else {
		ios.Success("Exported %d alias(es) to %s", len(aliases), filename)
	}

	return nil
}
