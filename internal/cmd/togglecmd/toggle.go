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
	Device      string
	ComponentID int // -1 means all components
	Factory     *cmdutil.Factory
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

By default, toggles all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).`,
		Example: `  # Toggle all components on a device
  shelly toggle living-room

  # Toggle specific switch (for multi-switch devices)
  shelly toggle dual-switch --id 1

  # Toggle a cover
  shelly toggle bedroom-blinds`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ComponentID, "id", -1, "Component ID to control (omit to control all)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Convert -1 (default) to nil (all components), otherwise pass the ID
	var componentID *int
	if opts.ComponentID >= 0 {
		componentID = &opts.ComponentID
	}

	var result *shelly.QuickResult
	err := cmdutil.RunWithSpinner(ctx, ios, "Toggling...", func(ctx context.Context) error {
		var opErr error
		result, opErr = svc.QuickToggle(ctx, opts.Device, componentID)
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
