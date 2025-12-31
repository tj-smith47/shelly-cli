// Package telemetry provides opt-in anonymous usage analytics for the CLI.
// When enabled, it sends command usage statistics to help improve the CLI.
// No device information, IP addresses, or personal data is collected.
package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// Endpoint is the telemetry collection endpoint.
// Set via ldflags at build time:
//
//	-ldflags "-X github.com/tj-smith47/shelly-cli/internal/telemetry.Endpoint=https://..."
//
// Falls back to SHELLY_TELEMETRY_ENDPOINT env var.
// If neither is set, telemetry is effectively disabled (no endpoint to send to).
var Endpoint string

// Event represents a telemetry event.
type Event struct {
	// Command is the command path (e.g., "device info", "switch on")
	Command string `json:"command"`

	// Success indicates whether the command completed successfully
	Success bool `json:"success"`

	// Duration is the command execution time in milliseconds
	Duration int64 `json:"duration_ms,omitempty"`

	// Version is the CLI version
	Version string `json:"version"`

	// OS is the operating system (e.g., "linux", "darwin", "windows")
	OS string `json:"os"`

	// Arch is the architecture (e.g., "amd64", "arm64")
	Arch string `json:"arch"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`
}

// Client handles telemetry event collection and submission.
type Client struct {
	endpoint   string
	httpClient *http.Client
	enabled    bool
	mu         sync.RWMutex

	// eventQueue holds events to be sent
	eventQueue chan Event
	done       chan struct{}
	wg         sync.WaitGroup

	// tickerDuration controls how often the sender checks for queued events.
	// Defaults to 10 seconds. Configurable for testing.
	tickerDuration time.Duration

	// marshalFunc is the JSON marshal function. Defaults to json.Marshal.
	// Can be overridden for testing error paths.
	marshalFunc func(v any) ([]byte, error)
}

var (
	globalClient     *Client
	globalClientOnce sync.Once
)

// getEndpoint returns the configured endpoint, checking ldflags then env var.
func getEndpoint() string {
	if Endpoint != "" {
		return Endpoint
	}
	return os.Getenv("SHELLY_TELEMETRY_ENDPOINT")
}

// DefaultClient returns the global telemetry client.
// The client is lazily initialized on first use.
func DefaultClient() *Client {
	globalClientOnce.Do(func() {
		globalClient = NewClient(getEndpoint())
	})
	return globalClient
}

// NewClient creates a new telemetry client.
func NewClient(endpoint string) *Client {
	c := &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		enabled:        false,
		eventQueue:     make(chan Event, 100),
		done:           make(chan struct{}),
		tickerDuration: 10 * time.Second,
		marshalFunc:    json.Marshal,
	}

	// Start background sender
	c.wg.Go(c.sender)

	return c
}

// NewClientWithOptions creates a new telemetry client with custom options.
// tickerDuration controls how often queued events are sent (default 10s).
func NewClientWithOptions(endpoint string, tickerDuration time.Duration) *Client {
	if tickerDuration <= 0 {
		tickerDuration = 10 * time.Second
	}
	c := &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		enabled:        false,
		eventQueue:     make(chan Event, 100),
		done:           make(chan struct{}),
		tickerDuration: tickerDuration,
		marshalFunc:    json.Marshal,
	}

	// Start background sender
	c.wg.Go(c.sender)

	return c
}

// SetEnabled enables or disables telemetry collection.
func (c *Client) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// IsEnabled returns whether telemetry is enabled.
func (c *Client) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// HasEndpoint returns whether a telemetry endpoint is configured.
func (c *Client) HasEndpoint() bool {
	return c.endpoint != ""
}

// Track records a command execution event.
// This is non-blocking and safe to call even if telemetry is disabled.
func (c *Client) Track(command string, success bool, duration time.Duration) {
	// Skip if disabled or no endpoint configured
	if !c.IsEnabled() || !c.HasEndpoint() {
		return
	}

	event := Event{
		Command:   command,
		Success:   success,
		Duration:  duration.Milliseconds(),
		Version:   version.Version,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now().UTC(),
	}

	// Non-blocking send - drop event if queue is full
	select {
	case c.eventQueue <- event:
	default:
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "telemetry queue full, dropping event", nil)
	}
}

// sender runs in the background and sends queued events.
func (c *Client) sender() {
	// Batch events and send periodically
	var batch []Event
	ticker := time.NewTicker(c.tickerDuration)
	defer ticker.Stop()

	for {
		select {
		case event := <-c.eventQueue:
			batch = append(batch, event)
			// Send immediately if batch is large enough
			if len(batch) >= 10 {
				c.sendBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				c.sendBatch(batch)
				batch = nil
			}

		case <-c.done:
			// Send remaining events before shutdown
			close(c.eventQueue)
			for event := range c.eventQueue {
				batch = append(batch, event)
			}
			if len(batch) > 0 {
				c.sendBatch(batch)
			}
			return
		}
	}
}

// sendBatch sends a batch of events to the telemetry endpoint.
func (c *Client) sendBatch(events []Event) {
	if len(events) == 0 || c.endpoint == "" {
		return
	}

	data, err := c.marshalFunc(events)
	if err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "marshal telemetry events", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(data))
	if err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "create telemetry request", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "shelly-cli/"+version.Version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "send telemetry", err)
		return
	}
	if cerr := resp.Body.Close(); cerr != nil {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "close telemetry response body", cerr)
	}
}

// Close gracefully shuts down the telemetry client.
func (c *Client) Close() {
	close(c.done)
	c.wg.Wait()
}

// Track is a convenience function that uses the default client.
func Track(command string, success bool, duration time.Duration) {
	DefaultClient().Track(command, success, duration)
}

// SetEnabled is a convenience function that uses the default client.
func SetEnabled(enabled bool) {
	DefaultClient().SetEnabled(enabled)
}

// IsEnabled is a convenience function that uses the default client.
func IsEnabled() bool {
	return DefaultClient().IsEnabled()
}

// HasEndpoint is a convenience function that uses the default client.
func HasEndpoint() bool {
	return DefaultClient().HasEndpoint()
}

// Close is a convenience function that uses the default client.
func Close() {
	DefaultClient().Close()
}

// GetCommandPath extracts the command path from a cobra command tree and args.
// Returns the subcommand path (e.g., "device info", "switch on") or "shelly" for root.
func GetCommandPath(rootCmd *cobra.Command, args []string) string {
	cmd, _, err := rootCmd.Find(args)
	if err != nil || cmd == nil || cmd == rootCmd {
		return "shelly"
	}

	// Build path from command chain
	var parts []string
	for c := cmd; c != nil && c != rootCmd; c = c.Parent() {
		parts = append([]string{c.Name()}, parts...)
	}
	return strings.Join(parts, " ")
}
