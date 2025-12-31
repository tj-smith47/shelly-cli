// Package set provides the config set subcommand for CLI settings.
package set

import (
	"fmt"
	"strings"

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

Use dot notation for nested values (e.g., "defaults.timeout=30s").`,
		Example: `  # Set default timeout
  shelly config set defaults.timeout=30s

  # Set output format
  shelly config set defaults.output=json

  # Set multiple values
  shelly config set defaults.timeout=30s defaults.output=json`,
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

	for _, arg := range opts.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format %q, expected key=value", arg)
		}

		key, value := parts[0], parts[1]
		if err := config.SetSetting(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}

		ios.Success("Set %s = %s", key, value)
	}

	return nil
}
