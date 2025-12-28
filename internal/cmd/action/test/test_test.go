package test

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

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "test <device> <event>" {
		t.Errorf("Use = %q, want 'test <device> <event>'", cmd.Use)
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

	wantAliases := []string{"trigger", "fire"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Fatalf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}
	for i, alias := range wantAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			errMsg:    "Use should not be empty",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			errMsg:    "RunE should be set",
		},
		{
			name:      "has Args",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			errMsg:    "Args should be set",
		},
		{
			name:      "has ValidArgsFunction",
			checkFunc: func(c *cobra.Command) bool { return c.ValidArgsFunction != nil },
			errMsg:    "ValidArgsFunction should be set for device completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if !tt.checkFunc(cmd) {
				t.Error(tt.errMsg)
			}
		})
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
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"device"},
			wantErr: true,
		},
		{
			name:    "two args",
			args:    []string{"device", "event"},
			wantErr: false,
		},
		{
			name:    "three args",
			args:    []string{"device", "event", "extra"},
			wantErr: true,
		},
		{
			name:    "IP address and event",
			args:    []string{"192.168.1.100", "out_on_url"},
			wantErr: false,
		},
		{
			name:    "device name and event",
			args:    []string{"living-room", "out_off_url"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	indexFlag := cmd.Flags().Lookup("index")
	if indexFlag == nil {
		t.Fatal("index flag not found")
	}
	if indexFlag.DefValue != "0" {
		t.Errorf("index default = %q, want 0", indexFlag.DefValue)
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	indexFlag := cmd.Flags().Lookup("index")
	if indexFlag.DefValue != "0" {
		t.Errorf("index default = %q, want 0", indexFlag.DefValue)
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
			name:    "index flag long",
			args:    []string{"--index", "1"},
			wantErr: false,
		},
		{
			name:    "index flag equals",
			args:    []string{"--index=2"},
			wantErr: false,
		},
		{
			name:    "index flag zero",
			args:    []string{"--index", "0"},
			wantErr: false,
		},
		{
			name:    "index flag with device args",
			args:    []string{"--index", "1", "device", "event"},
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

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Event:   "out_on_url",
	}

	if opts.Index != 0 {
		t.Errorf("Default Index = %d, want 0", opts.Index)
	}
}

func TestOptions_DeviceFieldSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-device",
		Event:   "out_on_url",
	}

	if opts.Device != "my-device" {
		t.Errorf("Device = %q, want 'my-device'", opts.Device)
	}
}

func TestOptions_EventFieldSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Event:   "out_off_url",
	}

	if opts.Event != "out_off_url" {
		t.Errorf("Event = %q, want 'out_off_url'", opts.Event)
	}
}

func TestOptions_IndexFieldSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Event:   "out_on_url",
		Index:   2,
	}

	if opts.Index != 2 {
		t.Errorf("Index = %d, want 2", opts.Index)
	}
}

func TestRun_Cancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Event:   "out_on_url",
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow the timeout to trigger
	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Event:   "out_on_url",
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestNewCommand_Execute(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-test-device", "out_on_url"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_ExecuteWithoutDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	// Should fail because no device argument provided
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error when no device argument provided")
	}
}

func TestNewCommand_ExecuteWithoutEvent(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"device"})

	// Should fail because no event argument provided
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error when no event argument provided")
	}
}

func TestNewCommand_ExecuteWithIndex(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"device", "out_on_url", "--index", "1"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	if len(cmd.Long) < 50 {
		t.Error("Long description seems too short")
	}
}

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	if len(cmd.Example) < 10 {
		t.Error("Example seems too short")
	}
}

func TestOptions_AllFieldsSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Event:   "out_on_url",
		Index:   1,
	}

	if opts.Factory == nil {
		t.Error("Factory should be set")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want 'test-device'", opts.Device)
	}
	if opts.Event != "out_on_url" {
		t.Errorf("Event = %q, want 'out_on_url'", opts.Event)
	}
	if opts.Index != 1 {
		t.Errorf("Index = %d, want 1", opts.Index)
	}
}

