// Package party provides the party command for fun light effects.
package party

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Duration time.Duration
	Interval time.Duration
	All      bool
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
			if opts.All {
				return runAll(cmd.Context(), f, opts)
			}
			if len(args) == 0 {
				return fmt.Errorf("specify device(s) or use --all")
			}
			return run(cmd.Context(), f, args, opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Duration, "duration", "d", 30*time.Second, "Party duration")
	cmd.Flags().DurationVarP(&opts.Interval, "interval", "i", 500*time.Millisecond, "Toggle interval")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Party with all registered devices")

	return cmd
}

func runAll(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	devices := config.ListDevices()
	if len(devices) == 0 {
		ios := f.IOStreams()
		ios.Warning("No devices registered. Run 'shelly discover mdns --register' first.")
		return nil
	}

	deviceList := make([]string, 0, len(devices))
	for name := range devices {
		deviceList = append(deviceList, name)
	}

	return run(ctx, f, deviceList, opts)
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

	// Random colors for RGB lights
	colors := []struct{ r, g, b int }{
		{255, 0, 0},     // Red
		{0, 255, 0},     // Green
		{0, 0, 255},     // Blue
		{255, 255, 0},   // Yellow
		{255, 0, 255},   // Magenta
		{0, 255, 255},   // Cyan
		{255, 128, 0},   // Orange
		{128, 0, 255},   // Purple
		{255, 255, 255}, // White
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	toggleState := false
	colorIndex := 0

	for {
		select {
		case <-partyCtx.Done():
			ios.Println("")
			ios.Success("Party's over!")

			// Turn lights back on
			for _, device := range devices {
				if err := svc.LightOn(ctx, device, 0); err != nil {
					// Silently ignore - might not be a light
					_ = svc.SwitchOn(ctx, device, 0) //nolint:errcheck // Best-effort cleanup
				}
			}
			return nil

		case <-ticker.C:
			toggleState = !toggleState

			for _, device := range devices {
				go func(dev string) {
					// Try as light first
					if toggleState {
						if err := svc.LightOn(partyCtx, dev, 0); err != nil {
							// Try as switch
							_ = svc.SwitchOn(partyCtx, dev, 0) //nolint:errcheck // Fallback, ignore
						}
					} else {
						if err := svc.LightOff(partyCtx, dev, 0); err != nil {
							// Try as switch
							_ = svc.SwitchOff(partyCtx, dev, 0) //nolint:errcheck // Fallback, ignore
						}
					}

					// Try to set random color for RGB lights (ignore errors for non-RGB)
					if toggleState {
						color := colors[rand.Intn(len(colors))]                       //nolint:gosec // Not crypto, just random colors
						_ = svc.RGBColor(partyCtx, dev, 0, color.r, color.g, color.b) //nolint:errcheck // Optional RGB
					}
				}(device)
			}

			colorIndex = (colorIndex + 1) % len(colors)
		}
	}
}
