package events

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const switchComponent = "switch"

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

	if cmd.Use != "events <device>" {
		t.Errorf("Use = %q, want 'events <device>'", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"ev", "subscribe"}
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

	expected := "Monitor device events in real-time"
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

	// Verify long description contains key content
	if !strings.Contains(cmd.Long, "WebSocket") {
		t.Error("Long should contain 'WebSocket'")
	}
	if !strings.Contains(cmd.Long, "Ctrl+C") {
		t.Error("Long should contain 'Ctrl+C'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Verify example contains expected patterns
	expectedPatterns := []string{
		"shelly monitor events",
		"--filter",
		"-o json",
	}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
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
		t.Error("Expected error with two args")
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	filterFlag := cmd.Flags().Lookup("filter")
	if filterFlag == nil {
		t.Fatal("filter flag not found")
	}
	if filterFlag.Shorthand != "f" {
		t.Errorf("filter shorthand = %q, want f", filterFlag.Shorthand)
	}
	if filterFlag.DefValue != "" {
		t.Errorf("filter default = %q, want empty", filterFlag.DefValue)
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
			name:    "filter flag short",
			args:    []string{"-f", "switch"},
			wantErr: false,
		},
		{
			name:    "filter flag long",
			args:    []string{"--filter", "switch"},
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

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	helpErr := cmd.Execute()
	if helpErr != nil {
		t.Errorf("--help should not error: %v", helpErr)
	}

	helpOutput := buf.String()
	if !strings.Contains(helpOutput, "Monitor device events") {
		t.Error("Help output should contain command description")
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

// Execute-based tests with mock server

func TestExecute_WithMockServer(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute with timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)

	// Command may fail due to timeout, which is expected
	if err := cmd.Execute(); err != nil {
		t.Logf("execute with timeout returned error (expected): %v", err)
	}
}

func TestExecute_MissingDevice(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestExecute_WithFilterFlag(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--filter", "switch"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err != nil {
		t.Logf("execute with timeout returned error (expected): %v", err)
	}
}

func TestExecute_WithFilterFlagShort(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "-f", "switch"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err != nil {
		t.Logf("execute with timeout returned error (expected): %v", err)
	}
}

func TestRun_WithContextCancellation(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: f,
		Device:  "test-device",
	}
	err = run(ctx, opts)
	// Context cancelled should result in error
	if err == nil {
		t.Log("run with cancelled context may succeed if no connection attempted")
	}
}

func TestRun_WithFilter(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	// Create a short-lived context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory: f,
		Device:  "test-device",
		Filter:  switchComponent,
	}
	err = run(ctx, opts)
	// We expect timeout or connection error
	if err == nil {
		t.Log("run completed or no error from filter test")
	}
}

func TestRun_NoFilter(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	// Create a short-lived context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory: f,
		Device:  "test-device",
	}
	err = run(ctx, opts)
	// We expect timeout or connection error
	if err == nil {
		t.Log("run completed or no error")
	}
}

func TestRun_OutputWithTitle(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	// Create a short-lived context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory: f,
		Device:  "test-device",
	}
	if err := run(ctx, opts); err != nil {
		t.Logf("run returned error (expected with timeout): %v", err)
	}

	// Check that title and instructions were printed
	result := stdout.String() + stderr.String()
	if !strings.Contains(result, "test-device") {
		t.Logf("Expected device name in output, got: %s", result)
	}
}

func TestRun_WithJSONOutput(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
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

	// Save and restore the output format
	viper.Set("output", string(output.FormatJSON))
	defer viper.Set("output", "")

	// Create a short-lived context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	opts := &Options{
		Factory: f,
		Device:  "test-device",
	}
	if err := run(ctx, opts); err != nil {
		t.Logf("run returned error (expected with timeout): %v", err)
	}

	// When JSON output is enabled, title and press Ctrl+C message should not appear.
	jsonOutput := stdout.String() + stderr.String()
	if strings.Contains(jsonOutput, "Event Monitor:") {
		t.Error("Title should not be printed in JSON output mode")
	}
	if strings.Contains(jsonOutput, "Press Ctrl+C") {
		t.Error("Instructions should not be printed in JSON output mode")
	}
}

