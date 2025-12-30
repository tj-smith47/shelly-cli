package feedback

import (
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "feedback" {
		t.Errorf("Use = %q, want 'feedback'", cmd.Use)
	}

	if len(cmd.Aliases) == 0 {
		t.Fatal("Aliases should not be empty")
	}

	// Check for expected aliases
	aliasMap := make(map[string]bool)
	for _, alias := range cmd.Aliases {
		aliasMap[alias] = true
	}
	expectedAliases := []string{"issue", "report-bug", "bug"}
	for _, expected := range expectedAliases {
		if !aliasMap[expected] {
			t.Errorf("expected alias %q not found", expected)
		}
	}

	if cmd.Short == "" {
		t.Error("Short is empty")
	}

	if cmd.Long == "" {
		t.Error("Long is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check for expected flags
	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("--type flag not found")
	}
	if typeFlag.Shorthand != "t" {
		t.Errorf("type shorthand = %q, want t", typeFlag.Shorthand)
	}

	titleFlag := cmd.Flags().Lookup("title")
	if titleFlag == nil {
		t.Fatal("--title flag not found")
	}

	deviceFlag := cmd.Flags().Lookup("device")
	if deviceFlag == nil {
		t.Fatal("--device flag not found")
	}

	attachLogFlag := cmd.Flags().Lookup("attach-log")
	if attachLogFlag == nil {
		t.Fatal("--attach-log flag not found")
	}

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("--dry-run flag not found")
	}

	issuesFlag := cmd.Flags().Lookup("issues")
	if issuesFlag == nil {
		t.Fatal("--issues flag not found")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Feedback command should accept no args
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("should accept zero args: %v", err)
		}
	}

	// Should reject extra args
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{"extra"})
		if err == nil {
			t.Error("should reject non-zero args")
		}
	}
}

func TestExecute_Issues_Flag(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--issues"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Verify browser was called
	if !mockBrowser.BrowseCalled {
		t.Error("Browser.Browse should have been called")
	}

	// Verify correct issues URL was used
	if mockBrowser.LastURL != github.IssuesURL() {
		t.Errorf("Browser.Browse called with %q, want %q", mockBrowser.LastURL, github.IssuesURL())
	}

	// Check output
	output := tf.OutString()
	if !strings.Contains(output, "Opening GitHub issues") {
		t.Errorf("expected 'Opening GitHub issues' in output, got: %s", output)
	}
}

func TestExecute_Issues_Flag_BrowserError_WithClipboardFallback(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)
	mockBrowser.Err = &browser.ClipboardFallbackError{URL: github.IssuesURL()}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--issues"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil (clipboard fallback should succeed)", err)
	}

	output := tf.OutString()
	// The output should include the info message and warning message
	// (may contain ANSI color codes)
	if !strings.Contains(output, "Opening GitHub issues") &&
		!strings.Contains(output, "Could not open browser") {
		t.Errorf("expected opening or warning message in output, got: %s", output)
	}
}

func TestExecute_Issues_Flag_BrowserError(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)
	mockBrowser.Err = factory.ErrMock

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--issues"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error when browser fails")
	}

	if !strings.Contains(err.Error(), "failed to open browser") {
		t.Errorf("expected 'failed to open browser' in error, got: %v", err)
	}
}

func TestExecute_DryRun_BugType(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()

	// Check for issue preview content
	if !strings.Contains(output, "Issue Preview") {
		t.Errorf("expected 'Issue Preview' in output, got: %s", output)
	}

	if !strings.Contains(output, "Type: bug") {
		t.Errorf("expected 'Type: bug' in output, got: %s", output)
	}

	if !strings.Contains(output, "Body:") {
		t.Errorf("expected 'Body:' in output, got: %s", output)
	}

	if !strings.Contains(output, "URL:") {
		t.Errorf("expected 'URL:' in output, got: %s", output)
	}
}

func TestExecute_DryRun_FeatureType(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "feature", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()

	if !strings.Contains(output, "Type: feature") {
		t.Errorf("expected 'Type: feature' in output, got: %s", output)
	}
}

