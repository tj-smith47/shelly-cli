// Package rollback provides the firmware rollback subcommand.
package rollback

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the firmware rollback command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "rollback <device>",
		Aliases: []string{"rb"},
		Short:   "Rollback to previous firmware",
		Long: `Rollback device firmware to the previous version.

This is only available when the device supports rollback (typically after
a recent firmware update or when in safe mode).`,
		Example: `  # Rollback firmware
  shelly firmware rollback living-room

  # Rollback without confirmation
  shelly firmware rollback living-room --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Check rollback availability
	var status *shelly.FirmwareStatus
	err := cmdutil.RunWithSpinner(ctx, ios, "Checking rollback availability...", func(ctx context.Context) error {
		var statusErr error
		status, statusErr = svc.GetFirmwareStatus(ctx, opts.Device)
		return statusErr
	})
	if err != nil {
		return fmt.Errorf("failed to get firmware status: %w", err)
	}

	if !status.CanRollback {
		ios.Warning("Rollback is not available for this device")
		ios.Info("Rollback is typically available after a recent update or when in safe mode.")
		return nil
	}

	// Confirm unless --yes
	ios.Warning("This will rollback the firmware to the previous version.")
	confirmed, confirmErr := opts.Factory.ConfirmAction("Proceed with rollback?", opts.Yes)
	if confirmErr != nil {
		return confirmErr
	}
	if !confirmed {
		ios.Warning("Rollback cancelled")
		return nil
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	return cmdutil.RunWithSpinner(ctx, ios, "Rolling back firmware...", func(ctx context.Context) error {
		if rollbackErr := svc.RollbackFirmware(ctx, opts.Device); rollbackErr != nil {
			return fmt.Errorf("failed to rollback: %w", rollbackErr)
		}
		ios.Success("Firmware rollback started on %s", opts.Device)
		ios.Info("The device will reboot automatically.")
		return nil
	})
}
