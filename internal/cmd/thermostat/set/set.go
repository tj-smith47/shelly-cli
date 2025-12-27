// Package set provides the thermostat set command.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory   *cmdutil.Factory
	Device    string
	ID        int
	Target    float64
	TargetSet bool // Tracks if --target was explicitly provided
	Mode      string
	Enable    bool
	Disable   bool
}

// NewCommand creates the thermostat set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"config", "configure"},
		Short:   "Set thermostat configuration",
		Long: `Set thermostat configuration options.

Allows setting:
- Target temperature (--target)
- Operating mode (--mode): heat, cool, or auto
- Enable/disable state (--enable/--disable)`,
		Example: `  # Set target temperature to 22°C
  shelly thermostat set gateway --target 22

  # Set mode to heat
  shelly thermostat set gateway --mode heat

  # Set target and mode together
  shelly thermostat set gateway --target 21 --mode auto

  # Enable thermostat
  shelly thermostat set gateway --enable`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.TargetSet = cmd.Flags().Changed("target")
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Thermostat component ID")
	cmd.Flags().Float64Var(&opts.Target, "target", 0, "Target temperature in Celsius")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "Operating mode (heat, cool, auto)")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable thermostat")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable thermostat")

	cmd.MarkFlagsMutuallyExclusive("enable", "disable")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate that at least one option is set
	if !opts.TargetSet && opts.Mode == "" && !opts.Enable && !opts.Disable {
		return fmt.Errorf("at least one of --target, --mode, --enable, or --disable must be specified")
	}

	// Validate mode
	if err := shelly.ValidateThermostatMode(opts.Mode, true); err != nil {
		return err
	}

	// Build config
	config := &components.ThermostatConfig{}
	changes := []string{}

	if opts.TargetSet {
		config.TargetC = &opts.Target
		changes = append(changes, fmt.Sprintf("target temperature: %.1f°C", opts.Target))
	}

	if opts.Mode != "" {
		config.ThermostatMode = &opts.Mode
		changes = append(changes, fmt.Sprintf("mode: %s", opts.Mode))
	}

	if opts.Enable {
		enable := true
		config.Enable = &enable
		changes = append(changes, "enabled: true")
	}

	if opts.Disable {
		disable := false
		config.Enable = &disable
		changes = append(changes, "enabled: false")
	}

	return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		thermostat := dev.Gen2().Thermostat(opts.ID)

		err := cmdutil.RunWithSpinner(ctx, ios, "Updating thermostat configuration...", func(ctx context.Context) error {
			return thermostat.SetConfig(ctx, config)
		})
		if err != nil {
			return fmt.Errorf("failed to set thermostat config: %w", err)
		}

		ios.Success("Thermostat %d configuration updated", opts.ID)
		for _, change := range changes {
			ios.Printf("  • %s\n", change)
		}

		return nil
	})
}
