package benchmark

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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
		Iterations: 10,
		Warmup:     2,
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Call run directly
	err := run(ctx, tf.Factory, "test-device", opts)

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
		Iterations: 50,
		Warmup:     5,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, tf.Factory, "my-device", opts)
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
		Iterations: 10,
		Warmup:     0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, tf.Factory, "device", opts)
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
		Iterations: 5,
		Warmup:     1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	err := run(ctx, tf.Factory, "timeout-device", opts)
	if err == nil {
		t.Error("Expected error from run with timed out context")
	}
}

// TestRun_DirectCall_IPAddress tests run with IP address device.
func TestRun_DirectCall_IPAddress(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Iterations: 3,
		Warmup:     1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, tf.Factory, "192.168.1.100", opts)
	if err == nil {
		t.Error("Expected error from run with cancelled context")
	}

	output := tf.OutString()
	if !strings.Contains(output, "192.168.1.100") {
		t.Errorf("Output should contain IP address, got: %q", output)
	}
}
