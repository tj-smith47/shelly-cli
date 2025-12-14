// Package enable provides the thermostat enable command.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/validate"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	Mode    string
}

// NewCommand creates the thermostat enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on", "start"},
		Short:   "Enable thermostat",
		Long: `Enable a thermostat component.

Optionally set the operating mode when enabling:
- heat: Heating mode (opens valve when below target)
- cool: Cooling mode (opens valve when above target)
- auto: Automatic mode (device decides based on conditions)`,
		Example: `  # Enable thermostat with current settings
  shelly thermostat enable gateway

  # Enable in heat mode
  shelly thermostat enable gateway --mode heat

  # Enable specific thermostat in auto mode
  shelly thermostat enable gateway --id 1 --mode auto`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Thermostat component ID")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "Operating mode (heat, cool, auto)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate mode if provided
	if err := validate.ValidateMode(opts.Mode, true); err != nil {
		return err
	}

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	thermostat := conn.Thermostat(opts.ID)

	ios.StartProgress("Enabling thermostat...")
	defer ios.StopProgress()

	// Enable the thermostat
	err = thermostat.Enable(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to enable thermostat: %w", err)
	}

	// Set mode if specified
	if opts.Mode != "" {
		err = thermostat.SetMode(ctx, opts.Mode)
		if err != nil {
			return fmt.Errorf("failed to set thermostat mode: %w", err)
		}
	}

	if opts.Mode != "" {
		ios.Success("Thermostat %d enabled in %s mode", opts.ID, opts.Mode)
	} else {
		ios.Success("Thermostat %d enabled", opts.ID)
	}

	return nil
}
