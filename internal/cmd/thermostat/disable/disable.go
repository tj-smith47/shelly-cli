// Package disable provides the thermostat disable command.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

// NewCommand creates the thermostat disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off", "stop"},
		Short:   "Disable thermostat",
		Long: `Disable a thermostat component.

When disabled, the thermostat will not actively control the valve
position based on temperature. The valve will typically remain
in its current position or close.`,
		Example: `  # Disable thermostat
  shelly thermostat disable gateway

  # Disable specific thermostat
  shelly thermostat disable gateway --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Thermostat component ID")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	thermostat := conn.Thermostat(opts.ID)

	ios.StartProgress("Disabling thermostat...")
	err = thermostat.Enable(ctx, false)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to disable thermostat: %w", err)
	}

	ios.Success("Thermostat %d disabled", opts.ID)
	return nil
}
