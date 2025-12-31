// Package jsonmetrics provides the JSON metrics output command.
package jsonmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds command options.
type Options struct {
	Factory    *cmdutil.Factory
	Devices    []string
	Continuous bool
	Interval   time.Duration
	Output     string
}

// NewCommand creates the JSON metrics command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:  f,
		Interval: 10 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "json",
		Aliases: []string{"j"},
		Short:   "Output metrics as JSON",
		Long: `Output device metrics in JSON format.

Outputs power, voltage, current, and energy metrics from all registered
devices (or a specified subset) in a machine-readable JSON format.

Use --continuous to stream metrics at regular intervals, or run once
for a single snapshot.`,
		Example: `  # Output metrics once to stdout
  shelly metrics json

  # Output for specific devices
  shelly metrics json --devices kitchen,living-room

  # Stream metrics every 10 seconds
  shelly metrics json --continuous --interval 10s

  # Save to file
  shelly metrics json --output metrics.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringSliceVar(&opts.Devices, "devices", nil, "Devices to include (default: all registered)")
	cmd.Flags().BoolVarP(&opts.Continuous, "continuous", "c", false, "Stream metrics continuously")
	cmd.Flags().DurationVarP(&opts.Interval, "interval", "i", opts.Interval, "Collection interval for continuous mode")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file (default: stdout)")

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

	// Determine output writer
	out := ios.Out
	if opts.Output != "" {
		cleanPath := filepath.Clean(opts.Output)
		file, err := os.Create(cleanPath)
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

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")

	if opts.Continuous {
		// Stream mode
		ticker := time.NewTicker(opts.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := encoder.Encode(svc.CollectJSONMetrics(ctx, devices)); err != nil {
					ios.DebugErr("encoding metrics", err)
				}
			}
		}
	}

	// Single shot
	return encoder.Encode(svc.CollectJSONMetrics(ctx, devices))
}
