// Package sleep provides the sleep command for turning devices off after a delay.
package sleep

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	Delay time.Duration
}

// NewCommand creates the sleep command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Delay: 5 * time.Minute,
	}

	cmd := &cobra.Command{
		Use:     "sleep <device>",
		Aliases: []string{"goodnight", "nap"},
		Short:   "Turn device off after a delay",
		Long: `Turn a device off after a specified delay.

Useful for:
  - Setting a sleep timer for lights
  - Scheduling devices to turn off
  - "Goodnight" automation

Press Ctrl+C to cancel before the delay expires.`,
		Example: `  # Turn off in 5 minutes (default)
  shelly sleep bedroom-light

  # Turn off in 30 minutes
  shelly sleep living-room -d 30m

  # Turn off in 1 hour
  shelly sleep hallway --delay 1h`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Delay, "delay", "d", 5*time.Minute, "Delay before turning off")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.Info("Sleep timer set for %s", device)
	ios.Info("Device will turn off in %v", opts.Delay)
	ios.Println("")
	ios.Info("Press Ctrl+C to cancel...")

	// Wait for the delay
	select {
	case <-ctx.Done():
		ios.Println("")
		ios.Warning("Sleep timer cancelled")
		return nil
	case <-time.After(opts.Delay):
		// Timer expired, turn off
	}

	ios.Println("")
	ios.Info("Turning off %s...", device)

	// Try QuickOff first (works for most devices)
	result, err := svc.QuickOff(ctx, device, nil)
	if err != nil {
		return fmt.Errorf("failed to turn off device: %w", err)
	}

	if result != nil && result.Count > 0 {
		ios.Success("Goodnight! %s is now off", device)
	} else {
		ios.Info("Command sent to %s", device)
	}

	return nil
}
