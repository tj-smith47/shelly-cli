// Package wake provides the wake command for turning devices on after a delay.
package wake

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

const simulateSunrise = "sunrise"

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Delay    time.Duration
	Simulate string
	Duration time.Duration
	// RampInterval is the shortest time between two brightness steps of a
	// sunrise ramp.
	RampInterval time.Duration
}

// NewCommand creates the wake command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:      f,
		Delay:        5 * time.Minute,
		RampInterval: 5 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "wake <device>",
		Aliases: []string{simulateSunrise, "wakeup"},
		Short:   "Turn device on after a delay",
		Long: `Turn a device on after a specified delay.

With --simulate sunrise, the device's lights turn on at 1% brightness when the
delay ends and rise to 100% over --duration (default 15m). Sunrise needs a
device with a dimmable light component (a dimmer, bulb, or light in white
mode); a device with only switches or relays is rejected with an error.

Useful for:
  - Waking up to lights
  - Scheduling devices to turn on
  - "Good morning" automation

Press Ctrl+C to cancel before the delay expires, or to stop a sunrise at its
current brightness.`,
		Example: `  # Turn on in 5 minutes (default)
  shelly wake bedroom-light

  # Turn on in 7 hours (alarm)
  shelly wake living-room -d 7h

  # Turn on in 30 seconds
  shelly wake kitchen --delay 30s

  # In 7 hours, fade the bedroom light up from 1% to 100% over 15 minutes
  shelly wake bedroom-light -d 7h --simulate sunrise --duration 15m`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if opts.Simulate == "" && cmd.Flags().Changed("duration") {
				return fmt.Errorf("--duration sets the length of a --simulate ramp; add --simulate %s or drop --duration", simulateSunrise)
			}
			if opts.Simulate != "" && opts.Simulate != simulateSunrise {
				return fmt.Errorf("unsupported --simulate %q (supported: %s)", opts.Simulate, simulateSunrise)
			}
			if opts.Simulate != "" && opts.Duration <= 0 {
				return fmt.Errorf("--duration must be greater than zero, got %v", opts.Duration)
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Delay, "delay", "d", 5*time.Minute, "Delay before turning on")
	cmd.Flags().StringVar(&opts.Simulate, "simulate", "", "Fade the light in instead of switching it on (supported: sunrise)")
	cmd.Flags().DurationVar(&opts.Duration, "duration", 15*time.Minute, "How long a --simulate ramp takes to reach full brightness")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// The light check runs before the delay so a device that cannot dim fails
	// now, before the wait, while the user can still fix it.
	var lightIDs []int
	if opts.Simulate == simulateSunrise {
		ids, err := svc.LightIDs(ctx, opts.Device)
		if err != nil {
			return fmt.Errorf("failed to read light components of %s: %w", opts.Device, err)
		}
		if len(ids) == 0 {
			return fmt.Errorf("--simulate sunrise needs a dimmable light, and %s has no light component with brightness control (switches and relays can only turn on or off)", opts.Device)
		}
		lightIDs = ids
	}

	ios.Info("Wake timer set for %s", opts.Device)
	if lightIDs != nil {
		ios.Info("Sunrise starts in %v and reaches full brightness %v after that", opts.Delay, opts.Duration)
	} else {
		ios.Info("Device will turn on in %v", opts.Delay)
	}
	ios.Println("")
	ios.Info("Press Ctrl+C to cancel...")

	select {
	case <-ctx.Done():
		ios.Println("")
		ios.Warning("Wake timer cancelled")
		return nil
	case <-time.After(opts.Delay):
	}

	ios.Println("")

	if lightIDs != nil {
		ios.Info("Sunrise: raising %s from 1%% to 100%% over %v...", opts.Device, opts.Duration)
		level, err := svc.LightRamp(ctx, opts.Device, lightIDs, 1, 100, opts.Duration, opts.RampInterval)
		select {
		case <-ctx.Done():
			ios.Warning("Sunrise stopped at %d%%", level)
			return nil
		default:
		}
		if err != nil {
			return fmt.Errorf("sunrise failed at %d%%: %w", level, err)
		}
		ios.Success("Good morning! %s is at full brightness", opts.Device)
		return nil
	}

	ios.Info("Turning on %s...", opts.Device)

	result, err := svc.QuickOn(ctx, opts.Device, nil)
	if err != nil {
		return fmt.Errorf("failed to turn on device: %w", err)
	}

	if result != nil && result.Count > 0 {
		ios.Success("Good morning! %s is now on", opts.Device)
	} else {
		ios.Info("Command sent to %s", opts.Device)
	}

	return nil
}
