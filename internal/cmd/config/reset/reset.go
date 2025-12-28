// Package reset provides the config reset subcommand for CLI settings.
package reset

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory *cmdutil.Factory
}

// NewCommand creates the config reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "reset",
		Aliases: []string{"clear"},
		Short:   "Reset CLI configuration to defaults",
		Long: `Reset the Shelly CLI configuration to default values.

This clears all custom settings from the config file. Device registrations,
aliases, and other data are preserved.`,
		Example: `  # Reset with confirmation
  shelly config reset

  # Reset without confirmation
  shelly config reset --yes`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	confirmed, err := opts.Factory.ConfirmAction("Reset CLI configuration to defaults?", opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Warning("Reset cancelled")
		return nil
	}

	if err := config.ResetSettings(); err != nil {
		return err
	}

	ios.Success("CLI configuration reset to defaults")
	return nil
}
