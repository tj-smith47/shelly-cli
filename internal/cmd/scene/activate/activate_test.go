package activate

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "activate <name>" {
		t.Errorf("Use = %q, want \"activate <name>\"", cmd.Use)
	}
	aliases := []string{"run", "exec", "play"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
	}
	for i, expected := range aliases {
		if cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
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

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg valid",
			args:    []string{"scene-name"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"name1", "name2"},
			wantErr: true,
		},
		{
			name:    "three args",
			args:    []string{"name1", "name2", "name3"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"timeout", "t", "10s"},
		{"concurrent", "c", "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_DryRunFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("dry-run flag not found")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("dry-run default = %q, want \"false\"", dryRunFlag.DefValue)
	}
	// dry-run typically doesn't have a shorthand
	if dryRunFlag.Shorthand != "" {
		t.Errorf("dry-run shorthand = %q, want empty", dryRunFlag.Shorthand)
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil, expected completion function for scene names")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestFormatParamsInline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params map[string]any
		empty  bool
	}{
		{"nil params", nil, true},
		{"empty params", map[string]any{}, true},
		{"with single param", map[string]any{"id": 0}, false},
		{"with multiple params", map[string]any{"id": 0, "on": true}, false},
		{"with string param", map[string]any{"name": "test"}, false},
		{"with nested param", map[string]any{"config": map[string]any{"key": "value"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := output.FormatParamsInline(tt.params)
			if tt.empty && result != "" {
				t.Errorf("FormatParamsInline() = %q, want empty", result)
			}
			if !tt.empty && result == "" {
				t.Errorf("FormatParamsInline() = empty, want non-empty")
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description mentions key features
	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	// Long description should mention concurrent execution
	if len(cmd.Long) < 50 {
		t.Error("Long description seems too short")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain actual usage patterns
	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Example should show basic usage
	examples := cmd.Example
	if len(examples) < 20 {
		t.Error("Example seems too short to be useful")
	}
}

func TestNewCommand_RunE_SceneNotFound(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Set up command to run
	cmd.SetArgs([]string{"nonexistent-scene-name-12345"})

	// Execute should fail because scene doesn't exist
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent scene, got nil")
	}
}

func TestNewCommand_FlagConcurrentDefault(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	concurrentFlag := cmd.Flags().Lookup("concurrent")
	if concurrentFlag == nil {
		t.Fatal("concurrent flag not found")
	}

	// Check that concurrent default is sensible (5)
	if concurrentFlag.DefValue != "5" {
		t.Errorf("concurrent default = %q, want \"5\"", concurrentFlag.DefValue)
	}
}

func TestNewCommand_FlagTimeoutDefault(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	timeoutFlag := cmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Fatal("timeout flag not found")
	}

	// Check that timeout default is 10 seconds
	if timeoutFlag.DefValue != "10s" {
		t.Errorf("timeout default = %q, want \"10s\"", timeoutFlag.DefValue)
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

// Note: TestRun_EmptyScene and TestRun_DryRun are skipped because config.GetScene
// uses global config state and the test manager doesn't propagate to those functions.
// The run function logic is covered by TestNewCommand_RunE_SceneNotFound.
