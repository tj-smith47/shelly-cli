// Package prometheus provides the Prometheus metrics exporter command.
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// MetricsCollector collects metrics from devices.
type MetricsCollector struct {
	svc     *shelly.Service
	devices []string
	ios     *iostreams.IOStreams

	mu      sync.RWMutex
	metrics map[string][]Metric
}

// Metric represents a single Prometheus metric.
type Metric struct {
	Name   string
	Value  float64
	Labels map[string]string
	Help   string
	Type   string
}

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

The exporter collects power, voltage, current, and energy metrics from
all registered devices (or a specified subset) and exposes them at /metrics.

Metrics exported:
  - shelly_power_watts: Current power consumption
  - shelly_voltage_volts: Voltage reading
  - shelly_current_amps: Current reading
  - shelly_energy_wh: Total energy consumption
  - shelly_device_online: Device online status (1=online, 0=offline)

Labels include: device, component_type, component_id`,
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

	collector := &MetricsCollector{
		svc:     svc,
		devices: devices,
		metrics: make(map[string][]Metric),
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

func (c *MetricsCollector) collect(ctx context.Context) {
	newMetrics := make(map[string][]Metric)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	var mu sync.Mutex

	for _, device := range c.devices {
		dev := device
		g.Go(func() error {
			metrics := c.collectDevice(ctx, dev)
			mu.Lock()
			newMetrics[dev] = metrics
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.ios.DebugErr("collecting device metrics", err)
		return
	}

	c.mu.Lock()
	c.metrics = newMetrics
	c.mu.Unlock()
}

func (c *MetricsCollector) collectDevice(ctx context.Context, device string) []Metric {
	var metrics []Metric

	// Device online metric
	online := 1.0

	// Collect PM components
	pmIDs, err := c.svc.ListPMComponents(ctx, device)
	if err == nil {
		for _, id := range pmIDs {
			status, err := c.svc.GetPMStatus(ctx, device, id)
			if err != nil {
				continue
			}

			labels := map[string]string{
				"device":         device,
				"component_type": "pm",
				"component_id":   fmt.Sprintf("%d", id),
			}

			metrics = append(metrics,
				Metric{
					Name:   "shelly_power_watts",
					Value:  status.APower,
					Labels: labels,
					Help:   "Current active power in watts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_voltage_volts",
					Value:  status.Voltage,
					Labels: labels,
					Help:   "Voltage in volts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_current_amps",
					Value:  status.Current,
					Labels: labels,
					Help:   "Current in amps",
					Type:   "gauge",
				},
			)

			if status.AEnergy != nil {
				metrics = append(metrics, Metric{
					Name:   "shelly_energy_wh",
					Value:  status.AEnergy.Total,
					Labels: labels,
					Help:   "Total energy in watt-hours",
					Type:   "counter",
				})
			}
		}
	} else {
		online = 0
	}

	// Collect PM1 components
	pm1IDs, err := c.svc.ListPM1Components(ctx, device)
	if err == nil {
		for _, id := range pm1IDs {
			status, err := c.svc.GetPM1Status(ctx, device, id)
			if err != nil {
				continue
			}

			labels := map[string]string{
				"device":         device,
				"component_type": "pm1",
				"component_id":   fmt.Sprintf("%d", id),
			}

			metrics = append(metrics,
				Metric{
					Name:   "shelly_power_watts",
					Value:  status.APower,
					Labels: labels,
					Help:   "Current active power in watts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_voltage_volts",
					Value:  status.Voltage,
					Labels: labels,
					Help:   "Voltage in volts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_current_amps",
					Value:  status.Current,
					Labels: labels,
					Help:   "Current in amps",
					Type:   "gauge",
				},
			)

			if status.AEnergy != nil {
				metrics = append(metrics, Metric{
					Name:   "shelly_energy_wh",
					Value:  status.AEnergy.Total,
					Labels: labels,
					Help:   "Total energy in watt-hours",
					Type:   "counter",
				})
			}
		}
	}

	// Collect EM components
	emIDs, err := c.svc.ListEMComponents(ctx, device)
	if err == nil {
		for _, id := range emIDs {
			status, err := c.svc.GetEMStatus(ctx, device, id)
			if err != nil {
				continue
			}

			labels := map[string]string{
				"device":         device,
				"component_type": "em",
				"component_id":   fmt.Sprintf("%d", id),
			}

			metrics = append(metrics,
				Metric{
					Name:   "shelly_power_watts",
					Value:  status.TotalActivePower,
					Labels: labels,
					Help:   "Current active power in watts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_voltage_volts",
					Value:  status.AVoltage,
					Labels: labels,
					Help:   "Phase A voltage in volts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_current_amps",
					Value:  status.TotalCurrent,
					Labels: labels,
					Help:   "Total current in amps",
					Type:   "gauge",
				},
			)
		}
	}

	// Collect EM1 components
	em1IDs, err := c.svc.ListEM1Components(ctx, device)
	if err == nil {
		for _, id := range em1IDs {
			status, err := c.svc.GetEM1Status(ctx, device, id)
			if err != nil {
				continue
			}

			labels := map[string]string{
				"device":         device,
				"component_type": "em1",
				"component_id":   fmt.Sprintf("%d", id),
			}

			metrics = append(metrics,
				Metric{
					Name:   "shelly_power_watts",
					Value:  status.ActPower,
					Labels: labels,
					Help:   "Current active power in watts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_voltage_volts",
					Value:  status.Voltage,
					Labels: labels,
					Help:   "Voltage in volts",
					Type:   "gauge",
				},
				Metric{
					Name:   "shelly_current_amps",
					Value:  status.Current,
					Labels: labels,
					Help:   "Current in amps",
					Type:   "gauge",
				},
			)
		}
	}

	// Add device online metric
	metrics = append(metrics, Metric{
		Name:  "shelly_device_online",
		Value: online,
		Labels: map[string]string{
			"device": device,
		},
		Help: "Device online status (1=online, 0=offline)",
		Type: "gauge",
	})

	return metrics
}

func (c *MetricsCollector) writeMetrics(w http.ResponseWriter) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Collect all metrics grouped by name
	byName := make(map[string][]Metric)
	for _, metrics := range c.metrics {
		for _, m := range metrics {
			byName[m.Name] = append(byName[m.Name], m)
		}
	}

	// Sort metric names for consistent output
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)

	// Write metrics in Prometheus format
	for _, name := range names {
		metrics := byName[name]
		if len(metrics) == 0 {
			continue
		}

		// Write HELP and TYPE once per metric name
		if _, err := fmt.Fprintf(w, "# HELP %s %s\n", name, metrics[0].Help); err != nil {
			c.ios.DebugErr("writing metric help", err)
			return
		}
		if _, err := fmt.Fprintf(w, "# TYPE %s %s\n", name, metrics[0].Type); err != nil {
			c.ios.DebugErr("writing metric type", err)
			return
		}

		for _, m := range metrics {
			if _, err := fmt.Fprintf(w, "%s%s %g\n", name, formatLabels(m.Labels), m.Value); err != nil {
				c.ios.DebugErr("writing metric value", err)
				return
			}
		}
	}
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%q", k, escapeLabel(v)))
	}
	sort.Strings(parts)

	return "{" + strings.Join(parts, ",") + "}"
}

func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}
