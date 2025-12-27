// Package clearcmd provides the action clear command.
package clearcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen1"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	Event   string
	Index   int
	Factory *cmdutil.Factory
}

// NewCommand creates the action clear command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "clear <device> <event>",
		Aliases: []string{"delete", "remove", "rm"},
		Short:   "Clear an action URL for a Gen1 device",
		Long: `Clear (remove) an action URL for a Gen1 Shelly device.

This removes the configured URL for the specified action, disabling the
HTTP callback for that event.

Gen1 devices support various action event types:
  Output events:    out_on_url, out_off_url
  Button events:    btn1_on_url, btn1_off_url, btn2_on_url, btn2_off_url
  Input events:     input_on_url, input_off_url
  Push events:      longpush_url, shortpush_url, double_shortpush_url, triple_shortpush_url
  Roller events:    roller_open_url, roller_close_url, roller_stop_url
  Sensor events:    motion_url, no_motion_url, flood_detected_url, etc.

Gen2+ devices use webhooks instead. See 'shelly webhook delete'.`,
		Example: `  # Clear output on action
  shelly action clear living-room out_on_url

  # Clear button long press action
  shelly action clear switch longpush_url

  # Clear action at a specific index
  shelly action clear relay out_on_url --index 1`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Event = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Index, "index", 0, "Action index (for multi-channel devices)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Parse the event type
	event := gen1.ActionEvent(opts.Event)

	// Clear the action
	err := svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if !dev.IsGen1() {
			ios.Warning("Device %s is not a Gen1 device", opts.Device)
			ios.Info("Gen2+ devices use webhooks. Try: shelly webhook delete %s", opts.Device)
			return fmt.Errorf("action URLs only available for Gen1 devices")
		}
		return dev.Gen1().ClearAction(ctx, opts.Index, event)
	})
	if err != nil {
		return err
	}

	ios.Success("Action %s cleared on %s", opts.Event, opts.Device)
	if opts.Index > 0 {
		ios.Info("  Index: %d", opts.Index)
	}

	return nil
}
