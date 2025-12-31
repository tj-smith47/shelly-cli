// Package show provides the config show subcommand for CLI settings.
package show

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the options for the config show command.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the config show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "show",
		Aliases: []string{"view", "dump"},
		Short:   "Display current CLI configuration",
		Long:    `Display the complete Shelly CLI configuration file contents.`,
		Example: `  # Show all configuration
  shelly config show

  # Output as JSON
  shelly config show -o json`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()
	settings := config.GetAllSettings()

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, settings)
	}

	return term.DisplayConfigTable(ios, settings)
}
