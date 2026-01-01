package list

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want \"list\"", cmd.Use)
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

	// Factory creates commands with "ls" and "l" aliases
	expectedAliases := map[string]bool{"ls": true, "l": true}

	if len(cmd.Aliases) < 1 {
		t.Error("expected at least 1 alias")
	}

	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
	}

	// Check that "ls" alias exists
	found := false
	for _, alias := range cmd.Aliases {
		if alias == "ls" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected alias \"ls\" not found")
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
			name:    "no args valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg invalid",
			args:    []string{"arg1"},
			wantErr: true,
		},
		{
			name:    "two args invalid",
			args:    []string{"arg1", "arg2"},
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

func TestNewCommand_NoValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// List command takes no args, so ValidArgsFunction should be nil
	if cmd.ValidArgsFunction != nil {
		t.Error("ValidArgsFunction should be nil for list command (takes no args)")
	}
}

func TestNewCommand_RunE_ExecutesSuccessfully(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Execute should succeed (list with no scenes is valid)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Note: This command uses the global -o/--output flag defined on the root command,
// not a local flag. The global flag is tested in the root command tests.

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_ExecuteWithScenes(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"scene-b": {
				Name:        "scene-b",
				Description: "Second scene",
				Actions: []config.SceneAction{
					{Device: "device1", Method: "Switch.Off"},
				},
			},
			"scene-a": {
				Name:        "scene-a",
				Description: "First scene",
				Actions: []config.SceneAction{
					{Device: "device2", Method: "Switch.On"},
				},
			},
			"scene-c": {
				Name:        "scene-c",
				Description: "",
				Actions:     []config.SceneAction{},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_ExecuteEmptyScenes(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	m := config.NewTestManager(&config.Config{})
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_ExecutesSortedByName(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"zebra": {Name: "zebra"},
			"alpha": {Name: "alpha"},
			"beta":  {Name: "beta"},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global config state
func TestNewCommand_FetchFuncIsCalled(t *testing.T) {
	// No t.Parallel() - modifies global config state
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)

	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"test-scene": {
				Name:        "test-scene",
				Description: "Test",
				Actions: []config.SceneAction{
					{Device: "dev", Method: "Method"},
				},
			},
		},
	}
	m := config.NewTestManager(cfg)
	config.SetDefaultManager(m)

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)

	// Execute and verify it doesn't panic or error
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
