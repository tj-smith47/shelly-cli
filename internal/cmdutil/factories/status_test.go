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

type TestComponentStatus struct {
	ID     int
	IsOn   bool
	Power  float64
	Bright int
}

func TestNewStatusCommand(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	fetcherCalled := false
	displayCalled := false
	var receivedDevice string
	var receivedID int

	opts := factories.StatusOpts[*TestComponentStatus]{
		Component: "Light",
		Aliases:   []string{"st", "s"},
		Fetcher: func(_ context.Context, _ *shelly.Service, device string, id int) (*TestComponentStatus, error) {
			fetcherCalled = true
			receivedDevice = device
			receivedID = id
			return &TestComponentStatus{ID: id, IsOn: true, Power: 5.5, Bright: 80}, nil
		},
		Display: func(_ *iostreams.IOStreams, status *TestComponentStatus) {
			displayCalled = true
		},
	}

	cmd := factories.NewStatusCommand(tf.Factory, opts)

	// Verify command metadata
	if cmd.Use != "status <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status <device>")
	}

	// Check for expected aliases
	hasSt := false
	for _, a := range cmd.Aliases {
		if a == "st" {
			hasSt = true
			break
		}
	}
	if !hasSt {
		t.Error("Aliases should contain 'st'")
	}

	// Verify --id flag exists
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("--id flag not found")
	}

	// Execute command
	cmd.SetArgs([]string{"my-light"})
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
	if receivedDevice != "my-light" {
		t.Errorf("device = %q, want %q", receivedDevice, "my-light")
	}
	if receivedID != 0 {
		t.Errorf("id = %d, want 0 (default)", receivedID)
	}
}

func TestNewStatusCommand_WithComponentID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	var receivedID int

	opts := factories.StatusOpts[*TestComponentStatus]{
		Component: "RGB",
		Aliases:   []string{"st"},
		Fetcher: func(_ context.Context, _ *shelly.Service, _ string, id int) (*TestComponentStatus, error) {
			receivedID = id
			return &TestComponentStatus{ID: id}, nil
		},
		Display: func(_ *iostreams.IOStreams, _ *TestComponentStatus) {},
	}

	cmd := factories.NewStatusCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"device", "--id", "2"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if receivedID != 2 {
		t.Errorf("id = %d, want 2", receivedID)
	}
}

func TestNewStatusCommand_FetcherError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	expectedErr := errors.New("device not responding")

	opts := factories.StatusOpts[*TestComponentStatus]{
		Component: "Switch",
		Aliases:   []string{"st"},
		Fetcher: func(_ context.Context, _ *shelly.Service, _ string, _ int) (*TestComponentStatus, error) {
			return nil, expectedErr
		},
		Display: func(_ *iostreams.IOStreams, _ *TestComponentStatus) {},
	}

	cmd := factories.NewStatusCommand(tf.Factory, opts)
	cmd.SetArgs([]string{"device"})

	err := cmd.Execute()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestNewStatusCommand_CustomSpinnerMsg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.StatusOpts[*TestComponentStatus]{
		Component:  "Cover",
		Aliases:    []string{"st"},
		SpinnerMsg: "Checking cover position...",
		Fetcher: func(_ context.Context, _ *shelly.Service, _ string, _ int) (*TestComponentStatus, error) {
			return &TestComponentStatus{}, nil
		},
		Display: func(_ *iostreams.IOStreams, _ *TestComponentStatus) {},
	}

	cmd := factories.NewStatusCommand(tf.Factory, opts)

	// The spinner message is used internally, so we just verify the command works
	cmd.SetArgs([]string{"device"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestNewStatusCommand_MissingDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := factories.StatusOpts[*TestComponentStatus]{
		Component: "Light",
		Aliases:   []string{"st"},
		Fetcher: func(_ context.Context, _ *shelly.Service, _ string, _ int) (*TestComponentStatus, error) {
			return &TestComponentStatus{}, nil
		},
		Display: func(_ *iostreams.IOStreams, _ *TestComponentStatus) {},
	}

	cmd := factories.NewStatusCommand(tf.Factory, opts)
	cmd.SetArgs([]string{}) // No device argument

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error when device is missing")
	}
}