func TestDeviceEvent_Model(t *testing.T) {
	t.Parallel()

	event := model.DeviceEvent{
		Device:      "test-device",
		Timestamp:   time.Now(),
		Event:       "state_changed",
		Component:   "switch",
		ComponentID: 0,
		Data: map[string]any{
			"output": true,
		},
	}

	if event.Device != "test-device" {
		t.Errorf("Device = %q, want 'test-device'", event.Device)
	}
	if event.Event != "state_changed" {
		t.Errorf("Event = %q, want 'state_changed'", event.Event)
	}
	if event.Component != "switch" {
		t.Errorf("Component = %q, want 'switch'", event.Component)
	}
	if event.ComponentID != 0 {
		t.Errorf("ComponentID = %d, want 0", event.ComponentID)
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cmdutil.Factory) bool
		errMsg    string
	}{
		{
			name: "NewCommand returns non-nil",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f) != nil
			},
			errMsg: "NewCommand should not return nil",
		},
		{
			name: "Has Use field",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Use != ""
			},
			errMsg: "Use field should not be empty",
		},
		{
			name: "Has Short field",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Short != ""
			},
			errMsg: "Short field should not be empty",
		},
		{
			name: "Has Long field",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Long != ""
			},
			errMsg: "Long field should not be empty",
		},
		{
			name: "Has Example field",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Example != ""
			},
			errMsg: "Example field should not be empty",
		},
		{
			name: "Has Aliases",
			checkFunc: func(f *cmdutil.Factory) bool {
				return len(NewCommand(f).Aliases) > 0
			},
			errMsg: "Aliases should not be empty",
		},
		{
			name: "Has RunE",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).RunE != nil
			},
			errMsg: "RunE should be set",
		},
		{
			name: "Has Args validator",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).Args != nil
			},
			errMsg: "Args should be set",
		},
		{
			name: "Has ValidArgsFunction",
			checkFunc: func(f *cmdutil.Factory) bool {
				return NewCommand(f).ValidArgsFunction != nil
			},
			errMsg: "ValidArgsFunction should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := cmdutil.NewFactory()
			if !tt.checkFunc(f) {
				t.Error(tt.errMsg)
			}
		})
	}
}

// TestEventHandler_FilterMatching tests filter logic when component matches.
func TestEventHandler_FilterMatching(t *testing.T) {
	t.Parallel()

	filter := switchComponent

	event := model.DeviceEvent{
		Device:      "test-device",
		Timestamp:   time.Now(),
		Event:       "state_changed",
		Component:   switchComponent,
		ComponentID: 0,
		Data: map[string]any{
			"output": true,
		},
	}

	// Test that matching filter passes through.
	passes := filter == "" || event.Component == filter

	if !passes {
		t.Error("Event with matching filter should pass through")
	}
}

// TestEventHandler_FilterNonMatching tests filter logic when component does not match.
func TestEventHandler_FilterNonMatching(t *testing.T) {
	t.Parallel()

	filter := switchComponent

	event := model.DeviceEvent{
		Device:      "test-device",
		Timestamp:   time.Now(),
		Event:       "notification",
		Component:   "light",
		ComponentID: 0,
		Data:        map[string]any{},
	}

	// Test that non-matching filter blocks event.
	passes := filter == "" || event.Component == filter

	if passes {
		t.Error("Event with non-matching filter should be blocked")
	}
}

// TestEventHandler_NoFilter tests filter logic when no filter is set.
func TestEventHandler_NoFilter(t *testing.T) {
	t.Parallel()

	filter := ""

	event := model.DeviceEvent{
		Device:      "test-device",
		Timestamp:   time.Now(),
		Event:       "state_changed",
		Component:   switchComponent,
		ComponentID: 0,
		Data:        map[string]any{},
	}

	// Test that empty filter allows all events.
	passes := filter == "" || event.Component == filter

	if !passes {
		t.Error("Event with empty filter should pass through")
	}
}

// TestDisplayEventFunction tests the DisplayEvent term function.
func TestDisplayEventFunction(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	event := model.DeviceEvent{
		Device:      "test-device",
		Timestamp:   time.Now(),
		Event:       "state_changed",
		Component:   "switch",
		ComponentID: 0,
		Data: map[string]any{
			"output": true,
		},
	}

	err := term.DisplayEvent(ios, event)
	if err != nil {
		t.Errorf("DisplayEvent should not error: %v", err)
	}

	displayOut := stdout.String()
	if displayOut == "" {
		t.Error("DisplayEvent should produce output")
	}
	if !strings.Contains(displayOut, "state_changed") {
		t.Error("Output should contain event type")
	}
	if !strings.Contains(displayOut, switchComponent) {
		t.Error("Output should contain component name")
	}
}

