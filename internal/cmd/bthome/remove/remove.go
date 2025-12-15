// Package remove provides the bthome remove command.
package remove

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	Yes     bool
}

// NewCommand creates the bthome remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "remove <device> <id>",
		Aliases: []string{"rm", "delete", "del"},
		Short:   "Remove a BTHome device",
		Long: `Remove a BTHome device from a Shelly gateway.

This removes the BTHomeDevice component and any associated BTHomeSensor
components. The physical device will no longer be tracked by the gateway.

Use 'shelly bthome list' to see device IDs.`,
		Example: `  # Remove BTHome device with ID 200
  shelly bthome remove living-room 200

  # Remove without confirmation
  shelly bthome remove living-room 200 --yes`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid device ID: %w", err)
			}
			opts.ID = id
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	ios.Warning("This will remove BTHome device %d and its sensors.", opts.ID)
	confirmed, err := opts.Factory.ConfirmAction("Continue?", opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Operation cancelled.")
		return nil
	}

	if err := svc.BTHomeRemoveDevice(ctx, opts.Device, opts.ID); err != nil {
		return err
	}

	ios.Success("BTHome device %d removed.", opts.ID)

	return nil
}
