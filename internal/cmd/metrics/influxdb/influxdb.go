// Package influxdb provides the InfluxDB line protocol metrics output command.
package influxdb

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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

// LineProtocolWriter writes metrics in InfluxDB line protocol format.
type LineProtocolWriter struct {
	out         io.Writer
	measurement string
	tags        map[string]string
	ios         *iostreams.IOStreams
}

// parseTags converts key=value pairs to a map.
func parseTags(tagPairs []string) map[string]string {
	tags := make(map[string]string)
	for _, pair := range tagPairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			tags[parts[0]] = parts[1]
		}
	}
	return tags
}

// setupOutput opens an output file or returns stdout.
func setupOutput(ios *iostreams.IOStreams, outputFile string) (io.Writer, func(), error) {
	if outputFile == "" {
		return ios.Out, func() {}, nil
	}

	cleanPath := filepath.Clean(outputFile)
	file, err := os.Create(cleanPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}
	cleanup := func() {
		if cerr := file.Close(); cerr != nil {
			ios.DebugErr("closing output file", cerr)
		}
	}
	return file, cleanup, nil
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
	tags := parseTags(opts.Tags)

	out, cleanup, err := setupOutput(ios, opts.Output)
	if err != nil {
		return err
	}
	defer cleanup()

	writer := &LineProtocolWriter{
		out:         out,
		measurement: opts.Measurement,
		tags:        tags,
		ios:         ios,
	}

	// Write points to output
	writePoints := func(points []export.InfluxDBPoint) {
		for _, p := range points {
			line := export.FormatInfluxDBPoint(p)
			if _, err := fmt.Fprintln(writer.out, line); err != nil {
				writer.ios.DebugErr("writing line", err)
			}
		}
	}

	if opts.Continuous {
		return runContinuous(ctx, svc, devices, opts.Measurement, tags, opts.Interval, writePoints)
	}

	// Single shot
	points := svc.CollectInfluxDBPointsMulti(ctx, devices, opts.Measurement, tags)
	writePoints(points)
	return nil
}

func runContinuous(ctx context.Context, svc *shelly.Service, devices []string, measurement string, tags map[string]string, interval time.Duration, writePoints func([]export.InfluxDBPoint)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			points := svc.CollectInfluxDBPointsMulti(ctx, devices, measurement, tags)
			writePoints(points)
		}
	}
}
