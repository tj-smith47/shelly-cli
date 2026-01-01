// Package show provides the log show subcommand.
package show

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Lines   int
}

// NewCommand creates the log show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Lines:   50,
	}

	cmd := &cobra.Command{
		Use:     "show",
		Aliases: []string{"view", "cat"},
		Short:   "Show recent log entries",
		Long:    `Show the most recent log entries from the CLI log file.`,
		Example: `  # Show last 50 lines (default)
  shelly log show

  # Show last 100 lines
  shelly log show -n 100`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Lines, "lines", "n", 50, "Number of lines to show")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	logPath, err := cmdutil.GetLogPath()
	if err != nil {
		return err
	}

	if _, err := config.Fs().Stat(logPath); err != nil {
		ios.Info("No log file found at: %s", logPath)
		ios.Info("Debug logging may not be enabled.")
		return nil
	}

	logLines, err := cmdutil.ReadLastLines(logPath, opts.Lines)
	if err != nil {
		return err
	}

	if len(logLines) == 0 {
		ios.Info("Log file is empty")
		return nil
	}

	for _, line := range logLines {
		ios.Println(line)
	}

	return nil
}
