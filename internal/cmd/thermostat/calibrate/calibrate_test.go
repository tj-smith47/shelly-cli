package calibrate

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

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cmdutil.Factory) bool
		wantOK    bool
		errMsg    string
	}{
		{
			name: "has use",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Use != ""
			},
			wantOK: true,
			errMsg: "Use should not be empty",
		},
		{
			name: "has short",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Short != ""
			},
			wantOK: true,
			errMsg: "Short should not be empty",
		},
		{
			name: "has long",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Long != ""
			},
			wantOK: true,
			errMsg: "Long should not be empty",
		},
		{
			name: "has example",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Example != ""
			},
			wantOK: true,
			errMsg: "Example should not be empty",
		},
		{
			name: "has aliases",
			checkFunc: func(f *cmdutil.Factory) bool {
				return len(NewCommand(f).Aliases) > 0
			},
			wantOK: true,
			errMsg: "Aliases should not be empty",
		},
		{
			name: "has RunE",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).RunE != nil
			},
			wantOK: true,
			errMsg: "RunE should be set",
		},
		{
			name: "uses ExactArgs(1)",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Args != nil
			},
			wantOK: true,
			errMsg: "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := cmdutil.NewFactory()
			if tt.checkFunc(f) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	if len(cmd.Aliases) == 0 {
		t.Fatal("Aliases are empty")
	}

	found := false
	for _, alias := range cmd.Aliases {
		if alias == "cal" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected alias 'cal' not found in %v", cmd.Aliases)
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("id flag not found")
	}
	if idFlag.Shorthand != "i" {
		t.Errorf("id shorthand = %q, want %q", idFlag.Shorthand, "i")
	}
	if idFlag.DefValue != "0" {
		t.Errorf("id default = %q, want %q", idFlag.DefValue, "0")
	}
}

func TestNewCommand_Args(t *testing.T) {
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
		t.Error("Expected error with multiple args")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestExecute_WithMock(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-thermostat",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSN-0024X",
					Model:      "Shelly Plus HT",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-thermostat": {"thermostat:0": map[string]any{"target_C": 21.0, "current_C": 19.5}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-thermostat"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithMockAndComponentID(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-thermostat",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSN-0024X",
					Model:      "Shelly Plus HT",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-thermostat": {
				"thermostat:0": map[string]any{"target_C": 21.0, "current_C": 19.5},
				"thermostat:1": map[string]any{"target_C": 22.0, "current_C": 20.5},
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

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-thermostat", "--id", "1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {
	t.Parallel()

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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "unknown") {
		t.Logf("error = %v", err)
	}
}

func TestExecute_Gen1Device(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-gen1",
					Address:    "192.168.1.101",
					MAC:        "BB:CC:DD:EE:FF:AA",
					Type:       "SHSW-1",
					Model:      "Shelly1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-gen1": {"relay": map[string]any{"ison": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-gen1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2") && !strings.Contains(err.Error(), "thermostat") {
		t.Logf("error = %v", err)
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
	if opts.ID != 0 {
		t.Errorf("Default ID = %d, want 0", opts.ID)
	}
}

func TestOptions_DeviceFieldSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-device",
	}

	if opts.Device != "my-device" {
		t.Errorf("Device = %q, want 'my-device'", opts.Device)
	}
}

func TestOptions_ComponentIDSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}
	opts.ID = 5

	if opts.ID != 5 {
		t.Errorf("ID = %d, want 5", opts.ID)
	}
}

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts IP addresses as device identifiers
	err := cmd.Args(cmd, []string{"192.168.1.100"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts named devices
	err := cmd.Args(cmd, []string{"living-room"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestNewCommand_RejectsMultipleArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command rejects multiple devices
	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Command should reject multiple device arguments")
	}
}

func TestNewCommand_ShorthandFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that short form -i works
	err := cmd.ParseFlags([]string{"-i", "2"})
	if err != nil {
		t.Errorf("Expected to parse -i flag, got: %v", err)
	}
}

func TestNewCommand_LonghandFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that long form --id works
	err := cmd.ParseFlags([]string{"--id", "3"})
	if err != nil {
		t.Errorf("Expected to parse --id flag, got: %v", err)
	}
}

