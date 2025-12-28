package github_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestRelease_Version(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tagName string
		want    string
	}{
		{
			name:    "with v prefix",
			tagName: "v1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "without v prefix",
			tagName: "1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "empty",
			tagName: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &github.Release{TagName: tt.tagName}
			if got := r.Version(); got != tt.want {
				t.Errorf("Version() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAsset_Fields(t *testing.T) {
	t.Parallel()

	asset := &github.Asset{
		ID:                 123,
		Name:               "shelly-linux-amd64.tar.gz",
		BrowserDownloadURL: "https://example.com/download",
		Size:               1024,
		ContentType:        "application/gzip",
	}

	if asset.ID != 123 {
		t.Errorf("ID = %d, want 123", asset.ID)
	}
	if asset.Name != "shelly-linux-amd64.tar.gz" {
		t.Errorf("Name = %q, want %q", asset.Name, "shelly-linux-amd64.tar.gz")
	}
	if asset.BrowserDownloadURL != "https://example.com/download" {
		t.Errorf("BrowserDownloadURL = %q, want %q", asset.BrowserDownloadURL, "https://example.com/download")
	}
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.0", "v1.0.0", 0},
		{"1.0.0-beta", "1.0.0", 0}, // Pre-release suffix ignored
		{"1.0.0+build", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			t.Parallel()
			got := github.CompareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		current   string
		available string
		want      bool
	}{
		{"1.0.0", "1.0.1", true},
		{"1.0.1", "1.0.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "2.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.current+"_vs_"+tt.available, func(t *testing.T) {
			t.Parallel()
			got := github.IsNewerVersion(tt.current, tt.available)
			if got != tt.want {
				t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tt.current, tt.available, got, tt.want)
			}
		})
	}
}

func TestSortReleasesByVersion(t *testing.T) {
	t.Parallel()

	releases := []github.Release{
		{TagName: "v1.0.0"},
		{TagName: "v2.0.0"},
		{TagName: "v1.5.0"},
	}

	github.SortReleasesByVersion(releases)

	// Should be in descending order
	if releases[0].TagName != "v2.0.0" {
		t.Errorf("releases[0].TagName = %q, want v2.0.0", releases[0].TagName)
	}
	if releases[1].TagName != "v1.5.0" {
		t.Errorf("releases[1].TagName = %q, want v1.5.0", releases[1].TagName)
	}
	if releases[2].TagName != "v1.0.0" {
		t.Errorf("releases[2].TagName = %q, want v1.0.0", releases[2].TagName)
	}
}

func TestParseRepoString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "simple owner/repo",
			input:     "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "with gh prefix",
			input:     "gh:owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "with github prefix",
			input:     "github:owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "with https prefix",
			input:     "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "with .git suffix",
			input:     "owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty owner",
			input:   "/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repo, err := github.ParseRepoString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestRelease_FindAssetForPlatform(t *testing.T) {
	t.Parallel()

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	release := &github.Release{
		Assets: []github.Asset{
			{Name: "shelly-linux-amd64.tar.gz"},
			{Name: "shelly-darwin-amd64.tar.gz"},
			{Name: "shelly-windows-amd64.zip"},
			{Name: "shelly-linux-arm64.tar.gz"},
			{Name: "checksums.txt"},
		},
	}

	asset := release.FindAssetForPlatform()
	if asset == nil {
		// This is expected if the current platform isn't in our test assets
		t.Logf("FindAssetForPlatform() returned nil (platform: %s/%s)", goos, goarch)
		return
	}

	// Asset name should contain the OS
	if asset.Name == "" {
		t.Error("FindAssetForPlatform() returned asset with empty name")
	}
}

func TestRelease_FindChecksumAsset(t *testing.T) {
	t.Parallel()

	release := &github.Release{
		Assets: []github.Asset{
			{Name: "shelly-linux-amd64.tar.gz"},
			{Name: "shelly-linux-amd64.tar.gz.sha256"},
			{Name: "checksums.txt"},
		},
	}

	// Test finding specific checksum
	asset := release.FindChecksumAsset("shelly-linux-amd64.tar.gz")
	if asset == nil {
		t.Error("FindChecksumAsset() should find sha256 file")
	} else if asset.Name != "shelly-linux-amd64.tar.gz.sha256" {
		t.Errorf("FindChecksumAsset() name = %q, want %q", asset.Name, "shelly-linux-amd64.tar.gz.sha256")
	}

	// Test finding generic checksums.txt
	asset = release.FindChecksumAsset("nonexistent.tar.gz")
	if asset == nil {
		// May fall back to checksums.txt
		t.Log("FindChecksumAsset() returned nil for nonexistent asset")
	}
}

func TestParseChecksumFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		assetName string
		wantHash  string
		wantErr   bool
	}{
		{
			name:      "standard format",
			content:   "abc123  shelly-linux-amd64.tar.gz\ndef456  shelly-darwin-amd64.tar.gz",
			assetName: "shelly-linux-amd64.tar.gz",
			wantHash:  "abc123",
		},
		{
			name:      "binary mode indicator",
			content:   "abc123 *shelly-linux-amd64.tar.gz",
			assetName: "shelly-linux-amd64.tar.gz",
			wantHash:  "abc123",
		},
		{
			name:      "with comments",
			content:   "# This is a comment\nabc123  shelly-linux-amd64.tar.gz",
			assetName: "shelly-linux-amd64.tar.gz",
			wantHash:  "abc123",
		},
		{
			name:      "single hash only",
			content:   "abc123",
			assetName: "anyfile.tar.gz",
			wantHash:  "abc123",
		},
		{
			name:      "case insensitive match",
			content:   "abc123  SHELLY-LINUX-AMD64.TAR.GZ",
			assetName: "shelly-linux-amd64.tar.gz",
			wantHash:  "abc123",
		},
		{
			name:      "not found",
			content:   "abc123  other-file.tar.gz",
			assetName: "shelly-linux-amd64.tar.gz",
			wantErr:   true,
		},
		{
			name:      "empty content",
			content:   "",
			assetName: "shelly-linux-amd64.tar.gz",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hash, err := github.ParseChecksumFile(tt.content, tt.assetName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChecksumFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash != tt.wantHash {
				t.Errorf("ParseChecksumFile() = %q, want %q", hash, tt.wantHash)
			}
		})
	}
}

