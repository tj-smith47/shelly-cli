// Package log provides the debug log command for Gen1 devices.
package log

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the debug log command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log <device>",
		Aliases: []string{"logs", "debug-log"},
		Short:   "Get device debug log (Gen1)",
		Long: `Get the debug log from a Gen1 Shelly device.

This command only works with Gen1 devices. Gen2+ devices use a
different logging mechanism via WebSocket or RPC.

Gen1 debug logs can help diagnose connectivity issues, action URL problems,
and other device behavior.`,
		Example: `  # Get debug log from a Gen1 device
  shelly debug log living-room-gen1

  # For Gen2+ devices, use RPC instead
  shelly debug rpc living-room Sys.GetStatus`,
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

	var logContent string
	err := svc.WithDevice(ctx, device, func(dev *shelly.DeviceClient) error {
		if !dev.IsGen1() {
			ios.Warning("Device %s is not a Gen1 device", device)
			ios.Info("Gen2+ devices use WebSocket/RPC for logging.")
			ios.Info("Try: shelly debug rpc %s Sys.GetStatus", device)
			return fmt.Errorf("debug log only available for Gen1 devices")
		}

		conn := dev.Gen1()
		var getErr error
		logContent, getErr = conn.GetDebugLog(ctx)
		return getErr
	})
	if err != nil {
		return err
	}

	if logContent == "" {
		ios.Info("Debug log is empty")
		return nil
	}

	ios.Println(logContent)
	return nil
}
