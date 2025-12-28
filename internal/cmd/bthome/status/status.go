// Package status provides the bthome status command.
package status

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory *cmdutil.Factory
	Device  string
	ID      int
	HasID   bool
}

// NewCommand creates the bthome status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device> [id]",
		Aliases: []string{"st", "info"},
		Short:   "Show BTHome device status",
		Long: `Show detailed status of BTHome devices.

Without an ID, shows the BTHome component status including any active
discovery scan. With an ID, shows detailed status of a specific
BTHomeDevice including signal strength, battery, and known objects.`,
		Example: `  # Show BTHome component status
  shelly bthome status living-room

  # Show specific device status
  shelly bthome status living-room 200

  # Output as JSON
  shelly bthome status living-room 200 --json`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) > 1 {
				id, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid device ID: %w", err)
				}
				opts.ID = id
				opts.HasID = true
			}
			return run(cmd.Context(), opts)
		},
	}

	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "text", "text", "json")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if opts.HasID {
		status, err := svc.Wireless().FetchBTHomeDeviceStatus(ctx, opts.Device, opts.ID)
		if err != nil {
			return err
		}

		if opts.Format == "json" {
			return output.JSON(ios.Out, status)
		}

		term.DisplayBTHomeDeviceStatus(ios, status)
		return nil
	}

	status, err := svc.Wireless().FetchBTHomeComponentStatus(ctx, opts.Device)
	if err != nil {
		return err
	}

	if opts.Format == "json" {
		return output.JSON(ios.Out, status)
	}

	term.DisplayBTHomeComponentStatus(ios, status)
	return nil
}
