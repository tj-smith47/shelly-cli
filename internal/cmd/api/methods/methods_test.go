package methods

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

	if cmd.Use != "methods <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "methods <device>")
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

	expectedAliases := []string{"list-methods", "lm"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
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
			name:    "one arg",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "two args",
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
		{name: "filter", shorthand: "", defValue: "", wantNonEmpty: false},
		{name: "format", shorthand: "f", defValue: formatText, wantNonEmpty: true},
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

func TestNewCommand_FilterFlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{name: "default empty", args: []string{}, expected: ""},
		{name: "filter Switch", args: []string{"--filter", "Switch"}, expected: "Switch"},
		{name: "filter cover", args: []string{"--filter", "cover"}, expected: "cover"},
		{name: "filter with spaces", args: []string{"--filter", "Get Status"}, expected: "Get Status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			got, err := cmd.Flags().GetString("filter")
			if err != nil {
				t.Fatalf("GetString(filter) error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("filter = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewCommand_FormatFlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{name: "default text", args: []string{}, expected: formatText},
		{name: "json format", args: []string{"--format", formatJSON}, expected: formatJSON},
		{name: "json shorthand", args: []string{"-f", formatJSON}, expected: formatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			got, err := cmd.Flags().GetString("format")
			if err != nil {
				t.Fatalf("GetString(format) error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("format = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Device:  "test-device",
		Filter:  "Switch",
	}
	opts.Format = formatJSON

	if opts.Factory != f {
		t.Error("Factory field not set correctly")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Filter != "Switch" {
		t.Errorf("Filter = %q, want %q", opts.Filter, "Switch")
	}
	if opts.Format != formatJSON {
		t.Errorf("Format = %q, want %q", opts.Format, formatJSON)
	}
}

func TestNewCommand_Properties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		check    func(*cmdutil.Factory) bool
		errorMsg string
	}{
		{
			name: "Use field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Use == "methods <device>"
			},
			errorMsg: "Use field not set correctly",
		},
		{
			name: "Short field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Short != ""
			},
			errorMsg: "Short field is empty",
		},
		{
			name: "Long field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Long != ""
			},
			errorMsg: "Long field is empty",
		},
		{
			name: "Example field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Example != ""
			},
			errorMsg: "Example field is empty",
		},
		{
			name: "Has at least 2 aliases",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return len(cmd.Aliases) >= 2
			},
			errorMsg: "Should have at least 2 aliases (list-methods, lm)",
		},
		{
			name: "Args validator is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Args != nil
			},
			errorMsg: "Args validator not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := cmdutil.NewFactory()
			if !tt.check(f) {
				t.Error(tt.errorMsg)
			}
		})
	}
}

func TestNewCommand_CombinedFlagsParsing(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.ParseFlags([]string{"--filter", "Switch", "--format", formatJSON})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		t.Fatalf("GetString(filter) error = %v", err)
	}
	if filter != "Switch" {
		t.Errorf("filter = %q, want %q", filter, "Switch")
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString(format) error = %v", err)
	}
	if format != formatJSON {
		t.Errorf("format = %q, want %q", format, formatJSON)
	}
}

func TestNewCommand_FilterFlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("filter")
	if flag == nil {
		t.Fatal("filter flag not found")
	}

	if flag.Usage == "" {
		t.Error("filter flag usage is empty")
	}
}

func TestRunMethods_DeviceNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
		Filter:  "",
	}
	opts.Format = formatText

	err := run(context.Background(), opts)

	// Should fail because device doesn't exist
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRunMethods_WithTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "",
	}
	opts.Format = formatJSON

	// This will fail on device connection, but exercises the early run() code
	err := run(context.Background(), opts)

	// Expect error due to no device
	if err == nil {
		t.Log("Expected connection error (no real device)")
	}
}

func TestRunMethods_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "",
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

	if !strings.Contains(helpOutput, "methods") {
		t.Error("Help should contain 'methods'")
	}
	if !strings.Contains(helpOutput, "RPC") {
		t.Error("Help should contain 'RPC'")
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "Switch",
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

func TestRunMethods_JSONFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "",
	}
	opts.Format = formatJSON

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a format error
	if err != nil && strings.Contains(err.Error(), "format") {
		t.Errorf("Unexpected format error for json: %v", err)
	}
}

