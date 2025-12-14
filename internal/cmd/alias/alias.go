// Package alias provides command alias management.
package alias

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/alias/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alias/exportcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alias/importcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alias/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alias/set"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the alias command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alias",
		Aliases: []string{"aliases"},
		Short:   "Manage command aliases",
		Long: `Create, list, and manage command aliases.

Aliases allow you to create shortcuts for frequently used commands.
They support argument interpolation with $1, $2, etc., and $@ for all arguments.

Shell command aliases are prefixed with ! and execute in your shell.`,
		Example: `  # List all aliases
  shelly alias list

  # Create a simple alias
  shelly alias set lights "batch on living-room kitchen bedroom"

  # Create an alias with arguments
  shelly alias set sw "switch $1 $2"
  # Usage: shelly sw on kitchen

  # Create a shell alias
  shelly alias set backup '!tar -czf shelly-backup.tar.gz ~/.config/shelly'

  # Delete an alias
  shelly alias delete lights

  # Export aliases to file
  shelly alias export aliases.yaml

  # Import aliases from file
  shelly alias import aliases.yaml`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(importcmd.NewCommand(f))
	cmd.AddCommand(exportcmd.NewCommand(f))

	return cmd
}
