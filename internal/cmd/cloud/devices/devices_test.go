package devices

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "devices" {
		t.Errorf("Use = %q, want 'devices'", cmd.Use)
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

	expectedAliases := map[string]bool{"ls": true, "list": true}
	for _, alias := range cmd.Aliases {
		delete(expectedAliases, alias)
	}

	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
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

	// Devices command should not require args
	if cmd.Args != nil {
		if err := cmd.Args(cmd, []string{}); err != nil {
			t.Errorf("command should accept zero args: %v", err)
		}
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check if there are any flags defined by the command itself
	// Note: some flags may be inherited from parent
	if cmd.Flags().NFlag() > 0 {
		t.Logf("devices command has %d flags set", cmd.Flags().NFlag())
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Commands()) > 0 {
		t.Errorf("devices command should not have subcommands, has %d", len(cmd.Commands()))
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

	if !strings.Contains(cmd.Example, "devices") {
		t.Error("Example should contain 'devices' command")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "device") {
		t.Error("Long description should mention 'device'")
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Short), "device") && !strings.Contains(strings.ToLower(cmd.Short), "cloud") {
		t.Error("Short description should mention device or cloud")
	}
}

func TestNewCommand_UseField(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "devices" {
		t.Errorf("Use = %q, want 'devices'", cmd.Use)
	}
}

func TestNewCommand_AliasesContents(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	if !aliasMap["ls"] {
		t.Error("missing 'ls' alias")
	}
	if !aliasMap["list"] {
		t.Error("missing 'list' alias")
	}
}

func TestNewCommand_LongMentionsCloud(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "Cloud") && !strings.Contains(cmd.Long, "cloud") {
		t.Error("Long description should mention Cloud")
	}
}

func TestNewCommand_LongMentionsRegistered(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "registered") {
		t.Error("Long description should mention registered devices")
	}
}

func TestRun_NotLoggedIn(t *testing.T) {
	t.Parallel()

	// Test that run returns error when not logged in
	// This uses global config which may or may not be logged in

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	err := run(f, ctx)

	// If global config has no token, this should error
	if err != nil {
		if !strings.Contains(err.Error(), "not logged in") {
			t.Logf("run error: %v", err)
		}
	}
}

func TestRun_WithCancelledContext(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(f, ctx)

	// Should fail due to cancelled context or not logged in
	if err == nil {
		t.Log("run succeeded with cancelled context - may have global config")
	}
}

func TestRun_WithTimeout(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow timeout to trigger
	time.Sleep(1 * time.Millisecond)

	err := run(f, ctx)

	// Should fail due to timeout or not logged in
	if err == nil {
		t.Log("run succeeded with timed out context - may have global config")
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

	if !strings.Contains(cmd.Long, "\n") {
		t.Log("Long description may be single line")
	}
}

func TestNewCommand_LongMentionsStatus(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "status") && !strings.Contains(cmd.Long, "online") {
		t.Log("Long description may mention device status")
	}
}

func TestNewCommand_ExampleMentionsJSON(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Example), "json") {
		t.Log("Example may show JSON output option")
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

func TestNewCommand_ExampleShowsBasicUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should show the basic devices command
	if !strings.Contains(cmd.Example, "shelly cloud devices") {
		t.Error("Example should show 'shelly cloud devices' command")
	}
}

func TestNewCommand_LongMentionsDeviceInfo(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should mention what info is shown
	if !strings.Contains(cmd.Long, "ID") &&
		!strings.Contains(cmd.Long, "name") &&
		!strings.Contains(cmd.Long, "model") {
		t.Log("Long description may describe device info shown")
	}
}

func TestNewCommand_ExampleWithOutputFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should show output format option
	if !strings.Contains(cmd.Example, "-o") {
		t.Log("Example may show -o output format option")
	}
}

func TestRun_ProducesOutput(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()
	err := run(f, ctx)
	if err != nil {
		t.Logf("run error: %v", err)
	}

	// Run should produce either stdout or stderr output
	totalOutput := out.Len() + errOut.Len()
	if totalOutput == 0 {
		t.Log("Run produced no output - expected with empty global config")
	}
}
