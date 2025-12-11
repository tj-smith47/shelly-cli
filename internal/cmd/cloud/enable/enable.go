// Package enable provides the cloud enable subcommand.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cloud enable command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <device>",
		Short: "Enable cloud connection",
		Long: `Enable the Shelly Cloud connection for a device.

Once enabled, the device will connect to Shelly Cloud for remote access
and monitoring.`,
		Example: `  # Enable cloud connection
  shelly cloud enable living-room`,
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

	return cmdutil.RunWithSpinner(ctx, ios, "Enabling cloud connection...", func(ctx context.Context) error {
		if err := svc.SetCloudEnabled(ctx, device, true); err != nil {
			return fmt.Errorf("failed to enable cloud: %w", err)
		}
		ios.Success("Cloud connection enabled on %s", device)
		return nil
	})
}
