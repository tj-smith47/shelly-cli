// Package toggle provides the switch toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:   "toggle <device>",
		Short: "Toggle switch state",
		Long:  `Toggle the state of a switch component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], switchID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &switchID, "Switch")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, switchID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Toggling switch...", func(ctx context.Context) error {
		status, err := svc.SwitchToggle(ctx, device, switchID)
		if err != nil {
			return err
		}

		state := "off"
		if status.Output {
			state = "on"
		}
		ios.Success("Switch %d toggled to %s", switchID, state)
		return nil
	})
}
