// Package enable provides the matter enable command.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the matter enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on", "activate"},
		Short:   "Enable Matter on a device",
		Long: `Enable Matter connectivity on a Shelly device.

When Matter is enabled, the device can be commissioned (added) to
Matter-compatible smart home ecosystems like Apple Home, Google Home,
or Amazon Alexa.

After enabling, use 'shelly matter code' to get the pairing code
for commissioning the device.`,
		Example: `  # Enable Matter
  shelly matter enable living-room

  # Then get pairing code
  shelly matter code living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		_, err := conn.Call(ctx, "Matter.SetConfig", map[string]any{
			"config": map[string]any{
				"enable": true,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to enable Matter: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	ios.Success("Matter enabled.")
	ios.Println()
	ios.Info("The device is now commissionable.")
	ios.Info("Get pairing code with: shelly matter code %s", opts.Device)

	return nil
}
