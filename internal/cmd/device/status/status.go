// Package status provides the device status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the device status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show device status",
		Long:    `Display the full status of a Shelly device including all components.`,
		Example: `  # Show status for a device
  shelly device status living-room

  # Using alias
  shelly dev st bedroom`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Resolve device to check if it's plugin-managed
	device, err := svc.ResolveWithGeneration(ctx, opts.Device)
	if err != nil {
		return err
	}

	// Check if this is a plugin-managed device
	if device.IsPluginManaged() {
		ios.StartProgress("Getting device status...")
		status, pluginErr := svc.GetPluginDeviceStatus(ctx, device)
		ios.StopProgress()
		if pluginErr != nil {
			return pluginErr
		}
		term.DisplayPluginDeviceStatus(ios, device, status)
		return nil
	}

	// Standard Shelly device status
	return cmdutil.RunDeviceStatus(ctx, ios, svc, opts.Device,
		"Getting device status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.DeviceStatus, error) {
			return svc.DeviceStatus(ctx, device)
		},
		term.DisplayDeviceStatus)
}
