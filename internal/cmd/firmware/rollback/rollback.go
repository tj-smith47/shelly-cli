// Package rollback provides the firmware rollback subcommand.
package rollback

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var yesFlag bool

// NewCommand creates the firmware rollback command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Check rollback availability
	ios.StartProgress("Checking rollback availability...")
	status, err := svc.GetFirmwareStatus(ctx, device)
	ios.StopProgress()
	if err != nil {
		return fmt.Errorf("failed to get firmware status: %w", err)
	}

	if !status.CanRollback {
		ios.Warning("Rollback is not available for this device")
		ios.Info("Rollback is typically available after a recent update or when in safe mode.")
		return nil
	}

	// Confirm unless --yes
	if !yesFlag {
		ios.Warning("This will rollback the firmware to the previous version.")
		confirmed, confirmErr := ios.Confirm("Proceed with rollback?", false)
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			ios.Warning("Rollback cancelled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	return cmdutil.RunWithSpinner(ctx, ios, "Rolling back firmware...", func(ctx context.Context) error {
		if rollbackErr := svc.RollbackFirmware(ctx, device); rollbackErr != nil {
			return fmt.Errorf("failed to rollback: %w", rollbackErr)
		}
		ios.Success("Firmware rollback started on %s", device)
		ios.Info("The device will reboot automatically.")
		return nil
	})
}
