// Package upgrade provides self-update functionality.
package upgrade

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

const (
	// GitHubRepo is the repository for releases.
	GitHubRepo = "tj-smith47/shelly-cli"
	// GitHubAPIURL is the GitHub API base URL.
	GitHubAPIURL = "https://api.github.com"
	// CheckInterval is the minimum time between update checks.
	CheckInterval = 24 * time.Hour
)

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// NewCommand creates the upgrade command.
func NewCommand() *cobra.Command {
	var checkOnly bool
	var targetVersion string
	var preRelease bool
	var force bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade shelly to the latest version",
		Long: `Check for and install updates to shelly.

By default, downloads and installs the latest stable release.
Use --check to only check for updates without installing.
Use --version to install a specific version.
Use --pre-release to include pre-release versions.

Examples:
  shelly upgrade              # Upgrade to latest stable
  shelly upgrade --check      # Check for updates only
  shelly upgrade --version v1.2.0   # Install specific version
  shelly upgrade --pre-release      # Include pre-releases`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentVersion := version.Short()

			// Get latest release
			release, err := getRelease(targetVersion, preRelease)
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if release == nil {
				fmt.Println("No releases found.")
				return nil
			}

			// Compare versions
			latestVersion := strings.TrimPrefix(release.TagName, "v")
			currentClean := strings.TrimPrefix(currentVersion, "v")

			if !force && latestVersion == currentClean {
				fmt.Printf("Already at latest version (%s)\n", currentVersion)
				return nil
			}

			// Show update info
			fmt.Printf("Current version: %s\n", currentVersion)
			fmt.Printf("Latest version:  %s\n", release.TagName)
			if release.Prerelease {
				fmt.Println("  (pre-release)")
			}
			fmt.Printf("Published:       %s\n", release.PublishedAt.Format("2006-01-02"))
			fmt.Println()

			if checkOnly {
				if latestVersion != currentClean {
					fmt.Println("Update available! Run 'shelly upgrade' to install.")
				}
				return nil
			}

			// Find appropriate asset
			asset := findAsset(release.Assets)
			if asset == nil {
				return fmt.Errorf("no compatible binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
			}

			// Confirm upgrade
			if !force {
				confirmed, err := output.Confirm(fmt.Sprintf("Upgrade to %s?", release.TagName), true)
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Println("Upgrade cancelled.")
					return nil
				}
			}

			// Download and install
			spinner := output.NewSpinner("Downloading update...")
			spinner.Start()

			if err := downloadAndInstall(release, asset); err != nil {
				spinner.StopWithError("Download failed")
				return fmt.Errorf("failed to install update: %w", err)
			}

			spinner.StopWithSuccess("Upgrade complete!")
			fmt.Printf("\nUpgraded to %s\n", release.TagName)
			fmt.Println("Restart shelly to use the new version.")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check for updates without installing")
	cmd.Flags().StringVarP(&targetVersion, "version", "V", "", "Install a specific version")
	cmd.Flags().BoolVar(&preRelease, "pre-release", false, "Include pre-release versions")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force upgrade even if already at latest")

	return cmd
}

// getRelease fetches a release from GitHub.
func getRelease(version string, preRelease bool) (*Release, error) {
	var url string
	if version != "" {
		// Specific version
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		url = fmt.Sprintf("%s/repos/%s/releases/tags/%s", GitHubAPIURL, GitHubRepo, version)
	} else {
		// Latest release
		url = fmt.Sprintf("%s/repos/%s/releases/latest", GitHubAPIURL, GitHubRepo)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		if version != "" {
			return nil, fmt.Errorf("version %s not found", version)
		}
		// No releases yet, try listing all
		if preRelease {
			return getLatestPreRelease()
		}
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	// If we want pre-releases and this is stable, check for newer pre-release
	if preRelease && !release.Prerelease {
		preRel, err := getLatestPreRelease()
		if err == nil && preRel != nil {
			if preRel.PublishedAt.After(release.PublishedAt) {
				return preRel, nil
			}
		}
	}

	return &release, nil
}

// getLatestPreRelease gets the latest pre-release.
func getLatestPreRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases", GitHubAPIURL, GitHubRepo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	for _, r := range releases {
		if !r.Draft {
			return &r, nil
		}
	}

	return nil, nil
}