func TestRun_WithDifferentEvents(t *testing.T) {
	t.Parallel()

	events := []string{
		"out_on_url",
		"out_off_url",
		"btn_on_url",
		"btn_off_url",
		"longpush_url",
		"shortpush_url",
	}

	for _, event := range events {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			opts := &Options{
				Factory: tf.Factory,
				Device:  "test-device",
				Event:   event,
			}

			// All should error due to cancelled context
			err := run(ctx, opts)
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}

func TestRun_WithDifferentIndexes(t *testing.T) {
	t.Parallel()

	indexes := []int{0, 1, 2, 3}

	for _, index := range indexes {
		t.Run("index_"+string(rune('0'+index)), func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			opts := &Options{
				Factory: tf.Factory,
				Device:  "test-device",
				Event:   "out_on_url",
				Index:   index,
			}

			// All should error due to cancelled context
			err := run(ctx, opts)
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Event:   "out_on_url",
	}

	// Verify Factory methods are accessible
	if opts.Factory.IOStreams() == nil {
		t.Error("IOStreams should be accessible from factory")
	}

	if opts.Factory.ShellyService() == nil {
		t.Error("ShellyService should be accessible from factory")
	}
}

func TestNewCommand_MultipleInstances(t *testing.T) {
	t.Parallel()

	// Verify multiple instances don't share state
	cmd1 := NewCommand(cmdutil.NewFactory())
	cmd2 := NewCommand(cmdutil.NewFactory())

	if cmd1 == cmd2 {
		t.Error("NewCommand should create distinct instances")
	}

	// Verify they have the same structure
	if cmd1.Use != cmd2.Use {
		t.Error("Both commands should have the same Use")
	}
}

func TestNewCommand_InheritedFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command doesn't inherit unexpected persistent flags
	if cmd.PersistentFlags().NFlag() != 0 {
		t.Errorf("Expected no persistent flags, got %d", cmd.PersistentFlags().NFlag())
	}
}

func TestNewCommand_LocalFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify local flags
	localFlags := cmd.LocalFlags()

	indexFlag := localFlags.Lookup("index")
	if indexFlag == nil {
		t.Error("index should be a local flag")
	}
}

func TestRun_ContextPropagation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a context with a deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "192.168.1.100",
		Event:   "out_on_url",
	}

	// Run should eventually fail due to context deadline or device unavailability
	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestNewCommand_PositionalArgsHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "valid device and event",
			args:       []string{"device1", "out_on_url"},
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "missing all args",
			args:       []string{},
			wantErr:    true,
			errMessage: "accepts 2 arg(s), received 0",
		},
		{
			name:       "only device provided",
			args:       []string{"device1"},
			wantErr:    true,
			errMessage: "accepts 2 arg(s), received 1",
		},
		{
			name:       "too many args",
			args:       []string{"device1", "event1", "extra"},
			wantErr:    true,
			errMessage: "accepts 2 arg(s), received 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_WithInvalidIP(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Use an unreachable IP to test error handling
	opts := &Options{
		Factory: tf.Factory,
		Device:  "0.0.0.0", // Unreachable IP
		Event:   "out_on_url",
	}

	// Create a short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Should fail trying to connect
	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with invalid IP")
	}
}

func TestRun_WithLocalhost(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Use localhost which won't have a Shelly device
	opts := &Options{
		Factory: tf.Factory,
		Device:  "127.0.0.1",
		Event:   "out_off_url",
		Index:   0,
	}

	// Create a short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Should fail trying to connect
	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with localhost (no Shelly device)")
	}
}

func TestOptions_ZeroValue(t *testing.T) {
	t.Parallel()

	// Test zero-value Options struct
	var opts Options

	if opts.Device != "" {
		t.Error("Zero Device should be empty string")
	}
	if opts.Event != "" {
		t.Error("Zero Event should be empty string")
	}
	if opts.Index != 0 {
		t.Error("Zero Index should be 0")
	}
	if opts.Factory != nil {
		t.Error("Zero Factory should be nil")
	}
}
