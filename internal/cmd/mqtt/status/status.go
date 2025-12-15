// Package status provides the mqtt status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the mqtt status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show MQTT status",
		Long: `Show the MQTT connection status for a device.

Displays whether MQTT is enabled and if the device is connected to the broker.`,
		Example: `  # Show MQTT status
  shelly mqtt status living-room

  # Output as JSON
  shelly mqtt status living-room -o json`,
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
		"Getting MQTT status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.MQTTStatus, error) {
			return svc.GetMQTTStatus(ctx, device)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *shelly.MQTTStatus) {
	ios.Title("MQTT Status")
	ios.Println()

	var state string
	if status.Connected {
		state = theme.StatusOK().Render("Connected")
	} else {
		state = theme.StatusError().Render("Disconnected")
	}
	ios.Printf("  Status: %s\n", state)
}
