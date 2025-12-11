// Package status provides the cover status subcommand.
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

// NewCommand creates the cover status command.
func NewCommand() *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:   "status <device>",
		Short: "Show cover status",
		Long:  `Show the current status of a cover component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0], coverID)
		},
	}

	cmd.Flags().IntVarP(&coverID, "id", "i", 0, "Cover ID (default 0)")

	return cmd
}

func run(cmd *cobra.Command, device string, coverID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Fetching cover status...")
	spin.Start()

	status, err := svc.CoverStatus(ctx, device, coverID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get cover status: %w", err)
	}

	return outputStatus(cmd, status)
}

func outputStatus(cmd *cobra.Command, status *model.CoverStatus) error {
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

func displayStatus(status *model.CoverStatus) {
	iostreams.Title("Cover %d Status", status.ID)
	fmt.Println()

	stateStyle := theme.StatusWarn()
	switch status.State {
	case "open":
		stateStyle = theme.StatusOK()
	case "closed":
		stateStyle = theme.StatusError()
	}
	fmt.Printf("  State:    %s\n", stateStyle.Render(status.State))

	if status.CurrentPosition != nil {
		fmt.Printf("  Position: %d%%\n", *status.CurrentPosition)
	}
	if status.Power != nil {
		fmt.Printf("  Power:    %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		fmt.Printf("  Voltage:  %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		fmt.Printf("  Current:  %.3f A\n", *status.Current)
	}
}
