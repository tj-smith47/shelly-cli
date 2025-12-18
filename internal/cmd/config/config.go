// Package config provides CLI configuration management commands.
package config

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/config/edit"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/get"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/path"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/reset"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/config/show"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the config command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"cfg"},
		Short:   "Manage CLI configuration",
		Long: `Manage Shelly CLI configuration settings.

Get, set, and edit CLI preferences like default timeout, output format,
and theme settings.

For device configuration, use: shelly device config <subcommand>`,
		Example: `  # View all CLI settings
  shelly config get

  # Get specific setting
  shelly config get defaults.timeout

  # Set a value
  shelly config set defaults.output=json

  # Open config in editor
  shelly config edit

  # Reset to defaults
  shelly config reset`,
	}

	cmd.AddCommand(edit.NewCommand(f))
	cmd.AddCommand(get.NewCommand(f))
	cmd.AddCommand(path.NewCommand(f))
	cmd.AddCommand(reset.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(show.NewCommand(f))

	return cmd
}
