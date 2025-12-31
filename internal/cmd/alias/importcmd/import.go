// Package importcmd provides the alias import command.
package importcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the import command.
type Options struct {
	Factory  *cmdutil.Factory
	Filename string
	Merge    bool
}

// NewCommand creates the alias import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Filename = args[0]
			return run(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Merge, "merge", "m", false, "Merge with existing aliases (skip conflicts)")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	imported, skipped, err := config.ImportAliases(opts.Filename, opts.Merge)
	if err != nil {
		return fmt.Errorf("failed to import aliases: %w", err)
	}

	ios.Success("Imported %d alias(es) from %s", imported, opts.Filename)
	if skipped > 0 {
		ios.Info("Skipped %d existing alias(es)", skipped)
	}

	return nil
}
