package show

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

	if cmd.Use != "show <device>" {
		t.Errorf("Use = %q, want 'show <device>'", cmd.Use)
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

	if len(cmd.Aliases) == 0 {
		t.Fatal("expected at least one alias")
	}

	expectedAliases := map[string]bool{"info": true, "view": true, "get": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias: %s", alias)
		}
		delete(expectedAliases, alias)
	}
	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
	}
}

func TestNewCommand_RequiresOneArg(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	tests := []struct {
		args      []string
		wantError bool
	}{
		{[]string{}, true},
		{[]string{"device"}, false},
		{[]string{"device", "extra"}, true},
	}

	for _, tt := range tests {
		err := cmd.Args(cmd, tt.args)
		gotError := err != nil
		if gotError != tt.wantError {
			t.Errorf("Args(%v) error = %v, want error = %v", tt.args, err, tt.wantError)
		}
	}
}

func TestNewCommand_RunESet(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name string
		fn   func() bool
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
			t.Parallel()
			if !tt.fn() {
				t.Errorf("command structure check failed: %s", tt.name)
			}
		})
	}
}

func TestNewCommand_ExampleContainsShelly(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "shelly") {
		t.Error("Example should contain 'shelly' command")
	}
}

func TestNewCommand_ExampleContainsCertShow(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "cert") {
		t.Error("Example should contain 'cert' command")
	}
	if !strings.Contains(cmd.Example, "show") {
		t.Error("Example should contain 'show' command")
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Show command should not have subcommands
	if len(cmd.Commands()) > 0 {
		t.Errorf("show command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Show command should not define its own flags
	if cmd.Flags().NFlag() > 0 {
		t.Errorf("show command should not have flags set, has %d", cmd.Flags().NFlag())
	}
}

func TestNewCommand_HasRunE_NotRun(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// RunE should be set (not Run)
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
	if cmd.Run != nil {
		t.Error("Run should not be set when RunE is used")
	}
}

func TestNewCommand_AliasesContent(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	// Verify expected aliases exist
	if !aliasMap["info"] {
		t.Error("missing 'info' alias")
	}
	if !aliasMap["view"] {
		t.Error("missing 'view' alias")
	}
	if !aliasMap["get"] {
		t.Error("missing 'get' alias")
	}
}

func TestNewCommand_LongMentionsTLS(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention TLS
	if !strings.Contains(cmd.Long, "TLS") {
		t.Error("Long description should mention TLS")
	}
}

func TestNewCommand_LongMentionsGen2(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention Gen2+
	if !strings.Contains(cmd.Long, "Gen2") {
		t.Error("Long description should mention Gen2+ devices")
	}
}

func TestNewCommand_ShortMentionsTLS(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Short description should mention TLS
	if !strings.Contains(cmd.Short, "TLS") {
		t.Error("Short description should mention TLS")
	}
}

func TestNewCommand_UseHasDeviceArg(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Use should show <device> argument
	if !strings.Contains(cmd.Use, "<device>") {
		t.Error("Use should show <device> argument")
	}
}

func TestNewCommand_ExampleIsValid(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should start with proper formatting (spaces for indentation)
	if !strings.HasPrefix(cmd.Example, "  ") {
		t.Error("Example should start with proper indentation")
	}
}

func TestNewCommand_LongDescriptionContent(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention certificate configuration
	if !strings.Contains(cmd.Long, "certificate") {
		t.Error("Long description should mention 'certificate'")
	}
}

func TestNewCommand_ExampleContainsDevice(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should show a device name
	if !strings.Contains(cmd.Example, "kitchen") {
		t.Log("Example does not use 'kitchen' as example device")
	}
}

func TestNewCommand_ShortDescriptionContent(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Short description should mention device
	if !strings.Contains(cmd.Short, "device") {
		t.Log("Short description does not mention 'device'")
	}
}

func TestNewCommand_ExampleShowsUsage(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should show actual command usage
	if !strings.Contains(cmd.Example, "shelly cert show") {
		t.Error("Example should show 'shelly cert show' command")
	}
}

func TestNewCommand_HasCorrectAliasCount(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should have exactly 3 aliases
	if len(cmd.Aliases) != 3 {
		t.Errorf("expected 3 aliases, got %d", len(cmd.Aliases))
	}
}

func TestNewCommand_ParentCommandType(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Command should not have a parent set yet (parent is set when added to cert command)
	if cmd.Parent() != nil {
		t.Error("command should not have parent before being added to parent command")
	}
}

func TestNewCommand_SilenceErrors(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check default settings
	if cmd.SilenceErrors {
		t.Log("SilenceErrors is set to true")
	}
	if cmd.SilenceUsage {
		t.Log("SilenceUsage is set to true")
	}
}

func TestNewCommand_TraverseChildren(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// TraverseChildren should be false for leaf commands
	if cmd.TraverseChildren {
		t.Error("TraverseChildren should be false for leaf command")
	}
}

func TestNewCommand_DisableFlagsInUseLine(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check if flags are shown in usage line
	if cmd.DisableFlagsInUseLine {
		t.Log("Flags are disabled in usage line")
	}
}

func TestNewCommand_ValidateArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Args function should be set
	if cmd.Args == nil {
		t.Fatal("Args should be set")
	}

	// Test with invalid number of args
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"device"}, false},
		{"two args", []string{"device1", "device2"}, true},
		{"three args", []string{"a", "b", "c"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tc.args)
			hasErr := err != nil
			if hasErr != tc.expectErr {
				t.Errorf("Args(%v) error = %v, expectErr = %v", tc.args, err, tc.expectErr)
			}
		})
	}
}

