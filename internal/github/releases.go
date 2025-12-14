// Package github provides GitHub API integration for downloading releases.
package github

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const (
	// GitHubAPIBaseURL is the base URL for GitHub API.
	GitHubAPIBaseURL = "https://api.github.com"
	// DefaultTimeout for API requests.
	DefaultTimeout = 30 * time.Second
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset (downloadable file).
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

// Client provides methods for interacting with GitHub releases.
type Client struct {
	httpClient *http.Client
	ios        *iostreams.IOStreams
}

// NewClient creates a new GitHub client.
func NewClient(ios *iostreams.IOStreams) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		ios: ios,
	}
}

// GetLatestRelease fetches the latest release for a repository.
func (c *Client) GetLatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", GitHubAPIBaseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "shelly-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.ios.DebugErr("closing response body", cerr)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository %s/%s not found or has no releases", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// FindBinaryAsset finds the appropriate binary asset for the current platform.
func (c *Client) FindBinaryAsset(release *Release, binaryName string) (*Asset, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Build list of patterns to match (in order of preference)
	patterns := []string{
		// Exact match: shelly-myext-linux-amd64
		fmt.Sprintf("%s-%s-%s", binaryName, goos, goarch),
		// With underscore: shelly-myext_linux_amd64
		fmt.Sprintf("%s_%s_%s", binaryName, goos, goarch),
		// Darwin as macos: shelly-myext-macos-amd64
		fmt.Sprintf("%s-%s-%s", binaryName, strings.Replace(goos, "darwin", "macos", 1), goarch),
		// x86_64 instead of amd64
		fmt.Sprintf("%s-%s-%s", binaryName, goos, strings.Replace(goarch, "amd64", "x86_64", 1)),
		// arm64 as aarch64
		fmt.Sprintf("%s-%s-%s", binaryName, goos, strings.Replace(goarch, "arm64", "aarch64", 1)),
	}

	// Also try with common extensions
	extensions := []string{"", ".exe", ".tar.gz", ".zip"}

	for _, pattern := range patterns {
		for _, ext := range extensions {
			candidate := pattern + ext
			for i := range release.Assets {
				asset := &release.Assets[i]
				if strings.EqualFold(asset.Name, candidate) {
					return asset, nil
				}
			}
		}
	}

	// Fallback: try to find any asset containing the OS and arch
	for i := range release.Assets {
		asset := &release.Assets[i]
		nameLower := strings.ToLower(asset.Name)
		if strings.Contains(nameLower, goos) && strings.Contains(nameLower, goarch) {
			return asset, nil
		}
	}

	// List available assets in error message
	assetNames := make([]string, 0, len(release.Assets))
	for _, asset := range release.Assets {
		assetNames = append(assetNames, asset.Name)
	}

	return nil, fmt.Errorf("no binary found for %s/%s; available assets: %v", goos, goarch, assetNames)
}

