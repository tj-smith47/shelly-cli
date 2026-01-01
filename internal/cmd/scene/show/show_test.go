package show

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const formatTable = "table"

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "show <name>" {
		t.Errorf("Use = %q, want \"show <name>\"", cmd.Use)
	}
	aliases := []string{"info", "get"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
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

	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"name"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"name", "extra"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
	if outputFlag.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want o", outputFlag.Shorthand)
	}
	if outputFlag.DefValue != formatTable {
		t.Errorf("output default = %q, want table", outputFlag.DefValue)
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
		{"with params", map[string]any{"id": 0}, false},
	}

	for _, tt := range tests {
		result := output.FormatParamsInline(tt.params)
		if tt.empty && result != "" {
			t.Errorf("%s: FormatParamsInline() = %q, want empty", tt.name, result)
		}
		if !tt.empty && result == "" {
			t.Errorf("%s: FormatParamsInline() = empty, want non-empty", tt.name)
		}
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for scene name completion")
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_SceneNotFound(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "nonexistent-scene",
	}

	err := run(opts)

	if err == nil {
		t.Error("expected error for nonexistent scene")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want to contain 'not found'", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_TableOutput(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"test-scene": {
				Name:        "test-scene",
				Description: "Test scene description",
				Actions: []config.SceneAction{
					{Device: "device1", Method: "Switch.Set", Params: map[string]any{"id": 0, "on": true}},
					{Device: "device2", Method: "Switch.Toggle", Params: nil},
				},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "test-scene",
	}
	opts.Format = formatTable

	err := run(opts)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output contains expected data
	allOutput := tf.OutString() + tf.ErrString()
	if !strings.Contains(allOutput, "test-scene") {
		t.Errorf("output = %q, want to contain 'test-scene'", allOutput)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_JSONOutput(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"json-scene": {
				Name:        "json-scene",
				Description: "JSON test scene",
				Actions: []config.SceneAction{
					{Device: "light", Method: "Light.Set", Params: map[string]any{"brightness": 50}},
				},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	// Capture stdout for JSON output
	var stdout bytes.Buffer
	tf := factory.NewTestFactory(t)

	// Manually set output
	opts := &Options{
		Factory: tf.Factory,
		Name:    "json-scene",
	}
	opts.Format = "json"

	err := run(opts)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// JSON output goes to os.Stdout, not the captured buffer
	// Just verify no error occurred
	_ = stdout
}

//nolint:paralleltest // Test modifies global config state
func TestRun_YAMLOutput(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"yaml-scene": {
				Name:        "yaml-scene",
				Description: "YAML test scene",
				Actions: []config.SceneAction{
					{Device: "switch", Method: "Switch.Toggle"},
				},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "yaml-scene",
	}
	opts.Format = "yaml"

	err := run(opts)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_Execute(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"exec-scene": {
				Name:        "exec-scene",
				Description: "Execute test",
				Actions:     []config.SceneAction{},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"exec-scene"})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_ExecuteWithOutputFlag(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"flag-scene": {
				Name: "flag-scene",
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"flag-scene", "-o", "json"})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_SceneWithNoActions(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"empty-scene": {
				Name:        "empty-scene",
				Description: "An empty scene",
				Actions:     []config.SceneAction{},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "empty-scene",
	}
	opts.Format = formatTable

	err := run(opts)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun_SceneWithNoDescription(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"no-desc-scene": {
				Name:        "no-desc-scene",
				Description: "",
				Actions: []config.SceneAction{
					{Device: "dev1", Method: "Method1"},
				},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "no-desc-scene",
	}
	opts.Format = formatTable

	err := run(opts)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
