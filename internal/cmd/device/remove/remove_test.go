package remove

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand_YesFlagAndHiddenForceAlias(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	yes := cmd.Flags().Lookup("yes")
	if yes == nil || yes.Shorthand != "y" {
		t.Fatalf("--yes/-y not registered: %+v", yes)
	}
	force := cmd.Flags().Lookup("force")
	if force == nil || !force.Hidden {
		t.Fatalf("--force must stay registered as a hidden alias: %+v", force)
	}
	if err := cmd.ParseFlags([]string{"--force"}); err != nil {
		t.Fatalf("ParseFlags(--force) error = %v", err)
	}
	if yes.Value.String() != "true" {
		t.Error("--force must set the same variable as --yes")
	}
}

//nolint:paralleltest // Installs the global config manager.
func TestRun_ConfirmGate(t *testing.T) {
	factory.SetupTestFs(t)
	mgr := config.NewTestManager(&config.Config{Devices: map[string]model.Device{
		"kitchen": {Name: "kitchen", Address: "192.168.1.10"},
	}})
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), &Options{Factory: tf.Factory, Name: "kitchen"})
	if !errors.Is(err, cmdutil.ErrConfirmationRequired) {
		t.Fatalf("no terminal, no --yes: error = %v, want ErrConfirmationRequired", err)
	}
	if _, ok := config.GetDevice("kitchen"); !ok {
		t.Fatal("device removed without confirmation")
	}

	if err := run(context.Background(), &Options{Factory: tf.Factory, Name: "kitchen", Yes: true}); err != nil {
		t.Fatalf("--yes: error = %v", err)
	}
	if _, ok := config.GetDevice("kitchen"); ok {
		t.Fatal("device still registered after --yes")
	}
}