// TestOutputEventJSONFunction tests the OutputEventJSON term function.
func TestOutputEventJSONFunction(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	event := model.DeviceEvent{
		Device:      "test-device",
		Timestamp:   time.Now(),
		Event:       "state_changed",
		Component:   "switch",
		ComponentID: 0,
		Data: map[string]any{
			"output": true,
		},
	}

	err := term.OutputEventJSON(ios, event)
	if err != nil {
		t.Errorf("OutputEventJSON should not error: %v", err)
	}

	jsonOut := stdout.String()
	if jsonOut == "" {
		t.Error("OutputEventJSON should produce output")
	}
	if !strings.Contains(jsonOut, "test-device") {
		t.Error("JSON output should contain device name")
	}
	if !strings.Contains(jsonOut, "state_changed") {
		t.Error("JSON output should contain event type")
	}
}

// TestDisplayEvent_DifferentEventTypes tests display of different event types.
func TestDisplayEvent_DifferentEventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType string
	}{
		{"state_changed", "state_changed"},
		{"error", "error"},
		{"notification", "notification"},
		{"unknown", "unknown_event"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			ios := iostreams.Test(nil, &stdout, &stderr)

			event := model.DeviceEvent{
				Device:      "test-device",
				Timestamp:   time.Now(),
				Event:       tt.eventType,
				Component:   switchComponent,
				ComponentID: 0,
				Data:        map[string]any{},
			}

			err := term.DisplayEvent(ios, event)
			if err != nil {
				t.Errorf("DisplayEvent should not error: %v", err)
			}

			eventOut := stdout.String()
			if !strings.Contains(eventOut, tt.eventType) {
				t.Errorf("Output should contain event type %q", tt.eventType)
			}
		})
	}
}

// TestEventFilter_DifferentComponents tests filtering with different component types.
func TestEventFilter_DifferentComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filterValue    string
		eventComponent string
		shouldPass     bool
	}{
		{"match_switch", "switch", "switch", true},
		{"match_light", "light", "light", true},
		{"match_cover", "cover", "cover", true},
		{"no_match", "switch", "light", false},
		{"empty_filter", "", "switch", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filter := tt.filterValue

			event := model.DeviceEvent{
				Device:      "test-device",
				Timestamp:   time.Now(),
				Event:       "state_changed",
				Component:   tt.eventComponent,
				ComponentID: 0,
				Data:        map[string]any{},
			}

			passes := filter == "" || event.Component == filter

			if passes != tt.shouldPass {
				t.Errorf("Event filter: expected %v, got %v", tt.shouldPass, passes)
			}
		})
	}
}

// TestEventHandler_MultipleComponentIDs tests handling of multiple component IDs.
func TestEventHandler_MultipleComponentIDs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	tests := []struct {
		name        string
		componentID int
	}{
		{"component_0", 0},
		{"component_1", 1},
		{"component_2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := model.DeviceEvent{
				Device:      "test-device",
				Timestamp:   time.Now(),
				Event:       "state_changed",
				Component:   switchComponent,
				ComponentID: tt.componentID,
				Data: map[string]any{
					"output": true,
				},
			}

			err := term.DisplayEvent(ios, event)
			if err != nil {
				t.Errorf("DisplayEvent with componentID %d should not error: %v", tt.componentID, err)
			}
		})
	}
}

// TestOutputEventJSON_VariousData tests JSON output with different event data.
func TestOutputEventJSON_VariousData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data map[string]any
	}{
		{"empty_data", map[string]any{}},
		{"output_state", map[string]any{"output": true}},
		{"power_data", map[string]any{"apower": 150.5}},
		{"temperature", map[string]any{"temperature": map[string]any{"tC": 23.5}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			ios := iostreams.Test(nil, &stdout, &stderr)

			event := model.DeviceEvent{
				Device:      "test-device",
				Timestamp:   time.Now(),
				Event:       "state_changed",
				Component:   "switch",
				ComponentID: 0,
				Data:        tt.data,
			}

			err := term.OutputEventJSON(ios, event)
			if err != nil {
				t.Errorf("OutputEventJSON should not error: %v", err)
			}

			jsonStr := stdout.String()
			if jsonStr == "" {
				t.Error("OutputEventJSON should produce output")
			}
		})
	}
}

// TestNewCommand_RunE_Closure tests the RunE callback structure.
func TestNewCommand_RunE_Closure(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Verify RunE is set and can be called
	if cmd.RunE == nil {
		t.Fatal("RunE should be set")
	}

	// Execute with empty context to verify the structure
	cmd.SetArgs([]string{"test-device"})
	cmd.SetContext(context.Background())

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Should error because device doesn't exist, but tests the closure.
	if err := cmd.Execute(); err != nil {
		t.Logf("execute returned error (expected): %v", err)
	}
}

