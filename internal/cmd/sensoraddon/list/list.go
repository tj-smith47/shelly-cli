// Package list provides the sensoraddon list command.
package list

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

// NewCommand creates the sensoraddon list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List configured peripherals",
		Long:    `List all configured Sensor Add-on peripherals on a device.`,
		Example: `  # List peripherals
  shelly sensoraddon list kitchen

  # JSON output
  shelly sensoraddon list kitchen -o json`,
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

	var peripherals []sensoraddon.Peripheral
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching peripherals...", func(ctx context.Context) error {
		var fetchErr error
		peripherals, fetchErr = svc.ListPeripherals(ctx, opts.Device)
		return fetchErr
	})
	if err != nil {
		return err
	}

	return cmdutil.PrintListResult(ios, peripherals, func(ios *iostreams.IOStreams, items []sensoraddon.Peripheral) {
		if len(items) == 0 {
			ios.NoResults("peripherals")
			return
		}

		ios.Title("Sensor Add-on Peripherals")
		ios.Println()

		for _, p := range items {
			typeStr := theme.Dim().Render("[" + string(p.Type) + "]")
			compStr := theme.Highlight().Render(p.Component)

			if p.Addr != nil {
				ios.Printf("  %s %s (addr: %s)\n", typeStr, compStr, *p.Addr)
			} else {
				ios.Printf("  %s %s\n", typeStr, compStr)
			}
		}

		ios.Println()
		ios.Count("peripheral", len(items))
	})
}
