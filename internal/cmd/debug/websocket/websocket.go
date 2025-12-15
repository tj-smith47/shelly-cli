// Package websocket provides the debug websocket command.
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Duration time.Duration
}

// NewCommand creates the debug websocket command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "websocket <device>",
		Aliases: []string{"ws", "events"},
		Short:   "Debug WebSocket connection",
		Long: `Debug WebSocket connection to a Shelly device.

This command shows WebSocket configuration and attempts to verify
connectivity. Gen2+ devices support WebSocket for real-time event
notifications.

Note: Full WebSocket event streaming requires the shelly-go WebSocket
transport which is not yet integrated into the CLI.`,
		Example: `  # Check WebSocket configuration
  shelly debug websocket living-room

  # Monitor for a specific duration (placeholder)
  shelly debug websocket living-room --duration 30s`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVar(&opts.Duration, "duration", 10*time.Second, "Monitoring duration (not yet implemented)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	ios.Println(theme.Bold().Render("WebSocket Configuration:"))
	ios.Println()

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		info := conn.Info()
		ios.Printf("  Device: %s (%s)\n", info.Model, info.ID)
		ios.Printf("  Generation: %d\n", info.Generation)
		ios.Println()

		// Get WebSocket config
		result, err := conn.Call(ctx, "Ws.GetConfig", nil)
		if err != nil {
			ios.Debug("Ws.GetConfig failed: %v", err)
			ios.Warning("WebSocket configuration not available (may not be supported)")
			ios.Println()
			tryFallbackWsConfig(ctx, conn, ios)
		} else {
			printJSONResult(ios, "WebSocket Config:", result)
		}

		// Get WebSocket status
		statusResult, statusErr := conn.Call(ctx, "Ws.GetStatus", nil)
		if statusErr != nil {
			ios.Debug("Ws.GetStatus failed: %v", statusErr)
		} else {
			printJSONResult(ios, "WebSocket Status:", statusResult)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	ios.Info("WebSocket event streaming not yet implemented.")
	ios.Info("Use 'shelly monitor events <device>' for event monitoring via polling.")

	return nil
}

// tryFallbackWsConfig attempts to get WebSocket config from Sys.GetConfig.
func tryFallbackWsConfig(ctx context.Context, conn *client.Client, ios *iostreams.IOStreams) {
	sysResult, sysErr := conn.Call(ctx, "Sys.GetConfig", nil)
	if sysErr != nil {
		return
	}

	jsonBytes, err := json.Marshal(sysResult)
	if err != nil {
		ios.DebugErr("failed to marshal sys config", err)
		return
	}

	var cfg map[string]any
	if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
		ios.DebugErr("failed to unmarshal sys config", err)
		return
	}

	ws, ok := cfg["ws"].(map[string]any)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("WebSocket (from Sys.GetConfig):"))
	for k, v := range ws {
		ios.Printf("    %s: %v\n", k, v)
	}
	ios.Println()
}

// printJSONResult prints a JSON result with a header.
func printJSONResult(ios *iostreams.IOStreams, header string, result any) {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ios.DebugErr("failed to marshal result", err)
		return
	}
	ios.Println("  " + theme.Highlight().Render(header))
	ios.Println(string(jsonBytes))
	ios.Println()
}
