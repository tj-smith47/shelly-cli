// Package path provides the config path subcommand.
package path

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the config path command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	dir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	ios.Printf("%s/config.yaml\n", dir)
	return nil
}
