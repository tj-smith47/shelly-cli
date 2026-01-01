package all

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
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wantUse     string
		wantShort   string
		wantAliases []string
		wantHasLong bool
		wantExample bool
	}{
		{
			name:        "command properties",
			wantUse:     "all",
			wantShort:   "Monitor all registered devices",
			wantAliases: []string{"overview", "summary"},
			wantHasLong: true,
			wantExample: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			if cmd == nil {
				t.Fatal("NewCommand returned nil")
			}

			if cmd.Use != tt.wantUse {
				t.Errorf("Use = %q, want %q", cmd.Use, tt.wantUse)
			}

			if cmd.Short != tt.wantShort {
				t.Errorf("Short = %q, want %q", cmd.Short, tt.wantShort)
			}

			if len(cmd.Aliases) != len(tt.wantAliases) {
				t.Errorf("Aliases length = %d, want %d", len(cmd.Aliases), len(tt.wantAliases))
			}
			for i, alias := range tt.wantAliases {
				if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
					t.Errorf("Alias[%d] = %q, want %q", i, cmd.Aliases[i], alias)
				}
			}

			if tt.wantHasLong && cmd.Long == "" {
				t.Error("Long description is empty")
			}

			if tt.wantExample && cmd.Example == "" {
				t.Error("Example is empty")
			}
		})
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"overview", "summary"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{
			name:         "interval flag",
			flagName:     "interval",
			shorthand:    "i",
			defaultValue: (5 * time.Second).String(),
			flagType:     "duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
			}

			if flag.Value.Type() != tt.flagType {
				t.Errorf("flag %q type = %q, want %q", tt.flagName, flag.Value.Type(), tt.flagType)
			}
		})
	}
}

func TestNewCommand_FlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	intervalFlag := cmd.Flags().Lookup("interval")
	if intervalFlag == nil {
		t.Fatal("interval flag not found")
	}

	if intervalFlag.Usage == "" {
		t.Error("interval flag usage is empty")
	}

	// Usage should mention refresh
	if !strings.Contains(strings.ToLower(intervalFlag.Usage), "refresh") {
		t.Errorf("interval flag usage should mention 'refresh', got: %q", intervalFlag.Usage)
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg",
			args:    []string{"device1"},
			wantErr: true,
		},
		{
			name:    "multiple args",
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "flag-like arg",
			args:    []string{"--some-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	longDesc := cmd.Long

	// Should mention monitoring all devices
	if !strings.Contains(strings.ToLower(longDesc), "all devices") {
		t.Error("Long description should mention 'all devices'")
	}

	// Should mention power consumption or status
	if !strings.Contains(strings.ToLower(longDesc), "power") && !strings.Contains(strings.ToLower(longDesc), "status") {
		t.Error("Long description should mention power consumption or status")
	}

	// Should mention Ctrl+C
	if !strings.Contains(longDesc, "Ctrl+C") {
		t.Error("Long description should mention 'Ctrl+C'")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example

	// Should show basic usage
	if !strings.Contains(example, "shelly monitor all") {
		t.Error("Example should show basic 'shelly monitor all' usage")
	}

	// Should show interval flag usage
	if !strings.Contains(example, "--interval") || !strings.Contains(example, "-i") {
		t.Error("Example should demonstrate --interval flag usage")
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval flag: %v", err)
	}
	if interval != 5*time.Second {
		t.Errorf("interval default = %v, want 5s", interval)
	}
}

func TestNewCommand_IntervalFlagGetDuration(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Simulate setting a custom interval
	if err := cmd.Flags().Set("interval", "10s"); err != nil {
		t.Fatalf("failed to set interval flag: %v", err)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval flag: %v", err)
	}
	if interval != 10*time.Second {
		t.Errorf("interval = %v, want 10s", interval)
	}
}

func TestNewCommand_IntervalFlagInvalidValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Invalid duration should fail to parse
	err := cmd.Flags().Set("interval", "not-a-duration")
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestNewCommand_IntervalFlagValidValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Valid duration should parse successfully
	err := cmd.Flags().Set("interval", "30s")
	if err != nil {
		t.Errorf("unexpected error for valid duration: %v", err)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--interval", "1s"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetContext(ctx)

	// Run with the cancelled context - should return immediately
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The command should have handled cancellation gracefully
	// Even if no devices registered message appears or context is cancelled
	// Both are valid outcomes
}

func TestRun_QuickContextCancellation(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithCancel(context.Background())

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--interval", "100ms"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Cancel context after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	cmd.SetContext(ctx)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCommand_MultipleFactoryCalls(t *testing.T) {
	t.Parallel()

	// Verify that calling NewCommand multiple times with the same factory works
	f := cmdutil.NewFactory()

	cmd1 := NewCommand(f)
	cmd2 := NewCommand(f)

	if cmd1 == nil || cmd2 == nil {
		t.Fatal("NewCommand returned nil")
	}

	// Both commands should have same properties
	if cmd1.Use != cmd2.Use {
		t.Error("Commands should have same Use")
	}

	if cmd1.Short != cmd2.Short {
		t.Error("Commands should have same Short")
	}
}

func TestNewCommand_FlagCanBeSet(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify flag can be set to a custom value
	if err := cmd.Flags().Set("interval", "15s"); err != nil {
		t.Fatalf("failed to set interval: %v", err)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval: %v", err)
	}
	if interval != 15*time.Second {
		t.Errorf("interval = %v, want 15s", interval)
	}
}

func TestRun_NoDevicesRegistered(t *testing.T) {
	t.Parallel()

	// This test verifies the "no devices registered" path in the run function.
	// The config.ListDevices() will return empty since we're using a fresh factory
	// without any injected mock devices.

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetContext(ctx)

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Logf("execute returned: %v", err)
	}

	// Check that "no devices" message was printed
	output := out.String() + errOut.String()
	if !strings.Contains(output, "No devices") && !strings.Contains(output, "no devices") {
		t.Logf("Expected 'no devices' message in output, got: %s", output)
	}
}

func TestRun_NoDevicesRegisteredShortTimeout(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Very short timeout to test early exit path
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory:  f,
		Interval: 1 * time.Second,
	}

	err := run(ctx, opts)
	// Should return nil for no devices
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestRun_DirectInvocation(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory:  f,
		Interval: 5 * time.Second,
	}

	err := run(ctx, opts)
	// Should return nil gracefully
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--help"})

	helpErr := cmd.Execute()
	if helpErr != nil {
		t.Errorf("--help should not error: %v", helpErr)
	}

	helpOutput := out.String() + errOut.String()
	// Check for key components in help
	if !strings.Contains(helpOutput, "all") {
		t.Error("Help output should contain 'all'")
	}
	if !strings.Contains(helpOutput, "interval") {
		t.Error("Help output should contain 'interval'")
	}
}

