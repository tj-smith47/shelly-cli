// Package boost provides the thermostat boost command.
package boost

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Duration time.Duration
	Cancel   bool
}

// NewCommand creates the thermostat boost command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "boost <device>",
		Aliases: []string{"turbo", "rapid"},
		Short:   "Activate boost mode",
		Long: `Activate boost mode on a thermostat.

Boost mode sets the valve to 100% for rapid heating. This is useful
when you need to quickly warm up a room.

Use --duration to specify how long boost mode should last.
Use --cancel to cancel an active boost.`,
		Example: `  # Activate boost with device default duration
  shelly thermostat boost gateway

  # Boost for 5 minutes
  shelly thermostat boost gateway --duration 5m

  # Boost for 30 minutes
  shelly thermostat boost gateway --duration 30m

  # Cancel active boost
  shelly thermostat boost gateway --cancel`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Thermostat component ID")
	cmd.Flags().DurationVarP(&opts.Duration, "duration", "d", 0, "Boost duration (e.g., 5m, 30m, 1h)")
	cmd.Flags().BoolVar(&opts.Cancel, "cancel", false, "Cancel active boost")

	cmd.MarkFlagsMutuallyExclusive("duration", "cancel")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	thermostat := conn.Thermostat(opts.ID)

	if opts.Cancel {
		err = cmdutil.RunWithSpinner(ctx, ios, "Cancelling boost mode...", func(ctx context.Context) error {
			return thermostat.CancelBoost(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to cancel boost: %w", err)
		}

		ios.Success("Boost mode cancelled on thermostat %d", opts.ID)
		return nil
	}

	// Activate boost
	durationSec := int(opts.Duration.Seconds())

	err = cmdutil.RunWithSpinner(ctx, ios, "Activating boost mode...", func(ctx context.Context) error {
		return thermostat.Boost(ctx, durationSec)
	})
	if err != nil {
		return fmt.Errorf("failed to activate boost: %w", err)
	}

	if durationSec > 0 {
		ios.Success("Boost mode activated on thermostat %d for %s", opts.ID, opts.Duration)
	} else {
		ios.Success("Boost mode activated on thermostat %d (device default duration)", opts.ID)
	}

	ios.Info("Valve is now at 100%% for rapid heating")
	ios.Info("Cancel with: shelly thermostat boost %s --cancel", opts.Device)

	return nil
}
