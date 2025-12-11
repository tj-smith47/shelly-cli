// Package status provides the device status subcommand.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the device status command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show device status",
		Long:    `Display the full status of a Shelly device including all components.`,
		Example: `  # Show status for a device
  shelly device status living-room

  # Using alias
  shelly dev st bedroom`,
		Args: cobra.ExactArgs(1),
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

	spin := iostreams.NewSpinner("Getting device status...")
	spin.Start()

	status, err := svc.DeviceStatus(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get device status: %w", err)
	}

	// Print device info
	iostreams.Info("Device: %s", theme.Bold().Render(status.Info.ID))
	iostreams.Info("Model: %s (Gen%d)", status.Info.Model, status.Info.Generation)
	iostreams.Info("Firmware: %s", status.Info.Firmware)
	iostreams.Info("")

	// Print status table
	table := output.NewTable("Component", "Value")

	for key, value := range status.Status {
		table.AddRow(key, formatValue(value))
	}

	table.Print()
	return nil
}

func formatValue(v any) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case map[string]any:
		return fmt.Sprintf("{%d fields}", len(val))
	case []any:
		return fmt.Sprintf("[%d items]", len(val))
	default:
		return fmt.Sprintf("%v", val)
	}
}
