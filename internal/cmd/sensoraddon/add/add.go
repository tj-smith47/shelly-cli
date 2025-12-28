// Package add provides the sensoraddon add command.
package add

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/sensoraddon"
)

// Options holds command options.
type Options struct {
	Device  string
	Type    string
	CID     int
	Addr    string
	Factory *cmdutil.Factory
}

// NewCommand creates the sensoraddon add command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "add <device> <type>",
		Aliases: []string{"create", "new"},
		Short:   "Add a peripheral",
		Long: `Add a Sensor Add-on peripheral to a device.

Peripheral types:
  ds18b20    - Dallas 1-Wire temperature sensor (requires --addr)
  dht22      - Temperature and humidity sensor
  digital_in - Digital input
  analog_in  - Analog input

Note: Changes require a device reboot to take effect.`,
		Example: `  # Add a DS18B20 sensor
  shelly sensoraddon add kitchen ds18b20 --addr "40:255:100:6:199:204:149:177"

  # Add a DHT22 sensor
  shelly sensoraddon add kitchen dht22

  # Add with specific component ID
  shelly sensoraddon add kitchen digital_in --cid 101`,
		Args: cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completion.DeviceNames()(cmd, args, toComplete)
			}
			if len(args) == 1 {
				return sensoraddon.ValidPeripheralTypes, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Type = strings.ToLower(args[1])
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.CID, "cid", 0, "Component ID (auto-assigned if not specified)")
	cmd.Flags().StringVar(&opts.Addr, "addr", "", "Sensor address (required for DS18B20)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.SensorAddonService()

	// Validate type
	var pType sensoraddon.PeripheralType
	switch opts.Type {
	case "ds18b20":
		pType = sensoraddon.TypeDS18B20
		if opts.Addr == "" {
			return fmt.Errorf("--addr is required for DS18B20 sensors")
		}
	case "dht22":
		pType = sensoraddon.TypeDHT22
	case "digital_in":
		pType = sensoraddon.TypeDigitalIn
	case "analog_in":
		pType = sensoraddon.TypeAnalogIn
	default:
		return fmt.Errorf("invalid type %q, must be one of: %s", opts.Type, strings.Join(sensoraddon.ValidPeripheralTypes, ", "))
	}

	addOpts := &sensoraddon.AddOptions{}
	if opts.CID > 0 {
		addOpts.CID = &opts.CID
	}
	if opts.Addr != "" {
		addOpts.Addr = &opts.Addr
	}

	var result map[string]any
	err := cmdutil.RunWithSpinner(ctx, ios, "Adding peripheral...", func(ctx context.Context) error {
		var addErr error
		result, addErr = svc.AddPeripheral(ctx, opts.Device, pType, addOpts)
		return addErr
	})
	if err != nil {
		return err
	}

	// List created components
	components := make([]string, 0, len(result))
	for comp := range result {
		components = append(components, comp)
	}

	ios.Success("Added %s peripheral: %s", opts.Type, strings.Join(components, ", "))
	ios.Warning("Device reboot required for changes to take effect")
	return nil
}
