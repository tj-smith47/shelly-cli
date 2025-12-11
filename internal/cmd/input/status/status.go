// Package status provides the input status subcommand.
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

// NewCommand creates the input status command.
func NewCommand() *cobra.Command {
	var inputID int

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show input status",
		Long:    `Display the status of an input component on a Shelly device.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], inputID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &inputID, "Input")

	return cmd
}

func run(ctx context.Context, device string, inputID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunStatus(ctx, ios, svc, device, inputID,
		"Getting input status...",
		func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.InputStatus, error) {
			return svc.InputStatus(ctx, device, id)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *model.InputStatus) {
	ios.Title("Input %d Status", status.ID)
	ios.Println()

	state := theme.StatusError().Render("inactive")
	if status.State {
		state = theme.StatusOK().Render("active")
	}
	ios.Printf("  State: %s\n", state)

	if status.Type != "" {
		ios.Printf("  Type:  %s\n", status.Type)
	}
}
