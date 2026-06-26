// Package set provides the config set subcommand for CLI settings.
package set

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the config set command.
type Options struct {
	Factory *cmdutil.Factory

	Args []string
}

// NewCommand creates the config set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <key>=<value>...",
		Aliases: []string{"write", "update"},
		Short:   "Set CLI configuration values",
		Long: `Set configuration values in the Shelly CLI config file.

Use dot notation for nested values (e.g. "discovery.timeout=30s").

A key and its value may be separated with "=", ":", or a space — these are
equivalent:

  shelly config set telemetry=true
  shelly config set telemetry:true
  shelly config set telemetry true

Values are stored using each setting's real type (e.g. telemetry as a boolean,
ratelimit.global.max_concurrent as an integer), so booleans and numbers behave
correctly.`,
		Example: `  # Enable anonymous usage telemetry (these are equivalent)
  shelly config set telemetry=true
  shelly config set telemetry true

  # Set discovery timeout (duration)
  shelly config set discovery.timeout=30s

  # Set output format
  shelly config set output=json

  # Set multiple values
  shelly config set discovery.timeout=30s output=json`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completion.SettingKeysWithEquals(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	pairs, err := cmdutil.ParseKeyValues(opts.Args)
	if err != nil {
		return err
	}

	for _, kv := range pairs {
		value, err := config.CoerceSettingValue(kv.Key, kv.Value)
		if err != nil {
			return err
		}
		if err := config.SetSetting(kv.Key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", kv.Key, err)
		}

		ios.Success("Set %s = %s", kv.Key, config.FormatSettingValue(value))
	}

	return nil
}
