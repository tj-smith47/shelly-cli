// Package deletecmd provides the alias delete command.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the alias delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"del", "rm", "remove"},
		Short:   "Delete a command alias",
		Long:    `Delete an existing command alias by name.`,
		Example: `  # Delete an alias
  shelly alias delete lights

  # Using short form
  shelly alias rm lights`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AliasNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(f, args[0])
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, name string) error {
	ios := f.IOStreams()

	// Check if alias exists
	if !config.IsAlias(name) {
		return fmt.Errorf("alias %q not found", name)
	}

	if err := config.RemoveAlias(name); err != nil {
		return fmt.Errorf("failed to delete alias: %w", err)
	}

	ios.Success("Deleted alias '%s'", name)
	return nil
}
