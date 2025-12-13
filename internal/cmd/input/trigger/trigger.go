// Package trigger provides the input trigger subcommand.
package trigger

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Event type constants.
const (
	EventSinglePush = "single_push"
	EventDoublePush = "double_push"
	EventLongPush   = "long_push"
)

// NewCommand creates the input trigger command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var inputID int
	var event string

	cmd := &cobra.Command{
		Use:   "trigger <device>",
		Short: "Trigger input event",
		Long: `Trigger an input event on a Shelly device.

Event types:
  single_push - Single button press
  double_push - Double button press
  long_push   - Long button press`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], inputID, event)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &inputID, "Input")
	cmd.Flags().StringVarP(&event, "event", "e", EventSinglePush, "Event type (single_push, double_push, long_push)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, inputID int, event string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Triggering input event...")

	err := svc.InputTrigger(ctx, device, inputID, event)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to trigger input event: %w", err)
	}

	ios.Success("Input %d triggered with event %q", inputID, event)
	return nil
}