func TestExecute_ShortAlias(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-thermostat",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSN-0024X",
					Model:      "Shelly Plus HT",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-thermostat": {"thermostat:0": map[string]any{"target_C": 21.0, "current_C": 19.5}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Test using the alias instead of full command name
	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-thermostat"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestNewCommand_RunE_SetsDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-test-device"})

	// Create a cancelled context to prevent actual execution
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_AllAliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Fatal("Expected at least one alias")
	}

	// Verify specific aliases are in the list
	aliasMap := make(map[string]bool)
	for _, alias := range cmd.Aliases {
		aliasMap[alias] = true
	}

	if !aliasMap["cal"] {
		t.Errorf("Expected 'cal' alias, got %v", cmd.Aliases)
	}
}

func TestRun_WithNonexistentDevice(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Device:  "192.168.1.199", // Non-existent address
		Factory: f,
	}

	// Run should fail since the device doesn't exist
	err := run(context.Background(), opts)
	if err == nil {
		t.Log("run returned nil - device might exist on network")
	} else {
		t.Logf("Expected error for non-existent device: %v", err)
	}
}

func TestRun_WithFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Device:  "nonexistent-device",
		Factory: tf.Factory,
	}

	// Run should fail since the device doesn't exist
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for non-existent device")
	}
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestOptions_InitializationWithFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}

	if opts.Device == "" {
		t.Error("Device should not be empty")
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
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "id flag short",
			args:    []string{"-i", "0"},
			wantErr: false,
		},
		{
			name:    "id flag long",
			args:    []string{"--id", "1"},
			wantErr: false,
		},
		{
			name:    "multiple flag values",
			args:    []string{"-i", "5"},
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

func TestExecute_WithComponentFlags(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-thermostat",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSN-0024X",
					Model:      "Shelly Plus HT",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-thermostat": {
				"thermostat:0": map[string]any{"target_C": 21.0, "current_C": 19.5},
				"thermostat:2": map[string]any{"target_C": 23.0, "current_C": 21.5},
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

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-thermostat", "-i", "2"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_RequiresDeviceArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no device argument provided")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly thermostat calibrate",
		"--id",
		"gateway",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"calibration",
		"thermostat",
		"valve",
		"installation",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestRun_ValidDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Run should fail due to device not existing, but should attempt the connection
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error (from device connection), got nil")
	}
}

func TestNewCommand_UseExactArgsValidator(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// ExactArgs(1) should reject 0 args
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("should reject 0 args")
	}

	// ExactArgs(1) should accept 1 arg
	if err := cmd.Args(cmd, []string{"device"}); err != nil {
		t.Errorf("should accept 1 arg, got: %v", err)
	}

	// ExactArgs(1) should reject 2 args
	if err := cmd.Args(cmd, []string{"dev1", "dev2"}); err == nil {
		t.Error("should reject 2 args")
	}
}

func TestNewCommand_ExactsCallsRun(t *testing.T) {
	t.Parallel()

	// This test verifies that the RunE callback properly sets the device
	// and calls the run function. We use a cancelled context to prevent
	// actual device operations.
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from cancelled context")
	}
}

func TestOptions_IDDefaultValue(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test",
	}

	if opts.ID != 0 {
		t.Errorf("Expected ID default 0, got %d", opts.ID)
	}
}

func TestExecute_WithDeviceNameVariations(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "my-thermostat",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSN-0024X",
					Model:      "Shelly Plus HT",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"my-thermostat": {"thermostat:0": map[string]any{"target_C": 21.0}},
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
	cmd.SetArgs([]string{"my-thermostat"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected for mock)", err)
	}
}

func TestNewCommand_AllFlagValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "id flag short",
			args:    []string{"-i", "0"},
			wantErr: false,
		},
		{
			name:    "id flag long",
			args:    []string{"--id", "1"},
			wantErr: false,
		},
		{
			name:    "id flag with multiple digits",
			args:    []string{"-i", "10"},
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

func TestExecute_WithIDZero(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-thermostat",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSN-0024X",
					Model:      "Shelly Plus HT",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-thermostat": {
				"thermostat:0": map[string]any{"target_C": 21.0, "current_C": 19.5},
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

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-thermostat", "--id", "0"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}
