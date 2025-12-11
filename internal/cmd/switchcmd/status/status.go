// Package status provides the switch status subcommand.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the switch status command.
func NewCommand() *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show switch status",
		Long:    `Show the current status of a switch component on the specified device.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0], switchID)
		},
	}

	cmd.Flags().IntVarP(&switchID, "id", "i", 0, "Switch ID (default 0)")

	return cmd
}

func run(cmd *cobra.Command, device string, switchID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Fetching switch status...")
	spin.Start()

	status, err := svc.SwitchStatus(ctx, device, switchID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get switch status: %w", err)
	}

	return outputStatus(cmd, status)
}

func outputStatus(cmd *cobra.Command, status *model.SwitchStatus) error {
	switch viper.GetString("output") {
	case string(output.FormatJSON):
		return output.JSON(cmd.OutOrStdout(), status)
	case string(output.FormatYAML):
		return output.YAML(cmd.OutOrStdout(), status)
	default:
		displayStatus(status)
		return nil
	}
}

func displayStatus(status *model.SwitchStatus) {
	iostreams.Title("Switch %d Status", status.ID)
	fmt.Println()

	state := theme.StatusError().Render("OFF")
	if status.Output {
		state = theme.StatusOK().Render("ON")
	}
	fmt.Printf("  State:   %s\n", state)

	if status.Power != nil {
		fmt.Printf("  Power:   %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		fmt.Printf("  Voltage: %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		fmt.Printf("  Current: %.3f A\n", *status.Current)
	}
	if status.Energy != nil {
		fmt.Printf("  Energy:  %.2f Wh\n", status.Energy.Total)
	}
}
