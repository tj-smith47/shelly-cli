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
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var inputID int

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show input status",
		Long:    `Display the status of an input component on a Shelly device.`,
		Example: `  # Show input status
  shelly input status living-room

  # Show specific input by ID
  shelly input status living-room --id 1

  # Output as JSON for scripting
  shelly input status living-room -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], inputID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &inputID, "Input")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, inputID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

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
