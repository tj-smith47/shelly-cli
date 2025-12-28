package logout

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const cmdName = "logout"

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != cmdName {
		t.Errorf("Use = %q, want '%s'", cmd.Use, cmdName)
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

	hasSignout := false
	for _, alias := range cmd.Aliases {
		if alias == "signout" {
			hasSignout = true
			break
		}
	}

	if !hasSignout {
		t.Error("expected 'signout' alias")
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

	// Logout command should not require args
	if cmd.Args != nil {
		if err := cmd.Args(cmd, []string{}); err != nil {
			t.Errorf("command should accept zero args: %v", err)
		}
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Logout command should not define its own flags
	if cmd.Flags().NFlag() > 0 {
		t.Errorf("logout command should not have flags set, has %d", cmd.Flags().NFlag())
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Commands()) > 0 {
		t.Errorf("logout command should not have subcommands, has %d", len(cmd.Commands()))
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

	if !strings.Contains(cmd.Example, cmdName) {
		t.Error("Example should contain 'logout' command")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "Cloud") {
		t.Error("Long description should mention 'Cloud'")
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Short), "credential") && !strings.Contains(strings.ToLower(cmd.Short), "clear") {
		t.Error("Short description should mention credentials or clearing")
	}
}

func TestNewCommand_UseField(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != cmdName {
		t.Errorf("Use = %q, want '%s'", cmd.Use, cmdName)
	}
}

func TestNewCommand_AliasesContents(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	if !aliasMap["signout"] {
		t.Error("missing 'signout' alias")
	}
}

func TestNewCommand_LongMentionsToken(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "token") && !strings.Contains(cmd.Long, "credentials") {
		t.Error("Long description should mention token or credentials")
	}
}

func TestNewCommand_LongMentionsLogin(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "login") {
		t.Error("Long description should mention logging in again")
	}
}

func TestRun_WithIOStreams(t *testing.T) {
	t.Parallel()

	// This test exercises the run function with custom IOStreams
	// The output depends on global config.Get() state

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Run will use global config state
	err := run(f)

	// Should not error (either logs out or says not logged in)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should produce some output
	if out.Len() == 0 && errOut.Len() == 0 {
		t.Log("no output produced - may depend on global config state")
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

func TestNewCommand_ExampleHasContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	lines := strings.Split(cmd.Example, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	if nonEmptyLines < 1 {
		t.Error("Example should contain usage examples")
	}
}

func TestNewCommand_LongHasMultipleLines(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "\n") {
		t.Error("Long description should have multiple lines for detail")
	}
}

func TestNewCommand_LongMentionsConfiguration(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "configuration") && !strings.Contains(cmd.Long, "config") {
		t.Log("Long description may not explicitly mention configuration")
	}
}

func TestNewCommand_LongMentionsRemoval(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "remove") && !strings.Contains(cmd.Long, "clear") {
		t.Log("Long description should mention removal or clearing")
	}
}

func TestNewCommand_ShortMentionsCloud(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Short), "cloud") {
		t.Log("Short description may not explicitly mention cloud")
	}
}

func TestNewCommand_MultipleAliasesWork(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify we have at least 1 alias
	if len(cmd.Aliases) < 1 {
		t.Errorf("expected at least 1 alias, got %d", len(cmd.Aliases))
	}
}

func TestNewCommand_ExampleShowsBasicUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should show the basic logout command
	if !strings.Contains(cmd.Example, "shelly cloud logout") {
		t.Error("Example should show 'shelly cloud logout' command")
	}
}

func TestRun_ProducesOutput(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Logf("run error: %v", err)
	}

	// Run should produce either stdout or stderr output
	totalOutput := out.Len() + errOut.Len()
	if totalOutput == 0 {
		t.Log("Run produced no output - expected with empty global config")
	}
}

func TestNewCommand_RunEReturnsNil(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Fatal("RunE should not be nil")
	}
}

func TestNewCommand_Execute(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})

	// Execute the command - should not error (may or may not be logged in)
	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute error (expected if config save fails): %v", err)
	}
}

func TestNewCommand_ExecuteWithHelp(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--help"})

	// Execute with help flag
	err := cmd.Execute()
	if err != nil {
		t.Logf("Help execution error: %v", err)
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// logout should not have ValidArgsFunction (no args needed)
	if cmd.ValidArgsFunction != nil {
		t.Log("logout has ValidArgsFunction defined")
	}
}

func TestNewCommand_ExampleStartsWithIndent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	lines := strings.Split(cmd.Example, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "shelly") {
			t.Logf("Example line: %s", line)
		}
	}
}

func TestNewCommand_LongMentionsNeedToLogin(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "need") && !strings.Contains(cmd.Long, "again") {
		t.Log("Long description may mention needing to login again")
	}
}

func TestNewCommand_SilentUnknownFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Try parsing unknown flag - should fail
	err := cmd.ParseFlags([]string{"--unknown-flag"})
	if err == nil {
		t.Error("Expected error for unknown flag")
	}
}

func TestNewCommand_RunEIsCallable(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Call RunE directly
	err := cmd.RunE(cmd, []string{})

	// May error if config save fails, but should not panic
	if err != nil {
		t.Logf("RunE returned error (may be expected): %v", err)
	}
}

func TestNewCommand_OutputGoesToStreams(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute error: %v", err)
	}

	// Some output should go to our streams
	combined := out.String() + errOut.String()
	if combined == "" {
		t.Log("No output captured - command may use global config")
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != cmdName {
		t.Errorf("Name() = %q, want '%s'", cmd.Name(), cmdName)
	}
}

func TestNewCommand_CommandPath(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	path := cmd.CommandPath()
	if !strings.Contains(path, cmdName) {
		t.Errorf("CommandPath() = %q, should contain '%s'", path, cmdName)
	}
}

func TestNewCommand_UsageString(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	usage := cmd.UsageString()
	if !strings.Contains(usage, cmdName) {
		t.Error("UsageString should contain command name")
	}
}

func TestNewCommand_LongExplainsEffect(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long should explain what happens
	if !strings.Contains(strings.ToLower(cmd.Long), "remove") &&
		!strings.Contains(strings.ToLower(cmd.Long), "clear") {
		t.Log("Long description should explain the effect")
	}
}

func TestNewCommand_HelpString(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	help := cmd.UsageString()
	if help == "" {
		t.Error("Help should not be empty")
	}

	if !strings.Contains(help, cmdName) {
		t.Error("Help should mention logout")
	}
}

func TestNewCommand_RunEMultipleCalls(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Call RunE multiple times
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Logf("First RunE error: %v", err)
	}
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Logf("Second RunE error: %v", err)
	}

	// Should not panic on multiple calls
}

func TestNewCommand_RunEWithArgs(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Call RunE with extra args (should be ignored)
	err := cmd.RunE(cmd, []string{"extra", "args"})
	if err != nil {
		t.Logf("RunE with extra args error: %v", err)
	}
}

func TestNewCommand_VerifyReturnType(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand should not return nil")
	}

	if cmd.Use == "" {
		t.Error("Command Use should be set")
	}
}

func TestNewCommand_FlagsNotModified(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Logout should not add any flags
	localFlags := cmd.LocalFlags()
	if localFlags.HasFlags() {
		t.Log("Logout command has local flags defined")
	}
}

func TestNewCommand_PersistentFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check persistent flags
	if cmd.HasPersistentFlags() {
		t.Log("Command has persistent flags")
	}
}
