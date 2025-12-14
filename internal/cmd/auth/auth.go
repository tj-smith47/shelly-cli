// Package auth provides device authentication commands.
package auth

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/auth/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/auth/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/auth/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the auth command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"authentication"},
		Short:   "Manage device authentication",
		Long: `Manage device authentication settings.

Enable, configure, or disable authentication for local device access.
When authentication is enabled, a username and password are required
for all device operations.`,
		Example: `  # Show authentication status
  shelly auth status living-room

  # Set authentication credentials
  shelly auth set living-room --user admin --password secret

  # Disable authentication
  shelly auth disable living-room`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(disable.NewCommand(f))

	return cmd
}
