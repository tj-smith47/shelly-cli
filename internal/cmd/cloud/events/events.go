// Package events provides the cloud events subcommand.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	filterDeviceFlag string
	filterEventFlag  string
	formatFlag       string
	rawFlag          bool
)

// NewCommand creates the cloud events command.
func NewCommand() *cobra.Command {
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
			return run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&filterDeviceFlag, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&filterEventFlag, "event", "", "Filter by event type")
	cmd.Flags().StringVar(&formatFlag, "format", "text", "Output format (text, json)")
	cmd.Flags().BoolVar(&rawFlag, "raw", false, "Output raw JSON messages")

	return cmd
}

// cloudEvent represents a parsed cloud WebSocket event.
type cloudEvent struct {
	Event     string          `json:"event"`
	DeviceID  string          `json:"device_id,omitempty"`
	Device    string          `json:"device,omitempty"`
	Status    json.RawMessage `json:"status,omitempty"`
	Settings  json.RawMessage `json:"settings,omitempty"`
	Online    *int            `json:"online,omitempty"`
	Timestamp int64           `json:"ts,omitempty"`
}

func (e *cloudEvent) getDeviceID() string {
	if e.DeviceID != "" {
		return e.DeviceID
	}
	return e.Device
}

func run(ctx context.Context) error {
	ios := iostreams.System()

	// Check if logged in
	cfg := config.Get()
	if cfg.Cloud.AccessToken == "" {
		ios.Error("Not logged in to Shelly Cloud")
		ios.Info("Use 'shelly cloud login' to authenticate")
		return fmt.Errorf("not logged in")
	}

	// Get WebSocket URL
	wsURL, err := buildWebSocketURL(cfg.Cloud.ServerURL, cfg.Cloud.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to build WebSocket URL: %w", err)
	}

	ios.Info("Connecting to Shelly Cloud...")

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

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

	// Read loop
	return readEvents(ctx, ios, conn)
}

func buildWebSocketURL(serverURL, token string) (string, error) {
	if serverURL == "" {
		// Try to extract from token
		parsedToken, err := shelly.ParseToken(token)
		if err != nil {
			return "", fmt.Errorf("no server URL and failed to parse token: %w", err)
		}
		serverURL = parsedToken.UserAPIURL
	}

	if serverURL == "" {
		return "", fmt.Errorf("no server URL available")
	}

	// Parse to get hostname
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	hostname := u.Hostname()
	if hostname == "" {
		hostname = serverURL
	}

	// Build WebSocket URL: wss://{host}:6113/shelly/wss/hk_sock?t={token}
	return fmt.Sprintf("wss://%s:6113/shelly/wss/hk_sock?t=%s", hostname, url.QueryEscape(token)), nil
}

func readEvents(ctx context.Context, ios *iostreams.IOStreams, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			ios.Println()
			ios.Info("Disconnected")
			return nil
		default:
		}

		// Set read deadline
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}

		_, message, readErr := conn.ReadMessage()
		if readErr != nil {
			return handleReadError(ctx, readErr)
		}

		// Handle the message
		if err := handleMessage(ios, message); err != nil {
			ios.Debug("Error handling message: %v", err)
		}
	}
}

func handleMessage(ios *iostreams.IOStreams, data []byte) error {
	// Raw output mode
	if rawFlag {
		ios.Println(string(data))
		return nil
	}

	// Parse the event
	var event cloudEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Apply filters
	if filterDeviceFlag != "" && event.getDeviceID() != filterDeviceFlag {
		return nil
	}
	if filterEventFlag != "" && !strings.Contains(event.Event, filterEventFlag) {
		return nil
	}

	// Output based on format
	switch formatFlag {
	case "json":
		formatted, err := json.Marshal(event)
		if err != nil {
			return err
		}
		ios.Println(string(formatted))
	default:
		displayEvent(ios, &event)
	}

	return nil
}

func displayEvent(ios *iostreams.IOStreams, event *cloudEvent) {
	timestamp := time.Now().Format("15:04:05")
	if event.Timestamp > 0 {
		timestamp = time.Unix(event.Timestamp, 0).Format("15:04:05")
	}

	deviceID := event.getDeviceID()
	if deviceID == "" {
		deviceID = "(unknown)"
	}

	switch event.Event {
	case "Shelly:Online":
		status := "offline"
		if event.Online != nil && *event.Online == 1 {
			status = "online"
		}
		ios.Printf("[%s] %s %s: %s\n", timestamp, event.Event, deviceID, status)

	case "Shelly:StatusOnChange":
		ios.Printf("[%s] %s %s\n", timestamp, event.Event, deviceID)
		if len(event.Status) > 0 {
			printIndentedJSON(ios, event.Status)
		}

	case "Shelly:Settings":
		ios.Printf("[%s] %s %s\n", timestamp, event.Event, deviceID)
		if len(event.Settings) > 0 {
			printIndentedJSON(ios, event.Settings)
		}

	default:
		ios.Printf("[%s] %s %s\n", timestamp, event.Event, deviceID)
	}
}

func printIndentedJSON(ios *iostreams.IOStreams, data json.RawMessage) {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	formatted, err := json.MarshalIndent(parsed, "  ", "  ")
	if err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	ios.Printf("  %s\n", string(formatted))
}

// handleReadError processes WebSocket read errors.
// Returns nil for expected closures (normal close, context cancelled).
func handleReadError(ctx context.Context, err error) error {
	// Check for expected closure scenarios
	if isExpectedClosure(ctx, err) {
		return nil
	}
	return fmt.Errorf("read error: %w", err)
}

// isExpectedClosure checks if the error represents a normal termination.
func isExpectedClosure(ctx context.Context, err error) bool {
	// Normal WebSocket closure
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		return true
	}
	// Context was cancelled (user pressed Ctrl+C)
	return ctx.Err() != nil
}
