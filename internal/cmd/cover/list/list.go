// Package list provides the cover list subcommand.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the cover list command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List cover components",
		Long:  `List all cover/roller components on the specified device with their current status.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cmd, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, cmd *cobra.Command, device string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*shelly.DefaultTimeout/10)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Fetching cover components...")
	spin.Start()

	covers, err := svc.CoverList(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to list cover components: %w", err)
	}

	if len(covers) == 0 {
		iostreams.NoResults("cover components")
		return nil
	}

	return outputList(cmd, covers)
}

func outputList(cmd *cobra.Command, covers []shelly.CoverInfo) error {
	switch viper.GetString("output") {
	case string(output.FormatJSON):
		return output.JSON(cmd.OutOrStdout(), covers)
	case string(output.FormatYAML):
		return output.YAML(cmd.OutOrStdout(), covers)
	default:
		printTable(covers)
		return nil
	}
}

func printTable(covers []shelly.CoverInfo) {
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
