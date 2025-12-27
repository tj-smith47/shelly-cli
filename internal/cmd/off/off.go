// Package off provides the quick off command.
package off

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

// NewCommand creates the off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "off <device>",
		Aliases: []string{"turn-off", "disable"},
		Short:   "Turn off a device (auto-detects type)",
		Long: `Turn off a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this closes them. For switches/lights/RGB, this turns them off.

By default, turns off all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).`,
		Example: `  # Turn off all components on a device
  shelly off living-room

  # Turn off specific switch (for multi-switch devices)
  shelly off dual-switch --id 1

  # Close a cover
  shelly off bedroom-blinds`,
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
	err := cmdutil.RunWithSpinner(ctx, ios, "Turning off...", func(ctx context.Context) error {
		var opErr error
		result, opErr = svc.QuickOff(ctx, opts.Device, opts.ComponentIDPointer())
		return opErr
	})
	if err != nil {
		return err
	}

	if result.Count == 1 {
		ios.Success("Device %q turned off", opts.Device)
	} else {
		ios.Success("Turned off %d components on %q", result.Count, opts.Device)
	}
	return nil
}
