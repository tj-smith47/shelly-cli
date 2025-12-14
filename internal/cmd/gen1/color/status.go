package color

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

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
		Short:   "Get color light status",
		Long:    `Get the current status of a Gen1 RGBW color light.`,
		Example: `  # Get color status
  shelly gen1 color status living-room

  # Output as JSON
  shelly gen1 color status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runColorStatus(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Color")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func runColorStatus(ctx context.Context, opts *StatusOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	color := gen1Client.Color(opts.ID)
	status, err := color.GetStatus(ctx)
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Color %d Status:", opts.ID)))
	ios.Println()

	stateStr := theme.Dim().Render("OFF")
	if status.IsOn {
		stateStr = theme.StatusOK().Render("ON")
	}
	ios.Printf("  State: %s\n", stateStr)

	if status.Mode != "" {
		ios.Printf("  Mode: %s\n", status.Mode)
	}

	// Display RGB values
	ios.Printf("  RGB: %s\n",
		theme.Highlight().Render(fmt.Sprintf("(%d, %d, %d)", status.Red, status.Green, status.Blue)))

	if status.White > 0 {
		ios.Printf("  White: %d\n", status.White)
	}

	ios.Printf("  Gain: %d%%\n", status.Gain)

	if status.Effect > 0 {
		ios.Printf("  Effect: %d\n", status.Effect)
	}

	if status.HasTimer {
		ios.Printf("  Timer: %s\n", theme.Highlight().Render(fmt.Sprintf("%d seconds remaining", status.TimerRemaining)))
	}

	return nil
}
