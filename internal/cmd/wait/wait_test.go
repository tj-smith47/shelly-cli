package wait

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testDevice = "test-device"

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "wait <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "wait <device>")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}
	if cmd.Short == "" || cmd.Long == "" {
		t.Error("Short and Long must be set")
	}
	if !strings.Contains(cmd.Example, "shelly wait kitchen") {
		t.Errorf("Example should show a device wait, got: %s", cmd.Example)
	}
	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should complete device names")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	defaults := map[string]string{"online": "true", "state": "", "id": "-1", "timeout": "2m0s", "interval": "2s"}
	for name, want := range defaults {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("flag --%s not found", name)
			continue
		}
		if flag.DefValue != want {
			t.Errorf("--%s default = %q, want %q", name, flag.DefValue, want)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one device", []string{"kitchen"}, false},
		{"two args", []string{"kitchen", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// The example scripts call the command in this exact form.
//
//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_ReturnsOnceDeviceAnswers(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       testDevice,
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
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

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{testDevice, "--online", "--timeout", "120s"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out := tf.OutString(); !strings.Contains(out, "is online") {
		t.Errorf("output should report the device online, got: %s", out)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_TimesOutWhenDeviceNeverAnswers(t *testing.T) {
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "nonexistent",
		Online:   true,
		Timeout:  50 * time.Millisecond,
		Interval: 10 * time.Millisecond,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if !strings.Contains(err.Error(), "did not come online within 50ms") {
		t.Errorf("error = %v, want a timeout message", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_StopsWhenCancelled(t *testing.T) {
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "nonexistent",
		Online:   true,
		Timeout:  time.Hour,
		Interval: time.Hour,
	}

	if err := run(ctx, opts); !errors.Is(err, context.Canceled) {
		t.Errorf("run() error = %v, want context.Canceled", err)
	}
}

func TestRun_RejectsOnlineFalse(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), &Options{Factory: tf.Factory, Device: "kitchen"})
	if err == nil {
		t.Fatal("expected --online=false to be rejected")
	}
}

func TestRun_RejectsUnknownState(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), &Options{Factory: tf.Factory, Device: "kitchen", Online: true, State: "dim"})
	if err == nil || !strings.Contains(err.Error(), "must be on or off") {
		t.Fatalf("run() error = %v, want an invalid --state error", err)
	}
}

// The example aliases call "wait --state on" and "wait --state off".
//
//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_WaitsForOutputState(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       testDevice,
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			testDevice: {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	on := &Options{
		Factory: tf.Factory, Device: testDevice, Online: true, State: stateOn,
		Timeout: 5 * time.Second, Interval: 10 * time.Millisecond,
	}
	on.ID = -1
	if err := run(context.Background(), on); err != nil {
		t.Fatalf("run(--state on) error = %v", err)
	}
	if out := tf.OutString(); !strings.Contains(out, "is on") {
		t.Errorf("output should report the device on, got: %s", out)
	}

	off := &Options{
		Factory: tf.Factory, Device: testDevice, Online: true, State: stateOff,
		Timeout: 50 * time.Millisecond, Interval: 10 * time.Millisecond,
	}
	off.ID = -1
	err = run(context.Background(), off)
	if err == nil || !strings.Contains(err.Error(), "did not turn off within 50ms") {
		t.Fatalf("run(--state off) error = %v, want a timeout", err)
	}
}
