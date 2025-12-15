// Package wake provides the wake command for turning devices on after a delay.
package wake

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

// NewCommand creates the wake command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Delay: 5 * time.Minute,
	}

	cmd := &cobra.Command{
		Use:     "wake <device>",
		Aliases: []string{"sunrise", "wakeup"},
		Short:   "Turn device on after a delay",
		Long: `Turn a device on after a specified delay.

Useful for:
  - Waking up to lights
  - Scheduling devices to turn on
  - "Good morning" automation

Press Ctrl+C to cancel before the delay expires.`,
		Example: `  # Turn on in 5 minutes (default)
  shelly wake bedroom-light

  # Turn on in 7 hours (alarm)
  shelly wake living-room -d 7h

  # Turn on in 30 seconds
  shelly wake kitchen --delay 30s`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Delay, "delay", "d", 5*time.Minute, "Delay before turning on")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.Info("Wake timer set for %s", device)
	ios.Info("Device will turn on in %v", opts.Delay)
	ios.Println("")
	ios.Info("Press Ctrl+C to cancel...")

	// Wait for the delay
	select {
	case <-ctx.Done():
		ios.Println("")
		ios.Warning("Wake timer cancelled")
		return nil
	case <-time.After(opts.Delay):
		// Timer expired, turn on
	}

	ios.Println("")
	ios.Info("Turning on %s...", device)

	// Try QuickOn first (works for most devices)
	result, err := svc.QuickOn(ctx, device, false)
	if err != nil {
		return fmt.Errorf("failed to turn on device: %w", err)
	}

	if result != nil && result.Count > 0 {
		ios.Success("Good morning! %s is now on", device)
	} else {
		ios.Info("Command sent to %s", device)
	}

	return nil
}
