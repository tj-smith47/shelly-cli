// Package status provides the thermostat status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the thermostat status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "show", "info"},
		Short:   "Show thermostat status",
		Long: `Show detailed status of a thermostat component.

Displays:
- Current temperature
- Target temperature
- Valve position (0-100%)
- Operating mode
- Boost/Override status
- Errors and flags`,
		Example: `  # Show thermostat status (default ID 0)
  shelly thermostat status gateway

  # Show specific thermostat
  shelly thermostat status gateway --id 1

  # Output as JSON
  shelly thermostat status gateway --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Thermostat")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var status *components.ThermostatStatus
	err := svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		thermostat := dev.Gen2().Thermostat(opts.ID)

		return cmdutil.RunWithSpinner(ctx, ios, "Getting thermostat status...", func(ctx context.Context) error {
			var statusErr error
			status, statusErr = thermostat.GetStatus(ctx)
			return statusErr
		})
	})
	if err != nil {
		return err
	}

	if opts.JSON {
		jsonBytes, jsonErr := json.MarshalIndent(status, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("failed to format JSON: %w", jsonErr)
		}
		ios.Println(string(jsonBytes))
		return nil
	}

	term.DisplayThermostatStatus(ios, status, opts.ID)
	return nil
}
