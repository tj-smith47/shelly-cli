// Package scan provides the sensoraddon scan command.
package scan

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/sensoraddon"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the sensoraddon scan command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "scan <device>",
		Aliases: []string{"discover", "find"},
		Short:   "Scan for OneWire devices",
		Long: `Scan for OneWire devices on the Sensor Add-on bus.

Currently only DS18B20 OneWire temperature sensors are supported.

Note: This will fail if a DHT22 sensor is in use, as DHT22 shares
the same GPIOs as the OneWire bus.`,
		Example: `  # Scan for OneWire devices
  shelly sensoraddon scan kitchen

  # JSON output
  shelly sensoraddon scan kitchen -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.SensorAddonService()

	var devices []sensoraddon.OneWireDevice
	err := cmdutil.RunWithSpinner(ctx, ios, "Scanning for OneWire devices...", func(ctx context.Context) error {
		var scanErr error
		devices, scanErr = svc.ScanOneWire(ctx, opts.Device)
		return scanErr
	})
	if err != nil {
		return err
	}

	return cmdutil.PrintListResult(ios, devices, func(ios *iostreams.IOStreams, items []sensoraddon.OneWireDevice) {
		if len(items) == 0 {
			ios.NoResults("OneWire devices")
			return
		}

		ios.Title("OneWire Devices")
		ios.Println()

		for _, dev := range items {
			typeStr := theme.Dim().Render("[" + dev.Type + "]")
			addrStr := theme.Highlight().Render(dev.Addr)

			if dev.Component != nil {
				ios.Printf("  %s %s -> %s\n", typeStr, addrStr, *dev.Component)
			} else {
				ios.Printf("  %s %s %s\n", typeStr, addrStr, theme.Dim().Render("(not linked)"))
			}
		}

		ios.Println()
		ios.Count("device", len(items))
	})
}
