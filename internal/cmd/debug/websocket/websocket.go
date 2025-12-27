// Package websocket provides the debug websocket command.
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Duration time.Duration
	Raw      bool
}

// NewCommand creates the debug websocket command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "websocket <device>",
		Aliases: []string{"ws", "events"},
		Short:   "Debug WebSocket connection and stream events",
		Long: `Debug WebSocket connection and stream real-time events from a Shelly device.

This command connects to a Gen2+ device via WebSocket and streams all
notifications (state changes, sensor updates, button presses, etc.) in real-time.

Gen2+ devices support WebSocket at ws://<device>/rpc for bidirectional
communication and event notifications.`,
		Example: `  # Stream events for 30 seconds (default)
  shelly debug websocket living-room

  # Stream events for 5 minutes
  shelly debug websocket living-room --duration 5m

  # Stream events indefinitely (until Ctrl+C)
  shelly debug websocket living-room --duration 0

  # Raw JSON output
  shelly debug websocket living-room --raw`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVar(&opts.Duration, "duration", 30*time.Second, "Monitoring duration (0 for indefinite)")
	cmd.Flags().BoolVar(&opts.Raw, "raw", false, "Output raw JSON events")

	return cmd
}

func run(ctx context.Context, opts *Options) error { //nolint:gocyclo // will fix soon
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Resolve device to get address
	resolved, err := svc.ResolveWithGeneration(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to resolve device: %w", err)
	}
	deviceHost := resolved.Address

	// Get device info and WebSocket configuration
	var deviceInfo struct {
		ID         string `json:"id"`
		Model      string `json:"model"`
		Generation int    `json:"gen"`
	}

	ios.Println(theme.Bold().Render("WebSocket Configuration:"))
	ios.Println()

	err = svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("WebSocket is only supported on Gen2+ devices (this is Gen1)")
		}

		conn := dev.Gen2()
		info := conn.Info()
		deviceInfo.ID = info.ID
		deviceInfo.Model = info.Model
		deviceInfo.Generation = info.Generation

		ios.Printf("  Device: %s (%s)\n", info.Model, info.ID)
		ios.Printf("  Generation: %d\n", info.Generation)
		ios.Println()

		// Get WebSocket config
		result, wsErr := conn.Call(ctx, "Ws.GetConfig", nil)
		if wsErr == nil {
			term.DisplayJSONResult(ios, "WebSocket Config:", result)
		} else {
			ios.Debug("Ws.GetConfig failed: %v", wsErr)
			ios.Warning("WebSocket configuration not available")
			ios.Println()
			// Fallback: try to extract ws config from Sys.GetConfig
			sysResult, sysErr := conn.Call(ctx, "Sys.GetConfig", nil)
			wsConfig := output.ExtractMapSection(sysResult, "ws")
			if sysErr == nil && wsConfig != nil {
				term.DisplayWebSocketFallbackConfig(ios, wsConfig)
			}
		}

		// Get WebSocket status
		statusResult, statusErr := conn.Call(ctx, "Ws.GetStatus", nil)
		if statusErr != nil {
			ios.Debug("Ws.GetStatus failed: %v", statusErr)
		} else {
			term.DisplayJSONResult(ios, "WebSocket Status:", statusResult)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Now connect via WebSocket for event streaming
	wsURL := fmt.Sprintf("ws://%s/rpc", deviceHost)
	ios.Println(theme.Bold().Render("Event Streaming:"))
	ios.Printf("  Connecting to %s\n", wsURL)
	ios.Println()

	// Build WebSocket options
	wsOpts := []transport.Option{
		transport.WithReconnect(true),
		transport.WithPingInterval(15 * time.Second),
	}

	// Check if device has auth configured
	cfg, cfgErr := opts.Factory.Config()
	if cfgErr != nil {
		ios.DebugErr("load config", cfgErr)
	}
	if cfg != nil {
		creds := cfg.GetAllDeviceCredentials()
		if cred, ok := creds[opts.Device]; ok && cred.Password != "" {
			wsOpts = append(wsOpts, transport.WithAuth(cred.Username, cred.Password))
		}
	}

	ws := transport.NewWebSocket(wsURL, wsOpts...)

	// Register state change callback
	ws.OnStateChange(func(state transport.ConnectionState) {
		term.DisplayWebSocketConnectionState(ios, state.String())
	})

	// Connect
	if err := ws.Connect(ctx); err != nil {
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}
	defer iostreams.CloseWithDebug("closing websocket", ws)

	// Subscribe to notifications
	eventCount := 0
	if err := ws.Subscribe(func(data json.RawMessage) {
		eventCount++
		timestamp := time.Now().Format("15:04:05")

		if opts.Raw {
			ios.Printf("[%s] %s\n", timestamp, string(data))
		} else {
			term.DisplayWebSocketEvent(ios, timestamp, data)
		}
	}); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Wait for duration or context cancellation
	if opts.Duration > 0 {
		ios.Info("Streaming events for %s (press Ctrl+C to stop)...", opts.Duration)
		ios.Println()

		select {
		case <-ctx.Done():
			ios.Println()
			ios.Info("Stopped by user")
		case <-time.After(opts.Duration):
			ios.Println()
			ios.Info("Duration reached")
		}
	} else {
		ios.Info("Streaming events indefinitely (press Ctrl+C to stop)...")
		ios.Println()

		<-ctx.Done()
		ios.Println()
		ios.Info("Stopped by user")
	}

	ios.Println()
	ios.Success("Received %d events", eventCount)

	return nil
}
