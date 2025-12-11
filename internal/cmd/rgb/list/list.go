// Package list provides the rgb list subcommand.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the rgb list command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List RGB components",
		Long:  `List all RGB light components on the specified device with their current status.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0])
		},
	}

	return cmd
}

func run(cmd *cobra.Command, device string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*shelly.DefaultTimeout/10)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Fetching RGB components...")
	spin.Start()

	rgbs, err := svc.RGBList(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to list RGB components: %w", err)
	}

	if len(rgbs) == 0 {
		iostreams.NoResults("RGB components")
		return nil
	}

	return outputList(cmd, rgbs)
}

func outputList(cmd *cobra.Command, rgbs []shelly.RGBInfo) error {
	switch viper.GetString("output") {
	case string(output.FormatJSON):
		return output.JSON(cmd.OutOrStdout(), rgbs)
	case string(output.FormatYAML):
		return output.YAML(cmd.OutOrStdout(), rgbs)
	default:
		printTable(rgbs)
		return nil
	}
}

func printTable(rgbs []shelly.RGBInfo) {
	t := output.NewTable("ID", "Name", "State", "Color", "Brightness", "Power")
	for _, rgb := range rgbs {
		name := rgb.Name
		if name == "" {
			name = fmt.Sprintf("rgb:%d", rgb.ID)
		}

		state := theme.StatusError().Render("OFF")
		if rgb.Output {
			state = theme.StatusOK().Render("ON")
		}

		color := fmt.Sprintf("R:%d G:%d B:%d", rgb.Red, rgb.Green, rgb.Blue)

		brightness := "-"
		if rgb.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", rgb.Brightness)
		}

		power := "-"
		if rgb.Power > 0 {
			power = fmt.Sprintf("%.1f W", rgb.Power)
		}

		t.AddRow(fmt.Sprintf("%d", rgb.ID), name, state, color, brightness, power)
	}
	t.Print()
}
