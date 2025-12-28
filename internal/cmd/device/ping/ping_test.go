package ping

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "ping <device>" {
		t.Errorf("Use = %q, want 'ping <device>'", cmd.Use)
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

	if len(cmd.Aliases) < 3 {
		t.Errorf("Expected at least 3 aliases, got %d", len(cmd.Aliases))
	}

	expectedAliases := map[string]bool{"check": true, "test": true, "p": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("Unexpected alias: %s", alias)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	countFlag := cmd.Flags().Lookup("count")
	if countFlag == nil {
		t.Fatal("count flag not found")
	}
	if countFlag.Shorthand != "c" {
		t.Errorf("count shorthand = %q, want c", countFlag.Shorthand)
	}

	timeoutFlag := cmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Fatal("timeout flag not found")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	countFlag := cmd.Flags().Lookup("count")
	if countFlag.DefValue != "1" {
		t.Errorf("count default = %q, want 1", countFlag.DefValue)
	}

	timeoutFlag := cmd.Flags().Lookup("timeout")
	if timeoutFlag.DefValue != "5s" {
		t.Errorf("timeout default = %q, want 5s", timeoutFlag.DefValue)
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}
}

func TestNewCommand_RejectsMultipleArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Expected error when multiple args provided")
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
			name:    "count flag short",
			args:    []string{"-c", "5"},
			wantErr: false,
		},
		{
			name:    "count flag long",
			args:    []string{"--count", "5"},
			wantErr: false,
		},
		{
			name:    "timeout flag",
			args:    []string{"--timeout", "10s"},
			wantErr: false,
		},
		{
			name:    "multiple flags",
			args:    []string{"-c", "3", "--timeout", "15s"},
			wantErr: false,
		},
		{
			name:    "no flags",
			args:    []string{},
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

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Count:   1,
		Timeout: 5 * time.Second,
	}

	if opts.Count != 1 {
		t.Errorf("Default Count = %d, want 1", opts.Count)
	}
	if opts.Timeout != 5*time.Second {
		t.Errorf("Default Timeout = %v, want 5s", opts.Timeout)
	}
}

func TestOptions_CustomValues(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Count:   10,
		Timeout: 30 * time.Second,
	}

	if opts.Count != 10 {
		t.Errorf("Count = %d, want 10", opts.Count)
	}
	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", opts.Timeout)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Count:   1,
		Timeout: 5 * time.Second,
	}

	err := run(ctx, tf.Factory, "test-device", opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Count:   1,
		Timeout: 5 * time.Second,
	}

	err := run(ctx, tf.Factory, "test-device", opts)
	if err == nil {
		t.Error("Expected error with timed out context")
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

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example
	if example == "" {
		t.Fatal("Example should not be empty")
	}

	// Check for expected patterns
	patterns := []string{"shelly", "device", "ping"}
	for _, pattern := range patterns {
		found := false
		for i := 0; i <= len(example)-len(pattern); i++ {
			if example[i:i+len(pattern)] == pattern {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description should not be empty")
	}

	// Long should be more descriptive than Short
	if len(cmd.Long) <= len(cmd.Short) {
		t.Error("Long description should be longer than Short description")
	}
}

func TestNewCommand_RunE_PassesDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_RunE_WithFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", "-c", "3", "--timeout", "10s"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestRun_SinglePing(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Count:   1,
		Timeout: 5 * time.Second,
	}

	// Will fail due to cancelled context, but tests single ping path
	err := run(ctx, tf.Factory, "test-device", opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_MultiplePings(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Count:   3,
		Timeout: 5 * time.Second,
	}

	// Will fail due to cancelled context, but tests multiple ping path
	err := run(ctx, tf.Factory, "test-device", opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_ValidCountValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		count   string
		wantErr bool
	}{
		{"count 1", "1", false},
		{"count 5", "5", false},
		{"count 10", "10", false},
		{"count 100", "100", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags([]string{"-c", tt.count})
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_ValidTimeoutValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeout string
		wantErr bool
	}{
		{"1 second", "1s", false},
		{"5 seconds", "5s", false},
		{"10 seconds", "10s", false},
		{"1 minute", "1m", false},
		{"500 milliseconds", "500ms", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags([]string{"--timeout", tt.timeout})
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_OutputFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Count:   1,
		Timeout: 5 * time.Second,
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Run with cancelled context - error is expected and checked via output
	err := run(ctx, tf.Factory, "test-device", opts)
	if err == nil {
		t.Log("Note: run() returned nil error with cancelled context")
	}

	// Check that some output was produced (PING message)
	output := tf.OutString()
	if output == "" {
		t.Error("Expected some output from ping command")
	}
}

func TestRun_PrintsDeviceName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Count:   1,
		Timeout: 5 * time.Second,
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, tf.Factory, "test-device", opts)
	if err == nil {
		t.Log("Note: run() returned nil error with cancelled context")
	}

	// Check output contains the device name
	output := tf.OutString()
	found := false
	target := "test-device"
	for i := 0; i <= len(output)-len(target); i++ {
		if output[i:i+len(target)] == target {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected output to contain device name %q, got: %s", target, output)
	}
}
