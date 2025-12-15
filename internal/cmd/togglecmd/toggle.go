// Package togglecmd provides the quick toggle command.
package togglecmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	All     bool
	Factory *cmdutil.Factory
}

// NewCommand creates the toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "toggle <device>",
		Aliases: []string{"flip", "switch"},
		Short:   "Toggle a device (auto-detects type)",
		Long: `Toggle a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this toggles between open and close based on current state.

Use --all to toggle all controllable components on the device.`,
		Example: `  # Toggle a switch or light
  shelly toggle living-room

  # Toggle all components on a device
  shelly toggle living-room --all

  # Toggle a cover
  shelly toggle bedroom-blinds`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Toggle all controllable components")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var result *shelly.QuickResult
	err := cmdutil.RunWithSpinner(ctx, ios, "Toggling...", func(ctx context.Context) error {
		var opErr error
		result, opErr = svc.QuickToggle(ctx, opts.Device, opts.All)
		return opErr
	})
	if err != nil {
		return err
	}

	if result.Count == 1 {
		ios.Success("Device %q toggled", opts.Device)
	} else {
		ios.Success("Toggled %d components on %q", result.Count, opts.Device)
	}
	return nil
}
