// Package path provides the config path subcommand.
package path

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the config path command.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the config path command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "path",
		Aliases: []string{"dir", "location"},
		Short:   "Show configuration file path",
		Long:    `Display the path to the Shelly CLI configuration file.`,
		Example: `  # Show config file path
  shelly config path

  # Open config directory in file manager
  open $(shelly config path | xargs dirname)`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	dir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	ios.Printf("%s/config.yaml\n", dir)
	return nil
}
