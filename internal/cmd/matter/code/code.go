// Package code provides the matter code command.
package code

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the matter code command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "code <device>",
		Aliases: []string{"pairing", "qr"},
		Short:   "Show Matter pairing code",
		Long: `Show the Matter pairing code for commissioning a device.

Displays the commissioning information needed to add the device
to a Matter fabric (Apple Home, Google Home, etc.):
- Manual pairing code (11-digit number)
- QR code data (for compatible apps)
- Discriminator and setup PIN

If the pairing code is not available via the API, check the device
label or web UI at http://<device-ip>/matter for the QR code.`,
		Example: `  # Show pairing code
  shelly matter code living-room

  # Output as JSON
  shelly matter code living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Resolve device IP for display
	deviceIP := ""
	if devCfg, ok := opts.Factory.ResolveDevice(opts.Device); ok && devCfg.Address != "" {
		deviceIP = devCfg.Address
	}

	// Check if device is commissionable
	commissionable, err := svc.Wireless().MatterIsCommissionable(ctx, opts.Device)
	if err != nil {
		return err
	}

	if !commissionable {
		ios.Warning("Device is not commissionable.")
		ios.Info("Enable Matter first: shelly matter enable %s", opts.Device)
		if opts.JSON {
			return output.JSON(ios.Out, model.CommissioningInfo{Available: false})
		}
		return nil
	}

	// Get commissioning code
	info, err := svc.Wireless().MatterGetCommissioningCode(ctx, opts.Device)
	if err != nil {
		ios.Debug("failed to get commissioning code: %v", err)
		// Code not available via API, show instructions
		if opts.JSON {
			return output.JSON(ios.Out, model.CommissioningInfo{Available: false})
		}
		term.DisplayNotAvailable(ios, deviceIP)
		return nil
	}

	if opts.JSON {
		return output.JSON(ios.Out, info)
	}

	term.DisplayCommissioningInfo(ios, info, deviceIP)
	return nil
}
