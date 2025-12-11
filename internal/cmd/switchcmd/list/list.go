// Package list provides the switch list subcommand.
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

// NewCommand creates the switch list command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List switch components",
		Long:  `List all switch components on the specified device with their current status.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0])
		},
	}

	return cmd
}

func run(cmd *cobra.Command, device string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*shelly.DefaultTimeout/10) // 15s
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Fetching switch components...")
	spin.Start()

	switches, err := svc.SwitchList(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to list switches: %w", err)
	}

	if len(switches) == 0 {
		iostreams.NoResults("switch components")
		return nil
	}

	return outputList(cmd, switches)
}

func outputList(cmd *cobra.Command, switches []shelly.SwitchInfo) error {
	switch viper.GetString("output") {
	case string(output.FormatJSON):
		return output.JSON(cmd.OutOrStdout(), switches)
	case string(output.FormatYAML):
		return output.YAML(cmd.OutOrStdout(), switches)
	default:
		printTable(switches)
		return nil
	}
}

func printTable(switches []shelly.SwitchInfo) {
	t := output.NewTable("ID", "Name", "State", "Power")
	for _, sw := range switches {
		name := sw.Name
		if name == "" {
			name = fmt.Sprintf("switch:%d", sw.ID)
		}

		state := theme.StatusError().Render("OFF")
		if sw.Output {
			state = theme.StatusOK().Render("ON")
		}

		power := "-"
		if sw.Power > 0 {
			power = fmt.Sprintf("%.1f W", sw.Power)
		}

		t.AddRow(fmt.Sprintf("%d", sw.ID), name, state, power)
	}
	t.Print()
}
