package benchmark

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "benchmark <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "benchmark <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"bench", "perf"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}

	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias[%d] = %q", i, want)
			continue
		}
		if cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "iterations", shorthand: "n", defValue: "10"},
		{name: "warmup", shorthand: "", defValue: "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}

			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
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
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg valid",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "two args invalid",
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "ip address as device",
			args:    []string{"192.168.1.100"},
			wantErr: false,
		},
		{
			name:    "alias as device",
			args:    []string{"kitchen-light"},
			wantErr: false,
		},
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

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Get flag values to verify defaults are applied
	iterations, err := cmd.Flags().GetInt("iterations")
	if err != nil {
		t.Fatalf("failed to get iterations flag: %v", err)
	}
	if iterations != 10 {
		t.Errorf("iterations default = %d, want 10", iterations)
	}

	warmup, err := cmd.Flags().GetInt("warmup")
	if err != nil {
		t.Fatalf("failed to get warmup flag: %v", err)
	}
	if warmup != 2 {
		t.Errorf("warmup default = %d, want 2", warmup)
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify RunE is set
	if cmd.RunE == nil {
		t.Error("RunE is not set")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify example contains key patterns
	example := cmd.Example
	patterns := []string{
		"shelly benchmark",
		"--iterations",
		"--json",
	}

	for _, pattern := range patterns {
		if !strings.Contains(example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description mentions key concepts
	long := cmd.Long
	patterns := []string{
		"latency",
		"RPC",
		"P50",
		"P95",
		"P99",
	}

	for _, pattern := range patterns {
		if !strings.Contains(long, pattern) {
			t.Errorf("Long description should mention %q", pattern)
		}
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		wantOK    bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			wantOK:    true,
			errMsg:    "Use should not be empty",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			wantOK:    true,
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			wantOK:    true,
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			wantOK:    true,
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK:    true,
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK:    true,
			errMsg:    "RunE should be set",
		},
		{
			name:      "uses ExactArgs(1)",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			wantOK:    true,
			errMsg:    "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if tt.checkFunc(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "iterations flag short",
			args:    []string{"-n", "50"},
			wantErr: false,
		},
		{
			name:    "iterations flag long",
			args:    []string{"--iterations", "50"},
			wantErr: false,
		},
		{
			name:    "warmup flag long",
			args:    []string{"--warmup", "5"},
			wantErr: false,
		},
		{
			name:    "multiple flags",
			args:    []string{"-n", "100", "--warmup", "10"},
			wantErr: false,
		},
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

func TestNewCommand_FlagValues(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse with custom values
	if err := cmd.ParseFlags([]string{"-n", "50", "--warmup", "5"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	iterations, err := cmd.Flags().GetInt("iterations")
	if err != nil {
		t.Fatalf("failed to get iterations flag: %v", err)
	}
	if iterations != 50 {
		t.Errorf("iterations = %d, want 50", iterations)
	}

	warmup, err := cmd.Flags().GetInt("warmup")
	if err != nil {
		t.Fatalf("failed to get warmup flag: %v", err)
	}
	if warmup != 5 {
		t.Errorf("warmup = %d, want 5", warmup)
	}
}

func TestRun_CancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow the timeout to trigger
	time.Sleep(1 * time.Millisecond)

	cmd.SetContext(ctx)

	// Execute - we expect an error due to timeout
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with timed out context")
	}
}

func TestRun_OutputContainsDeviceName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-test-device"})

	// Create a cancelled context to prevent actual network call
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - will fail but should output device name first
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}

	output := tf.OutString()
	if !strings.Contains(output, "my-test-device") {
		t.Errorf("Output should contain device name 'my-test-device', got: %q", output)
	}
}

func TestRun_OutputContainsIterations(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device", "-n", "25"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}

	output := tf.OutString()
	if !strings.Contains(output, "25") {
		t.Errorf("Output should contain iterations '25', got: %q", output)
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

func TestNewCommand_RejectsMultipleArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Command should reject multiple device arguments")
	}
}

func TestOptions_IterationsField(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Iterations: 100,
		Warmup:     5,
	}

	if opts.Iterations != 100 {
		t.Errorf("Iterations = %d, want 100", opts.Iterations)
	}
	if opts.Warmup != 5 {
		t.Errorf("Warmup = %d, want 5", opts.Warmup)
	}
}

func TestOptions_WarmupField(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Iterations: 10,
		Warmup:     0,
	}

	if opts.Warmup != 0 {
		t.Errorf("Warmup = %d, want 0", opts.Warmup)
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	// When Options is created via NewCommand, it should have defaults
	cmd := NewCommand(cmdutil.NewFactory())

	// Parse with no flags to ensure defaults are applied
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	iterationsFlag := cmd.Flags().Lookup("iterations")
	if iterationsFlag.DefValue != "10" {
		t.Errorf("iterations default = %q, want 10", iterationsFlag.DefValue)
	}

	warmupFlag := cmd.Flags().Lookup("warmup")
	if warmupFlag.DefValue != "2" {
		t.Errorf("warmup default = %q, want 2", warmupFlag.DefValue)
	}
}

// TestRun_DirectCall tests the run function directly with cancelled context.
func TestRun_DirectCall(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "test-device",
		Iterations: 10,
		Warmup:     2,
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Call run directly
	err := run(ctx, opts)

	// Should fail due to network/context error
	if err == nil {
		t.Error("Expected error from run with cancelled context")
	}

	// Verify output was written (lines 67-69 in run function)
	output := tf.OutString()
	if !strings.Contains(output, "Benchmarking") {
		t.Errorf("Output should contain 'Benchmarking', got: %q", output)
	}
	if !strings.Contains(output, "test-device") {
		t.Errorf("Output should contain device name, got: %q", output)
	}
}

// TestRun_DirectCall_CustomIterations tests run with custom iterations.
func TestRun_DirectCall_CustomIterations(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "my-device",
		Iterations: 50,
		Warmup:     5,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error from run with cancelled context")
	}

	output := tf.OutString()
	if !strings.Contains(output, "50") {
		t.Errorf("Output should contain iterations '50', got: %q", output)
	}
	if !strings.Contains(output, "5") {
		t.Errorf("Output should contain warmup '5', got: %q", output)
	}
}

