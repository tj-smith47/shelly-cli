package info

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	_ "github.com/tj-smith47/shelly-go/profiles/wave" // Register wave profiles.

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "info <model>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "info <model>")
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

	expectedAliases := []string{"show", "i"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}

	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias[%d] = %q", i, want)
			continue
		}
		if cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
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
		{name: "output", shorthand: "o", defValue: "table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
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
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"SNSW-001P16ZW"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"SNSW-001P16ZW", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction not set")
	}
}

func TestNewCommand_RunESet(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE not set")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	// Should mention Z-Wave device capabilities
	if !strings.Contains(cmd.Long, "Z-Wave") {
		t.Error("Long description should mention Z-Wave")
	}
}

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly zwave info") {
		t.Error("Example should contain 'shelly zwave info'")
	}
}

func TestNewCommand_CanBeAddedToParent(t *testing.T) {
	t.Parallel()

	parent := &cobra.Command{Use: "zwave"}
	child := NewCommand(cmdutil.NewFactory())

	parent.AddCommand(child)

	found := false
	for _, cmd := range parent.Commands() {
		if cmd.Name() == "info" {
			found = true
			break
		}
	}

	if !found {
		t.Error("info command was not added to parent")
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Factory: f,
	}

	if opts.Model != "SNSW-001P16ZW" {
		t.Errorf("Model = %q, want %q", opts.Model, "SNSW-001P16ZW")
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}
}

func TestRun_UnknownModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "UNKNOWN-MODEL-XYZ",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for unknown model")
	}

	if !strings.Contains(err.Error(), "unknown device model") {
		t.Errorf("Error should mention 'unknown device model', got: %v", err)
	}
}

func TestRun_InvalidModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		model       string
		wantErrPart string
	}{
		{
			name:        "unknown model",
			model:       "UNKNOWN-MODEL-XYZ",
			wantErrPart: "unknown device model",
		},
		{
			name:        "empty model",
			model:       "",
			wantErrPart: "unknown device model",
		},
		{
			name:        "random string",
			model:       "not-a-shelly",
			wantErrPart: "unknown device model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			opts := &Options{
				Model:   tt.model,
				Factory: tf.Factory,
			}

			err := run(opts)
			if err == nil {
				t.Error("Expected error for invalid model")
			}

			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Errorf("Error should contain %q, got: %v", tt.wantErrPart, err)
			}
		})
	}
}

func TestRun_ValidZWaveModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	// SNSW-001P16ZW is a valid Z-Wave model (Wave 1PM)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("Expected no error for valid Z-Wave model, got: %v", err)
	}

	// Verify output was produced
	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for valid Z-Wave model")
	}
}

func TestRun_OutputContainsDeviceInfo(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := tf.OutString()

	// Verify output contains expected sections
	if !strings.Contains(output, "Z-Wave") {
		t.Error("Output should contain 'Z-Wave'")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			errMsg:    "Use should not be empty",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			errMsg:    "RunE should be set",
		},
		{
			name:      "has Args",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			errMsg:    "Args should be set",
		},
		{
			name:      "has ValidArgsFunction",
			checkFunc: func(c *cobra.Command) bool { return c.ValidArgsFunction != nil },
			errMsg:    "ValidArgsFunction should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if !tt.checkFunc(cmd) {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_AliasesWork(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	for _, alias := range cmd.Aliases {
		if alias == "" {
			t.Error("Empty alias found")
		}
	}

	hasShowAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "show" {
			hasShowAlias = true
			break
		}
	}
	if !hasShowAlias {
		t.Error("Expected 'show' alias")
	}
}
