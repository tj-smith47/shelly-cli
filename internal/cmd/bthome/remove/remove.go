// Package remove provides the bthome remove command.
package remove

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
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
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if !opts.Yes {
		ios.Warning("This will remove BTHome device %d and its sensors.", opts.ID)
		ios.Printf("Continue? [y/N]: ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			ios.Debug("failed to read response: %v", err)
			return fmt.Errorf("operation cancelled")
		}

		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			ios.Info("Operation cancelled.")
			return nil
		}
	}

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		params := map[string]any{
			"id": opts.ID,
		}

		_, err := conn.Call(ctx, "BTHome.DeleteDevice", params)
		if err != nil {
			return fmt.Errorf("failed to remove device: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	ios.Success("BTHome device %d removed.", opts.ID)

	return nil
}
