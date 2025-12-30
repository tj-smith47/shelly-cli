package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	
)

// Result contains the result of an upgrade attempt.
type Result struct {
	Name       string
	OldVersion string
	NewVersion string
	Upgraded   bool
	Skipped    bool
	Error      error
}

// Upgrader handles plugin upgrade operations.
type Upgrader struct {
	registry *Registry
	loader   *Loader
	ios      *iostreams.IOStreams
}

// New creates a new plugin upgrader.
func New(registry *Registry, ios *iostreams.IOStreams) *Upgrader {
	return &Upgrader{
		registry: registry,
		loader:   NewLoader(),
		ios:      ios,
	}
}

// UpgradeAll upgrades all installed 
func (u *Upgrader) UpgradeAll(ctx context.Context) ([]Result, error) {
	extensionList, err := u.registry.List()
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(extensionList))
	for _, ext := range extensionList {
		result := u.upgradeExtension(ctx, ext.Name, ext.Version, ext.Manifest)
		results = append(results, result)
	}

	return results, nil
}

// UpgradeOne upgrades a single plugin by name.
func (u *Upgrader) UpgradeOne(ctx context.Context, name string) (Result, error) {
	if !u.registry.IsInstalled(name) {
		return Result{}, fmt.Errorf("extension %q is not installed", name)
	}

	plugin, err := u.loader.Find(name)
	if err != nil {
		return Result{}, fmt.Errorf("failed to find plugin: %w", err)
	}
	if plugin == nil {
		return Result{}, fmt.Errorf("plugin %q not found", name)
	}

	return u.upgradeExtension(ctx, name, plugin.Version, plugin.Manifest), nil
}

func (u *Upgrader) upgradeExtension(ctx context.Context, name, currentVersion string, manifest *Manifest) Result {
	result := Result{
		Name:       name,
		OldVersion: currentVersion,
	}

	binaryName := PluginPrefix + name

	// If we have a manifest with GitHub source, use it directly
	if manifest != nil && manifest.Source.Type == SourceTypeGitHub && manifest.Source.URL != "" {
		return u.upgradeFromManifest(ctx, name, binaryName, currentVersion, manifest)
	}

	// If manifest exists but isn't GitHub, skip
	if manifest != nil {
		result.Skipped = true
		result.Error = fmt.Errorf("cannot auto-upgrade: %s", manifest.UpgradeMessage())
		return result
	}

	// No manifest - fall back to heuristic search (legacy plugins)
	return u.upgradeWithHeuristics(ctx, name, binaryName, currentVersion)
}

func (u *Upgrader) upgradeFromManifest(ctx context.Context, name, binaryName, currentVersion string, manifest *Manifest) Result {
	result := Result{
		Name:       name,
		OldVersion: currentVersion,
	}

	// Parse owner/repo from GitHub URL
	owner, repo, err := parseGitHubURL(manifest.Source.URL)
	if err != nil {
		result.Error = err
		return result
	}

	client := github.NewClient(u.ios)
	release, err := client.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		result.Error = fmt.Errorf("failed to get latest release: %w", err)
		return result
	}

	// Compare versions
	latestVersion := release.TagName
	latestClean := strings.TrimPrefix(latestVersion, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")

	if latestClean == currentClean {
		result.NewVersion = currentVersion
		return result // Already up to date
	}

	result.NewVersion = latestVersion

	// Find and download the binary
	asset, err := client.FindBinaryAsset(release, binaryName)
	if err != nil {
		result.Error = fmt.Errorf("failed to find binary: %w", err)
		return result
	}

	binaryPath, cleanup, err := client.DownloadAndExtract(ctx, asset, binaryName)
	if err != nil {
		result.Error = fmt.Errorf("failed to download: %w", err)
		return result
	}
	defer cleanup()

	// Create updated manifest with new version info
	newManifest := NewManifest(name, ParseGitHubSource(owner+"/"+repo, release.TagName, asset.Name))
	newManifest.Version = latestClean
	newManifest.Description = manifest.Description
	newManifest.InstalledAt = manifest.InstalledAt // Preserve original install time
	newManifest.MarkUpdated()

	// Install with updated manifest
	if err := u.registry.InstallWithManifest(binaryPath, newManifest); err != nil {
		result.Error = fmt.Errorf("failed to install: %w", err)
		return result
	}

	result.Upgraded = true
	return result
}

func (u *Upgrader) upgradeWithHeuristics(ctx context.Context, name, binaryName, currentVersion string) Result {
	result := Result{
		Name:       name,
		OldVersion: currentVersion,
	}

	// Check common GitHub sources - heuristic approach for legacy plugins without manifests
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

		client := github.NewClient(u.ios)
		release, err := client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			lastErr = err
			continue
		}

		// Compare versions
		latestVersion := release.TagName
		latestClean := strings.TrimPrefix(latestVersion, "v")
		currentClean := strings.TrimPrefix(currentVersion, "v")

		if latestClean == currentClean {
			result.NewVersion = currentVersion
			return result // Already up to date
		}

		result.NewVersion = latestVersion

		// Find and download the binary
		asset, err := client.FindBinaryAsset(release, binaryName)
		if err != nil {
			result.Error = fmt.Errorf("failed to find binary: %w", err)
			return result
		}

		binaryPath, cleanup, err := client.DownloadAndExtract(ctx, asset, binaryName)
		if err != nil {
			result.Error = fmt.Errorf("failed to download: %w", err)
			return result
		}

		// Create manifest for previously manifest-less plugin
		newManifest := NewManifest(name, ParseGitHubSource(owner+"/"+repo, release.TagName, asset.Name))
		newManifest.Version = latestClean

		// Install with new manifest
		if err := u.registry.InstallWithManifest(binaryPath, newManifest); err != nil {
			cleanup()
			result.Error = fmt.Errorf("failed to install: %w", err)
			return result
		}
		cleanup()

		result.Upgraded = true
		return result
	}

	if lastErr != nil {
		result.Skipped = true
		result.Error = fmt.Errorf("no GitHub source found for %s: %w", name, lastErr)
		return result
	}

	result.Skipped = true
	result.Error = fmt.Errorf("no GitHub source found for %s", name)
	return result
}

// parseGitHubURL extracts owner and repo from a GitHub URL.
func parseGitHubURL(url string) (owner, repo string, err error) {
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	parts := strings.SplitN(url, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
	}
	return parts[0], parts[1], nil
}
