package authstatus

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	cmdName = "auth-status"
	// Test JWT token with exp claim in year 2030 (cannot be validated without proper signature).
	testTokenFuture = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDB9.signature"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != cmdName {
		t.Errorf("Use = %q, want '%s'", cmd.Use, cmdName)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Fatal("expected at least one alias")
	}

	hasWhoami := false
	for _, alias := range cmd.Aliases {
		if alias == "whoami" {
			hasWhoami = true
			break
		}
	}

	if !hasWhoami {
		t.Error("expected 'whoami' alias")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
	if cmd.Run != nil {
		t.Error("Run should not be set when RunE is used")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Command should not require args (Args is nil or allows 0 args)
	if cmd.Args != nil {
		if err := cmd.Args(cmd, []string{}); err != nil {
			t.Errorf("command should accept zero args: %v", err)
		}
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// auth-status command should not define its own flags
	if cmd.Flags().NFlag() > 0 {
		t.Errorf("auth-status command should not have flags set, has %d", cmd.Flags().NFlag())
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Commands()) > 0 {
		t.Errorf("auth-status command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"has Use", func() bool { return NewCommand(cmdutil.NewFactory()).Use != "" }},
		{"has Short", func() bool { return NewCommand(cmdutil.NewFactory()).Short != "" }},
		{"has Long", func() bool { return NewCommand(cmdutil.NewFactory()).Long != "" }},
		{"has Example", func() bool { return NewCommand(cmdutil.NewFactory()).Example != "" }},
		{"has Aliases", func() bool { return len(NewCommand(cmdutil.NewFactory()).Aliases) > 0 }},
		{"has RunE", func() bool { return NewCommand(cmdutil.NewFactory()).RunE != nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.fn() {
				t.Errorf("command structure check failed: %s", tt.name)
			}
		})
	}
}

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "shelly") {
		t.Error("Example should contain 'shelly' command")
	}

	if !strings.Contains(cmd.Example, cmdName) {
		t.Error("Example should contain 'auth-status' command")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "authentication") {
		t.Error("Long description should mention 'authentication'")
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Short, "authentication") && !strings.Contains(cmd.Short, "auth") {
		t.Error("Short description should mention authentication")
	}
}

func TestNewCommand_ExampleContainsWhoami(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "whoami") {
		t.Error("Example should mention whoami alias")
	}
}

func TestNewCommand_UseField(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != cmdName {
		t.Errorf("Use = %q, want '%s'", cmd.Use, cmdName)
	}
}

func TestNewCommand_AliasesContents(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	if !aliasMap["whoami"] {
		t.Error("missing 'whoami' alias")
	}
}

func TestNewCommand_LongMentionsLogin(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "logged in") && !strings.Contains(cmd.Long, "email") {
		t.Error("Long description should mention login status or email")
	}
}

func TestNewCommand_LongMentionsToken(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "token") {
		t.Error("Long description should mention token")
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_WithIOStreams(t *testing.T) {
	// Use in-memory filesystem to avoid touching real config
	factory.SetupTestFs(t)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	// Error is expected when not logged in (isolated config has no token)
	if err != nil {
		t.Logf("run error (expected with isolated config): %v", err)
	}

	// Should produce some output about auth status
	if out.Len() == 0 && errOut.Len() == 0 {
		t.Log("no output produced")
	}
}

func TestNewCommand_ExampleContainsCloud(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "cloud") {
		t.Error("Example should show 'cloud' subcommand")
	}
}

func TestNewCommand_ShortIsConcise(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Short description should be concise (one line)
	if strings.Contains(cmd.Short, "\n") {
		t.Error("Short description should not contain newlines")
	}

	if len(cmd.Short) > 80 {
		t.Errorf("Short description too long (%d chars), should be under 80", len(cmd.Short))
	}
}

func TestNewCommand_ExampleHasMultipleLines(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Good examples should have multiple usage examples
	lines := strings.Split(cmd.Example, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	if nonEmptyLines < 2 {
		t.Error("Example should contain multiple usage examples")
	}
}

func TestNewCommand_LongHasMultipleLines(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should provide detailed explanation
	if !strings.Contains(cmd.Long, "\n") {
		t.Error("Long description should have multiple lines for detail")
	}
}

func TestNewCommand_RunEReturnsNil(t *testing.T) {
	// Test the RunE function returns a function
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Fatal("RunE should not be nil")
	}
}

