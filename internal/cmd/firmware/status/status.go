// Package status provides the firmware status subcommand.
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

// NewCommand creates the firmware status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show firmware status",
		Long: `Show the current firmware status for a device.

Displays update status, available versions, and rollback availability.`,
		Example: `  # Show firmware status
  shelly firmware status living-room`,
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
		"Getting firmware status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.FirmwareStatus, error) {
			return svc.GetFirmwareStatus(ctx, device)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *shelly.FirmwareStatus) {
	ios.Println(theme.Bold().Render("Firmware Status"))
	ios.Println("")

	// Status
	statusStr := status.Status
	if statusStr == "" {
		statusStr = "idle"
	}
	ios.Printf("  Status:      %s\n", statusStr)

	// Update available
	ios.Printf("  Update:      %s\n", output.RenderAvailableState(status.HasUpdate, "up to date"))
	if status.HasUpdate && status.NewVersion != "" {
		ios.Printf("  New Version: %s\n", status.NewVersion)
	}

	// Progress (if updating)
	if status.Progress > 0 {
		ios.Printf("  Progress:    %d%%\n", status.Progress)
	}

	// Rollback
	ios.Printf("  Rollback:    %s\n", output.RenderAvailableState(status.CanRollback, "not available"))
}
