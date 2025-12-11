// Package list provides the script list subcommand.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the script list command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls"},
		Short:   "List scripts on a device",
		Long: `List all scripts on a Gen2+ Shelly device.

Shows script ID, name, enabled status, and running status.`,
		Example: `  # List all scripts
  shelly script list living-room`,
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

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Getting scripts...",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.ScriptInfo, error) {
			return svc.ListScripts(ctx, device)
		},
		displayScripts)
}

func displayScripts(ios *iostreams.IOStreams, scripts []shelly.ScriptInfo) {
	if len(scripts) == 0 {
		ios.Info("No scripts found on this device")
		return
	}

	table := output.NewTable("ID", "Name", "Enabled", "Running")
	for _, s := range scripts {
		name := s.Name
		if name == "" {
			name = theme.Dim().Render("(unnamed)")
		}
		enabled := theme.Dim().Render("no")
		if s.Enable {
			enabled = theme.StatusOK().Render("yes")
		}
		running := theme.Dim().Render("stopped")
		if s.Running {
			running = theme.StatusOK().Render("running")
		}
		table.AddRow(fmt.Sprintf("%d", s.ID), name, enabled, running)
	}
	table.PrintTo(ios.Out)
}
