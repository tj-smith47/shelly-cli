// Package prometheus provides the Prometheus metrics exporter command.
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Port     int
	Devices  []string
	Interval time.Duration
}

// NewCommand creates the prometheus metrics command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:  f,
		Port:     9090,
		Interval: 15 * time.Second,
	}

	cmd := &cobra.Command{
		Use:   "prometheus",
		Short: "Start Prometheus metrics exporter",
		Long: `Start an HTTP server that exports metrics in Prometheus format.

The exporter collects metrics from all registered devices (or a specified
subset) and exposes them at /metrics for Prometheus scraping.

Metrics exported:
  Power metering (PM/PM1/EM/EM1 components):
  - shelly_power_watts: Current power consumption
  - shelly_voltage_volts: Voltage reading
  - shelly_current_amps: Current reading
  - shelly_energy_wh_total: Total energy consumption

  System metrics:
  - shelly_device_online: Device reachability (1=online, 0=offline)
  - shelly_wifi_rssi: WiFi signal strength in dBm
  - shelly_uptime_seconds: Device uptime
  - shelly_temperature_celsius: Device temperature
  - shelly_ram_free_bytes: Free RAM
  - shelly_ram_total_bytes: Total RAM

  Component state:
  - shelly_switch_on: Switch state (1=on, 0=off)

Labels include: device, component, component_id, phase`,
		Example: `  # Start exporter on default port 9090
  shelly metrics prometheus

  # Start on custom port with specific devices
  shelly metrics prometheus --port 8080 --devices kitchen,living-room

  # Collect metrics every 30 seconds
  shelly metrics prometheus --interval 30s`,
		Aliases: []string{"prom"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Port, "port", opts.Port, "HTTP port for the exporter")
	cmd.Flags().StringSliceVar(&opts.Devices, "devices", nil, "Devices to include (default: all registered)")
	cmd.Flags().DurationVar(&opts.Interval, "interval", opts.Interval, "Metrics collection interval")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()
	cfg, err := opts.Factory.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get device list
	devices := opts.Devices
	if len(devices) == 0 {
		for name := range cfg.Devices {
			devices = append(devices, name)
		}
	}

	if len(devices) == 0 {
		ios.Warning("No devices found. Register devices using 'shelly device add' or specify --devices")
		return nil
	}

	sort.Strings(devices)

	collector := shelly.NewPrometheusCollector(svc, devices)

	// Initial collection
	collector.Collect(ctx)

	// Start background collection
	go func() {
		ticker := time.NewTicker(opts.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collector.Collect(ctx)
			}
		}
	}()

	// HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		if _, writeErr := w.Write([]byte(collector.FormatMetrics())); writeErr != nil {
			ios.DebugErr("writing metrics response", writeErr)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, writeErr := fmt.Fprintf(w, `<html><body>
<h1>Shelly Metrics Exporter</h1>
<p>Devices: %d</p>
<p><a href="/metrics">Metrics</a></p>
</body></html>`, len(devices)); writeErr != nil {
			ios.DebugErr("writing index page", writeErr)
		}
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", opts.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ios.Printf("Starting Prometheus exporter on http://localhost:%d/metrics\n", opts.Port)
	ios.Printf("Monitoring %d devices with %s collection interval\n", len(devices), opts.Interval)
	ios.Printf("Press Ctrl+C to stop\n")

	// Handle shutdown
	go func() {
		<-ctx.Done()
		// Use fresh context for shutdown since parent context is already cancelled
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			ios.DebugErr("server shutdown", shutdownErr)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
