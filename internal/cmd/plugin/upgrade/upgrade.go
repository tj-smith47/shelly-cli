// Package upgrade provides the extension upgrade command.
package upgrade

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/pluginupgrade"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the extension upgrade command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:     "upgrade [name]",
		Aliases: []string{"update"},
		Short:   "Upgrade extension(s)",
		Long: `Upgrade installed extension(s) to the latest version.

If a name is provided, upgrades only that extension.
Use --all to upgrade all installed extensions.

Note: This command checks GitHub for newer releases. Extensions must have been
originally installed from GitHub (the repo info is stored in plugin metadata).
For extensions installed from local files or URLs, you need to reinstall manually.`,
		Example: `  # Upgrade a specific extension (requires GitHub source)
  shelly extension upgrade myext

  # Check and upgrade all extensions
  shelly extension upgrade --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return run(cmd.Context(), f, name, all)
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Upgrade all installed extensions")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, name string, all bool) error {
	ios := f.IOStreams()

	registry, err := plugins.NewRegistry()
	if err != nil {
		return err
	}

	upgrader := pluginupgrade.New(registry, ios)

	if all {
		extensionList, listErr := registry.List()
		if listErr != nil {
			return listErr
		}
		if len(extensionList) == 0 {
			ios.Info("No extensions installed")
			return nil
		}

		for _, ext := range extensionList {
			ios.Printf("Checking %s...\n", ext.Name)
		}

		results, upgradeErr := upgrader.UpgradeAll(ctx)
		if upgradeErr != nil {
			return upgradeErr
		}

		// Convert to term type for display
		termResults := make([]term.PluginUpgradeResult, len(results))
		for i, r := range results {
			termResults[i] = term.PluginUpgradeResult{
				Name:       r.Name,
				OldVersion: r.OldVersion,
				NewVersion: r.NewVersion,
				Upgraded:   r.Upgraded,
				Skipped:    r.Skipped,
				Error:      r.Error,
			}
		}
		term.DisplayPluginUpgradeResults(ios, termResults)
		return nil
	}

	if name == "" {
		ios.Info("Specify an extension name or use --all to upgrade all extensions")
		return nil
	}

	result, err := upgrader.UpgradeOne(ctx, name)
	if err != nil {
		return err
	}
	term.DisplayPluginUpgradeResult(ios, term.PluginUpgradeResult{
		Name:       result.Name,
		OldVersion: result.OldVersion,
		NewVersion: result.NewVersion,
		Upgraded:   result.Upgraded,
		Skipped:    result.Skipped,
		Error:      result.Error,
	})
	return nil
}