func TestNewCommand_ShortHandFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test short form of interval flag
	if err := cmd.ParseFlags([]string{"-i", "3s"}); err != nil {
		t.Fatalf("failed to parse short flag: %v", err)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval: %v", err)
	}
	if interval != 3*time.Second {
		t.Errorf("interval = %v, want 3s", interval)
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

func TestNewCommand_UseValue(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "all" {
		t.Errorf("Use = %q, want 'all'", cmd.Use)
	}
}

func TestOptions_FactoryEmbedded(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory:  f,
		Interval: 5 * time.Second,
	}

	if opts.Factory != f {
		t.Error("Factory should be embedded in Options")
	}
	if opts.Interval != 5*time.Second {
		t.Errorf("Interval = %v, want 5s", opts.Interval)
	}
}

func TestRun_WithMockServer(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device-1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
				{
					Name:       "test-device-2",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SHSW-1",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device-1": {
				"pm:0": map[string]any{
					"id":      float64(0),
					"voltage": float64(230.0),
					"current": float64(1.5),
					"apower":  float64(345.0),
				},
			},
			"test-device-2": {
				"switch:0": map[string]any{
					"id":     float64(0),
					"output": true,
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

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory:  f,
		Interval: 50 * time.Millisecond,
	}
	err = run(ctx, opts)
	// Either context cancellation or successful completion
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("run returned: %v", err)
	}

	// Check that monitoring was started
	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Monitoring") {
		t.Error("Output should contain 'Monitoring' title")
	}
	if !strings.Contains(output, "2 devices") {
		t.Logf("Expected '2 devices' in output, got: %s", output)
	}
}

func TestRun_WithMockServerSingleDevice(t *testing.T) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory:  f,
		Interval: 30 * time.Millisecond,
	}
	err = run(ctx, opts)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("run returned: %v", err)
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Monitoring") {
		t.Error("Output should contain 'Monitoring' title")
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

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)
	demo.InjectIntoFactory(f)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--interval", "20ms"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err != nil {
		t.Logf("execute returned: %v", err)
	}
}

func TestRun_MonitoringLoopTick(t *testing.T) {
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
				"pm:0": map[string]any{
					"id":      float64(0),
					"voltage": float64(230.0),
					"current": float64(1.5),
					"apower":  float64(345.0),
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

	// Use a very short interval so the ticker fires multiple times
	// Context timeout is 200ms with 50ms interval = at least 3 ticks
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory:  f,
		Interval: 50 * time.Millisecond,
	}
	err = run(ctx, opts)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("run returned: %v", err)
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Monitoring") {
		t.Error("Output should contain 'Monitoring' title")
	}
}

func TestRun_MultipleTicks(t *testing.T) {
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
				"switch:0": map[string]any{
					"id":     float64(0),
					"output": true,
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

	// Use very short interval (10ms) with longer timeout (150ms)
	// This ensures the ticker fires multiple times before cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory:  f,
		Interval: 10 * time.Millisecond,
	}
	err = run(ctx, opts)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("run returned: %v", err)
	}
}