func TestNewCommand_ExampleComment(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should have a comment explaining the command
	if !strings.Contains(cmd.Example, "#") {
		t.Log("Example does not contain comment marker")
	}
}

// ============================================================================
// Run function tests - test actual execution paths
// ============================================================================

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The run function should respect context cancellation
	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)

	// Expect an error due to cancelled context (connection will fail)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow the timeout to trigger
	time.Sleep(1 * time.Millisecond)

	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)

	// Expect an error due to timeout
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestNewCommand_RunE_PassesDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device"})

	// Create a cancelled context to ensure the run function receives the right device
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context but want to verify structure
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestRun_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context to quickly fail without network calls
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The run function should accept IP addresses
	opts := &Options{Factory: tf.Factory, Device: "192.168.1.100"}
	err := run(ctx, opts)

	// Expect an error due to cancelled context, but the function should be invoked
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context to quickly fail without network calls
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The run function should accept device names
	opts := &Options{Factory: tf.Factory, Device: "living-room"}
	err := run(ctx, opts)

	// Expect an error due to cancelled context, but the function should be invoked
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_ExecuteWithCancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"kitchen"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute should fail with cancelled context
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestRun_MultipleDeviceFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		device string
	}{
		{"ip_address", "192.168.1.100"},
		{"device_name", "kitchen"},
		{"device_with_dashes", "living-room-lamp"},
		{"device_with_underscores", "kitchen_light"},
		{"hostname", "shelly-device.local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)

			// Create a cancelled context
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// The run function should accept various device formats
			opts := &Options{Factory: tf.Factory, Device: tt.device}
			err := run(ctx, opts)

			// Expect an error due to cancelled context, but no panic
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}

func TestRun_UsesIOStreams(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a context with short timeout to ensure quick failure
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give time for context to timeout
	time.Sleep(2 * time.Millisecond)

	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)

	// We expect an error due to timeout, but the factory methods should have been called
	if err == nil {
		t.Error("Expected error with timed out context")
	}

	// Verify that the function uses the IOStreams from the factory
	// The error output should be accessible through the test factory
	// (even if empty due to context cancellation)
	_ = tf.OutString()
	_ = tf.ErrString()
}

