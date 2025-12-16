package cmdutil_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

type testItem struct {
	Name  string `json:"name" yaml:"name"`
	Value int    `json:"value" yaml:"value"`
}

func TestNewConfigListCommand_Structure(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return nil
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {},
	})

	if cmd == nil {
		t.Fatal("NewConfigListCommand returned nil")
	}

	// Check command metadata
	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}
	if cmd.Short != "List widgets" {
		t.Errorf("Short = %q, want %q", cmd.Short, "List widgets")
	}

	// Check aliases
	aliases := cmd.Aliases
	expectedAliases := []string{"ls", "l"}
	if len(aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", aliases, expectedAliases)
	}
	for i, a := range expectedAliases {
		if aliases[i] != a {
			t.Errorf("Aliases[%d] = %q, want %q", i, aliases[i], a)
		}
	}

	// Check Args is NoArgs
	if cmd.Args == nil {
		t.Error("Args should be set (NoArgs)")
	}
}

func TestNewConfigListCommand_Execute_Empty(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return nil // Empty list
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {
			t.Error("DisplayFunc should not be called for empty list")
		},
	})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check default empty message
	if !strings.Contains(out.String(), "No widgets configured") {
		t.Errorf("output should contain default empty message, got: %s", out.String())
	}
}

func TestNewConfigListCommand_Execute_Empty_CustomMessage(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return nil
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {},
		EmptyMsg:    "No widgets found in config",
		HintMsg:     "Use 'shelly widget create' to add one.",
	})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "No widgets found in config") {
		t.Errorf("result should contain custom empty message, got: %s", result)
	}
	if !strings.Contains(result, "shelly widget create") {
		t.Errorf("result should contain hint message, got: %s", result)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestNewConfigListCommand_Execute_TableOutput(t *testing.T) {
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	displayCalled := false

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return []testItem{
				{Name: "foo", Value: 1},
				{Name: "bar", Value: 2},
			}
		},
		DisplayFunc: func(ios *iostreams.IOStreams, items []testItem) {
			displayCalled = true
			if len(items) != 2 {
				t.Errorf("items length = %d, want 2", len(items))
			}
			// Write something to verify display was called
			ios.Printf("Displayed widgets")
		},
	})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !displayCalled {
		t.Error("DisplayFunc should be called for table output")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestNewConfigListCommand_Execute_JSONOutput(t *testing.T) {
	// Set JSON format
	viper.Set("output", "json")
	defer viper.Set("output", "")

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	displayCalled := false

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return []testItem{
				{Name: "foo", Value: 1},
				{Name: "bar", Value: 2},
			}
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {
			displayCalled = true
		},
	})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if displayCalled {
		t.Error("DisplayFunc should NOT be called for JSON output")
	}

	// Check JSON output
	jsonOutput := out.String()
	if !strings.Contains(jsonOutput, `"name": "foo"`) {
		t.Errorf("JSON output should contain name field, got: %s", jsonOutput)
	}
	if !strings.Contains(jsonOutput, `"value": 1`) {
		t.Errorf("JSON output should contain value field, got: %s", jsonOutput)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestNewConfigListCommand_Execute_YAMLOutput(t *testing.T) {
	// Set YAML format
	viper.Set("output", "yaml")
	defer viper.Set("output", "")

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return []testItem{
				{Name: "foo", Value: 1},
			}
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {
			t.Error("DisplayFunc should NOT be called for YAML output")
		},
	})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check YAML output
	yamlOutput := out.String()
	if !strings.Contains(yamlOutput, "name: foo") {
		t.Errorf("YAML output should contain name field, got: %s", yamlOutput)
	}
	if !strings.Contains(yamlOutput, "value: 1") {
		t.Errorf("YAML output should contain value field, got: %s", yamlOutput)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestNewConfigListCommand_Execute_TemplateOutput(t *testing.T) {
	// Set template format
	viper.Set("output", "template")
	viper.Set("template", "{{range .}}{{.Name}}: {{.Value}}\n{{end}}")
	defer func() {
		viper.Set("output", "")
		viper.Set("template", "")
	}()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return []testItem{
				{Name: "foo", Value: 1},
				{Name: "bar", Value: 2},
			}
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {
			t.Error("DisplayFunc should NOT be called for template output")
		},
	})

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check template output
	templateOutput := out.String()
	if !strings.Contains(templateOutput, "foo: 1") {
		t.Errorf("template output should contain 'foo: 1', got: %s", templateOutput)
	}
	if !strings.Contains(templateOutput, "bar: 2") {
		t.Errorf("template output should contain 'bar: 2', got: %s", templateOutput)
	}
}

func TestNewConfigListCommand_RejectsArgs(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return nil
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {},
	})

	cmd.SetArgs([]string{"unexpected-arg"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should fail with unexpected arguments")
	}
}

func TestNewConfigListCommand_Examples(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := cmdutil.NewConfigListCommand(f, cmdutil.ConfigListOpts[testItem]{
		Resource: "widget",
		FetchFunc: func() []testItem {
			return nil
		},
		DisplayFunc: func(_ *iostreams.IOStreams, _ []testItem) {},
	})

	example := cmd.Example
	if !strings.Contains(example, "shelly widget list") {
		t.Errorf("Example should contain 'shelly widget list', got: %s", example)
	}
	if !strings.Contains(example, "-o json") {
		t.Errorf("Example should contain '-o json', got: %s", example)
	}
	if !strings.Contains(example, "-o yaml") {
		t.Errorf("Example should contain '-o yaml', got: %s", example)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestWantsStructured_Integration(t *testing.T) {
	// Test that WantsStructured works correctly for JSON, YAML, and template
	tests := []struct {
		format   string
		expected bool
	}{
		{"json", true},
		{"yaml", true},
		{"template", true},
		{"table", false},
		{"text", false},
		{"", false},
	}

	for _, tt := range tests {
		// Can't use t.Run with t.Parallel here due to viper concurrent map writes
		viper.Set("output", tt.format)

		got := output.WantsStructured()
		if got != tt.expected {
			t.Errorf("WantsStructured() for %q = %v, want %v", tt.format, got, tt.expected)
		}

		viper.Set("output", "")
	}
}
