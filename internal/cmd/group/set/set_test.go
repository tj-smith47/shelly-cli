package set

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "set <group>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <group>")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	brightnessFlag := cmd.Flags().Lookup("brightness")
	switch {
	case brightnessFlag == nil:
		t.Error("brightness flag not found")
	case brightnessFlag.Shorthand != "b":
		t.Errorf("brightness shorthand = %q, want %q", brightnessFlag.Shorthand, "b")
	case brightnessFlag.DefValue != "-1":
		t.Errorf("brightness default = %q, want %q", brightnessFlag.DefValue, "-1")
	}

	tempFlag := cmd.Flags().Lookup("temp")
	if tempFlag == nil {
		t.Error("temp flag not found")
	}

	onFlag := cmd.Flags().Lookup("on")
	if onFlag == nil {
		t.Error("on flag not found")
	}

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("id flag not found")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Args == nil {
		t.Fatal("Args validator not set")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"grp"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
}

func TestExecute_GroupNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"nonexistent-group", "-b", "50"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestExecute_EmptyGroup(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"empty-group": {Devices: []string{}},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"empty-group", "-b", "50"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty group")
	}
	if !strings.Contains(err.Error(), "no devices") {
		t.Errorf("expected 'no devices' error, got: %v", err)
	}
}
