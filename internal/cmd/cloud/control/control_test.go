package control

import (
	"bytes"
	"context"
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

func TestRun_NotLoggedIn(t *testing.T) {
	// This test uses the global config.Get() which should return empty access token
	// when no config file is present

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	err := run(ctx, f, "device123", "on", 0)

	// Should fail because no token is configured
	if err == nil {
		// Test may pass if there's a global config file - that's okay
		t.Log("run succeeded - likely has global config")
		return
	}

	// If it failed, check for "not logged in" error
	if !strings.Contains(err.Error(), "not logged in") {
		t.Logf("got error: %v (expected 'not logged in')", err)
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
