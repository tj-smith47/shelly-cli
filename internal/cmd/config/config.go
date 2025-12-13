// Package config provides device configuration commands.
package config

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/config/diff"
	configexport "github.com/tj-smith47/shelly-cli/internal/cmd/config/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/get"
	configimport "github.com/tj-smith47/shelly-cli/internal/cmd/config/import"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/reset"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/set"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the config command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"cfg"},
		Short:   "Manage device configuration",
		Long: `Manage device configuration settings.

Get, set, export, and import device configurations. Configuration includes
component settings, system parameters, and feature configurations.`,
		Example: `  # Get full device configuration
  shelly config get living-room

  # Get specific component configuration
  shelly config get living-room switch:0

  # Set configuration values
  shelly config set living-room switch:0 name="Main Light"

  # Export configuration to file
  shelly config export living-room config.json

  # Import configuration from file
  shelly config import living-room config.json --dry-run

  # Compare configuration with a file
  shelly config diff living-room config.json

  # Reset configuration to defaults
  shelly config reset living-room switch:0`,
	}

	cmd.AddCommand(get.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(diff.NewCommand(f))
	cmd.AddCommand(configexport.NewCommand(f))
	cmd.AddCommand(configimport.NewCommand(f))
	cmd.AddCommand(reset.NewCommand(f))

	return cmd
}