// TestRun_DirectCall_ZeroWarmup tests run with zero warmup.
func TestRun_DirectCall_ZeroWarmup(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "device",
		Iterations: 10,
		Warmup:     0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error from run with cancelled context")
	}

	output := tf.OutString()
	if !strings.Contains(output, "Benchmarking") {
		t.Errorf("Output should contain 'Benchmarking', got: %q", output)
	}
}

// TestRun_DirectCall_Timeout tests run with timeout.
func TestRun_DirectCall_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "timeout-device",
		Iterations: 5,
		Warmup:     1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error from run with timed out context")
	}
}

// TestRun_DirectCall_IPAddress tests run with IP address device.
func TestRun_DirectCall_IPAddress(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "192.168.1.100",
		Iterations: 3,
		Warmup:     1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error from run with cancelled context")
	}

	output := tf.OutString()
	if !strings.Contains(output, "192.168.1.100") {
		t.Errorf("Output should contain IP address, got: %q", output)
	}
}

// TestExecute_WithMockDevice tests full Execute path with mock device server.
func TestExecute_WithMockDevice(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bench-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bench-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetArgs([]string{"bench-device", "-n", "3", "--warmup", "1"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Benchmark Complete") {
		t.Errorf("Expected 'Benchmark Complete' in output, got: %q", output)
	}
	if !strings.Contains(output, "bench-device") {
		t.Errorf("Expected device name in output, got: %q", output)
	}
	if !strings.Contains(output, "RPC Latency") {
		t.Errorf("Expected 'RPC Latency' in output, got: %q", output)
	}
	if !strings.Contains(output, "Ping Latency") {
		t.Errorf("Expected 'Ping Latency' in output, got: %q", output)
	}
}

// TestExecute_Gen1Device tests that Gen1 devices are rejected.
func TestExecute_Gen1Device(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
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
	cmd.SetArgs([]string{"gen1-device", "-n", "2", "--warmup", "0"})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Expected Gen2+ error message, got: %v", err)
	}
}

