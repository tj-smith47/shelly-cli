// Package ping provides the device ping subcommand.
package ping

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the device ping command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping <device>",
		Short: "Check device connectivity",
		Long: `Check if a device is reachable and responding.

The ping command attempts to connect to the device and retrieve its info.
This is useful for verifying network connectivity and device availability.`,
		Example: `  # Ping a registered device
  shelly device ping living-room

  # Ping by IP address
  shelly device ping 192.168.1.100

  # Short form
  shelly dev ping office-switch`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	start := time.Now()
	spin := iostreams.NewSpinner("Pinging device...")
	spin.Start()

	info, err := svc.DevicePing(ctx, device)
	elapsed := time.Since(start)
	spin.Stop()

	if err != nil {
		iostreams.Error("Device %s is not reachable", device)
		return fmt.Errorf("ping failed: %w", err)
	}

	iostreams.Success("Device %s is online", info.ID)
	iostreams.Info("  Model: %s (Gen%d)", info.Model, info.Generation)
	iostreams.Info("  Response time: %v", elapsed.Round(time.Millisecond))

	return nil
}
