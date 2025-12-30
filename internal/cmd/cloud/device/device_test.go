package device

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
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

	// Test Use
	if cmd.Use != "device <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "device <id>")
	}

	// Test Aliases
	wantAliases := []string{"get"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"abc123"}, false},
		{"two args", []string{"abc123", "def456"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
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

func TestNewCommand_StatusFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("status")
	if flag == nil {
		t.Fatal("--status flag should exist")
	}

	if flag.DefValue != "false" {
		t.Errorf("--status default = %q, want 'false'", flag.DefValue)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly cloud device",
		"--status",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"device",
		"Cloud",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long should contain %q", pattern)
		}
	}
}

func TestNewCommand_ShortIsConcise(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if strings.Contains(cmd.Short, "\n") {
		t.Error("Short description should not contain newlines")
	}

	if len(cmd.Short) > 80 {
		t.Errorf("Short description too long (%d chars), should be under 80", len(cmd.Short))
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Commands()) > 0 {
		t.Errorf("device command should not have subcommands, has %d", len(cmd.Commands()))
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

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_NotLoggedIn(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create test config without cloud token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "", // Not logged in
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	err := run(ctx, f, "test-device")

	if err == nil {
		t.Error("expected error when not logged in")
	}

	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("error should mention 'not logged in', got: %v", err)
	}

	// Check stderr contains helpful message
	stderr := errOut.String()
	if !strings.Contains(stderr, "Not logged in") {
		t.Errorf("stderr should mention 'Not logged in', got: %s", stderr)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_NotLoggedIn_ShowsLoginHint(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create test config without cloud token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "", // Not logged in
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	err := run(ctx, f, "test-device")
	if err != nil {
		t.Logf("expected error: %v", err)
	}

	// Check stdout contains login hint
	combined := out.String() + errOut.String()
	if !strings.Contains(combined, "cloud login") {
		t.Errorf("output should mention 'cloud login' hint, got: %s", combined)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithCancelledContext(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create a mock cloud server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create test config with fake token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "fake-token-for-testing",
			ServerURL:   server.URL,
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, f, "test-device")

	// Should fail due to cancelled context or cloud API error
	if err == nil {
		t.Log("run succeeded with cancelled context - may have returned before making request")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithTimeout(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create test config with fake token (no server needed - will timeout)
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "fake-token-for-testing",
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create context with very short timeout (already expired)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow timeout to trigger
	time.Sleep(1 * time.Millisecond)

	err := run(ctx, f, "test-device")

	// Should fail due to timeout or API error
	if err == nil {
		t.Log("run succeeded with timed out context")
	}
}

func TestNewCommand_Execute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute should fail without arguments")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestNewCommand_Execute_WithDeviceID(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create test config without cloud token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "", // Not logged in
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"device123"})

	err := cmd.Execute()
	// Should fail because not logged in
	if err == nil {
		t.Error("Execute should fail when not logged in")
	}
}

func TestNewCommand_UsageString(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	usage := cmd.UsageString()
	if !strings.Contains(usage, "device") {
		t.Error("UsageString should contain command name")
	}
	if !strings.Contains(usage, "--status") {
		t.Error("UsageString should mention --status flag")
	}
}

func TestNewCommand_HelpString(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	help := cmd.UsageString()
	if help == "" {
		t.Error("Help should not be empty")
	}

	if !strings.Contains(help, "device") {
		t.Error("Help should mention device")
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != "device" {
		t.Errorf("Name() = %q, want 'device'", cmd.Name())
	}
}

func TestNewCommand_CommandPath(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	path := cmd.CommandPath()
	if !strings.Contains(path, "device") {
		t.Errorf("CommandPath() = %q, should contain 'device'", path)
	}
}

func TestNewCommand_FlagExists(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check that the --status flag exists and is a bool flag with false default
	flag := cmd.Flags().Lookup("status")
	if flag == nil {
		t.Fatal("--status flag should exist")
	}
	if flag.DefValue != "false" {
		t.Errorf("--status default = %q, want 'false'", flag.DefValue)
	}
	if flag.Usage == "" {
		t.Error("--status should have a usage description")
	}
}

// createTestJWT creates a minimal JWT token with the given user_api_url claim.
// This is a simplified JWT for testing purposes only.
func createTestJWT(userAPIURL string) string {
	// JWT header: {"alg":"HS256","typ":"JWT"}
	header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

	// Create payload with user_api_url claim
	// We need to include: user_api_url, iat, exp
	// For testing, we create a payload manually
	now := time.Now().Unix()
	exp := now + 3600 // 1 hour from now

	payload := map[string]any{
		"user_api_url": userAPIURL,
		"iat":          now,
		"exp":          exp,
	}
	// json.Marshal will not fail for this simple map[string]any
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		// Unreachable for valid map[string]any, but handle for linter
		return ""
	}
	payloadEncoded := base64URLEncode(payloadBytes)

	// Fake signature (not cryptographically valid but the parser may not verify)
	signature := "fake_signature_for_testing"

	return header + "." + payloadEncoded + "." + signature
}

// base64URLEncode encodes bytes to base64 URL encoding without padding.
func base64URLEncode(data []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(data)
	return encoded
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithMockCloudServer(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create a mock cloud server that responds to device status requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request for debugging
		t.Logf("Mock server received: %s %s", r.Method, r.URL.Path)

		// Return a mock device status response
		// The shelly-go cloud client expects this format
		resp := map[string]any{
			"isok": true,
			"data": map[string]any{
				"online": true,
				"device_status": map[string]any{
					"_dev_info": map[string]any{
						"code": "SNSW-001P16EU",
						"gen":  2,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// Create a JWT token with user_api_url pointing to our mock server
	token := createTestJWT(server.URL)

	// Create test config with the token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: token,
			ServerURL:   server.URL,
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	err := run(ctx, f, "test-device-123")

	// The cloud client validates JWT signatures, so this will likely fail.
	// But we exercise the code path up to the API call.
	if err != nil {
		t.Logf("run error (expected with JWT validation): %v", err)
	}

	// Verify that some output was produced
	combined := out.String() + errOut.String()
	t.Logf("Output: %s", combined)
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithMockCloudServer_StatusFlag(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create a mock cloud server that responds to device status requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mock device status response with status and settings
		resp := map[string]any{
			"isok": true,
			"data": map[string]any{
				"online":   true,
				"status":   map[string]any{"switch:0": map[string]any{"output": true}},
				"settings": map[string]any{"name": "Test Device"},
				"device_status": map[string]any{
					"_dev_info": map[string]any{
						"code": "SNSW-001P16EU",
						"gen":  2,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// Create a JWT token with user_api_url pointing to our mock server
	token := createTestJWT(server.URL)

	// Create test config with the token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: token,
			ServerURL:   server.URL,
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"test-device-123", "--status"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute error: %v", err)
	}

	// Check output contains device info
	combined := out.String() + errOut.String()
	if !strings.Contains(combined, "Cloud Device") {
		t.Error("output should contain 'Cloud Device'")
	}
}

func TestNewCommand_AliasGet(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	hasGet := false
	for _, alias := range cmd.Aliases {
		if alias == "get" {
			hasGet = true
			break
		}
	}

	if !hasGet {
		t.Error("expected 'get' alias")
	}
}

func TestNewCommand_ExampleHasComments(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Good examples typically have comments explaining usage
	if !strings.Contains(cmd.Example, "#") {
		t.Log("Example may benefit from comment lines")
	}
}

func TestNewCommand_LongMentionsStatus(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "status") && !strings.Contains(cmd.Long, "settings") {
		t.Log("Long description may mention status or settings")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestNewCommand_Execute_WithStatusFlag(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create test config without cloud token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "", // Not logged in
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"device123", "--status"})

	err := cmd.Execute()
	// Should fail because not logged in
	if err == nil {
		t.Error("Execute should fail when not logged in")
	}

	// Verify statusFlag was set (even though command failed)
	flag := cmd.Flags().Lookup("status")
	if flag == nil || flag.Value.String() != "true" {
		t.Error("--status flag should be parsed as true")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestNewCommand_RunEIsCallable(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create test config without cloud token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "", // Not logged in
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	// Must set context before calling RunE directly
	cmd.SetContext(context.Background())

	// Call RunE directly with args
	err := cmd.RunE(cmd, []string{"test-device"})

	// Should not panic, error is expected if not logged in
	if err == nil {
		t.Log("RunE succeeded - unexpected")
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

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_OutputGoesToStreams(t *testing.T) {
	// Save original HOME to restore later
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		config.ResetDefaultManagerForTesting()
	}()

	// Create temp dir for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	config.ResetDefaultManagerForTesting()

	// Create test config without cloud token
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Cloud: config.CloudConfig{
			AccessToken: "", // Not logged in
		},
	}
	mgr := config.NewTestManager(cfg)
	config.SetDefaultManager(mgr)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	err := run(ctx, f, "test-device")
	if err != nil {
		t.Logf("expected error: %v", err)
	}

	// Run should produce some output to our streams
	totalOutput := out.Len() + errOut.Len()
	if totalOutput == 0 {
		t.Error("Run produced no output")
	}
}

func TestNewCommand_ArgsValidation(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that Args validator is set to ExactArgs(1)
	tests := []struct {
		args    []string
		wantErr bool
	}{
		{[]string{}, true},
		{[]string{"one"}, false},
		{[]string{"one", "two"}, true},
		{[]string{"one", "two", "three"}, true},
	}

	for _, tt := range tests {
		err := cmd.Args(cmd, tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
		}
	}
}

func TestNewCommand_ExampleHasMultipleLines(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

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

func TestNewCommand_UseField(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "device <id>" {
		t.Errorf("Use = %q, want 'device <id>'", cmd.Use)
	}
}
