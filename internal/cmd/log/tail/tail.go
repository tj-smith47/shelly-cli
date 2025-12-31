// Package tail provides the log tail subcommand.
package tail

import (
	"bufio"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Follow  bool
}

// NewCommand creates the log tail command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "tail",
		Aliases: []string{"follow", "f"},
		Short:   "Tail log file",
		Long:    `Show and optionally follow the log file in real-time.`,
		Example: `  # Show last entries and follow
  shelly log tail -f

  # Just show last entries
  shelly log tail`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Follow, "follow", "f", false, "Follow log output")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	logPath, err := cmdutil.GetLogPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		ios.Info("No log file found at: %s", logPath)
		return nil
	}

	// Show last 20 lines
	logLines, err := cmdutil.ReadLastLines(logPath, 20)
	if err != nil {
		return err
	}

	for _, line := range logLines {
		ios.Println(line)
	}

	if !opts.Follow {
		return nil
	}

	ios.Info("Following log... (Ctrl+C to stop)")

	file, err := os.Open(logPath) //nolint:gosec // Log file path is from config dir
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing log file", file)

	// Seek to end
	if _, err := file.Seek(0, 2); err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		ios.Printf("%s", line)
	}
}
