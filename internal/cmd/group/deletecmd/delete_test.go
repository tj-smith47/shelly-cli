package deletecmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// setupTestManagerWithGroups creates a test manager and populates it with groups.
func setupTestManagerWithGroups(t *testing.T, groups map[string]config.Group) *config.Manager {
	t.Helper()
	cfg := &config.Config{
		Groups: groups,
	}
	return config.NewTestManager(cfg)
}

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "delete <group>" {
		t.Errorf("Use = %q, want \"delete <group>\"", cmd.Use)
	}
	aliases := []string{"rm", "del", "remove"}
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

	// Test requires exactly 1 argument
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

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want y", yesFlag.Shorthand)
	}
	if yesFlag.DefValue != "false" {
		t.Errorf("yes default = %q, want false", yesFlag.DefValue)
	}
}

func TestExecute_Help(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "delete") {
		t.Error("help output should contain 'delete'")
	}
	if !strings.Contains(output, "group") {
		t.Error("help output should contain 'group'")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_GroupNotFound(t *testing.T) {
	// Setup manager with no groups
	mgr := setupTestManagerWithGroups(t, map[string]config.Group{})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"nonexistent-group", "--yes"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent group")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_Success_EmptyGroup(t *testing.T) {
	// Setup manager with an empty group
	mgr := setupTestManagerWithGroups(t, map[string]config.Group{
		"empty-group": {Name: "empty-group", Devices: []string{}},
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"empty-group", "--yes"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}

	// Verify group was actually deleted
	if _, exists := mgr.GetGroup("empty-group"); exists {
		t.Error("group should have been deleted")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_Success_GroupWithDevices(t *testing.T) {
	// Setup manager with a group containing devices
	mgr := setupTestManagerWithGroups(t, map[string]config.Group{
		"downstairs": {Name: "downstairs", Devices: []string{"living-room", "kitchen"}},
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"downstairs", "--yes"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}

	// Verify group was actually deleted
	if _, exists := mgr.GetGroup("downstairs"); exists {
		t.Error("group should have been deleted")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_CancelConfirmation(t *testing.T) {
	// Setup manager with a group
	mgr := setupTestManagerWithGroups(t, map[string]config.Group{
		"my-group": {Name: "my-group", Devices: []string{"test-device"}},
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	// Without --yes and in non-TTY mode, confirmation should return false
	cmd := NewCommand(f)
	cmd.SetArgs([]string{"my-group"}) // Note: no --yes flag
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "cancelled") {
		t.Errorf("expected 'cancelled' in output, got: %s", output)
	}

	// Verify group still exists (not deleted because confirmation was cancelled)
	if _, exists := mgr.GetGroup("my-group"); !exists {
		t.Error("group should still exist after cancelled deletion")
	}
}

func TestExecute_ValidArgsFunction(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for group completion")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_MultipleGroups_DeleteOne(t *testing.T) {
	// Setup manager with multiple groups
	mgr := setupTestManagerWithGroups(t, map[string]config.Group{
		"group-one": {Name: "group-one", Devices: []string{"device-a"}},
		"group-two": {Name: "group-two", Devices: []string{"device-b"}},
	})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"group-one", "--yes"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute returned error: %v", err)
	}

	// Verify only group-one was deleted
	if _, exists := mgr.GetGroup("group-one"); exists {
		t.Error("group-one should have been deleted")
	}
	if _, exists := mgr.GetGroup("group-two"); !exists {
		t.Error("group-two should still exist")
	}
}
