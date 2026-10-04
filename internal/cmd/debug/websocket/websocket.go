// Package websocket provides the debug websocket command.
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// commandUse is the Use string for the debug websocket command.
const commandUse = "websocket <device>"

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
		Use:     commandUse,
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

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Resolve device and check generation
	resolved, err := svc.ResolveWithGeneration(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to resolve device: %w", err)
	}
	if resolved.Generation < 2 {
		return fmt.Errorf("WebSocket is only supported on Gen2+ devices (this is Gen%d)", resolved.Generation)
	}

	// Display device info
	term.DisplayWebSocketDeviceInfo(ios, resolved.Model, resolved.DisplayName(), resolved.Generation)

	// Get and display WebSocket configuration
	wsInfo, err := svc.GetWebSocketInfo(ctx, opts.Device)
	if err != nil {
		ios.DebugErr("get websocket info", err)
	} else {
		term.DisplayWebSocketInfo(ios, wsInfo.Config, wsInfo.Status)
	}

	wsURL := fmt.Sprintf("ws://%s/rpc", resolved.Address)

	// Connect to WebSocket
	ios.Println(theme.Bold().Render("Event Streaming:"))
	ios.Printf("  Connecting to %s\n", wsURL)
	ios.Println()

	ws, err := client.NewDeviceWebSocket(wsURL, resolved.Auth,
		transport.WithReconnect(true),
		transport.WithPingInterval(15*time.Second),
	)
	if err != nil {
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}
	// Notifications and reconnect state changes arrive on the transport's
	// goroutines; mu keeps their output and the event count from interleaving
	// with this goroutine's.
	var mu sync.Mutex
	ws.OnStateChange(func(state transport.ConnectionState) {
		mu.Lock()
		defer mu.Unlock()
		term.DisplayWebSocketConnectionState(ios, state.String())
	})

	if err := ws.Connect(ctx); err != nil {
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}
	defer iostreams.CloseWithDebug("closing websocket", ws)

	// Subscribe and stream events
	eventCount := 0
	if err := ws.Subscribe(func(data json.RawMessage) {
		mu.Lock()
		defer mu.Unlock()
		eventCount++
		timestamp := time.Now().Format("15:04:05")
		if opts.Raw {
			ios.Printf("[%s] %s\n", timestamp, string(data))
		} else {
			term.DisplayWebSocketEvent(ios, timestamp, data)
		}
	}); err != nil {
		ios.DebugErr("subscribe to websocket events", err)
	}

	// A device sends no notifications until it has answered one request frame,
	// and on a password-protected device that frame must authenticate.
	if _, err := client.StartDeviceNotifications(ctx, ws); err != nil {
		return fmt.Errorf("start event stream: %w", err)
	}

	mu.Lock()
	if opts.Duration > 0 {
		ios.Info("Streaming events for %s (press Ctrl+C to stop)...", opts.Duration)
	} else {
		ios.Info("Streaming events indefinitely (press Ctrl+C to stop)...")
	}
	ios.Println()
	mu.Unlock()

	stoppedByUser := true
	if opts.Duration > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(opts.Duration):
			stoppedByUser = false
		}
	} else {
		<-ctx.Done()
	}

	mu.Lock()
	defer mu.Unlock()
	ios.Println()
	if stoppedByUser {
		ios.Info("Stopped by user")
	} else {
		ios.Info("Duration reached")
	}
	ios.Println()
	ios.Success("Received %d events", eventCount)
	return nil
}
