// Package on provides the quick on command.
package on

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	flags.QuickComponentFlags
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "on <device>",
		Aliases: []string{"turn-on", "enable"},
		Short:   "Turn on a device (auto-detects type)",
		Long: `Turn on a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this opens them. For switches/lights/RGB, this turns them on.

By default, turns on all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).`,
		Example: `  # Turn on all components on a device
  shelly on living-room

  # Turn on specific switch (for multi-switch devices)
  shelly on dual-switch --id 1

  # Open a cover
  shelly on bedroom-blinds`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddQuickComponentFlags(cmd, &opts.QuickComponentFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	var result *shelly.QuickResult
	err := cmdutil.RunWithSpinner(ctx, ios, "Turning on...", func(ctx context.Context) error {
		var opErr error
		result, opErr = svc.QuickOn(ctx, opts.Device, opts.ComponentIDPointer())
		return opErr
	})
	if err != nil {
		return err
	}

	if result.Count == 1 {
		ios.Success("Device %q turned on", opts.Device)
	} else {
		ios.Success("Turned on %d components on %q", result.Count, opts.Device)
	}
	return nil
}
