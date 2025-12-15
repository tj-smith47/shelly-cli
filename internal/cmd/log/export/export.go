// Package export provides the log export subcommand.
package export

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	Output string
}

// NewCommand creates the log export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "export",
		Aliases: []string{"save", "backup"},
		Short:   "Export log file",
		Long:    `Export the log file to a specified location.`,
		Example: `  # Export to file
  shelly log export -o /tmp/shelly-debug.log

  # Export to stdout
  shelly log export`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path")

	return cmd
}

func run(f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	logPath, err := cmdutil.GetLogPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(logPath) //nolint:gosec // Log file path is from config dir
	if err != nil {
		if os.IsNotExist(err) {
			ios.Info("No log file found")
			return nil
		}
		return err
	}

	if opts.Output == "" {
		ios.Printf("%s", string(data))
		return nil
	}

	if err := os.WriteFile(opts.Output, data, 0o600); err != nil {
		return err
	}

	ios.Success("Log exported to: %s", opts.Output)
	return nil
}
