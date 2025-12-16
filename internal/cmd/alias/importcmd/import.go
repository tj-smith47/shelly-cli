// Package importcmd provides the alias import command.
package importcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the alias import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var merge bool

	cmd := &cobra.Command{
		Use:     "import <file>",
		Aliases: []string{"load"},
		Short:   "Import aliases from a YAML file",
		Long: `Import command aliases from a YAML file.

By default, existing aliases with the same name are overwritten.
Use --merge to skip existing aliases instead of overwriting them.

The file format is:
  aliases:
    name1: "command1"
    name2: "command2"
    shellalias: "!shell command"`,
		Example: `  # Import aliases (overwrite existing)
  shelly alias import aliases.yaml

  # Import without overwriting existing aliases
  shelly alias import aliases.yaml --merge`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(f, args[0], merge)
		},
	}

	cmd.Flags().BoolVarP(&merge, "merge", "m", false, "Merge with existing aliases (skip conflicts)")

	return cmd
}

func run(f *cmdutil.Factory, filename string, merge bool) error {
	ios := f.IOStreams()

	imported, skipped, err := config.ImportAliases(filename, merge)
	if err != nil {
		return fmt.Errorf("failed to import aliases: %w", err)
	}

	ios.Success("Imported %d alias(es) from %s", imported, filename)
	if skipped > 0 {
		ios.Info("Skipped %d existing alias(es)", skipped)
	}

	return nil
}
