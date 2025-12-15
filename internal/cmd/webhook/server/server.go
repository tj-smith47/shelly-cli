// Package server provides the webhook server subcommand.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	Port       int
	Interface  string
	LogJSON    bool
	AutoConfig bool
	Devices    []string
}

// webhookEntry represents a single received webhook.
type webhookEntry struct {
	Timestamp  time.Time         `json:"timestamp"`
	RemoteAddr string            `json:"remote_addr"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Query      map[string]string `json:"query"`
	Body       string            `json:"body,omitempty"`
}

// NewCommand creates the webhook server command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Port:      8080,
		Interface: "0.0.0.0",
	}

	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serve", "listen", "receiver"},
		Short:   "Start a local webhook receiver server",
		Long: `Start a local HTTP server to receive and log webhooks from Shelly devices.

This is useful for testing and debugging webhook configurations. The server
logs all incoming requests with their headers, query parameters, and body.

The server will display its URL which can be used to configure device webhooks.`,
		Example: `  # Start server on default port 8080
  shelly webhook server

  # Start on a specific port
  shelly webhook server --port 9000

  # Start with JSON logging for piping
  shelly webhook server --log-json

  # Auto-configure devices to send webhooks here
  shelly webhook server --auto-config --device kitchen --device bedroom`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVar(&opts.Interface, "interface", "0.0.0.0", "Network interface to bind to")
	cmd.Flags().BoolVar(&opts.LogJSON, "log-json", false, "Log webhooks as JSON (for piping)")
	cmd.Flags().BoolVar(&opts.AutoConfig, "auto-config", false, "Auto-configure devices to use this server")
	cmd.Flags().StringSliceVar(&opts.Devices, "device", nil, "Devices to auto-configure (with --auto-config)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Get local IP for display
	localIP := getLocalIP()
	serverURL := fmt.Sprintf("http://%s:%d", localIP, opts.Port)

	ios.Success("Webhook Server")
	ios.Println()
	ios.Info("Listening on: %s:%d", opts.Interface, opts.Port)
	ios.Info("Webhook URL: %s/webhook", serverURL)
	ios.Info("Configure your device webhooks to POST to: %s/webhook", serverURL)
	ios.Println()
	ios.Info("Press Ctrl+C to stop...")

	// Auto-configure devices if requested
	if opts.AutoConfig && len(opts.Devices) > 0 {
		ios.Info("Auto-configuring devices...")
		configureDevices(ctx, f, opts.Devices, serverURL)
		ios.Println()
	}

	// Create HTTP server
	server := &webhookServer{
		ios:     ios,
		logJSON: opts.LogJSON,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleWebhook)
	mux.HandleFunc("/webhook", server.handleWebhook)
	mux.HandleFunc("/health", server.handleHealth)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", opts.Interface, opts.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		ios.Println("")
		ios.Info("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			ios.DebugErr("shutdown", err)
		}

		ios.Success("Server stopped")
		return nil

	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}

type webhookServer struct {
	ios     *iostreams.IOStreams
	logJSON bool
	mu      sync.Mutex
	count   int
}

func (s *webhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()

	// Build webhook entry
	entry := webhookEntry{
		Timestamp:  time.Now(),
		RemoteAddr: r.RemoteAddr,
		Method:     r.Method,
		Path:       r.URL.Path,
		Headers:    make(map[string]string),
		Query:      make(map[string]string),
	}

	// Capture headers
	for key := range r.Header {
		entry.Headers[key] = r.Header.Get(key)
	}

	// Capture query params
	for key := range r.URL.Query() {
		entry.Query[key] = r.URL.Query().Get(key)
	}

	// Capture body
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil && len(body) > 0 {
			entry.Body = string(body)
		}
	}

	// Log the webhook
	if s.logJSON {
		s.logWebhookJSON(entry)
	} else {
		s.logWebhookPretty(count, entry)
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		s.ios.DebugErr("write response", err)
	}
}

func (s *webhookServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"healthy","webhooks_received":` + fmt.Sprint(s.count) + `}`)); err != nil {
		s.ios.DebugErr("write health response", err)
	}
}

func (s *webhookServer) logWebhookJSON(entry webhookEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		s.ios.DebugErr("marshal webhook", err)
		return
	}
	s.ios.Println(string(data))
}

func (s *webhookServer) logWebhookPretty(count int, entry webhookEntry) {
	s.ios.Success("[#%d] Webhook Received", count)
	s.ios.Printf("  Time:   %s\n", entry.Timestamp.Format(time.RFC3339))
	s.ios.Printf("  From:   %s\n", entry.RemoteAddr)
	s.ios.Printf("  Method: %s\n", entry.Method)
	s.ios.Printf("  Path:   %s\n", entry.Path)

	if len(entry.Query) > 0 {
		s.ios.Printf("  Query:\n")
		for k, v := range entry.Query {
			s.ios.Printf("    %s: %s\n", k, v)
		}
	}

	if entry.Body != "" {
		s.ios.Printf("  Body:\n")
		s.ios.Printf("    %s\n", formatBody(entry.Body))
	}

	s.ios.Println("")
}

// formatBody attempts to pretty-print JSON, falls back to raw body.
func formatBody(body string) string {
	var jsonBody map[string]any
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		return body
	}
	prettyBody, err := json.MarshalIndent(jsonBody, "    ", "  ")
	if err != nil {
		return body
	}
	return string(prettyBody)
}

func getLocalIP() string {
	// Try to get a non-loopback IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return "localhost"
}

func configureDevices(ctx context.Context, f *cmdutil.Factory, devices []string, serverURL string) {
	ios := f.IOStreams()
	svc := f.ShellyService()

	webhookURL := serverURL + "/webhook"

	for _, device := range devices {
		ios.Printf("  Configuring %s... ", device)

		conn, err := svc.Connect(ctx, device)
		if err != nil {
			ios.Printf("failed (connect: %v)\n", err)
			continue
		}

		// Create a webhook for all events
		params := map[string]any{
			"cid":    0,
			"enable": true,
			"event":  "*",
			"urls":   []string{webhookURL},
		}

		_, err = conn.Call(ctx, "Webhook.Create", params)
		iostreams.CloseWithDebug("closing auto-config connection", conn)

		if err != nil {
			ios.Printf("failed (%v)\n", err)
			continue
		}

		ios.Printf("done\n")
	}
}
