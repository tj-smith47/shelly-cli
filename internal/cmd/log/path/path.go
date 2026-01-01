// Package path provides the log path subcommand.
package path

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the log path command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "path",
		Aliases: []string{"where", "location"},
		Short:   "Show log file path",
		Long:    `Show the path to the CLI log file.`,
		Example: `  # Show log path
  shelly log path`,
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

	ios.Println(logPath)

	info, err := fs.Stat(logPath)
	if os.IsNotExist(err) {
		ios.Info("(file does not exist yet)")
	} else if err == nil {
		ios.Info("Size: %d bytes", info.Size())
		ios.Info("Modified: %s", info.ModTime().Format(time.RFC3339))
	}

	return nil
}
