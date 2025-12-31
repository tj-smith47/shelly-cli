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
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const (
	// DefaultTimeout for API requests.
	DefaultTimeout = 30 * time.Second
	// DefaultOwner is the repository owner for shelly-cli.
	DefaultOwner = "tj-smith47"
	// DefaultRepo is the repository name.
	DefaultRepo = "shelly-cli"
)

// GitHubAPIBaseURL is the base URL for GitHub API (var for testing).
var GitHubAPIBaseURL = "https://api.github.com"

// ErrNoReleases is returned when no releases are found.
var ErrNoReleases = errors.New("no releases found")

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Version returns the release version without the 'v' prefix.
func (r *Release) Version() string {
	return strings.TrimPrefix(r.TagName, "v")
}

// Asset represents a release asset (downloadable file).
type Asset struct {
	ID                 int64  `json:"id"`
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
		return nil, ErrNoReleases
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

// GetReleaseByTag fetches a specific release by tag name.
func (c *Client) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", GitHubAPIBaseURL, owner, repo, tag)

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
		return nil, fmt.Errorf("release %s not found", tag)
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

// ListReleases fetches all releases, optionally including prereleases.
func (c *Client) ListReleases(ctx context.Context, owner, repo string, includePrereleases bool) ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", GitHubAPIBaseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "shelly-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.ios.DebugErr("closing response body", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	// Filter out drafts and optionally prereleases
	filtered := make([]Release, 0, len(releases))
	for _, r := range releases {
		if r.Draft {
			continue
		}
		if r.Prerelease && !includePrereleases {
			continue
		}
		filtered = append(filtered, r)
	}

	return filtered, nil
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
	fs := getFs()

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
	if err := fs.MkdirAll(destDir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	outFile, err := fs.Create(destPath)
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
	fs := getFs()

	// Create temp directory for download
	tmpDir, err := afero.TempDir(fs, "", "shelly-download-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanup = func() {
		if rerr := fs.RemoveAll(tmpDir); rerr != nil {
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
	if err := fs.Chmod(binaryPath, 0o755); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to make executable: %w", err)
	}

	return binaryPath, cleanup, nil
}

// maxExtractSize is the maximum size for extracted files (100MB).
const maxExtractSize = 100 * 1024 * 1024

// extractTarGz extracts a .tar.gz file and returns the path to the binary.
func (c *Client) extractTarGz(archivePath, destDir, binaryName string) (string, error) {
	fs := getFs()

	file, err := fs.Open(archivePath)
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
	fs := getFs()

	// Read the zip file content using afero
	zipData, err := afero.ReadFile(fs, archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to read zip file: %w", err)
	}

	// Create a zip reader from the bytes
	reader, err := zip.NewReader(newBytesReaderAt(zipData), int64(len(zipData)))
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}

	return c.findBinaryInZip(reader, destDir, binaryName)
}

// bytesReaderAt wraps a byte slice to implement io.ReaderAt.
type bytesReaderAt struct {
	data []byte
}

func newBytesReaderAt(data []byte) *bytesReaderAt {
	return &bytesReaderAt{data: data}
}

func (b *bytesReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// findBinaryInZip searches for and extracts the binary from a zip reader.
func (c *Client) findBinaryInZip(reader *zip.Reader, destDir, binaryName string) (string, error) {
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
	fs := getFs()

	outFile, err := fs.Create(destPath)
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

// FindAssetForPlatform finds the appropriate asset for the current OS/arch.
func (r *Release) FindAssetForPlatform() *Asset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Normalize arch names
	archNames := []string{goarch}
	switch goarch {
	case "amd64":
		archNames = append(archNames, "x86_64", "64bit")
	case "386":
		archNames = append(archNames, "i386", "32bit")
	case "arm64":
		archNames = append(archNames, "aarch64")
	}

	// OS-specific extensions
	extensions := []string{".tar.gz", ".zip"}
	if goos == "windows" {
		extensions = []string{".zip", ".exe"}
	}

	for i := range r.Assets {
		asset := &r.Assets[i]
		name := strings.ToLower(asset.Name)

		// Skip checksums
		if strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".sha512") ||
			strings.Contains(name, "checksum") {
			continue
		}

		// Check OS
		if !strings.Contains(name, goos) {
			continue
		}

		// Check arch
		archMatch := false
		for _, a := range archNames {
			if strings.Contains(name, strings.ToLower(a)) {
				archMatch = true
				break
			}
		}
		if !archMatch {
			continue
		}

		// Check extension
		for _, ext := range extensions {
			if strings.HasSuffix(name, ext) {
				return asset
			}
		}
	}

	return nil
}

// FindChecksumAsset finds the checksum file for an asset.
func (r *Release) FindChecksumAsset(assetName string) *Asset {
	checksumNames := []string{
		assetName + ".sha256",
		assetName + ".sha256sum",
		"checksums.txt",
		"sha256sums.txt",
		"SHA256SUMS",
	}

	for i := range r.Assets {
		asset := &r.Assets[i]
		for _, name := range checksumNames {
			if strings.EqualFold(asset.Name, name) {
				return asset
			}
		}
	}

	return nil
}

// CompareVersions compares two semantic versions.
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func CompareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split into parts
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := range maxLen {
		var p1, p2 int
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	return 0
}

// parseVersion parses a version string into numeric parts.
func parseVersion(v string) []int {
	// Remove prerelease suffix (e.g., -beta.1)
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))

	for _, p := range parts {
		var num int
		_, err := fmt.Sscanf(p, "%d", &num)
		if err != nil {
			num = 0
		}
		result = append(result, num)
	}

	return result
}

// IsNewerVersion returns true if available is newer than current.
func IsNewerVersion(current, available string) bool {
	return CompareVersions(available, current) > 0
}

// SortReleasesByVersion sorts releases by version in descending order.
func SortReleasesByVersion(releases []Release) {
	sort.Slice(releases, func(i, j int) bool {
		return CompareVersions(releases[i].TagName, releases[j].TagName) > 0
	})
}

// ReleaseFetcher creates a function that fetches the latest version string.
// The returned function signature is compatible with version.ReleaseFetcher.
func ReleaseFetcher(ios *iostreams.IOStreams) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		rel, err := NewClient(ios).GetLatestRelease(ctx, DefaultOwner, DefaultRepo)
		if err != nil || rel == nil {
			return "", err
		}
		return rel.Version(), nil
	}
}

// ExtensionDownloadResult holds the result of downloading an extension from GitHub.
type ExtensionDownloadResult struct {
	LocalPath string
	TagName   string
	AssetName string
	Cleanup   func()
}

// DownloadExtensionRelease downloads the latest release of an extension from GitHub.
// The binaryPrefix is prepended to the repo name if not already present (e.g., "shelly-").
func (c *Client) DownloadExtensionRelease(ctx context.Context, owner, repo, binaryPrefix string) (*ExtensionDownloadResult, error) {
	binaryName := repo
	if !strings.HasPrefix(binaryName, binaryPrefix) {
		binaryName = binaryPrefix + repo
	}

	release, err := c.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	c.ios.Info("Found release: %s", release.TagName)

	asset, err := c.FindBinaryAsset(release, binaryName)
	if err != nil {
		return nil, fmt.Errorf("failed to find binary: %w", err)
	}

	binaryPath, cleanup, err := c.DownloadAndExtract(ctx, asset, binaryName)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	return &ExtensionDownloadResult{
		LocalPath: binaryPath,
		TagName:   release.TagName,
		AssetName: asset.Name,
		Cleanup:   cleanup,
	}, nil
}
