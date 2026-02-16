// Package disable provides the cloud disable subcommand.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the cloud disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewEnableDisableCommand(f, factories.EnableDisableOpts{
		Feature: "cloud connection",
		Enable:  false,
		Aliases: []string{"off", "disconnect"},
		Long: `Disable the Shelly Cloud connection for a device.

Once disabled, the device will only be accessible via local network.`,
		Example: `  # Disable cloud connection
  shelly cloud disable living-room`,
		ServiceFunc: func(ctx context.Context, f *cmdutil.Factory, device string) error {
			if err := f.ShellyService().SetCloudEnabled(ctx, device, false); err != nil {
				return fmt.Errorf("failed to disable cloud: %w", err)
			}
			return nil
		},
	})
}
