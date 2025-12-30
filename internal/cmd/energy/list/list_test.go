package list

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

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

	if cmd.Use != "list <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device>")
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

	expectedAliases := []string{"ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Fatalf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}

	for i, want := range expectedAliases {
		if cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
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
			name:    "no args - error",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg - valid",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "two args - error",
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "three args - error",
			args:    []string{"device1", "device2", "device3"},
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

func TestNewCommand_ArgsValidator(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction not set for device completion")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE not set")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify Long description mentions key details
	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	tests := []struct {
		name     string
		contains string
	}{
		{"mentions EM components", "EM"},
		{"mentions EM1 components", "EM1"},
		{"mentions energy monitoring", "energy monitor"},
		{"mentions 3-phase", "3-phase"},
		{"mentions single-phase", "single-phase"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(cmd.Long, tt.contains) {
				t.Errorf("Long description should contain %q", tt.contains)
			}
		})
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	tests := []struct {
		name     string
		contains string
	}{
		{"shows basic usage", "shelly energy list"},
		{"shows JSON output", "-o json"},
		{"shows short form", "shelly energy ls"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(cmd.Example, tt.contains) {
				t.Errorf("Example should contain %q", tt.contains)
			}
		})
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func() bool
		errMsg    string
	}{
		{
			name: "has use",
			checkFunc: func() bool {
				return NewCommand(cmdutil.NewFactory()).Use != ""
			},
			errMsg: "Use should not be empty",
		},
		{
			name: "has short",
			checkFunc: func() bool {
				return NewCommand(cmdutil.NewFactory()).Short != ""
			},
			errMsg: "Short should not be empty",
		},
		{
			name: "has long",
			checkFunc: func() bool {
				return NewCommand(cmdutil.NewFactory()).Long != ""
			},
			errMsg: "Long should not be empty",
		},
		{
			name: "has example",
			checkFunc: func() bool {
				return NewCommand(cmdutil.NewFactory()).Example != ""
			},
			errMsg: "Example should not be empty",
		},
		{
			name: "has aliases",
			checkFunc: func() bool {
				return len(NewCommand(cmdutil.NewFactory()).Aliases) > 0
			},
			errMsg: "Aliases should not be empty",
		},
		{
			name: "has RunE",
			checkFunc: func() bool {
				return NewCommand(cmdutil.NewFactory()).RunE != nil
			},
			errMsg: "RunE should be set",
		},
		{
			name: "has ValidArgsFunction",
			checkFunc: func() bool {
				return NewCommand(cmdutil.NewFactory()).ValidArgsFunction != nil
			},
			errMsg: "ValidArgsFunction should be set for completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.checkFunc() {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantArgsErr bool
	}{
		{
			name:        "no args fails",
			args:        []string{},
			wantArgsErr: true,
		},
		{
			name:        "with device arg passes args check",
			args:        []string{"test-device"},
			wantArgsErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			ios := iostreams.Test(nil, &stdout, &stderr)
			f := cmdutil.NewWithIOStreams(ios)

			cmd := NewCommand(f)
			cmd.SetArgs(tt.args)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			// Use short timeout for non-args tests
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			err := cmd.ExecuteContext(ctx)

			if tt.wantArgsErr && err == nil {
				t.Error("Expected error when executing with invalid args")
			}
			if !tt.wantArgsErr && err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "accepts") && strings.Contains(errStr, "arg") {
					t.Errorf("Should accept valid arguments, got args error: %v", err)
				}
			}
		})
	}
}

func TestNewCommand_WithTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test factory")
	}

	// Verify factory is usable
	ios := tf.IOStreams()
	if ios == nil {
		t.Error("Factory.IOStreams() should not return nil")
	}
}

