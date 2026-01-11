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

Use dot notation for nested values (e.g., "defaults.timeout").
Multiple keys can be deleted at once.

If a key has nested child values, confirmation is required unless --yes is provided.`,
		Example: `  # Delete a single setting
  shelly config delete defaults.timeout

  # Delete multiple settings
  shelly config delete defaults.timeout defaults.output

  # Delete a parent key with all children (with confirmation)
  shelly config delete defaults

  # Skip confirmation prompt
  shelly config delete defaults --yes

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
		// Check if this is a parent key with nested values
		children, isParent := config.IsParentSetting(key)
		if isParent && !opts.Yes {
			ios.Warning("Key %q has %d nested values:", key, len(children))
			for _, child := range children {
				ios.Printf("  - %s\n", child)
			}

			confirmed, err := ios.Confirm(fmt.Sprintf("Delete %q and all nested values?", key), false)
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				ios.Info("Skipped %s", key)
				continue
			}

			// Delete all child keys first
			for _, child := range children {
				if err := config.DeleteSetting(child); err != nil {
					// Log but continue - some might not exist in file
					ios.DebugErr("delete child setting", err)
				}
			}
		}

		if err := config.DeleteSetting(key); err != nil {
			// If it's a parent-only key (no direct value), it might not exist in file
			if isParent && strings.Contains(err.Error(), "not set") {
				ios.Success("Deleted %s (and %d nested values)", key, len(children))
				continue
			}
			return fmt.Errorf("failed to delete %s: %w", key, err)
		}

		if isParent {
			ios.Success("Deleted %s (and %d nested values)", key, len(children))
		} else {
			ios.Success("Deleted %s", key)
		}
	}

	return nil
}