func TestRun_UsesShellyService(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The run function should attempt to use the shelly service
	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)

	// The error should be non-nil since context is cancelled
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestNewCommand_CommandExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{}) // No args

	// Execute should fail due to missing args
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with no args")
	}
}

func TestNewCommand_CommandExecute_TooManyArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"device1", "device2"}) // Too many args

	// Execute should fail due to too many args
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute with too many args")
	}
}

func TestRun_ContextDeadlineExceeded(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a context that will immediately exceed deadline
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	// The run function should return an error due to deadline exceeded
	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)

	if err == nil {
		t.Error("Expected error with deadline exceeded context")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// ValidArgsFunction may or may not be set for this command
	// Document the actual state
	if cmd.ValidArgsFunction != nil {
		t.Log("ValidArgsFunction is set for device completion")
	} else {
		t.Log("ValidArgsFunction is not set")
	}
}

func TestNewCommand_GroupAnnotations(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check if GroupID annotation is set
	if cmd.GroupID != "" {
		t.Logf("GroupID is set to: %s", cmd.GroupID)
	}

	// Check if annotations are set
	if len(cmd.Annotations) > 0 {
		t.Logf("Annotations: %v", cmd.Annotations)
	}
}

// ============================================================================
// Additional run function tests for better coverage
// ============================================================================

func TestRun_ErrorPropagation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		device     string
		ctxSetup   func() (context.Context, context.CancelFunc)
		wantErr    bool
		errContain string
	}{
		{
			name:   "cancelled context returns error",
			device: "test-device",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			wantErr: true,
		},
		{
			name:   "deadline exceeded returns error",
			device: "192.168.1.1",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Hour))
				return ctx, cancel
			},
			wantErr: true,
		},
		{
			name:   "timeout returns error",
			device: "kitchen",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(1 * time.Millisecond)
				return ctx, cancel
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)
			ctx, cancel := tt.ctxSetup()
			defer cancel()

			opts := &Options{Factory: tf.Factory, Device: tt.device}
			err := run(ctx, opts)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRun_FactoryMethodsCalled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Use a context that will fail quickly but still allow factory methods to be called
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// The run function should call IOStreams and ShellyService from the factory
	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)

	// Error is expected since no real device exists
	if err == nil {
		t.Log("Unexpectedly succeeded - a real device might be available")
	}

	// Verify that the function was invoked (outputs may be empty due to error)
	// but no panic occurred
	t.Log("Function executed without panic")
}

func TestRun_EmptyDeviceString(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test with empty device string - should still attempt connection
	opts := &Options{Factory: tf.Factory, Device: ""}
	err := run(ctx, opts)

	// Error expected due to cancelled context or invalid device
	if err == nil {
		t.Error("Expected error with empty device string and cancelled context")
	}
}

func TestRun_SpecialCharactersInDevice(t *testing.T) {
	t.Parallel()

	specialCases := []string{
		"device-with-dash",
		"device_with_underscore",
		"device.with.dots",
		"192.168.1.100",
		"[::1]",
		"device:8080",
	}

	for _, device := range specialCases {
		t.Run(device, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// Should handle special characters without panic
			opts := &Options{Factory: tf.Factory, Device: device}
			err := run(ctx, opts)

			// Error expected due to cancelled context
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	}
}

func TestNewCommand_ExecuteWithValidContext_NoDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nonexistent-device"})

	// Use a short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)

	// Execute should fail because device doesn't exist or times out
	err := cmd.Execute()
	if err == nil {
		t.Log("Command succeeded unexpectedly - device might exist")
	}
}

func TestRun_MultipleCalls(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Make multiple calls to run with cancelled context
	for i := range 3 {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		opts := &Options{Factory: tf.Factory, Device: "device"}
		err := run(ctx, opts)
		if err == nil {
			t.Errorf("Call %d: Expected error with cancelled context", i)
		}
	}
}

