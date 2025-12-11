// Package status provides the rgb status subcommand.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the rgb status command.
func NewCommand() *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:   "status <device>",
		Short: "Show RGB status",
		Long:  `Show the current status of an RGB light component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0], rgbID)
		},
	}

	cmd.Flags().IntVarP(&rgbID, "id", "i", 0, "RGB ID (default 0)")

	return cmd
}

func run(cmd *cobra.Command, device string, rgbID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Fetching RGB status...")
	spin.Start()

	status, err := svc.RGBStatus(ctx, device, rgbID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get RGB status: %w", err)
	}

	return outputStatus(cmd, status)
}

func outputStatus(cmd *cobra.Command, status *model.RGBStatus) error {
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

func displayStatus(status *model.RGBStatus) {
	iostreams.Title("RGB %d Status", status.ID)
	fmt.Println()

	state := theme.StatusError().Render("OFF")
	if status.Output {
		state = theme.StatusOK().Render("ON")
	}
	fmt.Printf("  State:      %s\n", state)

	if status.RGB != nil {
		fmt.Printf("  Color:      R:%d G:%d B:%d\n",
			status.RGB.Red, status.RGB.Green, status.RGB.Blue)
	}
	if status.Brightness != nil {
		fmt.Printf("  Brightness: %d%%\n", *status.Brightness)
	}
	if status.Power != nil {
		fmt.Printf("  Power:      %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		fmt.Printf("  Voltage:    %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		fmt.Printf("  Current:    %.3f A\n", *status.Current)
	}
}
