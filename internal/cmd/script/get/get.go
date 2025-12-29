// Package get provides the script get subcommand.
package get

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory    *cmdutil.Factory
	Device     string
	ID         int
	ShowStatus bool
}

// NewCommand creates the script get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			opts.ID = id
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.ShowStatus, "status", false, "Show script status instead of code")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	if opts.ShowStatus {
		return cmdutil.RunWithSpinner(ctx, ios, "Getting script status...", func(ctx context.Context) error {
			status, err := svc.GetScriptStatus(ctx, opts.Device, opts.ID)
			if err != nil {
				return fmt.Errorf("failed to get script status: %w", err)
			}
			term.DisplayScriptStatus(ios, status)
			return nil
		})
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Getting script code...", func(ctx context.Context) error {
		code, err := svc.GetScriptCode(ctx, opts.Device, opts.ID)
		if err != nil {
			return fmt.Errorf("failed to get script code: %w", err)
		}
		term.DisplayScriptCode(ios, code)
		return nil
	})
}
