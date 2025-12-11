// Package list provides the light list subcommand.
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

// NewCommand creates the light list command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List light components",
		Long:  `List all light components on the specified device with their current status.`,
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

	spin := iostreams.NewSpinner("Fetching light components...")
	spin.Start()

	lights, err := svc.LightList(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to list lights: %w", err)
	}

	if len(lights) == 0 {
		iostreams.NoResults("light components")
		return nil
	}

	return outputList(cmd, lights)
}

func outputList(cmd *cobra.Command, lights []shelly.LightInfo) error {
	switch viper.GetString("output") {
	case string(output.FormatJSON):
		return output.JSON(cmd.OutOrStdout(), lights)
	case string(output.FormatYAML):
		return output.YAML(cmd.OutOrStdout(), lights)
	default:
		printTable(lights)
		return nil
	}
}

func printTable(lights []shelly.LightInfo) {
	t := output.NewTable("ID", "Name", "State", "Brightness", "Power")
	for _, lt := range lights {
		name := lt.Name
		if name == "" {
			name = fmt.Sprintf("light:%d", lt.ID)
		}

		state := theme.StatusError().Render("OFF")
		if lt.Output {
			state = theme.StatusOK().Render("ON")
		}

		brightness := "-"
		if lt.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", lt.Brightness)
		}

		power := "-"
		if lt.Power > 0 {
			power = fmt.Sprintf("%.1f W", lt.Power)
		}

		t.AddRow(fmt.Sprintf("%d", lt.ID), name, state, brightness, power)
	}
	t.Print()
}
