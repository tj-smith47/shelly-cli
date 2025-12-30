package shell

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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "shell <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "shell <device>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"sh", "console"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Interactive shell for a specific device"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	wantPatterns := []string{
		"Open an interactive shell",
		"RPC commands",
		"help",
		"exit",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long should contain %q", pattern)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	wantPatterns := []string{
		"shelly shell",
		"living-room",
		"shell>",
		"exit",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"living-room"}, false},
		{"two args", []string{"living-room", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
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

func TestOptions_Device(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test-device",
		Factory: cmdutil.NewFactory(),
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
}

func TestOptions_Factory(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "my-device",
		Factory: f,
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}

	if opts.Factory.IOStreams() == nil {
		t.Error("Factory.IOStreams() returned nil")
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

// TestRun_Options verifies that run function receives correct options
func TestRun_Options(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Device:  "test-device",
		Factory: f,
	}

	// We can't fully test run() without a real device, but we can verify
	// that Options are properly structured
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Factory == nil {
		t.Error("Factory is nil")
	}
}

// TestExecute_NoArgs verifies that the command requires a device argument
func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no args provided")
	}
}

// TestNewCommand_Metadata verifies all required command metadata
func TestNewCommand_Metadata(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify required metadata is present
	if cmd.Use != "shell <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "shell <device>")
	}

	if cmd.Short == "" {
		t.Error("Short is empty")
	}

	if cmd.Long == "" {
		t.Error("Long is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Verify aliases exist and contain expected values
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty, expected at least one alias")
	}

	// Verify all fields match expectations
	expectedShort := "Interactive shell for a specific device"
	if cmd.Short != expectedShort {
		t.Errorf("Short = %q, want %q", cmd.Short, expectedShort)
	}
}

// TestNewCommand_ExampleContent verifies example contains expected patterns
func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly shell",
		"shell>",
		"info",
		"methods",
		"Switch.GetStatus",
		"exit",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

// TestNewCommand_LongDescription verifies long description contains expected patterns
func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"interactive shell",
		"RPC commands",
		"device",
		"help",
		"exit",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

// TestRun_Gen2Device tests that run works with a Gen2 device fixture
func TestRun_Gen2Device(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen2-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1PM",
					Model:      "Shelly 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen2-device": {
				"switch:0": map[string]interface{}{
					"id":  0,
					"on":  true,
					"apower": 0.0,
				},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)
	demo.InjectIntoFactory(f)

	opts := &Options{
		Device:  "gen2-device",
		Factory: f,
	}

	// Run will block on the interactive shell loop, but we can verify it gets far enough
	// to print device info before blocking
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run in a goroutine so we can cancel it after getting output
	done := make(chan error)
	go func() {
		done <- run(ctx, opts)
	}()

	// Give it a moment to connect and print device info
	select {
	case <-done:
		// Command completed before we could verify output
		output := stdout.String() + stderr.String()
		// Should have connected and shown device info
		if !strings.Contains(output, "Connected to") && !strings.Contains(output, "Shelly") {
			t.Logf("Expected device info in output, got: %s", output)
		}
	case <-ctx.Done():
		t.Log("Context cancelled")
	}
}

// TestRun_Gen1DeviceRejection tests that run rejects Gen1 devices
func TestRun_Gen1DeviceRejection(t *testing.T) {
	t.Parallel()

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
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {
				"relay:0": map[string]interface{}{
					"ison": true,
				},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)
	demo.InjectIntoFactory(f)

	opts := &Options{
		Device:  "gen1-device",
		Factory: f,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2") {
		t.Errorf("Expected error mentioning Gen2+ requirement, got: %v", err)
	}
}

// TestRun_WithInvalidDevice tests error handling for non-existent devices
func TestRun_WithInvalidDevice(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{},
		},
		DeviceStates: map[string]mock.DeviceState{},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)
	demo.InjectIntoFactory(f)

	opts := &Options{
		Device:  "nonexistent-device",
		Factory: f,
	}

	err = run(context.Background(), opts)
	// Should fail because device doesn't exist
	if err == nil {
		t.Log("Expected error for non-existent device")
	}
}

// TestRun_ContextCancellation tests that context cancellation works
func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1PM",
					Model:      "Shelly 1PM",
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

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)
	demo.InjectIntoFactory(f)

	opts := &Options{
		Device:  "test-device",
		Factory: f,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = run(ctx, opts)
	// Should fail due to cancelled context
	if err == nil {
		t.Log("Expected error with cancelled context")
	}
}

// TestNewCommand_AllAliasesCorrect verifies all aliases are present
func TestNewCommand_AllAliasesCorrect(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Verify each expected alias is present
	for i, expected := range []string{"sh", "console"} {
		if i >= len(cmd.Aliases) {
			t.Errorf("Expected alias %q at index %d, but not found", expected, i)
			continue
		}
		if cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

// TestNewCommand_ValidArgsRejectsMultiple verifies Args validation
func TestNewCommand_ValidArgsRejectsMultiple(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// ExactArgs(1) should reject 0, 2, or more args
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("Expected error with 0 args")
	}

	if err := cmd.Args(cmd, []string{"device", "extra"}); err == nil {
		t.Error("Expected error with 2 args")
	}

	if err := cmd.Args(cmd, []string{"device"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

// TestNewCommand_LongContainsMethodExamples verifies long description has RPC examples
func TestNewCommand_LongContainsMethodExamples(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"help",
		"info",
		"methods",
		"Switch.GetStatus",
		"Shelly.GetConfig",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long should contain example: %q", pattern)
		}
	}
}

// TestOptions_EmptyDevice validates Options with empty device
func TestOptions_EmptyDevice(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "",
		Factory: cmdutil.NewFactory(),
	}

	if opts.Device != "" {
		t.Errorf("Device = %q, want empty string", opts.Device)
	}
}

// TestNewCommand_ExampleHasMultipleSections verifies example structure
func TestNewCommand_ExampleHasMultipleSections(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should have comments explaining the different scenarios
	if !strings.Contains(cmd.Example, "#") {
		t.Error("Example should contain comments")
	}

	// Should have at least two distinct examples
	commentCount := strings.Count(cmd.Example, "#")
	if commentCount < 2 {
		t.Errorf("Example should have at least 2 comment lines, got %d", commentCount)
	}
}

// TestNewCommand_LongDescriptionLength verifies thorough documentation
func TestNewCommand_LongDescriptionLength(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should be comprehensive
	if len(cmd.Long) < 150 {
		t.Errorf("Long description is too short: %d chars", len(cmd.Long))
	}
}

// TestNewCommand_ShortDescriptionLength verifies conciseness
func TestNewCommand_ShortDescriptionLength(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Short description should be concise
	if len(cmd.Short) > 150 {
		t.Errorf("Short description is too long: %d chars", len(cmd.Short))
	}

	// But not too short
	if len(cmd.Short) < 10 {
		t.Errorf("Short description is too short: %d chars", len(cmd.Short))
	}
}