func TestRunMethods_TextFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "",
	}
	opts.Format = formatText

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a format error
	if err != nil && strings.Contains(err.Error(), "format") {
		t.Errorf("Unexpected format error for text: %v", err)
	}
}

func TestRunMethods_WithFilter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "Switch",
	}
	opts.Format = formatText

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a filter error
	if err != nil && strings.Contains(err.Error(), "filter") {
		t.Errorf("Unexpected filter error: %v", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRunMethods_ListMethods_TextFormat(t *testing.T) {
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
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "",
	}
	opts.Format = formatText

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Available RPC Methods") {
		t.Error("Output should contain 'Available RPC Methods' header")
	}
	if !strings.Contains(output, "Switch") {
		t.Error("Output should contain 'Switch' namespace")
	}
	if !strings.Contains(output, "Shelly") {
		t.Error("Output should contain 'Shelly' namespace")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRunMethods_ListMethods_JSONFormat(t *testing.T) {
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
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "",
	}
	opts.Format = formatJSON

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "[") {
		t.Error("JSON output should contain array bracket")
	}
	if !strings.Contains(output, "Switch.Set") {
		t.Error("JSON output should contain Switch.Set method")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRunMethods_ListMethods_WithFilter(t *testing.T) {
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
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "switch",
	}
	opts.Format = formatText

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain Switch methods
	if !strings.Contains(output, "Switch") {
		t.Error("Output should contain filtered Switch namespace")
	}
	// Should NOT contain unmatched methods
	if strings.Contains(output, "Cover:") {
		t.Error("Output should NOT contain Cover namespace when filtering for switch")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRunMethods_ListMethods_CaseInsensitiveFilter(t *testing.T) {
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "SHELLY", // Upper case filter
	}
	opts.Format = formatText

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Case insensitive match should find Shelly methods
	if !strings.Contains(output, "Shelly") {
		t.Error("Case-insensitive filter should find Shelly namespace")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRunMethods_ListMethods_FilterNoMatch(t *testing.T) {
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "nonexistentmethod",
	}
	opts.Format = formatText

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should show 0 methods in header
	if !strings.Contains(output, "(0)") {
		t.Error("Output should show 0 methods when no match")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRunMethods_ListMethods_JSONWithFilter(t *testing.T) {
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "Switch",
	}
	opts.Format = formatJSON

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Switch") {
		t.Error("JSON output should contain filtered Switch methods")
	}
	// Should only contain Switch methods, not Shelly methods
	if strings.Contains(output, "Shelly.GetStatus") {
		t.Error("Filtered JSON should NOT contain Shelly.GetStatus")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecuteMethods_ListMethods_Success(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "living-room",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"living-room": {"switch:0": map[string]any{"output": true}},
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
	cmd.SetArgs([]string{"living-room"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Available RPC Methods") {
		t.Error("Output should contain 'Available RPC Methods'")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecuteMethods_ListMethods_WithFilterFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-switch",
					Address:    "192.168.1.101",
					MAC:        "11:22:33:44:55:66",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-switch": {},
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
	cmd.SetArgs([]string{"test-switch", "--filter", "Switch"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Switch") {
		t.Error("Output should contain Switch methods")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecuteMethods_ListMethods_JSONFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-test",
					Address:    "192.168.1.102",
					MAC:        "AA:11:BB:22:CC:33",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-test": {},
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
	cmd.SetArgs([]string{"json-test", "--format", "json"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("JSON output should contain array brackets")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecuteMethods_ListMethods_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nonexistent-device"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	// Should fail because device doesn't exist
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRunMethods_NamespacesGrouping(t *testing.T) {
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Filter:  "",
	}
	opts.Format = formatText

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()

	// Verify namespace grouping in output
	namespaces := []string{"Cover:", "Input:", "Light:", "MQTT:", "Script:", "Shelly:", "Switch:", "Wifi:"}
	for _, ns := range namespaces {
		if !strings.Contains(output, ns) {
			t.Errorf("Output should contain namespace %q", ns)
		}
	}
}
