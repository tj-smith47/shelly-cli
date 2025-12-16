// Package get provides the script get subcommand.
package get

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

var statusFlag bool

// NewCommand creates the script get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <device> <id>",
		Aliases: []string{"code"},
		Short:   "Get script code or status",
		Long: `Get the source code or status of a script on a Gen2+ device.

By default, displays the script source code. Use --status to show
detailed status including memory usage and errors.`,
		Example: `  # Get script code
  shelly script get living-room 1

  # Get script status
  shelly script get living-room 1 --status`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenScriptID(),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid script ID: %s", args[1])
			}
			return run(cmd.Context(), f, args[0], id)
		},
	}

	cmd.Flags().BoolVar(&statusFlag, "status", false, "Show script status instead of code")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	if statusFlag {
		return cmdutil.RunWithSpinner(ctx, ios, "Getting script status...", func(ctx context.Context) error {
			status, err := svc.GetScriptStatus(ctx, device, id)
			if err != nil {
				return fmt.Errorf("failed to get script status: %w", err)
			}
			displayStatus(ios, status)
			return nil
		})
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Getting script code...", func(ctx context.Context) error {
		code, err := svc.GetScriptCode(ctx, device, id)
		if err != nil {
			return fmt.Errorf("failed to get script code: %w", err)
		}
		displayCode(ios, code)
		return nil
	})
}

func displayStatus(ios *iostreams.IOStreams, status *shelly.ScriptStatus) {
	ios.Println(theme.Bold().Render("Script Status"))
	ios.Println("")

	ios.Printf("  ID:      %d\n", status.ID)
	ios.Printf("  Status:  %s\n", output.RenderRunningState(status.Running))
	ios.Println("")

	ios.Println(theme.Bold().Render("Memory"))
	ios.Printf("  Usage:   %d bytes\n", status.MemUsage)
	ios.Printf("  Peak:    %d bytes\n", status.MemPeak)
	ios.Printf("  Free:    %d bytes\n", status.MemFree)

	if len(status.Errors) > 0 {
		ios.Println("")
		ios.Println(theme.StatusError().Render("Errors:"))
		for _, e := range status.Errors {
			ios.Printf("  - %s\n", e)
		}
	}
}

func displayCode(ios *iostreams.IOStreams, code string) {
	if code == "" {
		ios.Info("Script has no code")
		return
	}
	ios.Println(code)
}
