package list

import (
	"bytes"
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

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}

	// Test Aliases
	wantAliases := []string{"ls", "l"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
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
		{"no args", []string{}, false},
		{"one arg", []string{"extra"}, true},
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test output flag if it exists
	flag := cmd.Flags().Lookup("output")
	if flag != nil {
		if flag.Shorthand != "o" {
			t.Errorf("--output shorthand = %q, want %q", flag.Shorthand, "o")
		}
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
		"shelly template list",
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
		"template",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestExecute_Default(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Output goes to IOStreams, check it there
	output := tf.TestIO.Out.String()
	if !strings.Contains(output, "No templates") && output == "" {
		t.Logf("output = %q", output)
	}
}

//nolint:paralleltest // Modifies global config state
func TestExecute_WithTemplates(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)
	tf.SetConfigManager(m)

	// Create some templates
	err := config.CreateDeviceTemplate(
		"template-one",
		"First template",
		"Shelly Plus 1PM",
		"",
		2,
		map[string]any{
			"switch:0": map[string]any{"name": "Switch 1"},
		},
		"device-1",
	)
	if err != nil {
		t.Fatalf("CreateDeviceTemplate: %v", err)
	}

	err = config.CreateDeviceTemplate(
		"template-two",
		"Second template",
		"Shelly Plus 2PM",
		"",
		2,
		map[string]any{
			"switch:0": map[string]any{"name": "Switch 2"},
		},
		"device-2",
	)
	if err != nil {
		t.Fatalf("CreateDeviceTemplate: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.TestIO.Out.String()
	// Should list the templates or show them in some format
	if strings.Contains(output, "No templates") {
		t.Error("expected templates to be listed, but got 'No templates'")
	}
}

//nolint:paralleltest // Modifies global config state
func TestExecute_SingleTemplate(t *testing.T) {
	// Reset config manager for isolated testing
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Create manager with properly initialized config
	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)
	tf.SetConfigManager(m)

	// Create a single template
	err := config.CreateDeviceTemplate(
		"single-template",
		"Only template",
		"Shelly Plus 1PM",
		"",
		2,
		map[string]any{
			"switch:0": map[string]any{"name": "Single Switch"},
		},
		"device-1",
	)
	if err != nil {
		t.Fatalf("CreateDeviceTemplate: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
