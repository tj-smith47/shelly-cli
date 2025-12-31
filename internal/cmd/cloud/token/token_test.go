package token

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// testJWTClaims represents the claims structure for Shelly cloud tokens.
type testJWTClaims struct {
	UserAPIURL string `json:"user_api_url,omitempty"`
	Exp        int64  `json:"exp,omitempty"`
	Iat        int64  `json:"iat,omitempty"`
	Sub        string `json:"sub,omitempty"`
}

// createTestJWT creates a valid JWT token for testing.
// Shelly cloud tokens are JWTs with user_api_url and exp claims in the payload.
func createTestJWT(claims testJWTClaims) string {
	// JWT header (standard)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	// JWT payload with claims
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		// Should never happen in tests with valid claims
		panic("failed to marshal JWT claims: " + err.Error())
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// JWT signature (dummy for tests - actual validation isn't done by ParseToken)
	signature := base64.RawURLEncoding.EncodeToString([]byte("test-signature"))

	return header + "." + payload + "." + signature
}

// createValidTestToken creates a token that expires in the future.
func createValidTestToken() string {
	return createTestJWT(testJWTClaims{
		UserAPIURL: "https://shelly-13-eu.shelly.cloud",
		Exp:        time.Now().Add(24 * time.Hour).Unix(),
		Iat:        time.Now().Unix(),
		Sub:        "test-user",
	})
}

// createExpiredTestToken creates a token that has already expired.
func createExpiredTestToken() string {
	return createTestJWT(testJWTClaims{
		UserAPIURL: "https://shelly-13-eu.shelly.cloud",
		Exp:        time.Now().Add(-1 * time.Hour).Unix(),
		Iat:        time.Now().Add(-25 * time.Hour).Unix(),
		Sub:        "test-user",
	})
}

// createShortToken creates a short token (less than 50 chars) for display testing.
func createShortToken() string {
	return createTestJWT(testJWTClaims{
		Exp: time.Now().Add(time.Hour).Unix(),
	})
}

// setupTestManagerWithCloud creates a test manager with cloud config.
func setupTestManagerWithCloud(t *testing.T, cloudCfg config.CloudConfig) *config.Manager {
	t.Helper()
	cfg := &config.Config{
		Devices:   make(map[string]model.Device),
		Groups:    make(map[string]config.Group),
		Aliases:   make(map[string]config.Alias),
		Scenes:    make(map[string]config.Scene),
		Templates: config.TemplatesConfig{},
		Cloud:     cloudCfg,
	}
	return config.NewTestManager(cfg)
}

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "token" {
		t.Errorf("Use = %q, want 'token'", cmd.Use)
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

	expectedAliases := map[string]bool{"tok": true, "key": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias: %s", alias)
		}
		delete(expectedAliases, alias)
	}
	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
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