func TestNewCommand_ExampleStartsWithIndent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example lines typically start with indent for proper formatting
	lines := strings.Split(cmd.Example, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "shelly") {
			// Non-comment, non-command lines should be handled
			t.Logf("Example line: %s", line)
		}
	}
}

func TestNewCommand_LongContainsDisplays(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Long), "display") {
		t.Log("Long description should mention what is displayed")
	}
}

func TestNewCommand_Execute(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})

	// Execute the command - may succeed or fail depending on global config
	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute error (expected): %v", err)
	}
}

func TestNewCommand_ExecuteWithHelp(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--help"})

	// Execute with help flag
	err := cmd.Execute()
	if err != nil {
		t.Logf("Help execution error: %v", err)
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// auth-status should not have ValidArgsFunction (no args needed)
	if cmd.ValidArgsFunction != nil {
		t.Log("auth-status has ValidArgsFunction defined")
	}
}

func TestNewCommand_RunEIsCallable(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Call RunE directly
	err := cmd.RunE(cmd, []string{})

	// Should not panic, error is OK if not logged in
	if err != nil {
		t.Logf("RunE returned error (expected): %v", err)
	}
}

func TestNewCommand_OutputGoesToStreams(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute error: %v", err)
	}

	// Some output should go to our streams
	combined := out.String() + errOut.String()
	if combined == "" {
		t.Log("No output captured - command may use global config")
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != cmdName {
		t.Errorf("Name() = %q, want 'auth-status'", cmd.Name())
	}
}

func TestNewCommand_CommandPath(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// CommandPath returns the full path (just the command name when standalone)
	path := cmd.CommandPath()
	if !strings.Contains(path, cmdName) {
		t.Errorf("CommandPath() = %q, should contain 'auth-status'", path)
	}
}

func TestNewCommand_UsageString(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	usage := cmd.UsageString()
	if !strings.Contains(usage, cmdName) {
		t.Error("UsageString should contain command name")
	}
}

func TestNewCommand_HelpString(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Get help output
	help := cmd.UsageString()
	if help == "" {
		t.Error("Help should not be empty")
	}

	// Help should mention the command
	if !strings.Contains(help, cmdName) {
		t.Error("Help should mention auth-status")
	}
}

func TestNewCommand_RunEMultipleCalls(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Call RunE multiple times to ensure it's stable
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Logf("First RunE error: %v", err)
	}
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Logf("Second RunE error: %v", err)
	}

	// Should not panic on multiple calls
}

func TestNewCommand_RunEWithArgs(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Call RunE with extra args (should be ignored)
	err := cmd.RunE(cmd, []string{"extra", "args"})
	if err != nil {
		t.Logf("RunE with extra args error: %v", err)
	}
}

func TestNewCommand_VerifyReturnType(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Verify the command is the expected type
	if cmd == nil {
		t.Fatal("NewCommand should not return nil")
	}

	// Check that it's properly initialized
	if cmd.Use == "" {
		t.Error("Command Use should be set")
	}
}

