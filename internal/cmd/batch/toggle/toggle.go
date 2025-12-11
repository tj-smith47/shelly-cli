// Package toggle provides the batch toggle subcommand.
package toggle

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the batch toggle command.
func NewCommand() *cobra.Command {
	var (
		groupName  string
		all        bool
		timeout    time.Duration
		switchID   int
		concurrent int
	)

	cmd := &cobra.Command{
		Use:   "toggle [device...]",
		Short: "Toggle devices",
		Long: `Toggle multiple devices simultaneously.

By default, toggles switch component 0 on each device.
Use --switch to specify a different component ID.

Devices can be specified by name directly, by group membership,
or by using --all to target all registered devices.`,
		Example: `  # Toggle all devices in a group
  shelly batch toggle --group living-room

  # Toggle specific devices
  shelly batch toggle light-1 light-2

  # Toggle all registered devices
  shelly batch toggle --all

  # Toggle switch 1 on all devices in group
  shelly batch toggle --group bedroom --switch 1

  # Control concurrency and timeout
  shelly batch toggle --all --concurrent 10 --timeout 30s`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := helpers.ResolveBatchTargets(groupName, all, args)
			if err != nil {
				return err
			}
			return run(cmd.Context(), targets, switchID, timeout, concurrent)
		},
	}

	cmd.Flags().StringVarP(&groupName, "group", "g", "", "Target device group")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Target all registered devices")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	cmd.Flags().IntVarP(&switchID, "switch", "s", 0, "Switch component ID")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Max concurrent operations")

	return cmd
}

func run(ctx context.Context, targets []string, switchID int, timeout time.Duration, concurrent int) error {
	if len(targets) == 0 {
		return fmt.Errorf("no target devices specified")
	}

	ios := iostreams.System()
	svc := shelly.NewService()

	ctx, cancel := context.WithTimeout(ctx, timeout*time.Duration(len(targets)))
	defer cancel()

	return cmdutil.RunBatch(ctx, ios, svc, targets, concurrent, func(ctx context.Context, svc *shelly.Service, device string) error {
		_, err := svc.SwitchToggle(ctx, device, switchID)
		return err
	})
}