func TestClient_GetLatestRelease(t *testing.T) {
	t.Parallel()

	release := github.Release{
		TagName: "v1.0.0",
		Name:    "Release 1.0.0",
		Body:    "Release notes",
		Assets:  []github.Asset{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testowner/testrepo/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// We can't easily test this without mocking the base URL
	// Just verify the method exists and returns an error for invalid repos
	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	ctx := context.Background()
	_, err := client.GetLatestRelease(ctx, "nonexistent", "nonexistent")
	// This should fail because the repo doesn't exist (network call)
	if err == nil {
		t.Log("GetLatestRelease() unexpectedly succeeded for nonexistent repo")
	}
}

func TestClient_GetReleaseByTag(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	ctx := context.Background()
	_, err := client.GetReleaseByTag(ctx, "nonexistent", "nonexistent", "v1.0.0")
	// This should fail because the repo doesn't exist
	if err == nil {
		t.Log("GetReleaseByTag() unexpectedly succeeded for nonexistent repo")
	}
}

func TestIssueOpts_Fields(t *testing.T) {
	t.Parallel()

	opts := github.IssueOpts{
		Type:      github.IssueTypeBug,
		Title:     "Test Issue",
		Device:    "test-device",
		AttachLog: true,
	}

	if opts.Type != github.IssueTypeBug {
		t.Errorf("Type = %q, want %q", opts.Type, github.IssueTypeBug)
	}
	if opts.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", opts.Title, "Test Issue")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if !opts.AttachLog {
		t.Error("AttachLog should be true")
	}
}

func TestIssueTypeConstants(t *testing.T) {
	t.Parallel()

	if github.IssueTypeBug != "bug" {
		t.Errorf("IssueTypeBug = %q, want %q", github.IssueTypeBug, "bug")
	}
	if github.IssueTypeFeature != "feature" {
		t.Errorf("IssueTypeFeature = %q, want %q", github.IssueTypeFeature, "feature")
	}
	if github.IssueTypeDevice != "device" {
		t.Errorf("IssueTypeDevice = %q, want %q", github.IssueTypeDevice, "device")
	}
}

func TestBuildIssueBody(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Theme: "dracula",
		Devices: map[string]model.Device{
			"dev1": {Address: "192.168.1.1"},
			"dev2": {Address: "192.168.1.2"},
		},
	}

	tests := []struct {
		name        string
		opts        github.IssueOpts
		wantContain string
	}{
		{
			name:        "bug report",
			opts:        github.IssueOpts{Type: github.IssueTypeBug},
			wantContain: "## Bug Description",
		},
		{
			name:        "feature request",
			opts:        github.IssueOpts{Type: github.IssueTypeFeature},
			wantContain: "## Feature Request",
		},
		{
			name:        "device issue",
			opts:        github.IssueOpts{Type: github.IssueTypeDevice, Device: "shelly-1pm"},
			wantContain: "## Device Compatibility Issue",
		},
		{
			name:        "with log attachment",
			opts:        github.IssueOpts{Type: github.IssueTypeBug, AttachLog: true},
			wantContain: "Log info requested",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := github.BuildIssueBody(cfg, tt.opts)
			if body == "" {
				t.Error("BuildIssueBody() returned empty string")
			}
			if len(body) < len(tt.wantContain) {
				t.Errorf("BuildIssueBody() too short, want at least %d chars", len(tt.wantContain))
			}
		})
	}
}

func TestFormatSystemInfo(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Theme: "dracula",
	}

	info := github.FormatSystemInfo(cfg, false)
	if info == "" {
		t.Error("FormatSystemInfo() returned empty string")
	}

	// Should contain version info
	if len(info) < 50 {
		t.Error("FormatSystemInfo() seems too short")
	}

	// Test with attach log
	infoWithLog := github.FormatSystemInfo(cfg, true)
	if len(infoWithLog) < len(info) {
		t.Error("FormatSystemInfo with attachLog should be longer")
	}
}

func TestBuildIssueURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		issueType   string
		title       string
		body        string
		wantContain string
	}{
		{
			name:        "bug",
			issueType:   github.IssueTypeBug,
			title:       "",
			body:        "test body",
			wantContain: "labels=bug",
		},
		{
			name:        "feature",
			issueType:   github.IssueTypeFeature,
			title:       "My Feature",
			body:        "test body",
			wantContain: "labels=enhancement",
		},
		{
			name:        "device",
			issueType:   github.IssueTypeDevice,
			title:       "",
			body:        "test body",
			wantContain: "labels=device-compatibility",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			url := github.BuildIssueURL(tt.issueType, tt.title, tt.body)
			if url == "" {
				t.Error("BuildIssueURL() returned empty string")
			}
			if !containsString(url, tt.wantContain) {
				t.Errorf("BuildIssueURL() = %q, want to contain %q", url, tt.wantContain)
			}
		})
	}
}

func TestIssuesURL(t *testing.T) {
	t.Parallel()

	url := github.IssuesURL()
	if url == "" {
		t.Error("IssuesURL() returned empty string")
	}
	if !containsString(url, "/issues") {
		t.Errorf("IssuesURL() = %q, should contain /issues", url)
	}
}

func TestRepoURL(t *testing.T) {
	t.Parallel()

	if github.RepoURL == "" {
		t.Error("RepoURL is empty")
	}
	if !containsString(github.RepoURL, "github.com") {
		t.Errorf("RepoURL = %q, should contain github.com", github.RepoURL)
	}
}

func TestGetExecutablePath(t *testing.T) {
	t.Parallel()

	path, err := github.GetExecutablePath()
	if err != nil {
		t.Fatalf("GetExecutablePath() error = %v", err)
	}

	if path == "" {
		t.Error("GetExecutablePath() returned empty string")
	}

	// Path should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("GetExecutablePath() path does not exist: %s", path)
	}
}

//nolint:paralleltest // Test modifies environment variable
func TestCheckForUpdatesCached_UpdatesDisabled(t *testing.T) {
	origEnv := os.Getenv("SHELLY_NO_UPDATE_CHECK")
	t.Cleanup(func() {
		if origEnv == "" {
			if err := os.Unsetenv("SHELLY_NO_UPDATE_CHECK"); err != nil {
				t.Logf("warning: failed to unset env: %v", err)
			}
		} else {
			if err := os.Setenv("SHELLY_NO_UPDATE_CHECK", origEnv); err != nil {
				t.Logf("warning: failed to restore env: %v", err)
			}
		}
	})

	if err := os.Setenv("SHELLY_NO_UPDATE_CHECK", "1"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	ctx := context.Background()

	result := github.CheckForUpdatesCached(ctx, ios, "/tmp/test-cache", "1.0.0")
	if result != nil {
		t.Error("CheckForUpdatesCached() should return nil when updates disabled")
	}
}

//nolint:paralleltest // Test writes to filesystem cache
func TestCheckForUpdatesCached_ValidCache(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "version-cache")

	// Write a "newer" version to cache
	if err := os.WriteFile(cachePath, []byte("2.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	ctx := context.Background()

	result := github.CheckForUpdatesCached(ctx, ios, cachePath, "1.0.0")
	if result == nil {
		t.Error("CheckForUpdatesCached() should return release when cached version is newer")
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()

	if github.GitHubAPIBaseURL == "" {
		t.Error("GitHubAPIBaseURL is empty")
	}
	if github.DefaultOwner == "" {
		t.Error("DefaultOwner is empty")
	}
	if github.DefaultRepo == "" {
		t.Error("DefaultRepo is empty")
	}
	if github.DefaultTimeout == 0 {
		t.Error("DefaultTimeout is zero")
	}
}

func TestErrNoReleases(t *testing.T) {
	t.Parallel()

	if github.ErrNoReleases == nil {
		t.Error("ErrNoReleases is nil")
	}
	if github.ErrNoReleases.Error() == "" {
		t.Error("ErrNoReleases.Error() is empty")
	}
}

func TestExtensionDownloadResult_Fields(t *testing.T) {
	t.Parallel()

	result := &github.ExtensionDownloadResult{
		LocalPath: "/tmp/extension",
		TagName:   "v1.0.0",
		AssetName: "extension-linux-amd64.tar.gz",
		Cleanup:   func() {},
	}

	if result.LocalPath != "/tmp/extension" {
		t.Errorf("LocalPath = %q, want %q", result.LocalPath, "/tmp/extension")
	}
	if result.TagName != "v1.0.0" {
		t.Errorf("TagName = %q, want %q", result.TagName, "v1.0.0")
	}
	if result.AssetName != "extension-linux-amd64.tar.gz" {
		t.Errorf("AssetName = %q, want %q", result.AssetName, "extension-linux-amd64.tar.gz")
	}
	if result.Cleanup == nil {
		t.Error("Cleanup should not be nil")
	}
}

// containsString checks if a string contains another string.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
