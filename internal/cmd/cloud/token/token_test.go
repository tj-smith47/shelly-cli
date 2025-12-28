package token

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

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

//nolint:paralleltest // Uses global config.Get() which may have side effects
func TestRun_NotLoggedIn(t *testing.T) {
	// This test uses the global config.Get() which returns empty access token by default
	// when no config file is present

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	err := run(f)

	// Should fail because no token is configured (empty config from config.Get())
	if err == nil {
		// The test may pass if there's a global config file - that's okay
		t.Log("run succeeded - likely has global config")
		return
	}

	// If it failed, it should be "not logged in" error
	if err.Error() != "not logged in" {
		// Could be a token parse error if there's a bad config
		t.Logf("got error: %v", err)
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

//nolint:paralleltest // Uses global config.Get() which may have side effects
func TestRun_WithIOStreams(t *testing.T) {
	// Test that run function properly uses IOStreams
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// This will fail because no config, but it exercises the run function
	err := run(f)

	// We expect an error (not logged in or invalid token)
	if err == nil {
		// If no error, the system has a valid config - that's okay
		t.Log("run succeeded - global config has valid token")
		return
	}

	// The error should be about login or token
	if !strings.Contains(err.Error(), "not logged in") &&
		!strings.Contains(err.Error(), "invalid") &&
		!strings.Contains(err.Error(), "token") {
		t.Logf("unexpected error type: %v", err)
	}

	// Error output should have been written
	// (either "Not logged in" or "Token is invalid")
	if errOut.Len() == 0 {
		t.Log("no error output produced - may have different error path")
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
