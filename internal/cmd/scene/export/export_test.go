package export

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "export <name> [file]" {
		t.Errorf("Use = %q, want \"export <name> [file]\"", cmd.Use)
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

	expectedAliases := []string{"save", "backup"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}
	for i, expected := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias %q", expected)
			continue
		}
		if cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
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
			name:    "two args valid",
			args:    []string{"scene-name", "output.yaml"},
			wantErr: false,
		},
		{
			name:    "three args invalid",
			args:    []string{"name", "file", "extra"},
			wantErr: true,
		},
		{
			name:    "four args invalid",
			args:    []string{"a", "b", "c", "d"},
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

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("format shorthand = %q, want \"f\"", formatFlag.Shorthand)
	}
	if formatFlag.DefValue != "yaml" {
		t.Errorf("format default = %q, want \"yaml\"", formatFlag.DefValue)
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

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description exists and is meaningful
	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	if len(cmd.Long) < 30 {
		t.Error("Long description seems too short")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Example should show meaningful patterns
	if len(cmd.Example) < 20 {
		t.Error("Example seems too short to be useful")
	}
}

func TestNewCommand_RunE_SceneNotFound(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Set up command to run with nonexistent scene
	cmd.SetArgs([]string{"nonexistent-scene-name-12345"})

	// Execute should fail because scene doesn't exist
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent scene, got nil")
	}
}

func TestNewCommand_FormatFlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}

	// Verify the flag has a meaningful usage description
	if formatFlag.Usage == "" {
		t.Error("format flag has no usage description")
	}
}
