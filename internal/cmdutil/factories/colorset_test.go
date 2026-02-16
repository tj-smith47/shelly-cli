package factories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testDevice = "test-device"

func TestNewColorSetCommand_RGB(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	called := false
	var gotDevice string
	var gotRed, gotGreen, gotBlue, gotWhite, gotBrightness int

	opts := factories.ColorSetOpts{
		Component: "RGB",
		HasWhite:  false,
		SetFunc: func(_ context.Context, _ *cmdutil.Factory, device string, _, red, green, blue, white, brightness int, _ bool) error {
			called = true
			gotDevice = device
			gotRed = red
			gotGreen = green
			gotBlue = blue
			gotWhite = white
			gotBrightness = brightness
			return nil
		},
	}

	cmd := factories.NewColorSetCommand(tf.Factory, opts)

	if cmd.Use != "set <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <device>")
	}

	// Verify --white flag does NOT exist for RGB
	if cmd.Flags().Lookup("white") != nil {
		t.Error("RGB command should not have --white flag")
	}

	cmd.SetArgs([]string{testDevice, "--red", "255", "--green", "128", "--blue", "0", "--brightness", "75"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !called {
		t.Error("SetFunc should have been called")
	}
	if gotDevice != testDevice {
		t.Errorf("device = %q, want %q", gotDevice, testDevice)
	}
	if gotRed != 255 {
		t.Errorf("red = %d, want 255", gotRed)
	}
	if gotGreen != 128 {
		t.Errorf("green = %d, want 128", gotGreen)
	}
	if gotBlue != 0 {
		t.Errorf("blue = %d, want 0", gotBlue)
	}
	if gotWhite != -1 {
		t.Errorf("white = %d, want -1 (unset for RGB)", gotWhite)
	}
	if gotBrightness != 75 {
		t.Errorf("brightness = %d, want 75", gotBrightness)
	}
}

func TestNewColorSetCommand_RGBW(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	called := false
	var gotWhite int

	opts := factories.ColorSetOpts{
		Component: "RGBW",
		HasWhite:  true,
		SetFunc: func(_ context.Context, _ *cmdutil.Factory, _ string, _, _, _, _, white, _ int, _ bool) error {
			called = true
			gotWhite = white
			return nil
		},
	}

	cmd := factories.NewColorSetCommand(tf.Factory, opts)

	// Verify --white flag exists for RGBW
	if cmd.Flags().Lookup("white") == nil {
		t.Error("RGBW command should have --white flag")
	}

	cmd.SetArgs([]string{testDevice, "--white", "200"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !called {
		t.Error("SetFunc should have been called")
	}
	if gotWhite != 200 {
		t.Errorf("white = %d, want 200", gotWhite)
	}
}

func TestNewColorSetCommand_Aliases(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.ColorSetOpts{
		Component: "RGB",
		SetFunc:   func(_ context.Context, _ *cmdutil.Factory, _ string, _, _, _, _, _, _ int, _ bool) error { return nil },
	}

	cmd := factories.NewColorSetCommand(tf.Factory, opts)

	hasColor := false
	for _, a := range cmd.Aliases {
		if a == "color" {
			hasColor = true
			break
		}
	}
	if !hasColor {
		t.Error("Aliases should contain 'color'")
	}
}

func TestNewColorSetCommand_ServiceError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	expectedErr := errors.New("set failed")

	opts := factories.ColorSetOpts{
		Component: "RGB",
		SetFunc: func(_ context.Context, _ *cmdutil.Factory, _ string, _, _, _, _, _, _ int, _ bool) error {
			return expectedErr
		},
	}

	cmd := factories.NewColorSetCommand(tf.Factory, opts)
	cmd.SetArgs([]string{testDevice})

	err := cmd.Execute()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestNewColorSetCommand_MissingArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.ColorSetOpts{
		Component: "RGB",
		SetFunc:   func(_ context.Context, _ *cmdutil.Factory, _ string, _, _, _, _, _, _ int, _ bool) error { return nil },
	}

	cmd := factories.NewColorSetCommand(tf.Factory, opts)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error when device argument is missing")
	}
}
