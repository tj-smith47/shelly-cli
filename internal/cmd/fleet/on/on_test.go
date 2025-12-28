package on

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

	if cmd.Use != "on [device...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "on [device...]")
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

	expectedAliases := []string{"turn-on", "enable"}
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

	// Check batch flags from BatchFlags struct
	flags := []struct {
		name      string
		shorthand string
	}{
		{"group", "g"},
		{"all", "a"},
		{"timeout", "t"},
		{"switch", "s"},
		{"concurrent", "c"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag %q not found", f.name)
			continue
		}
		if flag.Shorthand != f.shorthand {
			t.Errorf("flag %q shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

func TestNewCommand_GroupFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.Flags().Set("group", "living-room"); err != nil {
		t.Fatalf("failed to set group flag: %v", err)
	}

	val, err := cmd.Flags().GetString("group")
	if err != nil {
		t.Fatalf("failed to get group value: %v", err)
	}

	if val != "living-room" {
		t.Errorf("group = %q, want %q", val, "living-room")
	}
}

func TestNewCommand_AllFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.Flags().Set("all", "true"); err != nil {
		t.Fatalf("failed to set all flag: %v", err)
	}

	val, err := cmd.Flags().GetBool("all")
	if err != nil {
		t.Fatalf("failed to get all value: %v", err)
	}

	if !val {
		t.Error("all = false, want true")
	}
}

func TestNewCommand_AcceptsArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// on command doesn't have explicit Args validator (nil means cobra.ArbitraryArgs by default)
	// which allows any number of args. Verify the command has no explicit Args restriction.
	if cmd.Args != nil {
		// If Args is set, verify it accepts arbitrary args
		if err := cmd.Args(cmd, []string{}); err != nil {
			t.Errorf("expected no error for zero args, got: %v", err)
		}
		if err := cmd.Args(cmd, []string{"device1"}); err != nil {
			t.Errorf("expected no error for one arg, got: %v", err)
		}
		if err := cmd.Args(cmd, []string{"device1", "device2"}); err != nil {
			t.Errorf("expected no error for multiple args, got: %v", err)
		}
	}
	// nil Args is acceptable - cobra defaults to allowing any args
}

func TestRun_NoDevicesSpecified(t *testing.T) {
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
	cmd.SetArgs([]string{}) // No devices, no --all, no --group
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for no devices specified, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "no devices specified") {
		t.Errorf("error = %q, want to contain 'no devices specified'", errStr)
	}
}

func TestRun_MissingCredentials(t *testing.T) {
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
	cmd.SetArgs([]string{"device1"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "integrator credentials required") && !contains(errStr, "credentials") {
		t.Errorf("error = %q, want to contain 'integrator credentials required' or 'credentials'", errStr)
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
	cmd.SetArgs([]string{"device1"})
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

func TestRun_WithAllFlag(t *testing.T) {
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
	cmd.SetArgs([]string{"--all"})
	err := cmd.Execute()

	// Will fail auth, but --all flag should bypass "no devices" error
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	errStr := err.Error()
	if contains(errStr, "no devices specified") {
		t.Errorf("--all flag should bypass 'no devices' error, got: %v", err)
	}
}

func TestRun_WithGroupFlag(t *testing.T) {
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
	cmd.SetArgs([]string{"--group", "living-room"})
	err := cmd.Execute()

	// Will fail auth, but --group flag should bypass "no devices" error
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	errStr := err.Error()
	if contains(errStr, "no devices specified") {
		t.Errorf("--group flag should bypass 'no devices' error, got: %v", err)
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

func TestOptions_EmbedsBatchFlags(t *testing.T) {
	t.Parallel()

	// Verify Options embeds BatchFlags
	opts := &Options{}

	// Should be able to access BatchFlags fields
	opts.All = true
	opts.GroupName = "test-group"

	if !opts.All {
		t.Error("opts.All = false, want true")
	}
	if opts.GroupName != "test-group" {
		t.Errorf("opts.GroupName = %q, want %q", opts.GroupName, "test-group")
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
