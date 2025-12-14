// Package list provides the cover list subcommand.
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

// NewCommand creates the cover list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List cover components",
		Long:    `List all cover/roller components on the specified device with their current status.`,
		Example: `  # List all covers on a device
  shelly cover list bedroom

  # List covers with JSON output
  shelly cv ls bedroom -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*shelly.DefaultTimeout/10)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Fetching cover components...",
		"cover components",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.CoverInfo, error) {
			return svc.CoverList(ctx, device)
		},
		displayList)
}

func displayList(ios *iostreams.IOStreams, covers []shelly.CoverInfo) {
	t := output.NewTable("ID", "Name", "State", "Position", "Power")
	for _, cover := range covers {
		name := cover.Name
		if name == "" {
			name = fmt.Sprintf("cover:%d", cover.ID)
		}

		stateStyle := theme.StatusWarn()
		switch cover.State {
		case "open":
			stateStyle = theme.StatusOK()
		case "closed":
			stateStyle = theme.StatusError()
		}
		state := stateStyle.Render(cover.State)

		position := "-"
		if cover.Position >= 0 {
			position = fmt.Sprintf("%d%%", cover.Position)
		}

		power := "-"
		if cover.Power > 0 {
			power = fmt.Sprintf("%.1f W", cover.Power)
		}

		t.AddRow(fmt.Sprintf("%d", cover.ID), name, state, position, power)
	}
	t.Print()
}
