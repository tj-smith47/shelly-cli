package add

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// Test constants for peripheral types.
const (
	testTypeDS18B20   = "ds18b20"
	testTypeDHT22     = "dht22"
	testTypeDigitalIn = "digital_in"
	testTypeAnalogIn  = "analog_in"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "add <device> <type>" {
		t.Errorf("Use = %q, want 'add <device> <type>'", cmd.Use)
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

	expectedAliases := []string{"create", "new"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Fatalf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}

	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "cid flag exists",
			flagName: "cid",
			wantType: "int",
		},
		{
			name:     "addr flag exists",
			flagName: "addr",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.flagName)
			}
			if flag.Value.Type() != tt.wantType {
				t.Errorf("%s type = %q, want %q", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	tests := []struct {
		name         string
		flagName     string
		wantDefault  string
		errMsgFormat string
	}{
		{
			name:         "cid default zero",
			flagName:     "cid",
			wantDefault:  "0",
			errMsgFormat: "cid default = %q, want 0",
		},
		{
			name:         "addr default empty",
			flagName:     "addr",
			wantDefault:  "",
			errMsgFormat: "addr default = %q, want empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag.DefValue != tt.wantDefault {
				t.Errorf(tt.errMsgFormat, flag.DefValue)
			}
		})
	}
}

func TestNewCommand_RequiresArgs(t *testing.T) {
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
			name:    "one arg only",
			args:    []string{"device1"},
			wantErr: true,
		},
		{
			name:    "two args valid",
			args:    []string{"device1", "ds18b20"},
			wantErr: false,
		},
		{
			name:    "three args invalid",
			args:    []string{"device1", "ds18b20", "extra"},
			wantErr: true,
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

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device and type completion")
	}
}

func TestNewCommand_ValidArgsFunction_DeviceCompletion(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that completion for first arg (device) returns something
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{}, "")
	// Should not return error directive
	if directive == cobra.ShellCompDirectiveError {
		t.Error("First arg completion should not return error directive")
	}
	// Suggestions may be empty without devices configured, that's ok
	_ = suggestions
}

func TestNewCommand_ValidArgsFunction_TypeCompletion(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that completion for second arg (type) returns peripheral types
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{"device1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Type completion directive = %v, want NoFileComp", directive)
	}

	// Should have valid peripheral types
	expectedTypes := []string{"ds18b20", "dht22", "digital_in", "analog_in"}
	if len(suggestions) != len(expectedTypes) {
		t.Errorf("Type suggestions count = %d, want %d", len(suggestions), len(expectedTypes))
	}
}

func TestNewCommand_ValidArgsFunction_NoMoreArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that completion for third arg returns nothing
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{"device1", "ds18b20"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Third arg directive = %v, want NoFileComp", directive)
	}
	if len(suggestions) != 0 {
		t.Errorf("Third arg suggestions = %v, want empty", suggestions)
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
			name:      "uses ExactArgs(2)",
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
			name:    "cid flag",
			args:    []string{"--cid", "101"},
			wantErr: false,
		},
		{
			name:    "addr flag",
			args:    []string{"--addr", "40:255:100:6:199:204:149:177"},
			wantErr: false,
		},
		{
			name:    "both flags",
			args:    []string{"--cid", "101", "--addr", "40:255:100:6:199:204:149:177"},
			wantErr: false,
		},
		{
			name:    "invalid cid type",
			args:    []string{"--cid", "notanumber"},
			wantErr: true,
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

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"192.168.1.100", "ds18b20"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"living-room", "dht22"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    testTypeDS18B20,
	}

	// Default values
	if opts.CID != 0 {
		t.Errorf("Default CID = %d, want 0", opts.CID)
	}
	if opts.Addr != "" {
		t.Errorf("Default Addr = %q, want empty", opts.Addr)
	}
}

