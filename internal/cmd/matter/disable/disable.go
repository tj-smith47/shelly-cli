// Package disable provides the matter disable command.
package disable

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the matter disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off", "deactivate"},
		Short:   "Disable Matter on a device",
		Long: `Disable Matter connectivity on a Shelly device.

When Matter is disabled, the device will no longer be controllable
through Matter fabrics. Existing fabric pairings are preserved but
will not function until Matter is re-enabled.

To permanently remove all Matter pairings, use 'shelly matter reset'
instead.`,
		Example: `  # Disable Matter
  shelly matter disable living-room

  # Re-enable later
  shelly matter enable living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	err := cmdutil.RunWithSpinner(ctx, ios, "Disabling Matter...", func(ctx context.Context) error {
		return svc.MatterDisable(ctx, opts.Device)
	})
	if err != nil {
		return err
	}

	ios.Success("Matter disabled.")
	ios.Info("Fabric pairings are preserved but inactive.")
	ios.Info("Re-enable with: shelly matter enable %s", opts.Device)

	return nil
}
