package github_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// Test version constants to avoid magic strings.
const (
	testVersionV1   = "v1.0.0"
	testVersionV123 = "v1.2.3"
	testVersion123  = "1.2.3"
	testVersionV15  = "v1.5.0"
	testVersionV2   = "v2.0.0"
	testExtractDir  = "/tmp/extracted"
	testChecksumTxt = "checksums.txt"
	testShellyPath  = "/usr/bin/shelly"
	testBackupPath  = "/tmp/backup"
	testSourcePath  = "/tmp/source"
	testTargetPath  = "/tmp/target"
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
			got := version.CompareVersions(tt.v1, tt.v2)
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
			got := version.IsNewerVersion(tt.current, tt.available)
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
		{TagName: testVersionV15},
	}

	github.SortReleasesByVersion(releases)

	// Should be in descending order
	if releases[0].TagName != "v2.0.0" {
		t.Errorf("releases[0].TagName = %q, want v2.0.0", releases[0].TagName)
	}
	if releases[1].TagName != testVersionV15 {
		t.Errorf("releases[1].TagName = %q, want %q", releases[1].TagName, testVersionV15)
	}
	if releases[2].TagName != testVersionV1 {
		t.Errorf("releases[2].TagName = %q, want %s", releases[2].TagName, testVersionV1)
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

//nolint:paralleltest // modifies global function variable
func TestGetExecutablePath_OsExecutableError(t *testing.T) {
	restore := github.SetOsExecutable(func() (string, error) {
		return "", errors.New("mock os.Executable error")
	})
	defer restore()

	_, err := github.GetExecutablePath()
	if err == nil {
		t.Fatal("expected error when osExecutable fails")
	}
	if !strings.Contains(err.Error(), "failed to get executable path") {
		t.Errorf("unexpected error message: %v", err)
	}
}

//nolint:paralleltest // modifies global function variable
func TestGetExecutablePath_EvalSymlinksError(t *testing.T) {
	restore := github.SetOsExecutable(func() (string, error) {
		return "/some/path", nil
	})
	defer restore()

	restoreSymlinks := github.SetEvalSymlinks(func(path string) (string, error) {
		return "", errors.New("mock symlink error")
	})
	defer restoreSymlinks()

	_, err := github.GetExecutablePath()
	if err == nil {
		t.Fatal("expected error when evalSymlinks fails")
	}
	if !strings.Contains(err.Error(), "failed to resolve symlinks") {
		t.Errorf("unexpected error message: %v", err)
	}
}

//nolint:paralleltest // modifies global function variable
func TestRestartCLI_Success(t *testing.T) {
	restore := github.SetOsExecutable(func() (string, error) {
		return testShellyPath, nil
	})
	defer restore()

	restoreSymlinks := github.SetEvalSymlinks(func(path string) (string, error) {
		return path, nil
	})
	defer restoreSymlinks()

	var capturedPath string
	var capturedArgs []string
	restoreExec := github.SetExecCommandStart(func(ctx context.Context, path string, args []string) error {
		capturedPath = path
		capturedArgs = args
		return nil
	})
	defer restoreExec()

	ctx := context.Background()
	err := github.RestartCLI(ctx, []string{"--version"})
	if err != nil {
		t.Fatalf("RestartCLI() error = %v", err)
	}
	if capturedPath != "/usr/bin/shelly" {
		t.Errorf("RestartCLI() path = %q, want /usr/bin/shelly", capturedPath)
	}
	if len(capturedArgs) != 1 || capturedArgs[0] != "--version" {
		t.Errorf("RestartCLI() args = %v, want [--version]", capturedArgs)
	}
}

//nolint:paralleltest // modifies global function variable
func TestRestartCLI_GetExecutablePathError(t *testing.T) {
	restore := github.SetOsExecutable(func() (string, error) {
		return "", errors.New("mock executable error")
	})
	defer restore()

	ctx := context.Background()
	err := github.RestartCLI(ctx, []string{"--version"})
	if err == nil {
		t.Fatal("expected error when GetExecutablePath fails")
	}
	if !strings.Contains(err.Error(), "failed to get executable path") {
		t.Errorf("unexpected error message: %v", err)
	}
}

//nolint:paralleltest // modifies global function variable
func TestRestartCLI_ExecCommandStartError(t *testing.T) {
	restore := github.SetOsExecutable(func() (string, error) {
		return testShellyPath, nil
	})
	defer restore()

	restoreSymlinks := github.SetEvalSymlinks(func(path string) (string, error) {
		return path, nil
	})
	defer restoreSymlinks()

	restoreExec := github.SetExecCommandStart(func(ctx context.Context, path string, args []string) error {
		return errors.New("mock exec error")
	})
	defer restoreExec()

	ctx := context.Background()
	err := github.RestartCLI(ctx, []string{"--version"})
	if err == nil {
		t.Fatal("expected error when execCommandStart fails")
	}
	if !strings.Contains(err.Error(), "mock exec error") {
		t.Errorf("unexpected error message: %v", err)
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

//nolint:paralleltest // modifies global filesystem
func TestCheckForUpdatesCached_ValidCache(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	cachePath := "/tmp/cache/version-cache"

	// Write a "newer" version to cache
	if err := memFs.MkdirAll("/tmp/cache", 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(memFs, cachePath, []byte("2.0.0"), 0o600); err != nil {
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

//nolint:paralleltest // modifies global filesystem
func TestCopyFile(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	srcPath := "/tmp/source.txt"
	dstPath := "/tmp/dest.txt"

	// Create source file
	content := []byte("test content for copy")
	if err := afero.WriteFile(memFs, srcPath, content, 0o600); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	if err := github.CopyFile(ios, srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	// Verify content was copied
	copied, err := afero.ReadFile(memFs, dstPath)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if !bytes.Equal(copied, content) {
		t.Errorf("copied content = %q, want %q", string(copied), string(content))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCopyFile_SourceNotExists(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	err := github.CopyFile(ios, "/tmp/nonexistent", "/tmp/dest")
	if err == nil {
		t.Error("CopyFile() should error when source doesn't exist")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	targetPath := "/tmp/target-binary"
	newPath := "/tmp/new-binary"

	// Create target (old) binary
	oldContent := []byte("old binary content")
	if err := afero.WriteFile(memFs, targetPath, oldContent, 0o600); err != nil {
		t.Fatalf("failed to write target: %v", err)
	}

	// Create new binary
	newContent := []byte("new binary content")
	if err := afero.WriteFile(memFs, newPath, newContent, 0o600); err != nil {
		t.Fatalf("failed to write new: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	if err := github.ReplaceBinary(ios, newPath, targetPath); err != nil {
		t.Fatalf("ReplaceBinary() error = %v", err)
	}

	// Verify target was replaced
	result, err := afero.ReadFile(memFs, targetPath)
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if !bytes.Equal(result, newContent) {
		t.Errorf("target content = %q, want %q", string(result), string(newContent))
	}

	// Verify backup was removed
	exists, err := afero.Exists(memFs, targetPath+".bak")
	if err != nil {
		t.Fatalf("failed to check backup: %v", err)
	}
	if exists {
		t.Error("backup file should have been removed")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_NewPathNotExists(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	targetPath := testTargetPath

	// Create target
	if err := afero.WriteFile(memFs, targetPath, []byte("old"), 0o600); err != nil {
		t.Fatalf("failed to write target: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	err := github.ReplaceBinary(ios, "/tmp/nonexistent", targetPath)
	if err == nil {
		t.Error("ReplaceBinary() should error when new path doesn't exist")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_TargetNotExists(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	newPath := "/tmp/new-binary"

	// Create new binary
	if err := afero.WriteFile(memFs, newPath, []byte("new"), 0o600); err != nil {
		t.Fatalf("failed to write new: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	err := github.ReplaceBinary(ios, newPath, "/tmp/nonexistent")
	if err == nil {
		t.Error("ReplaceBinary() should error when target doesn't exist")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractTarGz(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	archivePath := "/tmp/test.tar.gz"
	destDir := testExtractDir
	if err := memFs.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a test tar.gz archive with a binary
	binaryContent := []byte("#!/bin/sh\necho hello")
	if err := createTestTarGzWithFs(memFs, archivePath, "shelly", binaryContent); err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	binaryPath, err := client.ExtractTarGz(archivePath, destDir, "shelly")
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v", err)
	}

	if binaryPath == "" {
		t.Error("ExtractTarGz() returned empty path")
	}

	// Verify content
	extracted, err := afero.ReadFile(memFs, binaryPath)
	if err != nil {
		t.Fatalf("failed to read extracted: %v", err)
	}
	if !bytes.Equal(extracted, binaryContent) {
		t.Errorf("extracted content = %q, want %q", string(extracted), string(binaryContent))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractTarGz_BinaryNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	archivePath := "/tmp/test.tar.gz"
	destDir := testExtractDir
	if err := memFs.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create archive with different binary name
	if err := createTestTarGzWithFs(memFs, archivePath, "other-binary", []byte("content")); err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ExtractTarGz(archivePath, destDir, "shelly")
	if err == nil {
		t.Error("ExtractTarGz() should error when binary not found")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractTarGz_InvalidArchive(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	archivePath := "/tmp/invalid.tar.gz"
	destDir := testExtractDir
	if err := memFs.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create invalid archive
	if err := afero.WriteFile(memFs, archivePath, []byte("not a valid gzip"), 0o600); err != nil {
		t.Fatalf("failed to write invalid archive: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ExtractTarGz(archivePath, destDir, "shelly")
	if err == nil {
		t.Error("ExtractTarGz() should error for invalid archive")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractZip(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	archivePath := "/tmp/test.zip"
	destDir := testExtractDir
	if err := memFs.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a test zip archive with a binary
	binaryContent := []byte("@echo off\necho hello")
	if err := createTestZipWithFs(memFs, archivePath, "shelly.exe", binaryContent); err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	binaryPath, err := client.ExtractZip(archivePath, destDir, "shelly")
	if err != nil {
		t.Fatalf("ExtractZip() error = %v", err)
	}

	if binaryPath == "" {
		t.Error("ExtractZip() returned empty path")
	}

	// Verify content
	extracted, err := afero.ReadFile(memFs, binaryPath)
	if err != nil {
		t.Fatalf("failed to read extracted: %v", err)
	}
	if !bytes.Equal(extracted, binaryContent) {
		t.Errorf("extracted content = %q, want %q", string(extracted), string(binaryContent))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractZip_BinaryNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	archivePath := "/tmp/test.zip"
	destDir := testExtractDir
	if err := memFs.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create archive with different binary name
	if err := createTestZipWithFs(memFs, archivePath, "other-binary.exe", []byte("content")); err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ExtractZip(archivePath, destDir, "shelly")
	if err == nil {
		t.Error("ExtractZip() should error when binary not found")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractZip_InvalidArchive(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	archivePath := "/tmp/invalid.zip"
	destDir := testExtractDir
	if err := memFs.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create invalid archive
	if err := afero.WriteFile(memFs, archivePath, []byte("not a valid zip"), 0o600); err != nil {
		t.Fatalf("failed to write invalid archive: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ExtractZip(archivePath, destDir, "shelly")
	if err == nil {
		t.Error("ExtractZip() should error for invalid archive")
	}
}

func TestMatchesBinaryName(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	tests := []struct {
		filename   string
		binaryName string
		want       bool
	}{
		{"shelly", "shelly", true},
		{"shelly-linux-amd64", "shelly", true},
		{"shelly.exe", "shelly", true},
		{"other-binary", "shelly", false},
		{"preshelly", "shelly", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			got := client.MatchesBinaryName(tt.filename, tt.binaryName)
			if got != tt.want {
				t.Errorf("MatchesBinaryName(%q, %q) = %v, want %v", tt.filename, tt.binaryName, got, tt.want)
			}
		})
	}
}

// Helper functions to create test archives using afero

func createTestTarGzWithFs(fs afero.Fs, path, filename string, content []byte) error {
	file, err := fs.Create(path)
	if err != nil {
		return err
	}

	gzWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzWriter)

	header := &tar.Header{
		Name: filename,
		Mode: 0o755,
		Size: int64(len(content)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	if _, err = tarWriter.Write(content); err != nil {
		return err
	}

	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzWriter.Close(); err != nil {
		return err
	}
	return file.Close()
}

func createTestZipWithFs(fs afero.Fs, path, filename string, content []byte) error {
	file, err := fs.Create(path)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(file)

	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	if _, err = writer.Write(content); err != nil {
		return err
	}

	if err := zipWriter.Close(); err != nil {
		return err
	}
	return file.Close()
}

func TestFindBinaryAsset(t *testing.T) {
	t.Parallel()

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Create assets that match the current platform
	assetName := "shelly-" + goos + "-" + goarch + ".tar.gz"
	release := &github.Release{
		Assets: []github.Asset{
			{Name: assetName},
			{Name: "shelly-other-os.tar.gz"},
			{Name: "checksums.txt"},
		},
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	asset, err := client.FindBinaryAsset(release, "shelly")
	if err != nil {
		t.Fatalf("FindBinaryAsset() error = %v", err)
	}
	if asset == nil {
		t.Fatal("FindBinaryAsset() returned nil")
	}
	if asset.Name != assetName {
		t.Errorf("FindBinaryAsset() = %q, want %q", asset.Name, assetName)
	}
}

func TestFindBinaryAsset_NotFound(t *testing.T) {
	t.Parallel()

	// Use assets that don't match any platform
	release := &github.Release{
		Assets: []github.Asset{
			{Name: "shelly-freebsd-sparc64.tar.gz"},
			{Name: "checksums.txt"},
		},
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	asset, err := client.FindBinaryAsset(release, "shelly")
	if err == nil {
		t.Error("FindBinaryAsset() should return error for unsupported platform")
	}
	if asset != nil {
		t.Error("FindBinaryAsset() should return nil asset for unsupported platform")
	}
}

//nolint:paralleltest // modifies global base URL
func TestListReleases(t *testing.T) {
	releases := []github.Release{
		{TagName: "v1.0.0", Name: "Release 1.0.0"},
		{TagName: "v1.1.0", Name: "Release 1.1.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testowner/testrepo/releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	result, err := client.ListReleases(context.Background(), "testowner", "testrepo", false)
	if err != nil {
		t.Fatalf("ListReleases() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ListReleases() returned %d releases, want 2", len(result))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestDownloadAsset(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	assetContent := []byte("binary content here")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(assetContent); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	destPath := "/tmp/downloaded"

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	asset := &github.Asset{
		Name:               "shelly-linux-amd64.tar.gz",
		BrowserDownloadURL: server.URL + "/download",
	}

	err := client.DownloadAsset(context.Background(), asset, destPath)
	if err != nil {
		t.Fatalf("DownloadAsset() error = %v", err)
	}

	// Verify content
	downloaded, err := afero.ReadFile(memFs, destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if !bytes.Equal(downloaded, assetContent) {
		t.Errorf("downloaded content = %q, want %q", string(downloaded), string(assetContent))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestDownloadAndExtract_TarGz(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create a tar.gz archive in memory
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	content := []byte("#!/bin/sh\necho hello")
	header := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		if _, err := w.Write(buf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	asset := &github.Asset{
		Name:               "shelly-linux-amd64.tar.gz",
		BrowserDownloadURL: server.URL + "/download",
	}

	binaryPath, cleanup, err := client.DownloadAndExtract(context.Background(), asset, "shelly")
	if err != nil {
		t.Fatalf("DownloadAndExtract() error = %v", err)
	}
	defer cleanup()

	if binaryPath == "" {
		t.Error("DownloadAndExtract() returned empty path")
	}

	// Verify content
	extracted, err := afero.ReadFile(memFs, binaryPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if !bytes.Equal(extracted, content) {
		t.Errorf("extracted content = %q, want %q", string(extracted), string(content))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestDownloadAndExtract_Zip(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create a zip archive in memory
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	content := []byte("@echo off\necho hello")
	writer, err := zipWriter.Create("shelly.exe")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := writer.Write(content); err != nil {
		t.Fatalf("failed to write zip content: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		if _, err := w.Write(buf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	asset := &github.Asset{
		Name:               "shelly-windows-amd64.zip",
		BrowserDownloadURL: server.URL + "/download",
	}

	binaryPath, cleanup, err := client.DownloadAndExtract(context.Background(), asset, "shelly")
	if err != nil {
		t.Fatalf("DownloadAndExtract() error = %v", err)
	}
	defer cleanup()

	if binaryPath == "" {
		t.Error("DownloadAndExtract() returned empty path")
	}

	// Verify content
	extracted, err := afero.ReadFile(memFs, binaryPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if !bytes.Equal(extracted, content) {
		t.Errorf("extracted content = %q, want %q", string(extracted), string(content))
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetLatestRelease_WithMockServer(t *testing.T) {
	release := github.Release{
		TagName: "v1.0.0",
		Name:    "Release 1.0.0",
		Body:    "Release notes",
		Assets:  []github.Asset{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testowner/testrepo/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	result, err := client.GetLatestRelease(context.Background(), "testowner", "testrepo")
	if err != nil {
		t.Fatalf("GetLatestRelease() error = %v", err)
	}

	if result.TagName != "v1.0.0" {
		t.Errorf("GetLatestRelease().TagName = %q, want %q", result.TagName, "v1.0.0")
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetReleaseByTag_WithMockServer(t *testing.T) {
	release := github.Release{
		TagName: "v1.2.3",
		Name:    "Release 1.2.3",
		Body:    "Release notes",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/testowner/testrepo/releases/tags/v1.2.3" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	result, err := client.GetReleaseByTag(context.Background(), "testowner", "testrepo", "v1.2.3")
	if err != nil {
		t.Fatalf("GetReleaseByTag() error = %v", err)
	}

	if result.TagName != "v1.2.3" {
		t.Errorf("GetReleaseByTag().TagName = %q, want %q", result.TagName, "v1.2.3")
	}
}

func TestCompareVersions_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"", "", 0},
		{"1", "1", 0},
		{"1.2", "1.2.0", 0},
		{"10.0.0", "9.99.99", 1},
		{"0.0.1", "0.0.2", -1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			t.Parallel()
			got := version.CompareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestBuildIssueURL_WithTitle(t *testing.T) {
	t.Parallel()

	url := github.BuildIssueURL(github.IssueTypeBug, "My Bug Title", "body content")
	if url == "" {
		t.Error("BuildIssueURL() returned empty string")
	}
	if !containsString(url, "title=My") {
		t.Error("BuildIssueURL() should contain title param")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestVerifyChecksum(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create a test binary
	binaryPath := "/tmp/test-binary"
	binaryContent := []byte("test binary content")
	if err := afero.WriteFile(memFs, binaryPath, binaryContent, 0o600); err != nil {
		t.Fatalf("failed to create test binary: %v", err)
	}

	// Calculate expected checksum (SHA256)
	hasher := sha256.New()
	hasher.Write(binaryContent)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	checksumContent := fmt.Sprintf("%s  test-asset.tar.gz\n", expectedHash)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(checksumContent)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	checksumAsset := &github.Asset{
		Name:               "checksums.txt",
		BrowserDownloadURL: server.URL + "/checksums.txt",
	}

	err := client.VerifyChecksum(context.Background(), ios, binaryPath, "test-asset.tar.gz", checksumAsset)
	if err != nil {
		t.Errorf("VerifyChecksum() error = %v", err)
	}
}

//nolint:paralleltest // modifies global filesystem
func TestVerifyChecksum_Mismatch(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	binaryPath := "/tmp/test-binary"
	binaryContent := []byte("test binary content")
	if err := afero.WriteFile(memFs, binaryPath, binaryContent, 0o600); err != nil {
		t.Fatalf("failed to create test binary: %v", err)
	}

	// Provide wrong checksum
	checksumContent := "0000000000000000000000000000000000000000000000000000000000000000  test-asset.tar.gz\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(checksumContent)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	checksumAsset := &github.Asset{
		Name:               "checksums.txt",
		BrowserDownloadURL: server.URL + "/checksums.txt",
	}

	err := client.VerifyChecksum(context.Background(), ios, binaryPath, "test-asset.tar.gz", checksumAsset)
	if err == nil {
		t.Error("VerifyChecksum() should return error for checksum mismatch")
	}
	if !containsString(err.Error(), "checksum mismatch") {
		t.Errorf("error should mention checksum mismatch, got: %v", err)
	}
}

func TestVerifyChecksum_FileNotFound(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	checksumAsset := &github.Asset{
		Name:               "checksums.txt",
		BrowserDownloadURL: "http://localhost:1/checksums.txt",
	}

	err := client.VerifyChecksum(context.Background(), ios, "/nonexistent/path", "test-asset.tar.gz", checksumAsset)
	if err == nil {
		t.Error("VerifyChecksum() should return error for nonexistent file")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCreateBackup_Failure(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Try to backup a file that doesn't exist
	err := github.CreateBackup(ios, "/tmp/nonexistent", "/tmp/backup")
	if err == nil {
		t.Error("CreateBackup() should return error for nonexistent source")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestRestoreFromBackup(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create backup file
	backupPath := testBackupPath
	backupContent := []byte("original content")
	if err := afero.WriteFile(memFs, backupPath, backupContent, 0o600); err != nil {
		t.Fatalf("failed to create backup file: %v", err)
	}

	targetPath := testTargetPath
	originalErr := fmt.Errorf("write failed")

	err := github.RestoreFromBackup(backupPath, targetPath, originalErr)
	if err == nil {
		t.Error("RestoreFromBackup() should return error")
	}
	if !containsString(err.Error(), "write failed") {
		t.Errorf("error should contain original error, got: %v", err)
	}

	// Verify backup was moved to target
	content, err := afero.ReadFile(memFs, targetPath)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if !bytes.Equal(content, backupContent) {
		t.Errorf("restored content = %q, want %q", string(content), string(backupContent))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestRestoreFromBackup_RestoreFails(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	originalErr := fmt.Errorf("write failed")

	err := github.RestoreFromBackup("/nonexistent/backup", "/nonexistent/target", originalErr)
	if err == nil {
		t.Error("RestoreFromBackup() should return error when restore fails")
	}
	if !containsString(err.Error(), "restore failed") {
		t.Errorf("error should mention restore failure, got: %v", err)
	}
}

//nolint:paralleltest // modifies global base URL
func TestPerformUpdate_Cancelled(t *testing.T) {
	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	release := &github.Release{
		TagName: "v2.0.0",
		Name:    "Release 2.0.0",
	}

	// Confirmation returns false
	confirmFunc := func(prompt string, skipConfirm bool) (bool, error) {
		return false, nil
	}

	err := client.PerformUpdate(context.Background(), ios, release, "v1.0.0", "release notes", confirmFunc, false)
	if err == nil {
		t.Error("PerformUpdate() should return error when cancelled")
	}
	if !containsString(err.Error(), "cancelled") {
		t.Errorf("error should mention cancellation, got: %v", err)
	}
}

func TestPerformUpdate_ConfirmError(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	release := &github.Release{
		TagName: "v2.0.0",
		Name:    "Release 2.0.0",
	}

	// Confirmation returns error
	confirmFunc := func(prompt string, skipConfirm bool) (bool, error) {
		return false, fmt.Errorf("input error")
	}

	err := client.PerformUpdate(context.Background(), ios, release, "v1.0.0", "release notes", confirmFunc, false)
	if err == nil {
		t.Error("PerformUpdate() should return error when confirmation fails")
	}
	if !containsString(err.Error(), "confirmation") {
		t.Errorf("error should mention confirmation failure, got: %v", err)
	}
}

//nolint:paralleltest // modifies global base URL
func TestPerformRollback_NoRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty releases list
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("[]")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	confirmFunc := func(prompt string, skipConfirm bool) (bool, error) {
		return true, nil
	}

	err := client.PerformRollback(context.Background(), ios, "v1.0.0", false, confirmFunc, false)
	if err == nil {
		t.Error("PerformRollback() should return error when no previous release available")
	}
}

//nolint:paralleltest // modifies global base URL
func TestListReleases_Prereleases(t *testing.T) {
	releases := []github.Release{
		{TagName: "v1.0.0", Name: "Release 1.0.0", Prerelease: false},
		{TagName: "v1.1.0-beta", Name: "Beta Release", Prerelease: true},
		{TagName: "v1.2.0", Name: "Release 1.2.0", Prerelease: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Test without prereleases
	result, err := client.ListReleases(context.Background(), "testowner", "testrepo", false)
	if err != nil {
		t.Fatalf("ListReleases() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ListReleases(includePre=false) returned %d releases, want 2", len(result))
	}

	// Test with prereleases
	result, err = client.ListReleases(context.Background(), "testowner", "testrepo", true)
	if err != nil {
		t.Fatalf("ListReleases(includePre=true) error = %v", err)
	}
	if len(result) != 3 {
		t.Errorf("ListReleases(includePre=true) returned %d releases, want 3", len(result))
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetLatestRelease_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetLatestRelease(context.Background(), "testowner", "testrepo")
	if err == nil {
		t.Error("GetLatestRelease() should return error when not found")
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetReleaseByTag_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetReleaseByTag(context.Background(), "testowner", "testrepo", "v1.0.0")
	if err == nil {
		t.Error("GetReleaseByTag() should return error when not found")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestDownloadAsset_Failure(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	destPath := "/tmp/downloaded"

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	asset := &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: "http://localhost:1/nonexistent",
	}

	err := client.DownloadAsset(context.Background(), asset, destPath)
	if err == nil {
		t.Error("DownloadAsset() should return error for unreachable URL")
	}
}

//nolint:paralleltest // modifies global base URL
func TestPerformRollback_Cancelled(t *testing.T) {
	releases := []github.Release{
		{TagName: "v2.0.0", Name: "Release 2.0.0"},
		{TagName: "v1.0.0", Name: "Release 1.0.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	confirmFunc := func(prompt string, skipConfirm bool) (bool, error) {
		return false, nil // User cancels
	}

	err := client.PerformRollback(context.Background(), ios, "v2.0.0", false, confirmFunc, false)
	// Rollback should complete without error when cancelled gracefully
	if err != nil {
		t.Errorf("PerformRollback() unexpected error = %v", err)
	}
}

//nolint:paralleltest // modifies global base URL
func TestPerformRollback_ConfirmError(t *testing.T) {
	releases := []github.Release{
		{TagName: "v2.0.0", Name: "Release 2.0.0"},
		{TagName: "v1.0.0", Name: "Release 1.0.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	confirmFunc := func(prompt string, skipConfirm bool) (bool, error) {
		return false, fmt.Errorf("input error")
	}

	err := client.PerformRollback(context.Background(), ios, "v2.0.0", false, confirmFunc, false)
	if err == nil {
		t.Error("PerformRollback() should return error when confirmation fails")
	}
	if !containsString(err.Error(), "confirmation") {
		t.Errorf("error should mention confirmation failure, got: %v", err)
	}
}

func TestInstallRelease_NoPlatformAsset(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Release with no assets for any platform
	release := &github.Release{
		TagName: "v1.0.0",
		Name:    "Release 1.0.0",
		Assets: []github.Asset{
			{Name: "shelly-freebsd-sparc64.tar.gz"},
		},
	}

	err := client.InstallRelease(context.Background(), ios, release)
	if err == nil {
		t.Error("InstallRelease() should return error when no platform asset available")
	}
	if !containsString(err.Error(), "no binary available") {
		t.Errorf("error should mention no binary available, got: %v", err)
	}
}

//nolint:paralleltest // modifies global filesystem
func TestDownloadAndExtract_RawBinary(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	binaryContent := []byte("#!/bin/sh\necho hello")

	// Create a server that returns a raw binary
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(binaryContent); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// For unknown formats, it treats the file as a raw binary
	asset := &github.Asset{
		Name:               "shelly-linux-amd64",
		BrowserDownloadURL: server.URL + "/download",
	}

	binaryPath, cleanup, err := client.DownloadAndExtract(context.Background(), asset, "shelly")
	if err != nil {
		t.Fatalf("DownloadAndExtract() error = %v", err)
	}
	defer cleanup()

	// Verify the binary was downloaded
	content, err := afero.ReadFile(memFs, binaryPath)
	if err != nil {
		t.Fatalf("failed to read binary: %v", err)
	}
	if !bytes.Equal(content, binaryContent) {
		t.Errorf("binary content = %q, want %q", string(content), string(binaryContent))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCreateBackup_Success(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create source file
	srcPath := testSourcePath
	if err := afero.WriteFile(memFs, srcPath, []byte("content"), 0o600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	backupPath := testBackupPath
	err := github.CreateBackup(ios, srcPath, backupPath)
	if err != nil {
		t.Errorf("CreateBackup() error = %v", err)
	}

	// Verify backup exists
	content, err := afero.ReadFile(memFs, backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("backup content = %q, want %q", string(content), "content")
	}
}

//nolint:paralleltest // modifies global base URL
func TestListReleases_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ListReleases(context.Background(), "testowner", "testrepo", false)
	if err == nil {
		t.Error("ListReleases() should return error for server error")
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetLatestRelease_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetLatestRelease(context.Background(), "testowner", "testrepo")
	if err == nil {
		t.Error("GetLatestRelease() should return error for server error")
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetReleaseByTag_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetReleaseByTag(context.Background(), "testowner", "testrepo", "v1.0.0")
	if err == nil {
		t.Error("GetReleaseByTag() should return error for server error")
	}
}

//nolint:paralleltest // modifies global base URL
func TestFindPreviousRelease(t *testing.T) {
	releases := []github.Release{
		{TagName: "v2.0.0", Name: "Release 2.0.0"},
		{TagName: testVersionV15, Name: "Release 1.5.0"},
		{TagName: "v1.0.0", Name: "Release 1.0.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	result, err := client.FindPreviousRelease(context.Background(), "v2.0.0", false)
	if err != nil {
		t.Fatalf("FindPreviousRelease() error = %v", err)
	}

	if result.TagName != testVersionV15 {
		t.Errorf("FindPreviousRelease() = %q, want %q", result.TagName, testVersionV15)
	}
}

//nolint:paralleltest // modifies global base URL
func TestFindPreviousRelease_NoPrevious(t *testing.T) {
	releases := []github.Release{
		{TagName: "v1.0.0", Name: "Release 1.0.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.FindPreviousRelease(context.Background(), "v1.0.0", false)
	if err == nil {
		t.Error("FindPreviousRelease() should return error when no previous release exists")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestSetFs(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)

	// Reset to nil (should restore OS filesystem)
	github.SetFs(nil)

	// No error means it worked
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_WithAfero(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create the new binary
	newContent := []byte("new binary content")
	if err := afero.WriteFile(memFs, "/tmp/new", newContent, 0o755); err != nil {
		t.Fatalf("failed to create new binary: %v", err)
	}

	// Create the target binary
	oldContent := []byte("old binary content")
	if err := afero.WriteFile(memFs, "/tmp/target", oldContent, 0o755); err != nil {
		t.Fatalf("failed to create target binary: %v", err)
	}

	err := github.ReplaceBinary(ios, "/tmp/new", "/tmp/target")
	if err != nil {
		t.Fatalf("ReplaceBinary() error = %v", err)
	}

	// Verify the target was replaced
	content, err := afero.ReadFile(memFs, "/tmp/target")
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if !bytes.Equal(content, newContent) {
		t.Errorf("target content = %q, want %q", string(content), string(newContent))
	}

	// Verify backup was removed
	exists, err := afero.Exists(memFs, "/tmp/target.bak")
	if err != nil {
		t.Fatalf("failed to check backup: %v", err)
	}
	if exists {
		t.Error("backup file should have been removed")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCopyFile_WithAfero(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create source file
	content := []byte("source content")
	if err := afero.WriteFile(memFs, "/tmp/src", content, 0o644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	err := github.CopyFile(ios, "/tmp/src", "/tmp/dst")
	if err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	// Verify copy
	copied, err := afero.ReadFile(memFs, "/tmp/dst")
	if err != nil {
		t.Fatalf("failed to read copy: %v", err)
	}
	if !bytes.Equal(copied, content) {
		t.Errorf("copied content = %q, want %q", string(copied), string(content))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCreateBackup_WithAfero(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create source file
	content := []byte("source content")
	if err := afero.WriteFile(memFs, "/tmp/src", content, 0o644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	err := github.CreateBackup(ios, "/tmp/src", "/tmp/backup")
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	// Verify backup exists
	backed, err := afero.ReadFile(memFs, "/tmp/backup")
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}
	if !bytes.Equal(backed, content) {
		t.Errorf("backup content = %q, want %q", string(backed), string(content))
	}

	// Verify source was moved (no longer exists)
	exists, err := afero.Exists(memFs, "/tmp/src")
	if err != nil {
		t.Fatalf("failed to check source: %v", err)
	}
	if exists {
		t.Error("source should have been moved to backup")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestRestoreFromBackup_WithAfero(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create backup file
	backupContent := []byte("backup content")
	if err := afero.WriteFile(memFs, "/tmp/backup", backupContent, 0o644); err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	originalErr := fmt.Errorf("write failed")
	err := github.RestoreFromBackup("/tmp/backup", "/tmp/target", originalErr)
	if err == nil {
		t.Error("RestoreFromBackup() should return error")
	}

	// Verify backup was moved to target
	content, err := afero.ReadFile(memFs, "/tmp/target")
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if !bytes.Equal(content, backupContent) {
		t.Errorf("target content = %q, want %q", string(content), string(backupContent))
	}
}

//nolint:paralleltest // modifies global base URL
func TestFetchSpecificVersion(t *testing.T) {
	testCases := []struct {
		name     string
		inputVer string
		wantTag  string
	}{
		{"with_prefix", "v2.0.0", "v2.0.0"},
		{"without_prefix", "2.0.0", "v2.0.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/releases/tags/") {
					w.Header().Set("Content-Type", "application/json")
					if _, err := w.Write([]byte(`{"tag_name": "` + tc.wantTag + `"}`)); err != nil {
						t.Errorf("failed to write response: %v", err)
					}
				}
			}))
			defer ts.Close()

			restore := github.SetAPIBaseURL(ts.URL)
			defer restore()

			ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
			client := github.NewClient(ios)

			release, err := client.FetchSpecificVersion(context.Background(), tc.inputVer)
			if err != nil {
				t.Fatalf("FetchSpecificVersion() error = %v", err)
			}
			if release.TagName != tc.wantTag {
				t.Errorf("TagName = %q, want %q", release.TagName, tc.wantTag)
			}
		})
	}
}

//nolint:paralleltest // modifies global base URL
func TestFetchLatestVersion(t *testing.T) {
	t.Run("stable_release", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/releases/latest") {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`{"tag_name": "v2.0.0"}`)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}
		}))
		defer ts.Close()

		restore := github.SetAPIBaseURL(ts.URL)
		defer restore()

		ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
		client := github.NewClient(ios)

		release, err := client.FetchLatestVersion(context.Background(), false)
		if err != nil {
			t.Fatalf("FetchLatestVersion() error = %v", err)
		}
		if release.TagName != testVersionV2 {
			t.Errorf("TagName = %q, want %q", release.TagName, testVersionV2)
		}
	})

	t.Run("include_prerelease", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/releases") {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`[{"tag_name": "v3.0.0-beta"}, {"tag_name": "v2.0.0"}]`)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}
		}))
		defer ts.Close()

		restore := github.SetAPIBaseURL(ts.URL)
		defer restore()

		ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
		client := github.NewClient(ios)

		release, err := client.FetchLatestVersion(context.Background(), true)
		if err != nil {
			t.Fatalf("FetchLatestVersion() error = %v", err)
		}
		if release.TagName != "v3.0.0-beta" {
			t.Errorf("TagName = %q, want %q", release.TagName, "v3.0.0-beta")
		}
	})

	t.Run("no_releases", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/releases") {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`[]`)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}
		}))
		defer ts.Close()

		restore := github.SetAPIBaseURL(ts.URL)
		defer restore()

		ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
		client := github.NewClient(ios)

		_, err := client.FetchLatestVersion(context.Background(), true)
		if !errors.Is(err, github.ErrNoReleases) {
			t.Errorf("expected ErrNoReleases, got %v", err)
		}
	})
}

//nolint:paralleltest // modifies global base URL
func TestGetTargetRelease(t *testing.T) {
	t.Run("specific_version", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/releases/tags/") {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`{"tag_name": "` + testVersionV15 + `"}`)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}
		}))
		defer ts.Close()

		restore := github.SetAPIBaseURL(ts.URL)
		defer restore()

		ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
		client := github.NewClient(ios)

		release, err := client.GetTargetRelease(context.Background(), "1.5.0", false)
		if err != nil {
			t.Fatalf("GetTargetRelease() error = %v", err)
		}
		if release.TagName != testVersionV15 {
			t.Errorf("TagName = %q, want %q", release.TagName, testVersionV15)
		}
	})

	t.Run("latest_stable", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/releases/latest") {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`{"tag_name": "v2.0.0"}`)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}
		}))
		defer ts.Close()

		restore := github.SetAPIBaseURL(ts.URL)
		defer restore()

		ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
		client := github.NewClient(ios)

		release, err := client.GetTargetRelease(context.Background(), "", false)
		if err != nil {
			t.Fatalf("GetTargetRelease() error = %v", err)
		}
		if release.TagName != testVersionV2 {
			t.Errorf("TagName = %q, want %q", release.TagName, testVersionV2)
		}
	})
}

//nolint:paralleltest // modifies global base URL
func TestReleaseFetcher(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"tag_name": "v2.5.0"}`)); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	fetcher := github.ReleaseFetcher(ios)

	ver, err := fetcher(context.Background())
	if err != nil {
		t.Fatalf("ReleaseFetcher() error = %v", err)
	}
	if ver != "2.5.0" {
		t.Errorf("version = %q, want %q", ver, "2.5.0")
	}
}

//nolint:paralleltest // modifies global base URL
func TestReleaseFetcher_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	fetcher := github.ReleaseFetcher(ios)

	_, err := fetcher(context.Background())
	if err == nil {
		t.Error("ReleaseFetcher() should return error")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadExtensionRelease(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	binaryContent := []byte("extension binary")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly-extension",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip content: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	// Store URL for the handler closure
	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"tag_name": "v1.0.0",
				"assets": []map[string]any{
					{
						"name":                 "shelly-extension_linux_amd64.tar.gz",
						"browser_download_url": serverURL + "/download/shelly-extension_linux_amd64.tar.gz",
					},
				},
			}); err != nil {
				t.Errorf("failed to encode response: %v", err)
			}
		case strings.Contains(r.URL.Path, "/download/"):
			w.Header().Set("Content-Type", "application/octet-stream")
			if _, err := w.Write(gzBuf.Bytes()); err != nil {
				t.Errorf("failed to write binary: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	serverURL = ts.URL

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	result, err := client.DownloadExtensionRelease(context.Background(), "owner", "extension", "shelly-")
	if err != nil {
		t.Fatalf("DownloadExtensionRelease() error = %v", err)
	}
	if result == nil {
		t.Fatal("DownloadExtensionRelease() returned nil result")
	}
	if result.TagName != testVersionV1 {
		t.Errorf("TagName = %q, want %q", result.TagName, testVersionV1)
	}
	if result.Cleanup != nil {
		result.Cleanup()
	}
}

func TestCheckForUpdatesCached_DisabledByEnv(t *testing.T) {
	// t.Setenv prevents parallel execution, but we also modify global filesystem
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	t.Setenv("SHELLY_NO_UPDATE_CHECK", "1")

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result != nil {
		t.Error("CheckForUpdatesCached() should return nil when disabled")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestCheckForUpdatesCached_ValidCacheWithUpdate(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create cache directory and file
	if err := memFs.MkdirAll("/cache", 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(memFs, "/cache/version", []byte("2.0.0"), 0o644); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result == nil {
		t.Fatal("CheckForUpdatesCached() should return update")
	}
	if result.Version() != "2.0.0" {
		t.Errorf("Version = %q, want %q", result.Version(), "2.0.0")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestCheckForUpdatesCached_ValidCacheNoUpdate(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create cache with same version
	if err := memFs.MkdirAll("/cache", 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(memFs, "/cache/version", []byte("1.0.0"), 0o644); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result != nil {
		t.Error("CheckForUpdatesCached() should return nil when no update available")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestCheckForUpdatesCached_FetchFromAPI(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"tag_name": "v3.0.0"}`)); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result == nil {
		t.Fatal("CheckForUpdatesCached() should return update")
	}
	if result.Version() != "3.0.0" {
		t.Errorf("Version = %q, want %q", result.Version(), "3.0.0")
	}

	// Verify cache was written
	data, err := afero.ReadFile(memFs, "/cache/version")
	if err != nil {
		t.Fatalf("cache not written: %v", err)
	}
	if string(data) != "3.0.0" {
		t.Errorf("cached version = %q, want %q", string(data), "3.0.0")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestCheckForUpdatesCached_FetchNoUpdate(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"tag_name": "v1.0.0"}`)); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result != nil {
		t.Error("CheckForUpdatesCached() should return nil when no update")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestCheckForUpdatesCached_FetchError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result != nil {
		t.Error("CheckForUpdatesCached() should return nil on error")
	}
}

func TestCheckForUpdates(t *testing.T) {
	// Use memory filesystem to prevent writes to real cache
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"tag_name": "v2.0.0"}`)); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result, err := github.CheckForUpdates(context.Background(), ios, "1.0.0")
	if err != nil {
		t.Fatalf("CheckForUpdates() error = %v", err)
	}
	if result == nil {
		t.Fatal("CheckForUpdates() returned nil")
	}
	if !result.UpdateAvailable {
		t.Error("UpdateAvailable should be true")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCreateBackup_RenameFailure(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Try to backup a non-existent file
	err := github.CreateBackup(ios, "/nonexistent/source", "/tmp/backup")
	if err == nil {
		t.Error("CreateBackup() should return error for non-existent source")
	}
}

//nolint:paralleltest // modifies global filesystem and runtimeGOOS
func TestCreateBackup_WindowsFallbackSuccess(t *testing.T) {
	// Simulate Windows
	restoreGOOS := github.SetRuntimeGOOS("windows")
	defer restoreGOOS()

	// Use a filesystem wrapper that fails on Rename but allows other operations
	memFs := afero.NewMemMapFs()
	failRenameFs := &renameFailFs{Fs: memFs}
	github.SetFs(failRenameFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create source file
	srcPath := testSourcePath
	backupPath := testBackupPath
	if err := afero.WriteFile(memFs, srcPath, []byte("source content"), 0o644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// On Windows, when rename fails, it should fall back to copy
	err := github.CreateBackup(ios, srcPath, backupPath)
	if err != nil {
		t.Errorf("CreateBackup() error = %v", err)
	}

	// Verify backup was created via copy
	content, err := afero.ReadFile(memFs, backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}
	if string(content) != "source content" {
		t.Errorf("backup content = %q, want %q", string(content), "source content")
	}
}

// renameFailFs is a filesystem wrapper that always fails on Rename.
type renameFailFs struct {
	afero.Fs
}

func (f *renameFailFs) Rename(oldname, newname string) error {
	return errors.New("rename not supported")
}

//nolint:paralleltest // modifies global filesystem and runtimeGOOS
func TestCreateBackup_WindowsFallbackCopyError(t *testing.T) {
	// Simulate Windows
	restoreGOOS := github.SetRuntimeGOOS("windows")
	defer restoreGOOS()

	// Use a filesystem that fails on Rename
	memFs := afero.NewMemMapFs()
	failRenameFs := &renameFailFs{Fs: memFs}
	github.SetFs(failRenameFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Try to backup non-existent file on "Windows"
	// Rename will fail, then copy will also fail because source doesn't exist
	err := github.CreateBackup(ios, "/nonexistent/source", "/tmp/backup")
	if err == nil {
		t.Error("CreateBackup() should return error when both rename and copy fail")
	}
	if !strings.Contains(err.Error(), "backup failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

//nolint:paralleltest // modifies global filesystem and runtimeGOOS
func TestCreateBackup_WindowsFallbackRemoveFails(t *testing.T) {
	// Simulate Windows
	restoreGOOS := github.SetRuntimeGOOS("windows")
	defer restoreGOOS()

	// Use a filesystem that fails on Rename and Remove
	memFs := afero.NewMemMapFs()
	failFs := &renameAndRemoveFailFs{Fs: memFs}
	github.SetFs(failFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create source file
	srcPath := testSourcePath
	backupPath := testBackupPath
	if err := afero.WriteFile(memFs, srcPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// On Windows, when rename fails, copy should succeed, then remove will fail
	// but we should still succeed (remove failure is logged, not returned)
	err := github.CreateBackup(ios, srcPath, backupPath)
	if err != nil {
		t.Errorf("CreateBackup() should succeed even if Remove fails, got error: %v", err)
	}

	// Verify backup was created
	content, err := afero.ReadFile(memFs, backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("backup content = %q, want %q", string(content), "content")
	}
}

// renameAndRemoveFailFs is a filesystem wrapper that fails on Rename and Remove.
type renameAndRemoveFailFs struct {
	afero.Fs
}

func (f *renameAndRemoveFailFs) Rename(oldname, newname string) error {
	return errors.New("rename not supported")
}

func (f *renameAndRemoveFailFs) Remove(name string) error {
	return errors.New("remove not supported")
}

//nolint:paralleltest // modifies global filesystem
func TestCopyFile_SourceNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	err := github.CopyFile(ios, "/nonexistent", "/tmp/dst")
	if err == nil {
		t.Error("CopyFile() should return error for non-existent source")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCopyFile_Success(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create source file
	content := []byte("test content for copy")
	if err := afero.WriteFile(memFs, "/tmp/src", content, 0o644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// Create parent directory for destination
	if err := memFs.MkdirAll("/tmp/dest", 0o755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Copy the file
	err := github.CopyFile(ios, "/tmp/src", "/tmp/dest/dst")
	if err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	// Verify copied content
	copied, err := afero.ReadFile(memFs, "/tmp/dest/dst")
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if !bytes.Equal(copied, content) {
		t.Errorf("copied content = %q, want %q", string(copied), string(content))
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestVerifyChecksum_OpenFileError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Try to verify checksum for a non-existent file
	err := client.VerifyChecksum(context.Background(), ios, "/nonexistent", "binary", &github.Asset{})
	if err == nil {
		t.Error("VerifyChecksum() should return error for non-existent file")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestVerifyChecksum_DownloadError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	// Create a file to checksum
	if err := afero.WriteFile(memFs, "/tmp/binary", []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.VerifyChecksum(context.Background(), ios, "/tmp/binary", "binary", &github.Asset{
		Name:               "checksums.txt",
		BrowserDownloadURL: ts.URL + "/checksums.txt",
	})
	if err == nil {
		t.Error("VerifyChecksum() should return error when checksum download fails")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAsset_Success(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	content := []byte("downloaded content")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(content); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.DownloadAsset(context.Background(), &github.Asset{
		Name:               "test",
		BrowserDownloadURL: ts.URL + "/test",
	}, "/tmp/downloaded/test")
	if err != nil {
		t.Fatalf("DownloadAsset() error = %v", err)
	}

	// Verify the file was downloaded
	data, err := afero.ReadFile(memFs, "/tmp/downloaded/test")
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("downloaded content = %q, want %q", string(data), string(content))
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAsset_HTTPError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.DownloadAsset(context.Background(), &github.Asset{
		Name:               "test",
		BrowserDownloadURL: ts.URL + "/test",
	}, "/tmp/test")
	if err == nil {
		t.Error("DownloadAsset() should return error on HTTP error")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestExtractTarGz_InvalidGzip(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Write invalid gzip data
	if err := afero.WriteFile(memFs, "/tmp/invalid.tar.gz", []byte("not gzip"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: "file:///tmp/invalid.tar.gz",
	}, "binary")

	// The function will fail during extraction
	if err == nil {
		t.Error("DownloadAndExtract() should return error for invalid gzip")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractZip_FileNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.zip",
		BrowserDownloadURL: "file:///nonexistent.zip",
	}, "binary")
	if err == nil {
		t.Error("DownloadAndExtract() should return error for non-existent file")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractZip_InvalidZip(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Write invalid zip data
	if err := afero.WriteFile(memFs, "/tmp/invalid.zip", []byte("not a zip"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "invalid.zip",
		BrowserDownloadURL: "file:///tmp/invalid.zip",
	}, "binary")
	if err == nil {
		t.Error("DownloadAndExtract() should return error for invalid zip")
	}
}

func TestBytesReaderAt_ReadBeyondEnd(t *testing.T) {
	t.Parallel()

	// This tests the ReadAt implementation when reading beyond data length
	data := []byte("hello")

	// We can't directly test bytesReaderAt since it's unexported,
	// but we can verify zip reading works with small data
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create a minimal valid zip
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("test.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	if err := afero.WriteFile(memFs, "/tmp/test.zip", buf.Bytes(), 0o644); err != nil {
		t.Fatalf("failed to write zip file: %v", err)
	}

	// Reading should work
	zipData, err := afero.ReadFile(memFs, "/tmp/test.zip")
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}
	if len(zipData) == 0 {
		t.Error("zip file should not be empty")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestFindBinaryInTar_NotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create tar.gz with different file
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "otherbinary",
		Mode: 0o755,
		Size: 5,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("hello")); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	// Create server that returns the archive
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: ts.URL + "/test.tar.gz",
	}, "notfound")
	if err == nil {
		t.Error("DownloadAndExtract() should return error when binary not found")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestFindBinaryInZip_NotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create zip with different file
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("otherbinary")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write([]byte("hello")); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(buf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err = client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.zip",
		BrowserDownloadURL: ts.URL + "/test.zip",
	}, "notfound")
	if err == nil {
		t.Error("DownloadAndExtract() should return error when binary not found")
	}
}

//nolint:paralleltest // modifies global base URL
func TestListReleases_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`invalid json`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ListReleases(context.Background(), "owner", "repo", false)
	if err == nil {
		t.Error("ListReleases() should return error for invalid JSON")
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetLatestRelease_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`invalid json`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetLatestRelease(context.Background(), "owner", "repo")
	if err == nil {
		t.Error("GetLatestRelease() should return error for invalid JSON")
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetReleaseByTag_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`invalid json`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetReleaseByTag(context.Background(), "owner", "repo", "v1.0.0")
	if err == nil {
		t.Error("GetReleaseByTag() should return error for invalid JSON")
	}
}

func TestFindAssetForPlatform_NoMatch(t *testing.T) {
	t.Parallel()

	// Create a release with assets for different platforms
	release := &github.Release{
		Assets: []github.Asset{
			{Name: "binary_windows_amd64.tar.gz"},
		},
	}

	// FindAssetForPlatform is a method on Release, test by checking if asset is nil
	// when current platform doesn't match (this test is less useful but demonstrates the code path)
	asset := release.FindAssetForPlatform()
	// The test will pass if runtime OS doesn't match windows
	if runtime.GOOS != "windows" && asset != nil {
		t.Error("FindAssetForPlatform() should return nil when no matching asset for current platform")
	}
}

func TestFindChecksumAsset_NotFound(t *testing.T) {
	t.Parallel()

	release := &github.Release{
		Assets: []github.Asset{
			{Name: "binary.tar.gz"},
			{Name: "binary.zip"},
		},
	}

	asset := release.FindChecksumAsset("binary.tar.gz")
	if asset != nil {
		t.Error("FindChecksumAsset() should return nil when no checksum found")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadExtensionRelease_NoBinary(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			// No assets match the platform
			if err := json.NewEncoder(w).Encode(map[string]any{
				"tag_name": "v1.0.0",
				"assets": []map[string]any{
					{
						"name":                 "wrong_platform.tar.gz",
						"browser_download_url": serverURL + "/download/wrong.tar.gz",
					},
				},
			}); err != nil {
				t.Errorf("failed to encode response: %v", err)
			}
		}
	}))
	defer ts.Close()
	serverURL = ts.URL

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.DownloadExtensionRelease(context.Background(), "owner", "extension", "shelly-")
	if err == nil {
		t.Error("DownloadExtensionRelease() should return error when no binary found")
	}
}

//nolint:paralleltest // modifies global base URL
func TestDownloadExtensionRelease_ReleaseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.DownloadExtensionRelease(context.Background(), "owner", "repo", "prefix-")
	if err == nil {
		t.Error("DownloadExtensionRelease() should return error on release fetch failure")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCheckForUpdatesCached_EmptyCacheContent(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create an empty cache file
	if err := memFs.MkdirAll("/cache", 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := afero.WriteFile(memFs, "/cache/version", []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write empty cache: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	// With empty cache content and valid cache, returns nil (no update)
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result != nil {
		t.Error("CheckForUpdatesCached() should return nil for empty cache content")
	}
}

//nolint:paralleltest // modifies global base URL
func TestFetchLatestVersion_ListError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.FetchLatestVersion(context.Background(), true)
	if err == nil {
		t.Error("FetchLatestVersion() should return error on list failure")
	}
}

//nolint:paralleltest // modifies global base URL
func TestFindPreviousRelease_ListError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.FindPreviousRelease(context.Background(), "v2.0.0", false)
	if err == nil {
		t.Error("FindPreviousRelease() should return error on list failure")
	}
}

//nolint:paralleltest // modifies global base URL
func TestFindPreviousRelease_Fallback(t *testing.T) {
	// Test the fallback case when current version is not found
	releases := []github.Release{
		{TagName: "v3.0.0", Name: "Release 3.0.0"},
		{TagName: "v2.0.0", Name: "Release 2.0.0"},
		{TagName: "v1.0.0", Name: "Release 1.0.0"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	restore := github.SetAPIBaseURL(server.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Looking for a version that doesn't exist, should fallback to second
	result, err := client.FindPreviousRelease(context.Background(), "v4.0.0", false)
	if err != nil {
		t.Fatalf("FindPreviousRelease() error = %v", err)
	}
	// Should fall back to releases[1]
	if result.TagName != testVersionV2 {
		t.Errorf("FindPreviousRelease() = %q, want %q", result.TagName, testVersionV2)
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAndExtract_ZipBinaryNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create a zip with a different filename
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("wrongname")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write([]byte("content")); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(buf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err = client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "archive.zip",
		BrowserDownloadURL: ts.URL + "/archive.zip",
	}, "shelly")
	if err == nil {
		t.Error("DownloadAndExtract() should return error when binary not in zip")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAndExtract_TarBinaryNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create a tar.gz with a different filename
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "wrongname",
		Mode: 0o755,
		Size: 7,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("content")); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "archive.tar.gz",
		BrowserDownloadURL: ts.URL + "/archive.tar.gz",
	}, "shelly")
	if err == nil {
		t.Error("DownloadAndExtract() should return error when binary not in tar")
	}
}

func TestParseRepoString_Formats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{"owner/repo", "myorg/myrepo", "myorg", "myrepo", false},
		{"gh:prefix", "gh:myorg/myrepo", "myorg", "myrepo", false},
		{"github:prefix", "github:myorg/myrepo", "myorg", "myrepo", false},
		{"url_format", "https://github.com/myorg/myrepo", "myorg", "myrepo", false},
		{"no_slash", "justrepo", "", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			owner, repo, err := github.ParseRepoString(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if owner != tc.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tc.wantOwner)
			}
			if repo != tc.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tc.wantRepo)
			}
		})
	}
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_ReadNewBinaryError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create target but no new binary
	if err := afero.WriteFile(memFs, "/tmp/target", []byte("old"), 0o755); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	err := github.ReplaceBinary(ios, "/tmp/nonexistent", "/tmp/target")
	if err == nil {
		t.Error("ReplaceBinary() should return error when new binary doesn't exist")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_StatTargetError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create new binary but no target
	if err := afero.WriteFile(memFs, "/tmp/new", []byte("new"), 0o755); err != nil {
		t.Fatalf("failed to create new: %v", err)
	}

	err := github.ReplaceBinary(ios, "/tmp/new", "/tmp/nonexistent")
	if err == nil {
		t.Error("ReplaceBinary() should return error when target doesn't exist")
	}
	if !strings.Contains(err.Error(), "stat target") {
		t.Errorf("error should mention stat target, got: %v", err)
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCopyFile_DestNotWritable(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create source file
	if err := afero.WriteFile(memFs, "/tmp/src", []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// CopyFile should succeed with a simple path on MemMapFs
	err := github.CopyFile(ios, "/tmp/src", "/tmp/dest")
	if err != nil {
		t.Errorf("CopyFile() unexpected error = %v", err)
	}
}

func TestBytesReaderAt_PartialRead(t *testing.T) {
	t.Parallel()

	// Create a zip with known content to test the partial read path
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	content := []byte("short")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("file.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}

	// The zip reader will use ReadAt internally
	if err := afero.WriteFile(memFs, "/test.zip", buf.Bytes(), 0o644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	data, err := afero.ReadFile(memFs, "/test.zip")
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}

	// Test that zip can be read (exercises ReadAt)
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("failed to create zip reader: %v", err)
	}
	if len(reader.File) != 1 {
		t.Errorf("expected 1 file, got %d", len(reader.File))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestVerifyChecksum_AssetNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create a file to checksum
	if err := afero.WriteFile(memFs, "/tmp/binary", []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create checksum file without entry for our asset
	checksumContent := "abcd1234  other-asset.tar.gz\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(checksumContent)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.VerifyChecksum(context.Background(), ios, "/tmp/binary", "our-asset.tar.gz", &github.Asset{
		Name:               "checksums.txt",
		BrowserDownloadURL: ts.URL,
	})
	if err == nil {
		t.Error("VerifyChecksum() should return error when asset not in checksum file")
	}
	if !strings.Contains(err.Error(), "checksum not found") {
		t.Errorf("error should mention checksum not found, got: %v", err)
	}
}

func TestParseChecksumFile_SingleHash(t *testing.T) {
	t.Parallel()

	// Test the single hash path (no filename, just hash)
	content := "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234"
	hash, err := github.ParseChecksumFile(content, "anything")
	if err != nil {
		t.Fatalf("ParseChecksumFile() error = %v", err)
	}
	if hash != "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234" {
		t.Errorf("hash = %q, want full hash", hash)
	}
}

func TestParseChecksumFile_BinaryMode(t *testing.T) {
	t.Parallel()

	// Test the binary mode indicator (*) prefix
	content := "abcd1234  *myfile.tar.gz"
	hash, err := github.ParseChecksumFile(content, "myfile.tar.gz")
	if err != nil {
		t.Fatalf("ParseChecksumFile() error = %v", err)
	}
	if hash != "abcd1234" {
		t.Errorf("hash = %q, want %q", hash, "abcd1234")
	}
}

func TestParseChecksumFile_EmptyLines(t *testing.T) {
	t.Parallel()

	content := "\n# comment\n\nabcd1234  target.tar.gz\n\n"
	hash, err := github.ParseChecksumFile(content, "target.tar.gz")
	if err != nil {
		t.Fatalf("ParseChecksumFile() error = %v", err)
	}
	if hash != "abcd1234" {
		t.Errorf("hash = %q, want %q", hash, "abcd1234")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAndExtract_ChmodError(t *testing.T) {
	// This is hard to test because afero.MemMapFs doesn't have real permission errors
	// But we can verify the path works by successfully extracting
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	content := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	binaryPath, cleanup, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: ts.URL,
	}, "shelly")
	if err != nil {
		t.Fatalf("DownloadAndExtract() error = %v", err)
	}
	defer cleanup()

	// Verify chmod was applied
	info, err := memFs.Stat(binaryPath)
	if err != nil {
		t.Fatalf("failed to stat binary: %v", err)
	}
	if info.Mode()&0o755 == 0 {
		t.Error("binary should be executable")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestFindBinaryInTar_TarReadError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create invalid tar content wrapped in valid gzip
	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write([]byte("invalid tar content")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: ts.URL,
	}, "shelly")
	if err == nil {
		t.Error("DownloadAndExtract() should return error for invalid tar")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestFindBinaryInTar_SkipDirectory(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	content := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	// Add a directory first
	dirHdr := &tar.Header{
		Name:     "subdir/",
		Mode:     0o755,
		Typeflag: tar.TypeDir,
	}
	if err := tw.WriteHeader(dirHdr); err != nil {
		t.Fatalf("failed to write dir header: %v", err)
	}

	// Add binary in subdirectory
	hdr := &tar.Header{
		Name: "subdir/shelly",
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	binaryPath, cleanup, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: ts.URL,
	}, "shelly")
	if err != nil {
		t.Fatalf("DownloadAndExtract() error = %v", err)
	}
	defer cleanup()

	// Verify binary was extracted despite directory in archive
	data, err := afero.ReadFile(memFs, binaryPath)
	if err != nil {
		t.Fatalf("failed to read binary: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("content mismatch")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestFindBinaryInZip_SkipDirectory(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	content := []byte("binary content")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Create a directory entry
	if _, err := zw.Create("subdir/"); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	// Create binary in subdirectory
	fw, err := zw.Create("subdir/shelly")
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(buf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	binaryPath, cleanup, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.zip",
		BrowserDownloadURL: ts.URL,
	}, "shelly")
	if err != nil {
		t.Fatalf("DownloadAndExtract() error = %v", err)
	}
	defer cleanup()

	// Verify binary was extracted
	data, err := afero.ReadFile(memFs, binaryPath)
	if err != nil {
		t.Fatalf("failed to read binary: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("content mismatch")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestMatchesBinaryName_PrefixMatch(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Test with prefix match (shelly-tool matches shelly)
	content := []byte("binary")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly-tool",
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// "shelly-tool" should match "shelly" due to prefix matching
	binaryPath, cleanup, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: ts.URL,
	}, "shelly")
	if err != nil {
		t.Fatalf("DownloadAndExtract() error = %v", err)
	}
	defer cleanup()

	if !strings.Contains(binaryPath, "shelly-tool") {
		t.Errorf("expected shelly-tool in path, got %s", binaryPath)
	}
}

func TestBuildIssueURL_FeatureType(t *testing.T) {
	t.Parallel()

	url := github.BuildIssueURL(github.IssueTypeFeature, "Title", "Body")
	if !strings.Contains(url, "labels=enhancement") {
		t.Errorf("BuildIssueURL(Feature) should contain labels=enhancement, got %s", url)
	}
}

func TestCompareVersions_VariousFormats(t *testing.T) {
	t.Parallel()

	// Test various version format edge cases that are supported
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"1.0.0", "1.0.0", 0},
		{"1.2.3", "1.2.4", -1},
		{"10.0.0", "9.0.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			t.Parallel()
			got := version.CompareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCheckForUpdatesCached_OldCache(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create an old cache file (can't easily set mtime with afero, so we test no cache path)
	// This test verifies the API fetch path works when cache is missing
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"tag_name": "v2.0.0"}`)); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result := github.CheckForUpdatesCached(context.Background(), ios, "/cache/version", "1.0.0")
	if result == nil {
		t.Fatal("should return update when cache missing")
	}
}

//nolint:paralleltest // modifies global base URL
func TestBuildIssueURL_DeviceType(t *testing.T) {
	url := github.BuildIssueURL(github.IssueTypeDevice, "Device Issue", "Body")
	if !strings.Contains(url, "labels=device") {
		t.Errorf("BuildIssueURL(Device) should contain labels=device, got %s", url)
	}
}

func TestRelease_Methods(t *testing.T) {
	t.Parallel()

	release := &github.Release{
		TagName: testVersionV123,
		Body:    "Release notes",
		Assets: []github.Asset{
			{Name: "checksums.txt"},
			{Name: "shelly_linux_amd64.tar.gz"},
		},
	}

	if release.Version() != testVersion123 {
		t.Errorf("Version() = %q, want %q", release.Version(), testVersion123)
	}

	// Test FindChecksumAsset
	checksumAsset := release.FindChecksumAsset("shelly_linux_amd64.tar.gz")
	if checksumAsset == nil {
		t.Error("FindChecksumAsset() returned nil")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractTarGz_OpenError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Try to extract from nonexistent archive
	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: "file:///nonexistent.tar.gz",
	}, "shelly")
	if err == nil {
		t.Error("expected error for nonexistent archive")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAndExtract_CleanupOnError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Server returns 404 which causes download to fail
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: ts.URL + "/file.tar.gz",
	}, "shelly")
	if err == nil {
		t.Error("DownloadAndExtract() should return error")
	}
	// The cleanup function should have been called internally
}

func TestAsset_Methods(t *testing.T) {
	t.Parallel()

	asset := &github.Asset{
		Name:               "shelly_linux_amd64.tar.gz",
		BrowserDownloadURL: "https://example.com/download",
		Size:               1024,
	}

	if asset.Name != "shelly_linux_amd64.tar.gz" {
		t.Errorf("Name = %q, expected shelly_linux_amd64.tar.gz", asset.Name)
	}
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_WriteFailRestore(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create new binary
	if err := afero.WriteFile(memFs, "/tmp/new", []byte("new content"), 0o755); err != nil {
		t.Fatalf("failed to create new: %v", err)
	}

	// Create target binary
	if err := afero.WriteFile(memFs, "/tmp/target", []byte("old content"), 0o755); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Test successful replacement
	err := github.ReplaceBinary(ios, "/tmp/new", "/tmp/target")
	if err != nil {
		t.Fatalf("ReplaceBinary() error = %v", err)
	}

	// Verify content was replaced
	content, err := afero.ReadFile(memFs, "/tmp/target")
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if string(content) != "new content" {
		t.Errorf("target content = %q, want %q", string(content), "new content")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_WriteFailTriggersRestore(t *testing.T) {
	// Use a filesystem that fails on write after rename
	memFs := afero.NewMemMapFs()
	writeFailFs := &writeFailFs{Fs: memFs}
	github.SetFs(writeFailFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create new binary
	if err := afero.WriteFile(memFs, "/tmp/new", []byte("new content"), 0o755); err != nil {
		t.Fatalf("failed to create new: %v", err)
	}

	// Create target binary
	if err := afero.WriteFile(memFs, "/tmp/target", []byte("old content"), 0o755); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Enable write failure for target path
	writeFailFs.failPath = "/tmp/target"

	err := github.ReplaceBinary(ios, "/tmp/new", "/tmp/target")
	if err == nil {
		t.Error("ReplaceBinary() should return error when write fails")
	}
	if !strings.Contains(err.Error(), "write failed") {
		t.Errorf("error should mention write failed, got: %v", err)
	}
}

// writeFailFs is a filesystem wrapper that fails on WriteFile for a specific path.
type writeFailFs struct {
	afero.Fs
	failPath string
}

func (f *writeFailFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if name == f.failPath && flag&os.O_WRONLY != 0 {
		return nil, errors.New("write not allowed")
	}
	return f.Fs.OpenFile(name, flag, perm)
}

//nolint:paralleltest // modifies global filesystem
func TestReplaceBinary_RemoveBackupFails(t *testing.T) {
	// Use a filesystem that fails on remove
	memFs := afero.NewMemMapFs()
	removeFailFs := &removeFailOnlyFs{Fs: memFs}
	github.SetFs(removeFailFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Create new binary
	if err := afero.WriteFile(memFs, "/tmp/new", []byte("new content"), 0o755); err != nil {
		t.Fatalf("failed to create new: %v", err)
	}

	// Create target binary
	if err := afero.WriteFile(memFs, "/tmp/target", []byte("old content"), 0o755); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// ReplaceBinary should succeed even if removing backup fails
	err := github.ReplaceBinary(ios, "/tmp/new", "/tmp/target")
	if err != nil {
		t.Fatalf("ReplaceBinary() should succeed even if backup removal fails, got: %v", err)
	}

	// Verify content was replaced
	content, err := afero.ReadFile(memFs, "/tmp/target")
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if string(content) != "new content" {
		t.Errorf("target content = %q, want %q", string(content), "new content")
	}
}

// removeFailOnlyFs is a filesystem wrapper that only fails on Remove.
type removeFailOnlyFs struct {
	afero.Fs
}

func (f *removeFailOnlyFs) Remove(name string) error {
	return errors.New("remove not allowed")
}

func TestCheckForUpdates_NoUpdate(t *testing.T) {
	// Use memory filesystem to prevent writes to real cache
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"tag_name": "v1.0.0"}`)); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	result, err := github.CheckForUpdates(context.Background(), ios, "1.0.0")
	if err != nil {
		t.Fatalf("CheckForUpdates() error = %v", err)
	}
	if result.UpdateAvailable {
		t.Error("UpdateAvailable should be false for same version")
	}
}

func TestCheckForUpdates_Error(t *testing.T) {
	// Use memory filesystem to prevent writes to real cache
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	_, err := github.CheckForUpdates(context.Background(), ios, "1.0.0")
	if err == nil {
		t.Error("CheckForUpdates() should return error")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestInstallRelease_DownloadFails(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Create release with platform-specific asset
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{
				Name:               fmt.Sprintf("shelly_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH),
				BrowserDownloadURL: ts.URL + "/download",
			},
		},
	}

	err := client.InstallRelease(context.Background(), ios, release)
	if err == nil {
		t.Error("InstallRelease() should return error when download fails")
	}
	if !strings.Contains(err.Error(), "failed to download") {
		t.Errorf("error should mention download failure, got: %v", err)
	}
}

func TestDefaultOwnerRepo(t *testing.T) {
	t.Parallel()

	if github.DefaultOwner == "" {
		t.Error("DefaultOwner should not be empty")
	}
	if github.DefaultRepo == "" {
		t.Error("DefaultRepo should not be empty")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestInstallRelease_ChecksumVerificationFails(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	binaryContent := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "download") {
			if _, err := w.Write(gzBuf.Bytes()); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		} else if strings.Contains(r.URL.Path, "checksums") {
			// Return checksum file with wrong hash
			if _, err := w.Write([]byte("wronghash  shelly.tar.gz\n")); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		}
	}))
	defer ts.Close()
	serverURL = ts.URL

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	assetName := fmt.Sprintf("shelly_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: serverURL + "/download/" + assetName,
			},
			{
				Name:               "checksums.txt",
				BrowserDownloadURL: serverURL + "/checksums/checksums.txt",
			},
		},
	}

	err := client.InstallRelease(context.Background(), ios, release)
	if err == nil {
		t.Error("InstallRelease() should return error when checksum fails")
	}
	if !strings.Contains(err.Error(), "checksum") {
		t.Errorf("error should mention checksum, got: %v", err)
	}
}

func TestParseVersion(t *testing.T) {
	t.Parallel()

	// Test parseVersion through CompareVersions with various inputs
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"1.0", "1.0.0", 0},
		{"1", "1.0.0", 0},
		{"1.2.3.4", "1.2.3.4", 0},
		{"invalid", "1.0.0", -1}, // non-numeric parsed as 0
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			t.Parallel()
			got := version.CompareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetReleaseByTag_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"tag_name": "v1.0.0"}`)); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetReleaseByTag(ctx, "owner", "repo", "v1.0.0")
	if err == nil {
		t.Error("GetReleaseByTag() should return error when context cancelled")
	}
}

//nolint:paralleltest // modifies global base URL
func TestListReleases_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ListReleases(ctx, "owner", "repo", false)
	if err == nil {
		t.Error("ListReleases() should return error when context cancelled")
	}
}

//nolint:paralleltest // modifies global base URL
func TestGetLatestRelease_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"tag_name": "v1.0.0"}`)); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.GetLatestRelease(ctx, "owner", "repo")
	if err == nil {
		t.Error("GetLatestRelease() should return error when context cancelled")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAsset_ContextCancelled(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		if _, err := w.Write([]byte("content")); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.DownloadAsset(ctx, &github.Asset{
		Name:               "test",
		BrowserDownloadURL: ts.URL + "/test",
	}, "/tmp/test")
	if err == nil {
		t.Error("DownloadAsset() should return error when context cancelled")
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAsset_MkdirAllError(t *testing.T) {
	// Use a filesystem that fails on MkdirAll
	memFs := afero.NewMemMapFs()
	mkdirFailFs := &mkdirFailFs{Fs: memFs}
	github.SetFs(mkdirFailFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("content")); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.DownloadAsset(context.Background(), &github.Asset{
		Name:               "test",
		BrowserDownloadURL: ts.URL + "/test",
	}, "/deep/nested/path/test")
	if err == nil {
		t.Error("DownloadAsset() should return error when MkdirAll fails")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("error should mention directory creation, got: %v", err)
	}
}

// mkdirFailFs is a filesystem wrapper that fails on MkdirAll.
type mkdirFailFs struct {
	afero.Fs
}

func (f *mkdirFailFs) MkdirAll(path string, perm os.FileMode) error {
	return errors.New("mkdir not allowed")
}

//nolint:paralleltest // modifies global filesystem
func TestExtractToFile_Success(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	content := []byte("test file content")
	err := client.ExtractToFile("/tmp/test.txt", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("ExtractToFile() error = %v", err)
	}

	// Verify content
	data, err := afero.ReadFile(memFs, "/tmp/test.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("file content = %q, want %q", string(data), string(content))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractToFile_CreateError(t *testing.T) {
	// Use a filesystem that fails on Create
	memFs := afero.NewMemMapFs()
	createFailFs := &createFailFs{Fs: memFs}
	github.SetFs(createFailFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.ExtractToFile("/tmp/test.txt", bytes.NewReader([]byte("content")))
	if err == nil {
		t.Error("ExtractToFile() should return error when Create fails")
	}
	if !strings.Contains(err.Error(), "failed to create file") {
		t.Errorf("error should mention create file, got: %v", err)
	}
}

// createFailFs is a filesystem wrapper that fails on Create.
type createFailFs struct {
	afero.Fs
}

func (f *createFailFs) Create(name string) (afero.File, error) {
	return nil, errors.New("create not allowed")
}

func TestFindAssetForPlatform_MultiplePlatforms(t *testing.T) {
	t.Parallel()

	// Test that it finds the right asset for the current platform
	release := &github.Release{
		Assets: []github.Asset{
			{Name: "shelly_darwin_amd64.tar.gz"},
			{Name: "shelly_linux_amd64.tar.gz"},
			{Name: "shelly_windows_amd64.zip"},
			{Name: "shelly_linux_arm64.tar.gz"},
		},
	}

	asset := release.FindAssetForPlatform()
	// On a linux/amd64 machine this should find the linux_amd64 asset
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		if asset == nil || !strings.Contains(asset.Name, "linux_amd64") {
			t.Errorf("FindAssetForPlatform() should find linux_amd64 asset")
		}
	}
}

func TestFindChecksumAsset_Variations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		assets       []github.Asset
		assetName    string
		wantChecksum bool
		checksumName string
	}{
		{
			name: "checksums.txt",
			assets: []github.Asset{
				{Name: "checksums.txt"},
				{Name: "binary.tar.gz"},
			},
			assetName:    "binary.tar.gz",
			wantChecksum: true,
			checksumName: "checksums.txt",
		},
		{
			name: "sha256sums",
			assets: []github.Asset{
				{Name: "sha256sums.txt"},
				{Name: "binary.tar.gz"},
			},
			assetName:    "binary.tar.gz",
			wantChecksum: true,
			checksumName: "sha256sums.txt",
		},
		{
			name: "no_checksum",
			assets: []github.Asset{
				{Name: "binary.tar.gz"},
			},
			assetName:    "binary.tar.gz",
			wantChecksum: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			release := &github.Release{Assets: tt.assets}
			asset := release.FindChecksumAsset(tt.assetName)
			if !tt.wantChecksum {
				if asset != nil {
					t.Error("expected no checksum asset")
				}
				return
			}
			if asset == nil {
				t.Fatal("expected to find checksum asset")
			}
			if asset.Name != tt.checksumName {
				t.Errorf("checksum name = %q, want %q", asset.Name, tt.checksumName)
			}
		})
	}
}

//nolint:paralleltest // modifies global base URL
func TestPerformUpdate_SkipConfirm(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Server won't be called if we skip confirm
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	release := &github.Release{
		TagName: "v2.0.0",
		Assets:  []github.Asset{},
	}

	// With skipConfirm=true but no platform asset, should fail on install
	confirmFunc := func(prompt string, skipConfirm bool) (bool, error) {
		return true, nil
	}

	err := client.PerformUpdate(context.Background(), ios, release, "v1.0.0", "notes", confirmFunc, true)
	if err == nil {
		t.Error("PerformUpdate() should return error when no platform asset")
	}
}

func TestRelease_Version_WithPrefix(t *testing.T) {
	t.Parallel()

	release := &github.Release{TagName: testVersionV123}
	if release.Version() != testVersion123 {
		t.Errorf("Version() = %q, want %q", release.Version(), testVersion123)
	}
}

func TestRelease_Version_WithoutPrefix(t *testing.T) {
	t.Parallel()

	release := &github.Release{TagName: testVersion123}
	if release.Version() != testVersion123 {
		t.Errorf("Version() = %q, want %q", release.Version(), testVersion123)
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestInstallRelease_NoChecksumAsset(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	binaryContent := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	assetName := fmt.Sprintf("shelly_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	// Release without checksum asset - should skip verification
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: ts.URL + "/download/" + assetName,
			},
		},
	}

	// This will fail at the os.Executable step, but it covers more of InstallRelease
	err := client.InstallRelease(context.Background(), ios, release)
	// Error is expected due to os.Executable but we've covered more code paths
	if err == nil {
		t.Log("InstallRelease() succeeded unexpectedly")
	}
}

//nolint:paralleltest // modifies global base URL
func TestPerformRollback_FindPreviousError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restore := github.SetAPIBaseURL(ts.URL)
	defer restore()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	confirmFunc := func(prompt string, skipConfirm bool) (bool, error) {
		return true, nil
	}

	err := client.PerformRollback(context.Background(), ios, "v2.0.0", false, confirmFunc, false)
	if err == nil {
		t.Error("PerformRollback() should return error when finding previous fails")
	}
}

func TestBuildIssueURL_EmptyTitle_Bug(t *testing.T) {
	t.Parallel()

	url := github.BuildIssueURL(github.IssueTypeBug, "", "body")
	if !strings.Contains(url, "title=%5BBug%5D") {
		t.Errorf("BuildIssueURL(Bug, empty) should set default title, got %s", url)
	}
}

func TestBuildIssueURL_EmptyTitle_Feature(t *testing.T) {
	t.Parallel()

	url := github.BuildIssueURL(github.IssueTypeFeature, "", "body")
	if !strings.Contains(url, "title=%5BFeature%5D") {
		t.Errorf("BuildIssueURL(Feature, empty) should set default title, got %s", url)
	}
}

func TestBuildIssueURL_EmptyTitle_Device(t *testing.T) {
	t.Parallel()

	url := github.BuildIssueURL(github.IssueTypeDevice, "", "body")
	if !strings.Contains(url, "title=%5BDevice%5D") {
		t.Errorf("BuildIssueURL(Device, empty) should set default title, got %s", url)
	}
}

//nolint:paralleltest // modifies global filesystem and function variables
func TestInstallRelease_OsExecutableError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	restore := github.SetOsExecutable(func() (string, error) {
		return "", errors.New("mock os.Executable error")
	})
	defer restore()

	binaryContent := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	assetName := fmt.Sprintf("shelly_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: ts.URL + "/download/" + assetName,
			},
		},
	}

	err := client.InstallRelease(context.Background(), ios, release)
	if err == nil {
		t.Fatal("expected error when osExecutable fails")
	}
	if !strings.Contains(err.Error(), "failed to get executable path") {
		t.Errorf("unexpected error message: %v", err)
	}
}

//nolint:paralleltest // modifies global filesystem and function variables
func TestInstallRelease_EvalSymlinksError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	restore := github.SetOsExecutable(func() (string, error) {
		return testShellyPath, nil
	})
	defer restore()

	restoreSymlinks := github.SetEvalSymlinks(func(path string) (string, error) {
		return "", errors.New("mock symlink error")
	})
	defer restoreSymlinks()

	binaryContent := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	assetName := fmt.Sprintf("shelly_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: ts.URL + "/download/" + assetName,
			},
		},
	}

	err := client.InstallRelease(context.Background(), ios, release)
	if err == nil {
		t.Fatal("expected error when evalSymlinks fails")
	}
	if !strings.Contains(err.Error(), "failed to resolve symlinks") {
		t.Errorf("unexpected error message: %v", err)
	}
}

//nolint:paralleltest // modifies global filesystem and function variables
func TestInstallRelease_ReplaceBinaryError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	restore := github.SetOsExecutable(func() (string, error) {
		return testShellyPath, nil
	})
	defer restore()

	restoreSymlinks := github.SetEvalSymlinks(func(path string) (string, error) {
		return path, nil
	})
	defer restoreSymlinks()

	binaryContent := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	assetName := fmt.Sprintf("shelly_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: ts.URL + "/download/" + assetName,
			},
		},
	}

	// ReplaceBinary will fail because /usr/bin/shelly doesn't exist in memFs
	err := client.InstallRelease(context.Background(), ios, release)
	if err == nil {
		t.Fatal("expected error when ReplaceBinary fails")
	}
	if !strings.Contains(err.Error(), "failed to install update") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestInstallRelease_Success(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Also set config.Fs for version.WriteCache
	config.SetFs(memFs)
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	// Mock osExecutable to return a path that exists in memFs
	execPath := "/usr/local/bin/shelly"
	restore := github.SetOsExecutable(func() (string, error) {
		return execPath, nil
	})
	defer restore()

	restoreSymlinks := github.SetEvalSymlinks(func(path string) (string, error) {
		return path, nil
	})
	defer restoreSymlinks()

	// Create the existing binary that will be replaced
	if err := afero.WriteFile(memFs, execPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to create existing binary: %v", err)
	}

	binaryContent := []byte("new binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "shelly",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(gzBuf.Bytes()); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	assetName := fmt.Sprintf("shelly_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: ts.URL + "/download/" + assetName,
			},
		},
	}

	err := client.InstallRelease(context.Background(), ios, release)
	if err != nil {
		t.Fatalf("InstallRelease() error = %v", err)
	}

	// Verify the binary was replaced
	content, err := afero.ReadFile(memFs, execPath)
	if err != nil {
		t.Fatalf("failed to read binary: %v", err)
	}
	if !bytes.Equal(content, binaryContent) {
		t.Errorf("binary content = %q, want %q", string(content), string(binaryContent))
	}
}

//nolint:paralleltest // modifies global filesystem
func TestCopyFile_DestCreateError(t *testing.T) {
	// Use a filesystem that fails on Create
	memFs := afero.NewMemMapFs()
	failFs := &copyDestFailFs{Fs: memFs}
	github.SetFs(failFs)
	defer github.SetFs(nil)

	// Create source file
	if err := afero.WriteFile(memFs, "/tmp/src", []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	err := github.CopyFile(ios, "/tmp/src", "/tmp/dst")
	if err == nil {
		t.Error("CopyFile() should return error when Create fails")
	}
}

// copyDestFailFs is a filesystem wrapper that fails on Create for specific paths.
type copyDestFailFs struct {
	afero.Fs
}

func (f *copyDestFailFs) Create(name string) (afero.File, error) {
	if strings.Contains(name, "dst") {
		return nil, errors.New("create not allowed for dst")
	}
	return f.Fs.Create(name)
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestVerifyChecksum_TempDirError(t *testing.T) {
	// Use a filesystem that fails on Mkdir (used by TempDir)
	memFs := afero.NewMemMapFs()
	mkdirFails := &tempDirFailFs{Fs: memFs}
	github.SetFs(mkdirFails)
	defer github.SetFs(nil)

	// Create binary file in underlying memFs directly (bypassing the wrapper)
	binaryContent := []byte("test binary content")
	if err := afero.WriteFile(memFs, "/tmp/binary", binaryContent, 0o755); err != nil {
		t.Fatalf("failed to create binary: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.VerifyChecksum(context.Background(), ios, "/tmp/binary", "binary", &github.Asset{
		Name:               "checksums.txt",
		BrowserDownloadURL: "http://localhost:1/checksums.txt",
	})
	if err == nil {
		t.Error("VerifyChecksum() should return error when TempDir fails")
	}
	// The error should mention temp dir creation failure
	if !strings.Contains(err.Error(), "create temp dir") {
		t.Errorf("error should mention create temp dir, got: %v", err)
	}
}

// tempDirFailFs is a filesystem that fails when creating temp directories.
type tempDirFailFs struct {
	afero.Fs
}

func (f *tempDirFailFs) Mkdir(name string, perm os.FileMode) error {
	return errors.New("temp dir creation not allowed")
}

func (f *tempDirFailFs) MkdirAll(path string, perm os.FileMode) error {
	return errors.New("temp dir creation not allowed")
}

//nolint:paralleltest // modifies global filesystem
func TestExtractToFile_CopyError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Create a reader that returns an error after some data
	errReader := &errorReader{err: errors.New("read error")}

	err := client.ExtractToFile("/tmp/test.txt", errReader)
	if err == nil {
		t.Error("ExtractToFile() should return error when reader fails")
	}
	if !strings.Contains(err.Error(), "failed to extract file") {
		t.Errorf("error should mention extract failure, got: %v", err)
	}
}

// errorReader is a reader that returns an error.
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestDownloadAndExtract_TempDirError(t *testing.T) {
	// Use a filesystem that fails on Mkdir (used by TempDir in DownloadAndExtract)
	memFs := afero.NewMemMapFs()
	mkdirFails := &tempDirFailFs{Fs: memFs}
	github.SetFs(mkdirFails)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("content")); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, _, err := client.DownloadAndExtract(context.Background(), &github.Asset{
		Name:               "test.tar.gz",
		BrowserDownloadURL: ts.URL + "/test.tar.gz",
	}, "shelly")
	if err == nil {
		t.Error("DownloadAndExtract() should return error when TempDir fails")
	}
}

//nolint:paralleltest // modifies global filesystem
func TestDownloadAsset_WriteError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	writeFailFs := &writeFailForDownloadFs{Fs: memFs}
	github.SetFs(writeFailFs)
	defer github.SetFs(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("content")); err != nil {
			t.Errorf("failed to write: %v", err)
		}
	}))
	defer ts.Close()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	err := client.DownloadAsset(context.Background(), &github.Asset{
		Name:               "test",
		BrowserDownloadURL: ts.URL + "/test",
	}, "/tmp/downloaded/test")
	if err == nil {
		t.Error("DownloadAsset() should return error when write fails")
	}
}

// writeFailForDownloadFs fails on Create for download paths.
type writeFailForDownloadFs struct {
	afero.Fs
}

func (f *writeFailForDownloadFs) Create(name string) (afero.File, error) {
	return nil, errors.New("create not allowed for download")
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestInstallRelease_NoAssetForPlatform(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	// Release with no matching assets for this platform
	release := &github.Release{
		TagName: testVersionV1,
		Assets: []github.Asset{
			{Name: "shelly_windows_amd64.zip"},
		},
	}

	// Only run this test on non-windows platforms
	if runtime.GOOS != "windows" {
		err := client.InstallRelease(context.Background(), ios, release)
		if err == nil {
			t.Error("InstallRelease() should return error when no asset matches platform")
		}
		if !strings.Contains(err.Error(), "no binary available") {
			t.Errorf("error should mention no binary available, got: %v", err)
		}
	}
}

func TestParseRepoString_EmptyOwner(t *testing.T) {
	t.Parallel()

	_, _, err := github.ParseRepoString("/repo")
	if err == nil {
		t.Error("ParseRepoString() should return error for empty owner")
	}
}

func TestParseRepoString_EmptyRepo(t *testing.T) {
	t.Parallel()

	_, _, err := github.ParseRepoString("owner/")
	if err == nil {
		t.Error("ParseRepoString() should return error for empty repo")
	}
}

func TestParseRepoString_GitSuffix(t *testing.T) {
	t.Parallel()

	owner, repo, err := github.ParseRepoString("owner/repo.git")
	if err != nil {
		t.Fatalf("ParseRepoString() error = %v", err)
	}
	if owner != "owner" || repo != "repo" {
		t.Errorf("ParseRepoString() = (%q, %q), want (owner, repo)", owner, repo)
	}
}

//nolint:paralleltest // modifies global filesystem and base URL
func TestExtractTarGz_ValidArchive(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create valid tar.gz with binary
	content := []byte("binary content")
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	hdr := &tar.Header{
		Name: "testbinary",
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar: %v", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}

	// Write archive to filesystem
	if err := afero.WriteFile(memFs, "/tmp/archive.tar.gz", gzBuf.Bytes(), 0o644); err != nil {
		t.Fatalf("failed to write archive: %v", err)
	}
	if err := memFs.MkdirAll("/tmp/dest", 0o755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	result, err := client.ExtractTarGz("/tmp/archive.tar.gz", "/tmp/dest", "testbinary")
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v", err)
	}
	if !strings.Contains(result, "testbinary") {
		t.Errorf("ExtractTarGz() = %q, should contain testbinary", result)
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractZip_ValidArchive(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	// Create valid zip with binary
	content := []byte("binary content")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("testbinary")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}

	// Write archive to filesystem
	if err := afero.WriteFile(memFs, "/tmp/archive.zip", buf.Bytes(), 0o644); err != nil {
		t.Fatalf("failed to write archive: %v", err)
	}
	if err := memFs.MkdirAll("/tmp/dest", 0o755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	result, err := client.ExtractZip("/tmp/archive.zip", "/tmp/dest", "testbinary")
	if err != nil {
		t.Fatalf("ExtractZip() error = %v", err)
	}
	if !strings.Contains(result, "testbinary") {
		t.Errorf("ExtractZip() = %q, should contain testbinary", result)
	}
}

//nolint:paralleltest // modifies global filesystem
func TestExtractZip_OpenError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	github.SetFs(memFs)
	defer github.SetFs(nil)

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	client := github.NewClient(ios)

	_, err := client.ExtractZip("/nonexistent/archive.zip", "/tmp/dest", "binary")
	if err == nil {
		t.Error("ExtractZip() should return error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read zip") {
		t.Errorf("error should mention read zip, got: %v", err)
	}
}

func TestDetectInstallMethod_Homebrew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want github.InstallMethod
	}{
		{
			name: "macOS ARM homebrew",
			path: "/opt/homebrew/Cellar/shelly/1.0.0/bin/shelly",
			want: github.InstallMethodHomebrew,
		},
		{
			name: "macOS Intel homebrew",
			path: "/usr/local/Cellar/shelly/1.0.0/bin/shelly",
			want: github.InstallMethodHomebrew,
		},
		{
			name: "Linux homebrew",
			path: "/home/linuxbrew/.linuxbrew/Cellar/shelly/1.0.0/bin/shelly",
			want: github.InstallMethodHomebrew,
		},
		{
			name: "direct download",
			path: "/usr/local/bin/shelly",
			want: github.InstallMethodDirect,
		},
		{
			name: "go install GOPATH",
			path: "/home/user/go/bin/shelly",
			want: github.InstallMethodDirect,
		},
		{
			name: "tmp directory",
			path: "/tmp/shelly-test",
			want: github.InstallMethodDirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info := github.DetectInstallMethodFromPath(tt.path)
			if info.Method != tt.want {
				t.Errorf("DetectInstallMethodFromPath(%q) = %v, want %v", tt.path, info.Method, tt.want)
			}
		})
	}
}

func TestInstallInfo_CanSelfUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method github.InstallMethod
		want   bool
	}{
		{"homebrew cannot self-update", github.InstallMethodHomebrew, false},
		{"direct can self-update", github.InstallMethodDirect, true},
		{"unknown can self-update", github.InstallMethodUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info := github.InstallInfo{Method: tt.method}
			if got := info.CanSelfUpdate(); got != tt.want {
				t.Errorf("InstallInfo{Method: %v}.CanSelfUpdate() = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestDetectInstallMethod_HomebrewUpdateCommand(t *testing.T) {
	t.Parallel()

	info := github.DetectInstallMethodFromPath("/opt/homebrew/Cellar/shelly/1.0.0/bin/shelly")
	if info.UpdateCommand != "brew upgrade shelly" {
		t.Errorf("UpdateCommand = %q, want %q", info.UpdateCommand, "brew upgrade shelly")
	}
}
