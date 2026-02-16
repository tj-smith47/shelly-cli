package factories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewEnableDisableCommand_Enable(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	called := false
	var receivedDevice string

	opts := factories.EnableDisableOpts{
		Feature:       "Modbus-TCP",
		Enable:        true,
		ExampleParent: "modbus",
		ServiceFunc: func(_ context.Context, _ *cmdutil.Factory, device string) error {
			called = true
			receivedDevice = device
			return nil
		},
	}

	cmd := factories.NewEnableDisableCommand(tf.Factory, opts)

	if cmd.Use != "enable <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "enable <device>")
	}

	hasOn := false
	for _, a := range cmd.Aliases {
		if a == "on" {
			hasOn = true
			break
		}
	}
	if !hasOn {
		t.Error("Aliases should contain 'on'")
	}

	cmd.SetArgs([]string{"test-device"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !called {
		t.Error("ServiceFunc should have been called")
	}
	if receivedDevice != "test-device" {
		t.Errorf("device = %q, want %q", receivedDevice, "test-device")
	}
}

func TestNewEnableDisableCommand_Disable(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	called := false

	opts := factories.EnableDisableOpts{
		Feature:       "cloud connection",
		Enable:        false,
		ExampleParent: "cloud",
		ServiceFunc: func(_ context.Context, _ *cmdutil.Factory, _ string) error {
			called = true
			return nil
		},
	}

	cmd := factories.NewEnableDisableCommand(tf.Factory, opts)

	if cmd.Use != "disable <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "disable <device>")
	}

	hasOff := false
	for _, a := range cmd.Aliases {
		if a == "off" {
			hasOff = true
			break
		}
	}
	if !hasOff {
		t.Error("Aliases should contain 'off'")
	}

	cmd.SetArgs([]string{"test-device"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !called {
		t.Error("ServiceFunc should have been called")
	}
}

func TestNewEnableDisableCommand_CustomAliases(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.EnableDisableOpts{
		Feature: "Matter",
		Enable:  true,
		Aliases: []string{"on", "activate"},
		ServiceFunc: func(_ context.Context, _ *cmdutil.Factory, _ string) error {
			return nil
		},
	}

	cmd := factories.NewEnableDisableCommand(tf.Factory, opts)

	hasActivate := false
	for _, a := range cmd.Aliases {
		if a == "activate" {
			hasActivate = true
			break
		}
	}
	if !hasActivate {
		t.Error("Aliases should contain 'activate'")
	}
}

func TestNewEnableDisableCommand_ServiceError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	expectedErr := errors.New("service unavailable")

	opts := factories.EnableDisableOpts{
		Feature: "Modbus-TCP",
		Enable:  true,
		ServiceFunc: func(_ context.Context, _ *cmdutil.Factory, _ string) error {
			return expectedErr
		},
	}

	cmd := factories.NewEnableDisableCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"test-device"})

	err := cmd.Execute()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestNewEnableDisableCommand_MissingArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.EnableDisableOpts{
		Feature: "Modbus-TCP",
		Enable:  true,
		ServiceFunc: func(_ context.Context, _ *cmdutil.Factory, _ string) error {
			return nil
		},
	}

	cmd := factories.NewEnableDisableCommand(tf.Factory, opts)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error when device argument is missing")
	}
}

func TestNewEnableDisableCommand_PostSuccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	postCalled := false

	opts := factories.EnableDisableOpts{
		Feature: "Matter",
		Enable:  true,
		ServiceFunc: func(_ context.Context, _ *cmdutil.Factory, _ string) error {
			return nil
		},
		PostSuccess: func(_ *cmdutil.Factory, _ string) {
			postCalled = true
		},
	}

	cmd := factories.NewEnableDisableCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"test-device"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !postCalled {
		t.Error("PostSuccess should have been called")
	}
}
