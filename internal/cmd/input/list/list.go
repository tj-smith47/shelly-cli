// Package list provides the input list subcommand.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the input list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List input components",
		Long:  `List all input components on a Shelly device with their current state.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Getting inputs...",
		"inputs",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.InputInfo, error) {
			return svc.InputList(ctx, device)
		},
		displayList)
}

func displayList(ios *iostreams.IOStreams, inputs []shelly.InputInfo) {
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
	ios.Count("input", len(inputs))
}
