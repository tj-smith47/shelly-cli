package factories_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewBatchComponentCommand_On(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1"},
		"dev2": {Name: "dev2", Address: "192.168.1.2"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	var callCount atomic.Int32

	opts := factories.BatchComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOn,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			callCount.Add(1)
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)

	// Verify command metadata
	if cmd.Use != "on [device...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "on [device...]")
	}

	hasEnable := false
	for _, a := range cmd.Aliases {
		if a == "enable" {
			hasEnable = true
			break
		}
	}
	if !hasEnable {
		t.Error("Aliases should contain 'enable'")
	}

	// Execute command with devices
	cmd.SetArgs([]string{"dev1", "dev2"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if callCount.Load() != 2 {
		t.Errorf("service func called %d times, want 2", callCount.Load())
	}
}

func TestNewBatchComponentCommand_Off(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"light1": {Name: "light1", Address: "192.168.1.10"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	called := false

	opts := factories.BatchComponentOpts{
		Component: "Light",
		Action:    factories.ActionOff,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			called = true
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)

	if cmd.Use != "off [device...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "off [device...]")
	}

	cmd.SetArgs([]string{"light1"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !called {
		t.Error("service func should have been called")
	}
}

func TestNewBatchComponentCommand_Toggle(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"rgb1": {Name: "rgb1", Address: "192.168.1.20"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	opts := factories.BatchComponentOpts{
		Component: "RGB",
		Action:    factories.ActionToggle,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)

	if cmd.Use != "toggle [device...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "toggle [device...]")
	}

	hasFlip := false
	for _, a := range cmd.Aliases {
		if a == "flip" {
			hasFlip = true
			break
		}
	}
	if !hasFlip {
		t.Error("Aliases should contain 'flip'")
	}
}

func TestNewBatchComponentCommand_MultipleDevices(t *testing.T) {
	t.Parallel()

	// Test with multiple devices passed as args
	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1"},
		"dev2": {Name: "dev2", Address: "192.168.1.2"},
		"dev3": {Name: "dev3", Address: "192.168.1.3"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	var callCount atomic.Int32

	opts := factories.BatchComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOn,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			callCount.Add(1)
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"dev1", "dev2", "dev3"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if callCount.Load() != 3 {
		t.Errorf("service func called %d times, want 3", callCount.Load())
	}
}

func TestNewBatchComponentCommand_Flags(t *testing.T) {
	t.Parallel()

	// Test that all expected flags are present
	tf := factory.NewTestFactory(t)

	opts := factories.BatchComponentOpts{
		Component: "Light",
		Action:    factories.ActionOff,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)

	// Check for expected flags
	expectedFlags := []string{"group", "all", "timeout", "light", "concurrent"}
	for _, name := range expectedFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("--%s flag not found", name)
		}
	}

	// Check short flags
	groupFlag := cmd.Flags().ShorthandLookup("g")
	if groupFlag == nil {
		t.Error("-g shorthand not found")
	}
	allFlag := cmd.Flags().ShorthandLookup("a")
	if allFlag == nil {
		t.Error("-a shorthand not found")
	}
	timeoutFlag := cmd.Flags().ShorthandLookup("t")
	if timeoutFlag == nil {
		t.Error("-t shorthand not found")
	}
	concurrentFlag := cmd.Flags().ShorthandLookup("c")
	if concurrentFlag == nil {
		t.Error("-c shorthand not found")
	}
}

func TestNewBatchComponentCommand_WithComponentID(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	var receivedID int

	opts := factories.BatchComponentOpts{
		Component: "Light",
		Action:    factories.ActionOn,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, id int) error {
			receivedID = id
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"dev1", "--light", "2"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if receivedID != 2 {
		t.Errorf("componentID = %d, want 2", receivedID)
	}
}

func TestNewBatchComponentCommand_SwitchShortFlag(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	var receivedID int

	opts := factories.BatchComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOn,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, id int) error {
			receivedID = id
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)

	// Switch should have -s shorthand
	switchFlag := cmd.Flags().Lookup("switch")
	if switchFlag == nil {
		t.Fatal("--switch flag not found")
	}
	if switchFlag.Shorthand != "s" {
		t.Errorf("switch flag shorthand = %q, want %q", switchFlag.Shorthand, "s")
	}

	cmd.SetArgs([]string{"dev1", "-s", "3"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if receivedID != 3 {
		t.Errorf("componentID = %d, want 3", receivedID)
	}
}

func TestNewBatchComponentCommand_NoTargets(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.BatchComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOn,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{}) // No targets

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error when no targets specified")
	}
}

func TestNewBatchComponentCommand_ServiceError(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	expectedErr := errors.New("device unreachable")

	opts := factories.BatchComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOn,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return expectedErr
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"dev1"})

	// Batch operations don't fail the whole command for individual device errors
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil (batch errors are logged)", err)
	}
}

func TestNewBatchComponentCommand_DefaultComponent(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	opts := factories.BatchComponentOpts{
		// Component is empty - should default to "Switch"
		Action: factories.ActionOn,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return nil
		},
	}

	cmd := factories.NewBatchComponentCommand(tf.Factory, opts)

	// Should have --switch flag with -s shorthand
	switchFlag := cmd.Flags().Lookup("switch")
	if switchFlag == nil {
		t.Error("--switch flag not found (default component should be Switch)")
	}
}
