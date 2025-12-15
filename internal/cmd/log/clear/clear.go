// Package clearcmd provides the log clear subcommand.
package clearcmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the log clear command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clear",
		Aliases: []string{"clean", "truncate"},
		Short:   "Clear log file",
		Long:    `Clear all entries from the CLI log file.`,
		Example: `  # Clear log file
  shelly log clear`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	logPath, err := cmdutil.GetLogPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		ios.Info("No log file to clear")
		return nil
	}

	if err := os.Truncate(logPath, 0); err != nil {
		return err
	}

	ios.Success("Log file cleared")
	return nil
}
