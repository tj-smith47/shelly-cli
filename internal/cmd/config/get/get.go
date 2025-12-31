// Package get provides the config get subcommand for CLI settings.
package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the options for the config get command.
type Options struct {
	Factory *cmdutil.Factory

	Key string
}

// NewCommand creates the config get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "get [key]",
		Aliases: []string{"read"},
		Short:   "Get a CLI configuration value",
		Long: `Get a configuration value from the Shelly CLI config file.

Use dot notation to access nested values (e.g., "defaults.timeout").
Without a key, shows all configuration values.`,
		Example: `  # Get all settings
  shelly config get

  # Get default timeout
  shelly config get defaults.timeout

  # Output as JSON
  shelly config get -o json`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.SettingKeys(),
		RunE: func(cmd *cobra.Command, args []string) error {
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

	if opts.Key == "" {
		settings := config.GetAllSettings()
		if output.WantsStructured() {
			return output.FormatOutput(ios.Out, settings)
		}
		return term.DisplayConfigTable(ios, settings)
	}

	value, ok := config.GetSetting(opts.Key)
	if !ok {
		return fmt.Errorf("configuration key %q not set", opts.Key)
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, map[string]any{opts.Key: value})
	}

	ios.Printf("%s: %s\n", opts.Key, config.FormatSettingValue(value))
	return nil
}
