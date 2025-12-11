// Package status provides the input status subcommand.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], inputID)
		},
	}

	cmd.Flags().IntVarP(&inputID, "id", "i", 0, "Input ID (default 0)")

	return cmd
}

func run(device string, inputID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Getting input status...")
	spin.Start()

	status, err := svc.InputStatus(ctx, device, inputID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get input status: %w", err)
	}

	state := theme.StatusError().Render("inactive")
	if status.State {
		state = theme.StatusOK().Render("active")
	}

	iostreams.Info("Input %d: %s", inputID, state)
	if status.Type != "" {
		iostreams.Info("Type: %s", status.Type)
	}

	return nil
}
