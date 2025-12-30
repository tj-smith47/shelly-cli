// Package log provides the log command for CLI logging management.
package log

import (
	"github.com/spf13/cobra"

	logclear "github.com/tj-smith47/shelly-cli/internal/cmd/log/clearcmd"
	logexport "github.com/tj-smith47/shelly-cli/internal/cmd/log/export"
	logpath "github.com/tj-smith47/shelly-cli/internal/cmd/log/path"
	logshow "github.com/tj-smith47/shelly-cli/internal/cmd/log/show"
	logtail "github.com/tj-smith47/shelly-cli/internal/cmd/log/tail"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the log command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Aliases: []string{"logs"},
		Short:   "Manage CLI logs",
		Long: `Manage Shelly CLI log files.

Log files are stored in the CLI config directory and contain
debug information about CLI operations.`,
		Example: `  # Show recent log entries
  shelly log show

  # Follow log in real-time
  shelly log tail

  # Show log file path
  shelly log path

  # Clear log file
  shelly log clear`,
	}

	cmd.AddCommand(logshow.NewCommand(f))
	cmd.AddCommand(logtail.NewCommand(f))
	cmd.AddCommand(logpath.NewCommand(f))
	cmd.AddCommand(logclear.NewCommand(f))
	cmd.AddCommand(logexport.NewCommand(f))

	return cmd
}
