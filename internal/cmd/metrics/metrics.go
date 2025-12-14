// Package metrics provides commands for exporting device metrics.
package metrics

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/metrics/influxdb"
	jsonmetrics "github.com/tj-smith47/shelly-cli/internal/cmd/metrics/json"
	"github.com/tj-smith47/shelly-cli/internal/cmd/metrics/prometheus"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the metrics command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Export device metrics",
		Long: `Export metrics from Shelly devices in various formats.

Supports multiple output formats for integration with monitoring systems:
  - Prometheus: Start an HTTP exporter for Prometheus scraping
  - JSON: Output metrics in JSON format for custom integrations
  - InfluxDB: Output in InfluxDB line protocol for time-series databases

All formats export: power, voltage, current, energy, temperature, and device status.`,
		Aliases: []string{"metric", "export-metrics"},
		Example: `  # Start Prometheus exporter
  shelly metrics prometheus --devices kitchen,bedroom

  # Export metrics as JSON
  shelly metric json kitchen

  # Export in InfluxDB line protocol
  shelly metrics influxdb kitchen`,
	}

	cmd.AddCommand(prometheus.NewCommand(f))
	cmd.AddCommand(jsonmetrics.NewCommand(f))
	cmd.AddCommand(influxdb.NewCommand(f))

	return cmd
}
