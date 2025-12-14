// Package status provides the cloud status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the cloud status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <device>",
		Short: "Show cloud connection status",
		Long: `Show the Shelly Cloud connection status for a device.

Displays whether the device is currently connected to Shelly Cloud.`,
		Example: `  # Show cloud status
  shelly cloud status living-room

  # Output as JSON
  shelly cloud status living-room -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Getting cloud status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.CloudStatus, error) {
			return svc.GetCloudStatus(ctx, device)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *shelly.CloudStatus) {
	ios.Title("Cloud Status")
	ios.Println()

	var state string
	if status.Connected {
		state = theme.StatusOK().Render("Connected")
	} else {
		state = theme.StatusError().Render("Disconnected")
	}
	ios.Printf("  Status: %s\n", state)
}
