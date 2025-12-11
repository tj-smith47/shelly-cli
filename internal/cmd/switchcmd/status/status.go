// Package status provides the switch status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the switch status command.
func NewCommand() *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show switch status",
		Long:    `Show the current status of a switch component on the specified device.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], switchID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &switchID, "Switch")

	return cmd
}

func run(ctx context.Context, device string, switchID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunStatus(ctx, ios, svc, device, switchID,
		"Fetching switch status...",
		func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.SwitchStatus, error) {
			return svc.SwitchStatus(ctx, device, id)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *model.SwitchStatus) {
	ios.Title("Switch %d Status", status.ID)
	ios.Println()

	state := theme.StatusError().Render("OFF")
	if status.Output {
		state = theme.StatusOK().Render("ON")
	}
	ios.Printf("  State:   %s\n", state)

	if status.Power != nil {
		ios.Printf("  Power:   %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage: %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current: %.3f A\n", *status.Current)
	}
	if status.Energy != nil {
		ios.Printf("  Energy:  %.2f Wh\n", status.Energy.Total)
	}
}
