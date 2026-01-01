// Package list provides the thermostat list command.
package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the thermostat list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List thermostats",
		Long: `List all thermostat components on a Shelly device.

Thermostat components are typically found on Shelly BLU TRV (Thermostatic
Radiator Valve) devices connected via BLU Gateway. Each thermostat has
an ID, enabled state, and target temperature.

Use 'shelly thermostat status' for detailed readings including current
temperature. Use 'shelly thermostat set' to adjust target temperature.

Output is formatted as styled text by default. Use --json for
structured output suitable for scripting.`,
		Example: `  # List thermostats
  shelly thermostat list gateway

  # Output as JSON
  shelly thermostat list gateway --json

  # Get enabled thermostats only
  shelly thermostat list gateway --json | jq '.[] | select(.enabled == true)'

  # Get target temperatures
  shelly thermostat list gateway --json | jq '.[] | {id, target_c}'

  # Find thermostats set above 22°C
  shelly thermostat list gateway --json | jq '.[] | select(.target_c > 22)'

  # Count active thermostats
  shelly thermostat list gateway --json | jq '[.[] | select(.enabled)] | length'

  # Short form
  shelly thermostat ls gateway`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "text", "text", "json")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var thermostats []model.ThermostatInfo
	err := svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		conn := dev.Gen2()

		// Get full device status to find thermostats
		result, callErr := conn.Call(ctx, "Shelly.GetStatus", nil)
		if callErr != nil {
			return fmt.Errorf("failed to get device status: %w", callErr)
		}

		// Parse status to find thermostat components
		jsonBytes, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal result: %w", marshalErr)
		}

		var fullStatus map[string]json.RawMessage
		if unmarshalErr := json.Unmarshal(jsonBytes, &fullStatus); unmarshalErr != nil {
			return fmt.Errorf("failed to parse status: %w", unmarshalErr)
		}

		thermostats = shelly.CollectThermostats(fullStatus)
		return nil
	})
	if err != nil {
		return err
	}

	// Output in requested format
	if opts.Format == "json" {
		jsonBytes, jsonErr := json.MarshalIndent(thermostats, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("failed to format JSON: %w", jsonErr)
		}
		ios.Println(string(jsonBytes))
		return nil
	}

	term.DisplayThermostats(ios, thermostats, opts.Device)
	return nil
}
