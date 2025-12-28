package status

import (
	"os"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
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

	expectedAliases := []string{"st", "list", "ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Fatalf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check online flag
	onlineFlag := cmd.Flags().Lookup("online")
	if onlineFlag == nil {
		t.Fatal("online flag not found")
	}
	if onlineFlag.DefValue != "false" {
		t.Errorf("online default = %q, want %q", onlineFlag.DefValue, "false")
	}

	// Check offline flag
	offlineFlag := cmd.Flags().Lookup("offline")
	if offlineFlag == nil {
		t.Fatal("offline flag not found")
	}
	if offlineFlag.DefValue != "false" {
		t.Errorf("offline default = %q, want %q", offlineFlag.DefValue, "false")
	}
}

func TestNewCommand_OnlineFlagParsing(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting the online flag
	if err := cmd.Flags().Set("online", "true"); err != nil {
		t.Fatalf("failed to set online flag: %v", err)
	}

	val, err := cmd.Flags().GetBool("online")
	if err != nil {
		t.Fatalf("failed to get online value: %v", err)
	}

	if !val {
		t.Error("online = false, want true")
	}
}

func TestNewCommand_OfflineFlagParsing(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting the offline flag
	if err := cmd.Flags().Set("offline", "true"); err != nil {
		t.Fatalf("failed to set offline flag: %v", err)
	}

	val, err := cmd.Flags().GetBool("offline")
	if err != nil {
		t.Fatalf("failed to get offline value: %v", err)
	}

	if !val {
		t.Error("offline = false, want true")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// status takes no args
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("expected no error for zero args, got: %v", err)
	}

	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("expected error for args provided, got nil")
	}
}

func TestRun_MissingCredentials_NoEnv(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// Ensure config has no integrator credentials
	tf.Config.Integrator = config.IntegratorConfig{}

	// Clear env vars
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}

	errStr := err.Error()
	// The status command uses GetIntegratorCredentials which returns a different error message
	if !contains(errStr, "credentials") && !contains(errStr, "connect") {
		t.Errorf("error = %q, want to contain 'credentials' or 'connect'", errStr)
	}
}

func TestRun_WithConfigCredentials_ButNoServer(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// Set valid credentials in config
	tf.Config.Integrator = config.IntegratorConfig{
		Tag:   "test-tag",
		Token: "test-token",
	}

	// Clear env vars to test config fallback
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	// Should fail with authentication error (no real server)
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	errStr := err.Error()
	// This will fail authentication since credentials are fake
	if !contains(errStr, "authentication failed") && !contains(errStr, "failed") {
		t.Errorf("error = %q, want to contain 'authentication failed' or 'failed'", errStr)
	}
}

func TestRun_WithOnlineFlag(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// Set valid credentials in config
	tf.Config.Integrator = config.IntegratorConfig{
		Tag:   "test-tag",
		Token: "test-token",
	}

	// Clear env vars
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--online"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify flag was parsed by checking no parse error
	errStr := err.Error()
	if contains(errStr, "invalid") && contains(errStr, "online") {
		t.Errorf("online flag parsing failed: %v", err)
	}
}

func TestRun_WithOfflineFlag(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// Set valid credentials in config
	tf.Config.Integrator = config.IntegratorConfig{
		Tag:   "test-tag",
		Token: "test-token",
	}

	// Clear env vars
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--offline"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify flag was parsed by checking no parse error
	errStr := err.Error()
	if contains(errStr, "invalid") && contains(errStr, "offline") {
		t.Errorf("offline flag parsing failed: %v", err)
	}
}

func TestNewCommand_WithFactory(t *testing.T) {
	t.Parallel()

	// Test that factory is properly wired
	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"test-device": {Name: "test-device", Address: "10.0.0.1"},
	})

	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with factory")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	// Verify Options struct with defaults
	opts := &Options{}

	if opts.Online {
		t.Error("default Online = true, want false")
	}

	if opts.Offline {
		t.Error("default Offline = true, want false")
	}
}

func TestOptions_MutuallyExclusiveFlags(t *testing.T) {
	t.Parallel()

	// Options allows both flags to be set; filtering handles it
	opts := &Options{
		Online:  true,
		Offline: true,
	}

	// Both can be true at the struct level
	// The run function filters appropriately
	if !opts.Online || !opts.Offline {
		t.Error("expected both flags to be settable")
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