// findAsset finds the appropriate asset for the current platform.
func findAsset(assets []Asset) *Asset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Build expected filename patterns
	patterns := []string{
		fmt.Sprintf("shelly_%s_%s", goos, goarch),
		fmt.Sprintf("shelly-%s-%s", goos, goarch),
	}

	// Map common arch names
	archAliases := map[string][]string{
		"amd64": {"x86_64", "x64"},
		"arm64": {"aarch64"},
	}

	if aliases, ok := archAliases[goarch]; ok {
		for _, alias := range aliases {
			patterns = append(patterns,
				fmt.Sprintf("shelly_%s_%s", goos, alias),
				fmt.Sprintf("shelly-%s-%s", goos, alias),
			)
		}
	}

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		for _, pattern := range patterns {
			if strings.Contains(name, strings.ToLower(pattern)) {
				// Skip checksum files
				if strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".md5") {
					continue
				}
				return &asset
			}
		}
	}

	return nil
}

// downloadAndInstall downloads and installs the update.
func downloadAndInstall(release *Release, asset *Asset) error {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "shelly-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download asset
	assetPath := filepath.Join(tmpDir, asset.Name)
	if err := downloadFile(asset.BrowserDownloadURL, assetPath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Try to download and verify checksum
	checksumURL := asset.BrowserDownloadURL + ".sha256"
	checksumPath := assetPath + ".sha256"
	if err := downloadFile(checksumURL, checksumPath); err == nil {
		if err := verifyChecksum(assetPath, checksumPath); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Extract binary
	var binaryPath string
	if strings.HasSuffix(asset.Name, ".tar.gz") || strings.HasSuffix(asset.Name, ".tgz") {
		binaryPath, err = extractTarGz(assetPath, tmpDir)
	} else if strings.HasSuffix(asset.Name, ".zip") {
		binaryPath, err = extractZip(assetPath, tmpDir)
	} else {
		// Assume it's the binary itself
		binaryPath = assetPath
	}
	if err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Replace binary
	if err := replaceBinary(binaryPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// downloadFile downloads a file from a URL.
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// verifyChecksum verifies a file's SHA256 checksum.
func verifyChecksum(filePath, checksumPath string) error {
	// Read expected checksum
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}

	// Parse checksum (format: "hash  filename" or just "hash")
	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return fmt.Errorf("empty checksum file")
	}
	expected := strings.ToLower(parts[0])

	// Calculate actual checksum
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))

	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

// extractTarGz extracts a .tar.gz archive and returns the path to the shelly binary.
func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var binaryPath string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Look for the shelly binary
		name := filepath.Base(header.Name)
		if name == "shelly" || name == "shelly.exe" {
			binaryPath = filepath.Join(destDir, name)
			outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("shelly binary not found in archive")
	}

	return binaryPath, nil
}

// extractZip extracts a .zip archive and returns the path to the shelly binary.
func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var binaryPath string
	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		// Look for the shelly binary
		name := filepath.Base(f.Name)
		if name == "shelly" || name == "shelly.exe" {
			binaryPath = filepath.Join(destDir, name)

			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, f.Mode())
			if err != nil {
				rc.Close()
				return "", err
			}

			_, err = io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()
			if err != nil {
				return "", err
			}
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("shelly binary not found in archive")
	}

	return binaryPath, nil
}

// replaceBinary replaces the current binary with the new one.
func replaceBinary(newPath, oldPath string) error {
	// On Windows, we can't replace a running binary directly
	// Instead, rename the old one and copy the new one
	if runtime.GOOS == "windows" {
		bakPath := oldPath + ".bak"
		if err := os.Rename(oldPath, bakPath); err != nil {
			return err
		}
		if err := copyFile(newPath, oldPath); err != nil {
			// Try to restore backup
			os.Rename(bakPath, oldPath)
			return err
		}
		os.Remove(bakPath)
		return nil
	}

	// On Unix, we can atomically replace via rename
	// First copy to same directory to ensure same filesystem
	tmpPath := oldPath + ".new"
	if err := copyFile(newPath, tmpPath); err != nil {
		return err
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, oldPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
