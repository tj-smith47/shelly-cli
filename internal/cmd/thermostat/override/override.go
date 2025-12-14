// Package override provides the thermostat override command.
package override

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Target   float64
	Duration time.Duration
	Cancel   bool
}

// NewCommand creates the thermostat override command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "override <device>",
		Aliases: []string{"temp-override", "manual"},
		Short:   "Override target temperature",
		Long: `Temporarily override the target temperature.

Override mode sets a different target temperature for a specified
duration. After the override expires, the thermostat returns to
its normal schedule.

This is useful for temporary temperature adjustments without
modifying the permanent schedule.`,
		Example: `  # Override to 25°C for 30 minutes
  shelly thermostat override gateway --target 25 --duration 30m

  # Override to 20°C for 2 hours
  shelly thermostat override gateway --target 20 --duration 2h

  # Override with device defaults
  shelly thermostat override gateway

  # Cancel active override
  shelly thermostat override gateway --cancel`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Thermostat component ID")
	cmd.Flags().Float64VarP(&opts.Target, "target", "t", 0, "Target temperature in Celsius")
	cmd.Flags().DurationVarP(&opts.Duration, "duration", "d", 0, "Override duration (e.g., 30m, 2h)")
	cmd.Flags().BoolVar(&opts.Cancel, "cancel", false, "Cancel active override")

	cmd.MarkFlagsMutuallyExclusive("target", "cancel")
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
		ios.StartProgress("Cancelling temperature override...")
		err = thermostat.CancelOverride(ctx)
		ios.StopProgress()

		if err != nil {
			return fmt.Errorf("failed to cancel override: %w", err)
		}

		ios.Success("Temperature override cancelled on thermostat %d", opts.ID)
		ios.Info("Thermostat will return to normal schedule")
		return nil
	}

	// Activate override
	durationSec := int(opts.Duration.Seconds())

	ios.StartProgress("Activating temperature override...")
	err = thermostat.Override(ctx, opts.Target, durationSec)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to activate override: %w", err)
	}

	// Build success message
	parts := []string{}
	if opts.Target > 0 {
		parts = append(parts, fmt.Sprintf("target: %.1f°C", opts.Target))
	}
	if opts.Duration > 0 {
		parts = append(parts, fmt.Sprintf("duration: %s", opts.Duration))
	}

	if len(parts) > 0 {
		ios.Success("Temperature override activated on thermostat %d", opts.ID)
		for _, p := range parts {
			ios.Printf("  • %s\n", p)
		}
	} else {
		ios.Success("Temperature override activated on thermostat %d (device defaults)", opts.ID)
	}

	ios.Info("Cancel with: shelly thermostat override %s --cancel", opts.Device)

	return nil
}
