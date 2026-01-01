// Package clearcmd provides the log clear subcommand.
package clearcmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the log clear command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "clear",
		Aliases: []string{"clean", "truncate"},
		Short:   "Clear log file",
		Long:    `Clear all entries from the CLI log file.`,
		Example: `  # Clear log file
  shelly log clear`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()
	fs := config.Fs()

	logPath, err := cmdutil.GetLogPath()
	if err != nil {
		return err
	}

	if _, err := fs.Stat(logPath); os.IsNotExist(err) {
		ios.Info("No log file to clear")
		return nil
	}

	// Open and truncate file
	file, err := fs.OpenFile(logPath, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	ios.Success("Log file cleared")
	return nil
}
