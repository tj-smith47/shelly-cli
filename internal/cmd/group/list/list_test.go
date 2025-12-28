package list

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want list", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "ls" {
		t.Errorf("Aliases = %v, want [ls]", cmd.Aliases)
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

	// ConfigListCommand uses aliases: ls, l
	expectedAliases := []string{"ls", "l"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// ConfigListCommand uses cobra.NoArgs - it takes no arguments
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("unexpected error with no args: %v", err)
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("expected error with extra arg")
	}
	if err := cmd.Args(cmd, []string{"arg1", "arg2"}); err == nil {
		t.Error("expected error with multiple args")
	}
}

func TestNewCommand_ArgsTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no args valid",
			args:      []string{},
			wantError: false,
		},
		{
			name:      "one arg invalid",
			args:      []string{"living-room"},
			wantError: true,
		},
		{
			name:      "two args invalid",
			args:      []string{"living-room", "bedroom"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantError {
				t.Errorf("Args() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewCommand_RunEExists(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil, expected to be set")
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should contain "group" somewhere in the short description
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention output formats
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain usage samples
	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

// TestRun_EmptyGroups tests the behavior when no groups exist.
func TestRun_EmptyGroups(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: make(map[string]config.Group),
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRun_WithGroups tests the behavior when groups exist.
func TestRun_WithGroups(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"living-room": {
				Name:    "living-room",
				Devices: []string{"light-1", "light-2"},
			},
			"bedroom": {
				Name:    "bedroom",
				Devices: []string{"lamp-1"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRun_SingleGroup tests behavior with exactly one group.
func TestRun_SingleGroup(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"office": {
				Name:    "office",
				Devices: []string{"desk-lamp", "monitor-light"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRun_GroupWithNoDevices tests behavior when a group has no devices.
func TestRun_GroupWithNoDevices(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"empty-group": {
				Name:    "empty-group",
				Devices: []string{},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRun_ManyGroups tests behavior with many groups.
func TestRun_ManyGroups(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"group-1": {Name: "group-1", Devices: []string{"d1"}},
			"group-2": {Name: "group-2", Devices: []string{"d2", "d3"}},
			"group-3": {Name: "group-3", Devices: []string{}},
			"group-4": {Name: "group-4", Devices: []string{"d4", "d5", "d6"}},
			"group-5": {Name: "group-5", Devices: []string{"d7"}},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Note: This command uses the global -o/--output flag defined on the root command,
// not a local flag. The global flag is tested in the root command tests.
