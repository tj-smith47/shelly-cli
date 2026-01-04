package factories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testDeviceName = "test-device"

func TestNewComponentCommand_On(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	actionCalled := false
	var receivedDevice string
	var receivedID int

	opts := factories.ComponentOpts{
		Component: "Light",
		Action:    factories.ActionOn,
		SimpleFunc: func(_ context.Context, _ *shelly.Service, device string, id int) error {
			actionCalled = true
			receivedDevice = device
			receivedID = id
			return nil
		},
	}

	cmd := factories.NewComponentCommand(tf.Factory, opts)

	// Verify command metadata
	if cmd.Use != "on <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "on <device>")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases should not be empty")
	}

	// Check for expected aliases
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

	// Verify --id flag exists
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("--id flag not found")
	}

	// Execute command
	cmd.SetArgs([]string{testDeviceName})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !actionCalled {
		t.Error("action should have been called")
	}
	if receivedDevice != testDeviceName {
		t.Errorf("device = %q, want %q", receivedDevice, testDeviceName)
	}
	if receivedID != 0 {
		t.Errorf("id = %d, want 0 (default)", receivedID)
	}
}

func TestNewComponentCommand_Off(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	actionCalled := false

	opts := factories.ComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOff,
		SimpleFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			actionCalled = true
			return nil
		},
	}

	cmd := factories.NewComponentCommand(tf.Factory, opts)

	if cmd.Use != "off <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "off <device>")
	}

	cmd.SetArgs([]string{"my-switch"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !actionCalled {
		t.Error("action should have been called")
	}
}

func TestNewComponentCommand_Toggle(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	toggleCalled := false

	opts := factories.ComponentOpts{
		Component: "RGB",
		Action:    factories.ActionToggle,
		ToggleFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) (bool, error) {
			toggleCalled = true
			return true, nil // Return that it's now ON
		},
	}

	cmd := factories.NewComponentCommand(tf.Factory, opts)

	if cmd.Use != "toggle <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "toggle <device>")
	}

	// Check for expected aliases
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

	cmd.SetArgs([]string{"rgb-device"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !toggleCalled {
		t.Error("toggle function should have been called")
	}
}

func TestNewComponentCommand_WithComponentID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	var receivedID int

	opts := factories.ComponentOpts{
		Component: "Light",
		Action:    factories.ActionOn,
		SimpleFunc: func(_ context.Context, _ *shelly.Service, _ string, id int) error {
			receivedID = id
			return nil
		},
	}

	cmd := factories.NewComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"device", "--id", "3"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if receivedID != 3 {
		t.Errorf("id = %d, want 3", receivedID)
	}
}

func TestNewComponentCommand_ActionError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	expectedErr := errors.New("device unavailable")

	opts := factories.ComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOn,
		SimpleFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return expectedErr
		},
	}

	cmd := factories.NewComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"device"})

	err := cmd.Execute()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestNewComponentCommand_ToggleError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	expectedErr := errors.New("toggle failed")

	opts := factories.ComponentOpts{
		Component: "RGB",
		Action:    factories.ActionToggle,
		ToggleFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) (bool, error) {
			return false, expectedErr
		},
	}

	cmd := factories.NewComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"device"})

	err := cmd.Execute()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestNewComponentCommand_MissingArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.ComponentOpts{
		Component:  "Light",
		Action:     factories.ActionOn,
		SimpleFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int) error { return nil },
	}

	cmd := factories.NewComponentCommand(tf.Factory, opts)
	cmd.SetArgs([]string{}) // No device argument

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error when device argument is missing")
	}
}

func TestNewListCommand(t *testing.T) {
	t.Parallel()

	type TestItem struct {
		ID   int
		Name string
	}

	tf := factory.NewTestFactory(t)
	fetcherCalled := false
	displayCalled := false

	opts := factories.ListOpts[TestItem]{
		Component: "Light",
		Long:      "List all light components on a device.",
		Example:   "  shelly light list mydevice",
		Fetcher: func(_ context.Context, _ *shelly.Service, _ string) ([]TestItem, error) {
			fetcherCalled = true
			return []TestItem{{ID: 0, Name: "Light 0"}, {ID: 1, Name: "Light 1"}}, nil
		},
		Display: func(_ *iostreams.IOStreams, items []TestItem) {
			displayCalled = true
		},
	}

	cmd := factories.NewListCommand(tf.Factory, opts)

	// Verify command metadata
	if cmd.Use != "list <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device>")
	}

	hasLs := false
	for _, a := range cmd.Aliases {
		if a == "ls" {
			hasLs = true
			break
		}
	}
	if !hasLs {
		t.Error("Aliases should contain 'ls'")
	}

	// Execute command
	cmd.SetArgs([]string{"my-device"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !fetcherCalled {
		t.Error("fetcher should have been called")
	}
	if !displayCalled {
		t.Error("display should have been called")
	}
}

func TestNewListCommand_EmptyList(t *testing.T) {
	t.Parallel()

	type TestItem struct {
		ID int
	}

	tf := factory.NewTestFactory(t)
	displayCalled := false

	opts := factories.ListOpts[TestItem]{
		Component: "Cover",
		Long:      "List covers",
		Example:   "  shelly cover list dev",
		Fetcher: func(_ context.Context, _ *shelly.Service, _ string) ([]TestItem, error) {
			return []TestItem{}, nil // Empty list
		},
		Display: func(_ *iostreams.IOStreams, _ []TestItem) {
			displayCalled = true
		},
	}

	cmd := factories.NewListCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"device"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if displayCalled {
		t.Error("display should not be called for empty list")
	}

	// Check that a "no results" message was printed
	output := tf.OutString()
	if output == "" {
		t.Error("should have printed 'no results' message")
	}
}

func TestNewListCommand_FetcherError(t *testing.T) {
	t.Parallel()

	type TestItem struct {
		ID int
	}

	tf := factory.NewTestFactory(t)
	expectedErr := errors.New("fetch failed")

	opts := factories.ListOpts[TestItem]{
		Component: "Input",
		Long:      "List inputs",
		Example:   "  shelly input list dev",
		Fetcher: func(_ context.Context, _ *shelly.Service, _ string) ([]TestItem, error) {
			return nil, expectedErr
		},
		Display: func(_ *iostreams.IOStreams, _ []TestItem) {},
	}

	cmd := factories.NewListCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"device"})

	err := cmd.Execute()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestActionConstants(t *testing.T) {
	t.Parallel()

	if factories.ActionOn != "on" {
		t.Errorf("ActionOn = %q, want %q", factories.ActionOn, "on")
	}
	if factories.ActionOff != "off" {
		t.Errorf("ActionOff = %q, want %q", factories.ActionOff, "off")
	}
	if factories.ActionToggle != "toggle" {
		t.Errorf("ActionToggle = %q, want %q", factories.ActionToggle, "toggle")
	}
}
