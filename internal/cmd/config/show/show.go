// Package show provides the config show subcommand for CLI settings.
package show

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the options for the config show command.
type Options struct {
	Factory *cmdutil.Factory
	Key     string
}

// NewCommand creates the config show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "show [key]",
		Aliases: []string{"view", "dump"},
		Short:   "Display current CLI configuration",
		Long: `Display CLI configuration settings.

Without arguments, shows all configuration.
With a key argument, shows only that specific setting (supports dot notation).`,
		Example: `  # Show all configuration
  shelly config show

  # Show specific setting
  shelly config show defaults.timeout

  # Show a section
  shelly config show defaults

  # Output as JSON
  shelly config show -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Key = args[0]
			}
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	var settings any
	if opts.Key != "" {
		val, ok := config.GetSetting(opts.Key)
		if !ok {
			return fmt.Errorf("setting %q not found", opts.Key)
		}
		// Wrap in map for consistent table display
		settings = map[string]any{opts.Key: val}
	} else {
		settings = config.GetAllSettings()
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, settings)
	}

	return term.DisplayConfigTable(ios, settings)
}
