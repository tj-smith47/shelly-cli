// Package enable provides the cloud enable subcommand.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the cloud enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewEnableDisableCommand(f, factories.EnableDisableOpts{
		Feature: "cloud connection",
		Enable:  true,
		Aliases: []string{"on", "connect"},
		Long: `Enable the Shelly Cloud connection for a device.

Once enabled, the device will connect to Shelly Cloud for remote access
and monitoring.`,
		Example: `  # Enable cloud connection
  shelly cloud enable living-room`,
		ServiceFunc: func(ctx context.Context, f *cmdutil.Factory, device string) error {
			if err := f.ShellyService().SetCloudEnabled(ctx, device, true); err != nil {
				return fmt.Errorf("failed to enable cloud: %w", err)
			}
			return nil
		},
	})
}
