package roller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// StatusOptions holds status command options.
type StatusOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	JSON    bool
}

func newStatusCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &StatusOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "get"},
		Short:   "Get roller status",
		Long:    `Get the current status of a Gen1 roller/cover.`,
		Example: `  # Get roller status
  shelly gen1 roller status living-room

  # Output as JSON
  shelly gen1 roller status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runRollerStatus(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Roller")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func runRollerStatus(ctx context.Context, opts *StatusOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	roller, err := gen1Client.Roller(opts.ID)
	if err != nil {
		return err
	}

	status, err := roller.GetStatus(ctx)
	if err != nil {
		return err
	}

	if opts.JSON {
		output, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("Roller %d Status:", opts.ID)))
	ios.Println()

	stateStr := formatState(status.State)
	ios.Printf("  State: %s\n", stateStr)

	if status.CurrentPos >= 0 && status.IsValid {
		ios.Printf("  Position: %d%%\n", status.CurrentPos)
	}

	if status.LastDirection != "" {
		ios.Printf("  Last Direction: %s\n", status.LastDirection)
	}

	if status.Power > 0 {
		ios.Printf("  Power: %.1f W\n", status.Power)
	}

	if status.Calibrating {
		ios.Printf("  Calibrating: %s\n", theme.Highlight().Render("Yes"))
	}

	if status.SafetySwitch {
		ios.Printf("  Safety Switch: %s\n", theme.StatusWarn().Render("Triggered"))
	}

	if status.Overtemperature {
		ios.Printf("  Overtemperature: %s\n", theme.StatusError().Render("Yes"))
	}

	return nil
}

func formatState(state string) string {
	switch state {
	case "open":
		return theme.StatusOK().Render("Opening")
	case "close":
		return theme.StatusOK().Render("Closing")
	case "stop":
		return theme.Dim().Render("Stopped")
	default:
		return theme.Dim().Render(state)
	}
}
