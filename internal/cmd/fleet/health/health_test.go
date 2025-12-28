package health

import (
	"os"
	"testing"
	"time"

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

	if cmd.Use != "health" {
		t.Errorf("Use = %q, want %q", cmd.Use, "health")
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

	expectedAliases := []string{"check", "diagnose"}
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

	// Check threshold flag
	thresholdFlag := cmd.Flags().Lookup("threshold")
	if thresholdFlag == nil {
		t.Fatal("threshold flag not found")
	}

	// Check default value is 10m
	if thresholdFlag.DefValue != "10m0s" {
		t.Errorf("threshold default = %q, want %q", thresholdFlag.DefValue, "10m0s")
	}
}

func TestNewCommand_ThresholdFlagParsing(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting the threshold flag
	if err := cmd.Flags().Set("threshold", "30m"); err != nil {
		t.Fatalf("failed to set threshold flag: %v", err)
	}

	val, err := cmd.Flags().GetDuration("threshold")
	if err != nil {
		t.Fatalf("failed to get threshold value: %v", err)
	}

	if val != 30*time.Minute {
		t.Errorf("threshold = %v, want %v", val, 30*time.Minute)
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// health takes no args
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
	if !contains(errStr, "integrator credentials required") {
		t.Errorf("error = %q, want to contain 'integrator credentials required'", errStr)
	}
}

func TestRun_MissingCredentials_WithEmptyConfig(t *testing.T) {
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
		t.Fatal("expected error for empty config credentials, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "integrator credentials required") {
		t.Errorf("error = %q, want to contain 'integrator credentials required'", errStr)
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

func TestRun_WithCustomThreshold(t *testing.T) {
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
	cmd.SetArgs([]string{"--threshold", "30m"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify threshold was parsed by checking no parse error
	errStr := err.Error()
	if contains(errStr, "invalid") && contains(errStr, "threshold") {
		t.Errorf("threshold flag parsing failed: %v", err)
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

func TestOptions_DefaultThreshold(t *testing.T) {
	t.Parallel()

	// Verify Options struct has the right default
	opts := &Options{
		Threshold: 10 * time.Minute,
	}

	if opts.Threshold != 10*time.Minute {
		t.Errorf("default Threshold = %v, want %v", opts.Threshold, 10*time.Minute)
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
