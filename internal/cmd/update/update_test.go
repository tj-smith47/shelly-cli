package update

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

const (
	betaChannel = "beta"
	testVersion = "v1.0.0"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "update" {
		t.Errorf("Use = %q, want %q", cmd.Use, "update")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, alias := range cmd.Aliases {
		aliasMap[alias] = true
	}

	expected := []string{"upgrade", "u"}
	for _, alias := range expected {
		if !aliasMap[alias] {
			t.Errorf("missing alias %q", alias)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly update") {
		t.Error("Example should contain 'shelly update'")
	}

	if !strings.Contains(cmd.Example, "--check") {
		t.Error("Example should contain '--check'")
	}

	if !strings.Contains(cmd.Example, "--version") {
		t.Error("Example should contain '--version'")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check --check flag
	checkFlag := cmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Fatal("--check flag not found")
	}
	if checkFlag.Shorthand != "c" {
		t.Errorf("--check shorthand = %q, want %q", checkFlag.Shorthand, "c")
	}

	// Check --version flag
	versionFlag := cmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Fatal("--version flag not found")
	}

	// Check --channel flag
	channelFlag := cmd.Flags().Lookup("channel")
	if channelFlag == nil {
		t.Fatal("--channel flag not found")
	}
	if channelFlag.DefValue != "stable" {
		t.Errorf("--channel default = %q, want %q", channelFlag.DefValue, "stable")
	}

	// Check --rollback flag
	rollbackFlag := cmd.Flags().Lookup("rollback")
	if rollbackFlag == nil {
		t.Fatal("--rollback flag not found")
	}

	// Check --yes flag
	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("--yes flag not found")
	}

	// Check --include-pre flag
	includePreFlag := cmd.Flags().Lookup("include-pre")
	if includePreFlag == nil {
		t.Fatal("--include-pre flag not found")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test flag parsing
	if err := cmd.Flags().Set("check", "true"); err != nil {
		t.Fatalf("failed to set check flag: %v", err)
	}
	if err := cmd.Flags().Set("version", testVersion); err != nil {
		t.Fatalf("failed to set version flag: %v", err)
	}
	if err := cmd.Flags().Set("channel", betaChannel); err != nil {
		t.Fatalf("failed to set channel flag: %v", err)
	}

	check, err := cmd.Flags().GetBool("check")
	if err != nil {
		t.Fatalf("get check flag: %v", err)
	}
	if !check {
		t.Error("check flag should be true")
	}

	ver, err := cmd.Flags().GetString("version")
	if err != nil {
		t.Fatalf("get version flag: %v", err)
	}
	if ver != testVersion {
		t.Errorf("version = %q, want %q", ver, testVersion)
	}

	channel, err := cmd.Flags().GetString("channel")
	if err != nil {
		t.Fatalf("get channel flag: %v", err)
	}
	if channel != betaChannel {
		t.Errorf("channel = %q, want %q", channel, betaChannel)
	}
}

func TestRun_DevelopmentVersion_Check(t *testing.T) {
	t.Parallel()

	// In test environment, version.Version is "dev" by default
	// which makes IsDevelopment() return true
	if !version.IsDevelopment() {
		t.Skip("Test requires development version")
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--check"})

	err := cmd.Execute()
	// Dev builds can check for updates without error
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should show update available (dev is older than any release)
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "Update available") {
		t.Errorf("ErrOutput = %q, want to contain 'Update available'", errOutput)
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Check {
		t.Error("Check should default to false")
	}
	if opts.Version != "" {
		t.Error("Version should default to empty")
	}
	if opts.Channel != "" {
		t.Error("Channel should default to empty (set by flag default)")
	}
	if opts.Rollback {
		t.Error("Rollback should default to false")
	}
	if opts.IncludePre {
		t.Error("IncludePre should default to false")
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory:    f,
		Check:      true,
		Version:    "v1.2.3",
		Channel:    betaChannel,
		Rollback:   true,
		IncludePre: true,
	}

	if opts.Factory != f {
		t.Error("Factory not set correctly")
	}
	if !opts.Check {
		t.Error("Check should be true")
	}
	if opts.Version != "v1.2.3" {
		t.Errorf("Version = %q, want %q", opts.Version, "v1.2.3")
	}
	if opts.Channel != betaChannel {
		t.Errorf("Channel = %q, want %q", opts.Channel, betaChannel)
	}
	if !opts.Rollback {
		t.Error("Rollback should be true")
	}
	if !opts.IncludePre {
		t.Error("IncludePre should be true")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNewCommand_InvalidFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--invalid-flag"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid flag")
	}
}

// mockRelease creates a mock GitHub release response.
func mockRelease(tagName string, prerelease bool) github.Release {
	return github.Release{
		TagName:    tagName,
		Name:       "Release " + tagName,
		Body:       "Release notes for " + tagName,
		Prerelease: prerelease,
		HTMLURL:    "https://github.com/test/repo/releases/" + tagName,
	}
}

// setupMockGitHubServer creates a test server that mocks the GitHub API.
// Cleanup is handled automatically via t.Cleanup.
func setupMockGitHubServer(t *testing.T, latestRelease *github.Release, releases []github.Release) {
	t.Helper()

	mux := http.NewServeMux()

	// Latest release endpoint
	mux.HandleFunc("/repos/tj-smith47/shelly-cli/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		if latestRelease == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(latestRelease); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	})

	// All releases endpoint
	mux.HandleFunc("/repos/tj-smith47/shelly-cli/releases", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	})

	// Specific release by tag endpoint
	mux.HandleFunc("/repos/tj-smith47/shelly-cli/releases/tags/", func(w http.ResponseWriter, r *http.Request) {
		tag := strings.TrimPrefix(r.URL.Path, "/repos/tj-smith47/shelly-cli/releases/tags/")
		for _, rel := range releases {
			if rel.TagName == tag {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(rel); err != nil {
					t.Logf("failed to encode response: %v", err)
				}
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})

	server := httptest.NewServer(mux)

	// Override the GitHub API base URL
	oldURL := github.GitHubAPIBaseURL
	github.GitHubAPIBaseURL = server.URL
	t.Cleanup(func() {
		github.GitHubAPIBaseURL = oldURL
		server.Close()
	})
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_CheckMode_UpdateAvailable(t *testing.T) {
	release := mockRelease("v99.0.0", false)
	setupMockGitHubServer(t, &release, []github.Release{release})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// In development mode, it will still fetch and display the version
	if !strings.Contains(output, "v99.0.0") {
		t.Errorf("Output should contain available version 'v99.0.0', got: %s", output)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_CheckMode_NoUpdateAvailable(t *testing.T) {
	release := mockRelease("v0.0.1", false)
	setupMockGitHubServer(t, &release, []github.Release{release})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should still succeed and show current status
	output := tf.OutString()
	if output == "" {
		t.Error("Expected some output")
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_NoReleasesFound(t *testing.T) {
	// Set up server with no latest release
	setupMockGitHubServer(t, nil, []github.Release{})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should show "No releases found" message
	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "No releases found") {
		t.Errorf("Output should contain 'No releases found', got: %s", output)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_AlreadyAtLatestVersion(t *testing.T) {
	// In development mode, this path won't be hit because of the dev version check
	// So we test via --check mode
	release := mockRelease("v0.0.1", false)
	setupMockGitHubServer(t, &release, []github.Release{release})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// In development mode with --check, it will show the version comparison
	output := tf.OutString()
	if output == "" {
		t.Error("Expected some output")
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_SpecificVersion(t *testing.T) {
	releases := []github.Release{
		mockRelease("v2.0.0", false),
		mockRelease("v1.5.0", false),
		mockRelease("v1.0.0", false),
	}
	setupMockGitHubServer(t, &releases[0], releases)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	// Request a specific version with --check to avoid download
	cmd.SetArgs([]string{"--version", "v1.5.0", "--check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "v1.5.0") {
		t.Errorf("Output should contain requested version 'v1.5.0', got: %s", output)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_IncludePrerelease(t *testing.T) {
	releases := []github.Release{
		mockRelease("v99.0.0-beta", true),
		mockRelease("v1.0.0", false),
	}
	// Latest stable is v1.0.0, but with --include-pre we get v99.0.0-beta
	latestStable := mockRelease("v1.0.0", false)
	setupMockGitHubServer(t, &latestStable, releases)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--include-pre", "--check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should show the prerelease version
	if !strings.Contains(output, "v99.0.0-beta") {
		t.Errorf("Output should contain prerelease version 'v99.0.0-beta', got: %s", output)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_BetaChannel(t *testing.T) {
	releases := []github.Release{
		mockRelease("v99.0.0-beta", true),
		mockRelease("v1.0.0", false),
	}
	latestStable := mockRelease("v1.0.0", false)
	setupMockGitHubServer(t, &latestStable, releases)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--channel", betaChannel, "--check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Beta channel should include prereleases
	if !strings.Contains(output, "v99.0.0-beta") {
		t.Errorf("Output should contain beta version 'v99.0.0-beta', got: %s", output)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global GitHubAPIBaseURL
func TestRun_GitHubAPIError(t *testing.T) {
	// Create a server that returns 500 errors
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	oldURL := github.GitHubAPIBaseURL
	github.GitHubAPIBaseURL = server.URL
	defer func() { github.GitHubAPIBaseURL = oldURL }()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--check"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for server error")
	}

	if !strings.Contains(err.Error(), "failed to fetch release") {
		t.Errorf("Error should contain 'failed to fetch release', got: %v", err)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global version.Version
func TestRun_NonDev_AlreadyAtLatestVersion(t *testing.T) {
	// Set a non-dev version to test the "already at latest" path
	oldVersion := version.Version
	version.Version = "v1.0.0"
	defer func() { version.Version = oldVersion }()

	// Mock returns v0.0.1, which is older than v1.0.0
	release := mockRelease("v0.0.1", false)
	setupMockGitHubServer(t, &release, []github.Release{release})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Already at latest") {
		t.Errorf("Output should contain 'Already at latest', got: %s", output)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global version.Version and GitHubAPIBaseURL
func TestRun_NonDev_Rollback(t *testing.T) {
	// Set a non-dev version
	oldVersion := version.Version
	version.Version = "v2.0.0"
	defer func() { version.Version = oldVersion }()

	// Mock releases for rollback
	releases := []github.Release{
		mockRelease("v2.0.0", false),
		mockRelease("v1.5.0", false),
		mockRelease("v1.0.0", false),
	}
	setupMockGitHubServer(t, &releases[0], releases)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	// Use --yes to skip confirmation, expect it to fail on download
	// since the mock server doesn't serve actual binary assets
	cmd.SetArgs([]string{"--rollback", "--yes"})

	err := cmd.Execute()
	// The rollback path is exercised but fails at InstallRelease
	// because the mock server doesn't serve actual binaries
	if err == nil {
		t.Fatal("Expected error from rollback (no binary in mock server)")
	}

	// Verify the error is about missing binary, not a different failure
	if !strings.Contains(err.Error(), "no binary") {
		t.Errorf("Expected 'no binary' error, got: %v", err)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global version.Version and GitHubAPIBaseURL
func TestRun_NonDev_UpdateWithConfirmation_Declined(t *testing.T) {
	// Set a non-dev version
	oldVersion := version.Version
	version.Version = "v1.0.0"
	defer func() { version.Version = oldVersion }()

	// Mock returns v2.0.0, which is newer than v1.0.0
	release := mockRelease("v2.0.0", false)
	setupMockGitHubServer(t, &release, []github.Release{release})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	// No --yes flag means confirmation is required
	// In non-TTY test mode, confirmation returns false by default
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	// PerformUpdate returns "update cancelled by user" error when declined
	if err == nil {
		t.Fatal("Expected 'update cancelled by user' error")
	}

	if !strings.Contains(err.Error(), "update cancelled by user") {
		t.Errorf("Expected 'update cancelled by user' error, got: %v", err)
	}
}
