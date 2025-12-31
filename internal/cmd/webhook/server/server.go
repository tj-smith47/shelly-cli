// Package server provides the webhook server subcommand.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/webhook"
)

// Options holds the command options.
type Options struct {
	Factory    *cmdutil.Factory
	AutoConfig bool
	Devices    []string
	Interface  string
	LogJSON    bool
	Port       int
}

// NewCommand creates the webhook server command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:   f,
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
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVar(&opts.Interface, "interface", "0.0.0.0", "Network interface to bind to")
	cmd.Flags().BoolVar(&opts.LogJSON, "log-json", false, "Log webhooks as JSON (for piping)")
	cmd.Flags().BoolVar(&opts.AutoConfig, "auto-config", false, "Auto-configure devices to use this server")
	cmd.Flags().StringSliceVar(&opts.Devices, "device", nil, "Devices to auto-configure (with --auto-config)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Get local IP for display
	localIP := webhook.GetLocalIP()
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
		svc := opts.Factory.ShellyService()
		webhook.ConfigureDevices(ctx, ios, svc, opts.Devices, serverURL)
		ios.Println()
	}

	// Create HTTP server
	server := webhook.NewServer(ios, opts.LogJSON)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", opts.Interface, opts.Port),
		Handler:           server.Handler(),
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

		// Use fresh context for shutdown since parent context is already cancelled
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
