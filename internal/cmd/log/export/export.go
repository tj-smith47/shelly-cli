// Package export provides the log export subcommand.
package export

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Output  string
}

// NewCommand creates the log export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			return run(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	logPath, err := cmdutil.GetLogPath()
	if err != nil {
		return err
	}

	fs := config.Fs()
	data, err := afero.ReadFile(fs, logPath)
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

	if err := afero.WriteFile(fs, opts.Output, data, 0o600); err != nil {
		return err
	}

	ios.Success("Log exported to: %s", opts.Output)
	return nil
}
