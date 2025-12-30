// Package events provides the cloud events subcommand.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	DeviceFilter string
	EventFilter  string
	Raw          bool
}

// makeEventHandler creates an event handler function for processing cloud events.
// This is extracted for testability.
func makeEventHandler(ios *iostreams.IOStreams, opts *Options) func(event *model.CloudEvent, raw []byte) error {
	return func(event *model.CloudEvent, raw []byte) error {
		// Raw output mode
		if opts.Raw {
			ios.Println(string(raw))
			return nil
		}

		// Output based on format
		switch opts.Format {
		case "json":
			formatted, jsonErr := json.Marshal(event)
			if jsonErr != nil {
				return jsonErr
			}
			ios.Println(string(formatted))
		default:
			term.DisplayCloudEvent(ios, event)
		}

		return nil
	}
}

// NewCommand creates the cloud events command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"watch", "subscribe"},
		Short:   "Subscribe to real-time cloud events",
		Long: `Subscribe to real-time events from the Shelly Cloud via WebSocket.

Displays events as they arrive from your cloud-connected devices.
Press Ctrl+C to stop listening.

Event types:
  Shelly:StatusOnChange  - Device status changed
  Shelly:Settings        - Device settings changed
  Shelly:Online          - Device came online/offline`,
		Example: `  # Watch all events
  shelly cloud events

  # Filter by device ID
  shelly cloud events --device abc123

  # Filter by event type
  shelly cloud events --event Shelly:Online

  # Output raw JSON
  shelly cloud events --raw

  # Output in JSON format
  shelly cloud events --format json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.DeviceFilter, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.EventFilter, "event", "", "Filter by event type")
	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "text", "text", "json")
	cmd.Flags().BoolVar(&opts.Raw, "raw", false, "Output raw JSON messages")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Check if logged in
	cfg := config.Get()
	if cfg.Cloud.AccessToken == "" {
		ios.Error("Not logged in to Shelly Cloud")
		ios.Info("Use 'shelly cloud login' to authenticate")
		return fmt.Errorf("not logged in")
	}

	// Get WebSocket URL
	wsURL, err := network.BuildCloudWebSocketURL(cfg.Cloud.ServerURL, cfg.Cloud.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to build WebSocket URL: %w", err)
	}

	ios.Info("Connecting to Shelly Cloud...")

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Create context that cancels on signal
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}

	conn, resp, dialErr := dialer.DialContext(ctx, wsURL, nil)
	if resp != nil && resp.Body != nil {
		iostreams.CloseWithDebug("closing websocket response body", resp.Body)
	}
	if dialErr != nil {
		return fmt.Errorf("failed to connect: %w", dialErr)
	}
	defer iostreams.CloseWithDebug("closing websocket connection", conn)

	ios.Success("Connected! Listening for events... (Ctrl+C to stop)")
	ios.Println()

	// Stream events
	streamOpts := network.CloudEventStreamOptions{
		DeviceFilter: opts.DeviceFilter,
		EventFilter:  opts.EventFilter,
		Raw:          opts.Raw,
	}

	err = network.StreamCloudEvents(ctx, conn, streamOpts, makeEventHandler(ios, opts))

	if err != nil {
		return err
	}

	ios.Println()
	ios.Info("Disconnected")
	return nil
}
