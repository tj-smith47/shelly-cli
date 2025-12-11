// Package list provides the input list subcommand.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the input list command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List input components",
		Long:  `List all input components on a Shelly device with their current state.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Getting inputs...")
	spin.Start()

	inputs, err := svc.InputList(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to list inputs: %w", err)
	}

	if len(inputs) == 0 {
		iostreams.Info("No inputs found on this device")
		return nil
	}

	table := output.NewTable("ID", "Name", "Type", "State")

	for _, input := range inputs {
		name := input.Name
		if name == "" {
			name = theme.Dim().Render("-")
		}

		state := theme.StatusError().Render("inactive")
		if input.State {
			state = theme.StatusOK().Render("active")
		}

		table.AddRow(
			fmt.Sprintf("%d", input.ID),
			name,
			input.Type,
			state,
		)
	}

	table.Print()
	iostreams.Count("input", len(inputs))

	return nil
}
