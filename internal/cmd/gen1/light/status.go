package light

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
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
		Short:   "Get light status",
		Long:    `Get the current status of a Gen1 light/dimmer.`,
		Example: `  # Get light status
  shelly gen1 light status living-room

  # Output as JSON
  shelly gen1 light status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runLightStatus(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Light")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func runLightStatus(ctx context.Context, opts *StatusOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	light, err := gen1Client.Light(opts.ID)
	if err != nil {
		return err
	}

	status, err := light.GetStatus(ctx)
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Light %d Status:", opts.ID)))
	ios.Println()

	stateStr := theme.Dim().Render("OFF")
	if status.IsOn {
		stateStr = theme.StatusOK().Render("ON")
	}
	ios.Printf("  State: %s\n", stateStr)
	ios.Printf("  Brightness: %d%%\n", status.Brightness)

	if status.Mode != "" {
		ios.Printf("  Mode: %s\n", status.Mode)
	}

	if status.Temp > 0 {
		ios.Printf("  Color Temp: %dK\n", status.Temp)
	}

	if status.HasTimer {
		ios.Printf("  Timer: %s\n", theme.Highlight().Render(fmt.Sprintf("%d seconds remaining", status.TimerRemaining)))
	}

	return nil
}