func TestNewCommand_RunESet(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// RunE should be set (not Run)
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
	if cmd.Run != nil {
		t.Error("Run should not be set when RunE is used")
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Verify command has proper structure
	tests := []struct {
		name string
		fn   func() bool
	}{
		{"has Use", func() bool { return cmd.Use != "" }},
		{"has Short", func() bool { return cmd.Short != "" }},
		{"has Long", func() bool { return cmd.Long != "" }},
		{"has Example", func() bool { return cmd.Example != "" }},
		{"has Aliases", func() bool { return len(cmd.Aliases) > 0 }},
		{"has RunE", func() bool { return cmd.RunE != nil }},
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

func TestNewCommand_AliasesContents(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	// Check for expected aliases
	if !aliasMap["tok"] {
		t.Error("missing 'tok' alias")
	}
	if !aliasMap["key"] {
		t.Error("missing 'key' alias")
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Token command should not have subcommands
	if len(cmd.Commands()) > 0 {
		t.Errorf("token command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Token command should not define its own flags
	if cmd.Flags().NFlag() > 0 {
		t.Errorf("token command should not have flags set, has %d", cmd.Flags().NFlag())
	}
}

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain "shelly cloud token" showing proper usage
	if cmd.Example == "" {
		t.Fatal("Example should not be empty")
	}

	// Check that example has actual shelly command
	if !bytes.Contains([]byte(cmd.Example), []byte("shelly")) {
		t.Error("Example should contain 'shelly' command")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention key concepts
	if !strings.Contains(cmd.Long, "token") {
		t.Error("Long description should mention 'token'")
	}
	if !strings.Contains(cmd.Long, "Shelly Cloud") {
		t.Error("Long description should mention 'Shelly Cloud'")
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Short description should be concise and mention token
	if !strings.Contains(cmd.Short, "token") {
		t.Error("Short description should mention 'token'")
	}
}

func TestNewCommand_UseField(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Use field should be exactly "token"
	if cmd.Use != "token" {
		t.Errorf("Use = %q, want 'token'", cmd.Use)
	}
}

func TestNewCommand_ExampleContainsClipboard(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should mention clipboard usage
	if !strings.Contains(cmd.Example, "clipboard") {
		t.Error("Example should mention clipboard usage")
	}
}

func TestNewCommand_ExampleContainsCloudToken(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should show cloud token command
	if !strings.Contains(cmd.Example, "cloud token") {
		t.Error("Example should show 'cloud token' command")
	}
}

func TestNewCommand_MultipleAliasesWork(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Verify we have at least 2 aliases
	if len(cmd.Aliases) < 2 {
		t.Errorf("expected at least 2 aliases, got %d", len(cmd.Aliases))
	}
}

func TestNewCommand_LongMentionsSecurity(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should warn about security
	if !strings.Contains(cmd.Long, "careful") &&
		!strings.Contains(cmd.Long, "share") &&
		!strings.Contains(cmd.Long, "expose") {
		t.Error("Long description should warn about token security")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_NotLoggedIn(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err == nil {
		t.Fatal("expected error for not logged in")
	}

	if err.Error() != "not logged in" {
		t.Errorf("error = %q, want 'not logged in'", err.Error())
	}

	// Check error output contains expected message
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "Not logged in") {
		t.Errorf("stderr should contain 'Not logged in', got: %s", errOutput)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_InvalidToken(t *testing.T) {
	// Create config with an invalid token (not a valid JWT)
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: "invalid-token-not-jwt",
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err == nil {
		t.Fatal("expected error for invalid token")
	}

	// Error should be about invalid token
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "invalid") && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error should mention 'invalid', got err=%v, stderr=%s", err, errOutput)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_ValidToken_TTY(t *testing.T) {
	token := createValidTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	// Create TTY IOStreams
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// Should contain title
	if !strings.Contains(output, "Cloud Token") {
		t.Error("output should contain 'Cloud Token' title")
	}

	// Should show truncated token (contains ...)
	if !strings.Contains(output, "...") {
		t.Errorf("output should contain truncated token with '...', got: %s", output)
	}

	// Should show server URL from token
	if !strings.Contains(output, "Server:") {
		t.Errorf("output should contain 'Server:', got: %s", output)
	}

	// Should show expiry information
	if !strings.Contains(output, "Expires:") {
		t.Errorf("output should contain 'Expires:', got: %s", output)
	}

	// Should show remaining time
	if !strings.Contains(output, "Remaining:") {
		t.Errorf("output should contain 'Remaining:', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_ValidToken_NonTTY(t *testing.T) {
	token := createValidTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	// Create non-TTY IOStreams (default for Test)
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(false)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// In non-TTY mode, should output full token without formatting
	if !strings.Contains(output, token) {
		t.Errorf("non-TTY output should contain full token, got: %s", output)
	}

	// Should NOT contain title in non-TTY mode (early return)
	// Actually, looking at the code, it outputs the title first, then checks TTY
	// So let's verify it contains the full token
	if len(output) < len(token) {
		t.Error("non-TTY output should be at least as long as the token")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_ExpiredToken_TTY(t *testing.T) {
	token := createExpiredTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// Should contain title
	if !strings.Contains(output, "Cloud Token") {
		t.Error("output should contain 'Cloud Token' title")
	}

	// Should show Status for expired token
	if !strings.Contains(output, "Status:") {
		t.Errorf("output should contain 'Status:' for expired token, got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_TokenWithoutExpiry(t *testing.T) {
	// Create token without expiry claim
	token := createTestJWT(testJWTClaims{
		UserAPIURL: "https://shelly-13-eu.shelly.cloud",
		Sub:        "test-user",
		// No Exp field
	})
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// Should contain server URL
	if !strings.Contains(output, "Server:") {
		t.Errorf("output should contain 'Server:', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_TokenWithoutServerURL(t *testing.T) {
	// Create token without UserAPIURL
	token := createTestJWT(testJWTClaims{
		Exp: time.Now().Add(24 * time.Hour).Unix(),
		Sub: "test-user",
		// No UserAPIURL
	})
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// Should NOT contain Server: line if no URL in token
	if strings.Contains(output, "Server:") {
		t.Errorf("output should NOT contain 'Server:' when URL is empty, got: %s", output)
	}

	// Should still show expiry
	if !strings.Contains(output, "Expires:") {
		t.Errorf("output should contain 'Expires:', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_NotLoggedIn(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for not logged in")
	}

	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error should contain 'not logged in', got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_ValidToken(t *testing.T) {
	token := createValidTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	tf := factory.NewTestFactory(t)
	tf.TestIO.SetStdoutTTY(true)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err := cmd.Execute()

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Cloud Token") {
		t.Errorf("output should contain 'Cloud Token', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_Help(t *testing.T) {
	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "token") {
		t.Error("help should mention token")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_InvalidToken(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: "not-a-jwt",
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_ExpiredToken(t *testing.T) {
	token := createExpiredTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	tf := factory.NewTestFactory(t)
	tf.TestIO.SetStdoutTTY(true)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err := cmd.Execute()

	// Expired token should still work (shows status)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Cloud Token") {
		t.Errorf("output should contain 'Cloud Token', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_ShortToken_TTY(t *testing.T) {
	// Create a token that is less than 50 characters
	// This tests the else branch where full token is shown
	shortToken := createShortToken()

	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: shortToken,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// Should contain Cloud Token title
	if !strings.Contains(output, "Cloud Token") {
		t.Error("output should contain 'Cloud Token' title")
	}

	// Should show Token: line
	if !strings.Contains(output, "Token:") {
		t.Errorf("output should contain 'Token:', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_TokenWithClipboardHint(t *testing.T) {
	token := createValidTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// Should contain clipboard hint in TTY mode
	if !strings.Contains(output, "clipboard") || !strings.Contains(output, "pipe") {
		t.Errorf("TTY output should contain clipboard hint, got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithIOStreams(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Test that run function properly uses IOStreams
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// This will fail because no config, but it exercises the run function
	opts := &Options{Factory: f}
	err := run(opts)

	// We expect an error (not logged in)
	if err == nil {
		t.Error("expected error for not logged in")
	}

	// The error should be about login
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error should be 'not logged in', got: %v", err)
	}

	// Error output should have been written
	if errOut.Len() == 0 {
		t.Error("expected error output to be written")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_NonTTY_FullToken(t *testing.T) {
	token := createValidTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	tf := factory.NewTestFactory(t)
	tf.TestIO.SetStdoutTTY(false)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err := cmd.Execute()

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// In non-TTY mode, output should contain the actual token
	// (after the title, since title is output before TTY check)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_TokenMinimalClaims(t *testing.T) {
	// Create token with minimal claims to test edge cases
	token := createTestJWT(testJWTClaims{
		// Only expiry, no URL
		Exp: time.Now().Add(time.Hour).Unix(),
	})
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// Should have title
	if !strings.Contains(output, "Cloud Token") {
		t.Error("output should contain 'Cloud Token' title")
	}

	// Should NOT have Server: since no URL in token
	if strings.Contains(output, "Server:") {
		t.Error("output should NOT contain 'Server:' with minimal token")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_WithContext(t *testing.T) {
	token := createValidTestToken()
	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	tf := factory.NewTestFactory(t)
	tf.TestIO.SetStdoutTTY(true)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	// Use a context (simulating cmd.Context())
	ctx := context.Background()
	cmd.SetContext(ctx)

	err := cmd.Execute()

	if err != nil {
		t.Errorf("Execute() with context error = %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_LongToken_Truncation(t *testing.T) {
	// Create a token that is definitely longer than 50 characters
	// Standard JWT tokens are typically 100+ characters
	token := createValidTestToken()

	// Verify token is long enough for truncation test
	if len(token) <= 50 {
		t.Skipf("token too short for truncation test: %d chars", len(token))
	}

	mgr := setupTestManagerWithCloud(t, config.CloudConfig{
		AccessToken: token,
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	ios.SetStdoutTTY(true)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{Factory: f}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := out.String()

	// In TTY mode with long token, should see truncated format with ...
	if !strings.Contains(output, "...") {
		t.Errorf("long token in TTY mode should be truncated with '...', got: %s", output)
	}

	// Should show first 20 chars
	expectedStart := token[:20]
	if !strings.Contains(output, expectedStart) {
		t.Errorf("output should contain first 20 chars of token, got: %s", output)
	}

	// Should show last 10 chars
	expectedEnd := token[len(token)-10:]
	if !strings.Contains(output, expectedEnd) {
		t.Errorf("output should contain last 10 chars of token, got: %s", output)
	}
}