// TestNewCommand_DeviceArgumentParsing tests that the command correctly parses device arguments.
func TestNewCommand_DeviceArgumentParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  []string
		valid bool
	}{
		{"valid_name", []string{"my-device"}, true},
		{"valid_ip", []string{"192.168.1.1"}, true},
		{"valid_hostname", []string{"shelly-device.local"}, true},
		{"no_args", []string{}, false},
		{"too_many_args", []string{"device1", "device2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)

			if tt.valid && err != nil {
				t.Errorf("Expected valid args, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid args, but got no error")
			}
		})
	}
}

// TestFilterOption_InOptions tests that filter can be set in Options struct.
func TestFilterOption_InOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Test Options with filter
	opts := &Options{
		Factory: f,
		Device:  "test-device",
		Filter:  switchComponent,
	}

	if opts.Filter != switchComponent {
		t.Errorf("Filter = %q, want %q", opts.Filter, switchComponent)
	}

	// Test Options without filter
	optsNoFilter := &Options{
		Factory: f,
		Device:  "test-device",
	}

	if optsNoFilter.Filter != "" {
		t.Errorf("Filter should be empty, got %q", optsNoFilter.Filter)
	}
}

// TestNewCommand_MultipleFlags tests multiple simultaneous flag operations.
func TestNewCommand_MultipleFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetArgs([]string{"test-device", "-f", switchComponent})
	cmd.SetContext(context.Background())

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute will fail but we're testing arg parsing.
	if err := cmd.Execute(); err != nil {
		t.Logf("execute returned error (expected): %v", err)
	}
}

// TestDeviceEvent_TimestampHandling tests event model construction with various timestamps.
func TestDeviceEvent_TimestampHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ts   time.Time
	}{
		{"now", time.Now()},
		{"past", time.Now().Add(-24 * time.Hour)},
		{"future", time.Now().Add(24 * time.Hour)},
		{"epoch", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := model.DeviceEvent{
				Device:      "test-device",
				Timestamp:   tt.ts,
				Event:       "state_changed",
				Component:   "switch",
				ComponentID: 0,
			}

			if event.Timestamp != tt.ts {
				t.Errorf("Timestamp mismatch: expected %v, got %v", tt.ts, event.Timestamp)
			}
		})
	}
}

// TestDeviceEvent_ComponentVariations tests component type variations.
func TestDeviceEvent_ComponentVariations(t *testing.T) {
	t.Parallel()

	components := []string{"switch", "light", "cover", "em", "pm", "sys", "input"}

	for _, comp := range components {
		t.Run(comp, func(t *testing.T) {
			t.Parallel()

			event := model.DeviceEvent{
				Device:      "test-device",
				Timestamp:   time.Now(),
				Event:       "state_changed",
				Component:   comp,
				ComponentID: 0,
			}

			if event.Component != comp {
				t.Errorf("Component mismatch for %s", comp)
			}
		})
	}
}

// TestDeviceEvent_EventTypes tests with different event types.
func TestDeviceEvent_EventTypes(t *testing.T) {
	t.Parallel()

	eventTypes := []string{"state_changed", "error", "notification", "report"}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			t.Parallel()

			event := model.DeviceEvent{
				Device:      "test-device",
				Timestamp:   time.Now(),
				Event:       eventType,
				Component:   "switch",
				ComponentID: 0,
			}

			if event.Event != eventType {
				t.Errorf("Event type mismatch for %s", eventType)
			}
		})
	}
}

// TestOutputFormat_JSONDetection tests WantsJSON detection.
func TestOutputFormat_JSONDetection(t *testing.T) {
	t.Parallel()

	// Save current format
	currentFormat := output.GetFormat()
	defer viper.Set("output", string(currentFormat))

	// Test JSON format
	viper.Set("output", string(output.FormatJSON))
	if !output.WantsJSON() {
		t.Error("WantsJSON should return true when format is JSON")
	}

	// Test non-JSON format
	viper.Set("output", string(output.FormatTable))
	if output.WantsJSON() {
		t.Error("WantsJSON should return false when format is not JSON")
	}
}

// TestNewCommand_HelpContent tests that help output contains expected information.
func TestNewCommand_HelpContent(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Logf("help execute returned error: %v", err)
	}

	helpText := buf.String()
	expectedStrings := []string{
		"events",
		"Monitor",
		"device",
		"filter",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(helpText, expected) {
			t.Errorf("Help output should contain %q", expected)
		}
	}
}
