// Package upgrade provides the extension upgrade command.
package upgrade

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
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

	if all {
		return upgradeAll(ctx, ios, registry)
	}

	if name == "" {
		ios.Info("Specify an extension name or use --all to upgrade all extensions")
		return nil
	}

	return upgradeOne(ctx, ios, registry, name)
}

func upgradeAll(ctx context.Context, ios *iostreams.IOStreams, registry *plugins.Registry) error {
	extensionList, err := registry.List()
	if err != nil {
		return err
	}

	if len(extensionList) == 0 {
		ios.Info("No extensions installed")
		return nil
	}

	var upgraded, skipped, failed int

	for _, ext := range extensionList {
		ios.Printf("Checking %s...\n", ext.Name)

		err := upgradeExtension(ctx, ios, registry, ext.Name, ext.Version)
		if err != nil {
			if strings.Contains(err.Error(), "no GitHub source") {
				ios.Warning("  Skipped: %v", err)
				skipped++
			} else {
				ios.Error("  Failed: %v", err)
				failed++
			}
			continue
		}
		upgraded++
	}

	ios.Printf("\nUpgrade complete: %d upgraded, %d skipped, %d failed\n", upgraded, skipped, failed)

	return nil
}

func upgradeOne(ctx context.Context, ios *iostreams.IOStreams, registry *plugins.Registry, name string) error {
	// Check if extension is installed
	if !registry.IsInstalled(name) {
		return fmt.Errorf("extension %q is not installed", name)
	}

	// Get current version
	loader := plugins.NewLoader()
	plugin, err := loader.Find(name)
	if err != nil {
		return fmt.Errorf("failed to find plugin: %w", err)
	}
	if plugin == nil {
		return fmt.Errorf("plugin %q not found", name)
	}

	return upgradeExtension(ctx, ios, registry, name, plugin.Version)
}

func upgradeExtension(ctx context.Context, ios *iostreams.IOStreams, registry *plugins.Registry, name, currentVersion string) error {
	// Try to find GitHub source from plugin name
	// Convention: plugins should be named shelly-<name> and come from a repo of the same name
	binaryName := plugins.PluginPrefix + name

	// Check common GitHub sources - this is a heuristic since we don't store origin
	// In the future, we could store origin in a manifest file
	possibleSources := []string{
		fmt.Sprintf("gh:tj-smith47/%s", binaryName), // Official plugins
		fmt.Sprintf("gh:%s/%s", name, binaryName),   // User repos
	}

	var lastErr error
	for _, source := range possibleSources {
		owner, repo, err := github.ParseRepoString(source)
		if err != nil {
			continue
		}

		client := github.NewClient(ios)
		release, err := client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			lastErr = err
			continue
		}

		// Compare versions
		latestVersion := release.TagName
		// Strip 'v' prefix for comparison
		latestClean := strings.TrimPrefix(latestVersion, "v")
		currentClean := strings.TrimPrefix(currentVersion, "v")

		if latestClean == currentClean {
			ios.Info("  Already at latest version: %s", currentVersion)
			return nil
		}

		ios.Info("  Upgrading from %s to %s", currentVersion, latestVersion)

		// Find and download the binary
		asset, err := client.FindBinaryAsset(release, binaryName)
		if err != nil {
			return fmt.Errorf("failed to find binary: %w", err)
		}

		binaryPath, cleanup, err := client.DownloadAndExtract(ctx, asset, binaryName)
		if err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}

		// Install (will overwrite existing)
		installErr := registry.Install(binaryPath)
		cleanup() // Clean up temp files immediately after use
		if installErr != nil {
			return fmt.Errorf("failed to install: %w", installErr)
		}

		ios.Success("  Upgraded %s to %s", name, latestVersion)
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("no GitHub source found for %s: %w", name, lastErr)
	}

	return fmt.Errorf("no GitHub source found for %s", name)
}
