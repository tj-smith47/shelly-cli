// Package status provides the ethernet status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the ethernet status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <device>",
		Short: "Show Ethernet status",
		Long: `Show the current Ethernet status for a device.

Displays connection status and IP address. Only available on Pro devices
with an Ethernet port.`,
		Example: `  # Show Ethernet status
  shelly ethernet status living-room-pro

  # Output as JSON
  shelly ethernet status living-room-pro -o json`,
		Args: cobra.ExactArgs(1),
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
		"Getting Ethernet status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.EthernetStatus, error) {
			return svc.GetEthernetStatus(ctx, device)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *shelly.EthernetStatus) {
	ios.Title("Ethernet Status")
	ios.Println()

	if status.IP != "" {
		ios.Printf("  Status:     Connected\n")
		ios.Printf("  IP Address: %s\n", status.IP)
	} else {
		ios.Printf("  Status:     Not connected\n")
	}
}
