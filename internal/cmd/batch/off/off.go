// Package off provides the batch off subcommand.
package off

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the batch off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		groupName  string
		all        bool
		timeout    time.Duration
		switchID   int
		concurrent int
	)

	cmd := &cobra.Command{
		Use:     "off [device...]",
		Aliases: []string{"disable"},
		Short:   "Turn off devices",
		Long: `Turn off multiple devices simultaneously.

By default, turns off switch component 0 on each device.
Use --switch to specify a different component ID.

Target devices can be specified multiple ways:
  - As arguments: device names or addresses
  - Via stdin: pipe device names (one per line or space-separated)
  - Via group: --group flag targets all devices in a group
  - Via all: --all flag targets all registered devices

Priority: explicit args > stdin > group > all

Stdin input supports comments (lines starting with #) and
blank lines are ignored, making it easy to use device lists
from files or other commands.`,
		Example: `  # Turn off specific devices
  shelly batch off light-1 light-2

  # Turn off all devices in a group
  shelly batch off --group living-room

  # Turn off all registered devices
  shelly batch off --all

  # Turn off switch 1 on all devices in group
  shelly batch off --group bedroom --switch 1

  # Control concurrency and timeout
  shelly batch off --all --concurrent 10 --timeout 30s

  # Pipe device names from a file
  cat devices.txt | shelly batch off

  # Pipe from device list command
  shelly device list -o json | jq -r '.[].name' | shelly batch off

  # Turn off only devices that are currently on
  shelly device list -o json | jq -r '.[] | select(.online) | .name' | shelly batch off`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := helpers.ResolveBatchTargets(groupName, all, args)
			if err != nil {
				return err
			}
			return run(cmd.Context(), f, targets, switchID, timeout, concurrent)
		},
	}

	cmd.Flags().StringVarP(&groupName, "group", "g", "", "Target device group")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Target all registered devices")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	cmd.Flags().IntVarP(&switchID, "switch", "s", 0, "Switch component ID")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Max concurrent operations")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, targets []string, switchID int, timeout time.Duration, concurrent int) error {
	if len(targets) == 0 {
		return fmt.Errorf("no target devices specified")
	}

	ios := f.IOStreams()
	svc := f.ShellyService()

	ctx, cancel := context.WithTimeout(ctx, timeout*time.Duration(len(targets)))
	defer cancel()

	return cmdutil.RunBatch(ctx, ios, svc, targets, concurrent, func(ctx context.Context, svc *shelly.Service, device string) error {
		return svc.SwitchOff(ctx, device, switchID)
	})
}
