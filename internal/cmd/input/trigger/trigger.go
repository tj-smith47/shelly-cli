// Package trigger provides the input trigger subcommand.
package trigger

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Event type constants.
const (
	EventSinglePush = "single_push"
	EventDoublePush = "double_push"
	EventLongPush   = "long_push"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory *cmdutil.Factory
	Device  string
	Event   string
}

// NewCommand creates the input trigger command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Event:   EventSinglePush,
	}

	cmd := &cobra.Command{
		Use:     "trigger <device>",
		Aliases: []string{"fire", "simulate"},
		Short:   "Trigger input event",
		Long: `Trigger an input event on a Shelly device.

Event types:
  single_push - Single button press
  double_push - Double button press
  long_push   - Long button press`,
		Example: `  # Trigger single push event
  shelly input trigger living-room

  # Trigger double push event
  shelly input trigger living-room --event double_push

  # Trigger long push on specific input
  shelly input trigger living-room --id 1 --event long_push`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Input")
	cmd.Flags().StringVarP(&opts.Event, "event", "e", EventSinglePush, "Event type (single_push, double_push, long_push)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	err := cmdutil.RunWithSpinner(ctx, ios, "Triggering input event...", func(ctx context.Context) error {
		return svc.InputTrigger(ctx, opts.Device, opts.ID, opts.Event)
	})
	if err != nil {
		return fmt.Errorf("failed to trigger input event: %w", err)
	}

	ios.Success("Input %d triggered with event %q", opts.ID, opts.Event)
	return nil
}
