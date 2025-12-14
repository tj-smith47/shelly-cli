// Package list provides the bthome list command.
package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmd/bthome/bthomeutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the bthome list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "devices"},
		Short:   "List BTHome devices",
		Long: `List all BTHome devices connected to a Shelly gateway.

Shows configured BTHomeDevice and BTHomeSensor components with their
current status, signal strength, and battery level.`,
		Example: `  # List all BTHome devices
  shelly bthome list living-room

  # Output as JSON
  shelly bthome list living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// BTHomeDeviceInfo holds information about a BTHome device.
type BTHomeDeviceInfo struct {
	ID         int     `json:"id"`
	Name       string  `json:"name,omitempty"`
	Addr       string  `json:"addr"`
	RSSI       *int    `json:"rssi,omitempty"`
	Battery    *int    `json:"battery,omitempty"`
	LastUpdate float64 `json:"last_update,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var devices []BTHomeDeviceInfo

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		// Get all component status to find BTHome devices
		result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		var status map[string]json.RawMessage
		if err := json.Unmarshal(jsonBytes, &status); err != nil {
			return fmt.Errorf("failed to parse status: %w", err)
		}

		// Collect BTHome device statuses using helper
		deviceStatuses := bthomeutil.CollectDevices(status, ios)

		// Enrich each device with config data (name, address)
		for _, devStatus := range deviceStatuses {
			configResult, cfgErr := conn.Call(ctx, "BTHomeDevice.GetConfig", map[string]any{"id": devStatus.ID})
			var name, addr string
			if cfgErr == nil {
				var cfg struct {
					Name *string `json:"name"`
					Addr string  `json:"addr"`
				}
				cfgBytes, cfgMarshalErr := json.Marshal(configResult)
				if cfgMarshalErr == nil && json.Unmarshal(cfgBytes, &cfg) == nil {
					addr = cfg.Addr
					if cfg.Name != nil {
						name = *cfg.Name
					}
				}
			}

			devices = append(devices, BTHomeDeviceInfo{
				ID:         devStatus.ID,
				Name:       name,
				Addr:       addr,
				RSSI:       devStatus.RSSI,
				Battery:    devStatus.Battery,
				LastUpdate: devStatus.LastUpdate,
			})
		}

		return nil
	})
	if err != nil {
		return err
	}

	if opts.JSON {
		output, err := json.MarshalIndent(devices, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if len(devices) == 0 {
		ios.Info("No BTHome devices found.")
		ios.Info("Use 'shelly bthome add %s' to discover new devices.", opts.Device)
		return nil
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("BTHome Devices (%d):", len(devices))))
	ios.Println()

	for _, dev := range devices {
		name := dev.Name
		if name == "" {
			name = fmt.Sprintf("Device %d", dev.ID)
		}

		ios.Printf("  %s\n", theme.Highlight().Render(name))
		ios.Printf("    ID: %d\n", dev.ID)
		if dev.Addr != "" {
			ios.Printf("    Address: %s\n", dev.Addr)
		}
		if dev.RSSI != nil {
			ios.Printf("    RSSI: %d dBm\n", *dev.RSSI)
		}
		if dev.Battery != nil {
			ios.Printf("    Battery: %d%%\n", *dev.Battery)
		}
		ios.Println()
	}

	return nil
}
