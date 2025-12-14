// Package jsonmetrics provides the JSON metrics output command.
package jsonmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// MetricsOutput represents the JSON output format.
type MetricsOutput struct {
	Timestamp time.Time       `json:"timestamp"`
	Devices   []DeviceMetrics `json:"devices"`
}

// DeviceMetrics represents metrics for a single device.
type DeviceMetrics struct {
	Device     string             `json:"device"`
	Online     bool               `json:"online"`
	Components []ComponentMetrics `json:"components,omitempty"`
}

// ComponentMetrics represents metrics for a single component.
type ComponentMetrics struct {
	Type    string  `json:"type"`
	ID      int     `json:"id"`
	Power   float64 `json:"power_w,omitempty"`
	Voltage float64 `json:"voltage_v,omitempty"`
	Current float64 `json:"current_a,omitempty"`
	Energy  float64 `json:"energy_wh,omitempty"`
}

// NewCommand creates the JSON metrics command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		devices    []string
		continuous bool
		interval   time.Duration
		output     string
	)

	cmd := &cobra.Command{
		Use:   "json",
		Short: "Output metrics as JSON",
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
			return run(cmd.Context(), f, devices, continuous, interval, output)
		},
	}

	cmd.Flags().StringSliceVar(&devices, "devices", nil, "Devices to include (default: all registered)")
	cmd.Flags().BoolVarP(&continuous, "continuous", "c", false, "Stream metrics continuously")
	cmd.Flags().DurationVarP(&interval, "interval", "i", 10*time.Second, "Collection interval for continuous mode")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, continuous bool, interval time.Duration, outputFile string) error {
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

	// Determine output writer
	out := ios.Out
	if outputFile != "" {
		cleanPath := filepath.Clean(outputFile)
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

	if continuous {
		// Stream mode
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				metrics := collectMetrics(ctx, svc, devices)
				if err := encoder.Encode(metrics); err != nil {
					ios.DebugErr("encoding metrics", err)
				}
			}
		}
	}

	// Single shot
	metrics := collectMetrics(ctx, svc, devices)
	return encoder.Encode(metrics)
}

func collectMetrics(ctx context.Context, svc *shelly.Service, devices []string) MetricsOutput {
	output := MetricsOutput{
		Timestamp: time.Now(),
		Devices:   make([]DeviceMetrics, len(devices)),
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	var mu sync.Mutex

	for i, device := range devices {
		idx := i
		dev := device
		g.Go(func() error {
			metrics := collectDeviceMetrics(ctx, svc, dev)
			mu.Lock()
			output.Devices[idx] = metrics
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return output
	}

	return output
}

func collectDeviceMetrics(ctx context.Context, svc *shelly.Service, device string) DeviceMetrics {
	metrics := DeviceMetrics{
		Device: device,
		Online: true,
	}

	// Collect each component type
	collectPMMetrics(ctx, svc, device, &metrics)
	collectPM1Metrics(ctx, svc, device, &metrics)
	collectEMMetrics(ctx, svc, device, &metrics)
	collectEM1Metrics(ctx, svc, device, &metrics)

	// Mark offline if no components found
	if len(metrics.Components) == 0 {
		metrics.Online = false
	}

	return metrics
}

func collectPMMetrics(ctx context.Context, svc *shelly.Service, device string, metrics *DeviceMetrics) {
	pmIDs, err := svc.ListPMComponents(ctx, device)
	if err != nil {
		return
	}
	for _, id := range pmIDs {
		status, err := svc.GetPMStatus(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentMetrics{
			Type:    "pm",
			ID:      id,
			Power:   status.APower,
			Voltage: status.Voltage,
			Current: status.Current,
		}
		if status.AEnergy != nil {
			comp.Energy = status.AEnergy.Total
		}
		metrics.Components = append(metrics.Components, comp)
	}
}

func collectPM1Metrics(ctx context.Context, svc *shelly.Service, device string, metrics *DeviceMetrics) {
	pm1IDs, err := svc.ListPM1Components(ctx, device)
	if err != nil {
		return
	}
	for _, id := range pm1IDs {
		status, err := svc.GetPM1Status(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentMetrics{
			Type:    "pm1",
			ID:      id,
			Power:   status.APower,
			Voltage: status.Voltage,
			Current: status.Current,
		}
		if status.AEnergy != nil {
			comp.Energy = status.AEnergy.Total
		}
		metrics.Components = append(metrics.Components, comp)
	}
}

func collectEMMetrics(ctx context.Context, svc *shelly.Service, device string, metrics *DeviceMetrics) {
	emIDs, err := svc.ListEMComponents(ctx, device)
	if err != nil {
		return
	}
	for _, id := range emIDs {
		status, err := svc.GetEMStatus(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentMetrics{
			Type:    "em",
			ID:      id,
			Power:   status.TotalActivePower,
			Voltage: status.AVoltage,
			Current: status.TotalCurrent,
		}
		metrics.Components = append(metrics.Components, comp)
	}
}

func collectEM1Metrics(ctx context.Context, svc *shelly.Service, device string, metrics *DeviceMetrics) {
	em1IDs, err := svc.ListEM1Components(ctx, device)
	if err != nil {
		return
	}
	for _, id := range em1IDs {
		status, err := svc.GetEM1Status(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentMetrics{
			Type:    "em1",
			ID:      id,
			Power:   status.ActPower,
			Voltage: status.Voltage,
			Current: status.Current,
		}
		metrics.Components = append(metrics.Components, comp)
	}
}
