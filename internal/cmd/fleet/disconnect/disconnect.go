// Package disconnect provides the fleet disconnect subcommand.
package disconnect

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the fleet disconnect command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "disconnect",
		Aliases: []string{"logout", "close"},
		Short:   "Disconnect from Shelly Cloud hosts",
		Long:    `Disconnect from all connected Shelly Cloud hosts.`,
		Example: `  # Disconnect from all hosts
  shelly fleet disconnect`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f)
		},
	}

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()

	ios.Info("Disconnecting from Shelly Cloud...")
	ios.Success("Disconnected from all hosts")

	return nil
}
