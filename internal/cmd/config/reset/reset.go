// Package reset provides the config reset subcommand for CLI settings.
package reset

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

var yesFlag bool

// NewCommand creates the config reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(f)
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	confirmed, err := f.ConfirmAction("Reset CLI configuration to defaults?", yesFlag)
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
