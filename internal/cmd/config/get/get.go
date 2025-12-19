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

// NewCommand creates the config get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(f, args)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, args []string) error {
	ios := f.IOStreams()

	if len(args) == 0 {
		settings := config.GetAllSettings()
		if output.WantsStructured() {
			return output.FormatOutput(ios.Out, settings)
		}
		return term.DisplayConfigTable(ios, settings)
	}

	key := args[0]
	value, ok := config.GetSetting(key)
	if !ok {
		return fmt.Errorf("configuration key %q not set", key)
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, map[string]any{key: value})
	}

	ios.Printf("%s: %s\n", key, config.FormatSettingValue(value))
	return nil
}