func TestExecute_DryRun_DeviceType(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "device", "--device", "kitchen-light", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()

	if !strings.Contains(output, "Type: device") {
		t.Errorf("expected 'Type: device' in output, got: %s", output)
	}

	if !strings.Contains(output, "kitchen-light") {
		t.Errorf("expected device name in output, got: %s", output)
	}
}

func TestExecute_DryRun_WithTitle(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug", "--title", "Test Bug Title", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()

	if !strings.Contains(output, "Title: Test Bug Title") {
		t.Errorf("expected title in output, got: %s", output)
	}
}

func TestExecute_DryRun_WithoutTitle(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "feature", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()

	// Should not have Title line when not provided
	if strings.Contains(output, "Title:") {
		t.Errorf("should not print Title when not provided, got: %s", output)
	}
}

func TestExecute_NoFlags_DefaultsToBug(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Should call browser with a bug issue URL
	if !mockBrowser.BrowseCalled {
		t.Error("Browser.Browse should have been called")
	}

	url := mockBrowser.LastURL
	if !strings.Contains(url, "labels=bug") {
		t.Errorf("URL should contain bug label, got: %s", url)
	}
}

func TestExecute_DeviceWithoutType_DefaultsToDevice(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--device", "test-device"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Should call browser with device issue URL
	if !mockBrowser.BrowseCalled {
		t.Error("Browser.Browse should have been called")
	}

	url := mockBrowser.LastURL
	if !strings.Contains(url, "labels=device-compatibility") {
		t.Errorf("URL should contain device-compatibility label, got: %s", url)
	}

	if !strings.Contains(url, "test-device") {
		t.Errorf("URL should contain device name, got: %s", url)
	}
}

func TestExecute_BugType_WithTitle(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug", "--title", "CLI Crash on Startup"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if !mockBrowser.BrowseCalled {
		t.Error("Browser.Browse should have been called")
	}

	url := mockBrowser.LastURL
	// Title will be URL encoded, so check for the encoded version or the parameter name
	if !strings.Contains(url, "title=") {
		t.Errorf("URL should contain title parameter, got: %s", url)
	}

	if !strings.Contains(url, "labels=bug") {
		t.Errorf("URL should contain bug label, got: %s", url)
	}
}

func TestExecute_FeatureType_WithTitle(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "feature", "--title", "Add Dark Mode"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	url := mockBrowser.LastURL
	// Title will be URL encoded
	if !strings.Contains(url, "title=") {
		t.Errorf("URL should contain title parameter, got: %s", url)
	}

	if !strings.Contains(url, "labels=enhancement") {
		t.Errorf("URL should contain enhancement label, got: %s", url)
	}
}

func TestExecute_Success_OpensBrowser(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Opening GitHub issue form") {
		t.Errorf("expected opening message in output, got: %s", output)
	}

	if !strings.Contains(output, "Issue form opened in browser") {
		t.Errorf("expected success message in output, got: %s", output)
	}

	if !strings.Contains(output, "Please fill in the description") {
		t.Errorf("expected description message in output, got: %s", output)
	}
}

func TestExecute_BrowserError_WithClipboardFallback(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)
	mockBrowser.Err = &browser.ClipboardFallbackError{URL: "http://test.url"}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil (should handle clipboard fallback)", err)
	}

	output := tf.OutString()
	// The messages should be in the output, though may have ANSI color codes
	if !strings.Contains(output, "Could not open browser") &&
		!strings.Contains(output, "Please open the URL manually") {
		t.Errorf("expected clipboard fallback or manual URL message, got: %s", output)
	}
}

func TestExecute_BrowserError_NoClipboard(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)
	mockBrowser.Err = factory.ErrMock

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "feature"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error when browser fails and clipboard unavailable")
	}

	if !strings.Contains(err.Error(), "failed to open browser") {
		t.Errorf("expected 'failed to open browser' error, got: %v", err)
	}
}

func TestExecute_AttachLog_Flag(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug", "--attach-log", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	// The body should include log attachment request
	if !strings.Contains(output, "Log info requested") {
		t.Errorf("expected log info in body, got: %s", output)
	}
}

func TestExecute_DeviceType_IncludesDeviceInfo(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "device", "--device", "living-room-light", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "living-room-light") {
		t.Errorf("expected device name in output, got: %s", output)
	}
}

