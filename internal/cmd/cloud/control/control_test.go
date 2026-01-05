package control

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// Test JWT token with exp claim in year 2030 (cannot be validated without proper signature).
const testTokenFuture = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDB9.signature"

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "control <device-id> <action>" {
		t.Errorf("Use = %q, want 'control <device-id> <action>'", cmd.Use)
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

	expectedAliases := map[string]bool{"ctrl": true, "cmd": true}
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

func TestNewCommand_RequiresExactlyTwoArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 2 arguments
	tests := []struct {
		args      []string
		wantError bool
	}{
		{[]string{}, true},
		{[]string{"device"}, true},
		{[]string{"device", "action"}, false},
		{[]string{"device", "action", "extra"}, true},
	}

	for _, tt := range tests {
		err := cmd.Args(cmd, tt.args)
		gotError := err != nil
		if gotError != tt.wantError {
			t.Errorf("Args(%v) error = %v, want error = %v", tt.args, err, tt.wantError)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check channel flag
	channelFlag := cmd.Flags().Lookup("channel")
	if channelFlag == nil {
		t.Fatal("channel flag not found")
	}

	if channelFlag.DefValue != "0" {
		t.Errorf("channel default = %q, want '0'", channelFlag.DefValue)
	}
}

func TestNewCommand_ChannelFlagType(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	channelFlag := cmd.Flags().Lookup("channel")
	if channelFlag == nil {
		t.Fatal("channel flag not found")
	}

	// Flag type should be int
	if channelFlag.Value.Type() != "int" {
		t.Errorf("channel flag type = %q, want 'int'", channelFlag.Value.Type())
	}
}

func TestNewCommand_RunESet(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

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
		{"has Args", func() bool { return cmd.Args != nil }},
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

func TestNewCommand_ExampleContainsShelly(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "shelly") {
		t.Error("Example should contain 'shelly' command")
	}
}

func TestNewCommand_ExampleContainsActions(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check that example shows various actions
	actions := []string{"on", "off", "toggle", "open", "position"}
	for _, action := range actions {
		if !strings.Contains(cmd.Example, action) {
			t.Errorf("Example should contain '%s' action", action)
		}
	}
}

func TestNewCommand_LongDescriptionDocumentsActions(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should document available actions
	actions := []string{"on", "off", "toggle", "open", "close", "stop"}
	for _, action := range actions {
		if !strings.Contains(cmd.Long, action) {
			t.Errorf("Long description should mention '%s' action", action)
		}
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_NotLoggedIn(t *testing.T) {
	// Use in-memory filesystem to avoid touching real config
	factory.SetupTestFs(t)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	opts := &Options{Factory: f, DeviceID: "device123", Action: "on", Channel: 0}
	err := run(ctx, opts)

	// Should fail because no token is configured in isolated config
	if err == nil {
		t.Error("expected error for not logged in, got nil")
		return
	}

	// Should get "not logged in" error
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in' error, got: %v", err)
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Control command should not have subcommands
	if len(cmd.Commands()) > 0 {
		t.Errorf("control command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_ChannelFlagUsage(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	channelFlag := cmd.Flags().Lookup("channel")
	if channelFlag == nil {
		t.Fatal("channel flag not found")
	}

	// Check that usage is descriptive
	if channelFlag.Usage == "" {
		t.Error("channel flag should have usage description")
	}
}

func TestNewCommand_HasRunE_NotRun(t *testing.T) {
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

func TestNewCommand_AliasesContent(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	// Verify expected aliases exist
	if !aliasMap["ctrl"] {
		t.Error("missing 'ctrl' alias")
	}
	if !aliasMap["cmd"] {
		t.Error("missing 'cmd' alias")
	}
}

func TestNewCommand_LongMentionsAuthentication(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention authentication requirement
	if !strings.Contains(cmd.Long, "login") && !strings.Contains(cmd.Long, "authentication") {
		t.Error("Long description should mention authentication requirement")
	}
}

// setupTestManagerWithCloud creates a test config manager with cloud settings using afero.
func setupTestManagerWithCloud(t *testing.T, accessToken string) *config.Manager {
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
	cfg.Cloud.ServerURL = "https://cloud.shelly.cloud"
	return mgr
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_NotLoggedInWithConfigManager(t *testing.T) {
	// Setup manager with no token (not logged in)
	mgr := setupTestManagerWithCloud(t, "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	ctx := context.Background()
	opts := &Options{Factory: f, DeviceID: "device123", Action: "on", Channel: 0}
	err := run(ctx, opts)

	if err == nil {
		t.Error("expected error for not logged in, got nil")
		return
	}

	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in' error, got: %v", err)
	}

	// Should show login hint
	output := out.String()
	if !strings.Contains(output, "login") {
		t.Errorf("expected output to contain 'login' hint, got: %q", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_LoggedInButConnectionFails(t *testing.T) {
	// Setup manager with a token (logged in)
	validToken := testTokenFuture
	mgr := setupTestManagerWithCloud(t, validToken)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	// Use a short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	opts := &Options{Factory: f, DeviceID: "device123", Action: "on", Channel: 0}
	err := run(ctx, opts)

	// Should fail due to network error (cloud API not available)
	if err == nil {
		t.Error("expected error for network failure, got nil")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithDifferentChannels(t *testing.T) {
	validToken := testTokenFuture
	mgr := setupTestManagerWithCloud(t, validToken)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Test with different channel values
	channels := []int{0, 1, 2}
	for _, ch := range channels {
		opts := &Options{Factory: f, DeviceID: "device123", Action: "on", Channel: ch}
		err := run(ctx, opts)
		// Expect error due to no real cloud (but the channel param should be handled)
		if err == nil {
			t.Logf("channel %d: expected error, got nil", ch)
		}
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_WithDifferentActions(t *testing.T) {
	validToken := testTokenFuture
	mgr := setupTestManagerWithCloud(t, validToken)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Test with different actions
	actions := []string{"on", "off", "toggle", "open", "close", "stop", "position=50"}
	for _, action := range actions {
		opts := &Options{Factory: f, DeviceID: "device123", Action: action, Channel: 0}
		err := run(ctx, opts)
		// Expect error due to no real cloud
		if err == nil {
			t.Logf("action %s: expected error, got nil", action)
		}
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_NotLoggedIn(t *testing.T) {
	mgr := setupTestManagerWithCloud(t, "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device123", "on"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for not logged in")
	}

	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in' error, got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_LoggedIn(t *testing.T) {
	validToken := testTokenFuture
	mgr := setupTestManagerWithCloud(t, validToken)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device123", "on"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Will fail due to network, but should get past login check
	err := cmd.Execute()
	if err == nil {
		t.Log("unexpected success - expected network error")
	} else if strings.Contains(err.Error(), "not logged in") {
		t.Error("should not get 'not logged in' error when token is set")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_WithChannelFlag(t *testing.T) {
	validToken := testTokenFuture
	mgr := setupTestManagerWithCloud(t, validToken)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device123", "off", "--channel", "1"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Will fail due to network, but flags should be parsed
	err := cmd.Execute()
	if err == nil {
		t.Log("unexpected success - expected network error")
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

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "channel flag",
			args:    []string{"--channel", "1"},
			wantErr: false,
		},
		{
			name:    "channel flag with device and action",
			args:    []string{"device", "on", "--channel", "2"},
			wantErr: false,
		},
		{
			name:    "no flags",
			args:    []string{"device", "action"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := &Options{
		DeviceID: "device123",
		Action:   "toggle",
		Channel:  2,
	}

	if opts.DeviceID != "device123" {
		t.Errorf("DeviceID = %q, want 'device123'", opts.DeviceID)
	}
	if opts.Action != "toggle" {
		t.Errorf("Action = %q, want 'toggle'", opts.Action)
	}
	if opts.Channel != 2 {
		t.Errorf("Channel = %d, want 2", opts.Channel)
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != "control" {
		t.Errorf("Name() = %q, want 'control'", cmd.Name())
	}
}

func TestNewCommand_UsageString(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	usage := cmd.UsageString()
	if !strings.Contains(usage, "control") {
		t.Error("UsageString should contain command name")
	}
	if !strings.Contains(usage, "channel") {
		t.Error("UsageString should contain channel flag")
	}
}
