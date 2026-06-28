package setaddress

import (
	"context"
	"slices"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "set-address <name> <address>" {
		t.Errorf("Use = %q, want 'set-address <name> <address>'", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "set-addr" {
		t.Errorf("Aliases = %v, want [set-addr readdress]", cmd.Aliases)
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
	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_RequiresTwoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.Args(cmd, []string{"only-one"}); err == nil {
		t.Error("expected error with one arg")
	}
	if err := cmd.Args(cmd, []string{"name", "10.0.0.1"}); err != nil {
		t.Errorf("expected no error with two args, got %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b", "c"}); err == nil {
		t.Error("expected error with three args")
	}
}

func TestNewCommand_HasNoVerifyFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("no-verify")
	if flag == nil {
		t.Fatal("no-verify flag not found")
	}
	if flag.DefValue != "false" {
		t.Errorf("no-verify default = %q, want false", flag.DefValue)
	}
}

//nolint:paralleltest // mutates the global default config manager via SetupTestFs
func TestRun_DeviceNotFound(t *testing.T) {
	factory.SetupTestFs(t)
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Name:     "nonexistent",
		Address:  "10.23.47.219",
		NoVerify: true,
	}

	if err := run(context.Background(), opts); err == nil {
		t.Error("expected error for a device that is not registered")
	}
}

//nolint:paralleltest // mutates the global default config manager via SetupTestFs
func TestRun_AlreadyAtAddressIsNoOp(t *testing.T) {
	factory.SetupTestFs(t)
	tf := factory.NewTestFactory(t)

	if err := config.RegisterDevice("gb", "10.23.47.219", 1, "", "", nil); err != nil {
		t.Fatalf("seed device: %v", err)
	}

	opts := &Options{
		Factory:  tf.Factory,
		Name:     "gb",
		Address:  "10.23.47.219",
		NoVerify: true,
	}

	if err := run(context.Background(), opts); err != nil {
		t.Errorf("run() error = %v, want nil for unchanged address", err)
	}
}

// TestRun_ChangesAddressAndPreservesGroupMembership is the core guarantee:
// re-addressing a device must update only its address while keeping it in every
// group — the safe alternative to remove+add, which drops group membership.
//
//nolint:paralleltest // mutates the global default config manager via SetupTestFs
func TestRun_ChangesAddressAndPreservesGroupMembership(t *testing.T) {
	factory.SetupTestFs(t)
	tf := factory.NewTestFactory(t)

	if err := config.RegisterDevice("gb", "10.23.47.229", 1, "", "", nil); err != nil {
		t.Fatalf("seed device: %v", err)
	}
	if err := config.CreateGroup("guest-bath-bulbs"); err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := config.AddDeviceToGroup("guest-bath-bulbs", "gb"); err != nil {
		t.Fatalf("add to group: %v", err)
	}

	opts := &Options{
		Factory:  tf.Factory,
		Name:     "gb",
		Address:  "10.23.47.219",
		NoVerify: true,
	}

	if err := run(context.Background(), opts); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	dev, ok := config.GetDevice("gb")
	if !ok {
		t.Fatal("device gb vanished after set-address")
	}
	if dev.Address != "10.23.47.219" {
		t.Errorf("address = %q, want 10.23.47.219", dev.Address)
	}

	group, ok := config.GetGroup("guest-bath-bulbs")
	if !ok {
		t.Fatal("group guest-bath-bulbs vanished after set-address")
	}
	if !slices.Contains(group.Devices, "gb") {
		t.Errorf("gb was dropped from its group; members = %v", group.Devices)
	}
}
