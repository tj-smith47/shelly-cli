// Package list provides the zigbee list command.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	JSON    bool
}

// NewCommand creates the zigbee list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "devices"},
		Short:   "List Zigbee-capable devices",
		Long: `List all Zigbee-capable Shelly devices on the network.

Scans configured devices to find those with Zigbee support
and shows their current Zigbee status.

Note: This only shows devices in your Shelly CLI config, not
devices paired to Zigbee coordinators.`,
		Example: `  # List Zigbee-capable devices
  shelly zigbee list

  # Output as JSON
  shelly zigbee list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, 60*shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	devices, err := svc.Wireless().ScanZigbeeDevices(ctx, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return term.OutputZigbeeDevicesJSON(ios, devices)
	}

	term.DisplayZigbeeDevices(ios, devices)
	return nil
}
