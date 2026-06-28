package off

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

	if cmd.Use != "off <group>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "off <group>")
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

	idFlag := cmd.Flags().Lookup("id")
	switch {
	case idFlag == nil:
		t.Error("id flag not found")
	case idFlag.DefValue != "-1":
		t.Errorf("id default = %q, want %q", idFlag.DefValue, "-1")
	}

	concurrentFlag := cmd.Flags().Lookup("concurrent")
	if concurrentFlag == nil {
		t.Error("concurrent flag not found")
	}
}

func TestExecute_GroupNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"nonexistent-group"})

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
	cmd.SetArgs([]string{"empty-group"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty group")
	}
	if !strings.Contains(err.Error(), "no devices") {
		t.Errorf("expected 'no devices' error, got: %v", err)
	}
}
