// Package enable provides the cloud enable subcommand.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the cloud enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on", "connect"},
		Short:   "Enable cloud connection",
		Long: `Enable the Shelly Cloud connection for a device.

Once enabled, the device will connect to Shelly Cloud for remote access
and monitoring.`,
		Example: `  # Enable cloud connection
  shelly cloud enable living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Enabling cloud connection...", func(ctx context.Context) error {
		if err := svc.SetCloudEnabled(ctx, device, true); err != nil {
			return fmt.Errorf("failed to enable cloud: %w", err)
		}
		ios.Success("Cloud connection enabled on %s", device)
		return nil
	})
}
