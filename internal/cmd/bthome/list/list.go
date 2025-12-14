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
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

BTHome is a Bluetooth Low Energy (BLE) protocol for sensors and devices.
Shelly BLU sensors (motion, door/window, button) connect to a gateway
device that collects their readings.

Shows configured BTHomeDevice components with their current status,
signal strength (RSSI), battery level, and last update time.

Use 'shelly bthome add' to discover and pair new devices.
Use 'shelly bthome sensors' to view sensor readings.

Output is formatted as styled text by default. Use --json for
structured output suitable for scripting.`,
		Example: `  # List all BTHome devices
  shelly bthome list living-room

  # Output as JSON
  shelly bthome list living-room --json

  # Get devices with low battery
  shelly bthome list living-room --json | jq '.[] | select(.battery != null and .battery < 20)'

  # Find devices with weak signal
  shelly bthome list living-room --json | jq '.[] | select(.rssi != null and .rssi < -80)'

  # Get device addresses (MAC)
  shelly bthome list living-room --json | jq -r '.[].addr'

  # List device names and IDs
  shelly bthome list living-room --json | jq '.[] | {name, id}'

  # Short form
  shelly bthome ls living-room`,
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

	devices, err := fetchDevices(ctx, svc, opts.Device, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputJSON(ios, devices)
	}

	displayDevices(ios, devices, opts.Device)
	return nil
}

func fetchDevices(ctx context.Context, svc *shelly.Service, device string, ios *iostreams.IOStreams) ([]BTHomeDeviceInfo, error) {
	var devices []BTHomeDeviceInfo

	err := svc.WithConnection(ctx, device, func(conn *client.Client) error {
		status, err := getDeviceStatus(ctx, conn)
		if err != nil {
			return err
		}

		deviceStatuses := bthomeutil.CollectDevices(status, ios)
		devices = enrichDevices(ctx, conn, deviceStatuses)
		return nil
	})

	return devices, err
}

func getDeviceStatus(ctx context.Context, conn *client.Client) (map[string]json.RawMessage, error) {
	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var status map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

func enrichDevices(ctx context.Context, conn *client.Client, deviceStatuses []bthomeutil.DeviceStatus) []BTHomeDeviceInfo {
	devices := make([]BTHomeDeviceInfo, 0, len(deviceStatuses))

	for _, devStatus := range deviceStatuses {
		name, addr := getDeviceConfig(ctx, conn, devStatus.ID)
		devices = append(devices, BTHomeDeviceInfo{
			ID:         devStatus.ID,
			Name:       name,
			Addr:       addr,
			RSSI:       devStatus.RSSI,
			Battery:    devStatus.Battery,
			LastUpdate: devStatus.LastUpdate,
		})
	}

	return devices
}

func getDeviceConfig(ctx context.Context, conn *client.Client, id int) (name, addr string) {
	configResult, err := conn.Call(ctx, "BTHomeDevice.GetConfig", map[string]any{"id": id})
	if err != nil {
		return "", ""
	}

	var cfg struct {
		Name *string `json:"name"`
		Addr string  `json:"addr"`
	}
	cfgBytes, err := json.Marshal(configResult)
	if err != nil {
		return "", ""
	}
	if json.Unmarshal(cfgBytes, &cfg) != nil {
		return "", ""
	}

	if cfg.Name != nil {
		name = *cfg.Name
	}
	return name, cfg.Addr
}

func outputJSON(ios *iostreams.IOStreams, devices []BTHomeDeviceInfo) error {
	output, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(output))
	return nil
}

func displayDevices(ios *iostreams.IOStreams, devices []BTHomeDeviceInfo, gatewayDevice string) {
	if len(devices) == 0 {
		ios.Info("No BTHome devices found.")
		ios.Info("Use 'shelly bthome add %s' to discover new devices.", gatewayDevice)
		return
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("BTHome Devices (%d):", len(devices))))
	ios.Println()

	for _, dev := range devices {
		displayDevice(ios, dev)
	}
}

func displayDevice(ios *iostreams.IOStreams, dev BTHomeDeviceInfo) {
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
