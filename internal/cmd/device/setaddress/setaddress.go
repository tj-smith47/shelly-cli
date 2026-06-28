// Package setaddress provides the device set-address subcommand.
package setaddress

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Name     string
	Address  string
	NoVerify bool
}

// NewCommand creates the device set-address command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set-address <name> <address>",
		Aliases: []string{"set-addr", "readdress"},
		Short:   "Update a registered device's IP address",
		Long: `Change the address of a device already in the registry.

Use this when a device moves to a new IP (DHCP lease change, subnet move, or a
manual re-IP). Only the address changes — the device's name, generation, model,
auth, and every group membership are preserved. This is the safe alternative to
'device remove' + 'device add', which would drop the device from its groups.

The new address is verified by default; pass --no-verify to pre-stage an address
for a device that is not yet reachable there.`,
		Example: `  # Re-point a device at its new IP
  shelly device set-address guest-bath 10.23.47.219

  # Pre-stage an address without a reachability check
  shelly device set-address guest-bath 10.23.47.219 --no-verify

  # Short form
  shelly dev set-addr bedroom 192.168.1.42`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			opts.Address = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.NoVerify, "no-verify", false, "Skip the reachability check at the new address")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	dev, exists := config.GetDevice(opts.Name)
	if !exists {
		return fmt.Errorf("device %q not found", opts.Name)
	}

	if dev.Address == opts.Address {
		ios.Info("%q is already at %s", opts.Name, opts.Address)
		return nil
	}

	if !opts.NoVerify {
		svc := opts.Factory.ShellyService()
		ios.StartProgress("Verifying device at new address...")
		_, err := svc.DeviceInfoAuto(ctx, opts.Address)
		ios.StopProgress()
		if err != nil {
			return fmt.Errorf("couldn't reach a device at %s: %w (use --no-verify to set it anyway)", opts.Address, err)
		}
	}

	if err := config.UpdateDeviceAddress(opts.Name, opts.Address); err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	ios.Success("%q is now at %s (was %s)", opts.Name, opts.Address, dev.Address)
	return nil
}
