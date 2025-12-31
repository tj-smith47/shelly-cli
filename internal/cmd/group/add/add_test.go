package add

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// setupTestManager creates a test manager with pre-initialized config for mutation tests.
func setupTestManager(t *testing.T) *config.Manager {
	t.Helper()
	return config.NewTestManager(&config.Config{})
}

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "add <group> <device>..." {
		t.Errorf("Use = %q, want \"add <group> <device>...\"", cmd.Use)
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

	expectedAliases := []string{"append", "include"}
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

	// Test requires minimum 2 arguments
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"group"}); err == nil {
		t.Error("expected error with only 1 arg")
	}
	if err := cmd.Args(cmd, []string{"group", "device"}); err != nil {
		t.Errorf("unexpected error with 2 args: %v", err)
	}
	if err := cmd.Args(cmd, []string{"group", "device1", "device2"}); err != nil {
		t.Errorf("unexpected error with 3 args: %v", err)
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
			name:      "no args",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "one arg only group",
			args:      []string{"living-room"},
			wantError: true,
		},
		{
			name:      "two args group and device",
			args:      []string{"living-room", "light-1"},
			wantError: false,
		},
		{
			name:      "three args group and two devices",
			args:      []string{"living-room", "light-1", "light-2"},
			wantError: false,
		},
		{
			name:      "many devices",
			args:      []string{"living-room", "d1", "d2", "d3", "d4", "d5"},
			wantError: false,
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

	expected := "Add devices to a group"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
	if len(cmd.Long) < 50 {
		t.Error("Long description seems too short")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
	if len(cmd.Example) < 50 {
		t.Error("Example seems too short")
	}
}

func TestRun_GroupNotFound(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: make(map[string]config.Group),
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	err := run(f, "nonexistent-group", []string{"device1"})
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
	if err.Error() != `group "nonexistent-group" not found` {
		t.Errorf("error = %q, want group not found error", err.Error())
	}
}

func TestRun_GroupNotFoundTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		groupName string
		devices   []string
		wantErr   string
	}{
		{
			name:      "nonexistent group",
			groupName: "nonexistent-group",
			devices:   []string{"device1"},
			wantErr:   `group "nonexistent-group" not found`,
		},
		{
			name:      "empty group name lookup",
			groupName: "missing",
			devices:   []string{"device1", "device2"},
			wantErr:   `group "missing" not found`,
		},
		{
			name:      "special characters in group name",
			groupName: "my-special-group",
			devices:   []string{"dev"},
			wantErr:   `group "my-special-group" not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				Groups: make(map[string]config.Group),
			}
			mgr := config.NewTestManager(cfg)

			out := &bytes.Buffer{}
			errOut := &bytes.Buffer{}
			ios := iostreams.Test(nil, out, errOut)

			f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

			err := run(f, tt.groupName, tt.devices)
			if err == nil {
				t.Fatal("expected error for nonexistent group")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRun_AddSingleDevice(t *testing.T) {
	t.Parallel()

	mgr := setupTestManager(t)
	if err := mgr.CreateGroup("living-room"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	err := run(f, "living-room", []string{"light-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group, ok := mgr.GetGroup("living-room")
	if !ok {
		t.Fatal("group not found after add")
	}
	if len(group.Devices) != 1 {
		t.Errorf("group devices = %d, want 1", len(group.Devices))
	}
	if group.Devices[0] != "light-1" {
		t.Errorf("group devices[0] = %q, want light-1", group.Devices[0])
	}
}

func TestRun_AddMultipleDevices(t *testing.T) {
	t.Parallel()

	mgr := setupTestManager(t)
	if err := mgr.CreateGroup("office"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	err := run(f, "office", []string{"lamp-1", "lamp-2", "switch-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group, ok := mgr.GetGroup("office")
	if !ok {
		t.Fatal("group not found after add")
	}
	if len(group.Devices) != 3 {
		t.Errorf("group devices = %d, want 3", len(group.Devices))
	}
}

func TestRun_AddDeviceAlreadyInGroup(t *testing.T) {
	t.Parallel()

	mgr := setupTestManager(t)
	if err := mgr.CreateGroup("bedroom"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}
	if err := mgr.AddDeviceToGroup("bedroom", "lamp-1"); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	err := run(f, "bedroom", []string{"lamp-1"})
	if err == nil {
		t.Fatal("expected error when no devices were added")
	}
	if err.Error() != "no devices were added" {
		t.Errorf("error = %q, want 'no devices were added'", err.Error())
	}
}

func TestRun_PartialAddWithWarnings(t *testing.T) {
	t.Parallel()

	mgr := setupTestManager(t)
	if err := mgr.CreateGroup("kitchen"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}
	if err := mgr.AddDeviceToGroup("kitchen", "light-1"); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	// Add one existing device and one new device
	err := run(f, "kitchen", []string{"light-1", "light-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group, ok := mgr.GetGroup("kitchen")
	if !ok {
		t.Fatal("group not found after add")
	}
	if len(group.Devices) != 2 {
		t.Errorf("group devices = %d, want 2", len(group.Devices))
	}
}

func TestRun_EmptyDeviceList(t *testing.T) {
	t.Parallel()

	mgr := setupTestManager(t)
	if err := mgr.CreateGroup("test-group"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	err := run(f, "test-group", []string{})
	if err == nil {
		t.Fatal("expected error with empty device list")
	}
	if err.Error() != "no devices were added" {
		t.Errorf("error = %q, want 'no devices were added'", err.Error())
	}
}

func TestNewCommand_HasMinimumArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	testCases := []struct {
		args    []string
		wantErr bool
	}{
		{[]string{}, true},
		{[]string{"one"}, true},
		{[]string{"one", "two"}, false},
		{[]string{"one", "two", "three"}, false},
	}

	for _, tc := range testCases {
		err := cmd.Args(cmd, tc.args)
		hasErr := err != nil
		if hasErr != tc.wantErr {
			t.Errorf("Args(%v) error = %v, wantErr = %v", tc.args, err, tc.wantErr)
		}
	}
}

func TestNewCommand_Execute(t *testing.T) {
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
	cmd.SetArgs([]string{"my-group", "device1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

func TestNewCommand_ExecuteSuccess(t *testing.T) {
	t.Parallel()

	mgr := setupTestManager(t)
	if err := mgr.CreateGroup("my-group"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"my-group", "device1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group, ok := mgr.GetGroup("my-group")
	if !ok {
		t.Fatal("group not found")
	}
	if len(group.Devices) != 1 {
		t.Errorf("group devices = %d, want 1", len(group.Devices))
	}
}
