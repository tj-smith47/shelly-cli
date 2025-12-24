// Package prometheus provides the Prometheus metrics exporter command.
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the prometheus metrics command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		port     int
		devices  []string
		interval time.Duration
	)

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
			return run(cmd.Context(), f, port, devices, interval)
		},
	}

	cmd.Flags().IntVar(&port, "port", 9090, "HTTP port for the exporter")
	cmd.Flags().StringSliceVar(&devices, "devices", nil, "Devices to include (default: all registered)")
	cmd.Flags().DurationVar(&interval, "interval", 15*time.Second, "Metrics collection interval")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, port int, devices []string, interval time.Duration) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get device list
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

	collector := &metricsCollector{
		svc:     svc,
		devices: devices,
		ios:     ios,
	}

	// Initial collection
	collector.collect(ctx)

	// Start background collection
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collector.collect(ctx)
			}
		}
	}()

	// HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		collector.writeMetrics(w)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := fmt.Fprintf(w, `<html><body>
<h1>Shelly Metrics Exporter</h1>
<p>Devices: %d</p>
<p><a href="/metrics">Metrics</a></p>
</body></html>`, len(devices)); err != nil {
			ios.DebugErr("writing index page", err)
		}
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ios.Printf("Starting Prometheus exporter on http://localhost:%d/metrics\n", port)
	ios.Printf("Monitoring %d devices with %s collection interval\n", len(devices), interval)
	ios.Printf("Press Ctrl+C to stop\n")

	// Handle shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			ios.DebugErr("server shutdown", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// metricsCollector collects and caches metrics from devices.
type metricsCollector struct {
	svc     *shelly.Service
	devices []string
	ios     *iostreams.IOStreams

	mu      sync.RWMutex
	metrics map[string]*shelly.PrometheusMetrics
}

func (c *metricsCollector) collect(ctx context.Context) {
	newMetrics := make(map[string]*shelly.PrometheusMetrics)

	g, ctx := errgroup.WithContext(ctx)
	// Use global rate limit for concurrency (service layer also enforces this)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	var mu sync.Mutex

	for _, device := range c.devices {
		dev := device
		g.Go(func() error {
			m, err := c.svc.CollectPrometheusMetrics(ctx, dev)
			if err != nil {
				c.ios.DebugErr(fmt.Sprintf("collecting metrics for %s", dev), err)
			}
			mu.Lock()
			newMetrics[dev] = m
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.ios.DebugErr("collecting device metrics", err)
	}

	c.mu.Lock()
	c.metrics = newMetrics
	c.mu.Unlock()
}

func (c *metricsCollector) writeMetrics(w http.ResponseWriter) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Combine all metrics
	combined := &shelly.PrometheusMetrics{}
	for _, m := range c.metrics {
		if m != nil {
			combined.Metrics = append(combined.Metrics, m.Metrics...)
		}
	}

	// Use service layer formatter
	output := shelly.FormatPrometheusMetrics(combined)
	if _, err := w.Write([]byte(output)); err != nil {
		c.ios.DebugErr("writing metrics response", err)
	}
}
