// Package set provides the action set command.
package set

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
	URL     string
	Index   int
	Enabled bool
	Factory *cmdutil.Factory
}

// NewCommand creates the action set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f, Enabled: true}

	cmd := &cobra.Command{
		Use:     "set <device> <event> <url>",
		Aliases: []string{"add", "configure"},
		Short:   "Set an action URL for a Gen1 device",
		Long: `Set an action URL for a Gen1 Shelly device.

Gen1 devices support various action event types:
  Output events:    out_on_url, out_off_url
  Button events:    btn1_on_url, btn1_off_url, btn2_on_url, btn2_off_url
  Input events:     input_on_url, input_off_url
  Push events:      longpush_url, shortpush_url, double_shortpush_url, triple_shortpush_url
  Roller events:    roller_open_url, roller_close_url, roller_stop_url
  Sensor events:    motion_url, no_motion_url, flood_detected_url, etc.
  System events:    overpower_url, overvoltage_url, overtemperature_url

Gen2+ devices use webhooks instead. See 'shelly webhook create'.`,
		Example: `  # Set output on action
  shelly action set living-room out_on_url "http://homeserver/api/light-on"

  # Set button long press action
  shelly action set switch longpush_url "http://homeserver/api/dim-lights"

  # Set action at a specific index (for multi-channel devices)
  shelly action set relay out_on_url "http://server/trigger" --index 1

  # Set action but leave it disabled
  shelly action set switch out_on_url "http://server/test" --disabled`,
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Event = args[1]
			opts.URL = args[2]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Index, "index", 0, "Action index (for multi-channel devices)")
	cmd.Flags().BoolVar(&opts.Enabled, "enabled", true, "Enable the action")
	cmd.Flags().BoolVar(&opts.Enabled, "disabled", false, "Disable the action (same as --enabled=false)")

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

	// Set the action URL
	err := svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if !dev.IsGen1() {
			ios.Warning("Device %s is not a Gen1 device", opts.Device)
			ios.Info("Gen2+ devices use webhooks. Try: shelly webhook create %s", opts.Device)
			return fmt.Errorf("action URLs only available for Gen1 devices")
		}
		return dev.Gen1().SetActionURL(ctx, opts.Index, event, opts.URL, opts.Enabled)
	})
	if err != nil {
		return err
	}

	status := "enabled"
	if !opts.Enabled {
		status = "disabled"
	}

	ios.Success("Action %s set on %s", opts.Event, opts.Device)
	ios.Info("  URL: %s", opts.URL)
	ios.Info("  Index: %d", opts.Index)
	ios.Info("  Status: %s", status)

	return nil
}
