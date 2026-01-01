package status

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "status <device>" {
		t.Errorf("Use = %q, want 'status <device>'", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"st", "watch"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Monitor device status in real-time"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "real-time") {
		t.Error("Long should contain 'real-time'")
	}
	if !strings.Contains(cmd.Long, "Ctrl+C") {
		t.Error("Long should contain 'Ctrl+C'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	expectedPatterns := []string{
		"shelly monitor status",
		"--interval",
		"--count",
	}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	intervalFlag := cmd.Flags().Lookup("interval")
	if intervalFlag == nil {
		t.Fatal("interval flag not found")
	}
	if intervalFlag.Shorthand != "i" {
		t.Errorf("interval shorthand = %q, want i", intervalFlag.Shorthand)
	}
	if intervalFlag.DefValue != "2s" {
		t.Errorf("interval default = %q, want 2s", intervalFlag.DefValue)
	}

	countFlag := cmd.Flags().Lookup("count")
	if countFlag == nil {
		t.Fatal("count flag not found")
	}
	if countFlag.Shorthand != "n" {
		t.Errorf("count shorthand = %q, want n", countFlag.Shorthand)
	}
	if countFlag.DefValue != "0" {
		t.Errorf("count default = %q, want 0", countFlag.DefValue)
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "interval flag short", args: []string{"-i", "5s"}, wantErr: false},
		{name: "interval flag long", args: []string{"--interval", "5s"}, wantErr: false},
		{name: "count flag short", args: []string{"-n", "10"}, wantErr: false},
		{name: "count flag long", args: []string{"--count", "10"}, wantErr: false},
		{name: "both flags", args: []string{"-i", "1s", "-n", "5"}, wantErr: false},
		{name: "no flags", args: []string{}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	helpErr := cmd.Execute()
	if helpErr != nil {
		t.Errorf("--help should not error: %v", helpErr)
	}

	helpOutput := buf.String()
	if !strings.Contains(helpOutput, "Monitor device status") {
		t.Error("Help output should contain command description")
	}
}

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	err := cmd.Args(cmd, []string{"192.168.1.100"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	err := cmd.Args(cmd, []string{"living-room"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestRun_WithMockServer(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {
				"em:0": map[string]any{
					"id":               float64(0),
					"a_current":        float64(1.5),
					"a_voltage":        float64(230.0),
					"a_act_power":      float64(345.0),
					"total_act_power":  float64(1035.0),
				},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)
	demo.InjectIntoFactory(f)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		Interval: 50 * time.Millisecond,
		Count:    2,
	}
	err = run(ctx, opts)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("run returned: %v", err)
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Monitoring") {
		t.Error("Output should contain 'Monitoring' title")
	}
	if !strings.Contains(output, "test-device") {
		t.Error("Output should contain device name")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)
	demo.InjectIntoFactory(f)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		Interval: 2 * time.Second,
	}
	err = run(ctx, opts)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Logf("run returned: %v", err)
	}
}

func TestRun_WithCountLimit(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)
	demo.InjectIntoFactory(f)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		Interval: 20 * time.Millisecond,
		Count:    1,
	}
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestExecute_WithMockServer(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--count", "1", "--interval", "20ms"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err != nil {
		t.Logf("execute returned: %v", err)
	}
}

func TestExecute_MissingDevice(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"nonexistent-device", "--count", "1", "--interval", "10ms"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Logf("execute returned (expected): %v", err)
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval flag: %v", err)
	}
	if interval != 2*time.Second {
		t.Errorf("interval default = %v, want 2s", interval)
	}

	count, err := cmd.Flags().GetInt("count")
	if err != nil {
		t.Fatalf("failed to get count flag: %v", err)
	}
	if count != 0 {
		t.Errorf("count default = %d, want 0", count)
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cmdutil.Factory) bool
		errMsg    string
	}{
		{
			name:      "NewCommand returns non-nil",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f) != nil },
			errMsg:    "NewCommand should not return nil",
		},
		{
			name:      "Has Use field",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f).Use != "" },
			errMsg:    "Use field should not be empty",
		},
		{
			name:      "Has Short field",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f).Short != "" },
			errMsg:    "Short field should not be empty",
		},
		{
			name:      "Has Long field",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f).Long != "" },
			errMsg:    "Long field should not be empty",
		},
		{
			name:      "Has Example field",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f).Example != "" },
			errMsg:    "Example field should not be empty",
		},
		{
			name:      "Has Aliases",
			checkFunc: func(f *cmdutil.Factory) bool { return len(NewCommand(f).Aliases) > 0 },
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "Has RunE",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f).RunE != nil },
			errMsg:    "RunE should be set",
		},
		{
			name:      "Has Args validator",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f).Args != nil },
			errMsg:    "Args should be set",
		},
		{
			name:      "Has ValidArgsFunction",
			checkFunc: func(f *cmdutil.Factory) bool { return NewCommand(f).ValidArgsFunction != nil },
			errMsg:    "ValidArgsFunction should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := cmdutil.NewFactory()
			if !tt.checkFunc(f) {
				t.Error(tt.errMsg)
			}
		})
	}
}