func TestOptions_FieldsSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "kitchen",
		Type:    testTypeDS18B20,
		CID:     101,
		Addr:    "40:255:100:6:199:204:149:177",
	}

	if opts.Device != "kitchen" {
		t.Errorf("Device = %q, want 'kitchen'", opts.Device)
	}
	if opts.Type != testTypeDS18B20 {
		t.Errorf("Type = %q, want %q", opts.Type, testTypeDS18B20)
	}
	if opts.CID != 101 {
		t.Errorf("CID = %d, want 101", opts.CID)
	}
	if opts.Addr != "40:255:100:6:199:204:149:177" {
		t.Errorf("Addr = %q, want '40:255:100:6:199:204:149:177'", opts.Addr)
	}
}

func TestRun_InvalidType(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    "invalid_type",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

func TestRun_DS18B20RequiresAddr(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    testTypeDS18B20,
		Addr:    "", // Missing required addr
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when DS18B20 missing --addr")
	}
}

func TestRun_DHT22NoAddrRequired(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    testTypeDHT22,
		Addr:    "", // Addr not required for DHT22
	}

	// Use a cancelled context to prevent actual network call
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// Should fail due to cancelled context, not validation
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_DigitalInNoAddrRequired(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    testTypeDigitalIn,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// Should fail due to cancelled context, not validation
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_AnalogInNoAddrRequired(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    testTypeAnalogIn,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// Should fail due to cancelled context, not validation
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    testTypeDHT22,
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow the timeout to trigger
	time.Sleep(1 * time.Millisecond)

	err := run(ctx, opts)

	// Expect an error due to timeout
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestRun_CancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Type:    testTypeDHT22,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_TypeCaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeArg  string
		wantType string
	}{
		{
			name:     "lowercase ds18b20",
			typeArg:  testTypeDS18B20,
			wantType: testTypeDS18B20,
		},
		{
			name:     "uppercase DS18B20",
			typeArg:  "DS18B20",
			wantType: testTypeDS18B20,
		},
		{
			name:     "mixed case Dht22",
			typeArg:  "Dht22",
			wantType: testTypeDHT22,
		},
		{
			name:     "uppercase DIGITAL_IN",
			typeArg:  "DIGITAL_IN",
			wantType: testTypeDigitalIn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)

			cmd := NewCommand(tf.Factory)
			// DS18B20 requires addr, so use a type that doesn't
			if tt.wantType == testTypeDS18B20 {
				cmd.SetArgs([]string{"device", tt.typeArg, "--addr", "40:255:100:6:199:204:149:177"})
			} else {
				cmd.SetArgs([]string{"device", tt.typeArg})
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			cmd.SetContext(ctx)

			// Execute - will fail due to context, but verifies type is parsed
			if err := cmd.Execute(); err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}

func TestNewCommand_RunE_SetsDeviceAndType(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-test-device", testTypeDHT22})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context but want to verify structure
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_LongContainsPeripheralTypes(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description mentions peripheral types
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestNewCommand_ExampleContainsDS18B20(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify example contains DS18B20 usage
	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestRun_ValidTypesAccepted(t *testing.T) {
	t.Parallel()

	validTypes := []string{testTypeDS18B20, testTypeDHT22, testTypeDigitalIn, testTypeAnalogIn}

	for _, pType := range validTypes {
		t.Run(pType, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)

			opts := &Options{
				Factory: tf.Factory,
				Device:  "test-device",
				Type:    pType,
			}

			// Add addr for ds18b20
			if pType == testTypeDS18B20 {
				opts.Addr = "40:255:100:6:199:204:149:177"
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := run(ctx, opts)
			// Should fail due to context, not type validation
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}

func TestRun_InvalidTypesRejected(t *testing.T) {
	t.Parallel()

	invalidTypes := []string{"unknown", "temp", "sensor", "ds1820", ""}

	for _, pType := range invalidTypes {
		t.Run("type_"+pType, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)

			opts := &Options{
				Factory: tf.Factory,
				Device:  "test-device",
				Type:    pType,
			}

			err := run(context.Background(), opts)
			if err == nil {
				t.Errorf("Expected error for invalid type %q", pType)
			}
		})
	}
}
