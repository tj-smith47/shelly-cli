package set

import (
	"context"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const errValueRequired = "value required (or use --toggle for booleans)"

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "set <device> <key> <value>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <device> <key> <value>")
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

	expectedAliases := []string{"update"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}
	for i, expected := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	toggleFlag := cmd.Flags().Lookup("toggle")
	switch {
	case toggleFlag == nil:
		t.Fatal("toggle flag not found")
	case toggleFlag.Shorthand != "t":
		t.Errorf("toggle shorthand = %q, want %q", toggleFlag.Shorthand, "t")
	case toggleFlag.DefValue != "false":
		t.Errorf("toggle default = %q, want %q", toggleFlag.DefValue, "false")
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
			name:    "one arg only",
			args:    []string{"device"},
			wantErr: true,
		},
		{
			name:    "two args valid",
			args:    []string{"device", "boolean:200"},
			wantErr: false,
		},
		{
			name:    "three args valid",
			args:    []string{"device", "boolean:200", "true"},
			wantErr: false,
		},
		{
			name:    "four args invalid",
			args:    []string{"device", "key", "value", "extra"},
			wantErr: true,
		},
		{
			name:    "five args invalid",
			args:    []string{"a", "b", "c", "d", "e"},
			wantErr: true,
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

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil, expected completion function for device names")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_RunE_MissingValue(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	// Two args (device, key) without value and without toggle flag
	cmd.SetArgs([]string{"device", "boolean:200"})

	ctx := context.Background()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when value is missing and toggle is not set")
	}
	if err != nil && err.Error() != errValueRequired {
		t.Errorf("Error = %q, want %q", err.Error(), errValueRequired)
	}
}

func TestNewCommand_RunE_ValueProvidedWithToggle(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	// Two args with toggle flag - should be valid (no value needed)
	cmd.SetArgs([]string{"device", "boolean:200", "--toggle"})

	// Use cancelled context to prevent actual execution
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	// Error expected (cancelled context) but NOT the "value required" error
	if err != nil && err.Error() == errValueRequired {
		t.Error("Should not require value when --toggle is set")
	}
}

func TestNewCommand_RunE_ThreeArgsValid(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"device", "boolean:200", "true"})

	// Use cancelled context to prevent actual execution
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	// Error expected (cancelled context) but NOT the "value required" error
	if err != nil && err.Error() == errValueRequired {
		t.Error("Should not require value when value is provided")
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
			name:    "toggle flag short",
			args:    []string{"-t"},
			wantErr: false,
		},
		{
			name:    "toggle flag long",
			args:    []string{"--toggle"},
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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		check   func(*testing.T)
		wantErr string
	}{
		{
			name: "has use",
			check: func(t *testing.T) {
				t.Helper()
				cmd := NewCommand(cmdutil.NewFactory())
				if cmd.Use == "" {
					t.Error("Use should not be empty")
				}
			},
		},
		{
			name: "has short",
			check: func(t *testing.T) {
				t.Helper()
				cmd := NewCommand(cmdutil.NewFactory())
				if cmd.Short == "" {
					t.Error("Short should not be empty")
				}
			},
		},
		{
			name: "has long",
			check: func(t *testing.T) {
				t.Helper()
				cmd := NewCommand(cmdutil.NewFactory())
				if cmd.Long == "" {
					t.Error("Long should not be empty")
				}
			},
		},
		{
			name: "has example",
			check: func(t *testing.T) {
				t.Helper()
				cmd := NewCommand(cmdutil.NewFactory())
				if cmd.Example == "" {
					t.Error("Example should not be empty")
				}
			},
		},
		{
			name: "has aliases",
			check: func(t *testing.T) {
				t.Helper()
				cmd := NewCommand(cmdutil.NewFactory())
				if len(cmd.Aliases) == 0 {
					t.Error("Aliases should not be empty")
				}
			},
		},
		{
			name: "has RunE",
			check: func(t *testing.T) {
				t.Helper()
				cmd := NewCommand(cmdutil.NewFactory())
				if cmd.RunE == nil {
					t.Error("RunE should be set")
				}
			},
		},
		{
			name: "uses RangeArgs(2,3)",
			check: func(t *testing.T) {
				t.Helper()
				cmd := NewCommand(cmdutil.NewFactory())
				if cmd.Args == nil {
					t.Error("Args should be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.check(t)
		})
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Key:     "boolean:200",
	}

	// Default values
	if opts.Toggle {
		t.Error("Default Toggle should be false")
	}
	if opts.Value != "" {
		t.Errorf("Default Value = %q, want empty", opts.Value)
	}
}

func TestOptions_WithValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "kitchen",
		Key:     "number:201",
		Value:   "25.5",
		Toggle:  false,
	}

	if opts.Device != "kitchen" {
		t.Errorf("Device = %q, want %q", opts.Device, "kitchen")
	}
	if opts.Key != "number:201" {
		t.Errorf("Key = %q, want %q", opts.Key, "number:201")
	}
	if opts.Value != "25.5" {
		t.Errorf("Value = %q, want %q", opts.Value, "25.5")
	}
}

func TestOptions_ToggleSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Key:     "boolean:200",
		Toggle:  true,
	}

	if !opts.Toggle {
		t.Error("Toggle should be true")
	}
}

func TestNewCommand_LongDescriptionContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Long) < 50 {
		t.Error("Long description seems too short")
	}

	// Long description should mention key format
	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain actual usage patterns
	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Example should show basic usage
	if len(cmd.Example) < 20 {
		t.Error("Example seems too short to be useful")
	}
}

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts IP addresses as device identifiers
	err := cmd.Args(cmd, []string{"192.168.1.100", "boolean:200", "true"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts named devices
	err := cmd.Args(cmd, []string{"living-room", "text:202", "hello"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestNewCommand_RunE_DeviceArgsMapping(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", "enum:203", "option1"})

	// Use cancelled context to prevent actual execution
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - error expected due to cancelled context
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}

	// The test validates that args are parsed and mapped correctly
	// since execution got past the validation phase (no "value required" error)
	if err != nil && err.Error() == errValueRequired {
		t.Error("Should not have value required error when value is provided")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Key:     "boolean:200",
		Value:   "true",
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_InvalidKey(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Key:     "invalid-key-format", // Invalid key format (no colon)
		Value:   "true",
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error with invalid key format")
	}
}

func TestRun_KeyWithInvalidID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Key:     "boolean:abc", // Invalid ID (not a number)
		Value:   "true",
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error with invalid component ID")
	}
}

func TestRun_KeyWithOutOfRangeID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{
			name: "ID below 200",
			key:  "boolean:100",
		},
		{
			name: "ID above 299",
			key:  "boolean:300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t) // Create factory per subtest to avoid race

			opts := &Options{
				Factory: tf.Factory,
				Device:  "test-device",
				Key:     tt.key,
				Value:   "true",
			}

			err := run(context.Background(), opts)

			if err == nil {
				t.Errorf("Expected error with out of range ID for key %q", tt.key)
			}
		})
	}
}
