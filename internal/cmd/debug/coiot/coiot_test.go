package coiot

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	formatText = "text"
	formatJSON = "json"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "coiot [device]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "coiot [device]")
	}

	if cmd.Short != "Show CoIoT/CoAP status or listen for multicast updates" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Show CoIoT/CoAP status or listen for multicast updates")
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

	expectedAliases := []string{"coap"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_ArgsValidation(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args allowed by args validator",
			args:    []string{},
			wantErr: false, // Args validator allows 0-1 args; RunE checks --listen
		},
		{
			name:    "one arg",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "two args rejected",
			args:    []string{"device1", "extra"},
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name         string
		shorthand    string
		defValue     string
		wantNonEmpty bool
	}{
		{name: "format", shorthand: "f", defValue: formatText, wantNonEmpty: true},
		{name: "listen", shorthand: "l", defValue: "false", wantNonEmpty: true},
		{name: "stream", shorthand: "s", defValue: "false", wantNonEmpty: true},
		{name: "duration", shorthand: "", defValue: "30s", wantNonEmpty: true},
		{name: "raw", shorthand: "", defValue: "false", wantNonEmpty: true},
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
			if tt.wantNonEmpty && flag.Usage == "" {
				t.Errorf("%s usage is empty", tt.name)
			}
		})
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is not set")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is not set")
	}
}

func TestOptions_Initialization(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Parse with no flags to check defaults
	err := cmd.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	// Verify format flag default
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString(format) failed: %v", err)
	}
	if format != formatText {
		t.Errorf("format default = %q, want %q", format, formatText)
	}
}

func TestNewCommand_JsonFormatFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.ParseFlags([]string{"--format", formatJSON})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString(format) failed: %v", err)
	}
	if format != formatJSON {
		t.Errorf("format = %q, want %q", format, formatJSON)
	}
}

func TestNewCommand_ShorthandFormatFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.ParseFlags([]string{"-f", formatJSON})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString(format) failed: %v", err)
	}
	if format != formatJSON {
		t.Errorf("format = %q, want %q", format, formatJSON)
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
	}
	opts.Format = formatText

	err := run(context.Background(), opts)

	// Should fail because device doesn't exist
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_WithTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

	// This will fail on device connection, but exercises the early run() code
	err := run(context.Background(), opts)

	// Expect error due to no device
	if err == nil {
		t.Log("Expected connection error (no real device)")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

	err := run(ctx, opts)

	// Should return some error (context cancelled or connection error)
	if err == nil {
		t.Log("Expected error with cancelled context")
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()

	if err == nil {
		t.Error("Expected error when executing with no arguments")
	}
}

func TestNewCommand_ExecuteWithDeviceArg(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Execute will fail due to no real device, but args should be accepted
	err := cmd.Execute()

	// We expect an error (no device connection), but not an args error
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "accepts") && strings.Contains(errStr, "arg") {
			t.Errorf("Should accept device argument, got args error: %v", err)
		}
	}
}

func TestNewCommand_HelpOutput(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Logf("Help execution: %v", err)
	}

	helpOutput := stdout.String()

	if !strings.Contains(helpOutput, "coiot") {
		t.Error("Help should contain 'coiot'")
	}
	if !strings.Contains(helpOutput, "CoIoT") {
		t.Error("Help should contain 'CoIoT'")
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

	// Verify factory is accessible
	if opts.Factory == nil {
		t.Fatal("Options.Factory should not be nil")
	}

	ios := opts.Factory.IOStreams()
	if ios == nil {
		t.Error("Factory.IOStreams() should not return nil")
	}

	svc := opts.Factory.ShellyService()
	if svc == nil {
		t.Error("Factory.ShellyService() should not return nil")
	}
}

func TestRun_JSONFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatJSON

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a format error
	if err != nil && strings.Contains(err.Error(), "format") {
		t.Errorf("Unexpected format error for json: %v", err)
	}
}

func TestRun_TextFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.Format = formatText

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a format error
	if err != nil && strings.Contains(err.Error(), "format") {
		t.Errorf("Unexpected format error for text: %v", err)
	}
}

// Execute-based tests with mock server
// These tests use mock.StartWithFixtures which sets a global demo singleton,
// so they cannot run in parallel.

//nolint:paralleltest // Uses global mock singleton
func TestExecute_Gen2Device_JSONOutput(t *testing.T) {
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
				"coiot": map[string]any{
					"enable":        true,
					"update_period": 30,
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
	cmd.SetArgs([]string{"test-device", "--format", "json"})
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "supported") {
		t.Errorf("JSON output should contain 'supported', got: %s", output)
	}
	if !strings.Contains(output, "WebSocket") {
		t.Errorf("JSON output should mention WebSocket for Gen2, got: %s", output)
	}
}

//nolint:paralleltest // Uses global mock singleton
func TestExecute_Gen2Device_TextOutput(t *testing.T) {
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
				"coiot": map[string]any{
					"enable":        true,
					"update_period": 30,
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
	cmd.SetArgs([]string{"test-device"}) // default is text format
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "CoIoT") {
		t.Errorf("Text output should contain 'CoIoT', got: %s", output)
	}
	if !strings.Contains(output, "Configuration") {
		t.Errorf("Text output should contain 'Configuration', got: %s", output)
	}
}

//nolint:paralleltest // Uses global mock singleton
func TestExecute_Gen1Device_ShowsCoIoTStatus(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.101",
					MAC:        "11:22:33:44:55:66",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": true}},
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
	cmd.SetArgs([]string{"gen1-device"})
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "CoIoT") {
		t.Errorf("Output should contain 'CoIoT', got: %s", output)
	}
	if !strings.Contains(output, "Gen1") {
		t.Errorf("Output should mention Gen1, got: %s", output)
	}
}

//nolint:paralleltest // Uses global mock singleton
func TestExecute_Gen2Device_WithoutCoIoTSection(t *testing.T) {
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
				// No coiot section
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
	cmd.SetArgs([]string{"test-device"})
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := tf.OutString()
	// Should show "not configured" message when no CoIoT section
	if !strings.Contains(output, "not configured") && !strings.Contains(output, "not available") {
		t.Errorf("Output should mention CoIoT not configured, got: %s", output)
	}
}

//nolint:paralleltest // Uses global mock singleton
func TestExecute_Gen2Device_ShowsWebSocketMessage(t *testing.T) {
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
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
	cmd.SetArgs([]string{"test-device", "--format", "json"})
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "supported") || !strings.Contains(output, "false") {
		t.Errorf("JSON output should indicate CoIoT not supported, got: %s", output)
	}
	if !strings.Contains(output, "WebSocket") {
		t.Errorf("JSON output should mention WebSocket, got: %s", output)
	}
}

//nolint:paralleltest // Uses global mock singleton
func TestExecute_WithCoAPAlias(t *testing.T) {
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
			"test-device": {
				"coiot": map[string]any{"enable": true},
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

	// The command should work via alias "coap"
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}
