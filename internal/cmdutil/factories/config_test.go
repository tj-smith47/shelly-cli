package factories_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// =============================================================================
// ConfigDeleteCommand Tests
// =============================================================================

func TestNewConfigDeleteCommand_Structure(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:   "scene",
		ExistsFunc: func(_ string) (any, bool) { return nil, true },
		DeleteFunc: func(_ string) error { return nil },
	})

	if cmd == nil {
		t.Fatal("NewConfigDeleteCommand returned nil")
	}

	// Check command metadata
	if cmd.Use != "delete <scene>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <scene>")
	}
	if cmd.Short != "Delete a scene" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Delete a scene")
	}

	// Check aliases
	aliases := cmd.Aliases
	expectedAliases := []string{"rm", "del", "remove"}
	if len(aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", aliases, expectedAliases)
	}
	for i, a := range expectedAliases {
		if aliases[i] != a {
			t.Errorf("Aliases[%d] = %q, want %q", i, aliases[i], a)
		}
	}

	// Check --yes flag exists (confirmation commands)
	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("--yes flag should exist for commands with confirmation")
	}
}

func TestNewConfigDeleteCommand_SkipConfirmation(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:         "alias",
		SkipConfirmation: true,
		ExistsFunc:       func(_ string) (any, bool) { return nil, true },
		DeleteFunc:       func(_ string) error { return nil },
	})

	// Check --yes flag does NOT exist (skip confirmation commands)
	flag := cmd.Flags().Lookup("yes")
	if flag != nil {
		t.Error("--yes flag should NOT exist when SkipConfirmation is true")
	}
}

func TestNewConfigDeleteCommand_Execute_Success(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	var deletedName string

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(name string) (any, bool) {
			return map[string]string{"name": name}, true
		},
		DeleteFunc: func(name string) error {
			deletedName = name
			return nil
		},
	})

	cmd.SetArgs([]string{"movie-night", "--yes"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if deletedName != "movie-night" {
		t.Errorf("deletedName = %q, want %q", deletedName, "movie-night")
	}

	// Check success output
	if !strings.Contains(out.String(), "Scene") && !strings.Contains(out.String(), "deleted") {
		t.Errorf("output should contain success message, got: %s", out.String())
	}
}

func TestNewConfigDeleteCommand_Execute_SkipConfirmation(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	var deletedName string

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:         "alias",
		SkipConfirmation: true,
		ExistsFunc: func(name string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(name string) error {
			deletedName = name
			return nil
		},
	})

	// No --yes flag needed when SkipConfirmation is true
	cmd.SetArgs([]string{"lights"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if deletedName != "lights" {
		t.Errorf("deletedName = %q, want %q", deletedName, "lights")
	}
}

func TestNewConfigDeleteCommand_Execute_NotFound(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, false // Not found
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
	})

	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed for nonexistent resource")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should contain 'not found'", err.Error())
	}
}

func TestNewConfigDeleteCommand_Execute_DeleteError(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			return errors.New("disk full")
		},
	})

	cmd.SetArgs([]string{"movie-night", "--yes"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed")
	}

	if !strings.Contains(err.Error(), "failed to delete") {
		t.Errorf("error = %q, should contain 'failed to delete'", err.Error())
	}
}

func TestNewConfigDeleteCommand_Execute_Cancelled(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	deleteWasCalled := false

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			deleteWasCalled = true
			return nil
		},
	})

	// Without --yes, in non-TTY mode, confirmation returns false (default)
	cmd.SetArgs([]string{"movie-night"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if deleteWasCalled {
		t.Error("DeleteFunc should not be called when confirmation is denied")
	}

	// Check cancellation message
	if !strings.Contains(out.String(), "cancelled") {
		t.Errorf("output should contain cancellation message, got: %s", out.String())
	}
}

func TestNewConfigDeleteCommand_Execute_WithInfoFunc(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	type scene struct {
		Name    string
		Actions int
	}

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(name string) (any, bool) {
			return scene{Name: name, Actions: 5}, true
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
		InfoFunc: func(resource any, name string) string {
			if _, ok := resource.(scene); !ok {
				return "invalid type"
			}
			return "Delete scene \"" + name + "\" with 5 action(s)?"
		},
	})

	cmd.SetArgs([]string{"movie-night", "--yes"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestNewConfigDeleteCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
	})

	// No args - should fail
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed without name argument")
	}
}

func TestNewConfigDeleteCommand_WithValidArgsFunc(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ValidArgsFunc: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"scene1", "scene2"}, cobra.ShellCompDirectiveNoFileComp
		},
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
	})

	// Verify ValidArgsFunction is set
	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set when ValidArgsFunc is provided")
	}
}

// =============================================================================
// ConfigListCommand Tests
// =============================================================================

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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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

	cmd := factories.NewConfigListCommand(f, factories.ConfigListOpts[testItem]{
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
