// Package log provides the debug log command for Gen1 devices.
package log

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the debug log command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "log <device>",
		Aliases: []string{"logs", "debug-log"},
		Short:   "Get device debug log (Gen1)",
		Long: `Get the debug log from a Gen1 Shelly device.

Note: This command only works with Gen1 devices. Gen2+ devices use a
different logging mechanism via RPC.

For Gen2+ devices, use:
  shelly debug rpc <device> Sys.GetStatus

Gen1 debug logs can help diagnose connectivity issues, action URL problems,
and other device behavior.`,
		Example: `  # Get debug log from a Gen1 device
  shelly debug log living-room-gen1

  # For Gen2+ devices, use RPC instead
  shelly debug rpc living-room Sys.GetStatus`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()

	// Gen1 debug log is not yet implemented in shelly-go
	// This is a placeholder that documents the workaround
	ios.Warning("Gen1 debug log retrieval is not yet implemented.")
	ios.Println()
	ios.Info("For Gen1 devices, access the debug log via HTTP:")
	ios.Info("  curl http://<device-ip>/debug/log")
	ios.Println()
	ios.Info("For Gen2+ devices, use the RPC interface:")
	ios.Info("  shelly debug rpc %s Sys.GetStatus", opts.Device)
	ios.Info("  shelly debug rpc %s Shelly.GetStatus", opts.Device)

	// Try to get basic info to verify connection
	svc := opts.Factory.ShellyService()
	info, err := svc.DeviceInfo(ctx, opts.Device)
	if err != nil {
		ios.Debug("could not connect to device: %v", err)
		return fmt.Errorf("could not connect to %s: %w", opts.Device, err)
	}

	ios.Println()
	ios.Info("Device %s is a Gen%d device (%s)", opts.Device, info.Generation, info.Model)

	if info.Generation == 1 {
		// Get the resolved address for the curl hint
		deviceCfg, cfgErr := config.ResolveDevice(opts.Device)
		if cfgErr == nil && deviceCfg.Address != "" {
			ios.Info("Try: curl http://%s/debug/log", deviceCfg.Address)
		} else {
			ios.Info("Try: curl http://<device-ip>/debug/log")
		}
	} else {
		ios.Info("Try: shelly debug rpc %s Sys.GetStatus", opts.Device)
	}

	return nil
}
