// Package path provides the log path subcommand.
package path

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the log path command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "path",
		Aliases: []string{"where", "location"},
		Short:   "Show log file path",
		Long:    `Show the path to the CLI log file.`,
		Example: `  # Show log path
  shelly log path`,
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

	ios.Println(logPath)

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		ios.Info("(file does not exist yet)")
	} else {
		info, err := os.Stat(logPath)
		if err == nil {
			ios.Info("Size: %d bytes", info.Size())
			ios.Info("Modified: %s", info.ModTime().Format(time.RFC3339))
		}
	}

	return nil
}
