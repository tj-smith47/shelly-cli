// Package config provides device configuration subcommands under device.
package config

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/device/config/diff"
	configexport "github.com/tj-smith47/shelly-cli/internal/cmd/device/config/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/config/get"
	configimport "github.com/tj-smith47/shelly-cli/internal/cmd/device/config/importcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/config/reset"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/config/set"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the device config command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	getCmd := get.NewCommand(f)

	cmd := &cobra.Command{
		Use:     "config [device] [component]",
		Aliases: []string{"cfg"},
		Short:   "Manage device configuration",
		Long: `Manage device configuration settings.

When called with a device argument (and optional component), delegates to
"config get" to show the device configuration.

Get, set, export, and import device configurations. Configuration includes
component settings, system parameters, and feature configurations.`,
		Example: `  # Get full device configuration (shorthand)
  shelly device config living-room

  # Get specific component configuration (shorthand)
  shelly device config living-room switch:0

  # Get full device configuration (explicit)
  shelly device config get living-room

  # Set configuration values
  shelly device config set living-room switch:0 name="Main Light"

  # Export configuration to file
  shelly device config export living-room config.json

  # Import configuration from file
  shelly device config import living-room config.json --dry-run

  # Compare configuration with a file
  shelly device config diff living-room config.json

  # Reset configuration to defaults
  shelly device config reset living-room switch:0`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			// Delegate to "get" subcommand
			getCmd.SetArgs(args)
			getCmd.SetContext(cmd.Context())
			return getCmd.Execute()
		},
	}

	cmd.AddCommand(diff.NewCommand(f))
	cmd.AddCommand(configexport.NewCommand(f))
	cmd.AddCommand(getCmd)
	cmd.AddCommand(configimport.NewCommand(f))
	cmd.AddCommand(reset.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))

	return cmd
}
