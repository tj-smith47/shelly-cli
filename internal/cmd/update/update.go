// Package update provides the update command for self-updating the CLI.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

type options struct {
	Factory    *cmdutil.Factory
	Check      bool
	Version    string
	Channel    string
	Rollback   bool
	Yes        bool
	IncludePre bool
}

// NewCommand creates the update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &options{Factory: f}

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade", "u"},
		Short:   "Update shelly to the latest version",
		Long: `Update shelly to the latest version.

By default, downloads and installs the latest stable release from GitHub.
Use --check to only check for updates without installing.
Use --version to install a specific version.`,
		Example: `  # Check for updates
  shelly update --check

  # Update to latest version
  shelly update

  # Update to a specific version
  shelly update --version v1.2.0

  # Update with pre-releases
  shelly update --include-pre

  # Update without confirmation
  shelly update --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Check, "check", "c", false, "Check for updates without installing")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Install a specific version")
	cmd.Flags().StringVar(&opts.Channel, "channel", "stable", "Release channel (stable, beta)")
	cmd.Flags().BoolVar(&opts.Rollback, "rollback", false, "Rollback to previous version")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.IncludePre, "include-pre", false, "Include pre-release versions")

	return cmd
}

func run(ctx context.Context, opts *options) error {
	ios := opts.Factory.IOStreams()
	ghClient := github.NewClient(ios)

	currentVersion := version.Version
	if err := validateCurrentVersion(ios, currentVersion, opts.Check); err != nil {
		return err
	}

	// Handle rollback
	if opts.Rollback {
		return handleRollback(ctx, ghClient, currentVersion, opts)
	}

	// Get target release
	targetRelease, err := getTargetRelease(ctx, ios, ghClient, opts)
	if err != nil {
		if errors.Is(err, errNoReleasesFound) {
			return nil // Not an error condition, message already printed
		}
		return err
	}

	availableVersion := targetRelease.Version()

	// Compare versions
	if opts.Check {
		return showUpdateStatus(ios, currentVersion, availableVersion, targetRelease)
	}

	// Check if update is needed
	if !github.IsNewerVersion(currentVersion, availableVersion) && opts.Version == "" {
		ios.Printf("Already at latest version (%s)\n", currentVersion)
		return nil
	}

	// Show update info and confirm
	if err := confirmUpdate(opts, currentVersion, availableVersion, targetRelease); err != nil {
		return err
	}

	// Perform update
	return performUpdate(ctx, ios, ghClient, targetRelease)
}

func validateCurrentVersion(ios *iostreams.IOStreams, currentVersion string, checkOnly bool) error {
	if currentVersion == "" || currentVersion == "dev" {
		ios.Warning("Development version detected, cannot determine update status")
		if !checkOnly {
			return fmt.Errorf("cannot update development version")
		}
	}
	return nil
}

func getTargetRelease(ctx context.Context, ios *iostreams.IOStreams, ghClient *github.Client, opts *options) (*github.Release, error) {
	if opts.Version != "" {
		return fetchSpecificVersion(ctx, ghClient, opts.Version)
	}
	return fetchLatestVersion(ctx, ios, ghClient, opts.IncludePre || opts.Channel == "beta")
}

func fetchSpecificVersion(ctx context.Context, ghClient *github.Client, ver string) (*github.Release, error) {
	tag := ver
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}
	release, err := ghClient.GetReleaseByTag(ctx, github.DefaultOwner, github.DefaultRepo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version %s: %w", ver, err)
	}
	return release, nil
}

// errNoReleasesFound is a sentinel error for when no releases are found (but it's not an error condition).
var errNoReleasesFound = errors.New("no releases found")

func fetchLatestVersion(ctx context.Context, ios *iostreams.IOStreams, ghClient *github.Client, includePre bool) (*github.Release, error) {
	release, err := getLatestRelease(ctx, ghClient, includePre)
	if err != nil {
		if errors.Is(err, github.ErrNoReleases) {
			ios.Info("No releases found")
			return nil, errNoReleasesFound
		}
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	return release, nil
}

func confirmUpdate(opts *options, currentVersion, availableVersion string, release *github.Release) error {
	ios := opts.Factory.IOStreams()

	ios.Printf("\nCurrent version: %s\n", currentVersion)
	ios.Printf("Available version: %s\n", availableVersion)

	if release.Body != "" {
		ios.Printf("\nRelease notes:\n%s\n", formatReleaseNotes(release.Body))
	}

	confirm, err := opts.Factory.ConfirmAction("\nProceed with update?", opts.Yes)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	if !confirm {
		ios.Info("Update cancelled")
		return fmt.Errorf("update cancelled by user")
	}
	return nil
}

func getLatestRelease(ctx context.Context, ghClient *github.Client, includePre bool) (*github.Release, error) {
	if !includePre {
		return ghClient.GetLatestRelease(ctx, github.DefaultOwner, github.DefaultRepo)
	}

	// Get all releases and find the latest (including pre)
	releases, err := ghClient.ListReleases(ctx, github.DefaultOwner, github.DefaultRepo, true)
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, github.ErrNoReleases
	}

	github.SortReleasesByVersion(releases)
	return &releases[0], nil
}

func showUpdateStatus(ios *iostreams.IOStreams, currentVersion, availableVersion string, release *github.Release) error {
	if github.IsNewerVersion(currentVersion, availableVersion) {
		ios.Printf("Update available: %s -> %s\n", currentVersion, availableVersion)
		ios.Printf("  Run 'shelly update' to install\n")
		if release.HTMLURL != "" {
			ios.Printf("  Release page: %s\n", release.HTMLURL)
		}
	} else {
		ios.Printf("Already at latest version (%s)\n", currentVersion)
	}
	return nil
}

func handleRollback(ctx context.Context, ghClient *github.Client, currentVersion string, opts *options) error {
	ios := opts.Factory.IOStreams()

	// Get all releases to find the previous version
	releases, err := ghClient.ListReleases(ctx, github.DefaultOwner, github.DefaultRepo, opts.IncludePre)
	if err != nil {
		return fmt.Errorf("failed to list releases: %w", err)
	}

	if len(releases) < 2 {
		return fmt.Errorf("no previous version available for rollback")
	}

	github.SortReleasesByVersion(releases)

	// Find current version index and get previous
	var previousRelease *github.Release
	for i, r := range releases {
		if r.Version() == strings.TrimPrefix(currentVersion, "v") && i+1 < len(releases) {
			previousRelease = &releases[i+1]
			break
		}
	}

	if previousRelease == nil {
		// Fallback to second release
		previousRelease = &releases[1]
	}

	ios.Printf("Rolling back from %s to %s\n", currentVersion, previousRelease.Version())

	confirm, err := opts.Factory.ConfirmAction("Proceed with rollback?", opts.Yes)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	if !confirm {
		ios.Info("Rollback cancelled")
		return nil
	}

	return performUpdate(ctx, ios, ghClient, previousRelease)
}

func performUpdate(ctx context.Context, ios *iostreams.IOStreams, ghClient *github.Client, release *github.Release) error {
	ios.StartProgress(fmt.Sprintf("Downloading shelly %s...", release.Version()))

	// Find the appropriate asset for this platform
	asset := release.FindAssetForPlatform()
	if asset == nil {
		ios.StopProgress()
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download and extract
	binaryPath, cleanup, err := ghClient.DownloadAndExtract(ctx, asset, "shelly")
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to download: %w", err)
	}
	defer cleanup()

	ios.StopProgress()

	// Verify checksum if available
	checksumAsset := release.FindChecksumAsset(asset.Name)
	if checksumAsset != nil {
		ios.StartProgress("Verifying checksum...")
		if err := verifyChecksum(ctx, ios, ghClient, binaryPath, asset.Name, checksumAsset); err != nil {
			ios.StopProgress()
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		ios.StopProgress()
		ios.Success("Checksum verified")
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Replace the binary
	ios.StartProgress("Installing update...")
	if err := replaceBinary(ios, binaryPath, execPath); err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to install update: %w", err)
	}
	ios.StopProgress()

	ios.Success("Successfully updated to shelly %s", release.Version())
	ios.Info("Restart shelly to use the new version")

	return nil
}

func verifyChecksum(ctx context.Context, ios *iostreams.IOStreams, ghClient *github.Client, binaryPath, assetName string, checksumAsset *github.Asset) error {
	// Calculate checksum of downloaded file
	f, err := os.Open(binaryPath) //nolint:gosec // G304: binaryPath is from controlled temp directory
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			ios.DebugErr("closing binary file", cerr)
		}
	}()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return fmt.Errorf("calculate hash: %w", err)
	}
	actualHash := hex.EncodeToString(hasher.Sum(nil))

	// Download and parse checksum file
	tmpDir, err := os.MkdirTemp("", "shelly-checksum-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if rerr := os.RemoveAll(tmpDir); rerr != nil {
			ios.DebugErr("removing temp dir", rerr)
		}
	}()

	checksumPath := filepath.Join(tmpDir, checksumAsset.Name)
	if err := ghClient.DownloadAsset(ctx, checksumAsset, checksumPath); err != nil {
		return fmt.Errorf("download checksum: %w", err)
	}

	content, err := os.ReadFile(checksumPath) //nolint:gosec // G304: checksumPath is from controlled temp directory
	if err != nil {
		return fmt.Errorf("read checksum: %w", err)
	}

	expectedHash, err := parseChecksumFile(string(content), assetName)
	if err != nil {
		return err
	}

	if !strings.EqualFold(actualHash, expectedHash) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func parseChecksumFile(content, assetName string) (string, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: "hash  filename" or "hash filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			filename := parts[len(parts)-1]

			// Handle "*filename" format (binary mode indicator)
			filename = strings.TrimPrefix(filename, "*")

			if strings.EqualFold(filepath.Base(filename), assetName) {
				return hash, nil
			}
		} else if len(parts) == 1 {
			// Single hash (assume it's for this file)
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum not found for %s in checksum file", assetName)
}

func replaceBinary(ios *iostreams.IOStreams, newPath, targetPath string) error {
	// Read the new binary
	newBinary, err := os.ReadFile(newPath) //nolint:gosec // G304: newPath is from controlled temp directory
	if err != nil {
		return fmt.Errorf("read new binary: %w", err)
	}

	// Get permissions of the old binary
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("stat target: %w", err)
	}
	mode := info.Mode()

	// Create backup
	backupPath := targetPath + ".bak"
	if err := createBackup(ios, targetPath, backupPath); err != nil {
		return err
	}

	// Write new binary
	if err := os.WriteFile(targetPath, newBinary, mode); err != nil {
		return restoreFromBackup(backupPath, targetPath, err)
	}

	// Remove backup
	if rerr := os.Remove(backupPath); rerr != nil {
		ios.DebugErr("removing backup", rerr)
	}

	return nil
}

func createBackup(ios *iostreams.IOStreams, targetPath, backupPath string) error {
	if err := os.Rename(targetPath, backupPath); err != nil {
		// On Windows, we might need to copy instead
		if runtime.GOOS != "windows" {
			return fmt.Errorf("backup failed: %w", err)
		}

		if copyErr := copyFile(ios, targetPath, backupPath); copyErr != nil {
			return fmt.Errorf("backup failed: %w", copyErr)
		}
		// Try to remove original (might fail if in use)
		if rerr := os.Remove(targetPath); rerr != nil {
			ios.DebugErr("removing original binary", rerr)
		}
	}
	return nil
}

func restoreFromBackup(backupPath, targetPath string, writeErr error) error {
	if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
		return fmt.Errorf("write failed (%w) and restore failed: %w", writeErr, restoreErr)
	}
	return fmt.Errorf("write failed: %w", writeErr)
}

func copyFile(ios *iostreams.IOStreams, src, dst string) error {
	source, err := os.Open(src) //nolint:gosec // G304: src is the current executable path
	if err != nil {
		return err
	}
	defer func() {
		if cerr := source.Close(); cerr != nil {
			ios.DebugErr("closing source file", cerr)
		}
	}()

	destination, err := os.Create(dst) //nolint:gosec // G304: dst is backup path derived from executable
	if err != nil {
		return err
	}
	defer func() {
		if cerr := destination.Close(); cerr != nil {
			ios.DebugErr("closing destination file", cerr)
		}
	}()

	_, err = io.Copy(destination, source)
	return err
}

func formatReleaseNotes(body string) string {
	// Truncate if too long
	const maxLen = 500
	if len(body) > maxLen {
		body = body[:maxLen] + "..."
	}

	// Indent each line
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}

	return strings.Join(lines, "\n")
}

// CheckForUpdatesCached checks for updates using a cached result.
// Returns nil if no update is available or check is disabled.
func CheckForUpdatesCached(ctx context.Context, ios *iostreams.IOStreams, cachePath string) *github.Release {
	// Check if updates are disabled
	if os.Getenv("SHELLY_NO_UPDATE_CHECK") != "" {
		return nil
	}

	// Check cache
	cacheValid := false
	if info, err := os.Stat(cachePath); err == nil {
		// Cache is valid for 24 hours
		cacheValid = time.Since(info.ModTime()) < 24*time.Hour
	}

	if cacheValid {
		// Read cached version
		data, err := os.ReadFile(cachePath) //nolint:gosec // G304: cachePath is from known config directory
		if err == nil {
			cachedVersion := strings.TrimSpace(string(data))
			if cachedVersion != "" && github.IsNewerVersion(version.Version, cachedVersion) {
				// Return a minimal release with just the version
				return &github.Release{TagName: "v" + cachedVersion}
			}
			return nil
		}
	}

	// Perform check in background with short timeout
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ghClient := github.NewClient(ios)
	release, err := ghClient.GetLatestRelease(checkCtx, github.DefaultOwner, github.DefaultRepo)
	if err != nil {
		return nil
	}

	// Update cache
	if merr := os.MkdirAll(filepath.Dir(cachePath), 0o750); merr != nil {
		ios.DebugErr("creating cache directory", merr)
	}
	if werr := os.WriteFile(cachePath, []byte(release.Version()), 0o600); werr != nil {
		ios.DebugErr("writing update cache", werr)
	}

	if github.IsNewerVersion(version.Version, release.Version()) {
		return release
	}

	return nil
}
