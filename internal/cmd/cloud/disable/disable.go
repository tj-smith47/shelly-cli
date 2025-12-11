// Package disable provides the cloud disable subcommand.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cloud disable command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <device>",
		Short: "Disable cloud connection",
		Long: `Disable the Shelly Cloud connection for a device.

Once disabled, the device will only be accessible via local network.`,
		Example: `  # Disable cloud connection
  shelly cloud disable living-room`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunWithSpinner(ctx, ios, "Disabling cloud connection...", func(ctx context.Context) error {
		if err := svc.SetCloudEnabled(ctx, device, false); err != nil {
			return fmt.Errorf("failed to disable cloud: %w", err)
		}
		ios.Success("Cloud connection disabled on %s", device)
		return nil
	})
}