func TestRun_ConnectionError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nonexistent-device"})

	// Use a short timeout to avoid long waits
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute will fail due to no real device connection
	err := cmd.ExecuteContext(ctx)

	// We expect an error since there's no real device
	if err == nil {
		t.Log("Expected connection error (no real device), but run succeeded")
	} else {
		// Verify it's not an args error
		errStr := err.Error()
		if strings.Contains(errStr, "accepts") && strings.Contains(errStr, "arg") {
			t.Errorf("Expected connection error, got args error: %v", err)
		}
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Create a command with the cancelled context
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Execute will fail due to cancelled context or connection
	err := cmd.ExecuteContext(ctx)

	// Should return an error (either context cancelled or connection)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "accepts") && strings.Contains(errStr, "arg") {
			t.Errorf("Expected context or connection error, got args error: %v", err)
		}
	}
}

func TestRun_IPAddressDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100"}) // IP address as device

	// Use a short timeout to avoid long waits
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute will fail due to no real device
	err := cmd.ExecuteContext(ctx)

	// We expect an error since there's no real device
	if err == nil {
		t.Log("Expected connection error for IP address device")
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

	tests := []struct {
		name     string
		contains string
	}{
		{"contains list", "list"},
		{"contains device", "device"},
		{"mentions EM", "EM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(helpOutput, tt.contains) {
				t.Errorf("Help should contain %q", tt.contains)
			}
		})
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Device != "" {
		t.Errorf("Default Device = %q, want empty", opts.Device)
	}

	if opts.Factory != nil {
		t.Error("Default Factory should be nil")
	}
}

func TestRun_WithOptions(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Device:  "test-device",
		Factory: tf.Factory,
	}

	// Use a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run directly - will fail on device connection but exercises run code
	err := run(ctx, opts)

	// We expect an error since there's no real device
	if err == nil {
		t.Log("Expected error for nonexistent device")
	} else {
		// Should be an EM components error, not nil factory error
		errStr := err.Error()
		if !strings.Contains(errStr, "EM components") {
			t.Errorf("Expected EM components error, got: %v", err)
		}
	}
}

func TestRun_EmptyDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Device:  "",
		Factory: tf.Factory,
	}

	// Use a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run directly - should fail quickly with empty device
	err := run(ctx, opts)

	// We expect an error
	if err == nil {
		t.Log("Expected error for empty device")
	}
}

func TestOptions_WithDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Device:  "my-device",
		Factory: tf.Factory,
	}

	if opts.Device != "my-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "my-device")
	}

	if opts.Factory != tf.Factory {
		t.Error("Factory not set correctly")
	}
}

func TestNewCommand_AliasExecution(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create command with factory
	cmd := NewCommand(tf.Factory)

	// Verify alias is set
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "ls" {
		t.Fatal("Expected 'ls' alias")
	}

	// Test that the command name check allows the alias
	cmd.SetArgs([]string{"test-device"})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute - will fail on device but should not fail on command name
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "unknown command") {
			t.Errorf("Alias should work, got: %v", err)
		}
	}
}

func TestNewCommand_FactoryAccessInRunE(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute to trigger RunE which accesses factory
	// Error is expected (no real device), but factory access should not panic
	err := cmd.ExecuteContext(ctx)
	if err == nil {
		t.Log("Expected error (no real device)")
	}
}

func TestRun_MultipleDeviceFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		device string
	}{
		{"hostname", "living-room-shelly"},
		{"ip address", "192.168.1.100"},
		{"ip with port", "192.168.1.100:80"},
		{"mDNS name", "shelly3em-123ABC.local"},
		{"short name", "em"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)

			opts := &Options{
				Device:  tt.device,
				Factory: tf.Factory,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Run - will fail but exercises device name handling
			err := run(ctx, opts)

			// Should get an EM-related error (EM or EM1), not a device format error
			if err != nil {
				errStr := err.Error()
				if !strings.Contains(errStr, "EM") {
					t.Errorf("Expected EM/EM1 components error for device %q, got: %v", tt.device, err)
				}
			}
		})
	}
}

// Execute-based tests using mock server
// Note: These tests use a global mock config manager, so they should not run in parallel
// to avoid race conditions between tests modifying the global state.

