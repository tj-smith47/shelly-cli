// Package coiot provides Gen1 CoIoT real-time monitoring commands.
package coiot

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Timeout time.Duration
	Follow  bool
}

// NewCommand creates the gen1 coiot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "coiot",
		Aliases: []string{"monitor", "watch", "listen"},
		Short:   "Monitor Gen1 devices via CoIoT",
		Long: `Monitor Gen1 Shelly devices via CoIoT multicast.

CoIoT (Constrained Application Protocol for IoT) is a protocol used by
Gen1 Shelly devices to broadcast their status on the local network via
multicast UDP to 224.0.1.187:5683.

This command listens for these broadcasts and displays real-time updates
as devices announce their status (typically every 30 seconds to 2 minutes).

Note: CoIoT works best with Gen1 devices. Gen2+ devices use mDNS and
outbound WebSocket connections instead.`,
		Example: `  # Monitor for 30 seconds
  shelly gen1 coiot

  # Monitor for 5 minutes
  shelly gen1 coiot --timeout 5m

  # Monitor continuously (Ctrl+C to stop)
  shelly gen1 coiot --follow`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", 30*time.Second, "Monitoring duration")
	cmd.Flags().BoolVarP(&opts.Follow, "follow", "f", false, "Monitor continuously until interrupted")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	if opts.Follow {
		ios.Info("Monitoring CoIoT broadcasts. Press Ctrl+C to stop...")
	} else {
		ios.Info("Monitoring CoIoT broadcasts for %s...", opts.Timeout)
	}
	ios.Println()

	// Set up context with timeout or signal handling
	var cancel context.CancelFunc
	if opts.Follow {
		ctx, cancel = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	} else {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}
	defer cancel()

	// Create CoIoT discoverer
	coiotDiscoverer := discovery.NewCoIoTDiscoverer()
	defer func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			ios.Debug("stopping CoIoT discoverer: %v", err)
		}
	}()

	// Start listening
	devices, err := coiotDiscoverer.DiscoverWithContext(ctx)
	if err != nil && ctx.Err() == nil {
		return err
	}

	ios.Println()
	displaySummary(ios, devices)

	return nil
}

func displaySummary(ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice) {
	if len(devices) == 0 {
		ios.Info("No CoIoT broadcasts received.")
		ios.Info("Make sure you are on the same network as your Gen1 devices.")
		return
	}

	ios.Println(theme.Bold().Render("Devices discovered via CoIoT:"))
	ios.Println()

	for _, d := range devices {
		ios.Printf("  %s\n", theme.Highlight().Render(d.Name))
		ios.Printf("    Address: %s\n", d.Address)
		if d.Model != "" {
			ios.Printf("    Model:   %s\n", d.Model)
		}
		if d.MACAddress != "" {
			ios.Printf("    MAC:     %s\n", d.MACAddress)
		}
		ios.Println()
	}

	ios.Success("Received broadcasts from %d device(s)", len(devices))
}
