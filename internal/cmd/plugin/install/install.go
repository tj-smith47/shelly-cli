// Package install provides the extension install command.
package install

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/download"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Force   bool
	Source  string
}

// NewCommand creates the extension install command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "install <source>",
		Aliases: []string{"add"},
		Short:   "Install an extension",
		Long: `Install an extension from a local file, URL, or GitHub repository.

Supported sources:
  - Local file: ./path/to/shelly-myext
  - GitHub repo: gh:user/shelly-myext
  - HTTP URL: https://example.com/shelly-myext

The extension must be named with the shelly- prefix.`,
		Example: `  # Install from local file
  shelly extension install ./shelly-myext

  # Install from GitHub (downloads latest release binary)
  shelly extension install gh:user/shelly-myext

  # Install from URL
  shelly extension install https://example.com/shelly-myext

  # Force reinstall
  shelly extension install ./shelly-myext --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Source = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force reinstall even if already installed")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	registry, err := plugins.NewRegistry()
	if err != nil {
		return err
	}

	var (
		localPath string
		pluginSrc plugins.Source
		version   string
		cleanup   func()
	)

	// Determine source type and get local path
	switch {
	case strings.HasPrefix(opts.Source, "gh:") || strings.HasPrefix(opts.Source, "github:"):
		owner, repo, parseErr := github.ParseRepoString(opts.Source)
		if parseErr != nil {
			return fmt.Errorf("failed to parse GitHub source: %w", parseErr)
		}

		client := github.NewClient(ios)
		var result *github.ExtensionDownloadResult
		downloadErr := cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Downloading from %s/%s...", owner, repo), func(ctx context.Context) error {
			var derr error
			result, derr = client.DownloadExtensionRelease(ctx, owner, repo, plugins.PluginPrefix)
			return derr
		})
		if downloadErr != nil {
			return fmt.Errorf("failed to download from GitHub: %w", downloadErr)
		}

		localPath = result.LocalPath
		cleanup = result.Cleanup
		pluginSrc = plugins.ParseGitHubSource(owner+"/"+repo, result.TagName, result.AssetName)
		version = strings.TrimPrefix(result.TagName, "v")

	case strings.HasPrefix(opts.Source, "http://") || strings.HasPrefix(opts.Source, "https://"):
		var result *download.Result
		downloadErr := cmdutil.RunWithSpinner(ctx, ios, "Downloading extension...", func(ctx context.Context) error {
			var derr error
			result, derr = download.FromURL(ctx, opts.Source)
			return derr
		})
		if downloadErr != nil {
			return fmt.Errorf("failed to download extension: %w", downloadErr)
		}
		localPath = result.LocalPath
		cleanup = result.Cleanup
		pluginSrc = plugins.ParseURLSource(opts.Source)

	default:
		localPath = opts.Source
		pluginSrc = plugins.ParseLocalSource(opts.Source)
	}

	if cleanup != nil {
		defer cleanup()
	}

	// Get extension name from filename
	filename := filepath.Base(localPath)
	if !strings.HasPrefix(filename, plugins.PluginPrefix) {
		return fmt.Errorf("extension must be named with prefix %q (got %q)", plugins.PluginPrefix, filename)
	}

	extName := strings.TrimPrefix(filename, plugins.PluginPrefix)

	// Check if already installed
	if registry.IsInstalled(extName) && !opts.Force {
		return fmt.Errorf("extension %q is already installed (use --force to reinstall)", extName)
	}

	// Create manifest with source information
	manifest := plugins.NewManifest(extName, pluginSrc)
	if version != "" {
		manifest.Version = version
	}

	// Install with manifest
	if err := registry.InstallWithManifest(localPath, manifest); err != nil {
		return err
	}

	ios.Success("Installed extension '%s'", extName)
	return nil
}
