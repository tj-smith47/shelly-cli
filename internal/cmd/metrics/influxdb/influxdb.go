// Package influxdb provides the InfluxDB line protocol metrics output command.
package influxdb

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// Options holds command options.
type Options struct {
	Factory     *cmdutil.Factory
	Devices     []string
	Continuous  bool
	Interval    time.Duration
	Output      string
	Measurement string
	Tags        []string
}

// NewCommand creates the InfluxDB metrics command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:     f,
		Interval:    10 * time.Second,
		Measurement: "shelly",
	}

	cmd := &cobra.Command{
		Use:   "influxdb",
		Short: "Output metrics in InfluxDB line protocol",
		Long: `Output device metrics in InfluxDB line protocol format.

Outputs power, voltage, current, and energy metrics from all registered
devices (or a specified subset) in InfluxDB line protocol format suitable
for piping to InfluxDB or Telegraf.

Format: measurement,tags field=value,field=value timestamp

Use --continuous to stream metrics at regular intervals.`,
		Example: `  # Output metrics once to stdout
  shelly metrics influxdb

  # Output for specific devices with custom measurement name
  shelly metrics influxdb --devices kitchen --measurement home_power

  # Stream metrics every 10 seconds
  shelly metrics influxdb --continuous --interval 10s

  # Add custom tags
  shelly metrics influxdb --tags location=home,floor=1

  # Pipe directly to InfluxDB (requires influx CLI)
  shelly metrics influxdb | influx write -b mybucket`,
		Aliases: []string{"influx", "line"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringSliceVar(&opts.Devices, "devices", nil, "Devices to include (default: all registered)")
	cmd.Flags().BoolVarP(&opts.Continuous, "continuous", "c", false, "Stream metrics continuously")
	cmd.Flags().DurationVarP(&opts.Interval, "interval", "i", opts.Interval, "Collection interval for continuous mode")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&opts.Measurement, "measurement", "m", opts.Measurement, "Measurement name")
	cmd.Flags().StringSliceVarP(&opts.Tags, "tags", "t", nil, "Additional tags (key=value)")

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
	tags := export.ParseTags(opts.Tags)

	// Setup output destination
	out := ios.Out
	if opts.Output != "" {
		cleanPath := filepath.Clean(opts.Output)
		file, err := config.Fs().Create(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				ios.DebugErr("closing output file", cerr)
			}
		}()
		out = file
	}

	// Write points to output
	writePoints := func(points []export.InfluxDBPoint) {
		for _, p := range points {
			line := export.FormatInfluxDBPoint(p)
			if _, err := fmt.Fprintln(out, line); err != nil {
				ios.DebugErr("writing line", err)
			}
		}
	}

	if opts.Continuous {
		return svc.StreamInfluxDBPoints(ctx, devices, opts.Measurement, tags, opts.Interval, writePoints)
	}

	// Single shot
	points := svc.CollectInfluxDBPointsMulti(ctx, devices, opts.Measurement, tags)
	writePoints(points)
	return nil
}
