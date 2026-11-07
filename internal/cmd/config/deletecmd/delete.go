// Package deletecmd provides the config delete subcommand for CLI settings.
package deletecmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the config delete command.
type Options struct {
	flags.ConfirmFlags
	Factory *cmdutil.Factory

	Keys []string
}

// NewCommand creates the config delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "delete <key>...",
		Aliases: []string{"del", "rm", "remove", "unset"},
		Short:   "Delete CLI configuration values",
		Long: `Delete configuration values from the Shelly CLI config file.

Use dot notation for nested values (e.g., "discovery.timeout").
Multiple keys can be deleted at once.

If a key has nested child values, confirmation is required unless --yes is provided.`,
		Example: `  # Delete a single setting
  shelly config delete discovery.timeout

  # Delete multiple settings
  shelly config delete discovery.timeout discovery.network

  # Delete a parent key with all children (with confirmation)
  shelly config delete discovery

  # Skip confirmation prompt
  shelly config delete discovery --yes

  # Using alias
  shelly config rm editor`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completion.SettingKeys(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Keys = args
			return run(opts)
		},
	}

	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	for _, key := range opts.Keys {
		children, isParent := config.IsParentSetting(key)
		if !isParent {
			if err := config.DeleteSetting(key); err != nil {
				return fmt.Errorf("failed to delete %s: %w", key, err)
			}
			ios.Success("Deleted %s", key)
			continue
		}

		if !opts.Yes {
			ios.Warning("Key %q has %d nested values:", key, len(children))
			for _, child := range children {
				ios.Printf("  - %s\n", child)
			}
		}
		confirmed, err := opts.Factory.ConfirmAction(fmt.Sprintf("Delete %q and all nested values?", key), opts.Yes)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}

		for _, child := range children {
			if err := config.DeleteSetting(child); err != nil {
				// A child listed by viper may come from defaults rather than the file.
				ios.DebugErr("delete child setting", err)
			}
		}
		// A parent key with no direct value of its own is reported as "not set".
		if err := config.DeleteSetting(key); err != nil && !strings.Contains(err.Error(), "not set") {
			return fmt.Errorf("failed to delete %s: %w", key, err)
		}
		ios.Success("Deleted %s (and %d nested values)", key, len(children))
	}

	return nil
}