// DownloadAsset downloads a release asset to a local file.
func (c *Client) DownloadAsset(ctx context.Context, asset *Asset, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "shelly-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.ios.DebugErr("closing response body", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Create destination file
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	outFile, err := os.Create(destPath) //nolint:gosec // G304: destPath is from controlled temp directory
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			c.ios.DebugErr("closing output file", cerr)
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DownloadAndExtract downloads a release binary and extracts it if needed.
// Returns the path to the extracted binary.
func (c *Client) DownloadAndExtract(ctx context.Context, asset *Asset, binaryName string) (binaryPath string, cleanup func(), err error) {
	// Create temp directory for download
	tmpDir, err := os.MkdirTemp("", "shelly-download-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanup = func() {
		if rerr := os.RemoveAll(tmpDir); rerr != nil {
			c.ios.DebugErr("removing temp directory", rerr)
		}
	}

	downloadPath := filepath.Join(tmpDir, asset.Name)

	// Download the asset
	if err := c.DownloadAsset(ctx, asset, downloadPath); err != nil {
		cleanup()
		return "", nil, err
	}

	// Check if it needs extraction
	switch {
	case strings.HasSuffix(asset.Name, ".tar.gz") || strings.HasSuffix(asset.Name, ".tgz"):
		binaryPath, err = c.extractTarGz(downloadPath, tmpDir, binaryName)
	case strings.HasSuffix(asset.Name, ".zip"):
		binaryPath, err = c.extractZip(downloadPath, tmpDir, binaryName)
	default:
		// Assume it's already the binary
		binaryPath = downloadPath
	}

	if err != nil {
		cleanup()
		return "", nil, err
	}

	// Make it executable - downloaded binaries must be executable
	if err := os.Chmod(binaryPath, 0o755); err != nil { //nolint:gosec // G302: binaries require executable permissions
		cleanup()
		return "", nil, fmt.Errorf("failed to make executable: %w", err)
	}

	return binaryPath, cleanup, nil
}

// maxExtractSize is the maximum size for extracted files (100MB).
const maxExtractSize = 100 * 1024 * 1024

// extractTarGz extracts a .tar.gz file and returns the path to the binary.
func (c *Client) extractTarGz(archivePath, destDir, binaryName string) (string, error) {
	file, err := os.Open(archivePath) //nolint:gosec // archivePath is from trusted temp directory
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			c.ios.DebugErr("closing archive file", cerr)
		}
	}()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if cerr := gzReader.Close(); cerr != nil {
			c.ios.DebugErr("closing gzip reader", cerr)
		}
	}()

	return c.findBinaryInTar(tar.NewReader(gzReader), destDir, binaryName)
}

// findBinaryInTar searches for and extracts the binary from a tar reader.
func (c *Client) findBinaryInTar(tarReader *tar.Reader, destDir, binaryName string) (string, error) {
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		filename := filepath.Base(header.Name)
		if !c.matchesBinaryName(filename, binaryName) {
			continue
		}

		destPath := filepath.Join(destDir, filename)
		if err := c.extractToFile(destPath, tarReader); err != nil {
			return "", err
		}
		return destPath, nil
	}

	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}

// extractZip extracts a .zip file and returns the path to the binary.
func (c *Client) extractZip(archivePath, destDir, binaryName string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() {
		if cerr := reader.Close(); cerr != nil {
			c.ios.DebugErr("closing zip reader", cerr)
		}
	}()

	return c.findBinaryInZip(reader, destDir, binaryName)
}

// findBinaryInZip searches for and extracts the binary from a zip reader.
func (c *Client) findBinaryInZip(reader *zip.ReadCloser, destDir, binaryName string) (string, error) {
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		filename := filepath.Base(file.Name)
		if !c.matchesBinaryName(filename, binaryName) {
			continue
		}

		destPath := filepath.Join(destDir, filename)
		if err := c.extractZipEntry(file, destPath); err != nil {
			return "", err
		}
		return destPath, nil
	}

	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}

// extractZipEntry extracts a single file from a zip archive.
func (c *Client) extractZipEntry(file *zip.File, destPath string) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			c.ios.DebugErr("closing zip entry", cerr)
		}
	}()

	return c.extractToFile(destPath, rc)
}

// matchesBinaryName checks if a filename matches the expected binary name.
func (c *Client) matchesBinaryName(filename, binaryName string) bool {
	return filename == binaryName || strings.HasPrefix(filename, binaryName)
}

// extractToFile extracts data from a reader to a file with size limit.
func (c *Client) extractToFile(destPath string, r io.Reader) error {
	outFile, err := os.Create(destPath) //nolint:gosec // destPath is in trusted temp directory
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			c.ios.DebugErr("closing output file", cerr)
		}
	}()

	_, err = io.CopyN(outFile, r, maxExtractSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}

// ParseRepoString parses a "gh:owner/repo" or "owner/repo" string.
func ParseRepoString(s string) (owner, repo string, err error) {
	s = strings.TrimPrefix(s, "gh:")
	s = strings.TrimPrefix(s, "github:")
	s = strings.TrimPrefix(s, "https://github.com/")

	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: expected 'owner/repo', got %q", s)
	}

	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repository format: owner and repo cannot be empty")
	}

	return owner, repo, nil
}