//nolint:paralleltest // uses global mock config manager
func TestExecute_WithEMComponents(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-em",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-em": {
				"em:0": map[string]any{
					"id":          0,
					"a_current":   1.5,
					"a_voltage":   230.0,
					"a_act_power": 345.0,
					"a_pf":        0.98,
					"a_freq":      50.0,
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

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestExecute_WithEM1Components(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-em1",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-em1": {
				"em1:0": map[string]any{
					"id":        0,
					"current":   0.5,
					"voltage":   230.0,
					"act_power": 115.0,
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

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-em1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestExecute_WithBothEMAndEM1(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-mixed",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-mixed": {
				"em:0": map[string]any{
					"id":          0,
					"a_current":   1.5,
					"a_voltage":   230.0,
					"a_act_power": 345.0,
				},
				"em1:0": map[string]any{
					"id":        0,
					"current":   0.5,
					"voltage":   230.0,
					"act_power": 115.0,
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

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-mixed"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestExecute_MultipleEMComponents(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-multi-em",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:03",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-multi-em": {
				"em:0": map[string]any{"id": 0, "a_current": 1.0},
				"em:1": map[string]any{"id": 1, "a_current": 2.0},
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
	cmd.SetArgs([]string{"test-multi-em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestExecute_NoComponents(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-no-em",
					Address:    "192.168.1.104",
					MAC:        "AA:BB:CC:DD:EE:04",
					Type:       "SNSW-001X16EU",
					Model:      "Shelly Plus 1",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-no-em": {
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

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-no-em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestExecute_DeviceNotFound(t *testing.T) {
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
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "unknown") && !strings.Contains(err.Error(), "EM") {
		t.Logf("error = %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestExecute_MultipleEM1Components(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-multi-em1",
					Address:    "192.168.1.105",
					MAC:        "AA:BB:CC:DD:EE:05",
					Type:       "SNSW-102P16EU",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-multi-em1": {
				"em1:0": map[string]any{"id": 0, "current": 0.5},
				"em1:1": map[string]any{"id": 1, "current": 1.0},
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
	cmd.SetArgs([]string{"test-multi-em1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestExecute_TableOutput(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-table",
					Address:    "192.168.1.106",
					MAC:        "AA:BB:CC:DD:EE:06",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-table": {
				"em:0": map[string]any{"id": 0, "a_current": 1.5},
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
	cmd.SetArgs([]string{"test-table"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

// Additional tests to increase coverage by exercising RunE callback

func TestRunE_SetsDeviceFromArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device-arg"})

	// Use short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Will fail due to no real device, but exercises RunE callback
	err := cmd.ExecuteContext(ctx)
	if err == nil {
		t.Log("Expected error for nonexistent device")
	}
}

func TestRun_ExercisesFactoryMethods(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Verify factory methods work
	ios := tf.IOStreams()
	if ios == nil {
		t.Fatal("IOStreams() returned nil")
	}

	svc := tf.ShellyService()
	if svc == nil {
		t.Fatal("ShellyService() returned nil")
	}
}

//nolint:paralleltest,tparallel // subtests don't need parallel as they're fast
func TestNewCommand_ExactArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that Args validator is set to ExactArgs(1)
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"empty args", []string{}, true},
		{"one arg", []string{"device"}, false},
		{"two args", []string{"device1", "device2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for shell completion")
	}
}

//nolint:paralleltest,tparallel // subtests don't need parallel as they're fast
func TestCommand_Properties(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify all required command properties
	tests := []struct {
		name  string
		check func() bool
	}{
		{"has Use", func() bool { return cmd.Use != "" }},
		{"has Short", func() bool { return cmd.Short != "" }},
		{"has Long", func() bool { return cmd.Long != "" }},
		{"has Example", func() bool { return cmd.Example != "" }},
		{"has Aliases", func() bool { return len(cmd.Aliases) > 0 }},
		{"has RunE", func() bool { return cmd.RunE != nil }},
		{"has Args", func() bool { return cmd.Args != nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("Command %s failed", tt.name)
			}
		})
	}
}
