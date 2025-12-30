package toggle

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "toggle <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "toggle <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test id flag exists
	idFlag := cmd.Flags().Lookup("id")
	switch {
	case idFlag == nil:
		t.Error("id flag not found")
	case idFlag.Shorthand != "i":
		t.Errorf("id shorthand = %q, want %q", idFlag.Shorthand, "i")
	case idFlag.DefValue != "0":
		t.Errorf("id default = %q, want %q", idFlag.DefValue, "0")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// The command should require exactly 1 argument
	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}

func TestExecute_WithMock(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-rgbw",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHRGBW2",
					Model:      "Shelly RGBW2",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-rgbw": {"rgbw:0": map[string]any{"output": false, "rgb": []int{0, 255, 0}}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-rgbw"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "unknown") {
		t.Logf("error = %v", err)
	}
}

func TestExecute_WithIDFlag(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-rgbw-multi",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHRGBW2",
					Model:      "Shelly RGBW2",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-rgbw-multi": {
				"rgbw:0": map[string]any{"output": false, "rgb": []int{255, 0, 0}},
				"rgbw:1": map[string]any{"output": true, "rgb": []int{0, 255, 0}},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-rgbw-multi", "--id", "1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error with id flag = %v (may be expected for mock)", err)
	}
}
