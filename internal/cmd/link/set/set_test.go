package set

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

func setupTestManager(t *testing.T, devices ...string) *config.Manager {
	t.Helper()
	cfg := &config.Config{}
	mgr := config.NewTestManager(cfg)
	for _, name := range devices {
		cfg.Devices[name] = model.Device{Name: name, Address: "192.168.1.1"}
	}
	return mgr
}

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "set <child-device> <parent-device>" {
		t.Errorf("Use = %q", cmd.Use)
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases should not be empty")
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_Success(t *testing.T) {
	mgr := setupTestManager(t, "bulb-duo", "bedroom-2pm")
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f, ChildDevice: "bulb-duo", ParentDevice: "bedroom-2pm", SwitchID: 0}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error: %v", err)
	}

	// Verify link was created
	link, ok := mgr.GetLink("bulb-duo")
	if !ok {
		t.Fatal("link should have been created")
	}
	if link.ParentDevice != "bedroom-2pm" {
		t.Errorf("parent = %q, want %q", link.ParentDevice, "bedroom-2pm")
	}

	output := out.String()
	if !strings.Contains(output, "Link set") {
		t.Errorf("output should contain 'Link set', got: %s", output)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_SelfLink(t *testing.T) {
	mgr := setupTestManager(t, "device-a")
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f, ChildDevice: "device-a", ParentDevice: "device-a"}
	err := run(opts)
	if err == nil {
		t.Fatal("expected error for self-link")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestRun_DeviceNotFound(t *testing.T) {
	mgr := setupTestManager(t, "child")
	config.SetDefaultManager(mgr)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	opts := &Options{Factory: f, ChildDevice: "child", ParentDevice: "nonexistent"}
	err := run(opts)
	if err == nil {
		t.Fatal("expected error for nonexistent parent")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found', got: %v", err)
	}
}

func TestNewCommand_Execute_Help(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Requires exactly 2 args
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"child"}); err == nil {
		t.Error("expected error with 1 arg")
	}
	if err := cmd.Args(cmd, []string{"child", "parent"}); err != nil {
		t.Errorf("unexpected error with 2 args: %v", err)
	}
}
