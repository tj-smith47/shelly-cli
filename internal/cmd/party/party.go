// Package party provides the party command for fun light effects.
package party

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	flags.DeviceTargetFlags
	Duration time.Duration
	Interval time.Duration
}

// NewCommand creates the party command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Duration: 30 * time.Second,
		Interval: 500 * time.Millisecond,
	}

	cmd := &cobra.Command{
		Use:     "party [device...]",
		Aliases: []string{"disco", "strobe", "rave"},
		Short:   "Party mode - flash lights!",
		Long: `Start a party mode that rapidly toggles lights.

This is a fun command that makes your lights flash for a set duration.
For RGB lights, it cycles through random colors.

Use Ctrl+C to stop early.`,
		Example: `  # Party with all devices for 30 seconds
  shelly party --all

  # Party with specific devices for 1 minute
  shelly party light-1 light-2 -d 1m

  # Fast strobe effect (200ms interval)
  shelly party --all -i 200ms`,
		RunE: func(cmd *cobra.Command, args []string) error {
			devices := args
			if opts.All {
				registered := config.ListDevices()
				if len(registered) == 0 {
					f.IOStreams().Warning("No devices registered. Run 'shelly discover mdns --register' first.")
					return nil
				}
				devices = make([]string, 0, len(registered))
				for name := range registered {
					devices = append(devices, name)
				}
			} else if len(args) == 0 {
				return fmt.Errorf("specify device(s) or use --all")
			}
			return run(cmd.Context(), f, devices, opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Duration, "duration", "d", 30*time.Second, "Party duration")
	cmd.Flags().DurationVarP(&opts.Interval, "interval", "i", 500*time.Millisecond, "Toggle interval")
	flags.AddAllOnlyFlag(cmd, &opts.DeviceTargetFlags)

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.Success("Party mode starting!")
	ios.Info("Duration: %v, Interval: %v", opts.Duration, opts.Interval)
	ios.Info("Devices: %d", len(devices))
	ios.Println("")
	ios.Info("Press Ctrl+C to stop...")
	ios.Println("")

	// Create a context with timeout for the party duration
	partyCtx, cancel := context.WithTimeout(ctx, opts.Duration)
	defer cancel()

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	toggleState := false

	for {
		select {
		case <-partyCtx.Done():
			ios.Println("")
			ios.Success("Party's over!")

			// Turn lights back on
			for _, device := range devices {
				if err := svc.LightOn(ctx, device, 0); err != nil {
					// Silently ignore - might not be a light, try switch
					if switchErr := svc.SwitchOn(ctx, device, 0); switchErr != nil {
						ios.DebugErr("party cleanup for "+device, switchErr)
					}
				}
			}
			return nil

		case <-ticker.C:
			toggleState = !toggleState

			for _, device := range devices {
				go func(dev string) {
					svc.PartyToggleDevice(partyCtx, ios, dev, toggleState)
				}(device)
			}
		}
	}
}
