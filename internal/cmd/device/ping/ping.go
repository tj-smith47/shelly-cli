// Package ping provides the device ping subcommand.
package ping

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Count   int
	Device  string
	Timeout time.Duration
}

// NewCommand creates the device ping command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Count:   1,
		Timeout: 5 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "ping <device>",
		Aliases: []string{"check", "test", "p"},
		Short:   "Check device connectivity",
		Long: `Check if a device is reachable and responding.

The ping command attempts to connect to the device and retrieve its info.
This is useful for verifying network connectivity and device availability.

Use -c to send multiple pings and show statistics.`,
		Example: `  # Ping a registered device
  shelly device ping living-room

  # Ping by IP address
  shelly device ping 192.168.1.100

  # Ping multiple times with statistics
  shelly device ping kitchen -c 5

  # Ping with custom timeout
  shelly device ping slow-device --timeout 10s`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Count, "count", "c", 1, "Number of pings to send")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 5*time.Second, "Timeout for each ping")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	ios.Info("PING %s", opts.Device)

	var totalTime time.Duration
	successCount := 0

	for i := range opts.Count {
		pingCtx, cancel := context.WithTimeout(ctx, opts.Timeout)

		start := time.Now()
		info, err := svc.DevicePing(pingCtx, opts.Device)
		elapsed := time.Since(start)
		cancel()

		if err != nil {
			ios.Warning("Request %d: %s", i+1, tuierrors.FormatErrorMsg(err))
		} else {
			successCount++
			totalTime += elapsed
			if opts.Count == 1 {
				// Single ping - show device info
				ios.Success("Reply from %s: time=%v", opts.Device, elapsed.Round(time.Millisecond))
				ios.Info("  Model: %s (Gen%d)", info.Model, info.Generation)
				ios.Info("  App: %s", info.App)
				ios.Info("  Firmware: %s", info.Firmware)
				ios.Info("  MAC: %s", model.NormalizeMAC(info.MAC))
			} else {
				// Multiple pings - show sequence number
				ios.Success("Reply from %s: seq=%d time=%v", opts.Device, i+1, elapsed.Round(time.Millisecond))
			}
		}

		// Small delay between pings (but not after last one)
		if i < opts.Count-1 && opts.Count > 1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Show statistics for multiple pings
	if opts.Count > 1 {
		ios.Println("")
		ios.Printf("--- %s ping statistics ---\n", opts.Device)
		ios.Printf("%d packets transmitted, %d received, %.0f%% packet loss\n",
			opts.Count, successCount, float64(opts.Count-successCount)/float64(opts.Count)*100)
		if successCount > 0 {
			avgTime := totalTime / time.Duration(successCount)
			ios.Printf("rtt avg = %v\n", avgTime.Round(time.Millisecond))
		}
	}

	if successCount == 0 {
		ios.Println("")
		ios.Error("Device %q is not responding", opts.Device)
		ios.Info("Ensure the device is powered on and connected to your network")
		return fmt.Errorf("device %s is unreachable", opts.Device)
	}

	return nil
}