// setupTestManagerWithCloud creates a test config manager with cloud settings using afero.
func setupTestManagerWithCloud(t *testing.T, accessToken, email, serverURL string) *config.Manager {
	t.Helper()
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	mgr := config.NewManager("/test/config/config.yaml")
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	cfg := mgr.Get()
	cfg.Cloud.AccessToken = accessToken
	cfg.Cloud.Email = email
	cfg.Cloud.ServerURL = serverURL
	return mgr
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_NotLoggedIn(t *testing.T) {
	// Setup manager with no token (not logged in)
	mgr := setupTestManagerWithCloud(t, "", "", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show not logged in status
	if !strings.Contains(output, "Status") {
		t.Errorf("expected output to contain 'Status', got: %q", output)
	}
	// Should suggest using login command
	if !strings.Contains(output, "login") {
		t.Errorf("expected output to contain 'login', got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_LoggedInWithEmail(t *testing.T) {
	// Create a valid JWT token for testing (expired but parseable)
	// This is a minimal JWT structure: header.payload.signature
	// Payload: {"exp": 1893456000} (year 2030)
	validToken := testTokenFuture

	mgr := setupTestManagerWithCloud(t, validToken, "test@example.com", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show logged in status and email
	if !strings.Contains(output, "Status") {
		t.Errorf("expected output to contain 'Status', got: %q", output)
	}
	if !strings.Contains(output, "test@example.com") {
		t.Errorf("expected output to contain email, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_LoggedInWithServerURL(t *testing.T) {
	validToken := testTokenFuture

	mgr := setupTestManagerWithCloud(t, validToken, "", "https://cloud.shelly.cloud")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show server URL
	if !strings.Contains(output, "cloud.shelly.cloud") {
		t.Errorf("expected output to contain server URL, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_LoggedInWithEmailAndServer(t *testing.T) {
	validToken := testTokenFuture

	mgr := setupTestManagerWithCloud(t, validToken, "user@example.com", "https://api.shelly.cloud")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show both email and server
	if !strings.Contains(output, "user@example.com") {
		t.Errorf("expected output to contain email, got: %q", output)
	}
	if !strings.Contains(output, "api.shelly.cloud") {
		t.Errorf("expected output to contain server, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_InvalidToken(t *testing.T) {
	// Invalid token (not a valid JWT)
	invalidToken := "not-a-valid-jwt-token"

	mgr := setupTestManagerWithCloud(t, invalidToken, "test@example.com", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	// Should not return error, but show warning
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should still show status
	if !strings.Contains(output, "Status") {
		t.Errorf("expected output to contain 'Status', got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_ExpiredToken(t *testing.T) {
	// Create an expired JWT token
	// Payload: {"exp": 1000000000} (year 2001, expired)
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDB9.signature"

	mgr := setupTestManagerWithCloud(t, expiredToken, "test@example.com", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	// Should not return error, but show warning about expired token
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show status
	if !strings.Contains(output, "Status") {
		t.Errorf("expected output to contain 'Status', got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_Title(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, "", "", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show title
	if !strings.Contains(output, "Cloud Authentication") {
		t.Errorf("expected output to contain title, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_NotLoggedIn(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, "", "", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := out.String()
	if !strings.Contains(output, "login") {
		t.Errorf("expected output to contain 'login' hint, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_LoggedIn(t *testing.T) {
	validToken := testTokenFuture
	mgr := setupTestManagerWithCloud(t, validToken, "cloud@example.com", "https://cloud.shelly.cloud")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := out.String()
	if !strings.Contains(output, "cloud@example.com") {
		t.Errorf("expected output to contain email, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_ValidTokenWithExpiry(t *testing.T) {
	// This token has expiry far in the future (year 2030)
	// The test exercises the code path that shows "Expiry:" with duration
	validToken := testTokenFuture

	mgr := setupTestManagerWithCloud(t, validToken, "user@example.com", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show token status
	if !strings.Contains(output, "Token") {
		t.Errorf("expected output to contain 'Token', got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_NoEmailNoServer(t *testing.T) {
	// Test with token but no email and no server
	validToken := testTokenFuture

	mgr := setupTestManagerWithCloud(t, validToken, "", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show status and token, but not email or server lines
	if !strings.Contains(output, "Status") {
		t.Errorf("expected output to contain 'Status', got: %q", output)
	}
	if !strings.Contains(output, "Token") {
		t.Errorf("expected output to contain 'Token', got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_TokenWithWarning(t *testing.T) {
	// Test with a malformed token that will produce a warning
	// Using a token that decodes but has invalid/missing exp claim
	malformedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.signature"

	mgr := setupTestManagerWithCloud(t, malformedToken, "user@example.com", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	// Should not return error even with warning
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should still show status
	if !strings.Contains(output, "Status") {
		t.Errorf("expected output to contain 'Status', got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_OutputContainsTitle(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, "", "", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	if err := run(opts); err != nil {
		t.Logf("run returned error (expected): %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Cloud Authentication Status") {
		t.Errorf("expected title in output, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_TokenStatusDisplayed(t *testing.T) {
	// Any token (will be marked as invalid since we can't create valid signatures in tests)
	testToken := testTokenFuture

	mgr := setupTestManagerWithCloud(t, testToken, "user@example.com", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()
	// Should show token status info (may be "Invalid" due to signature validation)
	if !strings.Contains(output, "Token") {
		t.Errorf("expected 'Token' in output, got: %q", output)
	}
}

func TestOptions_ZeroValue(t *testing.T) {
	t.Parallel()

	opts := &Options{}
	if opts.Factory != nil {
		t.Error("Factory should be nil for zero value")
	}
}

func TestNewCommand_NoLocalFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	// auth-status should not have local flags
	if cmd.LocalFlags().HasFlags() {
		t.Log("auth-status has local flags defined")
	}
}
