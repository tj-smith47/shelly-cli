// Package ping provides the ping command for device connectivity testing.
package ping

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	Count   int
	Timeout time.Duration
}

// NewCommand creates the ping command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Count:   1,
		Timeout: 5 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "ping <device>",
		Aliases: []string{"p", "test"},
		Short:   "Test device connectivity",
		Long: `Test connectivity to a Shelly device.

Sends a request to the device's RPC API and measures response time.
This is useful for:
  - Checking if a device is online
  - Measuring network latency
  - Verifying device configuration

The device can be specified by registered name or IP address.`,
		Example: `  # Ping a device once
  shelly ping kitchen-light

  # Ping multiple times
  shelly ping kitchen-light -c 5

  # Ping with custom timeout
  shelly ping 192.168.1.100 --timeout 10s`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Count, "count", "c", 1, "Number of pings to send")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 5*time.Second, "Timeout for each ping")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, identifier string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.Info("PING %s", identifier)

	var totalTime time.Duration
	successCount := 0

	for i := range opts.Count {
		pingCtx, cancel := context.WithTimeout(ctx, opts.Timeout)

		start := time.Now()
		info, err := svc.DevicePing(pingCtx, identifier)
		elapsed := time.Since(start)
		cancel()

		if err != nil {
			ios.Printf("Request %d: timeout or error - %v\n", i+1, err)
		} else {
			successCount++
			totalTime += elapsed
			if opts.Count == 1 {
				// Single ping - show device info
				ios.Success("Reply from %s: time=%v", identifier, elapsed.Round(time.Millisecond))
				ios.Info("  Model: %s", info.Model)
				ios.Info("  App: %s", info.App)
				ios.Info("  Firmware: %s", info.Firmware)
				ios.Info("  MAC: %s", info.MAC)
			} else {
				// Multiple pings - show sequence number
				ios.Success("Reply from %s: seq=%d time=%v", identifier, i+1, elapsed.Round(time.Millisecond))
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
		ios.Printf("--- %s ping statistics ---\n", identifier)
		ios.Printf("%d packets transmitted, %d received, %.0f%% packet loss\n",
			opts.Count, successCount, float64(opts.Count-successCount)/float64(opts.Count)*100)
		if successCount > 0 {
			avgTime := totalTime / time.Duration(successCount)
			ios.Printf("rtt avg = %v\n", avgTime.Round(time.Millisecond))
		}
	}

	if successCount == 0 {
		return fmt.Errorf("device %s is unreachable", identifier)
	}

	return nil
}
