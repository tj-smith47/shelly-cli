// Package list provides the zigbee list command.
package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	JSON    bool
}

// NewCommand creates the zigbee list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "devices"},
		Short:   "List Zigbee-capable devices",
		Long: `List all Zigbee-capable Shelly devices on the network.

Scans configured devices to find those with Zigbee support
and shows their current Zigbee status.

Note: This only shows devices in your Shelly CLI config, not
devices paired to Zigbee coordinators.`,
		Example: `  # List Zigbee-capable devices
  shelly zigbee list

  # Output as JSON
  shelly zigbee list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// ZigbeeDevice represents a Zigbee-capable device.
type ZigbeeDevice struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Model        string `json:"model,omitempty"`
	Enabled      bool   `json:"zigbee_enabled"`
	NetworkState string `json:"network_state,omitempty"`
	EUI64        string `json:"eui64,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, 60*shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	devices, err := scanZigbeeDevices(ctx, svc, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputZigbeeJSON(ios, devices)
	}

	displayZigbeeDevices(ios, devices)
	return nil
}

func scanZigbeeDevices(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams) ([]ZigbeeDevice, error) {
	cfg := config.Get()
	if cfg == nil {
		return nil, fmt.Errorf("no configuration found; run 'shelly init' first")
	}

	devices := make([]ZigbeeDevice, 0, len(cfg.Devices))

	for name, dev := range cfg.Devices {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		device, ok := checkZigbeeSupport(ctx, svc, name, dev, ios)
		if !ok {
			continue
		}

		if device.Enabled {
			enrichZigbeeStatus(ctx, svc, dev.Address, &device)
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func checkZigbeeSupport(ctx context.Context, svc *shelly.Service, name string, dev model.Device, ios *iostreams.IOStreams) (ZigbeeDevice, bool) {
	result, err := svc.RawRPC(ctx, dev.Address, "Zigbee.GetConfig", nil)
	if err != nil {
		ios.Debug("device %s does not support Zigbee: %v", name, err)
		return ZigbeeDevice{}, false
	}

	var cfg struct {
		Enable bool `json:"enable"`
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		ios.Debug("failed to marshal result for %s: %v", name, err)
		return ZigbeeDevice{}, false
	}
	if json.Unmarshal(jsonBytes, &cfg) != nil {
		return ZigbeeDevice{}, false
	}

	return ZigbeeDevice{
		Name:    name,
		Address: dev.Address,
		Model:   dev.Model,
		Enabled: cfg.Enable,
	}, true
}

func enrichZigbeeStatus(ctx context.Context, svc *shelly.Service, address string, device *ZigbeeDevice) {
	statusResult, err := svc.RawRPC(ctx, address, "Zigbee.GetStatus", nil)
	if err != nil {
		return
	}

	var status struct {
		NetworkState string `json:"network_state"`
		EUI64        string `json:"eui64"`
	}
	statusBytes, err := json.Marshal(statusResult)
	if err != nil {
		return
	}
	if json.Unmarshal(statusBytes, &status) == nil {
		device.NetworkState = status.NetworkState
		device.EUI64 = status.EUI64
	}
}

func outputZigbeeJSON(ios *iostreams.IOStreams, devices []ZigbeeDevice) error {
	output, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(output))
	return nil
}

func displayZigbeeDevices(ios *iostreams.IOStreams, devices []ZigbeeDevice) {
	if len(devices) == 0 {
		ios.Info("No Zigbee-capable devices found.")
		ios.Info("Zigbee is supported on Gen4 devices.")
		return
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("Zigbee-Capable Devices (%d):", len(devices))))
	ios.Println()

	for _, dev := range devices {
		displayZigbeeDevice(ios, dev)
	}
}

func displayZigbeeDevice(ios *iostreams.IOStreams, dev ZigbeeDevice) {
	ios.Printf("  %s\n", theme.Highlight().Render(dev.Name))
	ios.Printf("    Address: %s\n", dev.Address)
	if dev.Model != "" {
		ios.Printf("    Model: %s\n", dev.Model)
	}

	enabledStr := theme.Dim().Render("Disabled")
	if dev.Enabled {
		enabledStr = theme.StatusOK().Render("Enabled")
	}
	ios.Printf("    Zigbee: %s\n", enabledStr)

	if dev.NetworkState != "" {
		stateStyle := theme.Dim()
		if dev.NetworkState == "joined" {
			stateStyle = theme.StatusOK()
		}
		ios.Printf("    State: %s\n", stateStyle.Render(dev.NetworkState))
	}
	if dev.EUI64 != "" {
		ios.Printf("    EUI64: %s\n", dev.EUI64)
	}
	ios.Println()
}
