package disconnect

import (
	"os"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "disconnect" {
		t.Errorf("Use = %q, want %q", cmd.Use, "disconnect")
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

	expectedAliases := []string{"logout", "close"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}

	for i, alias := range expectedAliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
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
		{
			name:    "no args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg",
			args:    []string{"arg1"},
			wantErr: true,
		},
		{
			name:    "multiple args",
			args:    []string{"arg1", "arg2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_RunEExists(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is not set")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "WebSocket") {
		t.Error("Long should contain 'WebSocket'")
	}

	if !strings.Contains(cmd.Long, "SHELLY_INTEGRATOR_TAG") {
		t.Error("Long should contain 'SHELLY_INTEGRATOR_TAG'")
	}

	if !strings.Contains(cmd.Long, "SHELLY_INTEGRATOR_TOKEN") {
		t.Error("Long should contain 'SHELLY_INTEGRATOR_TOKEN'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "shelly fleet disconnect") {
		t.Error("Example should contain 'shelly fleet disconnect'")
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
	// The disconnect command should return a credentials error
	if !strings.Contains(errStr, "credentials") && !strings.Contains(errStr, "integrator") {
		t.Errorf("error = %q, want to contain 'credentials' or 'integrator'", errStr)
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
	if !strings.Contains(errStr, "authentication") && !strings.Contains(errStr, "failed") {
		t.Errorf("error = %q, want to contain 'authentication' or 'failed'", errStr)
	}
}

func TestNewCommand_NoLocalFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The disconnect command doesn't add its own local flags
	if cmd.HasLocalFlags() {
		t.Error("disconnect command should not have local flags")
	}
}

func TestNewCommand_WithFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with factory")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}
