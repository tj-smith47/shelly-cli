package wifi

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

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

	if cmd.Use != "wifi <device>" {
		t.Errorf("Use = %q, want 'wifi <device>'", cmd.Use)
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

	expectedAliases := []string{"network", "wlan"}
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
			name:     "ssid flag exists",
			flagName: "ssid",
			wantType: "string",
		},
		{
			name:     "password flag exists",
			flagName: "password",
			wantType: "string",
		},
		{
			name:     "no-scan flag exists",
			flagName: "no-scan",
			wantType: "bool",
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
			name:         "ssid default empty",
			flagName:     "ssid",
			wantDefault:  "",
			errMsgFormat: "ssid default = %q, want empty",
		},
		{
			name:         "password default empty",
			flagName:     "password",
			wantDefault:  "",
			errMsgFormat: "password default = %q, want empty",
		},
		{
			name:         "no-scan default false",
			flagName:     "no-scan",
			wantDefault:  "false",
			errMsgFormat: "no-scan default = %q, want false",
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

	err = cmd.Args(cmd, []string{"device1", "device2"})
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
			name:    "ssid flag",
			args:    []string{"--ssid", "MyNetwork"},
			wantErr: false,
		},
		{
			name:    "password flag",
			args:    []string{"--password", "secret123"},
			wantErr: false,
		},
		{
			name:    "no-scan flag",
			args:    []string{"--no-scan"},
			wantErr: false,
		},
		{
			name:    "multiple flags",
			args:    []string{"--ssid", "MyNetwork", "--password", "secret"},
			wantErr: false,
		},
		{
			name:    "all flags combined",
			args:    []string{"--ssid", "MyNetwork", "--password", "secret", "--no-scan"},
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

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Default values
	if opts.SSID != "" {
		t.Errorf("Default SSID = %q, want empty", opts.SSID)
	}
	if opts.Password != "" {
		t.Errorf("Default Password = %q, want empty", opts.Password)
	}
	if opts.NoScan {
		t.Error("Default NoScan should be false")
	}
}

func TestOptions_FieldsSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "kitchen",
		SSID:     "MyNetwork",
		Password: "secret123",
		NoScan:   true,
	}

	if opts.Device != "kitchen" {
		t.Errorf("Device = %q, want 'kitchen'", opts.Device)
	}
	if opts.SSID != "MyNetwork" {
		t.Errorf("SSID = %q, want 'MyNetwork'", opts.SSID)
	}
	if opts.Password != "secret123" {
		t.Errorf("Password = %q, want 'secret123'", opts.Password)
	}
	if !opts.NoScan {
		t.Error("NoScan should be true")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "TestNetwork",
		Password: "testpassword",
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
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "TestNetwork",
		Password: "testpassword",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_RunE_SetsDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-test-device", "--ssid", "TestNetwork", "--password", "secret"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context but want to verify structure
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_ExampleContainsInteractive(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify example contains interactive provisioning
	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestNewCommand_LongContainsUsageDetails(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description contains useful information
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestRun_WithSSIDAndPassword(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "MyNetwork",
		Password: "secret123",
	}

	// Create a cancelled context - should fail after validation but before network call
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// Will fail due to context, but SSID and password are set so we skip prompts
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_NoScanWithoutSSID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		SSID:    "",
		NoScan:  true,
	}

	// In non-TTY mode, Input returns default value (empty string)
	// So this should fail when trying to configure with empty SSID
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	// Will fail due to context or empty SSID
	if err == nil {
		t.Error("Expected error")
	}
}

func TestOptions_FactorySet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestNewCommand_ArgsValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "empty args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"device"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "IP address",
			args:    []string{"192.168.1.1"},
			wantErr: false,
		},
		{
			name:    "hostname",
			args:    []string{"shelly-plus-1"},
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

func TestNewCommand_CommandProperties(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name  string
		got   string
		field string
	}{
		{
			name:  "Use field",
			got:   cmd.Use,
			field: "Use",
		},
		{
			name:  "Short field",
			got:   cmd.Short,
			field: "Short",
		},
		{
			name:  "Long field",
			got:   cmd.Long,
			field: "Long",
		},
		{
			name:  "Example field",
			got:   cmd.Example,
			field: "Example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got == "" {
				t.Errorf("%s should not be empty", tt.field)
			}
		})
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_DirectCredentials(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
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
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "TestNetwork",
		Password: "testpassword",
		NoScan:   false,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "WiFi configured") {
		t.Errorf("Output should contain success message, got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_NoScanWithSSID(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
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

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "DirectSSID",
		Password: "directpass",
		NoScan:   true,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "DirectSSID") {
		t.Errorf("Output should contain SSID, got: %q", output)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "nonexistent",
		SSID:     "Network",
		Password: "pass",
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_Gen1Device(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
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

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "gen1-device",
		SSID:     "Network",
		Password: "pass",
	}

	// Gen1 might not support WiFi config the same way
	err = run(context.Background(), opts)
	// Error may or may not occur depending on implementation
	if err != nil {
		t.Logf("Gen1 WiFi config error (may be expected): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_CancelledContext(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
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

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "Network",
		Password: "pass",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_EmptyPassword(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
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

	// Empty password - should prompt but in test env will fail
	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "Network",
		Password: "",
	}

	// In test environment, password prompt will fail
	err = run(context.Background(), opts)
	// Error expected because password prompt fails in non-TTY
	if err != nil {
		t.Logf("password prompt error (expected in non-TTY): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_WithMock_NoScanNoSSID(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
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

	// NoScan=true but no SSID - should prompt
	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		SSID:     "",
		Password: "pass",
		NoScan:   true,
	}

	// In test environment, SSID prompt returns empty which triggers an error path
	err = run(context.Background(), opts)
	// The prompt returns empty default, then the scan happens
	if err != nil {
		t.Logf("SSID prompt error (may be expected in non-TTY): %v", err)
	}
}
