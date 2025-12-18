// Package status provides the device status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Getting device status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.DeviceStatus, error) {
			return svc.DeviceStatus(ctx, device)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *shelly.DeviceStatus) {
	ios.Info("Device: %s", theme.Bold().Render(status.Info.ID))
	ios.Info("Model: %s (Gen%d)", status.Info.Model, status.Info.Generation)
	ios.Info("Firmware: %s", status.Info.Firmware)
	ios.Println()

	table := output.NewTable("Component", "Value")
	for key, value := range status.Status {
		table.AddRow(key, output.FormatDisplayValue(value))
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print device status table", err)
	}
}
