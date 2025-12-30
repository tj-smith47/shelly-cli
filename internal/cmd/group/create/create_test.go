package create

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// setupTestManager creates a manager with temp dir for tests that need mutations.
func setupTestManager(t *testing.T) *config.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	mgr := config.NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return mgr
}

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "create <name>" {
		t.Errorf("Use = %q, want \"create <name>\"", cmd.Use)
	}
	aliases := []string{"new", "add-group"}
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

// Note: TestRun_CreateGroup is skipped because config.CreateGroup modifies
// global config state which persists across test runs, causing "already exists"
// errors in parallel test execution.

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly group create",
		"shelly group new",
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
		"group",
		"unique",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestNewCommand_Execute_Success(t *testing.T) {
	mgr := setupTestManager(t)
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"test-group"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify group was created
	group, ok := mgr.GetGroup("test-group")
	if !ok {
		t.Fatal("group should have been created")
	}
	if group.Name != "test-group" {
		t.Errorf("group name = %q, want %q", group.Name, "test-group")
	}
	if len(group.Devices) != 0 {
		t.Errorf("group devices = %d, want 0", len(group.Devices))
	}

	// Verify output
	output := out.String()
	if !strings.Contains(output, "created") {
		t.Errorf("output should contain 'created', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestNewCommand_Execute_GroupAlreadyExists(t *testing.T) {
	mgr := setupTestManager(t)
	config.SetDefaultManager(mgr)

	// Create the group first
	if err := mgr.CreateGroup("existing-group"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"existing-group"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for existing group")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should contain 'already exists', got: %v", err)
	}
}

func TestNewCommand_Execute_InvalidName_Empty(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{}) // No args

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no args")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestNewCommand_Execute_InvalidName_Separator(t *testing.T) {
	mgr := setupTestManager(t)
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"invalid/name"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid name with separator")
	}
	if !strings.Contains(err.Error(), "path separator") || !strings.Contains(err.Error(), "colon") {
		t.Errorf("error should mention path separators or colons, got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestNewCommand_Execute_SuccessOutput(t *testing.T) {
	mgr := setupTestManager(t)
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"living-room"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()

	// Check for success message
	if !strings.Contains(output, "living-room") {
		t.Errorf("output should contain group name, got: %s", output)
	}

	// Check for hint about adding devices
	if !strings.Contains(output, "shelly group add") {
		t.Errorf("output should contain hint about adding devices, got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_Success(t *testing.T) {
	mgr := setupTestManager(t)
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	err := run(f, "kitchen")
	if err != nil {
		t.Fatalf("run() error: %v", err)
	}

	// Verify group was created
	group, ok := mgr.GetGroup("kitchen")
	if !ok {
		t.Fatal("group should have been created")
	}
	if group.Name != "kitchen" {
		t.Errorf("group name = %q, want %q", group.Name, "kitchen")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_GroupAlreadyExists(t *testing.T) {
	mgr := setupTestManager(t)
	config.SetDefaultManager(mgr)

	// Create group first
	if err := mgr.CreateGroup("bedroom"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	err := run(f, "bedroom")
	if err == nil {
		t.Fatal("expected error for existing group")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should contain 'already exists', got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_InvalidName(t *testing.T) {
	mgr := setupTestManager(t)
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	tests := []struct {
		name    string
		wantErr string
	}{
		{"path/with/slashes", "path separator"},
		{"path\\with\\backslash", "path separator"},
		{"name:colon", "colon"},
	}

	for _, tt := range tests {
		err := run(f, tt.name)
		if err == nil {
			t.Errorf("run(%q) expected error", tt.name)
			continue
		}
		if !strings.Contains(err.Error(), tt.wantErr) {
			t.Errorf("run(%q) error = %v, want to contain %q", tt.name, err, tt.wantErr)
		}
	}
}
