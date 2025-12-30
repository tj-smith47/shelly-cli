package command

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "command <method> [params-json] [device...]" {
		t.Errorf("Use = %q, want \"command <method> [params-json] [device...]\"", cmd.Use)
	}
	aliases := []string{"cmd", "rpc"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
	}
	for i, alias := range aliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
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

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test requires minimum 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"Shelly.GetStatus"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"Switch.Set", "{\"id\":0}"}); err != nil {
		t.Errorf("unexpected error with 2 args: %v", err)
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flags := []struct {
		name      string
		shorthand string
	}{
		{"group", "g"},
		{"all", "a"},
		{"timeout", "t"},
		{"concurrent", "c"},
		{"output", "o"},
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

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		defValue string
	}{
		{"group", ""},
		{"all", "false"},
		{"timeout", "10s"},
		{"concurrent", "5"},
		{"output", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
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

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly batch command",
		"Shelly.GetStatus",
		"Switch.Set",
		"--group",
		"--all",
		"-o yaml",
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
		"RPC",
		"method",
		"JSON",
		"devices",
		"group",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		GroupName:  "test-group",
		All:        true,
		Timeout:    30 * time.Second,
		Concurrent: 10,
	}

	if opts.GroupName != "test-group" {
		t.Errorf("GroupName = %q, want %q", opts.GroupName, "test-group")
	}

	if !opts.All {
		t.Error("All should be true")
	}

	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", opts.Timeout, 30*time.Second)
	}

	if opts.Concurrent != 10 {
		t.Errorf("Concurrent = %d, want %d", opts.Concurrent, 10)
	}
}

func TestIsJSONObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{"{}", true},
		{`{"id":0}`, true},
		{`{"on":true,"id":0}`, true},
		{"", false},
		{"hello", false},
		{"[1,2,3]", false}, // Arrays don't count
		{"123", false},
	}

	for _, tt := range tests {
		got := utils.IsJSONObject(tt.input)
		if got != tt.want {
			t.Errorf("IsJSONObject(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
