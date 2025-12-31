// Package list provides the action list command.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen1"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the action list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "show"},
		Short:   "List action URLs for a Gen1 device",
		Long: `List all configured action URLs for a Gen1 Shelly device.

Gen1 devices support various action types that trigger HTTP callbacks:
  - out_on_url, out_off_url: Output state change actions
  - btn_on_url, btn_off_url: Button toggle actions
  - longpush_url, shortpush_url: Button press duration actions
  - roller_open_url, roller_close_url, roller_stop_url: Roller actions

Gen2+ devices use webhooks instead. See 'shelly webhook list'.`,
		Example: `  # List actions for a device
  shelly action list living-room

  # JSON output
  shelly action list living-room -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var actions *gen1.ActionSettings
	err := svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if !dev.IsGen1() {
			ios.Warning("Device %s is not a Gen1 device", opts.Device)
			ios.Info("Gen2+ devices use webhooks. Try: shelly webhook list %s", opts.Device)
			return fmt.Errorf("action URLs only available for Gen1 devices")
		}

		var getErr error
		actions, getErr = dev.Gen1().GetActions(ctx)
		return getErr
	})
	if err != nil {
		return err
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, actions)
	}

	term.DisplayGen1Actions(ios, actions)
	return nil
}