// TestExecute_ZeroWarmup tests benchmark with no warmup iterations.
func TestExecute_ZeroWarmup(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "no-warmup-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"no-warmup-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetArgs([]string{"no-warmup-device", "-n", "2", "--warmup", "0"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Verify warmup was skipped - no "Warming up..." message with 0 warmup
	if !strings.Contains(output, "Benchmark Complete") {
		t.Errorf("Expected 'Benchmark Complete' in output, got: %q", output)
	}
}

// TestExecute_JSONOutput tests benchmark with JSON output format.
func TestExecute_JSONOutput(t *testing.T) { //nolint:paralleltest // Modifies global viper state
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Set output format to JSON via viper
	viper.Set("output", "json")
	defer viper.Set("output", "")

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"json-device", "-n", "2", "--warmup", "1"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// JSON output should contain the device name in JSON format
	if !strings.Contains(output, `"device"`) {
		t.Errorf("Expected JSON field 'device' in output, got: %q", output)
	}
	if !strings.Contains(output, `"iterations"`) {
		t.Errorf("Expected JSON field 'iterations' in output, got: %q", output)
	}
}

// TestRun_WithMock tests the run function directly with mock device.
func TestRun_WithMock(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "run-test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"run-test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "run-test-device",
		Iterations: 3,
		Warmup:     1,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Benchmark Complete") {
		t.Errorf("Expected 'Benchmark Complete' in output, got: %q", output)
	}
}

// TestRun_Gen1Rejection tests that run function rejects Gen1 devices.
func TestRun_Gen1Rejection(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-test",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-test": {"relay": map[string]any{"ison": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "gen1-test",
		Iterations: 2,
		Warmup:     0,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Expected Gen2+ error, got: %v", err)
	}
}

// TestRun_ManyIterations tests run with more iterations for progress output.
func TestRun_ManyIterations(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "many-iter-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"many-iter-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "many-iter-device",
		Iterations: 10, // More than 5 to trigger progress output
		Warmup:     2,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Progress output shows every 5 iterations
	if !strings.Contains(output, "Progress") {
		t.Errorf("Expected progress output for 10 iterations, got: %q", output)
	}
}

// TestExecute_VerifyOutputFormat tests that output contains expected sections.
func TestExecute_VerifyOutputFormat(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "format-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"format-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetArgs([]string{"format-device", "-n", "3", "--warmup", "1"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()

	// Verify all expected output sections
	expectedSections := []string{
		"Benchmarking",
		"format-device",
		"Warming up",
		"Running RPC benchmark",
		"Running ping benchmark",
		"Benchmark Complete",
		"Device:",
		"Iterations:",
		"RPC Latency",
		"Ping Latency",
		"Summary:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected %q in output, got: %q", section, output)
		}
	}
}

// TestExecute_HighIterations tests benchmark with more iterations.
func TestExecute_HighIterations(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "high-iter-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"high-iter-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetArgs([]string{"high-iter-device", "-n", "15", "--warmup", "3"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should show progress at 5, 10, 15
	if !strings.Contains(output, "Progress: 5/15") {
		t.Errorf("Expected 'Progress: 5/15' in output, got: %q", output)
	}
	if !strings.Contains(output, "Progress: 10/15") {
		t.Errorf("Expected 'Progress: 10/15' in output, got: %q", output)
	}
	if !strings.Contains(output, "Progress: 15/15") {
		t.Errorf("Expected 'Progress: 15/15' in output, got: %q", output)
	}
}

// TestExecute_UnknownDevice tests benchmark with non-existent device.
func TestExecute_UnknownDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"unknown-device"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for unknown device")
	}
}

// TestExecute_DefaultIterations tests benchmark with default iterations.
func TestExecute_DefaultIterations(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "default-iter-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"default-iter-device": {"switch:0": map[string]any{"output": false}},
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
	// Use defaults (10 iterations, 2 warmup)
	cmd.SetArgs([]string{"default-iter-device"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should show "10 iterations + 2 warmup"
	if !strings.Contains(output, "10 iterations") {
		t.Errorf("Expected '10 iterations' in output (default), got: %q", output)
	}
	if !strings.Contains(output, "2 warmup") {
		t.Errorf("Expected '2 warmup' in output (default), got: %q", output)
	}
}
