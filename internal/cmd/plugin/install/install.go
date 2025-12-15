// Package install provides the extension install command.
package install

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// NewCommand creates the extension install command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var force bool

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
			return run(cmd.Context(), f, args[0], force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall even if already installed")

	return cmd
}

// installResult holds the result of downloading/preparing a plugin for installation.
type installResult struct {
	localPath string
	source    plugins.Source
	version   string
	cleanup   func()
}

func run(ctx context.Context, f *cmdutil.Factory, source string, force bool) error {
	ios := f.IOStreams()

	registry, err := plugins.NewRegistry()
	if err != nil {
		return err
	}

	var result *installResult

	// Determine source type and get local path
	switch {
	case strings.HasPrefix(source, "gh:") || strings.HasPrefix(source, "github:"):
		// GitHub source
		var gerr error
		result, gerr = installFromGitHub(ctx, ios, source)
		if gerr != nil {
			return fmt.Errorf("failed to download from GitHub: %w", gerr)
		}
		if result.cleanup != nil {
			defer result.cleanup()
		}

	case strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://"):
		// URL source
		var derr error
		result, derr = downloadFromURL(ctx, ios, source)
		if derr != nil {
			return fmt.Errorf("failed to download extension: %w", derr)
		}
		if result.cleanup != nil {
			defer result.cleanup()
		}

	default:
		// Local file
		result = &installResult{
			localPath: source,
			source:    plugins.ParseLocalSource(source),
		}
	}

	// Get extension name from filename
	filename := filepath.Base(result.localPath)
	if !strings.HasPrefix(filename, plugins.PluginPrefix) {
		return fmt.Errorf("extension must be named with prefix %q (got %q)", plugins.PluginPrefix, filename)
	}

	extName := strings.TrimPrefix(filename, plugins.PluginPrefix)

	// Check if already installed
	if registry.IsInstalled(extName) && !force {
		return fmt.Errorf("extension %q is already installed (use --force to reinstall)", extName)
	}

	// Create manifest with source information
	manifest := plugins.NewManifest(extName, result.source)
	if result.version != "" {
		manifest.Version = result.version
	}

	// Install with manifest
	if err := registry.InstallWithManifest(result.localPath, manifest); err != nil {
		return err
	}

	ios.Success("Installed extension '%s'", extName)
	return nil
}

// installFromGitHub downloads an extension from a GitHub repository.
func installFromGitHub(ctx context.Context, ios *iostreams.IOStreams, source string) (*installResult, error) {
	owner, repo, err := github.ParseRepoString(source)
	if err != nil {
		return nil, err
	}

	// The binary name is the repo name (should be shelly-something)
	binaryName := repo
	if !strings.HasPrefix(binaryName, plugins.PluginPrefix) {
		binaryName = plugins.PluginPrefix + repo
	}

	ios.StartProgress(fmt.Sprintf("Fetching latest release from %s/%s...", owner, repo))

	client := github.NewClient(ios)

	release, err := client.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		ios.StopProgress()
		return nil, err
	}

	ios.StopProgress()
	ios.Info("Found release: %s", release.TagName)

	// Find the appropriate binary for this platform
	asset, err := client.FindBinaryAsset(release, binaryName)
	if err != nil {
		return nil, err
	}

	ios.StartProgress(fmt.Sprintf("Downloading %s...", asset.Name))

	binaryPath, cleanup, err := client.DownloadAndExtract(ctx, asset, binaryName)
	if err != nil {
		ios.StopProgress()
		return nil, err
	}

	ios.StopProgress()

	// Build source info with repo string (e.g., "owner/repo")
	repoStr := owner + "/" + repo
	return &installResult{
		localPath: binaryPath,
		source:    plugins.ParseGitHubSource(repoStr, release.TagName, asset.Name),
		version:   strings.TrimPrefix(release.TagName, "v"),
		cleanup:   cleanup,
	}, nil
}

// downloadFromURL downloads an extension from a URL.
func downloadFromURL(ctx context.Context, ios *iostreams.IOStreams, downloadURL string) (*installResult, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "shelly-ext-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		if rerr := os.Remove(tmpPath); rerr != nil {
			ios.DebugErr("removing temp file", rerr)
		}
	}

	ios.StartProgress("Downloading extension...")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		cleanup()
		ios.StopProgress()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "shelly-cli")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cleanup()
		ios.StopProgress()
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			ios.DebugErr("closing response body", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		cleanup()
		ios.StopProgress()
		return nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if cerr := tmpFile.Close(); cerr != nil && err == nil {
		err = cerr
	}

	if err != nil {
		cleanup()
		ios.StopProgress()
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Make executable - extensions must be executable binaries
	if err := os.Chmod(tmpPath, 0o755); err != nil { //nolint:gosec // G302: extensions require executable permissions
		cleanup()
		ios.StopProgress()
		return nil, fmt.Errorf("failed to make executable: %w", err)
	}

	ios.StopProgress()

	return &installResult{
		localPath: tmpPath,
		source:    plugins.ParseURLSource(downloadURL),
		cleanup:   cleanup,
	}, nil
}