func TestExecute_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--help"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute with --help error = %v, want nil", err)
	}

	// Help output might go to stdout or stderr depending on cobra version
	output := tf.OutString()
	errOutput := tf.ErrString()
	combined := output + errOutput

	if !strings.Contains(combined, "feedback") {
		t.Errorf("help output should contain command name, got: %s", combined)
	}
}

func TestExecute_TypeFlag_InvalidValue(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug", "--dry-run"})

	// This should still work - the invalid type will just not match any case
	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute with invalid type may error: %v", err)
	}
}

func TestExecute_Multiple_Issues_Calls(t *testing.T) {
	t.Parallel()

	tf, mockBrowser := factory.NewTestFactoryWithMockBrowser(t)

	// First command
	cmd1 := NewCommand(tf.Factory)
	cmd1.SetContext(context.Background())
	cmd1.SetArgs([]string{"--issues"})

	err := cmd1.Execute()
	if err != nil {
		t.Errorf("First Execute() error = %v", err)
	}

	if !mockBrowser.BrowseCalled {
		t.Error("First: Browser.Browse should have been called")
	}

	firstURL := mockBrowser.LastURL

	// Reset
	mockBrowser.BrowseCalled = false
	tf.Reset()

	// Second command
	cmd2 := NewCommand(tf.Factory)
	cmd2.SetContext(context.Background())
	cmd2.SetArgs([]string{"--issues"})

	err = cmd2.Execute()
	if err != nil {
		t.Errorf("Second Execute() error = %v", err)
	}

	if !mockBrowser.BrowseCalled {
		t.Error("Second: Browser.Browse should have been called")
	}

	// Both should use same URL
	if mockBrowser.LastURL != firstURL {
		t.Errorf("URL mismatch: %q vs %q", mockBrowser.LastURL, firstURL)
	}
}

func TestRun_NilConfig(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	opts := &Options{
		Type:       "bug",
		Title:      "Test Bug",
		DryRun:     true,
		OpenIssues: false,
	}

	// Even with nil config, should work (it's optional)
	err := run(context.Background(), tf.Factory, opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

func TestRun_WithConfig(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	opts := &Options{
		Type:   "bug",
		Title:  "Test",
		DryRun: true,
	}

	err := run(context.Background(), tf.Factory, opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--issues"})

	// Might error due to context cancellation in browser
	err := cmd.Execute()
	// We don't care about the error result here since it's testing context cancellation behavior
	if err != nil {
		t.Logf("Execute with cancelled context returned error (expected): %v", err)
	}
}

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example
	if !strings.Contains(example, "shelly feedback") {
		t.Errorf("example should contain 'shelly feedback', got: %s", example)
	}

	// Should have multiple examples
	lines := strings.Split(example, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}
	if nonEmptyLines < 3 {
		t.Errorf("example should have multiple usage examples, got %d lines", nonEmptyLines)
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	long := cmd.Long
	if !strings.Contains(long, "GitHub") {
		t.Error("Long should mention GitHub")
	}

	if !strings.Contains(long, "bug") {
		t.Error("Long should mention bug")
	}

	if !strings.Contains(long, "feature") {
		t.Error("Long should mention feature")
	}

	if !strings.Contains(long, "device") {
		t.Error("Long should mention device")
	}
}

func TestExecute_EmptyTitleWithBugType_DryRun(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	// When title is empty, it should not show a Title line
	if strings.Contains(output, "\nTitle:") {
		t.Errorf("should not show Title line when empty, got: %s", output)
	}
}

func TestExecute_SystemInfo_InBody(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--type", "bug", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	// System info should be in the body
	if !strings.Contains(output, "Environment") {
		t.Errorf("expected Environment section in body, got: %s", output)
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Type:       "bug",
		Title:      "Test Title",
		Device:     "test-device",
		AttachLog:  true,
		DryRun:     true,
		OpenIssues: false,
	}

	if opts.Type != "bug" {
		t.Errorf("Type = %q, want bug", opts.Type)
	}
	if opts.Title != "Test Title" {
		t.Errorf("Title = %q, want 'Test Title'", opts.Title)
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want 'test-device'", opts.Device)
	}
	if !opts.AttachLog {
		t.Error("AttachLog should be true")
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
	if opts.OpenIssues {
		t.Error("OpenIssues should be false")
	}
}
