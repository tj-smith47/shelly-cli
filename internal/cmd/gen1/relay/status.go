package relay

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
		Short:   "Get relay status",
		Long:    `Get the current status of a Gen1 relay switch.`,
		Example: `  # Get relay status
  shelly gen1 relay status living-room

  # Get status of relay 1 (second relay)
  shelly gen1 relay status living-room --id 1

  # Output as JSON
  shelly gen1 relay status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runStatus(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Relay")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func runStatus(ctx context.Context, opts *StatusOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	relay, err := gen1Client.Relay(opts.ID)
	if err != nil {
		return err
	}

	status, err := relay.GetStatus(ctx)
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Relay %d Status:", opts.ID)))
	ios.Println()

	stateStr := theme.Dim().Render("OFF")
	if status.IsOn {
		stateStr = theme.StatusOK().Render("ON")
	}
	ios.Printf("  State: %s\n", stateStr)

	if status.Source != "" {
		ios.Printf("  Source: %s\n", status.Source)
	}

	if status.HasTimer {
		ios.Printf("  Timer: %s\n", theme.Highlight().Render(fmt.Sprintf("%d seconds remaining", status.TimerRemaining)))
	}

	if status.Overpower {
		ios.Printf("  Overpower: %s\n", theme.StatusError().Render("Yes"))
	}

	return nil
}