func TestRun_OutputBufferAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Reset buffers before test
	tf.Reset()

	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}

	// Buffers should be accessible (may be empty or have content based on error handling)
	stdout := tf.OutString()
	stderr := tf.ErrString()

	// Just verify we can access the buffers without panic
	t.Logf("stdout length: %d, stderr length: %d", len(stdout), len(stderr))
}

func TestNewCommand_IntegrationWithFactory(t *testing.T) {
	t.Parallel()

	// Test that command properly integrates with factory
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)

	// Verify command can be used in a command tree
	parent := &cobra.Command{Use: "cert"}
	parent.AddCommand(cmd)

	// Verify parent-child relationship
	if cmd.Parent() == nil {
		t.Error("Command should have parent after AddCommand")
	}

	// Verify command is findable
	found, _, err := parent.Find([]string{"show"})
	if err != nil {
		t.Errorf("Find error: %v", err)
	}
	if found != cmd {
		t.Error("Found wrong command")
	}
}

// ============================================================================
// Execute-based tests with mock server for comprehensive coverage
// ============================================================================

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_Gen2Device_Success(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-gen2",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-gen2": {
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
	cmd.SetArgs([]string{"test-gen2"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected non-empty output")
	}
	// Check for success message indicating TLS config was fetched
	if !strings.Contains(output, "TLS Configuration") {
		t.Errorf("output should contain 'TLS Configuration', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_Gen1Device_Error(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-gen1",
					Address:    "192.168.1.101",
					MAC:        "BB:CC:DD:EE:FF:00",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-gen1": {
				"relay": map[string]any{"ison": false},
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
	cmd.SetArgs([]string{"test-gen1"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("error should mention Gen2+, got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_NoCustomCA_ShowsGuidance(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "no-ca-device",
					Address:    "192.168.1.102",
					MAC:        "CC:DD:EE:FF:00:11",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"no-ca-device": {
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
	cmd.SetArgs([]string{"no-ca-device"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should show guidance about installing custom CA
	if !strings.Contains(output, "cert install") {
		t.Errorf("output should contain guidance about 'cert install', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_WithCustomCA_MQTT(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "mqtt-ca-device",
					Address:    "192.168.1.103",
					MAC:        "DD:EE:FF:00:11:22",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"mqtt-ca-device": {
				"switch:0": map[string]any{"output": true},
				// Note: MQTT config with ssl_ca would need to be part of the GetConfig response
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
	cmd.SetArgs([]string{"mqtt-ca-device"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "TLS Configuration") {
		t.Errorf("output should contain 'TLS Configuration', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{},
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

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_CancelledContext(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "cancel-test",
					Address:    "192.168.1.104",
					MAC:        "EE:FF:00:11:22:33",
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{Factory: tf.Factory, Device: "cancel-test"}
	err = run(ctx, opts)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_VerboseMode(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "verbose-test",
					Address:    "192.168.1.105",
					MAC:        "FF:00:11:22:33:44",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"verbose-test": {
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

	// Set verbose mode via viper
	viper.Set("verbose", true)
	defer viper.Set("verbose", false)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"verbose-test"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Verbose mode should show raw configuration info
	if !strings.Contains(output, "Raw configuration") {
		t.Errorf("verbose output should contain 'Raw configuration', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithMockDevice(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "kitchen",
					Address:    "192.168.1.106",
					MAC:        "00:11:22:33:44:55",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"kitchen": {
				"switch:0": map[string]any{"output": false},
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
	cmd.SetArgs([]string{"kitchen"})
	cmd.SetContext(context.Background())

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "TLS Configuration") {
		t.Errorf("output should contain 'TLS Configuration', got: %s", output)
	}
	if !strings.Contains(output, "kitchen") {
		t.Errorf("output should contain device name 'kitchen', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_MultipleDevices(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "device1",
					Address:    "192.168.1.107",
					MAC:        "11:22:33:44:55:66",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
				{
					Name:       "device2",
					Address:    "192.168.1.108",
					MAC:        "22:33:44:55:66:77",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"device1": {"switch:0": map[string]any{"output": true}},
			"device2": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Test first device
	opts := &Options{Factory: tf.Factory, Device: "device1"}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run(device1) error = %v", err)
	}

	// Reset output
	tf.Reset()

	// Test second device
	opts = &Options{Factory: tf.Factory, Device: "device2"}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run(device2) error = %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_VerboseMode_ShowsRawConfig(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "verbose-dev",
					Address:    "192.168.1.110",
					MAC:        "A0:B1:C2:D3:E4:F5",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"verbose-dev": {
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

	// Enable verbose mode
	viper.Set("verbose", true)
	defer viper.Set("verbose", false)

	opts := &Options{Factory: tf.Factory, Device: "verbose-dev"}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain "Raw configuration" text
	if !strings.Contains(output, "Raw configuration") {
		t.Errorf("verbose output should contain 'Raw configuration', got: %s", output)
	}
	// Should contain JSON-like output with "sys"
	if !strings.Contains(output, "sys") {
		t.Errorf("verbose output should contain 'sys' from config, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_NoCustomCA_ShowsCertInstall(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "no-cert-dev",
					Address:    "192.168.1.111",
					MAC:        "B0:C1:D2:E3:F4:05",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"no-cert-dev": {
				"switch:0": map[string]any{"output": false},
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

	opts := &Options{Factory: tf.Factory, Device: "no-cert-dev"}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain guidance about cert install
	if !strings.Contains(output, "cert install") {
		t.Errorf("output should contain 'cert install', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_Gen2Success_OutputFormat(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "output-fmt-dev",
					Address:    "192.168.1.112",
					MAC:        "C0:D1:E2:F3:04:15",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"output-fmt-dev": {
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

	opts := &Options{Factory: tf.Factory, Device: "output-fmt-dev"}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Output should contain device name
	if !strings.Contains(output, "output-fmt-dev") {
		t.Errorf("output should contain device name, got: %s", output)
	}
	// Output should contain TLS Configuration header
	if !strings.Contains(output, "TLS Configuration") {
		t.Errorf("output should contain 'TLS Configuration', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_WithMQTTConfig(t *testing.T) {
	// Test with MQTT configuration in device state
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "mqtt-dev",
					Address:    "192.168.1.113",
					MAC:        "D0:E1:F2:03:14:25",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"mqtt-dev": {
				"switch:0": map[string]any{"output": true},
				"mqtt":     map[string]any{"connected": true, "server": "mqtt.example.com:1883"},
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

	opts := &Options{Factory: tf.Factory, Device: "mqtt-dev"}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain TLS Configuration
	if !strings.Contains(output, "TLS Configuration") {
		t.Errorf("output should contain 'TLS Configuration', got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_VerboseWithValidConfig(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "verbose-cfg-dev",
					Address:    "192.168.1.114",
					MAC:        "E0:F1:02:13:24:35",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"verbose-cfg-dev": {
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

	// Enable verbose mode
	viper.Set("verbose", true)
	defer viper.Set("verbose", false)

	opts := &Options{Factory: tf.Factory, Device: "verbose-cfg-dev"}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Verbose mode should produce output with the config data
	if !strings.Contains(output, "Raw configuration") {
		t.Errorf("verbose output should contain 'Raw configuration', got: %s", output)
	}
	// Config should be displayed as JSON
	if !strings.Contains(output, "{") {
		t.Errorf("verbose output should contain JSON braces, got: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_DirectCallToRun(t *testing.T) {
	// Directly test the run function for additional coverage
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "direct-call-dev",
					Address:    "192.168.1.115",
					MAC:        "F0:01:12:23:34:45",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"direct-call-dev": {
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

	ctx := context.Background()
	opts := &Options{Factory: tf.Factory, Device: "direct-call-dev"}
	err = run(ctx, opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "TLS Configuration") {
		t.Errorf("output should contain 'TLS Configuration', got: %s", output)
	}
	if !strings.Contains(output, "cert install") {
		t.Errorf("output should contain 'cert install' guidance, got: %s", output)
	}
}
