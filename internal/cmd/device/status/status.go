// Package status provides the device status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the device status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, deviceName string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Resolve device to check if it's plugin-managed
	device, err := svc.ResolveWithGeneration(ctx, deviceName)
	if err != nil {
		return err
	}

	// Check if this is a plugin-managed device
	if device.IsPluginManaged() {
		return runPluginStatus(ctx, ios, svc, device)
	}

	// Standard Shelly device status
	return cmdutil.RunDeviceStatus(ctx, ios, svc, deviceName,
		"Getting device status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.DeviceStatus, error) {
			return svc.DeviceStatus(ctx, device)
		},
		term.DisplayDeviceStatus)
}

// runPluginStatus fetches and displays status for a plugin-managed device.
func runPluginStatus(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device model.Device) error {
	ios.StartProgress("Getting device status...")

	status, err := svc.GetPluginDeviceStatus(ctx, device)
	ios.StopProgress()

	if err != nil {
		return err
	}

	return printPluginStatus(ios, device, status)
}

// printPluginStatus displays the plugin device status.
func printPluginStatus(ios *iostreams.IOStreams, device model.Device, status *plugins.DeviceStatusResult) error {
	term.DisplayPluginDeviceStatus(ios, device, status)
	return nil
}
