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
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
					toggleDevice(partyCtx, svc, ios, dev, toggleState, colors)
				}(device)
			}

			colorIndex = (colorIndex + 1) % len(colors)
		}
	}
}

// toggleDevice handles toggling a single device on or off with fallback to switch.
func toggleDevice(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, dev string, on bool, colors []struct{ r, g, b int }) {
	if on {
		toggleDeviceOn(ctx, svc, ios, dev, colors)
	} else {
		toggleDeviceOff(ctx, svc, ios, dev)
	}
}

// toggleDeviceOn turns a device on with light/switch fallback and sets random color.
func toggleDeviceOn(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, dev string, colors []struct{ r, g, b int }) {
	if err := svc.LightOn(ctx, dev, 0); err != nil {
		// Try as switch (expected to fail for non-switch devices)
		if switchErr := svc.SwitchOn(ctx, dev, 0); switchErr != nil {
			ios.DebugErr("party toggle on "+dev, switchErr)
		}
	}

	// Try to set random color for RGB lights (expected to fail for non-RGB)
	color := colors[rand.Intn(len(colors))] //nolint:gosec // Not crypto, just random colors
	if rgbErr := svc.RGBColor(ctx, dev, 0, color.r, color.g, color.b); rgbErr != nil {
		ios.DebugErr("party RGB "+dev, rgbErr)
	}
}

// toggleDeviceOff turns a device off with light/switch fallback.
func toggleDeviceOff(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, dev string) {
	if err := svc.LightOff(ctx, dev, 0); err != nil {
		// Try as switch (expected to fail for non-switch devices)
		if switchErr := svc.SwitchOff(ctx, dev, 0); switchErr != nil {
			ios.DebugErr("party toggle off "+dev, switchErr)
		}
	}
}
